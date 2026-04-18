package walker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thitipong-PP/code-without-token/internal/ignore"
)

// ──────────────────────────────────────────────
// Test helpers
// ──────────────────────────────────────────────

// scaffold creates a temp directory tree from a map of relative-path → content.
func scaffold(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("scaffold mkdir %s: %v", full, err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("scaffold write %s: %v", full, err)
		}
	}
	return root
}

// cdTemp changes the working directory to a fresh temp dir for the duration of
// the test (ignore.Load reads .gitignore / .aiignore from CWD).
func cdTemp(t *testing.T) string {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	return dir
}

// cdInto changes CWD to an existing directory for the duration of the test.
// Used when we need Walk(".") so WalkDir passes relative paths to ShouldIgnore,
// letting path-based rules like "internal/ignore" match correctly.
func cdInto(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

// emptyRules returns Rules with no extra entries (only built-in defaults).
func emptyRules(t *testing.T) *ignore.Rules {
	t.Helper()
	cdTemp(t)
	return ignore.Load()
}

// rulesFrom writes content to .gitignore inside a fresh CWD and returns the
// loaded Rules.
func rulesFrom(t *testing.T, gitignore string) *ignore.Rules {
	t.Helper()
	dir := cdTemp(t)
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	return ignore.Load()
}

// rulesFromAI writes content to .aiignore inside a fresh CWD.
func rulesFromAI(t *testing.T, aiignore string) *ignore.Rules {
	t.Helper()
	dir := cdTemp(t)
	if err := os.WriteFile(filepath.Join(dir, ".aiignore"), []byte(aiignore), 0644); err != nil {
		t.Fatalf("write .aiignore: %v", err)
	}
	return ignore.Load()
}

// rulesFromDir writes a .gitignore into an existing directory, changes CWD to
// that directory, and loads Rules. Use this together with Walk(".") so that
// WalkDir emits relative paths, allowing path-based rules like "internal/ignore"
// to match correctly against the paths ShouldIgnore receives.
func rulesFromDir(t *testing.T, dir, gitignore string) *ignore.Rules {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatalf("write .gitignore in %s: %v", dir, err)
	}
	cdInto(t, dir)
	return ignore.Load()
}

// outputLines splits the walk output into non-blank lines.
func outputLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// hasName reports whether any output line ends with "- <name>".
func hasName(output, name string) bool {
	for _, l := range outputLines(output) {
		if strings.HasSuffix(strings.TrimSpace(l), "- "+name) {
			return true
		}
	}
	return false
}

// indentOf returns the leading-space count of the first line whose trimmed
// suffix matches "- <name>", or -1 if not found.
func indentOf(output, name string) int {
	for _, l := range outputLines(output) {
		if strings.HasSuffix(strings.TrimSpace(l), "- "+name) {
			return len(l) - len(strings.TrimLeft(l, " "))
		}
	}
	return -1
}

// ──────────────────────────────────────────────
// Group 1 – Basic output structure
// ──────────────────────────────────────────────

// The root directory itself must appear as the first line.
func TestWalk_RootAppearsFirst(t *testing.T) {
	root := scaffold(t, map[string]string{"main.go": ""})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ls := outputLines(out)
	if len(ls) == 0 {
		t.Fatal("output is empty")
	}
	// The first line must contain the root's basename.
	// We don't assert zero indent because t.TempDir() returns an absolute path
	// with its own separator depth; the root line inherits that depth.
	rootBase := filepath.Base(root)
	first := ls[0]
	if !strings.HasSuffix(strings.TrimSpace(first), "- "+rootBase) {
		t.Errorf("first line should end with %q, got %q", "- "+rootBase, first)
	}
}

// A single file at root level must appear in output.
func TestWalk_SingleFileAtRoot(t *testing.T) {
	root := scaffold(t, map[string]string{"main.go": ""})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasName(out, "main.go") {
		t.Errorf("expected main.go in output:\n%s", out)
	}
}

// Multiple flat files must all appear.
func TestWalk_MultipleFilesAtRoot(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":   "",
		"readme.md": "",
		"go.mod":    "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"main.go", "readme.md", "go.mod"} {
		if !hasName(out, name) {
			t.Errorf("expected %q in output:\n%s", name, out)
		}
	}
}

