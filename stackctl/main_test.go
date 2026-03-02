package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	cOlive, cRust, cMuted, cReset = "", "", "", ""
}

func TestHealthHandler(t *testing.T) {
	cfg := config{port: "8080", runtime: "podman"}
	handler := healthHandler(cfg)

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

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("got status %q, want ok", resp.Status)
	}
	if resp.Runtime != "podman" {
		t.Errorf("got runtime %q, want podman", resp.Runtime)
	}
}

func TestHealthHandlerDocker(t *testing.T) {
	cfg := config{port: "9090", runtime: "docker"}
	handler := healthHandler(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	var resp HealthResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Runtime != "docker" {
		t.Errorf("got runtime %q, want docker", resp.Runtime)
	}
}

func TestInfoHandler(t *testing.T) {
	handler := infoHandler()

	req := httptest.NewRequest("GET", "/info", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}

	var resp InfoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.System == "" {
		t.Error("system field is empty")
	}
}

func TestStacksHandlerNoRuntime(t *testing.T) {
	cfg := config{port: "8080", runtime: "nonexistent-runtime"}
	handler := stacksHandler(cfg)

	req := httptest.NewRequest("GET", "/stacks", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want 500 for missing runtime", rec.Code)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	writeJSON(rec, data)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("got Content-Type %q, want application/json", ct)
	}

	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("got %q, want value", result["key"])
	}
}

func TestRunCmd(t *testing.T) {
	result := runCmd("echo", "hello")
	if result != "hello" {
		t.Errorf("got %q, want hello", result)
	}
}

func TestRunCmdFails(t *testing.T) {
	result := runCmd("nonexistent-command-xyz")
	if result != "" {
		t.Errorf("got %q, want empty for failed command", result)
	}
}
