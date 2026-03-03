// Package theme defines color themes for mddeck rendering.
package theme

import (
	"github.com/miskun/mddeck/internal/ansi"
)

// Theme defines the visual styling for a deck.
type Theme struct {
	Name string

	// Base colors
	Fg string // foreground ANSI escape
	Bg string // background ANSI escape (empty for default)

	// Semantic colors
	Accent    string // accent color (headings, links, bullets)
	BoldFg    string // bold text foreground (falls back to Accent if empty)
	Muted     string // muted color (blockquotes, notes)
	CodeFg    string // code foreground
	CodeBg    string // code background indicator
	ErrorFg   string // error foreground

	// Heading styles
	H1Style string
	H2Style string
	H3Style string

	// UI elements
	SlideNumStyle  string
	NotesStyle     string
	TimerStyle     string
	HelpStyle      string
	HRStyle        string
	BlockquoteChar string
	BulletChar     string
}

// Default is the built-in default theme.
var Default = Theme{
	Name:           "default",
	Fg:             ansi.FgWhite,
	Bg:             "",
	Accent:         ansi.FgCyan,
	BoldFg:         ansi.FgCyan,
	Muted:          ansi.FgBrightBlack,
	CodeFg:         ansi.FgGreen,
	CodeBg:         "",
	ErrorFg:        ansi.FgRed,
	H1Style:        ansi.Bold + ansi.FgCyan,
	H2Style:        ansi.Bold + ansi.FgBrightCyan,
	H3Style:        ansi.Bold + ansi.FgBrightWhite,
	SlideNumStyle:  ansi.Dim + ansi.FgBrightBlack,
	NotesStyle:     ansi.Italic + ansi.FgYellow,
	TimerStyle:     ansi.FgBrightBlack,
	HelpStyle:      ansi.Dim,
	HRStyle:        ansi.FgBrightBlack,
	BlockquoteChar: "│ ",
	BulletChar:     "• ",
}

// Dark is a dark theme variant.
// Body: #aaacae, Bold: #f1f3f5, Headings: #3d90ce.
var Dark = Theme{
	Name:           "dark",
	Fg:             ansi.FgRGB(170, 172, 174), // #aaacae
	Bg:             "",
	Accent:         ansi.FgRGB(61, 144, 206),  // #3d90ce
	BoldFg:         ansi.FgRGB(241, 243, 245), // #f1f3f5
	Muted:          ansi.FgBrightBlack,
	CodeFg:         ansi.FgBrightGreen,
	CodeBg:         "",
	ErrorFg:        ansi.FgBrightRed,
	H1Style:        ansi.Bold + ansi.FgRGB(61, 144, 206), // #3d90ce
	H2Style:        ansi.Bold + ansi.FgRGB(61, 144, 206), // #3d90ce
	H3Style:        ansi.Bold + ansi.FgRGB(61, 144, 206), // #3d90ce
	SlideNumStyle:  ansi.Dim + ansi.FgBrightBlack,
	NotesStyle:     ansi.Italic + ansi.FgBrightYellow,
	TimerStyle:     ansi.FgBrightBlack,
	HelpStyle:      ansi.Dim,
	HRStyle:        ansi.FgBrightBlack,
	BlockquoteChar: "│ ",
	BulletChar:     "• ",
}

// Light is a light-background theme.
var Light = Theme{
	Name:           "light",
	Fg:             ansi.FgBlack,
	Bg:             "",
	Accent:         ansi.FgBlue,
	BoldFg:         ansi.FgBlue,
	Muted:          ansi.FgBrightBlack,
	CodeFg:         ansi.FgRed,
	CodeBg:         "",
	ErrorFg:        ansi.FgRed,
	H1Style:        ansi.Bold + ansi.FgBlue,
	H2Style:        ansi.Bold + ansi.FgCyan,
	H3Style:        ansi.Bold + ansi.FgBlack,
	SlideNumStyle:  ansi.Dim + ansi.FgBrightBlack,
	NotesStyle:     ansi.Italic + ansi.FgMagenta,
	TimerStyle:     ansi.FgBrightBlack,
	HelpStyle:      ansi.Dim,
	HRStyle:        ansi.FgBrightBlack,
	BlockquoteChar: "│ ",
	BulletChar:     "• ",
}

// themes is the registry of available themes.
var themes = map[string]Theme{
	"default": Default,
	"dark":    Dark,
	"light":   Light,
}

// Get returns a theme by name, falling back to Default.
func Get(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return Default
}

// List returns all available theme names.
func List() []string {
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	return names
}
