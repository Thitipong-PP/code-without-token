package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// defaultIgnored contains exact names that are always skipped.
var defaultIgnored = []string{
	".git",
	"node_modules",
	".next",
	"dist",
	"build",
}

// Rules holds the complete set of ignore patterns for the current project.
type Rules struct {
	exact    map[string]bool // fast path: literal names like ".DS_Store"
	patterns []string        // glob patterns like "*.exe", "vendor"
}

// Load reads both .gitignore and .aiignore from the current directory
// and merges them with the hardcoded defaults. It never fails — missing
// files are silently skipped.
func Load() *Rules {
	r := &Rules{
		exact: make(map[string]bool),
	}

	for _, name := range defaultIgnored {
		r.exact[name] = true
	}

	r.loadIgnoreFile(".gitignore")
	r.loadIgnoreFile(".aiignore")

	return r
}

// ShouldIgnore reports whether the given file or directory name
// matches any active ignore rule — exact names or glob patterns.
func (r *Rules) ShouldIgnore(name string) bool {
	// Fast path: exact match
	if r.exact[name] {
		return true
	}

	// Pattern match: e.g. *.exe, vendor, .vscode
	for _, pattern := range r.patterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// loadIgnoreFile parses any .gitignore-style file and adds its entries to the rules.
// It silently returns if the file does not exist or cannot be read.
func (r *Rules) loadIgnoreFile(path string) {
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

		// Use fast-path exact map for plain names, patterns slice for globs
		if isPattern(line) {
			r.patterns = append(r.patterns, line)
		} else {
			r.exact[line] = true
		}
	}
}

// isPattern reports whether a line requires glob matching.
// Plain names (e.g. "vendor", ".DS_Store") go into the exact map;
// anything with *, ?, or [ goes into the patterns slice.
func isPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}