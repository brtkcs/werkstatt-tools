package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	olive    = "\033[38;2;168;200;48m"
	red      = "\033[38;2;255;107;74m"
	muted    = "\033[38;2;88;80;72m"
	reset    = "\033[0m"
	dataFile = "kvault.json"
)

var store = make(map[string]string)

// save – memóriából fájlba
func save() {
	data, _ := json.MarshalIndent(store, "", "  ")
	os.WriteFile(dataFile, data, 0o644)
}

// load – fájlból memóriába
func load() {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return // nincs fájl, üres store-ral indulunk
	}
	json.Unmarshal(data, &store)
}

func main() {
	// Induláskor visszatöltjük a fájlból
	load()
	fmt.Printf("\n%skvault%s → :8080  (%d kulcs betöltve)\n\n", olive, reset, len(store))

	http.HandleFunc("/keys/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/keys/")

		if key == "" {
			http.Error(w, "hiányzó kulcs", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			value, exists := store[key]
			if !exists {
				http.Error(w, "nem található", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"key": key, "value": value})

		case "POST":
			var body struct {
				Value string `json:"value"`
			}
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				http.Error(w, "hibás JSON", http.StatusBadRequest)
				return
			}
			store[key] = body.Value
			save()
			fmt.Printf("%s SET%s %s%s%s = %s\n", olive, reset, muted, key, reset, body.Value)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"key": key, "value": body.Value})

		case "DELETE":
			_, exists := store[key]
			if !exists {
				http.Error(w, "nem található", http.StatusNotFound)
				return
			}
			delete(store, key)
			save()
			fmt.Printf("%s DEL%s %s%s%s\n", red, reset, muted, key, reset)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"key": key, "deleted": "true"})

		default:
			http.Error(w, "nem támogatott method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store)
	})

	http.ListenAndServe(":8080", nil)
}
