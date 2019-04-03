package correlation

import (
	"context"
	"net/http"

	"github.com/samkreter/go-core/log"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	// activityIDContextKey the activityID context key
	activityIDContextKey = contextKey("activityID")

	// correlationIDContextKey  the correlationID context key
	correlationIDContextKey = contextKey("correlationID")

	//contextMetadataHeadersConextKey the ms request headers
	contextMetadataHeadersConextKey = contextKey("contextMetadataHeaders")
)

var (
	// CorrelationIDHeader the correlation ID Header
	CorrelationIDHeader = "correlation-request-id"

	// UserAgentHeader the UserAgent header
	UserAgentHeader = "User-Agent"

	// AcceptedLanguageHeader the AcceptedLanguage header
	AcceptedLanguageHeader = "Accept-Language"

	// RequestIDHeader stores the request id for the request
	RequestIDHeader = "request-id"
)

// ContextMatadataHeaders stores header information into the context
type ContextMatadataHeaders map[string]string

// Add adds non emtpy string values to the metadata
func (m ContextMatadataHeaders) Add(req *http.Request, key string) {
	val := req.Header.Get(key)
	if val != "" {
		m[key] = val
	}
}

// Get returns the value related to the given key.
func (m ContextMatadataHeaders) Get(key string) string {
	if val, ok := m[key]; ok {
		return val
	}

	return ""
}

// FromReq populates the metadataHeader from an http request
func (m ContextMatadataHeaders) FromReq(req *http.Request) {
	m.Add(req, UserAgentHeader)
	m.Add(req, AcceptedLanguageHeader)
}

func generateGUID() string {
	guid := uuid.NewV4()
	return guid.String()
}

// AddCorrelationLogger adds the correlation information to the context logger
func AddCorrelationLogger(ctx context.Context) context.Context {
	logger := log.G(ctx).WithFields(logrus.Fields{
		"correlationID": GetCorrelationID(ctx),
		"activityID":    GetActivityID(ctx),
	})

	return log.WithLogger(ctx, logger)
}

// SetCorrelationID sets the correlation ID in the context
func SetCorrelationID(ctx context.Context, correlationID string) context.Context {
	ctx = context.WithValue(ctx, correlationIDContextKey, correlationID)
	return AddCorrelationLogger(ctx)
}

// GetCorrelationID gets the correlation ID from a context
func GetCorrelationID(ctx context.Context) string {
	correlationID, ok := ctx.Value(correlationIDContextKey).(string)
	if !ok {
		return ""
	}
	return correlationID
}

// SetActivityID sets the current activity ID in the context or generates a new one
func SetActivityID(ctx context.Context, activityID string) context.Context {
	if activityID == "" {
		activityID = generateGUID()
	}

	ctx = context.WithValue(ctx, activityIDContextKey, activityID)

	// Update logger with the updated activity ID
	return AddCorrelationLogger(ctx)
}

// GetActivityID gets the activity ID from a context
func GetActivityID(ctx context.Context) string {
	activityID, ok := ctx.Value(activityIDContextKey).(string)
	if !ok {
		return ""
	}
	return activityID
}

// GetMetadataHeaders gets the metadata headers from a context
func GetMetadataHeaders(ctx context.Context) ContextMatadataHeaders {
	metadataHeaders, ok := ctx.Value(contextMetadataHeadersConextKey).(ContextMatadataHeaders)
	if !ok {
		return nil
	}
	return metadataHeaders
}

// CreateCtxFromRequest serialize http request headers into a context
func CreateCtxFromRequest(req *http.Request) context.Context {
	ctx := req.Context()

	correlationID := req.Header.Get(CorrelationIDHeader)
	if correlationID == "" {
		correlationID = generateGUID()
	}

	ctx = SetCorrelationID(ctx, correlationID)

	ctx = SetActivityID(ctx, "")

	metadataHeaders := make(ContextMatadataHeaders)
	metadataHeaders.FromReq(req)
	ctx = context.WithValue(ctx, contextMetadataHeadersConextKey, metadataHeaders)

	// Add correlation fields to the logger
	ctx = AddCorrelationLogger(ctx)

	return ctx
}

// AddHeadersFromContext Add metadata headers from the context into the request headers
// only adds headers that are not already set.
func AddHeadersFromContext(ctx context.Context, req *http.Request) {
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		req.Header.Set(CorrelationIDHeader, correlationID)
	}

	req.Header.Set(RequestIDHeader, generateGUID())

	metadataHeaders := GetMetadataHeaders(ctx)
	if metadataHeaders == nil {
		return
	}

	for key, val := range metadataHeaders {
		if val != "" && req.Header.Get(key) == "" {
			req.Header.Set(key, val)
		}
	}
}
