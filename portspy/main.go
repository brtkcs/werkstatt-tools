package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

var (
	cOlive  = "\033[38;2;168;200;48m"
	cPurple = "\033[38;2;139;92;246m"
	cMuted  = "\033[38;2;88;80;72m"
	cReset  = "\033[0m"
)

// Result holds the outcome of a single port probe
type Result struct {
	Port     int
	Open     bool
	Duration time.Duration
}

func main() {
	host := flag.String("host", "localhost", "target host or IP")
	start := flag.Int("start", 1, "start port")
	end := flag.Int("end", 1024, "end port")
	timeout := flag.Int("timeout", 1, "timeout in seconds")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "portspy - concurrent TCP port scanner\n\n")
		fmt.Fprintf(os.Stderr, "Usage: portspy [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  portspy -host 10.0.0.1\n")
		fmt.Fprintf(os.Stderr, "  portspy -host 192.168.1.1 -start 1 -end 65535 -timeout 2\n")
	}
	flag.Parse()

	if *noColor {
		cOlive, cPurple, cMuted, cReset = "", "", "", ""
	}

	if *start < 1 || *end > 65535 || *start > *end {
		fmt.Fprintf(os.Stderr, "invalid port range: %d-%d\n", *start, *end)
		os.Exit(2)
	}

	dur := time.Duration(*timeout) * time.Second
	os.Exit(run(*host, *start, *end, dur))
}

// run executes the scan and returns an exit code
func run(host string, start, end int, timeout time.Duration) int {
	fmt.Printf("\n%sportspy%s %s->%s %s [%d-%d]\n\n", cOlive, cReset, cMuted, cReset, host, start, end)

	open := scanPorts(host, start, end, timeout)

	for _, r := range open {
		fmt.Printf("  %s%-6d%s %sopen%s  %s%v%s\n", cPurple, r.Port, cReset, cOlive, cReset, cMuted, r.Duration.Round(time.Millisecond), cReset)
	}

	fmt.Printf("\n  %s%d open ports%s\n\n", cOlive, len(open), cReset)
	return 0
}

// scanPorts probes all ports in range concurrently and returns sorted open results
func scanPorts(host string, start, end int, timeout time.Duration) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, end-start+1)

	for port := start; port <= end; port++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			results <- probePort(host, p, timeout)
		}(port)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var open []Result
	for r := range results {
		if r.Open {
			open = append(open, r)
		}
	}

	sort.Slice(open, func(i, j int) bool {
		return open[i].Port < open[j].Port
	})

	return open
}

// probePort checks if a single TCP port is open
func probePort(host string, port int, timeout time.Duration) Result {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, timeout)
	dur := time.Since(start)

	if err != nil {
		return Result{Port: port, Open: false, Duration: dur}
	}
	conn.Close()
	return Result{Port: port, Open: true, Duration: dur}
}
