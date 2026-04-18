package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var defaultIgnored = []string{
	".git",
	"node_modules",
	".next",
	"dist",
	"build",
}

type Rules struct {
	exact    map[string]bool
	patterns []string
}

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

// ShouldIgnore reports whether the entry should be excluded.
// path is the full relative path (e.g. "internal/ignore/ignore.go").
// name is the bare filename (e.g. "ignore.go").
func (r *Rules) ShouldIgnore(path, name string) bool {
	path = filepath.ToSlash(path)

	// 1. Exact match on bare name or full path
	if r.exact[name] || r.exact[path] {
		return true
	}

	// 2. Exact entry used as a path prefix
	//    "internal/ignore" should match "internal/ignore/ignore.go"
	for entry := range r.exact {
		if strings.HasPrefix(path, entry+"/") {
			return true
		}
	}

	// 3. Glob pattern against name and full path
	for _, pattern := range r.patterns {
		if matchPattern(pattern, name) || matchPattern(pattern, path) {
			return true
		}
	}

	return false
}

func (r *Rules) loadIgnoreFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.Trim(filepath.ToSlash(line), "/")
		if isPattern(line) {
			r.patterns = append(r.patterns, line)
		} else {
			r.exact[line] = true
		}
	}
}

func matchPattern(pattern, target string) bool {
	if matched, err := filepath.Match(pattern, target); err == nil && matched {
		return true
	}
	if strings.HasPrefix(target, pattern+"/") {
		return true
	}
	return false
}

func isPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}