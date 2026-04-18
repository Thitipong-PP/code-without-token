package walker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thitipong-PP/code-without-token/internal/ignore"
)

func scaffold(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", full, err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
	return root
}

func emptyRules(t *testing.T) *ignore.Rules {
	t.Helper()
	orig, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(orig) })
	return ignore.Load()
}

func rulesWithEntries(t *testing.T, gitignore string) *ignore.Rules {
	t.Helper()
	orig, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(orig) })
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644)
	return ignore.Load()
}

func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

func containsName(output, name string) bool {
	for _, l := range lines(output) {
		if strings.HasSuffix(strings.TrimSpace(l), "- "+name) {
			return true
		}
	}
	return false
}

// --- Basic output ---

func TestWalkReturnsRootEntry(t *testing.T) {
	root := scaffold(t, map[string]string{"main.go": ""})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines(out)) == 0 {
		t.Error("expected at least one line in output")
	}
}

func TestWalkShowsFiles(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go": "", "readme.md": "", "go.mod": "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"main.go", "readme.md", "go.mod"} {
		if !containsName(out, name) {
			t.Errorf("expected %q in output", name)
		}
	}
}

func TestWalkShowsNestedFiles(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/builder/builder.go": "",
		"internal/cli/flags.go":       "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"internal", "builder", "builder.go", "cli", "flags.go"} {
		if !containsName(out, name) {
			t.Errorf("expected %q in output", name)
		}
	}
}

func TestWalkIndentsByDepth(t *testing.T) {
	root := scaffold(t, map[string]string{"internal/main.go": ""})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, l := range lines(out) {
		if strings.HasSuffix(strings.TrimSpace(l), "- main.go") {
			if !strings.HasPrefix(l, "  ") {
				t.Errorf("expected main.go to be indented, got %q", l)
			}
			return
		}
	}
	t.Error("main.go not found in output")
}

// --- Name-based ignoring ---

func TestWalkIgnoresExactFileName(t *testing.T) {
	root := scaffold(t, map[string]string{"main.go": "", "readme.md": ""})
	out, err := Walk(root, rulesWithEntries(t, "readme.md\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if containsName(out, "readme.md") {
		t.Error("readme.md should be excluded")
	}
	if !containsName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

func TestWalkIgnoresGlobPattern(t *testing.T) {
	root := scaffold(t, map[string]string{
		"app.exe": "", "lib.dll": "", "main.go": "",
	})
	out, err := Walk(root, rulesWithEntries(t, "*.exe\n*.dll\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"app.exe", "lib.dll"} {
		if containsName(out, name) {
			t.Errorf("%q should be excluded by glob", name)
		}
	}
	if !containsName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

func TestWalkIgnoresDirectory(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go": "", "vendor/lib/a.go": "",
	})
	out, err := Walk(root, rulesWithEntries(t, "vendor/\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"vendor", "lib", "a.go"} {
		if containsName(out, name) {
			t.Errorf("%q should not appear", name)
		}
	}
	if !containsName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

func TestWalkIgnoredDirectoryContentsNotWalked(t *testing.T) {
	root := scaffold(t, map[string]string{
		"node_modules/express/index.js": "", "main.go": "",
	})
	out, err := Walk(root, rulesWithEntries(t, "node_modules\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"node_modules", "express", "index.js"} {
		if containsName(out, name) {
			t.Errorf("%q should not appear — inside ignored dir", name)
		}
	}
}

// --- Path-based ignoring (the new behaviour) ---

func TestWalkIgnoresExactPath(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/ignore/ignore.go":      "",
		"internal/ignore/ignore_test.go": "",
		"internal/walker/walker.go":      "",
	})
	out, err := Walk(root, rulesWithEntries(t, "internal/ignore\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The ignore dir and its contents should be gone
	for _, name := range []string{"ignore.go", "ignore_test.go"} {
		if containsName(out, name) {
			t.Errorf("%q should be excluded via path rule", name)
		}
	}
	// Sibling package should still appear
	if !containsName(out, "walker.go") {
		t.Error("walker.go should still appear")
	}
}

func TestWalkIgnoresGlobPath(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/ignore/ignore.go":      "",
		"internal/ignore/ignore_test.go": "",
		"internal/walker/walker.go":      "",
	})
	out, err := Walk(root, rulesWithEntries(t, "internal/ignore/*\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"ignore.go", "ignore_test.go"} {
		if containsName(out, name) {
			t.Errorf("%q should be excluded by internal/ignore/*", name)
		}
	}
	if !containsName(out, "walker.go") {
		t.Error("walker.go should still appear")
	}
}

func TestWalkPathRuleDoesNotMatchSiblings(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/ignore/ignore.go": "",
		"internal/walker/walker.go": "",
		"readme.md":                 "",
	})
	// Only ignore the ignore package by path
	out, err := Walk(root, rulesWithEntries(t, "internal/ignore\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsName(out, "walker.go") {
		t.Error("walker.go should not be affected by internal/ignore rule")
	}
	if !containsName(out, "readme.md") {
		t.Error("readme.md should not be affected by internal/ignore rule")
	}
}

// --- Edge cases ---

func TestWalkEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines(out)) != 1 {
		t.Errorf("expected 1 line for empty dir, got %d: %q", len(lines(out)), out)
	}
}

func TestWalkReturnsErrorOnBadRoot(t *testing.T) {
	_, err := Walk("/nonexistent/path/xyz", emptyRules(t))
	if err == nil {
		t.Error("expected error for nonexistent root")
	}
}