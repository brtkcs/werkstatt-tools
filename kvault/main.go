package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	cOlive = "\033[38;2;168;200;48m"
	cRust  = "\033[38;2;255;107;74m"
	cMuted = "\033[38;2;88;80;72m"
	cReset = "\033[0m"
)

// storage abstracts key-value persistence for testing
type storage interface {
	get(key string) (string, bool)
	set(key, value string)
	del(key string) bool
	all() map[string]string
	count() int
}

// fileStore is the real implementation backed by a JSON file
type fileStore struct {
	mu   sync.RWMutex
	data map[string]string
	path string
}

func newFileStore(path string) *fileStore {
	s := &fileStore{
		data: make(map[string]string),
		path: path,
	}
	s.load()
	return s
}

func (s *fileStore) load() {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	json.Unmarshal(raw, &s.data)
}

func (s *fileStore) save() {
	raw, _ := json.MarshalIndent(s.data, "", "  ")
	os.WriteFile(s.path, raw, 0o644)
}

func (s *fileStore) get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *fileStore) set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	s.save()
}

func (s *fileStore) del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; !ok {
		return false
	}
	delete(s.data, key)
	s.save()
	return true
}

func (s *fileStore) all() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]string, len(s.data))
	for k, v := range s.data {
		cp[k] = v
	}
	return cp
}

func (s *fileStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func main() {
	port := flag.String("port", "8080", "listen port")
	dataFile := flag.String("data", "kvault.json", "path to data file")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "kvault - key-value store with REST API\n\n")
		fmt.Fprintf(os.Stderr, "Usage: kvault [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Endpoints:\n")
		fmt.Fprintf(os.Stderr, "  GET    /keys         list all keys\n")
		fmt.Fprintf(os.Stderr, "  GET    /keys/{key}   get value\n")
		fmt.Fprintf(os.Stderr, "  POST   /keys/{key}   set value {\"value\":\"...\"}\n")
		fmt.Fprintf(os.Stderr, "  DELETE /keys/{key}   delete key\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *noColor {
		cOlive, cRust, cMuted, cReset = "", "", "", ""
	}

	store := newFileStore(*dataFile)

	http.HandleFunc("/keys/", keyHandler(store))
	http.HandleFunc("/keys", listHandler(store))

	addr := ":" + *port
	fmt.Printf("\n%skvault%s %s->%s %s (%d keys loaded)\n\n", cOlive, cReset, cMuted, cReset, addr, store.count())

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", cRust, cReset, err)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func keyHandler(store storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/keys/")
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			value, ok := store.get(key)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"key": key, "value": value})

		case http.MethodPost:
			var body struct {
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			store.set(key, body.Value)
			fmt.Printf("%s SET%s %s%s%s = %s\n", cOlive, cReset, cMuted, key, cReset, body.Value)
			writeJSON(w, http.StatusCreated, map[string]string{"key": key, "value": body.Value})

		case http.MethodDelete:
			if !store.del(key) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			fmt.Printf("%s DEL%s %s%s%s\n", cRust, cReset, cMuted, key, cReset)
			writeJSON(w, http.StatusOK, map[string]string{"key": key, "deleted": "true"})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listHandler(store storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.all())
	}
}
