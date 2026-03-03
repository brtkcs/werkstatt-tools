package main

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	cOlive, cPurple, cMuted, cReset = "", "", "", ""
}

// startListener creates a TCP listener on a random port, returns port number and cleanup func
func startListener(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

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
	return port, func() { ln.Close() }
}

func TestProbePortOpen(t *testing.T) {
	port, cleanup := startListener(t)
	defer cleanup()

	result := probePort("127.0.0.1", port, 500*time.Millisecond)

	if !result.Open {
		t.Errorf("expected port %d to be open", port)
	}
	if result.Port != port {
		t.Errorf("got port %d, want %d", result.Port, port)
	}
	if result.Duration == 0 {
		t.Error("expected non-zero duration")
	}
}

func TestProbePortClosed(t *testing.T) {
	result := probePort("127.0.0.1", 39871, 200*time.Millisecond)

	if result.Open {
		t.Error("expected port 39871 to be closed")
	}
}

func TestScanPortsFindsOpen(t *testing.T) {
	port, cleanup := startListener(t)
	defer cleanup()

	// Scan a small range that includes our listener
	open := scanPorts("127.0.0.1", port, port, 500*time.Millisecond)

	if len(open) != 1 {
		t.Fatalf("expected 1 open port, got %d", len(open))
	}
	if open[0].Port != port {
		t.Errorf("got port %d, want %d", open[0].Port, port)
	}
}

func TestScanPortsNoneOpen(t *testing.T) {
	// Scan a range where nothing should be listening
	open := scanPorts("127.0.0.1", 39870, 39875, 200*time.Millisecond)

	if len(open) != 0 {
		t.Errorf("expected 0 open ports, got %d: %v", len(open), open)
	}
}

func TestScanPortsSorted(t *testing.T) {
	port1, cleanup1 := startListener(t)
	defer cleanup1()
	port2, cleanup2 := startListener(t)
	defer cleanup2()

	low, high := port1, port2
	if low > high {
		low, high = high, low
	}

	open := scanPorts("127.0.0.1", low, high, 500*time.Millisecond)

	if len(open) < 2 {
		t.Skipf("expected at least 2 open ports, got %d (ports may overlap)", len(open))
	}

	for i := 1; i < len(open); i++ {
		if open[i].Port < open[i-1].Port {
			t.Errorf("results not sorted: port %d before %d", open[i-1].Port, open[i].Port)
		}
	}
}

func TestScanPortsMultipleListeners(t *testing.T) {
	port1, cleanup1 := startListener(t)
	defer cleanup1()
	port2, cleanup2 := startListener(t)
	defer cleanup2()
	port3, cleanup3 := startListener(t)
	defer cleanup3()

	// Find the range
	ports := []int{port1, port2, port3}
	min, max := ports[0], ports[0]
	for _, p := range ports {
		if p < min {
			min = p
		}
		if p > max {
			max = p
		}
	}

	open := scanPorts("127.0.0.1", min, max, 500*time.Millisecond)

	if len(open) < 3 {
		t.Errorf("expected at least 3 open ports, got %d", len(open))
	}
}
