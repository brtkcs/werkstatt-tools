package main

import (
	"flag"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// Result – egy port scan eredménye
type Result struct {
	Port     int
	Open     bool
	Duration time.Duration
}

func main() {
	// flag – CLI argumentumok kezelése
	// Használat: go run main.go -host 10.0.0.19 -start 1 -end 1024
	host := flag.String("host", "localhost", "cél host vagy IP")
	start := flag.Int("start", 1, "kezdő port")
	end := flag.Int("end", 1024, "záró port")
	timeout := flag.Int("timeout", 1, "timeout másodpercben")
	flag.Parse()

	fmt.Printf("\n portspy – %s [%d-%d]\n\n", *host, *start, *end)

	var wg sync.WaitGroup
	// Channel – goroutine-ok ezen küldik vissza az eredményt
	// Mint egy cső: az egyik végén betolod, a másikon kijön
	results := make(chan Result, *end-*start+1)

	for port := *start; port <= *end; port++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			address := fmt.Sprintf("%s:%d", *host, p)
			startTime := time.Now()

			conn, err := net.DialTimeout("tcp", address, time.Duration(*timeout)*time.Second)
			duration := time.Since(startTime)

			if err != nil {
				results <- Result{Port: p, Open: false, Duration: duration}
				return
			}
			conn.Close()
			results <- Result{Port: p, Open: true, Duration: duration}
		}(port)
	}

	// Háttérben várjuk meg a goroutine-okat, aztán zárjuk a channel-t
	go func() {
		wg.Wait()
		close(results)
	}()

	// Összegyűjtjük és rendezzük az eredményeket
	var open []Result
	for r := range results {
		if r.Open {
			open = append(open, r)
		}
	}

	sort.Slice(open, func(i, j int) bool {
		return open[i].Port < open[j].Port
	})

	for _, r := range open {
		fmt.Printf("  %-6d nyitott  %v\n", r.Port, r.Duration.Round(time.Millisecond))
	}

	fmt.Printf("\n  %d nyitott port\n\n", len(open))
}
