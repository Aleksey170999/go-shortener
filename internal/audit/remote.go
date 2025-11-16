package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type RemoteAudit struct {
	url        string
	httpClient *http.Client
}

func NewRemoteAudit(url string) *RemoteAudit {
	return &RemoteAudit{
		url: url,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

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
