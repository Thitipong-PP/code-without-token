package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func writeFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(body), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return full
}

// ──────────────────────────────────────────────
// ReadFiles
// ──────────────────────────────────────────────

func TestReadFiles_EmptyInputReturnsEmptySlice(t *testing.T) {
	got := ReadFiles([]string{})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestReadFiles_SingleFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "main.go", "package main\n")

	got := ReadFiles([]string{path})
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].Name != path {
		t.Errorf("Name: expected %q, got %q", path, got[0].Name)
	}
	if got[0].Content != "package main\n" {
		t.Errorf("Content: expected %q, got %q", "package main\n", got[0].Content)
	}
}

func TestReadFiles_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	p1 := writeFile(t, dir, "a.go", "aaa")
	p2 := writeFile(t, dir, "b.go", "bbb")
	p3 := writeFile(t, dir, "c.go", "ccc")

	got := ReadFiles([]string{p1, p2, p3})
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}
	contents := map[string]string{p1: "aaa", p2: "bbb", p3: "ccc"}
	for _, f := range got {
		want, ok := contents[f.Name]
		if !ok {
			t.Errorf("unexpected file name %q", f.Name)
		}
		if f.Content != want {
			t.Errorf("file %q: expected content %q, got %q", f.Name, want, f.Content)
		}
	}
}

func TestReadFiles_NonexistentFileSkipped(t *testing.T) {
	got := ReadFiles([]string{"/nonexistent/path/xyz.go"})
	if len(got) != 0 {
		t.Errorf("expected empty slice for unreadable file, got %v", got)
	}
}

func TestReadFiles_MixOfValidAndInvalidFiles(t *testing.T) {
	dir := t.TempDir()
	good := writeFile(t, dir, "good.go", "good content")

	got := ReadFiles([]string{good, "/nonexistent/bad.go"})
	if len(got) != 1 {
		t.Fatalf("expected 1 result (bad file skipped), got %d", len(got))
	}
	if got[0].Name != good {
		t.Errorf("expected %q, got %q", good, got[0].Name)
	}
}

func TestReadFiles_EmptyFileIncluded(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty.go", "")

	got := ReadFiles([]string{path})
	if len(got) != 1 {
		t.Fatalf("expected 1 result for empty file, got %d", len(got))
	}
	if got[0].Content != "" {
		t.Errorf("expected empty content, got %q", got[0].Content)
	}
}

func TestReadFiles_ContentMatchesFileBytes(t *testing.T) {
	dir := t.TempDir()
	body := "package main\n\nfunc main() {}\n"
	path := writeFile(t, dir, "main.go", body)

	got := ReadFiles([]string{path})
	if got[0].Content != body {
		t.Errorf("content mismatch:\nwant %q\ngot  %q", body, got[0].Content)
	}
}

func TestReadFiles_PreservesOrder(t *testing.T) {
	dir := t.TempDir()
	p1 := writeFile(t, dir, "first.go", "1")
	p2 := writeFile(t, dir, "second.go", "2")
	p3 := writeFile(t, dir, "third.go", "3")

	got := ReadFiles([]string{p1, p2, p3})
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}
	order := []string{p1, p2, p3}
	for i, p := range order {
		if got[i].Name != p {
			t.Errorf("position %d: expected %q, got %q", i, p, got[i].Name)
		}
	}
}

// ──────────────────────────────────────────────
// FormatFile
// ──────────────────────────────────────────────

func TestFormatFile_ContainsFileName(t *testing.T) {
	f := File{Name: "main.go", Content: "package main"}
	got := FormatFile(f)
	if !strings.Contains(got, "main.go") {
		t.Errorf("expected output to contain file name, got:\n%s", got)
	}
}

func TestFormatFile_ContainsContent(t *testing.T) {
	f := File{Name: "main.go", Content: "package main\nfunc main() {}"}
	got := FormatFile(f)
	if !strings.Contains(got, "package main") {
		t.Errorf("expected output to contain file content, got:\n%s", got)
	}
}

func TestFormatFile_HasCodeFences(t *testing.T) {
	f := File{Name: "main.go", Content: "x"}
	got := FormatFile(f)
	if !strings.Contains(got, "```") {
		t.Errorf("expected markdown code fences in output, got:\n%s", got)
	}
}

func TestFormatFile_HasSeparatorHeader(t *testing.T) {
	f := File{Name: "go.mod", Content: "module foo"}
	got := FormatFile(f)
	if !strings.Contains(got, "--- Content of go.mod ---") {
		t.Errorf("expected separator header, got:\n%s", got)
	}
}

func TestFormatFile_EmptyContentStillFormats(t *testing.T) {
	f := File{Name: "empty.go", Content: ""}
	got := FormatFile(f)
	if !strings.Contains(got, "empty.go") {
		t.Errorf("expected file name in output even for empty content, got:\n%s", got)
	}
	if !strings.Contains(got, "```") {
		t.Errorf("expected code fences even for empty content, got:\n%s", got)
	}
}

func TestFormatFile_OutputFormat(t *testing.T) {
	f := File{Name: "main.go", Content: "package main"}
	got := FormatFile(f)
	want := "\n\n--- Content of main.go ---\n```\npackage main\n```\n"
	if got != want {
		t.Errorf("format mismatch:\nwant %q\ngot  %q", want, got)
	}
}