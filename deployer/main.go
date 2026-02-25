package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	olive  = "\033[38;2;168;200;48m"
	red    = "\033[38;2;255;107;74m"
	purple = "\033[38;2;139;92;246m"
	muted  = "\033[38;2;88;80;72m"
	reset  = "\033[0m"
)

type Stack struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Config struct {
	Runtime string  `yaml:"runtime"`
	Stacks  []Stack `yaml:"stacks"`
}

// stackStatus – fut-e a stack?
func stackStatus(s Stack, runtime string) (string, bool) {
	cmd := exec.Command(runtime, "compose", "ps", "--format", "json")
	cmd.Dir = s.Path
	output, err := cmd.Output()
	if err != nil {
		return "hiba", false
	}
	// Ha üres vagy "[]" → nem fut
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" || trimmed == "[]" || trimmed == "null" {
		return "leállva", false
	}
	return "fut", true
}

func main() {
	data, err := os.ReadFile("deployer.yaml")
	if err != nil {
		fmt.Println("Nem találom a deployer.yaml fájlt")
		return
	}

	var config Config
	yaml.Unmarshal(data, &config)

	action := "status"
	target := ""
	if len(os.Args) > 1 {
		action = os.Args[1]
	}
	if len(os.Args) > 2 {
		target = os.Args[2]
	}

	fmt.Printf("\n%sdeployer%s – %s", olive, reset, action)
	if target != "" {
		fmt.Printf(" [%s]", target)
	}
	fmt.Println("\n")

	// Szűrés – ha van target, csak azt a stack-ot kezeljük
	stacks := config.Stacks
	if target != "" {
		stacks = nil
		for _, s := range config.Stacks {
			if s.Name == target {
				stacks = append(stacks, s)
			}
		}
		if len(stacks) == 0 {
			fmt.Printf("  %s✗%s nem találom: %s\n\n", red, reset, target)
			return
		}
	}

	switch action {
	case "status":
		for _, s := range stacks {
			status, running := stackStatus(s, config.Runtime)
			if running {
				fmt.Printf("  %s●%s %-20s %s%s%s\n", olive, reset, s.Name, muted, status, reset)
			} else {
				fmt.Printf("  %s●%s %-20s %s%s%s\n", red, reset, s.Name, muted, status, reset)
			}
		}

	case "up":
		for _, s := range stacks {
			fmt.Printf("  %s▲%s %s  indítás...\n", olive, reset, s.Name)
			cmd := exec.Command(config.Runtime, "compose", "up", "-d")
			cmd.Dir = s.Path
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

	case "down":
		for _, s := range stacks {
			fmt.Printf("  %s▼%s %s  leállítás...\n", red, reset, s.Name)
			cmd := exec.Command(config.Runtime, "compose", "down")
			cmd.Dir = s.Path
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

	default:
		fmt.Printf("  Használat: deployer [status|up|down] [stack név]\n")
	}

	fmt.Println()
}
