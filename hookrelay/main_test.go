package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	cOlive, cRust, cMuted, cReset = "", "", "", ""
}

// mockLogger captures saved webhooks for testing
type mockLogger struct {
	saved []Webhook
	err   error
}

func (m *mockLogger) save(hook Webhook) error {
	if m.err != nil {
		return m.err
	}
	m.saved = append(m.saved, hook)
	return nil
}

func TestHealthHandler(t *testing.T) {
	handler := healthHandler()

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("got Content-Type %q, want application/json", ct)
	}
}

func TestWebhookHandlerValid(t *testing.T) {
	mock := &mockLogger{}
	handler := webhookHandler(mock)

	body := `{"event":"push","repo":"werkstatt","branch":"main"}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}

	if len(mock.saved) != 1 {
		t.Fatalf("expected 1 saved webhook, got %d", len(mock.saved))
	}
	if mock.saved[0].Event != "push" {
		t.Errorf("got event %q, want push", mock.saved[0].Event)
	}
	if mock.saved[0].Repo != "werkstatt" {
		t.Errorf("got repo %q, want werkstatt", mock.saved[0].Repo)
	}
}

func TestWebhookHandlerGetRejected(t *testing.T) {
	mock := &mockLogger{}
	handler := webhookHandler(mock)

	req := httptest.NewRequest("GET", "/webhook", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got status %d, want 405", rec.Code)
	}
}

func TestWebhookHandlerInvalidJSON(t *testing.T) {
	mock := &mockLogger{}
	handler := webhookHandler(mock)

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want 400", rec.Code)
	}
}

func TestWebhookHandlerMissingEvent(t *testing.T) {
	mock := &mockLogger{}
	handler := webhookHandler(mock)

	body := `{"repo":"werkstatt","branch":"main"}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want 400 for missing event", rec.Code)
	}
}

func TestWebhookHandlerEmptyBody(t *testing.T) {
	mock := &mockLogger{}
	handler := webhookHandler(mock)

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(""))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want 400 for empty body", rec.Code)
	}
}

func TestFileLoggerSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	log := fileLogger{path: path}

	hook := Webhook{Event: "push", Repo: "test", Branch: "main"}
	if err := log.save(hook); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var entry LogEntry
	if err := json.Unmarshal(data[:len(data)-1], &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if entry.Event != "push" {
		t.Errorf("got event %q, want push", entry.Event)
	}
	if entry.Repo != "test" {
		t.Errorf("got repo %q, want test", entry.Repo)
	}
}

func TestFileLoggerAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	log := fileLogger{path: path}

	log.save(Webhook{Event: "push", Repo: "one", Branch: "main"})
	log.save(Webhook{Event: "tag", Repo: "two", Branch: "v1.0"})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("got %d lines, want 2", len(lines))
	}
}