// Intermediate directories must appear in output.
func TestWalk_DirectoriesAppearInOutput(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/builder/builder.go": "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"internal", "builder", "builder.go"} {
		if !hasName(out, name) {
			t.Errorf("expected %q in output:\n%s", name, out)
		}
	}
}

// An empty directory (root only) must produce exactly one output line.
func TestWalk_EmptyDirectory_OneLineOutput(t *testing.T) {
	root := t.TempDir()
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ls := outputLines(out)
	if len(ls) != 1 {
		t.Errorf("expected 1 line for empty dir, got %d:\n%s", len(ls), out)
	}
}

// ──────────────────────────────────────────────
// Group 2 – Indentation correctness
// ──────────────────────────────────────────────

// A file nested one level deep must be indented more than its parent directory.
func TestWalk_IndentIncreasesByDepth(t *testing.T) {
	root := scaffold(t, map[string]string{"internal/main.go": ""})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dirIndent := indentOf(out, "internal")
	fileIndent := indentOf(out, "main.go")
	if dirIndent == -1 {
		t.Fatal("internal not found in output")
	}
	if fileIndent == -1 {
		t.Fatal("main.go not found in output")
	}
	if fileIndent <= dirIndent {
		t.Errorf("main.go indent (%d) should be greater than internal indent (%d)", fileIndent, dirIndent)
	}
}

// Each additional nesting level must add exactly two spaces of indentation.
func TestWalk_IndentTwoSpacesPerLevel(t *testing.T) {
	root := scaffold(t, map[string]string{"a/b/c.go": ""})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	aIndent := indentOf(out, "a")
	bIndent := indentOf(out, "b")
	cIndent := indentOf(out, "c.go")

	if aIndent == -1 || bIndent == -1 || cIndent == -1 {
		t.Fatalf("missing entries in output:\n%s", out)
	}
	if bIndent-aIndent != 2 {
		t.Errorf("expected 2-space step from a→b, got %d", bIndent-aIndent)
	}
	if cIndent-bIndent != 2 {
		t.Errorf("expected 2-space step from b→c.go, got %d", cIndent-bIndent)
	}
}

// Sibling files at the same depth must have equal indentation.
func TestWalk_SiblingsHaveEqualIndent(t *testing.T) {
	root := scaffold(t, map[string]string{
		"pkg/alpha.go": "",
		"pkg/beta.go":  "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	aIndent := indentOf(out, "alpha.go")
	bIndent := indentOf(out, "beta.go")
	if aIndent == -1 || bIndent == -1 {
		t.Fatalf("missing entries in output:\n%s", out)
	}
	if aIndent != bIndent {
		t.Errorf("alpha.go indent %d ≠ beta.go indent %d", aIndent, bIndent)
	}
}

// ──────────────────────────────────────────────
// Group 3 – Error handling
// ──────────────────────────────────────────────

// Walking a nonexistent root must return a non-nil error.
func TestWalk_NonexistentRoot_ReturnsError(t *testing.T) {
	_, err := Walk("/nonexistent/path/xyz_abc", emptyRules(t))
	if err == nil {
		t.Error("expected error for nonexistent root, got nil")
	}
}

// The error message must include the bad path.
func TestWalk_ErrorMessageContainsPath(t *testing.T) {
	badPath := "/nonexistent/path/xyz_abc"
	_, err := Walk(badPath, emptyRules(t))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), badPath) {
		t.Errorf("error %q should mention path %q", err.Error(), badPath)
	}
}

