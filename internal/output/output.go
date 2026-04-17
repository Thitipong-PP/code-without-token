package output

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
)

// Print displays the generated context to stdout with a visible border.
func Print(text string) {
	fmt.Println("\n================ GENERATED CONTEXT ================")
	fmt.Println(text)
	fmt.Println("===================================================")
}

// CopyToClipboard asks the user for confirmation and, if accepted,
// copies the given text to the system clipboard.
func CopyToClipboard(text string) {
	fmt.Print("\nDo you want to copy the result to clipboard? (Y/n): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	// Default to yes on empty input
	if input != "" && input != "y" && input != "yes" {
		return
	}

	if err := clipboard.WriteAll(text); err != nil {
		fmt.Println("\n--- Failed to copy to clipboard. Printing output instead ---")
		return
	}

	fmt.Println("Successfully generated context and copied to clipboard!")
	fmt.Println("You can now press Ctrl+V / Cmd+V in your AI chat.")
}