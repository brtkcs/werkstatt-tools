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

// memStore is an in-memory storage for testing
type memStore struct {
	data map[string]string
}

func newMemStore() *memStore {
	return &memStore{data: make(map[string]string)}
}

func (m *memStore) get(key string) (string, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (m *memStore) set(key, value string) {
	m.data[key] = value
}

func (m *memStore) del(key string) bool {
	if _, ok := m.data[key]; !ok {
		return false
	}
	delete(m.data, key)
	return true
}

func (m *memStore) all() map[string]string {
	cp := make(map[string]string, len(m.data))
	for k, v := range m.data {
		cp[k] = v
	}
	return cp
}

func (m *memStore) count() int {
	return len(m.data)
}

func TestKeyHandlerSet(t *testing.T) {
	store := newMemStore()
	handler := keyHandler(store)

	body := `{"value":"secret123"}`
	req := httptest.NewRequest("POST", "/keys/apikey", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got status %d, want 201", rec.Code)
	}

	v, ok := store.get("apikey")
	if !ok || v != "secret123" {
		t.Errorf("got %q, want secret123", v)
	}
}

func TestKeyHandlerGet(t *testing.T) {
	store := newMemStore()
	store.set("token", "abc")
	handler := keyHandler(store)

	req := httptest.NewRequest("GET", "/keys/token", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["value"] != "abc" {
		t.Errorf("got value %q, want abc", resp["value"])
	}
}

func TestKeyHandlerGetNotFound(t *testing.T) {
	store := newMemStore()
	handler := keyHandler(store)

	req := httptest.NewRequest("GET", "/keys/missing", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want 404", rec.Code)
	}
}

func TestKeyHandlerDelete(t *testing.T) {
	store := newMemStore()
	store.set("temp", "value")
	handler := keyHandler(store)

	req := httptest.NewRequest("DELETE", "/keys/temp", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}

	if _, ok := store.get("temp"); ok {
		t.Error("key should be deleted")
	}
}

func TestKeyHandlerDeleteNotFound(t *testing.T) {
	store := newMemStore()
	handler := keyHandler(store)

	req := httptest.NewRequest("DELETE", "/keys/missing", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want 404", rec.Code)
	}
}

func TestKeyHandlerMissingKey(t *testing.T) {
	store := newMemStore()
	handler := keyHandler(store)

	req := httptest.NewRequest("GET", "/keys/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want 400", rec.Code)
	}
}

func TestKeyHandlerInvalidJSON(t *testing.T) {
	store := newMemStore()
	handler := keyHandler(store)

	req := httptest.NewRequest("POST", "/keys/test", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want 400", rec.Code)
	}
}

func TestKeyHandlerMethodNotAllowed(t *testing.T) {
	store := newMemStore()
	handler := keyHandler(store)

	req := httptest.NewRequest("PUT", "/keys/test", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got status %d, want 405", rec.Code)
	}
}

func TestListHandler(t *testing.T) {
	store := newMemStore()
	store.set("a", "1")
	store.set("b", "2")
	handler := listHandler(store)

	req := httptest.NewRequest("GET", "/keys", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("got %d keys, want 2", len(resp))
	}
}

func TestListHandlerEmpty(t *testing.T) {
	store := newMemStore()
	handler := listHandler(store)

	req := httptest.NewRequest("GET", "/keys", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("got %d keys, want 0", len(resp))
	}
}

func TestFileStorePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.json")

	s1 := newFileStore(path)
	s1.set("key1", "value1")
	s1.set("key2", "value2")

	// Load into a new store from same file
	s2 := newFileStore(path)
	v, ok := s2.get("key1")
	if !ok || v != "value1" {
		t.Errorf("got %q, want value1", v)
	}
	if s2.count() != 2 {
		t.Errorf("got %d keys, want 2", s2.count())
	}
}

func TestFileStoreDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.json")

	s := newFileStore(path)
	s.set("temp", "data")
	s.del("temp")

	// Verify file reflects deletion
	raw, _ := os.ReadFile(path)
	var data map[string]string
	json.Unmarshal(raw, &data)
	if _, ok := data["temp"]; ok {
		t.Error("key should be deleted from file")
	}
}
