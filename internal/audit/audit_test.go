package audit

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuditWriter is a mock implementation of AuditWriter for testing
type MockAuditWriter struct {
	events []AuditEvent
}

func (m *MockAuditWriter) Write(_ context.Context, e AuditEvent) {
	m.events = append(m.events, e)
}

func TestAuditManager_LogEvent(t *testing.T) {
	// Create a test context
	ctx := context.Background()

	// Create a mock writer
	mockWriter := &MockAuditWriter{}

	// Create audit manager and register the mock writer
	manager := NewAuditManager()
	manager.RegisterWriter(mockWriter)

	// Test data
	action := "test_action"
	userID := "test_user"
	url := "http://example.com"

	// Log an event
	beforeLog := time.Now().Unix()
	manager.LogEvent(ctx, action, userID, url)
	afterLog := time.Now().Unix()

	// Wait a bit for the async write to complete
	time.Sleep(100 * time.Millisecond)

	// Verify the event was logged
	require.Len(t, mockWriter.events, 1)
	event := mockWriter.events[0]
	assert.Equal(t, action, event.Action)
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, url, event.URL)
	assert.GreaterOrEqual(t, event.TimeStamp, int(beforeLog))
	assert.LessOrEqual(t, event.TimeStamp, int(afterLog))
}

func TestAuditManager_ConcurrentWrites(t *testing.T) {
	ctx := context.Background()
	manager := NewAuditManager()

	// Use an atomic counter to track the number of events written
	var eventCount int32

	// Create multiple writers
	for i := 0; i < 5; i++ {
		writer := &countingWriter{count: &eventCount}
		manager.RegisterWriter(writer)
	}

	// Log multiple events concurrently
	for i := 0; i < 100; i++ {
		go manager.LogEvent(ctx, "concurrent_action", "user1", "http://example.com")
	}

	// Wait for all events to be processed
	time.Sleep(500 * time.Millisecond)

	// Verify all events were processed
	assert.Equal(t, int32(500), eventCount) // 100 events * 5 writers
}

// countingWriter is a test implementation of AuditWriter that counts events
type countingWriter struct {
	count *int32
}

func (w *countingWriter) Write(ctx context.Context, e AuditEvent) {
	atomic.AddInt32(w.count, 1)
}

func TestFileAudit_Write(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "audit.log")

	// Create file audit
	fileAudit := NewFileAudit(filePath)

	// Test data
	ctx := context.Background()
	event := AuditEvent{
		TimeStamp: int(time.Now().Unix()),
		Action:    "test_action",
		UserID:    "test_user",
		URL:       "http://example.com",
	}

	// Write event
	fileAudit.Write(ctx, event)

	// Read the file
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// Parse the JSON
	var loggedEvent AuditEvent
	err = json.Unmarshal(data, &loggedEvent)
	require.NoError(t, err)

	// Verify the event was written correctly
	assert.Equal(t, event, loggedEvent)
}

func TestFileAudit_WriteError(t *testing.T) {
	// Create a file in a non-existent directory to cause an error
	fileAudit := NewFileAudit("/non/existent/path/audit.log")

	// This should not panic
	fileAudit.Write(context.Background(), AuditEvent{})
}

func TestRemoteAudit_Write(t *testing.T) {
	// Start a test HTTP server
	server := startTestHTTPServer(t)
	defer server.Close()

	// Create remote audit with test server URL
	remoteAudit := NewRemoteAudit(server.URL)

	// Test data
	ctx := context.Background()
	event := AuditEvent{
		TimeStamp: int(time.Now().Unix()),
		Action:    "test_action",
		UserID:    "test_user",
		URL:       "http://example.com",
	}

	// Write event
	remoteAudit.Write(ctx, event)

	// The test server will verify the request
}

// testHTTPServer is a simple HTTP server for testing RemoteAudit
type testHTTPServer struct {
	t             *testing.T
	server        *http.Server
	URL           string
	eventReceived chan AuditEvent
}

func startTestHTTPServer(t *testing.T) *testHTTPServer {
	s := &testHTTPServer{
		t:             t,
		eventReceived: make(chan AuditEvent, 1),
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Verify content type
		contentType := r.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType)

		// Parse request body
		var event AuditEvent
		err := json.NewDecoder(r.Body).Decode(&event)
		assert.NoError(t, err)

		// Send the event to the channel
		select {
		case s.eventReceived <- event:
		default:
			t.Log("Event channel full, dropping event")
		}

		// Verify required fields
		assert.NotEmpty(t, event.Action)
		assert.NotEmpty(t, event.UserID)
		assert.NotEmpty(t, event.URL)
		assert.NotZero(t, event.TimeStamp)

		w.WriteHeader(http.StatusOK)
	})

	s.server = &http.Server{
		Addr:    ":0", // Let the OS choose a free port
		Handler: handler,
	}

	// Start the server in a goroutine
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s.URL = "http://" + listener.Addr().String()

	go func() {
		err := s.server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			t.Logf("Test server error: %v", err)
		}
	}()

	return s
}

func (s *testHTTPServer) Close() {
	if s.server != nil {
		s.server.Close()
	}
}
