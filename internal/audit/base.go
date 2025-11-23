package audit

import "context"

// AuditEvent represents an audit log entry containing information about a user action.
// It includes the timestamp, action type, user ID, and the URL involved.
type AuditEvent struct {
	TimeStamp int    `json:"ts"`      // Unix timestamp of when the event occurred
	Action    string `json:"action"`  // The action performed (e.g., "create", "delete", "update")
	UserID    string `json:"user_id"` // ID of the user who performed the action
	URL       string `json:"url"`     // The URL that was affected by the action
}

// AuditWriter defines the interface for writing audit events to a specific destination.
// Implementations should handle the actual writing logic, such as file I/O or network requests.
type AuditWriter interface {
	Write(ctx context.Context, e AuditEvent)
}
