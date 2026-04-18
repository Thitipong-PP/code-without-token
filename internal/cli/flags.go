package cli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Thitipong-PP/code-without-token/internal/ignore"
	"github.com/Thitipong-PP/code-without-token/internal/walker"
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
		// ... existing usage prints ...
	}

	flag.Parse()

	task := strings.TrimSpace(*taskPtr)
	includesRaw := strings.TrimSpace(*includePtr)

	// Interactive mode: no flags and no arguments provided at all
	if task == "" && includesRaw == "" && len(os.Args) == 1 {
		selected := RunMainMenu()

		switch selected {
		case "Run Task":
			// If they only chose to run a task, a task is REQUIRED.
			task = promptTask(true)

		case "Select Files":
			rules := ignore.Load()
			files, err := walker.ListFiles(".", rules)
			if err != nil {
				fmt.Printf("Error listing files: %v\n", err)
				os.Exit(1)
			}

			if len(files) == 0 {
				fmt.Println("No selectable files found.")
				os.Exit(0)
			}

			selectedFiles := RunFilePicker(files)
			if len(selectedFiles) == 0 {
				fmt.Println("No files selected. Exiting...")
				os.Exit(0)
			}

			includesRaw = strings.Join(selectedFiles, ",")
			fmt.Printf("\nSelected %d file(s).\n", len(selectedFiles))

			// Since they have files, the task is now OPTIONAL.
			task = promptTask(false)

		case "Exit", "":
			fmt.Println("\nExiting...")
			os.Exit(0)
		}
	}

	return Config{
		Task:     task,
		Includes: parseIncludes(includesRaw),
	}
}

// promptTask asks the user interactively for a task description.
// If required is true, it forces the user to enter text.
func promptTask(required bool) string {
	if required {
		fmt.Print("What is your task?: ")
	} else {
		fmt.Print("What is your task? (Press Enter to skip and just copy files): ")
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	task := strings.TrimSpace(input)

	if required && task == "" {
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