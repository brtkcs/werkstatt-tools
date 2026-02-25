package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	olive = "\033[38;2;168;200;48m"
	red   = "\033[38;2;255;107;74m"
	muted = "\033[38;2;88;80;72m"
	reset = "\033[0m"
)

type Config struct {
	Interval  int      `yaml:"interval"`
	Processes []string `yaml:"processes"`
}

func isRunning(name string) bool {
	cmd := exec.Command("pgrep", "-x", name)
	return cmd.Run() == nil
}

func main() {
	data, err := os.ReadFile("dmon.yaml")
	if err != nil {
		fmt.Println("Nem találom a dmon.yaml fájlt")
		return
	}

	var config Config
	yaml.Unmarshal(data, &config)

	interval := time.Duration(config.Interval) * time.Second
	fmt.Printf("\n%sdmon%s – %d process (%v)\n\n", olive, reset, len(config.Processes), interval)

	// Előző állapot – ebből tudjuk mi változott
	prev := make(map[string]bool)

	for {
		for _, p := range config.Processes {
			running := isRunning(p)

			// Első futásnál mindig kiírjuk
			old, exists := prev[p]
			if !exists {
				if running {
					fmt.Printf("  %s●%s %s%s%s  fut\n", olive, reset, muted, p, reset)
				} else {
					fmt.Printf("  %s●%s %s%s%s  nem fut\n", red, reset, muted, p, reset)
				}
			} else if old != running {
				// Állapot változott!
				ts := time.Now().Format("15:04:05")
				if running {
					fmt.Printf("  %s%s%s  %s▲%s %s  elindult\n", muted, ts, reset, olive, reset, p)
				} else {
					fmt.Printf("  %s%s%s  %s▼%s %s  leállt\n", muted, ts, reset, red, reset, p)
				}
			}

			prev[p] = running
		}

		time.Sleep(interval)
	}
}
