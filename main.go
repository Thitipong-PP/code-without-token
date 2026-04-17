package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
)

// Load .gitignore entries into a map for quick lookup
func loadGitignore() map[string]bool {
	ignoreMap := map[string]bool{
		".git":         true,
		"node_modules": true,
		".next":        true,
		"dist":         true,
		"build":        true,
	}

	file, err := os.Open(".gitignore")
	if err != nil {
		return ignoreMap
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.Trim(line, "/")
		ignoreMap[line] = true
	}
	return ignoreMap
}

func main() {
	// Input handling
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

	task := *taskPtr

	if task == "" {
		fmt.Print("What is your task?: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		task = strings.TrimSpace(input)
	}

	if task == "" {
		task = "analyze the code and suggest improvements"
	}

	ignoreDirs := loadGitignore()

	var builder strings.Builder
	builder.WriteString("Here is the project structure:\n")

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && ignoreDirs[d.Name()] {
			return filepath.SkipDir
		}

		depth := strings.Count(path, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)
		builder.WriteString(fmt.Sprintf("%s- %s\n", indent, d.Name()))
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	includes := strings.TrimSpace(*includePtr)
	if includes != "" {
		filesToInclude := strings.Split(includes, ",")
		for _, fileName := range filesToInclude {
			cleanFileName := strings.TrimSpace(fileName)
			content, err := os.ReadFile(cleanFileName)
			if err == nil {
				builder.WriteString(fmt.Sprintf("\n\n--- Content of %s ---\n```\n%s\n```\n", cleanFileName, string(content)))
			} else {
				fmt.Printf(" Warning: Could not read file '%s'\n", cleanFileName)
			}
		}
	}

	builder.WriteString(fmt.Sprintf("\nCurrently, I am working on this project. I want to: %s.\n", task))
	builder.WriteString("To avoid any side effects, please analyze this structure and tell me exactly which files you need to see the code from?")

	finalText := builder.String()

	fmt.Println("\n================ GENERATED CONTEXT ================")
	fmt.Println(finalText)
	fmt.Println("===================================================")

	fmt.Print("\nDo you want to copy the result to clipboard? (Y/n): ")
	confirmReader := bufio.NewReader(os.Stdin)
	copyInput, _ := confirmReader.ReadString('\n')
	copyInput = strings.TrimSpace(strings.ToLower(copyInput))

	if copyInput == "" || copyInput == "y" || copyInput == "yes" {
		err = clipboard.WriteAll(finalText)
		if err != nil {
			fmt.Println("\n--- Failed to copy to clipboard. Printing output instead ---")
		} else {
			fmt.Println("Successfully generated context and copied to clipboard!")
			fmt.Println("You can now press Ctrl+V / Cmd+V in your AI chat.")
		}
	}
}