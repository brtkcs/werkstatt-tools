package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	cOlive, cPurple, cRust, cMuted, cReset = "", "", "", "", ""
}

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{
			name:  "simple list",
			input: "22,80,443",
			want:  []int{22, 80, 443},
		},
		{
			name:  "with spaces",
			input: "22, 80, 443",
			want:  []int{22, 80, 443},
		},
		{
			name:  "duplicates removed",
			input: "80,443,80,22",
			want:  []int{22, 80, 443},
		},
		{
			name:  "sorted output",
			input: "8080,22,443,80",
			want:  []int{22, 80, 443, 8080},
		},
		{
			name:  "single port",
			input: "22",
			want:  []int{22},
		},
		{
			name:    "invalid port string",
			input:   "22,abc,443",
			wantErr: true,
		},
		{
			name:    "port too high",
			input:   "22,99999",
			wantErr: true,
		},
		{
			name:    "port zero",
			input:   "0,22",
			wantErr: true,
		},
		{
			name:  "trailing comma",
			input: "22,80,",
			want:  []int{22, 80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePorts(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d ports, want %d", len(got), len(tt.want))
			}
			for i, p := range got {
				if p != tt.want[i] {
					t.Errorf("port[%d] = %d, want %d", i, p, tt.want[i])
				}
			}
		})
	}
}

func TestFormatPorts(t *testing.T) {
	result := formatPorts([]int{22, 80, 443})
	if result != "22, 80, 443" {
		t.Errorf("got %q, want %q", result, "22, 80, 443")
	}
}

func TestFormatPortsEmpty(t *testing.T) {
	result := formatPorts([]int{})
	if result != "" {
		t.Errorf("got %q, want empty", result)
	}
}

func TestScanHostLocalListener(t *testing.T) {
	// Start a temporary TCP listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Accept connections in background
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	// Extract port number
	addr := ln.Addr().String()
	portStr := addr[strings.LastIndex(addr, ":")+1:]
	port, _ := strconv.Atoi(portStr)

	host := scanHost("127.0.0.1", 500*time.Millisecond, []int{port})

	if host.IP != "127.0.0.1" {
		t.Errorf("got IP %q, want 127.0.0.1", host.IP)
	}
	if len(host.Ports) != 1 || host.Ports[0] != port {
		t.Errorf("got ports %v, want [%d]", host.Ports, port)
	}
}

func TestScanHostClosedPorts(t *testing.T) {
	// Scan ports that should not be open on localhost
	host := scanHost("127.0.0.1", 200*time.Millisecond, []int{39871, 39872})

	if len(host.Ports) != 0 {
		t.Errorf("expected no open ports, got %v", host.Ports)
	}
}

func TestIsAliveWithListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	addr := ln.Addr().String()
	portStr := addr[strings.LastIndex(addr, ":")+1:]
	port, _ := strconv.Atoi(portStr)

	alive := isAlive("127.0.0.1", 500*time.Millisecond, []int{port})
	if !alive {
		t.Error("expected host to be alive")
	}
}

func TestIsAliveNoListener(t *testing.T) {
	alive := isAlive("127.0.0.1", 200*time.Millisecond, []int{39871})
	if alive {
		t.Error("expected host to not be alive on closed port")
	}
}

func TestScanHostWithHTTPServer(t *testing.T) {
	// Use httptest as a real service to scan
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	addr := srv.Listener.Addr().String()
	portStr := addr[strings.LastIndex(addr, ":")+1:]
	port, _ := strconv.Atoi(portStr)

	host := scanHost("127.0.0.1", 500*time.Millisecond, []int{port, 39999})

	if len(host.Ports) != 1 {
		t.Fatalf("expected 1 open port, got %d: %v", len(host.Ports), host.Ports)
	}
	if host.Ports[0] != port {
		t.Errorf("got port %d, want %d", host.Ports[0], port)
	}
}
