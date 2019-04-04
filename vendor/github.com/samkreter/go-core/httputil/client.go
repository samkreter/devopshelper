package httputil

import (
	"net/http"
	"time"

	"github.com/samkreter/go-core/correlation"
	"github.com/samkreter/go-core/log"

	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
)

const (
	defaultHTTPClientTimeout = time.Second * 30
)

// NewHTTPClient creates a new http client with tracing and logging enabled
func NewHTTPClient(correlationEnabled, loggingEnabled, tracingEnabled bool) *http.Client {
	var transport http.RoundTripper
	transport = &http.Transport{}

	// Add outgoing request logging transport
	if loggingEnabled {
		transport = &LogTransport{
			Transport: transport,
		}
	}

	// Add correlation propegation transport
	if correlationEnabled {
		transport = &CorrelationTransport{
			Transport: transport,
		}
	}

	// Add tracing transport
	if tracingEnabled {
		transport = &ochttp.Transport{
			Base: transport,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   defaultHTTPClientTimeout,
	}
}

// CorrelationTransport implements http.RoundTripper.
// When set as Transport of http.Client, it executes HTTP requests with correlation propegation.
type CorrelationTransport struct {
	Transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper and adds correlation propegation the client requests
func (t *CorrelationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	correlation.AddHeadersFromContext(req.Context(), req)
	if t.Transport != nil {
		return t.Transport.RoundTrip(req)
	}

	return http.DefaultTransport.RoundTrip(req)
}

// LogTransport implements http.RoundTripper.
// When set as Transport of http.Client, it executes HTTP requests with logging.
type LogTransport struct {
	Transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper and adds logging the client requests
func (t *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	endLogging := StartLogOutgoingRequest(req)

	resp, err := t.transport().RoundTrip(req)

	endLogging(resp, err)

	return resp, err
}

func (t *LogTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}

// StartLogOutgoingRequest logs the outgoing requests. Returns the end function when
// the request is finished
func StartLogOutgoingRequest(req *http.Request) (endReqLog func(resp *http.Response, err error)) {
	ctx := req.Context()

	fields := logrus.Fields{
		"httpMethod":    req.Method,
		"targetUri":     req.URL.String(),
		"hostName":      req.Host,
		"correlationID": correlation.GetCorrelationID(ctx),
		"activityID":    correlation.GetActivityID(ctx),
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "" {
		fields["contentType"] = contentType
	}

	log.G(ctx).WithFields(fields).Debug("Outgoing Http Request Started")

	startTime := time.Now()

	return func(resp *http.Response, err error) {
		fields["contentLength"] = resp.ContentLength
		fields["httpStatusCode"] = resp.StatusCode
		fields["durationInMilliseconds"] = time.Now().Sub(startTime)

		log.G(ctx).WithFields(fields).Debug("Outgoing Http Request Ended")
	}
}
