package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

// FileAudit implements the AuditWriter interface for writing audit events to a file.
// It provides thread-safe file operations with proper resource management.
type FileAudit struct {
	filePath string     // Path to the audit log file
	mu       sync.Mutex // Mutex to ensure thread-safe file operations
}

// NewFileAudit creates a new FileAudit instance with the specified file path.
// The file will be created if it doesn't exist, and new entries will be appended to it.
func NewFileAudit(filePath string) *FileAudit {
	return &FileAudit{
		filePath: filePath,
	}
}

// Write persists an audit event to the log file in JSON format.
// It handles context cancellation and ensures thread-safe file operations.
// Each event is written as a new line in the file.
func (a *FileAudit) Write(ctx context.Context, e AuditEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	select {
	case <-ctx.Done():
		return
	default:
		file, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer file.Close()

		enc := json.NewEncoder(file)
		enc.Encode(e)
	}
}
