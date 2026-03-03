package render

import (
	"sort"
	"strings"

	a "github.com/miskun/mddeck/internal/ansi"
)

// screenBuf is a virtual frame buffer for flicker-free diff-based rendering.
//
// Advanced TUIs (Bubble Tea, Textual, etc.) avoid flicker by:
//  1. Composing the full frame in memory
//  2. Comparing new lines against the previous frame
//  3. Only rewriting lines that actually changed
//
// This minimizes terminal I/O and eliminates intermediate blank states.
//
// Usage:
//
//	scr := newScreenBuf(width, height)
//	scr.Set(row, col, styledContent)  // 0-based coordinates
//	lines := scr.Lines()              // composed lines for diffing
type screenBuf struct {
	rows   [][]screenSegment
	width  int
	height int
}

// screenSegment represents a piece of styled content placed at a column.
type screenSegment struct {
	col     int
	content string
	visLen  int
}

// newScreenBuf creates a frame buffer for the given viewport dimensions.
func newScreenBuf(width, height int) *screenBuf {
	return &screenBuf{
		rows:   make([][]screenSegment, height),
		width:  width,
		height: height,
	}
}

// Set places ANSI-styled content at the given 0-based row and column.
// Multiple calls to the same row at different columns are composed correctly
// (e.g. multi-column layouts). Calls are ignored if row is out of bounds.
func (sb *screenBuf) Set(row, col int, content string) {
	if row < 0 || row >= sb.height || col < 0 {
		return
	}
	sb.rows[row] = append(sb.rows[row], screenSegment{
		col:     col,
		content: content,
		visLen:  a.VisibleLen(content),
	})
}

// Lines returns the composed rows as a slice of strings.
// Rows are NOT padded to viewport width — EraseLine in the render functions
// handles clearing remnants. This keeps output small and avoids writing
// characters past the actual terminal width during a resize race.
func (sb *screenBuf) Lines() []string {
	lines := make([]string, sb.height)
	for i := 0; i < sb.height; i++ {
		if len(sb.rows[i]) == 0 {
			lines[i] = ""
		} else {
			lines[i] = sb.composeRow(i)
		}
	}
	return lines
}

// RenderDiff produces output that only rewrites lines that differ between
// prev and next. Returns the minimal escape sequence to update the terminal.
func RenderDiff(prev, next []string, baseFg string, width int) string {
	var buf strings.Builder
	if baseFg != "" {
		buf.WriteString(baseFg)
	}
	for i := 0; i < len(next); i++ {
		if i >= len(prev) || prev[i] != next[i] {
			buf.WriteString(a.CursorTo(i+1, 1)) // 1-based row
			buf.WriteString(a.EraseLine)          // clear remnants
			buf.WriteString(next[i])
		}
	}
	buf.WriteString(a.Reset)
	return buf.String()
}

// RenderFull writes all lines sequentially (CursorHome + lines joined by \r\n).
// Used for first render or when the viewport size changes.
// Each line is preceded by EraseLine to prevent artifacts from resize race
// conditions (terminal may be a different size than when we measured).
func RenderFull(lines []string, baseFg string) string {
	var buf strings.Builder
	totalLen := 6 // CursorHome + Reset
	for _, l := range lines {
		totalLen += len(l) + len(a.EraseLine) + 2
	}
	buf.Grow(totalLen)
	buf.WriteString(a.CursorHome)
	if baseFg != "" {
		buf.WriteString(baseFg)
	}
	for i, line := range lines {
		buf.WriteString(a.EraseLine)
		buf.WriteString(line)
		if i < len(lines)-1 {
			buf.WriteString("\r\n")
		}
	}
	buf.WriteString(a.Reset)
	return buf.String()
}

// composeRow merges all segments on a row into a single string padded to width.
// Segments are sorted by column; gaps between segments are filled with spaces.
// Content is cropped to viewport width to prevent terminal auto-wrap.
func (sb *screenBuf) composeRow(rowIdx int) string {
	segs := sb.rows[rowIdx]

	// Stable sort by column so insertion order breaks ties (last wins for dedup).
	sort.SliceStable(segs, func(i, j int) bool {
		return segs[i].col < segs[j].col
	})

	var buf strings.Builder
	col := 0

	for _, seg := range segs {
		if col >= sb.width {
			break
		}
		if seg.col > col {
			// Fill gap, but don't exceed viewport width.
			gap := seg.col - col
			if gap > sb.width-col {
				gap = sb.width - col
			}
			buf.WriteString(strings.Repeat(" ", gap))
			col += gap
		}
		if col >= sb.width {
			break
		}
		// Crop segment content if it would overflow viewport.
		remaining := sb.width - col
		if seg.visLen > remaining {
			buf.WriteString(a.Truncate(seg.content, remaining))
			buf.WriteString(a.Reset) // prevent style bleeding
			col = sb.width
		} else {
			buf.WriteString(seg.content)
			col += seg.visLen
		}
	}

	// No trailing padding — EraseLine clears the rest of each row.

	return buf.String()
}
