package main

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	olive  = "\033[38;2;168;200;48m"
	purple = "\033[38;2;139;92;246m"
	muted  = "\033[38;2;88;80;72m"
	reset  = "\033[0m"
)

type Host struct {
	IP    string
	Ports []int
}

// Gyakori portok – nem az összeset, csak az érdekeseket
var commonPorts = []int{22, 53, 80, 443, 631, 3000, 3306, 5432, 8080, 8443, 9090}

func scanHost(ip string, timeout time.Duration) Host {
	host := Host{IP: ip}

	var wg sync.WaitGroup
	var mu sync.Mutex
	// Mutex – "zár". Egyszerre csak egy goroutine írhat a ports slice-ba.
	// Nélküle a párhuzamos írás összekeveredne.

	for _, port := range commonPorts {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%d", ip, p)
			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err != nil {
				return
			}
			conn.Close()

			mu.Lock()
			host.Ports = append(host.Ports, p)
			mu.Unlock()
		}(port)
	}

	wg.Wait()
	sort.Ints(host.Ports)
	return host
}

func isAlive(ip string, timeout time.Duration) bool {
	for _, port := range commonPorts {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

func main() {
	subnet := "10.0.0"
	timeout := 500 * time.Millisecond

	fmt.Printf("\n%snetmapper%s – %s.0/24\n\n", olive, reset, subnet)
	fmt.Printf("%shost discovery...%s\n", muted, reset)

	// 1. fázis – ki él?
	var wg sync.WaitGroup
	aliveIPs := make(chan string, 254)

	for i := 1; i <= 254; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ip := fmt.Sprintf("%s.%d", subnet, n)
			if isAlive(ip, timeout) {
				aliveIPs <- ip
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(aliveIPs)
	}()

	var alive []string
	for ip := range aliveIPs {
		alive = append(alive, ip)
	}
	sort.Strings(alive)

	// 2. fázis – mit futtatnak?
	fmt.Printf("%s%d host, port scan...%s\n\n", muted, len(alive), reset)

	for _, ip := range alive {
		host := scanHost(ip, timeout)
		fmt.Printf("  %s●%s %-15s", olive, reset, host.IP)
		for _, p := range host.Ports {
			fmt.Printf("  %s%d%s", purple, p, reset)
		}
		fmt.Println()
	}

	fmt.Printf("\n  %s%d eszköz%s\n\n", olive, len(alive), reset)
}
