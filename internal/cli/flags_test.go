package cli

import (
	"testing"
)

// ──────────────────────────────────────────────
// parseIncludes
// ──────────────────────────────────────────────

func TestParseIncludes_EmptyStringReturnsEmptySlice(t *testing.T) {
	got := parseIncludes("")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestParseIncludes_SingleFile(t *testing.T) {
	got := parseIncludes("main.go")
	if len(got) != 1 || got[0] != "main.go" {
		t.Errorf("expected [main.go], got %v", got)
	}
}

func TestParseIncludes_MultipleFiles(t *testing.T) {
	got := parseIncludes("main.go,go.mod,readme.md")
	want := []string{"main.go", "go.mod", "readme.md"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: expected %q, got %q", i, w, got[i])
		}
	}
}

func TestParseIncludes_WhitespaceAroundEntries(t *testing.T) {
	got := parseIncludes("  main.go , go.mod ,  readme.md  ")
	want := []string{"main.go", "go.mod", "readme.md"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: expected %q, got %q", i, w, got[i])
		}
	}
}

func TestParseIncludes_EmptySegmentsSkipped(t *testing.T) {
	// Double commas or trailing comma produce empty segments — they must be dropped.
	got := parseIncludes("main.go,,go.mod,")
	want := []string{"main.go", "go.mod"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: expected %q, got %q", i, w, got[i])
		}
	}
}

func TestParseIncludes_WhitespaceOnlySegmentsSkipped(t *testing.T) {
	got := parseIncludes("main.go,   ,go.mod")
	want := []string{"main.go", "go.mod"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParseIncludes_SingleFileWithSpaces(t *testing.T) {
	got := parseIncludes("  main.go  ")
	if len(got) != 1 || got[0] != "main.go" {
		t.Errorf("expected [main.go], got %v", got)
	}
}

func TestParseIncludes_PathsPreserved(t *testing.T) {
	got := parseIncludes("internal/builder/builder.go,internal/cli/flags.go")
	want := []string{"internal/builder/builder.go", "internal/cli/flags.go"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: expected %q, got %q", i, w, got[i])
		}
	}
}

func TestParseIncludes_OnlyCommasReturnsEmptySlice(t *testing.T) {
	got := parseIncludes(",,,")
	if len(got) != 0 {
		t.Errorf("expected empty slice for input ',,,', got %v", got)
	}
}