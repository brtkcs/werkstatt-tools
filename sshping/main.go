package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	red   = "\033[38;2;255;107;74m"
	olive = "\033[38;2;168;200;48m"
	muted = "\033[38;2;88;80;72m"
	reset = "\033[0m"
)

type Host struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type Config struct {
	Hosts []Host `yaml:"hosts"`
}

type Result struct {
	Host     Host
	Up       bool
	Duration time.Duration
}

func main() {
	data, err := os.ReadFile("hosts.yaml")
	if err != nil {
		fmt.Println("Nem találom a hosts.yaml fájlt")
		return
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("YAML hiba:", err)
		return
	}

	fmt.Printf("\n%s sshping%s – %d host\n\n", olive, reset, len(config.Hosts))

	timeout := 2 * time.Second
	results := make(chan Result, len(config.Hosts))
	var wg sync.WaitGroup

	for _, host := range config.Hosts {
		wg.Add(1)
		go func(h Host) {
			defer wg.Done()

			address := fmt.Sprintf("%s:%d", h.Address, h.Port)
			start := time.Now()

			conn, err := net.DialTimeout("tcp", address, timeout)
			duration := time.Since(start)

			if err != nil {
				results <- Result{Host: h, Up: false, Duration: duration}
				return
			}
			conn.Close()
			results <- Result{Host: h, Up: true, Duration: duration}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	up := 0
	for r := range results {
		address := fmt.Sprintf("%s:%d", r.Host.Address, r.Host.Port)
		if r.Up {
			fmt.Printf("  %s✓%s %-20s %s%s%s  %v\n", olive, reset, r.Host.Name, muted, address, reset, r.Duration.Round(time.Millisecond))
			up++
		} else {
			fmt.Printf("  %s✗%s %-20s %s%s%s  nem elérhető\n", red, reset, r.Host.Name, muted, address, reset)
		}
	}

	fmt.Printf("\n  %s%d/%d elérhető%s\n\n", olive, up, len(config.Hosts), reset)
}
