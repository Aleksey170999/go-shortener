// Package middlewares provides HTTP middleware functions for the application.
// This file implements request/response compression using gzip.
package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter wraps http.ResponseWriter to provide gzip compression.
// It implements http.ResponseWriter interface and can be used to compress
// HTTP responses on-the-fly.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter creates a new compressWriter that wraps the provided
// http.ResponseWriter. The returned writer will compress all data written to it
// using gzip compression.
//
// Parameters:
//   - w: The original http.ResponseWriter to wrap
//
// Returns:
//   - *compressWriter: A new compressWriter instance
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the header map of the underlying http.ResponseWriter.
// This method is part of the http.ResponseWriter interface implementation.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes compressed data to the underlying gzip.Writer.
// Implements the io.Writer interface.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code.
// It sets the Content-Encoding header to gzip if not already set.
// Implements the http.ResponseWriter interface.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close flushes any pending compressed data and closes the gzip.Writer.
// This method should be called to ensure all data is properly written.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader wraps an io.ReadCloser to provide gzip decompression.
// It implements the io.ReadCloser interface and can be used to decompress
// gzip-encoded request bodies.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader creates a new compressReader that wraps the provided
// io.ReadCloser. The returned reader will decompress gzipped data read from it.
//
// Parameters:
//   - r: The original io.ReadCloser to wrap
//
// Returns:
//   - *compressReader: A new compressReader instance
//   - error: An error if the gzip reader cannot be created
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads decompressed data from the underlying gzip.Reader.
// Implements the io.Reader interface.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the gzip.Reader and the underlying io.ReadCloser.
// This method should always be called to prevent resource leaks.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware is an HTTP middleware that provides transparent gzip compression
// for HTTP responses and decompression for HTTP requests.
//
// For responses:
//   - Checks if the client accepts gzip encoding (Accept-Encoding: gzip)
//   - If so, compresses the response body and sets appropriate headers
//
// For requests:
//   - Checks if the request body is gzipped (Content-Encoding: gzip)
//   - If so, decompresses the request body transparently
//
// The middleware preserves the original request/response when compression is
// not needed or not supported.
func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasGzipEncoding(r.Header.Get("Content-Encoding")) {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip data", http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer func() { _ = cr.Close() }()
		}

		ow := w
		supportsGzip := hasGzipEncoding(r.Header.Get("Accept-Encoding"))
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer func() { _ = cw.Close() }()
		}

		h.ServeHTTP(ow, r)
	})
}

// hasGzipEncoding checks if the given header string contains 'gzip' encoding.
// It's used to check Accept-Encoding and Content-Encoding headers.
//
// Parameters:
//   - header: The header value to check (e.g., "gzip, deflate, br")
//
// Returns:
//   - bool: true if 'gzip' is present in the header, false otherwise
func hasGzipEncoding(header string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.TrimSpace(strings.ToLower(part)) == "gzip" {
			return true
		}
	}
	return false
}
