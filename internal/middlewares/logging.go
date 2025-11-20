// Package middlewares provides HTTP middleware functions for the application.
// This file implements request logging functionality using zap logger.
package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData holds information about the HTTP response that will be logged.
// It tracks the status code and response size for logging purposes.
type responseData struct {
	status int // HTTP status code
	size   int // Response body size in bytes
}

// loggingResponseWriter wraps http.ResponseWriter to capture response information.
// It implements http.ResponseWriter interface to intercept Write and WriteHeader calls.
type loggingResponseWriter struct {
	http.ResponseWriter               // Embedded ResponseWriter
	responseData        *responseData // Pointer to store response metadata
}

// Write intercepts calls to the underlying ResponseWriter's Write method.
// It captures the response size and ensures a default status code is set.
// Implements the io.Writer interface.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.responseData.status == 0 {
		r.responseData.status = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader intercepts calls to the underlying ResponseWriter's WriteHeader method.
// It captures the status code for logging purposes.
// Implements the http.ResponseWriter interface.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging creates a middleware that logs HTTP request details using the provided zap.Logger.
// The middleware logs the following information for each request:
//   - HTTP method
//   - Request URI
//   - Response status code
//   - Response size in bytes
//   - Request duration in milliseconds
//
// The middleware wraps the response writer to capture the status code and response size.
//
// Parameters:
//   - logger: A configured zap.Logger instance for structured logging
//
// Returns:
//   - A middleware function that can be used with http.Handler
func WithLogging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			uri := r.URL.RequestURI()
			if uri == "" {
				uri = r.URL.Path
			}

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			next.ServeHTTP(&lw, r)

			duration := time.Since(start)
			logger.Sugar().Infow("request completed",
				"method", r.Method,
				"uri", uri,
				"status", responseData.status,
				"response_size", responseData.size,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}
