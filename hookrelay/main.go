package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	cOlive = "\033[38;2;168;200;48m"
	cRust  = "\033[38;2;255;107;74m"
	cMuted = "\033[38;2;88;80;72m"
	cReset = "\033[0m"
)

// Webhook represents an incoming webhook payload
type Webhook struct {
	Event  string `json:"event"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

// LogEntry is what gets written to the log file
type LogEntry struct {
	Time   string `json:"time"`
	Event  string `json:"event"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

// logger abstracts log file writing for testing
type logger interface {
	save(hook Webhook) error
}

// fileLogger writes to a JSON lines file
type fileLogger struct {
	path string
}

func (l fileLogger) save(hook Webhook) error {
	entry := LogEntry{
		Time:   time.Now().Format("2006-01-02 15:04:05"),
		Event:  hook.Event,
		Repo:   hook.Repo,
		Branch: hook.Branch,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(append(data, '\n'))
	return err
}

func main() {
	port := flag.String("port", "8080", "listen port")
	logFile := flag.String("log", "webhooks.log", "path to log file")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "hookrelay - webhook receiver with JSON log\n\n")
		fmt.Fprintf(os.Stderr, "Usage: hookrelay [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Endpoints:\n")
		fmt.Fprintf(os.Stderr, "  GET  /health    health check\n")
		fmt.Fprintf(os.Stderr, "  POST /webhook   receive webhook payload\n\n")
		fmt.Fprintf(os.Stderr, "Payload format:\n")
		fmt.Fprintf(os.Stderr, `  {"event":"push","repo":"myrepo","branch":"main"}`)
		fmt.Fprintf(os.Stderr, "\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *noColor {
		cOlive, cRust, cMuted, cReset = "", "", "", ""
	}

	log := fileLogger{path: *logFile}

	http.HandleFunc("/health", healthHandler())
	http.HandleFunc("/webhook", webhookHandler(log))

	addr := ":" + *port
	fmt.Printf("\n%shookrelay%s %s->%s %s (log: %s)\n\n", cOlive, cReset, cMuted, cReset, addr, *logFile)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", cRust, cReset, err)
	}
}

func logRequest(method, path, msg string) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("%s%s%s %s%-4s%s %s  %s\n", cMuted, ts, cReset, cOlive, method, cReset, path, msg)
}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r.Method, "/health", "OK")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	}
}

func webhookHandler(log logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logRequest(r.Method, "/webhook", cRust+"POST only"+cReset)
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		var hook Webhook
		if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
			logRequest("POST", "/webhook", cRust+"invalid JSON"+cReset)
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if hook.Event == "" {
			logRequest("POST", "/webhook", cRust+"missing event field"+cReset)
			http.Error(w, "missing event field", http.StatusBadRequest)
			return
		}

		logRequest("POST", "/webhook", fmt.Sprintf("%s -> %s/%s", hook.Event, hook.Repo, hook.Branch))

		if err := log.save(hook); err != nil {
			logRequest("POST", "/webhook", cRust+"log write failed"+cReset)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"received"}`)
	}
}
