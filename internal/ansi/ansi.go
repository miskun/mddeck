// Package ansi provides ANSI escape sequence handling and safety filtering.
package ansi

import (
	"os"
	"regexp"
	"strings"
)

// SGR (Select Graphic Rendition) regex matches valid color/style sequences.
var sgrRegex = regexp.MustCompile(`\x1b\[[\d;]*m`)

// allEscapeRegex matches any ANSI escape sequence.
var allEscapeRegex = regexp.MustCompile(`\x1b\[[^a-zA-Z]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b\][^\x1b]*\x1b\\`)

// StripUnsafe removes all non-SGR ANSI sequences from text, keeping only
// colors, bold, underline, reset, etc.
func StripUnsafe(text string) string {
	// First, find all SGR sequences and their positions
	sgrMatches := sgrRegex.FindAllStringIndex(text, -1)
	sgrSet := make(map[int]bool)
	for _, m := range sgrMatches {
		for i := m[0]; i < m[1]; i++ {
			sgrSet[i] = true
		}
	}

	// Find all escape sequences
	allMatches := allEscapeRegex.FindAllStringIndex(text, -1)

	if len(allMatches) == 0 {
		return text
	}

	var result strings.Builder
	result.Grow(len(text))
	lastEnd := 0

	for _, m := range allMatches {
		// Copy text before this escape
		result.WriteString(text[lastEnd:m[0]])

		// Check if this is an SGR sequence (keep it)
		if sgrSet[m[0]] {
			result.WriteString(text[m[0]:m[1]])
		}
		// Otherwise strip it (don't write)

		lastEnd = m[1]
	}

	// Copy remaining text
	result.WriteString(text[lastEnd:])

	return result.String()
}

// StripAll removes all ANSI escape sequences from text.
func StripAll(text string) string {
	return allEscapeRegex.ReplaceAllString(text, "")
}

// VisibleLen returns the visible character count (excluding ANSI sequences).
func VisibleLen(text string) int {
	clean := StripAll(text)
	count := 0
	for range clean {
		count++
	}
	return count
}

// Truncate truncates styled text to maxVis visible characters, preserving
// all ANSI escape sequences. Guarantees the result has at most maxVis visible
// characters, so rows never exceed viewport width and cause terminal auto-wrap.
func Truncate(s string, maxVis int) string {
	if maxVis <= 0 {
		return ""
	}
	var buf strings.Builder
	vis := 0
	inEsc := false

	for _, r := range s {
		if inEsc {
			buf.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			buf.WriteRune(r)
			inEsc = true
			continue
		}
		if vis >= maxVis {
			break
		}
		buf.WriteRune(r)
		vis++
	}

	return buf.String()
}

// TruncateEllipsis truncates an ANSI-styled string to maxVis visible
// characters and appends "…" if the string was truncated.
func TruncateEllipsis(s string, maxVis int) string {
	if maxVis <= 0 {
		return ""
	}
	vl := VisibleLen(s)
	if vl <= maxVis {
		return s
	}
	if maxVis <= 1 {
		return "…"
	}
	return Truncate(s, maxVis-1) + "…"
}

// Common ANSI escape codes for styling.
const (
	Reset         = "\x1b[0m"
	Bold          = "\x1b[1m"
	Dim           = "\x1b[2m"
	Italic        = "\x1b[3m"
	Underline     = "\x1b[4m"
	Reverse       = "\x1b[7m"
	Strikethrough = "\x1b[9m"

	// Foreground colors
	FgBlack   = "\x1b[30m"
	FgRed     = "\x1b[31m"
	FgGreen   = "\x1b[32m"
	FgYellow  = "\x1b[33m"
	FgBlue    = "\x1b[34m"
	FgMagenta = "\x1b[35m"
	FgCyan    = "\x1b[36m"
	FgWhite   = "\x1b[37m"

	// Bright foreground
	FgBrightBlack   = "\x1b[90m"
	FgBrightRed     = "\x1b[91m"
	FgBrightGreen   = "\x1b[92m"
	FgBrightYellow  = "\x1b[93m"
	FgBrightBlue    = "\x1b[94m"
	FgBrightMagenta = "\x1b[95m"
	FgBrightCyan    = "\x1b[96m"
	FgBrightWhite   = "\x1b[97m"

	// Background colors
	BgBlack   = "\x1b[40m"
	BgRed     = "\x1b[41m"
	BgGreen   = "\x1b[42m"
	BgYellow  = "\x1b[43m"
	BgBlue    = "\x1b[44m"
	BgMagenta = "\x1b[45m"
	BgCyan    = "\x1b[46m"
	BgWhite   = "\x1b[47m"

	// Cursor/screen control
	ClearScreen = "\x1b[2J"
	EraseLine   = "\x1b[2K"
	CursorHome  = "\x1b[H"
	HideCursor  = "\x1b[?25l"
	ShowCursor  = "\x1b[?25h"

	// Synchronized output (DEC mode 2026) — tells the terminal to
	// hold display updates until the full frame is received.
	BeginSync = "\x1b[?2026h"
	EndSync   = "\x1b[?2026l"
)

