package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	cOlive  = "\033[38;2;168;200;48m"
	cRust   = "\033[38;2;255;107;74m"
	cPurple = "\033[38;2;139;92;246m"
	cMuted  = "\033[38;2;88;80;72m"
	cReset  = "\033[0m"
)

// Stack represents a compose stack definition
type Stack struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// Config holds the deployer configuration
type Config struct {
	Runtime string  `yaml:"runtime"`
	Stacks  []Stack `yaml:"stacks"`
}

// runner abstracts command execution for testing
type runner interface {
	run(runtime, dir string, args ...string) (string, error)
	runAttached(runtime, dir string, args ...string) error
}

// execRunner is the real command executor
type execRunner struct{}

func (e execRunner) run(runtime, dir string, args ...string) (string, error) {
	cmd := exec.Command(runtime, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func (e execRunner) runAttached(runtime, dir string, args ...string) error {
	cmd := exec.Command(runtime, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	configFile := flag.String("c", "deployer.yaml", "path to config file")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "deployer - manage compose stacks from a single config\n\n")
		fmt.Fprintf(os.Stderr, "Usage: deployer [flags] <command> [stack]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  status    show stack status (default)\n")
		fmt.Fprintf(os.Stderr, "  up        start stacks\n")
		fmt.Fprintf(os.Stderr, "  down      stop stacks\n")
		fmt.Fprintf(os.Stderr, "  restart   restart stacks (down + up)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  deployer status\n")
		fmt.Fprintf(os.Stderr, "  deployer up vaultwarden\n")
		fmt.Fprintf(os.Stderr, "  deployer restart\n")
		fmt.Fprintf(os.Stderr, "  deployer -c /etc/deployer.yaml status\n")
	}
	flag.Parse()

	if *noColor {
		cOlive, cRust, cPurple, cMuted, cReset = "", "", "", "", ""
	}

	action := "status"
	target := ""
	args := flag.Args()
	if len(args) > 0 {
		action = args[0]
	}
	if len(args) > 1 {
		target = args[1]
	}

	os.Exit(run(*configFile, action, target, execRunner{}))
}

// loadConfig reads and parses the YAML config file
func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("config not found: %s", path)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %v", err)
	}

	if cfg.Runtime == "" {
		cfg.Runtime = "podman"
	}

	return cfg, nil
}

// validateConfig checks that the config has required fields
func validateConfig(cfg Config) error {
	if len(cfg.Stacks) == 0 {
		return fmt.Errorf("no stacks defined in config")
	}
	for i, s := range cfg.Stacks {
		if s.Name == "" {
			return fmt.Errorf("stack %d has no name", i+1)
		}
		if s.Path == "" {
			return fmt.Errorf("stack %q has no path", s.Name)
		}
	}
	return nil
}

// filterStacks returns matching stacks, or error if target not found
func filterStacks(stacks []Stack, target string) ([]Stack, error) {
	if target == "" {
		return stacks, nil
	}

	for _, s := range stacks {
		if s.Name == target {
			return []Stack{s}, nil
		}
	}
	return nil, fmt.Errorf("stack not found: %s", target)
}

// stackStatus checks if a compose stack is running
func stackStatus(s Stack, runtime string, r runner) (string, bool) {
	output, err := r.run(runtime, s.Path, "compose", "ps", "--format", "json")
	if err != nil {
		return "error", false
	}
	if output == "" || output == "[]" || output == "null" {
		return "stopped", false
	}
	return "running", true
}

// run executes the deployer action and returns an exit code
func run(configFile, action, target string, r runner) int {
	cfg, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", cRust, cReset, err)
		return 2
	}

	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", cRust, cReset, err)
		return 2
	}

	stacks, err := filterStacks(cfg.Stacks, target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", cRust, cReset, err)
		return 1
	}

	fmt.Printf("\n%sdeployer%s %s->%s %s", cOlive, cReset, cMuted, cReset, action)
	if target != "" {
		fmt.Printf(" [%s]", target)
	}
	fmt.Println()

	switch action {
	case "status":
		for _, s := range stacks {
			status, running := stackStatus(s, cfg.Runtime, r)
			icon, color := "*", cRust
			if running {
				color = cOlive
			}
			fmt.Printf("  %s%s%s %-20s %s%s%s\n", color, icon, cReset, s.Name, cMuted, status, cReset)
		}

	case "up":
		for _, s := range stacks {
			fmt.Printf("  %s^%s %s  starting...\n", cOlive, cReset, s.Name)
			if err := r.runAttached(cfg.Runtime, s.Path, "compose", "up", "-d"); err != nil {
				fmt.Fprintf(os.Stderr, "  %s✗%s %s failed: %v\n", cRust, cReset, s.Name, err)
			}
		}

	case "down":
		for _, s := range stacks {
			fmt.Printf("  %sv%s %s  stopping...\n", cRust, cReset, s.Name)
			if err := r.runAttached(cfg.Runtime, s.Path, "compose", "down"); err != nil {
				fmt.Fprintf(os.Stderr, "  %s✗%s %s failed: %v\n", cRust, cReset, s.Name, err)
			}
		}

	case "restart":
		for _, s := range stacks {
			fmt.Printf("  %sv%s %s  stopping...\n", cRust, cReset, s.Name)
			r.runAttached(cfg.Runtime, s.Path, "compose", "down")
			fmt.Printf("  %s^%s %s  starting...\n", cOlive, cReset, s.Name)
			if err := r.runAttached(cfg.Runtime, s.Path, "compose", "up", "-d"); err != nil {
				fmt.Fprintf(os.Stderr, "  %s✗%s %s failed: %v\n", cRust, cReset, s.Name, err)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s (use status, up, down, restart)\n", action)
		return 2
	}

	fmt.Println()
	return 0
}
