package audit

import "context"

type AuditEvent struct {
	TimeStamp int    `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id"`
	URL       string `json:"url"`
}
type AuditWriter interface {
	Write(ctx context.Context, e AuditEvent)
}
