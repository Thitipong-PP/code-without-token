package builder

import (
	"fmt"
	"strings"

	"github.com/Thitipong-PP/code-without-token/internal/cli"
	"github.com/Thitipong-PP/code-without-token/internal/content"
	"github.com/Thitipong-PP/code-without-token/internal/ignore"
	"github.com/Thitipong-PP/code-without-token/internal/walker"
)

// Build assembles the final context string from the given config.
// It walks the project structure when a task is set, appends file
// contents when includes are provided, and returns the complete prompt.
func Build(cfg cli.Config) (string, error) {
	var sb strings.Builder

	if cfg.Task != "" {
		rules := ignore.Load()

		structure, err := walker.Walk(".", rules)
		if err != nil {
			return "", fmt.Errorf("building structure: %w", err)
		}

		sb.WriteString("Here is the project structure:\n")
		sb.WriteString(structure)
		sb.WriteString(fmt.Sprintf(
			"\nCurrently, I am working on this project. I want to: %s.\n",
			cfg.Task,
		))
		sb.WriteString("To avoid any side effects, please analyze this structure and tell me exactly which files you need to see the code from?")
	}

	if len(cfg.Includes) > 0 {
		files := content.ReadFiles(cfg.Includes)
		for _, f := range files {
			sb.WriteString(content.FormatFile(f))
		}
	}

	return sb.String(), nil
}