// A valid walk must return a nil error.
func TestWalk_ValidRoot_NoError(t *testing.T) {
	root := scaffold(t, map[string]string{"file.go": ""})
	_, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ──────────────────────────────────────────────
// Group 4 – .gitignore-based ignoring
// ──────────────────────────────────────────────

// An exact filename listed in .gitignore must be excluded.
func TestWalk_GitIgnore_ExactFileName(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":   "",
		"secret.txt": "",
	})
	out, err := Walk(root, rulesFrom(t, "secret.txt\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasName(out, "secret.txt") {
		t.Error("secret.txt should be excluded")
	}
	if !hasName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

// A glob pattern must exclude all matching files.
func TestWalk_GitIgnore_GlobPattern(t *testing.T) {
	root := scaffold(t, map[string]string{
		"app.exe":  "",
		"lib.dll":  "",
		"main.go":  "",
	})
	out, err := Walk(root, rulesFrom(t, "*.exe\n*.dll\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"app.exe", "lib.dll"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by glob", name)
		}
	}
	if !hasName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

// An ignored directory must not appear, and its contents must not be walked.
func TestWalk_GitIgnore_DirectorySkipped(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":            "",
		"vendor/lib/a.go":    "",
		"vendor/lib/b.go":    "",
	})
	out, err := Walk(root, rulesFrom(t, "vendor\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"vendor", "lib", "a.go", "b.go"} {
		if hasName(out, name) {
			t.Errorf("%q should not appear — inside ignored dir", name)
		}
	}
	if !hasName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

// A path-based rule must ignore matching subtree without affecting siblings.
// Walk(".") is used so WalkDir emits relative paths that match the rule.
func TestWalk_GitIgnore_ExactPath(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/ignore/ignore.go":      "",
		"internal/ignore/ignore_test.go": "",
		"internal/walker/walker.go":      "",
	})
	rules := rulesFromDir(t, root, "internal/ignore\n")
	out, err := Walk(".", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"ignore.go", "ignore_test.go"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded via path rule", name)
		}
	}
	if !hasName(out, "walker.go") {
		t.Error("walker.go should not be affected by internal/ignore rule")
	}
}

