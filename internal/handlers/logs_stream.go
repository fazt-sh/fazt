package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
)

// LogStreamManager manages SSE connections for log streaming
type LogStreamManager struct {
	mu       sync.RWMutex
	channels map[string][]chan LogEvent
}

// LogEvent represents a log event for streaming
type LogEvent struct {
	ID        int64  `json:"id"`
	SiteID    string `json:"site_id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	CreatedAt string `json:"created_at"`
}

var logManager = &LogStreamManager{
	channels: make(map[string][]chan LogEvent),
}

// GetLogStreamManager returns the global log stream manager
func GetLogStreamManager() *LogStreamManager {
	return logManager
}

// Subscribe adds a channel to receive logs for a site
func (m *LogStreamManager) Subscribe(siteID string, ch chan LogEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[siteID] = append(m.channels[siteID], ch)
}

// Unsubscribe removes a channel from a site's log stream
func (m *LogStreamManager) Unsubscribe(siteID string, ch chan LogEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	channels := m.channels[siteID]
	for i, c := range channels {
		if c == ch {
			m.channels[siteID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
}

// Broadcast sends a log event to all subscribers
func (m *LogStreamManager) Broadcast(event LogEvent) {
	m.mu.RLock()
	channels := m.channels[event.SiteID]
	m.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// PersistLog saves a log entry to the database and broadcasts it
func PersistLog(siteID, level, message string) error {
	db := database.GetDB()
	if db == nil {
		return nil
	}

	result, err := db.Exec(`
		INSERT INTO site_logs (site_id, level, message)
		VALUES (?, ?, ?)
	`, siteID, level, message)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	event := LogEvent{
		ID:        id,
		SiteID:    siteID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	logManager.Broadcast(event)
	return nil
}

// LogStreamHandler handles SSE connections for log streaming
func LogStreamHandler(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		http.Error(w, "site_id required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create channel for this connection
	ch := make(chan LogEvent, 100)
	logManager.Subscribe(siteID, ch)
	defer logManager.Unsubscribe(siteID, ch)
	defer close(ch)

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"site_id\":\"%s\"}\n\n", siteID)
	flusher.Flush()

	// Stream logs
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: log\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}
