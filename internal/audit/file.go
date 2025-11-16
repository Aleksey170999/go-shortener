package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type FileAudit struct {
	filePath string
	mu       sync.Mutex
}

func NewFileAudit(filePath string) *FileAudit {
	return &FileAudit{
		filePath: filePath,
	}
}

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
