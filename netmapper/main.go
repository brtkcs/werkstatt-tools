package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	cOlive  = "\033[38;2;168;200;48m"
	cPurple = "\033[38;2;139;92;246m"
	cRust   = "\033[38;2;255;107;74m"
	cMuted  = "\033[38;2;88;80;72m"
	cReset  = "\033[0m"
)

// Host represents a discovered network host and its open ports
type Host struct {
	IP    string
	Ports []int
}

// defaultPorts are commonly used service ports
var defaultPorts = []int{22, 53, 80, 443, 631, 3000, 3306, 5432, 8080, 8443, 9090}

func main() {
	subnet := flag.String("subnet", "192.168.1", "subnet prefix (first 3 octets)")
	timeout := flag.Int("timeout", 500, "connection timeout in milliseconds")
	ports := flag.String("ports", "", "comma-separated port list (default: common ports)")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "netmapper - network host discovery and port scanner\n\n")
		fmt.Fprintf(os.Stderr, "Usage: netmapper [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Scans a /24 subnet for live hosts, then checks open ports on each.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDefault ports: %s\n", formatPorts(defaultPorts))
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  netmapper -subnet 10.0.0\n")
		fmt.Fprintf(os.Stderr, "  netmapper -subnet 192.168.1 -timeout 1000\n")
		fmt.Fprintf(os.Stderr, "  netmapper -ports 22,80,443,8080\n")
	}
	flag.Parse()

	if *noColor {
		cOlive, cPurple, cRust, cMuted, cReset = "", "", "", "", ""
	}

	scanPorts := defaultPorts
	if *ports != "" {
		parsed, err := parsePorts(*ports)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%serror:%s %v\n", cRust, cReset, err)
			os.Exit(2)
		}
		scanPorts = parsed
	}

	dur := time.Duration(*timeout) * time.Millisecond
	os.Exit(run(*subnet, dur, scanPorts))
}

// run executes the scan and returns an exit code
func run(subnet string, timeout time.Duration, ports []int) int {
	fmt.Printf("\n%snetmapper%s %s->%s %s.0/24\n\n", cOlive, cReset, cMuted, cReset, subnet)
	fmt.Printf("%sdiscovering hosts...%s\n", cMuted, cReset)

	alive := discoverHosts(subnet, timeout, ports)

	fmt.Printf("%s%d hosts found, scanning ports...%s\n\n", cMuted, len(alive), cReset)

	if len(alive) == 0 {
		fmt.Printf("  %sno hosts found%s\n\n", cMuted, cReset)
		return 0
	}

	for _, ip := range alive {
		host := scanHost(ip, timeout, ports)
		fmt.Printf("  %s*%s %-15s", cOlive, cReset, host.IP)
		for _, p := range host.Ports {
			fmt.Printf("  %s%d%s", cPurple, p, cReset)
		}
		fmt.Println()
	}

	fmt.Printf("\n  %s%d hosts%s\n\n", cOlive, len(alive), cReset)
	return 0
}

// discoverHosts scans a /24 subnet and returns sorted list of live IPs
func discoverHosts(subnet string, timeout time.Duration, ports []int) []string {
	var wg sync.WaitGroup
	aliveIPs := make(chan string, 254)

	for i := 1; i <= 254; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ip := fmt.Sprintf("%s.%d", subnet, n)
			if isAlive(ip, timeout, ports) {
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
	return alive
}

// isAlive checks if a host responds on any of the given ports
func isAlive(ip string, timeout time.Duration, ports []int) bool {
	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

// scanHost scans all given ports on a single host concurrently
func scanHost(ip string, timeout time.Duration, ports []int) Host {
	host := Host{IP: ip}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, p), timeout)
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

// parsePorts converts a comma-separated string to a sorted int slice
func parsePorts(s string) ([]int, error) {
	var ports []int
	seen := make(map[int]bool)

	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		p, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", trimmed)
		}
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("port out of range: %d", p)
		}
		if !seen[p] {
			ports = append(ports, p)
			seen[p] = true
		}
	}

	sort.Ints(ports)
	return ports, nil
}

// formatPorts returns a human-readable port list
func formatPorts(ports []int) string {
	s := make([]string, len(ports))
	for i, p := range ports {
		s[i] = strconv.Itoa(p)
	}
	return strings.Join(s, ", ")
}