// Fg256 returns a 256-color foreground escape.
func Fg256(n int) string {
	return "\x1b[38;5;" + itoa(n) + "m"
}

// Bg256 returns a 256-color background escape.
func Bg256(n int) string {
	return "\x1b[48;5;" + itoa(n) + "m"
}

// FgRGB returns a foreground color escape for the given RGB values.
// If the terminal supports 24-bit true color, emits \x1b[38;2;r;g;bm.
// Otherwise, approximates to the nearest 256-color palette entry.
func FgRGB(r, g, b int) string {
	if hasTrueColor() {
		return "\x1b[38;2;" + itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"
	}
	return Fg256(rgbTo256(r, g, b))
}

// BgRGB returns a background color escape for the given RGB values.
// If the terminal supports 24-bit true color, emits \x1b[48;2;r;g;bm.
// Otherwise, approximates to the nearest 256-color palette entry.
func BgRGB(r, g, b int) string {
	if hasTrueColor() {
		return "\x1b[48;2;" + itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"
	}
	return Bg256(rgbTo256(r, g, b))
}

// trueColorSupport caches the result of true-color detection.
var trueColorSupport *bool

// hasTrueColor checks whether the terminal supports 24-bit true color.
// Checks COLORTERM for "truecolor" or "24bit".
func hasTrueColor() bool {
	if trueColorSupport != nil {
		return *trueColorSupport
	}
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	result := ct == "truecolor" || ct == "24bit"
	trueColorSupport = &result
	return result
}

// rgbTo256 converts an RGB color to the nearest xterm 256-color palette index.
func rgbTo256(r, g, b int) int {
	// Check if it's close to a greyscale ramp entry (232-255)
	// 24 shades from grey8 (#080808) to grey93 (#eeeeee)
	if r == g && g == b {
		if r < 8 {
			return 16 // black
		}
		if r > 238 {
			return 231 // white
		}
		return 232 + (r-8)*24/230
	}

	// Map to the 6x6x6 color cube (indices 16-231)
	ri := colorCubeIndex(r)
	gi := colorCubeIndex(g)
	bi := colorCubeIndex(b)
	cubeIdx := 16 + 36*ri + 6*gi + bi

	// Also check the nearest greyscale entry
	grey := (r + g + b) / 3
	var greyIdx int
	if grey < 8 {
		greyIdx = 16
	} else if grey > 238 {
		greyIdx = 231
	} else {
		greyIdx = 232 + (grey-8)*24/230
	}

	// Compare which is closer: cube or greyscale
	cubeDist := colorDist(r, g, b, cubeIdx)
	greyDist := colorDist(r, g, b, greyIdx)

	if greyDist < cubeDist {
		return greyIdx
	}
	return cubeIdx
}

// colorCubeIndex maps a 0-255 channel value to a 6-level cube index.
func colorCubeIndex(v int) int {
	// The 6 cube levels are: 0, 95, 135, 175, 215, 255
	if v < 48 {
		return 0
	}
	if v < 115 {
		return 1
	}
	return (v-35)/40 // maps 115→2, 155→3, 195→4, 235→5
}

// cubeLevels are the RGB values for each of the 6 color cube indices.
var cubeLevels = [6]int{0, 95, 135, 175, 215, 255}

// colorDist returns the squared distance between an RGB color and a 256-palette entry.
func colorDist(r, g, b, idx int) int {
	var pr, pg, pb int
	if idx < 16 {
		// Standard colors — rough approximation
		return 1<<30 - 1 // large distance, prefer cube/grey
	} else if idx < 232 {
		ci := idx - 16
		pr = cubeLevels[ci/36]
		pg = cubeLevels[(ci%36)/6]
		pb = cubeLevels[ci%6]
	} else {
		// Greyscale ramp: 232-255
		g := 8 + (idx-232)*10
		pr, pg, pb = g, g, g
	}
	dr := r - pr
	dg := g - pg
	db := b - pb
	return dr*dr + dg*dg + db*db
}

// CursorTo moves cursor to given row, col (1-based).
func CursorTo(row, col int) string {
	return "\x1b[" + itoa(row) + ";" + itoa(col) + "H"
}

func itoa(n int) string {
	if n < 0 {
		n = 0
	}
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// ProcessArtBlock processes an art block's content, applying safety filtering if needed.
func ProcessArtBlock(content string, safe bool) string {
	if safe {
		return StripUnsafe(content)
	}
	return content
}

// ParseEscapes converts literal \033 and \e to actual escape characters.
func ParseEscapes(s string) string {
	s = strings.ReplaceAll(s, "\\033", "\x1b")
	s = strings.ReplaceAll(s, "\\e", "\x1b")
	s = strings.ReplaceAll(s, "\\x1b", "\x1b")
	s = strings.ReplaceAll(s, "\\x1B", "\x1b")
	return s
}
