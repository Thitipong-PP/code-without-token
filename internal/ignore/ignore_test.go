package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func chdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { os.Chdir(original) })
}

func tempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	chdir(t, dir)
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

// si is shorthand for ShouldIgnore with explicit path and name (distinct values).
func si(r *Rules, path, name string) bool {
	return r.ShouldIgnore(path, name)
}

// siName calls ShouldIgnore with name as both path and name — only appropriate
// for bare-name tests where path-prefix matching is not under test.
func siName(r *Rules, name string) bool {
	return r.ShouldIgnore(name, name)
}

// ──────────────────────────────────────────────
// Group 1 – Built-in defaults
// ──────────────────────────────────────────────

// Every entry in defaultIgnored must be ignored with no ignore files present.
func TestDefaults_AllIgnored(t *testing.T) {
	tempDir(t)
	r := Load()
	for _, name := range defaultIgnored {
		if !siName(r, name) {
			t.Errorf("default %q should be ignored", name)
		}
	}
}

// Default entries must work as path prefixes so nested children are excluded.
// e.g. "node_modules/express/index.js" must be excluded because "node_modules"
// is a default exact entry and prefix-matching applies to all exact entries.
func TestDefaults_AsPathPrefix(t *testing.T) {
	tempDir(t)
	r := Load()

	cases := []struct{ path, name string }{
		{"node_modules/express/index.js", "index.js"},
		{"node_modules/lodash/lodash.js", "lodash.js"},
		{".git/HEAD", "HEAD"},
		{".git/refs/heads/main", "main"},
		{"dist/bundle.js", "bundle.js"},
		{"build/output.bin", "output.bin"},
		{".next/cache/data.json", "data.json"},
	}
	for _, c := range cases {
		if !si(r, c.path, c.name) {
			t.Errorf("expected path %q (name %q) to be ignored via default prefix rule", c.path, c.name)
		}
	}
}

// Non-default names must not be ignored when no ignore files are present.
func TestDefaults_NonDefaultNotIgnored(t *testing.T) {
	tempDir(t)
	r := Load()
	for _, name := range []string{"main.go", "readme.md", "go.mod", "internal", "src"} {
		if siName(r, name) {
			t.Errorf("%q should not be ignored by default", name)
		}
	}
}

// ──────────────────────────────────────────────
// Group 2 – Exact name matching
// ──────────────────────────────────────────────

// Bare names from .gitignore must be matched by name alone.
func TestExactName_GitIgnore(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "ai-context\n.DS_Store\nThumbs.db\n")
	r := Load()
	for _, name := range []string{"ai-context", ".DS_Store", "Thumbs.db"} {
		if !siName(r, name) {
			t.Errorf("expected %q to be ignored via .gitignore exact name", name)
		}
	}
}

// Bare names from .aiignore must be matched by name alone.
func TestExactName_AIIgnore(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".aiignore", "secrets\nlogs\n")
	r := Load()
	for _, name := range []string{"secrets", "logs"} {
		if !siName(r, name) {
			t.Errorf("expected %q to be ignored via .aiignore exact name", name)
		}
	}
}

// Exact name must not cause false positives on unrelated names.
func TestExactName_NoFalsePositive(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor\n")
	r := Load()
	for _, name := range []string{"vendor2", "pre-vendor", "main.go"} {
		if siName(r, name) {
			t.Errorf("%q should not be matched by exact rule 'vendor'", name)
		}
	}
}

// ──────────────────────────────────────────────
// Group 3 – Path-based exact matching
// ──────────────────────────────────────────────

// A path rule must match when the full path is provided.
func TestExactPath_FullPathMatches(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "internal/ignore\n")
	r := Load()
	if !si(r, "internal/ignore", "ignore") {
		t.Error("full path 'internal/ignore' should be ignored")
	}
}

