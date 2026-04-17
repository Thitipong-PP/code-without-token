package cli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds all parsed input from the user.
type Config struct {
	Task     string
	Includes []string
}

// Parse reads flags and, if needed, falls into interactive mode.
// Returns a filled Config or exits on invalid input.
func Parse() Config {
	taskPtr := flag.String("task", "", "What you want the AI to do (e.g. -task \"add user auth\")")
	includePtr := flag.String("include", "", "Comma-separated files to include content from (e.g. main.go,utils.go)")

	flag.Usage = func() {
		fmt.Println("CODE WITHOUT TOKEN: Generate project context for AI assistance")
		fmt.Println("\nUsage:")
		fmt.Println("  ai-context [flags] (or run without flags for Interactive Mode)")
		fmt.Println("\nExamples:")
		fmt.Println("  ai-context")
		fmt.Println("  ai-context -task \"add login api\"")
		fmt.Println("  ai-context -task \"fix bug\" -include \"main.go,go.mod\"")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	task := strings.TrimSpace(*taskPtr)
	includesRaw := strings.TrimSpace(*includePtr)

	// Interactive mode: no flags provided at all
	if task == "" && includesRaw == "" {
		task = promptTask()
	}

	return Config{
		Task:     task,
		Includes: parseIncludes(includesRaw),
	}
}

// promptTask asks the user interactively for a task description.
func promptTask() string {
	fmt.Print("What is your task?: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	task := strings.TrimSpace(input)

	if task == "" {
		fmt.Println("Error: Task cannot be empty.")
		os.Exit(1)
	}

	return task
}

// parseIncludes splits a comma-separated file list into a clean slice.
// Returns an empty slice if the input is blank.
func parseIncludes(raw string) []string {
	if raw == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		if clean := strings.TrimSpace(p); clean != "" {
			result = append(result, clean)
		}
	}

	return result
}