package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

var (
	cOlive = "\033[38;2;168;200;48m"
	cRust  = "\033[38;2;255;107;74m"
	cMuted = "\033[38;2;88;80;72m"
	cReset = "\033[0m"
)

// config holds runtime settings
type config struct {
	port    string
	runtime string
}

// HealthResponse is the /health endpoint response
type HealthResponse struct {
	Status  string `json:"status"`
	Runtime string `json:"runtime"`
}

// InfoResponse is the /info endpoint response
type InfoResponse struct {
	System string `json:"system"`
	Uptime string `json:"uptime"`
	Memory string `json:"memory"`
	Disk   string `json:"disk"`
}

func main() {
	port := flag.String("port", "8080", "listen port")
	runtime := flag.String("runtime", "podman", "container runtime: docker or podman")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "stackctl - system info REST API\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: stackctl [flags]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Endpoints:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  GET /health   health check\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  GET /stacks   running containers (JSON)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  GET /info     system information\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *noColor {
		cOlive, cRust, cMuted, cReset = "", "", "", ""
	}

	cfg := config{port: *port, runtime: *runtime}

	http.HandleFunc("/health", healthHandler(cfg))
	http.HandleFunc("/stacks", stacksHandler(cfg))
	http.HandleFunc("/info", infoHandler())

	addr := ":" + cfg.port
	fmt.Printf("\n%sstackctl%s %s->%s %s (runtime: %s)\n\n", cOlive, cReset, cMuted, cReset, addr, cfg.runtime)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%serror:%s %v\n", cRust, cReset, err)
	}
}

func logRequest(method, path, msg string) {
	fmt.Printf("%s%-4s%s %s  %s\n", cOlive, method, cReset, path, msg)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func runCmd(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func healthHandler(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r.Method, "/health", "OK")
		writeJSON(w, HealthResponse{
			Status:  "ok",
			Runtime: cfg.runtime,
		})
	}
}

func stacksHandler(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command(cfg.runtime, "ps", "--format", "json")
		output, err := cmd.Output()
		if err != nil {
			logRequest(r.Method, "/stacks", cRust+"error"+cReset)
			http.Error(w, fmt.Sprintf(`{"error":"%s not available"}`, cfg.runtime), http.StatusInternalServerError)
			return
		}

		logRequest(r.Method, "/stacks", "OK")
		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	}
}

func infoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := InfoResponse{
			System: runCmd("uname", "-a"),
			Uptime: runCmd("uptime", "-p"),
			Memory: runCmd("free", "-h"),
			Disk:   runCmd("df", "-h", "/"),
		}

		logRequest(r.Method, "/info", "OK")
		writeJSON(w, info)
	}
}
