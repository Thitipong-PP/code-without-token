package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	taskPtr := flag.String("task", "", "What you want the AI to do (e.g. -task \"add user auth\")")
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

	ignoreDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		".next":        true,
		"dist":         true,
	}

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
		return
	}

	builder.WriteString(fmt.Sprintf("\nCurrently, I am working on this project. I want to: %s.\n", task))
	builder.WriteString("To avoid any side effects, please analyze this structure and tell me exactly which files you need to see the code from?")

	fmt.Println(builder.String())
}