// A path rule must NOT match the bare name alone (path-specific, not name-global).
func TestExactPath_BareNameDoesNotMatch(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "internal/ignore\n")
	r := Load()
	// The bare name "ignore" with path "ignore" must NOT be excluded.
	if si(r, "ignore", "ignore") {
		t.Error("bare name 'ignore' should not match path-specific rule 'internal/ignore'")
	}
}

// A path rule must match children via prefix expansion.
func TestExactPath_PrefixMatchesChildren(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "internal/ignore\n")
	r := Load()
	cases := []struct{ path, name string }{
		{"internal/ignore/ignore.go", "ignore.go"},
		{"internal/ignore/ignore_test.go", "ignore_test.go"},
		{"internal/ignore/sub/deep.go", "deep.go"},
	}
	for _, c := range cases {
		if !si(r, c.path, c.name) {
			t.Errorf("path %q should be excluded by prefix rule 'internal/ignore'", c.path)
		}
	}
}

// A path rule must not spill over into sibling directories.
func TestExactPath_SiblingNotAffected(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "internal/ignore\n")
	r := Load()
	siblings := []struct{ path, name string }{
		{"internal/walker/walker.go", "walker.go"},
		{"internal/builder/builder.go", "builder.go"},
		{"readme.md", "readme.md"},
	}
	for _, c := range siblings {
		if si(r, c.path, c.name) {
			t.Errorf("sibling path %q should not be affected by rule 'internal/ignore'", c.path)
		}
	}
}

// A path rule from .aiignore must also apply prefix matching to its children.
func TestExactPath_AIIgnorePrefixMatchesChildren(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".aiignore", "internal/secret\n")
	r := Load()
	if !si(r, "internal/secret/key.go", "key.go") {
		t.Error("child of .aiignore path rule should be excluded")
	}
	if si(r, "internal/public/api.go", "api.go") {
		t.Error("sibling should not be affected by .aiignore path rule")
	}
}

// ──────────────────────────────────────────────
// Group 4 – Glob pattern matching
// ──────────────────────────────────────────────

// Standard extension globs must match the intended file types.
func TestGlob_ExtensionPatterns(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "*.exe\n*.exe~\n*.dll\n*.so\n*.dylib\n")
	r := Load()

	hits := []string{"app.exe", "app.exe~", "lib.dll", "lib.so", "lib.dylib"}
	for _, name := range hits {
		if !siName(r, name) {
			t.Errorf("expected %q to be ignored by glob", name)
		}
	}
	misses := []string{"main.go", "readme.md", "app.exe.bak", "libso", "dll"}
	for _, name := range misses {
		if siName(r, name) {
			t.Errorf("%q should NOT be ignored by extension glob", name)
		}
	}
}

// The ? wildcard must match exactly one character.
func TestGlob_QuestionMarkWildcard(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "file?.go\n")
	r := Load()
	if !siName(r, "file1.go") {
		t.Error("file1.go should match file?.go")
	}
	if !siName(r, "fileA.go") {
		t.Error("fileA.go should match file?.go")
	}
	// Two characters in place of ? must not match.
	if siName(r, "file12.go") {
		t.Error("file12.go should not match file?.go")
	}
}

// The [bracket] wildcard must match the character set.
func TestGlob_BracketWildcard(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "file[0-9].go\n")
	r := Load()
	if !siName(r, "file0.go") {
		t.Error("file0.go should match file[0-9].go")
	}
	if !siName(r, "file9.go") {
		t.Error("file9.go should match file[0-9].go")
	}
	if siName(r, "fileA.go") {
		t.Error("fileA.go should not match file[0-9].go")
	}
}

// A glob path pattern must match files under the specified directory.
func TestGlob_PathPattern(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "internal/ignore/*\n")
	r := Load()
	if !si(r, "internal/ignore/ignore.go", "ignore.go") {
		t.Error("internal/ignore/ignore.go should match internal/ignore/*")
	}
	if !si(r, "internal/ignore/ignore_test.go", "ignore_test.go") {
		t.Error("internal/ignore/ignore_test.go should match internal/ignore/*")
	}
}

