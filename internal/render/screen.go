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
	// Padding background: only applied outside the content stage bounds.
	padBg  string // background color escape (empty = none)
	stageL int    // stage left column (cols before this = padding)
	stageR int    // stage right column (cols at/after this = padding)
	stageT int    // stage top row (rows before this = padding)
	stageB int    // stage bottom row (rows at/after this = padding)
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
		stageR: width,  // default: full viewport (no padding)
		stageB: height,
	}
}

// SetPadding configures the padding background and content stage bounds.
// Areas outside [stageL, stageT, stageR, stageB) are filled with padBg.
func (sb *screenBuf) SetPadding(padBg string, stageL, stageT, stageR, stageB int) {
	sb.padBg = padBg
	sb.stageL = stageL
	sb.stageT = stageT
	sb.stageR = stageR
	sb.stageB = stageB
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
// When padBg is set, padding areas are filled with the background color.
// Content area gaps (gutters, etc.) use the terminal default background.
func (sb *screenBuf) Lines() []string {
	lines := make([]string, sb.height)
	for i := 0; i < sb.height; i++ {
		if len(sb.rows[i]) == 0 {
			lines[i] = sb.emptyRow(i)
		} else {
			lines[i] = sb.composeRow(i)
		}
	}
	return lines
}

// emptyRow returns the content for a row with no segments.
func (sb *screenBuf) emptyRow(rowIdx int) string {
	if sb.padBg == "" {
		return ""
	}
	inStage := rowIdx >= sb.stageT && rowIdx < sb.stageB
	if !inStage {
		// Full padding row
		return sb.padBg + strings.Repeat(" ", sb.width) + a.Reset
	}
	// Content row with no segments: left pad + empty content + right pad
	var buf strings.Builder
	sb.writeGap(&buf, 0, sb.width, true)
	return buf.String()
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

// composeRow merges all segments on a row into a single string.
// Segments are sorted by column; gaps are filled with spaces.
// When padBg is set, gaps in the padding zone use padBg; gaps within
// the content stage use terminal default background.
func (sb *screenBuf) composeRow(rowIdx int) string {
	segs := sb.rows[rowIdx]

	// Stable sort by column so insertion order breaks ties (last wins for dedup).
	sort.SliceStable(segs, func(i, j int) bool {
		return segs[i].col < segs[j].col
	})

	var buf strings.Builder
	col := 0
	hasPadBg := sb.padBg != ""
	inStage := rowIdx >= sb.stageT && rowIdx < sb.stageB

	for _, seg := range segs {
		if col >= sb.width {
			break
		}
		if seg.col > col {
			// Fill gap, but don't exceed viewport width.
			end := seg.col
			if end > sb.width {
				end = sb.width
			}
			sb.writeGap(&buf, col, end, inStage)
			col = end
		}
		if col >= sb.width {
			break
		}
		// Crop segment content if it would overflow viewport.
		remaining := sb.width - col
		if seg.visLen > remaining {
			buf.WriteString(a.Truncate(seg.content, remaining))
			buf.WriteString(a.Reset)
			col = sb.width
		} else {
			buf.WriteString(seg.content)
			col += seg.visLen
		}
	}

	// Fill trailing space
	if hasPadBg && col < sb.width {
		sb.writeGap(&buf, col, sb.width, inStage)
	}

	return buf.String()
}

// writeGap fills columns [from, to) with the appropriate background.
// On content rows (inStage=true), padding zones (left/right margins) use padBg
// while the content area uses plain spaces. On padding rows, everything uses padBg.
func (sb *screenBuf) writeGap(buf *strings.Builder, from, to int, inStage bool) {
	if from >= to {
		return
	}
	if sb.padBg == "" {
		buf.WriteString(strings.Repeat(" ", to-from))
		return
	}
	if !inStage {
		// Entire row is padding — fill with padBg
		buf.WriteString(sb.padBg)
		buf.WriteString(strings.Repeat(" ", to-from))
		buf.WriteString(a.Reset)
		return
	}

	// Content row: split into up to 3 zones (left pad, content, right pad)
	cur := from

	// Left padding zone (cur < stageL)
	if cur < sb.stageL && cur < to {
		end := to
		if end > sb.stageL {
			end = sb.stageL
		}
		buf.WriteString(sb.padBg)
		buf.WriteString(strings.Repeat(" ", end-cur))
		buf.WriteString(a.Reset)
		cur = end
	}

	// Content zone (stageL <= cur < stageR)
	if cur >= sb.stageL && cur < sb.stageR && cur < to {
		end := to
		if end > sb.stageR {
			end = sb.stageR
		}
		buf.WriteString(strings.Repeat(" ", end-cur))
		cur = end
	}

	// Right padding zone (cur >= stageR)
	if cur >= sb.stageR && cur < to {
		buf.WriteString(sb.padBg)
		buf.WriteString(strings.Repeat(" ", to-cur))
		buf.WriteString(a.Reset)
	}
}
