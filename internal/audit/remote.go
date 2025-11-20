package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// RemoteAudit implements the AuditWriter interface for sending audit events to a remote HTTP endpoint.
// It uses a configurable HTTP client with timeout settings for reliable event delivery.
type RemoteAudit struct {
	url        string       // The target URL where audit events will be sent
	httpClient *http.Client // HTTP client with configured timeout settings
}

// NewRemoteAudit creates a new RemoteAudit instance with the specified endpoint URL.
// The HTTP client is configured with a 5-second timeout by default.
// The URL should be the full endpoint where audit events should be posted.
func NewRemoteAudit(url string) *RemoteAudit {
	return &RemoteAudit{
		url: url,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Write sends an audit event to the configured remote endpoint as a JSON payload.
// The request includes proper content-type headers and handles context cancellation.
// Failures during the HTTP request or response are silently ignored to prevent
// blocking the main application flow.
func (a *RemoteAudit) Write(ctx context.Context, e AuditEvent) {
	select {
	case <-ctx.Done():
		return
	default:
		jsonData, err := json.Marshal(e)
		if err != nil {
			return
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			a.url,
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.httpClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}
}