// A glob path pattern must not match files in sibling directories.
func TestGlob_PathPatternNoSiblingSpill(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "internal/ignore/*\n")
	r := Load()
	if si(r, "internal/walker/walker.go", "walker.go") {
		t.Error("internal/walker/walker.go should not match internal/ignore/*")
	}
}

// A glob from .aiignore must work identically to one from .gitignore.
func TestGlob_AIIgnore(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".aiignore", "*.secret\n*.key\n")
	r := Load()
	if !siName(r, "prod.secret") {
		t.Error("prod.secret should be ignored via .aiignore glob")
	}
	if !siName(r, "id_rsa.key") {
		t.Error("id_rsa.key should be ignored via .aiignore glob")
	}
	if siName(r, "main.go") {
		t.Error("main.go should not be ignored")
	}
}

// ──────────────────────────────────────────────
// Group 5 – loadIgnoreFile parsing
// ──────────────────────────────────────────────

// Comment lines (# prefix) must be silently skipped.
func TestParsing_CommentsSkipped(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "# this is a comment\n#vendor\n.DS_Store\n")
	r := Load()
	if siName(r, "# this is a comment") {
		t.Error("comment text should not become a rule")
	}
	if siName(r, "#vendor") {
		t.Error("commented-out entry should not become a rule")
	}
	if !siName(r, ".DS_Store") {
		t.Error(".DS_Store should still be ignored")
	}
}

// Blank and whitespace-only lines must be silently skipped.
func TestParsing_BlankAndWhitespaceLinesSkipped(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "\n   \n\t\nvendor\n\n")
	r := Load()
	// Whitespace lines must not create phantom rules.
	if siName(r, "") || siName(r, "   ") || siName(r, "\t") {
		t.Error("blank/whitespace lines should not produce rules")
	}
	if !siName(r, "vendor") {
		t.Error("vendor should still be ignored")
	}
}

// Trailing slashes must be stripped so "vendor/" behaves like "vendor".
func TestParsing_TrailingSlashStripped(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor/\n.vscode/\n")
	r := Load()
	if !siName(r, "vendor") {
		t.Error("vendor/ should normalise to vendor")
	}
	if !siName(r, ".vscode") {
		t.Error(".vscode/ should normalise to .vscode")
	}
}

// Leading slashes must also be stripped.
func TestParsing_LeadingSlashStripped(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "/vendor\n")
	r := Load()
	if !siName(r, "vendor") {
		t.Error("/vendor should normalise to vendor")
	}
}

// Both leading and trailing slashes must be stripped together.
func TestParsing_BothSlashesStripped(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "/vendor/\n")
	r := Load()
	if !siName(r, "vendor") {
		t.Error("/vendor/ should normalise to vendor")
	}
}

// Duplicate entries must not cause errors or unexpected behaviour.
func TestParsing_DuplicateEntriesHarmless(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor\nvendor\n*.exe\n*.exe\n")
	r := Load()
	if !siName(r, "vendor") {
		t.Error("vendor should be ignored even when listed twice")
	}
	if !siName(r, "app.exe") {
		t.Error("app.exe should be ignored even when pattern listed twice")
	}
}

// Missing ignore files must not cause Load to fail or affect defaults.
func TestParsing_MissingIgnoreFilesAreSkipped(t *testing.T) {
	tempDir(t) // fresh dir with no .gitignore or .aiignore
	r := Load()
	// Defaults must still be intact.
	for _, name := range defaultIgnored {
		if !siName(r, name) {
			t.Errorf("default %q must survive missing ignore files", name)
		}
	}
}

// Empty ignore files must not disturb defaults or produce spurious rules.
func TestParsing_EmptyIgnoreFilesHarmless(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "")
	writeFile(t, dir, ".aiignore", "")
	r := Load()
	for _, name := range defaultIgnored {
		if !siName(r, name) {
			t.Errorf("default %q should survive empty ignore files", name)
		}
	}
	if siName(r, "main.go") {
		t.Error("main.go should not be ignored when files are empty")
	}
}