// A glob path rule must exclude all files under the matched path.
// Walk(".") is used so WalkDir emits relative paths that match the rule.
func TestWalk_GitIgnore_GlobPath(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/ignore/ignore.go":      "",
		"internal/ignore/ignore_test.go": "",
		"internal/walker/walker.go":      "",
	})
	rules := rulesFromDir(t, root, "internal/ignore/*\n")
	out, err := Walk(".", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"ignore.go", "ignore_test.go"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by internal/ignore/*", name)
		}
	}
	if !hasName(out, "walker.go") {
		t.Error("walker.go should still appear")
	}
}

// A path rule must not accidentally exclude files in sibling directories.
func TestWalk_GitIgnore_PathRuleDoesNotMatchSiblings(t *testing.T) {
	root := scaffold(t, map[string]string{
		"internal/ignore/ignore.go": "",
		"internal/walker/walker.go": "",
		"readme.md":                 "",
	})
	rules := rulesFromDir(t, root, "internal/ignore\n")
	out, err := Walk(".", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasName(out, "walker.go") {
		t.Error("walker.go should not be affected")
	}
	if !hasName(out, "readme.md") {
		t.Error("readme.md should not be affected")
	}
}

// Comments and blank lines in .gitignore must be ignored.
func TestWalk_GitIgnore_CommentsAndBlanksIgnored(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":   "",
		"readme.md": "",
	})
	gitignore := "# this is a comment\n\n   \n# another comment\n"
	out, err := Walk(root, rulesFrom(t, gitignore))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"main.go", "readme.md"} {
		if !hasName(out, name) {
			t.Errorf("expected %q in output; comments/blanks should not cause exclusion", name)
		}
	}
}

// ──────────────────────────────────────────────
// Group 5 – .aiignore-based ignoring
// ──────────────────────────────────────────────

// Entries in .aiignore must also be excluded.
func TestWalk_AIIgnore_ExcludesFile(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":    "",
		"secrets.go": "",
	})
	out, err := Walk(root, rulesFromAI(t, "secrets.go\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasName(out, "secrets.go") {
		t.Error("secrets.go should be excluded via .aiignore")
	}
	if !hasName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

// A glob pattern in .aiignore must exclude all matching files.
func TestWalk_AIIgnore_GlobPattern(t *testing.T) {
	root := scaffold(t, map[string]string{
		"data.csv":  "",
		"data2.csv": "",
		"main.go":   "",
	})
	out, err := Walk(root, rulesFromAI(t, "*.csv\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"data.csv", "data2.csv"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by .aiignore glob", name)
		}
	}
	if !hasName(out, "main.go") {
		t.Error("main.go should still appear")
	}
}

// ──────────────────────────────────────────────
// Group 6 – Built-in default ignores
// ──────────────────────────────────────────────

// .git directory must be excluded by default with no ignore files present.
func TestWalk_Default_GitDirExcluded(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":       "",
		".git/HEAD":     "ref: refs/heads/main",
		".git/config":   "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{".git", "HEAD", "config"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by default rules", name)
		}
	}
}

// node_modules must be excluded by default.
func TestWalk_Default_NodeModulesExcluded(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":                        "",
		"node_modules/express/index.js":  "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"node_modules", "express", "index.js"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by default rules", name)
		}
	}
}

// dist directory must be excluded by default.
func TestWalk_Default_DistExcluded(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":           "",
		"dist/bundle.js":    "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"dist", "bundle.js"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by default rules", name)
		}
	}
}

// build directory must be excluded by default.
func TestWalk_Default_BuildExcluded(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":          "",
		"build/output.bin": "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"build", "output.bin"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by default rules", name)
		}
	}
}

// .next directory must be excluded by default.
func TestWalk_Default_NextDirExcluded(t *testing.T) {
	root := scaffold(t, map[string]string{
		"main.go":              "",
		".next/cache/data.json": "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{".next", "cache", "data.json"} {
		if hasName(out, name) {
			t.Errorf("%q should be excluded by default rules", name)
		}
	}
}

// ──────────────────────────────────────────────
// Group 7 – Output format correctness
// ──────────────────────────────────────────────

// Every non-blank output line must match the "  …- <name>" format.
func TestWalk_OutputFormat_EveryLineHasDash(t *testing.T) {
	root := scaffold(t, map[string]string{
		"pkg/util.go": "",
		"main.go":     "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, l := range outputLines(out) {
		trimmed := strings.TrimLeft(l, " ")
		if !strings.HasPrefix(trimmed, "- ") {
			t.Errorf("line does not start with '- ' after indent: %q", l)
		}
	}
}

// Output must end with a newline (no trailing garbage).
func TestWalk_OutputFormat_TrailingNewline(t *testing.T) {
	root := scaffold(t, map[string]string{"main.go": ""})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("output should end with newline, got %q", out)
	}
}

// An empty directory must still produce non-empty output (the root line).
func TestWalk_OutputFormat_EmptyDirNotEmptyString(t *testing.T) {
	root := t.TempDir()
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("output for empty dir should not be blank")
	}
}

// ──────────────────────────────────────────────
// Group 8 – Deep / complex trees
// ──────────────────────────────────────────────

// A deeply nested (4-level) structure must be fully represented.
func TestWalk_DeepNesting(t *testing.T) {
	root := scaffold(t, map[string]string{
		"a/b/c/d/deep.go": "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"a", "b", "c", "d", "deep.go"} {
		if !hasName(out, name) {
			t.Errorf("expected %q in deep-nest output:\n%s", name, out)
		}
	}
}

// Files in multiple sibling directories must all appear.
func TestWalk_MultipleSiblingDirs(t *testing.T) {
	root := scaffold(t, map[string]string{
		"cmd/main.go":       "",
		"internal/app.go":   "",
		"pkg/util.go":       "",
	})
	out, err := Walk(root, emptyRules(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"cmd", "main.go", "internal", "app.go", "pkg", "util.go"} {
		if !hasName(out, name) {
			t.Errorf("expected %q in output:\n%s", name, out)
		}
	}
}

// Ignoring one subtree must leave all other subtrees intact.
// Walk(".") is used so WalkDir emits relative paths that match the path rule.
func TestWalk_IgnoreOneSubtree_OthersIntact(t *testing.T) {
	root := scaffold(t, map[string]string{
		"cmd/main.go":            "",
		"internal/secret/key.go": "",
		"internal/public/api.go": "",
		"pkg/util.go":            "",
	})
	rules := rulesFromDir(t, root, "internal/secret\n")
	out, err := Walk(".", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasName(out, "key.go") {
		t.Error("key.go should be excluded")
	}
	for _, name := range []string{"cmd", "main.go", "internal", "public", "api.go", "pkg", "util.go"} {
		if !hasName(out, name) {
			t.Errorf("%q should still appear:\n%s", name, out)
		}
	}
}