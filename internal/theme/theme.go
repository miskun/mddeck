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

	// Alert / callout colors
	AlertNote      string // NOTE callout color
	AlertTip       string // TIP callout color
	AlertImportant string // IMPORTANT callout color
	AlertWarning   string // WARNING callout color
	AlertCaution   string // CAUTION callout color

	// Padding / chrome
	PadBg string // background color for padding area (empty = terminal default)
	PadFg string // foreground color for padding area (empty = use Fg)

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
// Colors defined as 24-bit RGB; auto-falls back to 256-color when needed.
var Default = Theme{
	Name:           "default",
	Fg:             ansi.FgRGB(204, 204, 204),              // #cccccc white
	Bg:             "",
	Accent:         ansi.FgRGB(0, 187, 187),                // #00bbbb cyan
	BoldFg:         ansi.FgRGB(0, 187, 187),                // #00bbbb cyan
	Muted:          ansi.FgRGB(102, 102, 102),              // #666666 grey
	CodeFg:         ansi.FgRGB(0, 187, 0),                  // #00bb00 green
	CodeBg:         "",
	ErrorFg:        ansi.FgRGB(205, 49, 49),                // #cd3131 red
	H1Style:        ansi.Bold + ansi.FgRGB(0, 187, 187),    // cyan
	H2Style:        ansi.Bold + ansi.FgRGB(85, 255, 255),   // #55ffff bright cyan
	H3Style:        ansi.Bold + ansi.FgRGB(255, 255, 255),  // #ffffff white
	AlertNote:      ansi.FgRGB(68, 147, 248),               // #4493f8 blue
	AlertTip:       ansi.FgRGB(63, 185, 80),                // #3fb950 green
	AlertImportant: ansi.FgRGB(163, 113, 247),              // #a371f7 purple
	AlertWarning:   ansi.FgRGB(210, 153, 34),               // #d29922 amber
	AlertCaution:   ansi.FgRGB(248, 81, 73),                // #f85149 red
	SlideNumStyle:  ansi.Dim + ansi.FgRGB(102, 102, 102),   // grey
	NotesStyle:     ansi.Italic + ansi.FgRGB(187, 187, 0),  // #bbbb00 yellow
	TimerStyle:     ansi.FgRGB(102, 102, 102),              // grey
	HelpStyle:      ansi.Dim,
	HRStyle:        ansi.FgRGB(102, 102, 102),              // grey
	BlockquoteChar: "│ ",
	BulletChar:     "• ",
}

// Dark is a dark theme variant.
// Body: grey, Bold: near-white, Headings: steel blue.
// Colors defined as 24-bit RGB; auto-falls back to 256-color when needed.
var Dark = Theme{
	Name:           "dark",
	Fg:             ansi.FgRGB(170, 172, 174),              // #aaacae grey
	Bg:             "",
	Accent:         ansi.FgRGB(61, 144, 206),               // #3d90ce steel blue
	BoldFg:         ansi.FgRGB(241, 243, 245),              // #f1f3f5 near-white
	Muted:          ansi.FgRGB(102, 102, 102),              // #666666 grey
	CodeFg:         ansi.FgRGB(85, 255, 85),                // #55ff55 bright green
	CodeBg:         "",
	ErrorFg:        ansi.FgRGB(255, 85, 85),                // #ff5555 bright red
	H1Style:        ansi.Bold + ansi.FgRGB(61, 144, 206),   // steel blue
	H2Style:        ansi.Bold + ansi.FgRGB(61, 144, 206),   // steel blue
	H3Style:        ansi.Bold + ansi.FgRGB(61, 144, 206),   // steel blue
	AlertNote:      ansi.FgRGB(68, 147, 248),               // #4493f8 blue
	AlertTip:       ansi.FgRGB(63, 185, 80),                // #3fb950 green
	AlertImportant: ansi.FgRGB(163, 113, 247),              // #a371f7 purple
	AlertWarning:   ansi.FgRGB(210, 153, 34),               // #d29922 amber
	AlertCaution:   ansi.FgRGB(248, 81, 73),                // #f85149 red
	SlideNumStyle:  ansi.Dim + ansi.FgRGB(102, 102, 102),   // grey
	NotesStyle:     ansi.Italic + ansi.FgRGB(255, 255, 85), // #ffff55 bright yellow
	TimerStyle:     ansi.FgRGB(102, 102, 102),              // grey
	HelpStyle:      ansi.Dim,
	HRStyle:        ansi.FgRGB(102, 102, 102),              // grey
	BlockquoteChar: "│ ",
	BulletChar:     "• ",
}

// Light is a light-background theme.
// Colors defined as 24-bit RGB; auto-falls back to 256-color when needed.
var Light = Theme{
	Name:           "light",
	Fg:             ansi.FgRGB(30, 30, 30),                 // #1e1e1e near-black
	Bg:             "",
	Accent:         ansi.FgRGB(4, 81, 165),                 // #0451a5 blue
	BoldFg:         ansi.FgRGB(4, 81, 165),                 // #0451a5 blue
	Muted:          ansi.FgRGB(102, 102, 102),              // #666666 grey
	CodeFg:         ansi.FgRGB(163, 21, 21),                // #a31515 red
	CodeBg:         "",
	ErrorFg:        ansi.FgRGB(205, 49, 49),                // #cd3131 red
	H1Style:        ansi.Bold + ansi.FgRGB(4, 81, 165),     // blue
	H2Style:        ansi.Bold + ansi.FgRGB(0, 187, 187),    // #00bbbb cyan
	H3Style:        ansi.Bold + ansi.FgRGB(30, 30, 30),     // near-black
	AlertNote:      ansi.FgRGB(9, 105, 218),                // #0969da blue
	AlertTip:       ansi.FgRGB(26, 127, 55),                // #1a7f37 green
	AlertImportant: ansi.FgRGB(130, 80, 223),               // #8250df purple
	AlertWarning:   ansi.FgRGB(154, 103, 0),                // #9a6700 amber
	AlertCaution:   ansi.FgRGB(207, 34, 46),                // #cf222e red
	SlideNumStyle:  ansi.Dim + ansi.FgRGB(102, 102, 102),   // grey
	NotesStyle:     ansi.Italic + ansi.FgRGB(175, 0, 219),  // #af00db magenta
	TimerStyle:     ansi.FgRGB(102, 102, 102),              // grey
	HelpStyle:      ansi.Dim,
	HRStyle:        ansi.FgRGB(102, 102, 102),              // grey
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
