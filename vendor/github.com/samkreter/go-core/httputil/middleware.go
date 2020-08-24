package httputil

import (
	"net/http"
	"time"

	"github.com/samkreter/go-core/correlation"
	"github.com/samkreter/go-core/log"

	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
)

const (
	incomingRequestStart = "HttpIncomingRequestStart"
	incomingRequestEnd   = "HttpIncomingRequestStart"
)

// HandlerConfig holds configuration for which middle to enable
type HandlerConfig struct {
	CorrelationEnabled bool
	LoggingEnabled     bool
	TracingEnabled     bool
}

// SetUpHandler adds logging, tracing and correlation for incoming requests
func SetUpHandler(handler http.Handler, config *HandlerConfig) http.Handler {

	// Adding distributed tracing
	if config.TracingEnabled {
		handler = TracingMiddleware(handler)
	}

	// Add incoming request logging
	if config.LoggingEnabled {
		handler = IncomingRequestLoggingMiddleware(handler)
	}

	// Add correlation propogation
	// Note(sakreter) this must be the last handler returned to ensure the correlation
	// information is in the context for the following handlers
	if config.CorrelationEnabled {
		handler = CorrelationMiddleware(handler)
	}

	return handler
}

// TracingMiddleware adds tracing Middleware to the handler
func TracingMiddleware(handler http.Handler) http.Handler {
	return &ochttp.Handler{
		Handler:     handler,
		Propagation: &b3.HTTPFormat{}}
}

// CorrelationMiddleware adds correlation Middleware to the handler
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		corrCtx := correlation.CreateCtxFromRequest(req)

		next.ServeHTTP(w, req.WithContext(corrCtx))
	})
}

// IncomingRequestLoggingMiddleware add incoming request logging to the handler
// TODO(sakreter): add support for operationName and apiVersion
func IncomingRequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		fields := logrus.Fields{
			"httpMethod":    req.Method,
			"targetUri":     req.URL.String(),
			"hostName":      req.Host,
			"correlationID": correlation.GetCorrelationID(ctx),
			"activityID":    correlation.GetActivityID(ctx),
			"taskName":      "StartIncomingRequest",
		}

		contentType := req.Header.Get("Content-Type")
		if contentType != "" {
			fields["contentType"] = contentType
		}

		rr := &responseRecorder{w: w}

		log.G(ctx).WithFields(fields).Info("Incoming request Start")

		startTime := time.Now()

		defer func() {
			fields["contentLength"] = rr.contentLength
			fields["httpStatusCode"] = rr.statusCode
			fields["durationInMilliseconds"] = time.Now().Sub(startTime)

			log.G(ctx).WithFields(fields).Info("Incoming request End")
		}()

		next.ServeHTTP(rr, req.WithContext(ctx))
	})
}

type responseRecorder struct {
	contentLength int
	statusCode    int
	w             http.ResponseWriter
}

func (r *responseRecorder) Header() http.Header { return r.w.Header() }

func (r *responseRecorder) Write(p []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	n, err := r.w.Write(p)
	r.contentLength += n
	return n, err
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.w.WriteHeader(statusCode)
}
