package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	cOlive = "\033[38;2;168;200;48m"
	cRust  = "\033[38;2;255;107;74m"
	cMuted = "\033[38;2;88;80;72m"
	cReset = "\033[0m"
)

// Config holds the monitoring configuration
type Config struct {
	Interval  int      `yaml:"interval"`
	Processes []string `yaml:"processes"`
}

// StateChange represents a process state transition
type StateChange struct {
	Name    string
	Running bool
	Changed bool
	First   bool
}

// checker abstracts process checking for testing
type checker interface {
	isRunning(name string) bool
}

// pgrepChecker uses pgrep to check processes
type pgrepChecker struct{}

func (p pgrepChecker) isRunning(name string) bool {
	return exec.Command("pgrep", "-x", name).Run() == nil
}

func main() {
	configFile := flag.String("c", "dmon.yaml", "path to config file")
	once := flag.Bool("once", false, "run once and exit (for scripting)")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dmon - process monitor with state change detection\n\n")
		fmt.Fprintf(os.Stderr, "Usage: dmon [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Monitors processes listed in a YAML config and reports\n")
		fmt.Fprintf(os.Stderr, "when they start or stop.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfig format (dmon.yaml):\n")
		fmt.Fprintf(os.Stderr, "  interval: 5\n")
		fmt.Fprintf(os.Stderr, "  processes:\n")
		fmt.Fprintf(os.Stderr, "    - sshd\n")
		fmt.Fprintf(os.Stderr, "    - podman\n")
	}
	flag.Parse()

	if *noColor {
		cOlive, cRust, cMuted, cReset = "", "", "", ""
	}

	cfg, err := loadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", cRust, cReset, err)
		os.Exit(2)
	}

	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", cRust, cReset, err)
		os.Exit(2)
	}

	interval := time.Duration(cfg.Interval) * time.Second
	fmt.Printf("\n%sdmon%s %s->%s %d processes (%v)\n\n", cOlive, cReset, cMuted, cReset, len(cfg.Processes), interval)

	chk := pgrepChecker{}
	prev := make(map[string]bool)

	if *once {
		changes := checkAll(cfg.Processes, prev, chk)
		printChanges(changes)
		return
	}

	for {
		changes := checkAll(cfg.Processes, prev, chk)
		printChanges(changes)
		for _, c := range changes {
			prev[c.Name] = c.Running
		}
		time.Sleep(interval)
	}
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

	if cfg.Interval <= 0 {
		cfg.Interval = 5
	}

	return cfg, nil
}

// validateConfig checks required fields
func validateConfig(cfg Config) error {
	if len(cfg.Processes) == 0 {
		return fmt.Errorf("no processes defined in config")
	}
	return nil
}

// checkAll checks all processes and returns state changes
func checkAll(processes []string, prev map[string]bool, chk checker) []StateChange {
	var changes []StateChange

	for _, name := range processes {
		running := chk.isRunning(name)
		old, exists := prev[name]

		change := StateChange{
			Name:    name,
			Running: running,
			First:   !exists,
			Changed: exists && old != running,
		}
		changes = append(changes, change)
	}

	return changes
}

// printChanges outputs state changes
func printChanges(changes []StateChange) {
	for _, c := range changes {
		if c.First {
			if c.Running {
				fmt.Printf("  %s*%s %s%s%s  running\n", cOlive, cReset, cMuted, c.Name, cReset)
			} else {
				fmt.Printf("  %s*%s %s%s%s  not running\n", cRust, cReset, cMuted, c.Name, cReset)
			}
		} else if c.Changed {
			ts := time.Now().Format("15:04:05")
			if c.Running {
				fmt.Printf("  %s%s%s  %s^%s %s  started\n", cMuted, ts, cReset, cOlive, cReset, c.Name)
			} else {
				fmt.Printf("  %s%s%s  %sv%s %s  stopped\n", cMuted, ts, cReset, cRust, cReset, c.Name)
			}
		}
	}
}
