package main

import (
	"bytes"
	"strings"
	"testing"
)

// reset puts the flag-backed globals into their default state so each case
// starts from a known configuration.
func reset() {
	perLine = false
	leftOnly = false
	rightOnly = false
	noNewline = false
}

func TestTrimAll(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"spaces both sides", "   hello   ", "hello\n"},
		{"newlines and tabs", "\n\t\r\n  hello world \t\n\n", "hello world\n"},
		{"inner whitespace kept", "  a  b\tc  ", "a  b\tc\n"},
		{"multiline keeps inner lines", "\n  first\n  second  \n\n", "first\n  second\n"},
		{"already trimmed", "hello", "hello\n"},
		{"only whitespace", " \t\n ", "\n"},
		{"empty input", "", "\n"},
		{"unicode whitespace", "  hi  ", "hi\n"},
		{"unicode content preserved", "  héllo wörld — ok  ", "héllo wörld — ok\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reset()
			var out bytes.Buffer
			if err := run(strings.NewReader(tt.input), &out); err != nil {
				t.Fatalf("run() error: %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("run(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNoNewline(t *testing.T) {
	reset()
	noNewline = true

	var out bytes.Buffer
	if err := run(strings.NewReader("  hello  \n"), &out); err != nil {
		t.Fatalf("run() error: %v", err)
	}
	if got := out.String(); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestLeftRight(t *testing.T) {
	tests := []struct {
		name        string
		left, right bool
		want        string
	}{
		{"left only", true, false, "hello  \n"},
		{"right only", false, true, "  hello\n"},
		{"both flags means both sides", true, true, "hello\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reset()
			leftOnly, rightOnly = tt.left, tt.right

			var out bytes.Buffer
			if err := run(strings.NewReader("  hello  "), &out); err != nil {
				t.Fatalf("run() error: %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrimLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"every line trimmed", "  a  \n\tb\t\n  c", "a\nb\nc"},
		{"trailing newline preserved", "  a  \n  b  \n", "a\nb\n"},
		{"blank lines preserved", "a\n   \n\nb\n", "a\n\n\nb\n"},
		{"crlf stripped", "  a  \r\n  b  \r\n", "a\nb\n"},
		{"single line without newline", "   a   ", "a"},
		{"empty input", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reset()
			perLine = true

			var out bytes.Buffer
			if err := run(strings.NewReader(tt.input), &out); err != nil {
				t.Fatalf("run() error: %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("run(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrimLinesNoNewline(t *testing.T) {
	reset()
	perLine = true
	noNewline = true

	var out bytes.Buffer
	if err := run(strings.NewReader("  a  \n  b  \n"), &out); err != nil {
		t.Fatalf("run() error: %v", err)
	}
	if got := out.String(); got != "a\nb" {
		t.Errorf("got %q, want %q", got, "a\nb")
	}
}

// A line longer than bufio's default buffer must survive intact — this is why
// the implementation uses bufio.Reader instead of bufio.Scanner.
func TestTrimLinesLongLine(t *testing.T) {
	reset()
	perLine = true

	long := strings.Repeat("x", 128*1024)
	var out bytes.Buffer
	if err := run(strings.NewReader("   "+long+"   \n"), &out); err != nil {
		t.Fatalf("run() error: %v", err)
	}
	if got := out.String(); got != long+"\n" {
		t.Errorf("long line mangled: got %d bytes, want %d", len(got), len(long)+1)
	}
}
