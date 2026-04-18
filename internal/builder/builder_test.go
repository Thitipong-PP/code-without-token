package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thitipong-PP/code-without-token/internal/cli"
)

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

// setupDir creates a temp directory, writes the given files into it,
// changes CWD to it (restored on cleanup), and returns the dir path.
// ignore.Load() and walker.Walk(".") both depend on CWD.
func setupDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	for rel, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	return dir
}

// ──────────────────────────────────────────────
// Task-only mode
// ──────────────────────────────────────────────

// When only Task is set, Build must produce structure + task prompt with no error.
func TestBuild_TaskOnly_ReturnsStructureAndPrompt(t *testing.T) {
	setupDir(t, map[string]string{"main.go": "package main"})

	out, err := Build(cli.Config{Task: "add user auth"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Here is the project structure:") {
		t.Error("expected structure header in output")
	}
	if !strings.Contains(out, "add user auth") {
		t.Error("expected task text in output")
	}
}

// Output must contain the ai-context template hint.
func TestBuild_TaskOnly_ContainsTemplateHint(t *testing.T) {
	setupDir(t, map[string]string{"main.go": ""})

	out, err := Build(cli.Config{Task: "fix bug"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ai-context -include") {
		t.Error("expected ai-context template hint in output")
	}
}

// Output must contain the "analyze this structure" instruction.
func TestBuild_TaskOnly_ContainsAnalyzeInstruction(t *testing.T) {
	setupDir(t, map[string]string{"main.go": ""})

	out, err := Build(cli.Config{Task: "refactor"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "analyze this structure") {
		t.Error("expected analyze instruction in output")
	}
}

// The file tree must appear between the structure header and the task line.
func TestBuild_TaskOnly_StructureContainsFiles(t *testing.T) {
	setupDir(t, map[string]string{
		"main.go":   "",
		"go.mod":    "",
		"readme.md": "",
	})

	out, err := Build(cli.Config{Task: "add tests"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"main.go", "go.mod", "readme.md"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q to appear in structure output", name)
		}
	}
}

// ──────────────────────────────────────────────
// Includes-only mode
// ──────────────────────────────────────────────

// When only Includes is set (no Task), Build must return file content with no structure header.
func TestBuild_IncludesOnly_ReturnsFileContent(t *testing.T) {
	dir := setupDir(t, map[string]string{"main.go": "package main\n"})

	out, err := Build(cli.Config{Includes: []string{filepath.Join(dir, "main.go")}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "package main") {
		t.Error("expected file content in output")
	}
}

// Structure header must NOT appear when Task is empty.
func TestBuild_IncludesOnly_NoStructureHeader(t *testing.T) {
	dir := setupDir(t, map[string]string{"main.go": "x"})

	out, err := Build(cli.Config{Includes: []string{filepath.Join(dir, "main.go")}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Here is the project structure:") {
		t.Error("structure header should not appear when Task is empty")
	}
}

// Multiple included files must all appear in output.
func TestBuild_IncludesOnly_MultipleFiles(t *testing.T) {
	dir := setupDir(t, map[string]string{
		"a.go": "package a",
		"b.go": "package b",
	})

	out, err := Build(cli.Config{
		Includes: []string{
			filepath.Join(dir, "a.go"),
			filepath.Join(dir, "b.go"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "package a") {
		t.Error("expected a.go content")
	}
	if !strings.Contains(out, "package b") {
		t.Error("expected b.go content")
	}
}

// ──────────────────────────────────────────────
// Combined mode (Task + Includes)
// ──────────────────────────────────────────────

// When both Task and Includes are set, output must contain both structure prompt and file content.
func TestBuild_TaskAndIncludes_BothPresent(t *testing.T) {
	dir := setupDir(t, map[string]string{"main.go": "package main\n"})

	out, err := Build(cli.Config{
		Task:     "add tests",
		Includes: []string{filepath.Join(dir, "main.go")},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Here is the project structure:") {
		t.Error("expected structure header")
	}
	if !strings.Contains(out, "package main") {
		t.Error("expected file content")
	}
	if !strings.Contains(out, "add tests") {
		t.Error("expected task text")
	}
}

// ──────────────────────────────────────────────
// Empty config
// ──────────────────────────────────────────────

// An empty Config must return an empty string with no error.
func TestBuild_EmptyConfig_ReturnsEmptyString(t *testing.T) {
	setupDir(t, map[string]string{"main.go": ""})

	out, err := Build(cli.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty string for empty config, got %q", out)
	}
}

// ──────────────────────────────────────────────
// Task text appears verbatim in output
// ──────────────────────────────────────────────

func TestBuild_TaskAppearsVerbatim(t *testing.T) {
	setupDir(t, map[string]string{"main.go": ""})
	task := "implement OAuth2 login with refresh tokens"

	out, err := Build(cli.Config{Task: task})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, task) {
		t.Errorf("expected task %q to appear verbatim in output", task)
	}
}