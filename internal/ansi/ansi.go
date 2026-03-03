// Package ansi provides ANSI escape sequence handling and safety filtering.
package ansi

import (
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

// FgRGB returns a 24-bit true-color foreground escape.
func FgRGB(r, g, b int) string {
	return "\x1b[38;2;" + itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"
}

// BgRGB returns a 24-bit true-color background escape.
func BgRGB(r, g, b int) string {
	return "\x1b[48;2;" + itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"
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
