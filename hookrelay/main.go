package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	olive = "\033[38;2;168;200;48m"
	muted = "\033[38;2;88;80;72m"
	red   = "\033[38;2;255;107;74m"
	reset = "\033[0m"
)

type Webhook struct {
	Event  string `json:"event"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

// LogEntry – amit fájlba mentünk
type LogEntry struct {
	Time   string `json:"time"`
	Event  string `json:"event"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

func log(method, path, msg string) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("%s%s%s %s%-4s%s %s  %s\n", muted, ts, reset, olive, method, reset, path, msg)
}

// saveLog – hozzáfűz egy sort a webhooks.log fájlhoz
func saveLog(hook Webhook) {
	entry := LogEntry{
		Time:   time.Now().Format("2006-01-02 15:04:05"),
		Event:  hook.Event,
		Repo:   hook.Repo,
		Branch: hook.Branch,
	}

	// JSON-né alakítjuk az entry-t
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// os.OpenFile – append módban nyit (hozzáfűz, nem felülír)
	// O_APPEND: a végére ír
	// O_CREATE: ha nem létezik, létrehozza
	// O_WRONLY: csak írásra
	f, err := os.OpenFile("webhooks.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	// Egy sor = egy JSON + újsor karakter
	f.Write(append(data, '\n'))
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log(r.Method, "/health", "OK")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			log(r.Method, "/webhook", red+"csak POST"+reset)
			http.Error(w, "csak POST", http.StatusMethodNotAllowed)
			return
		}

		var hook Webhook
		err := json.NewDecoder(r.Body).Decode(&hook)
		if err != nil {
			log("POST", "/webhook", red+"hibás JSON"+reset)
			http.Error(w, "hibás JSON", http.StatusBadRequest)
			return
		}

		log("POST", "/webhook", fmt.Sprintf("%s → %s/%s", hook.Event, hook.Repo, hook.Branch))
		saveLog(hook)
		fmt.Fprintln(w, "OK")
	})

	fmt.Printf("\n%shookrelay%s → :8080\n\n", olive, reset)
	http.ListenAndServe(":8080", nil)
}
