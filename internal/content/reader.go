package content

import (
	"fmt"
	"os"
)

// File holds the name and raw text content of a single included file.
type File struct {
	Name    string
	Content string
}

// ReadFiles attempts to read each file in the given list.
// Files that cannot be read are skipped; a warning is printed for each.
// Returns the successfully read files.
func ReadFiles(paths []string) []File {
	results := make([]File, 0, len(paths))

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Warning: could not read file %q: %v\n", path, err)
			continue
		}

		results = append(results, File{
			Name:    path,
			Content: string(data),
		})
	}

	return results
}

// FormatFile returns the standard markdown block representation of a file,
// matching the format expected by most AI assistants.
func FormatFile(f File) string {
	return fmt.Sprintf("\n\n--- Content of %s ---\n```\n%s\n```\n", f.Name, f.Content)
}