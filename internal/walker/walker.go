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

		// Skip ignored entries — directories are skipped entirely (no descend),
		// files are simply omitted from output.
		if rules.ShouldIgnore(path, d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
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

// ListFiles traverses the directory tree and returns a flat list of file paths,
// skipping any entry that matches the given ignore rules. Directories are omitted.
func ListFiles(root string, rules *ignore.Rules) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if rules.ShouldIgnore(path, d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only append files, not directories, to the selection list
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("listing files %q: %w", root, err)
	}

	return files, nil
}