package audit

import (
	"context"
	"sync"
	"time"
)

type AuditManager struct {
	writers []AuditWriter
	mu      sync.Mutex
}

func NewAuditManager() *AuditManager {
	return &AuditManager{
		writers: make([]AuditWriter, 0),
	}
}

func (am *AuditManager) RegisterWriter(writer AuditWriter) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.writers = append(am.writers, writer)
}

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
