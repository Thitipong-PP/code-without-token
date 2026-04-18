package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

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

// --- Defaults ---

func TestDefaultsAlwaysIgnored(t *testing.T) {
	tempDir(t)
	r := Load()
	for _, name := range defaultIgnored {
		if !r.ShouldIgnore(name) {
			t.Errorf("expected %q to be ignored by default", name)
		}
	}
}

func TestNoIgnoreFilesPresent(t *testing.T) {
	tempDir(t)
	r := Load()
	if !r.ShouldIgnore(".git") {
		t.Error("expected .git to be ignored")
	}
	if r.ShouldIgnore("main.go") {
		t.Error("main.go should not be ignored")
	}
}

// --- Exact names ---

func TestExactNameFromGitignore(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "ai-context\n.DS_Store\nThumbs.db\n")
	r := Load()
	for _, name := range []string{"ai-context", ".DS_Store", "Thumbs.db"} {
		if !r.ShouldIgnore(name) {
			t.Errorf("expected %q to be ignored", name)
		}
	}
}

func TestExactNameFromAiignore(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".aiignore", "secrets\nlogs\n")
	r := Load()
	for _, name := range []string{"secrets", "logs"} {
		if !r.ShouldIgnore(name) {
			t.Errorf("expected %q to be ignored via .aiignore", name)
		}
	}
}

// --- Glob patterns ---

func TestGlobExePattern(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "*.exe\n*.exe~\n*.dll\n*.so\n*.dylib\n")
	r := Load()

	hits := []string{"app.exe", "app.exe~", "lib.dll", "lib.so", "lib.dylib"}
	for _, name := range hits {
		if !r.ShouldIgnore(name) {
			t.Errorf("expected %q to be ignored by glob pattern", name)
		}
	}

	misses := []string{"main.go", "readme.md", "app.exe.bak"}
	for _, name := range misses {
		if r.ShouldIgnore(name) {
			t.Errorf("%q should NOT be ignored", name)
		}
	}
}

func TestGlobFromAiignore(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".aiignore", "*.secret\n*.key\n")
	r := Load()

	if !r.ShouldIgnore("prod.secret") {
		t.Error("expected prod.secret to be ignored")
	}
	if !r.ShouldIgnore("id_rsa.key") {
		t.Error("expected id_rsa.key to be ignored")
	}
	if r.ShouldIgnore("main.go") {
		t.Error("main.go should not be ignored")
	}
}

// --- Directory-style entries (trailing slash stripped) ---

func TestDirectoryEntryStripsSlash(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor/\n.vscode/\n")
	r := Load()

	if !r.ShouldIgnore("vendor") {
		t.Error("expected vendor/ (normalized) to be ignored")
	}
	if !r.ShouldIgnore(".vscode") {
		t.Error("expected .vscode/ (normalized) to be ignored")
	}
}

// --- Merge behaviour ---

func TestBothFilesAreMerged(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor/\n*.exe\n")
	writeFile(t, dir, ".aiignore", "secrets\n*.key\n")
	r := Load()

	if !r.ShouldIgnore("vendor") {
		t.Error("expected vendor from .gitignore")
	}
	if !r.ShouldIgnore("app.exe") {
		t.Error("expected *.exe glob from .gitignore")
	}
	if !r.ShouldIgnore("secrets") {
		t.Error("expected secrets from .aiignore")
	}
	if !r.ShouldIgnore("prod.key") {
		t.Error("expected *.key glob from .aiignore")
	}
}

func TestOnlyGitignorePresent(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor\n")
	r := Load()
	if !r.ShouldIgnore("vendor") {
		t.Error("expected vendor")
	}
	if r.ShouldIgnore("secrets") {
		t.Error("secrets should not be ignored")
	}
}

func TestOnlyAiignorePresent(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".aiignore", "secrets\n")
	r := Load()
	if !r.ShouldIgnore("secrets") {
		t.Error("expected secrets")
	}
	if r.ShouldIgnore("vendor") {
		t.Error("vendor should not be ignored")
	}
}

// --- Parsing edge cases ---

func TestCommentsAndBlankLinesSkipped(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "# Mac\n\n.DS_Store\n")
	r := Load()
	if r.ShouldIgnore("# Mac") {
		t.Error("comment should not be an entry")
	}
	if !r.ShouldIgnore(".DS_Store") {
		t.Error("expected .DS_Store")
	}
}

func TestDuplicateEntriesAreHarmless(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "vendor\nvendor\n*.exe\n*.exe\n")
	r := Load()
	if !r.ShouldIgnore("vendor") {
		t.Error("expected vendor")
	}
	if !r.ShouldIgnore("app.exe") {
		t.Error("expected app.exe")
	}
}

func TestDefaultsNotOverriddenByEmptyFiles(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, ".gitignore", "")
	writeFile(t, dir, ".aiignore", "")
	r := Load()
	for _, name := range defaultIgnored {
		if !r.ShouldIgnore(name) {
			t.Errorf("default %q should still be ignored", name)
		}
	}
}

// --- isPattern helper ---

func TestIsPattern(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"*.exe", true},
		{"*.exe~", true},
		{"file[0].go", true},
		{"file?.go", true},
		{"vendor", false},
		{".DS_Store", false},
		{"ai-context", false},
	}
	for _, c := range cases {
		if got := isPattern(c.input); got != c.want {
			t.Errorf("isPattern(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}