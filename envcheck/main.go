package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Werkstatt dark palette
var (
	cReset  = "\033[0m"
	cOlive  = "\033[38;2;168;200;48m"
	cRust   = "\033[38;2;212;101;74m"
	cPurple = "\033[38;2;123;108;176m"
	cMuted  = "\033[38;2;88;80;72m"
	cBold   = "\033[1m"
)

// Entry represents a parsed key-value pair from an env file
type Entry struct {
	Key   string
	Value string
	Line  int
}

// Problem represents a validation issue
type Problem struct {
	Level   string // "error" or "warn"
	Line    int
	Message string
}

func main() {
	envFile := flag.String("f", ".env", "path to .env file")
	exampleFile := flag.String("e", ".env.example", "path to .env.example file")
	quiet := flag.Bool("q", false, "quiet mode: only show errors")
	noColor := flag.Bool("no-color", false, "disable colored output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "envcheck - validate .env files against .env.example\n\n")
		fmt.Fprintf(os.Stderr, "Usage: envcheck [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExit codes:\n")
		fmt.Fprintf(os.Stderr, "  0  no errors\n")
		fmt.Fprintf(os.Stderr, "  1  validation errors found\n")
		fmt.Fprintf(os.Stderr, "  2  file not found\n")
	}
	flag.Parse()

	if *noColor {
		cReset, cOlive, cRust, cPurple, cMuted, cBold = "", "", "", "", "", ""
	}

	os.Exit(run(*envFile, *exampleFile, *quiet))
}

// run executes validation and returns an exit code
func run(envFile, exampleFile string, quiet bool) int {
	if !quiet {
		fmt.Printf("\n%s%s envcheck%s\n", cBold, cOlive, cReset)
		fmt.Printf("%s──────────────────────────────────%s\n", cMuted, cReset)
	}

	// Parse .env
	entries, problems := parseEnvFile(envFile)
	if entries == nil {
		fmt.Fprintf(os.Stderr, "%s✗ %s not found%s\n", cRust, envFile, cReset)
		return 2
	}

	// Duplicates
	problems = append(problems, findDuplicates(entries)...)

	// Empty values
	problems = append(problems, findEmptyValues(entries)...)

	// Compare with example
	exampleEntries, _ := parseEnvFile(exampleFile)
	if exampleEntries != nil {
		problems = append(problems, compareWithExample(entries, exampleEntries, envFile, exampleFile)...)
	} else if !quiet {
		fmt.Printf("%s⚠ %s not found, skipping comparison%s\n\n", cMuted, exampleFile, cReset)
	}

	// Output
	printProblems(problems, quiet)
	if !quiet {
		printSummary(entries, problems)
	}

	for _, p := range problems {
		if p.Level == "error" {
			return 1
		}
	}
	return 0
}

// parseEnvFile reads and parses an env file into entries
func parseEnvFile(path string) ([]Entry, []Problem) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil
	}
	defer file.Close()

	var entries []Entry
	var problems []Problem
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		trimmed := strings.TrimSpace(scanner.Text())

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		eqIdx := strings.Index(trimmed, "=")
		if eqIdx == -1 {
			problems = append(problems, Problem{
				Level:   "error",
				Line:    lineNum,
				Message: fmt.Sprintf("invalid syntax: %s", trimmed),
			})
			continue
		}

		key := strings.TrimSpace(trimmed[:eqIdx])
		value := strings.TrimSpace(trimmed[eqIdx+1:])
		value = strings.Trim(value, `"'`)

		entries = append(entries, Entry{Key: key, Value: value, Line: lineNum})
	}

	return entries, problems
}

// findDuplicates returns errors for keys appearing more than once
func findDuplicates(entries []Entry) []Problem {
	var problems []Problem
	seen := make(map[string]int)

	for _, e := range entries {
		if firstLine, exists := seen[e.Key]; exists {
			problems = append(problems, Problem{
				Level:   "error",
				Line:    e.Line,
				Message: fmt.Sprintf("duplicate key: %s (first: line %d)", e.Key, firstLine),
			})
		} else {
			seen[e.Key] = e.Line
		}
	}
	return problems
}

// findEmptyValues returns warnings for keys with no value
func findEmptyValues(entries []Entry) []Problem {
	var problems []Problem
	for _, e := range entries {
		if e.Value == "" {
			problems = append(problems, Problem{
				Level:   "warn",
				Line:    e.Line,
				Message: fmt.Sprintf("empty value: %s", e.Key),
			})
		}
	}
	return problems
}

// compareWithExample checks for missing and extra keys between env and example
func compareWithExample(entries, exampleEntries []Entry, envFile, exampleFile string) []Problem {
	var problems []Problem

	envKeys := make(map[string]bool)
	for _, e := range entries {
		envKeys[e.Key] = true
	}

	exampleKeys := make(map[string]bool)
	for _, e := range exampleEntries {
		exampleKeys[e.Key] = true
	}

	for _, e := range exampleEntries {
		if !envKeys[e.Key] {
			problems = append(problems, Problem{
				Level:   "error",
				Message: fmt.Sprintf("missing key: %s (defined in %s)", e.Key, exampleFile),
			})
		}
	}

	for _, e := range entries {
		if !exampleKeys[e.Key] {
			problems = append(problems, Problem{
				Level:   "warn",
				Line:    e.Line,
				Message: fmt.Sprintf("extra key: %s (not in %s)", e.Key, exampleFile),
			})
		}
	}

	return problems
}

func printProblems(problems []Problem, quiet bool) {
	for _, p := range problems {
		if quiet && p.Level != "error" {
			continue
		}

		icon, color := "⚠", cPurple
		if p.Level == "error" {
			icon, color = "✗", cRust
		}

		if p.Line > 0 {
			fmt.Printf("%s%s %s:%d %s%s\n", color, icon, cMuted, p.Line, color+p.Message, cReset)
		} else {
			fmt.Printf("%s%s %s%s\n", color, icon, p.Message, cReset)
		}
	}
}

func printSummary(entries []Entry, problems []Problem) {
	errors, warns := 0, 0
	for _, p := range problems {
		if p.Level == "error" {
			errors++
		} else {
			warns++
		}
	}

	fmt.Printf("\n%s──────────────────────────────────%s\n", cMuted, cReset)
	fmt.Printf("%s%d keys%s", cOlive, len(entries), cReset)

	if errors > 0 {
		fmt.Printf("  %s%d errors%s", cRust, errors, cReset)
	}
	if warns > 0 {
		fmt.Printf("  %s%d warnings%s", cPurple, warns, cReset)
	}
	if errors == 0 && warns == 0 {
		fmt.Printf("  %s✓ all good%s", cOlive, cReset)
	}
	fmt.Println()
}
