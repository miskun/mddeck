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

func TestRgbTo256(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b int
		want    int
	}{
		{"pure black", 0, 0, 0, 16},
		{"pure white", 255, 255, 255, 231},
		{"mid grey", 128, 128, 128, 244}, // greyscale ramp
		{"pure red", 255, 0, 0, 196},     // cube 5,0,0 → 16+180=196
		{"pure green", 0, 255, 0, 46},    // cube 0,5,0 → 16+30=46
		{"pure blue", 0, 0, 255, 21},     // cube 0,0,5 → 16+5=21
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rgbTo256(tt.r, tt.g, tt.b)
			if got != tt.want {
				t.Errorf("rgbTo256(%d,%d,%d) = %d, want %d", tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestFgRGBFallback(t *testing.T) {
	// Force no true-color support
	f := false
	trueColorSupport = &f
	defer func() { trueColorSupport = nil }()

	got := FgRGB(255, 0, 0)
	want := Fg256(196)
	if got != want {
		t.Errorf("FgRGB(255,0,0) with no truecolor = %q, want %q", got, want)
	}
}

func TestFgRGBTrueColor(t *testing.T) {
	// Force true-color support
	tr := true
	trueColorSupport = &tr
	defer func() { trueColorSupport = nil }()

	got := FgRGB(170, 172, 174)
	want := "\x1b[38;2;170;172;174m"
	if got != want {
		t.Errorf("FgRGB(170,172,174) with truecolor = %q, want %q", got, want)
	}
}