// Only .gitignore present — .aiignore absence must not affect .gitignore rules.
func TestParsing_OnlyGitIgnorePresent(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor\n")
	r := Load()
	if !siName(r, "vendor") {
		t.Error("vendor from .gitignore should be ignored")
	}
	if siName(r, "secrets") {
		t.Error("secrets should not be ignored when only .gitignore is present")
	}
}

// Only .aiignore present — .gitignore absence must not affect .aiignore rules.
func TestParsing_OnlyAIIgnorePresent(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".aiignore", "secrets\n")
	r := Load()
	if !siName(r, "secrets") {
		t.Error("secrets from .aiignore should be ignored")
	}
	if siName(r, "vendor") {
		t.Error("vendor should not be ignored when only .aiignore is present")
	}
}

// Both files present — rules from both must be merged into one Rules set.
func TestParsing_BothFilesMerged(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor/\n*.exe\n")
	writeFile(t, dir, ".aiignore", "secrets\n*.key\n")
	r := Load()
	for _, name := range []string{"vendor", "app.exe", "secrets", "prod.key"} {
		if !siName(r, name) {
			t.Errorf("%q should be ignored after merging both files", name)
		}
	}
}

// ──────────────────────────────────────────────
// Group 6 – isPattern unit tests
// ──────────────────────────────────────────────

func TestIsPattern(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		// Patterns — contain glob metacharacters
		{"*.exe", true},
		{"*.exe~", true},
		{"file[0].go", true},
		{"file[0-9].go", true},
		{"file?.go", true},
		{"internal/ignore/*", true},
		{"**/*.go", true},
		// Non-patterns — no glob metacharacters
		{"vendor", false},
		{".DS_Store", false},
		{"ai-context", false},
		{"internal/ignore", false},
		{"build", false},
		{".git", false},
		{"", false},
	}
	for _, c := range cases {
		got := isPattern(c.input)
		if got != c.want {
			t.Errorf("isPattern(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

// ──────────────────────────────────────────────
// Group 7 – matchPattern unit tests
// ──────────────────────────────────────────────

func TestMatchPattern_StarGlob(t *testing.T) {
	if !matchPattern("*.go", "main.go") {
		t.Error("*.go should match main.go")
	}
	if matchPattern("*.go", "main.txt") {
		t.Error("*.go should not match main.txt")
	}
}

func TestMatchPattern_QuestionGlob(t *testing.T) {
	if !matchPattern("file?.go", "file1.go") {
		t.Error("file?.go should match file1.go")
	}
	if matchPattern("file?.go", "file10.go") {
		t.Error("file?.go should not match file10.go")
	}
}

func TestMatchPattern_BracketGlob(t *testing.T) {
	if !matchPattern("file[0-9].go", "file3.go") {
		t.Error("file[0-9].go should match file3.go")
	}
	if matchPattern("file[0-9].go", "fileA.go") {
		t.Error("file[0-9].go should not match fileA.go")
	}
}

// The prefix branch: matchPattern("a/b", "a/b/c.go") must return true even
// though filepath.Match("a/b", "a/b/c.go") returns false.
func TestMatchPattern_PrefixBranch(t *testing.T) {
	if !matchPattern("vendor/lib", "vendor/lib/a.go") {
		t.Error("prefix branch: vendor/lib should match vendor/lib/a.go")
	}
	if matchPattern("vendor/lib", "vendor/lib2/a.go") {
		t.Error("prefix branch must not match vendor/lib2/a.go for pattern vendor/lib")
	}
}

// matchPattern must not match an unrelated target.
func TestMatchPattern_NoSpuriousMatch(t *testing.T) {
	if matchPattern("*.exe", "main.go") {
		t.Error("*.exe should not match main.go")
	}
	if matchPattern("vendor", "pre-vendor") {
		t.Error("vendor pattern should not match pre-vendor")
	}
}