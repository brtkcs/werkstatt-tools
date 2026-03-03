package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	cOlive = "\033[38;2;168;200;48m"
	cRust  = "\033[38;2;255;107;74m"
	cMuted = "\033[38;2;88;80;72m"
	cReset = "\033[0m"
)

// Host represents a target from the config file
type Host struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// Config holds the hosts list
type Config struct {
	Hosts []Host `yaml:"hosts"`
}

// Result holds the outcome of pinging a single host
type Result struct {
	Host     Host
	Up       bool
	Duration time.Duration
}

// dialer abstracts TCP connection for testing
type dialer interface {
	dial(address string, timeout time.Duration) (bool, time.Duration)
}

// tcpDialer is the real implementation
type tcpDialer struct{}

func (d tcpDialer) dial(address string, timeout time.Duration) (bool, time.Duration) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	dur := time.Since(start)
	if err != nil {
		return false, dur
	}
	conn.Close()
	return true, dur
}

func main() {
	configFile := flag.String("c", "hosts.yaml", "path to hosts config file")
	timeout := flag.Int("timeout", 2, "timeout in seconds")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "sshping - check host reachability from a YAML config\n\n")
		fmt.Fprintf(os.Stderr, "Usage: sshping [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfig format (hosts.yaml):\n")
		fmt.Fprintf(os.Stderr, "  hosts:\n")
		fmt.Fprintf(os.Stderr, "    - name: my server\n")
		fmt.Fprintf(os.Stderr, "      address: 10.0.0.1\n")
		fmt.Fprintf(os.Stderr, "      port: 22\n")
	}
	flag.Parse()

	if *noColor {
		cOlive, cRust, cMuted, cReset = "", "", "", ""
	}

	dur := time.Duration(*timeout) * time.Second
	os.Exit(run(*configFile, dur, tcpDialer{}))
}

// loadConfig reads and parses the YAML config
func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("config not found: %s", path)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %v", err)
	}

	return cfg, nil
}

// validateConfig checks that hosts have required fields
func validateConfig(cfg Config) error {
	if len(cfg.Hosts) == 0 {
		return fmt.Errorf("no hosts defined in config")
	}
	for i, h := range cfg.Hosts {
		if h.Name == "" {
			return fmt.Errorf("host %d has no name", i+1)
		}
		if h.Address == "" {
			return fmt.Errorf("host %q has no address", h.Name)
		}
		if h.Port < 1 || h.Port > 65535 {
			return fmt.Errorf("host %q has invalid port: %d", h.Name, h.Port)
		}
	}
	return nil
}

// pingHosts checks all hosts concurrently and returns results
func pingHosts(hosts []Host, timeout time.Duration, d dialer) []Result {
	results := make(chan Result, len(hosts))
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(h Host) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%d", h.Address, h.Port)
			up, dur := d.dial(addr, timeout)
			results <- Result{Host: h, Up: up, Duration: dur}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var out []Result
	for r := range results {
		out = append(out, r)
	}
	return out
}

// run executes the ping check and returns an exit code
func run(configFile string, timeout time.Duration, d dialer) int {
	cfg, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", cRust, cReset, err)
		return 2
	}

	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", cRust, cReset, err)
		return 2
	}

	fmt.Printf("\n%ssshping%s %s->%s %d hosts\n\n", cOlive, cReset, cMuted, cReset, len(cfg.Hosts))

	results := pingHosts(cfg.Hosts, timeout, d)

	up := 0
	down := 0
	for _, r := range results {
		addr := fmt.Sprintf("%s:%d", r.Host.Address, r.Host.Port)
		if r.Up {
			fmt.Printf("  %s*%s %-20s %s%s%s  %v\n", cOlive, cReset, r.Host.Name, cMuted, addr, cReset, r.Duration.Round(time.Millisecond))
			up++
		} else {
			fmt.Printf("  %s✗%s %-20s %s%s%s  unreachable\n", cRust, cReset, r.Host.Name, cMuted, addr, cReset)
			down++
		}
	}

	fmt.Printf("\n  %s%d/%d reachable%s\n\n", cOlive, up, len(cfg.Hosts), cReset)

	if down > 0 {
		return 1
	}
	return 0
}
