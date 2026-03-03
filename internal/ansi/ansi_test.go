package ansi

import (
	"testing"
)

func TestStripUnsafe(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "SGR preserved",
			input: "\x1b[32mgreen\x1b[0m",
			want:  "\x1b[32mgreen\x1b[0m",
		},
		{
			name:  "cursor movement stripped",
			input: "\x1b[2Jhello",
			want:  "hello",
		},
		{
			name:  "mixed SGR and non-SGR",
			input: "\x1b[1mbold\x1b[0m\x1b[2Jcleared",
			want:  "\x1b[1mbold\x1b[0mcleared",
		},
		{
			name:  "plain text unchanged",
			input: "hello world",
			want:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripUnsafe(tt.input)
			if got != tt.want {
				t.Errorf("StripUnsafe(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripAll(t *testing.T) {
	input := "\x1b[32mgreen\x1b[0m text"
	want := "green text"
	got := StripAll(input)
	if got != want {
		t.Errorf("StripAll(%q) = %q, want %q", input, got, want)
	}
}

func TestVisibleLen(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"\x1b[32mhello\x1b[0m", 5},
		{"", 0},
		{"\x1b[1m\x1b[32mbold green\x1b[0m", 10},
	}

	for _, tt := range tests {
		got := VisibleLen(tt.input)
		if got != tt.want {
			t.Errorf("VisibleLen(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseEscapes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`\033[32m`, "\x1b[32m"},
		{`\e[32m`, "\x1b[32m"},
		{`\x1b[32m`, "\x1b[32m"},
		{"no escapes", "no escapes"},
	}

	for _, tt := range tests {
		got := ParseEscapes(tt.input)
		if got != tt.want {
			t.Errorf("ParseEscapes(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
