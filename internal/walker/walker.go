package walker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Thitipong-PP/code-without-token/internal/ignore"
)

// Walk traverses the directory tree starting at root, skipping any entry
// that matches the given ignore rules. It returns a formatted tree string
// (one entry per line, indented by depth) or an error.
func Walk(root string, rules *ignore.Rules) (string, error) {
	var sb strings.Builder

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories entirely (don't descend into them)
		if d.IsDir() && rules.ShouldIgnore(d.Name()) {
			return filepath.SkipDir
		}

		depth := strings.Count(path, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)
		sb.WriteString(fmt.Sprintf("%s- %s\n", indent, d.Name()))

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("walking directory %q: %w", root, err)
	}

	return sb.String(), nil
}