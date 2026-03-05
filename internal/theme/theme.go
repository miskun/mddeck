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

	// Heading styles (content headings — used in body regions)
	H1Style string
	H2Style string
	H3Style string

	// Title styles (slide titles, distinct from content headings)
	TitleStyle      string // main title on title/section centered slides (fallback: H1Style)
	SlideTitleStyle string // title row heading on grid layouts (fallback: H2Style)

	// Heading/title margin-bottom (blank lines after element, default 1)
	TitleMargin      int
	SlideTitleMargin int
	H1Margin         int
	H2Margin         int
	H3Margin         int
	H4Margin         int
	H5Margin         int
	H6Margin         int

	// Alert / callout colors
	AlertNote      string // NOTE callout color
	AlertTip       string // TIP callout color
	AlertImportant string // IMPORTANT callout color
	AlertWarning   string // WARNING callout color
	AlertCaution   string // CAUTION callout color

	// Padding / chrome
	// PadBg paints the area outside the slide content stage (the centering
	// margins from aspect ratio). Set to an ANSI background escape like
	// ansi.BgRGB(28, 28, 34) to make the slide boundary visible.
	// Empty (default) = transparent, uses the terminal's own background.
	PadBg string
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
	H1Style:        ansi.Bold + ansi.FgRGB(0, 187, 187),    // accent (content heading)
	H2Style:        ansi.Bold + ansi.FgRGB(0, 187, 187),    // accent (content heading)
	H3Style:        ansi.Bold + ansi.FgRGB(0, 187, 187),    // accent (content heading)
	TitleStyle:      ansi.Bold + ansi.FgRGB(0, 187, 187),    // cyan — main title
	SlideTitleStyle: ansi.Bold + ansi.FgRGB(85, 255, 255),   // bright cyan — grid title row
	TitleMargin:      1,
	SlideTitleMargin: 1,
	H1Margin:         1,
	H2Margin:         1,
	H3Margin:         1,
	H4Margin:         1,
	H5Margin:         1,
	H6Margin:         1,
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
	H1Style:        ansi.Bold + ansi.FgRGB(61, 144, 206),   // accent (content heading)
	H2Style:        ansi.Bold + ansi.FgRGB(61, 144, 206),   // accent (content heading)
	H3Style:        ansi.Bold + ansi.FgRGB(61, 144, 206),   // accent (content heading)
	TitleStyle:      ansi.Bold + ansi.FgRGB(61, 144, 206),   // steel blue — main title
	SlideTitleStyle: ansi.Bold + ansi.FgRGB(120, 186, 240),  // bright steel blue — grid title row
	TitleMargin:      1,
	SlideTitleMargin: 1,
	H1Margin:         1,
	H2Margin:         1,
	H3Margin:         1,
	H4Margin:         1,
	H5Margin:         1,
	H6Margin:         1,
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
	H1Style:        ansi.Bold + ansi.FgRGB(4, 81, 165),     // accent (content heading)
	H2Style:        ansi.Bold + ansi.FgRGB(4, 81, 165),     // accent (content heading)
	H3Style:        ansi.Bold + ansi.FgRGB(4, 81, 165),     // accent (content heading)
	TitleStyle:      ansi.Bold + ansi.FgRGB(4, 81, 165),     // blue — main title
	SlideTitleStyle: ansi.Bold + ansi.FgRGB(0, 187, 187),    // cyan — grid title row
	TitleMargin:      1,
	SlideTitleMargin: 1,
	H1Margin:         1,
	H2Margin:         1,
	H3Margin:         1,
	H4Margin:         1,
	H5Margin:         1,
	H6Margin:         1,
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

// GetTitleStyle returns TitleStyle, falling back to H1Style.
func (t Theme) GetTitleStyle() string {
	if t.TitleStyle != "" {
		return t.TitleStyle
	}
	return t.H1Style
}

// GetSlideTitleStyle returns SlideTitleStyle, falling back to H2Style.
func (t Theme) GetSlideTitleStyle() string {
	if t.SlideTitleStyle != "" {
		return t.SlideTitleStyle
	}
	return t.H2Style
}

// GetHeadingMargin returns the margin-bottom (blank lines after) for a content
// heading at the given level. Built-in themes set all margins to 1.
func (t Theme) GetHeadingMargin(level int) int {
	switch level {
	case 1:
		return t.H1Margin
	case 2:
		return t.H2Margin
	case 3:
		return t.H3Margin
	case 4:
		return t.H4Margin
	case 5:
		return t.H5Margin
	case 6:
		return t.H6Margin
	default:
		return 1
	}
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
