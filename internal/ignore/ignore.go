package ignore

import (
	"bufio"
	"os"
	"strings"
)

// defaultIgnored contains directories that are always skipped
// regardless of what .gitignore says.
var defaultIgnored = []string{
	".git",
	"node_modules",
	".next",
	"dist",
	"build",
}

// Rules holds the complete set of ignore entries for the current project.
type Rules struct {
	entries map[string]bool
}

// Load reads .gitignore from the current directory and merges it with
// the hardcoded defaults. It never fails — if .gitignore is missing
// it simply returns the defaults.
func Load() *Rules {
	r := &Rules{
		entries: make(map[string]bool),
	}

	for _, name := range defaultIgnored {
		r.entries[name] = true
	}

	r.loadGitignoreFile(".gitignore")

	return r
}

// ShouldIgnore reports whether the given file or directory name
// matches any of the active ignore rules.
func (r *Rules) ShouldIgnore(name string) bool {
	return r.entries[name]
}

// loadGitignoreFile parses a .gitignore file and adds its entries to the rules.
func (r *Rules) loadGitignoreFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Normalize: strip leading/trailing slashes
		line = strings.Trim(line, "/")
		r.entries[line] = true
	}
}