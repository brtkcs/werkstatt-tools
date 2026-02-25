package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

const (
	olive = "\033[38;2;168;200;48m"
	red   = "\033[38;2;255;107;74m"
	reset = "\033[0m"
)

func log(method, path, msg string) {
	fmt.Printf("%s%-4s%s %s  %s\n", olive, method, reset, path, msg)
}

func trimNL(b []byte) string {
	return strings.TrimSpace(string(b))
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log(r.Method, "/health", "OK")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	http.HandleFunc("/stacks", func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("podman", "ps", "--format", "json")
		output, err := cmd.Output()
		if err != nil {
			log(r.Method, "/stacks", red+"hiba"+reset)
			http.Error(w, "podman hiba", http.StatusInternalServerError)
			return
		}

		log(r.Method, "/stacks", "OK")
		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	})

	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		uname, _ := exec.Command("uname", "-a").Output()
		mem, _ := exec.Command("free", "-h").Output()
		disk, _ := exec.Command("df", "-h", "/").Output()
		uptime, _ := exec.Command("uptime", "-p").Output()

		info := fmt.Sprintf("{\"system\": %q, \"uptime\": %q, \"memory\": %q, \"disk\": %q}",
			trimNL(uname),
			trimNL(uptime),
			trimNL(mem),
			trimNL(disk),
		)

		log(r.Method, "/info", "OK")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, info)
	})

	fmt.Printf("\n%sstackctl%s → :8080\n\n", olive, reset)
	http.ListenAndServe(":8080", nil)
}
