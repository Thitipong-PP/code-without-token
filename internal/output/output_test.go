package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// ──────────────────────────────────────────────
// Helper: capture stdout
// ──────────────────────────────────────────────

// captureStdout redirects os.Stdout to a pipe, runs fn, and returns everything
// that was written to stdout during fn's execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return buf.String()
}

// captureStdin replaces os.Stdin with a reader that yields the given input,
// runs fn, then restores os.Stdin.
func captureStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	fmt.Fprint(w, input)
	w.Close()

	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = orig })
	fn()
}

// ──────────────────────────────────────────────
// Print
// ──────────────────────────────────────────────

func TestPrint_ContainsText(t *testing.T) {
	out := captureStdout(t, func() {
		Print("hello world")
	})
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected text in output, got:\n%s", out)
	}
}

func TestPrint_HasTopBorder(t *testing.T) {
	out := captureStdout(t, func() {
		Print("x")
	})
	if !strings.Contains(out, "================") {
		t.Errorf("expected border in output, got:\n%s", out)
	}
}

func TestPrint_EmptyStringStillPrintsBorders(t *testing.T) {
	out := captureStdout(t, func() {
		Print("")
	})
	if !strings.Contains(out, "================") {
		t.Errorf("expected borders even for empty text, got:\n%s", out)
	}
}

func TestPrint_TextAppearsVerbatim(t *testing.T) {
	text := "Here is the project structure:\n- .\n  - main.go\n"
	out := captureStdout(t, func() {
		Print(text)
	})
	if !strings.Contains(out, text) {
		t.Errorf("expected verbatim text in output, got:\n%s", out)
	}
}

// ──────────────────────────────────────────────
// CopyToClipboard — decline path
// ──────────────────────────────────────────────

// When the user answers "n", the function must return without copying.
// We can't assert clipboard state portably, but we can assert stdout behaviour.
func TestCopyToClipboard_DeclineWithN(t *testing.T) {
	out := captureStdout(t, func() {
		captureStdin(t, "n\n", func() {
			CopyToClipboard("some text")
		})
	})
	// Must NOT print the success message.
	if strings.Contains(out, "Successfully generated") {
		t.Error("should not print success message when user declines")
	}
}

func TestCopyToClipboard_DeclineWithNo(t *testing.T) {
	out := captureStdout(t, func() {
		captureStdin(t, "no\n", func() {
			CopyToClipboard("some text")
		})
	})
	if strings.Contains(out, "Successfully generated") {
		t.Error("should not print success message when user declines with 'no'")
	}
}

func TestCopyToClipboard_DeclineWithArbitraryInput(t *testing.T) {
	out := captureStdout(t, func() {
		captureStdin(t, "nope\n", func() {
			CopyToClipboard("some text")
		})
	})
	if strings.Contains(out, "Successfully generated") {
		t.Error("should not print success message for arbitrary non-yes input")
	}
}

// ──────────────────────────────────────────────
// CopyToClipboard — prompt output
// ──────────────────────────────────────────────

// The prompt asking the user must always be printed regardless of answer.
func TestCopyToClipboard_PrintsPrompt(t *testing.T) {
	out := captureStdout(t, func() {
		captureStdin(t, "n\n", func() {
			CopyToClipboard("text")
		})
	})
	if !strings.Contains(out, "copy the result to clipboard") {
		t.Errorf("expected clipboard prompt in output, got:\n%s", out)
	}
}