package audit

import (
	"context"
	"sync"
	"time"
)

// AuditManager coordinates multiple AuditWriter instances to handle audit logging.
// It provides thread-safe registration of writers and concurrent event logging.
type AuditManager struct {
	writers []AuditWriter // List of registered audit writers
	mu      sync.Mutex    // Mutex to protect concurrent access to writers slice
}

// NewAuditManager creates and initializes a new AuditManager instance.
// The returned manager starts with no registered writers; use RegisterWriter to add them.
func NewAuditManager() *AuditManager {
	return &AuditManager{
		writers: make([]AuditWriter, 0),
	}
}

// RegisterWriter adds a new AuditWriter to the list of writers that will receive audit events.
// This method is thread-safe and can be called concurrently.
func (am *AuditManager) RegisterWriter(writer AuditWriter) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.writers = append(am.writers, writer)
}

// LogEvent creates and dispatches an audit event to all registered writers.
// The event is sent asynchronously to each writer, and context cancellation is respected.
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - action: The type of action being logged (e.g., "url_created", "url_deleted")
//   - userID: ID of the user who performed the action
//   - url: The URL that was affected by the action
func (am *AuditManager) LogEvent(ctx context.Context, action, userID, url string) {
	event := AuditEvent{
		TimeStamp: int(time.Now().Unix()),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}

	am.mu.Lock()
	writers := make([]AuditWriter, len(am.writers))
	copy(writers, am.writers)
	am.mu.Unlock()

	for _, writer := range writers {
		go writer.Write(ctx, event)
	}
}
