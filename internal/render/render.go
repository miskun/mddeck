// Package render converts parsed markdown blocks into styled terminal output.
package render

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	a "github.com/miskun/mddeck/internal/ansi"
	"github.com/miskun/mddeck/internal/layout"
	"github.com/miskun/mddeck/internal/model"
	"github.com/miskun/mddeck/internal/theme"
)

// Renderer renders slides to terminal output.
type Renderer struct {
	Theme   theme.Theme
	Deck    *model.Deck
	Wrap    bool
	TabSize int
	SafeANSI bool
}

// NewRenderer creates a new renderer for a deck.
func NewRenderer(deck *model.Deck, th theme.Theme) *Renderer {
	return &Renderer{
		Theme:    th,
		Deck:     deck,
		Wrap:     deck.Meta.GetWrap(),
		TabSize:  deck.Meta.GetTabSize(),
		SafeANSI: deck.Meta.GetSafeAnsi(),
	}
}

// filterBlocksByStep returns only blocks whose Step <= the given step.
func filterBlocksByStep(blocks []model.Block, step int) []model.Block {
	var result []model.Block
	for _, b := range blocks {
		if b.Step <= step {
			result = append(result, b)
		}
	}
	return result
}

// RenderSlide renders a single slide and returns composed lines for diff-based output.
func (r *Renderer) RenderSlide(slide *model.Slide, vp layout.Viewport, step int) []string {
	// Filter blocks to only those visible at the current step
	visibleBlocks := filterBlocksByStep(slide.Blocks, step)

	// Create a copy of the slide with filtered blocks for layout/rendering
	visSlide := &model.Slide{
		Meta:   slide.Meta,
		Blocks: visibleBlocks,
		Notes:  slide.Notes,
		Index:  slide.Index,
		Steps:  slide.Steps,
	}

	lr := layout.ComputeLayout(visSlide, vp, &r.Deck.Meta)
	scr := newScreenBuf(vp.Width, vp.Height)

	// Configure padding background with content stage bounds
	if r.Theme.PadBg != "" && len(lr.Regions) > 0 {
		stageL := lr.Regions[0].X
		stageT := lr.Regions[0].Y
		stageR := lr.Regions[0].X + lr.Regions[0].Width
		stageB := lr.Regions[0].Y + lr.Regions[0].Height
		for _, reg := range lr.Regions[1:] {
			if reg.X < stageL {
				stageL = reg.X
			}
			if reg.Y < stageT {
				stageT = reg.Y
			}
			if reg.X+reg.Width > stageR {
				stageR = reg.X + reg.Width
			}
			if reg.Y+reg.Height > stageB {
				stageB = reg.Y + reg.Height
			}
		}
		scr.SetPadding(r.Theme.PadBg, stageL, stageT, stageR, stageB)
	}

	switch lr.Mode {
	case model.LayoutTitle:
		r.renderTitle(visSlide, lr.Regions[0], scr)
	case model.LayoutCenter:
		r.renderCentered(visSlide, lr.Regions[0], scr)
	default:
		// All layouts (built-in and custom) use the same grid renderer.
		// Single region → render directly. Multiple regions → distribute blocks.
		if len(lr.Regions) > 1 {
			r.renderGrid(visSlide, lr, scr)
		} else if len(lr.Regions) == 1 {
			r.renderSingleRegion(visSlide.Blocks, lr.Regions[0], scr)
		}
	}

	// Footer bar: left | center | right across the bottom row
	total := len(r.Deck.Slides)
	footer := r.Deck.Meta.Footer

	// Right section: custom text or default slide counter
	rightText := footer.Right
	if rightText == "" {
		if slide.Steps > 0 {
			rightText = fmt.Sprintf(" %d / %d [%d/%d] ", slide.Index+1, total, step+1, slide.Steps+1)
		} else {
			rightText = fmt.Sprintf(" %d / %d ", slide.Index+1, total)
		}
	}

	// Left section
	leftText := footer.Left

	// Center section
	centerText := footer.Center

	style := r.Theme.SlideNumStyle

	// Place left-aligned text
	if leftText != "" {
		padded := " " + leftText + " "
		scr.Set(vp.Height-1, 0, style+padded+a.Reset)
	}

	// Place center-aligned text
	if centerText != "" {
		padded := " " + centerText + " "
		centerCol := (vp.Width - len(padded)) / 2
		if centerCol < 0 {
			centerCol = 0
		}
		scr.Set(vp.Height-1, centerCol, style+padded+a.Reset)
	}

	// Place right-aligned text
	rightCol := vp.Width - len(rightText)
	if rightCol < 0 {
		rightCol = 0
	}
	scr.Set(vp.Height-1, rightCol, style+rightText+a.Reset)

	return scr.Lines()
}

// renderTitle renders a title-layout slide (centered vertically and horizontally).
func (r *Renderer) renderTitle(slide *model.Slide, region layout.Region, scr *screenBuf) {
	lines := r.renderBlocks(slide.Blocks, region.Width)

	// Center vertically
	startY := region.Y + (region.Height-len(lines))/2
	if startY < region.Y {
		startY = region.Y
	}

	for i, line := range lines {
		if i >= region.Height {
			break
		}
		row := startY + i
		// Crop to region width to prevent terminal auto-wrap
		line = r.cropLine(line, region.Width)
		// Center horizontally
		visLen := a.VisibleLen(line)
		padLeft := (region.Width - visLen) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		scr.Set(row, region.X+padLeft, line)
	}
}

// renderCentered renders a center-layout slide (centered vertically and horizontally).
func (r *Renderer) renderCentered(slide *model.Slide, region layout.Region, scr *screenBuf) {
	lines := r.renderBlocks(slide.Blocks, region.Width)

	// Center vertically
	startY := region.Y + (region.Height-len(lines))/2
	if startY < region.Y {
		startY = region.Y
	}

	for i, line := range lines {
		if i >= region.Height {
			break
		}
		row := startY + i
		line = r.cropLine(line, region.Width)
		// Center horizontally
		visLen := a.VisibleLen(line)
		padLeft := (region.Width - visLen) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		scr.Set(row, region.X+padLeft, line)
	}
}

// renderGrid renders blocks distributed across N regions in a grid layout.
// Major blocks (heading + its content) are distributed round-robin across regions.
func (r *Renderer) renderGrid(slide *model.Slide, lr layout.LayoutResult, scr *screenBuf) {
	majors := layout.SplitBlocksIntoMajor(slide.Blocks)
	nRegions := len(lr.Regions)

	// Distribute major blocks round-robin into region buckets
	regionBlocks := make([][]model.Block, nRegions)
	for i, maj := range majors {
		idx := i % nRegions
		if maj.Heading.Type != 0 {
			regionBlocks[idx] = append(regionBlocks[idx], maj.Heading)
		}
		regionBlocks[idx] = append(regionBlocks[idx], maj.Content...)
	}

	// If no major blocks detected, put everything in the first region
	if len(majors) == 0 {
		regionBlocks[0] = slide.Blocks
	}

	for i, blocks := range regionBlocks {
		if len(blocks) > 0 {
			r.renderInRegion(blocks, lr.Regions[i], scr)
		}
	}
}

// renderSingleRegion renders all blocks into a single region.
func (r *Renderer) renderSingleRegion(blocks []model.Block, region layout.Region, scr *screenBuf) {
	r.renderInRegion(blocks, region, scr)
}

// renderInRegion renders blocks within a specific region.
func (r *Renderer) renderInRegion(blocks []model.Block, region layout.Region, scr *screenBuf) {
	lines := r.renderBlocks(blocks, region.Width)
	startY := region.Y

	overflow := len(lines) > region.Height
	limit := region.Height
	if overflow {
		limit = region.Height - 1 // leave last row for overflow indicator
	}
	if limit > len(lines) {
		limit = len(lines)
	}

	for i := 0; i < limit; i++ {
		scr.Set(startY+i, region.X, r.cropLine(lines[i], region.Width))
	}

	if overflow {
		scr.Set(startY+region.Height-1, region.X, r.Theme.Muted+"↓"+a.Reset)
	}
}

// renderBlocks converts blocks into styled lines.
func (r *Renderer) renderBlocks(blocks []model.Block, width int) []string {
	var lines []string

	for i, block := range blocks {
		blockLines := r.renderBlock(block, width)
		lines = append(lines, blockLines...)

		// Add spacing between blocks, but not between consecutive
		// list blocks of the same type (e.g. incremental list items).
		if i < len(blocks)-1 {
			next := blocks[i+1]
			sameList := isListType(block.Type) && block.Type == next.Type
			if !sameList {
				lines = append(lines, "")
			}
		}
	}

	return lines
}

// renderBlock renders a single block into styled lines.
// isListType returns true for list block types.
func isListType(t model.BlockType) bool {
	return t == model.BlockUnorderedList || t == model.BlockOrderedList || t == model.BlockTaskList
}

func (r *Renderer) renderBlock(block model.Block, width int) []string {
	switch block.Type {
	case model.BlockHeading:
		return r.renderHeading(block, width)
	case model.BlockParagraph:
		return r.renderParagraph(block, width)
	case model.BlockUnorderedList:
		return r.renderUnorderedList(block, width)
	case model.BlockOrderedList:
		return r.renderOrderedList(block, width)
	case model.BlockBlockquote:
		return r.renderBlockquote(block, width)
	case model.BlockAlert:
		return r.renderAlert(block, width)
	case model.BlockTaskList:
		return r.renderTaskList(block, width)
	case model.BlockTable:
		return r.renderTable(block, width)
	case model.BlockFencedCode:
		return r.renderCode(block, width)
	case model.BlockANSIArt, model.BlockASCIIArt, model.BlockBrailleArt:
		return r.renderArtBlock(block)
	case model.BlockHorizontalRule:
		return r.renderHR(width)
	default:
		return []string{block.Raw}
	}
}

// renderHeading renders a heading with appropriate styling.
func (r *Renderer) renderHeading(block model.Block, width int) []string {
	var style string
	var prefix string

	switch block.Level {
	case 1:
		style = r.Theme.H1Style
		prefix = ""
	case 2:
		style = r.Theme.H2Style
		prefix = ""
	default:
		style = r.Theme.H3Style
		prefix = ""
	}

	text := style + prefix + block.Raw + a.Reset
	return []string{text}
}

// renderParagraph renders a paragraph with optional wrapping.
// Embedded newlines (from hard line breaks: trailing \ or two spaces) are preserved.
func (r *Renderer) renderParagraph(block model.Block, width int) []string {
	// Split on embedded newlines first (from hard line breaks)
	segments := strings.Split(block.Raw, "\n")
	var lines []string

	for _, seg := range segments {
		if r.Wrap && width > 0 {
			wrapped := wrapText(seg, width)
			for _, wl := range wrapped {
				lines = append(lines, r.applyInlineStyles(wl))
			}
		} else {
			lines = append(lines, r.applyInlineStyles(seg))
		}
	}
	return lines
}

// parseListItem extracts the depth and text from a depth-prefixed list item.
// Items are stored as "DEPTH:text" where DEPTH is a digit.
func parseListItem(item string) (int, string) {
	if len(item) >= 2 && item[0] >= '0' && item[0] <= '9' && item[1] == ':' {
		return int(item[0] - '0'), item[2:]
	}
	return 0, item
}

// renderUnorderedList renders an unordered list with nesting support.
func (r *Renderer) renderUnorderedList(block model.Block, width int) []string {
	var lines []string
	bullets := []string{"• ", "◦ ", "▪ "} // different bullets per depth

	for _, item := range block.Lines {
		depth, text := parseListItem(item)
		indent := strings.Repeat("  ", depth+1) // base indent + nesting
		bulletIdx := depth % len(bullets)
		bullet := r.Theme.Accent + bullets[bulletIdx] + a.Reset
		prefixWidth := 2*(depth+1) + 2 // indentation + bullet

		// Split on hard line breaks (embedded newlines from trailing backslash)
		subLines := strings.Split(text, "\n")
		for si, sub := range subLines {
			if r.Wrap && width > prefixWidth {
				wrapped := wrapText(sub, width-prefixWidth)
				for j, wl := range wrapped {
					styled := r.applyInlineStyles(wl)
					if si == 0 && j == 0 {
						lines = append(lines, indent+bullet+styled)
					} else {
						lines = append(lines, indent+"  "+styled)
					}
				}
			} else {
				styled := r.applyInlineStyles(sub)
				if si == 0 {
					lines = append(lines, indent+bullet+styled)
				} else {
					lines = append(lines, indent+"  "+styled)
				}
			}
		}
	}
	return lines
}

// renderOrderedList renders an ordered list with nesting support.
func (r *Renderer) renderOrderedList(block model.Block, width int) []string {
	var lines []string
	// Track per-depth counters for ordering
	counters := make(map[int]int)
	// If this block was split from a larger list, start at the right number
	if block.ListStart > 0 {
		counters[0] = block.ListStart - 1
	}

	for _, item := range block.Lines {
		depth, text := parseListItem(item)
		counters[depth]++
		// Reset deeper counters when we go back up
		for d := depth + 1; d <= 9; d++ {
			delete(counters, d)
		}

		indent := strings.Repeat("  ", depth+1)
		num := fmt.Sprintf("%s%d.%s ", r.Theme.Accent, counters[depth], a.Reset)
		prefix := indent + num
		prefixWidth := 2*(depth+1) + 3 // indentation + "N. "

		// Split on hard line breaks (embedded newlines from trailing backslash)
		subLines := strings.Split(text, "\n")
		for si, sub := range subLines {
			if r.Wrap && width > prefixWidth {
				wrapped := wrapText(sub, width-prefixWidth)
				for j, wl := range wrapped {
					styled := r.applyInlineStyles(wl)
					if si == 0 && j == 0 {
						lines = append(lines, prefix+styled)
					} else {
						lines = append(lines, strings.Repeat(" ", prefixWidth)+styled)
					}
				}
			} else {
				styled := r.applyInlineStyles(sub)
				if si == 0 {
					lines = append(lines, prefix+styled)
				} else {
					lines = append(lines, strings.Repeat(" ", prefixWidth)+styled)
				}
			}
		}
	}
	return lines
}

// renderBlockquote renders a blockquote.
func (r *Renderer) renderBlockquote(block model.Block, width int) []string {
	var lines []string
	indicator := r.Theme.Muted + r.Theme.BlockquoteChar
	// Account for indicator width when wrapping
	contentWidth := width - 2 // "│ " is 2 chars

	for _, line := range block.Lines {
		if r.Wrap && contentWidth > 0 {
			wrapped := wrapText(line, contentWidth)
			for _, wl := range wrapped {
				text := r.applyInlineStyles(wl)
				lines = append(lines, indicator+text+a.Reset)
			}
		} else {
			text := r.applyInlineStyles(line)
			lines = append(lines, indicator+text+a.Reset)
		}
	}
	return lines
}

// alertStyle returns the icon and color for an alert type.
func (r *Renderer) alertStyle(alertType string) (string, string) {
	switch alertType {
	case "NOTE":
		return "▪", r.Theme.AlertNote
	case "TIP":
		return "▪", r.Theme.AlertTip
	case "IMPORTANT":
		return "▪", r.Theme.AlertImportant
	case "WARNING":
		return "▪", r.Theme.AlertWarning
	case "CAUTION":
		return "▪", r.Theme.AlertCaution
	default:
		return "ℹ", r.Theme.Accent
	}
}

// renderAlert renders an alert/callout block with styled prefix.
func (r *Renderer) renderAlert(block model.Block, width int) []string {
	var lines []string
	icon, color := r.alertStyle(block.Language)

	// Title line
	title := color + a.Bold + icon + " " + block.Language + a.Reset
	bar := color + "│ " + a.Reset

	lines = append(lines, bar+title)

	// Account for bar width when wrapping
	contentWidth := width - 2 // "│ " is 2 chars

	for _, line := range block.Lines {
		if line == "" {
			lines = append(lines, bar)
		} else if r.Wrap && contentWidth > 0 {
			wrapped := wrapText(line, contentWidth)
			for _, wl := range wrapped {
				text := r.applyInlineStyles(wl)
				lines = append(lines, bar+text)
			}
		} else {
			text := r.applyInlineStyles(line)
			lines = append(lines, bar+text)
		}
	}
	return lines
}

// parseTaskItem extracts depth, checked state, and text from a task list item.
// Items are stored as "DEPTH:C:text" where C is 1 (checked) or 0 (unchecked).
func parseTaskItem(item string) (int, bool, string) {
	if len(item) >= 4 && item[0] >= '0' && item[0] <= '9' && item[1] == ':' && item[3] == ':' {
		depth := int(item[0] - '0')
		checked := item[2] == '1'
		return depth, checked, item[4:]
	}
	return 0, false, item
}

// renderTaskList renders a task list with checkboxes.
func (r *Renderer) renderTaskList(block model.Block, width int) []string {
	var lines []string

	for _, item := range block.Lines {
		depth, checked, text := parseTaskItem(item)
		indent := strings.Repeat("  ", depth+1)

		var checkbox string
		if checked {
			checkbox = r.Theme.Accent + "☑" + a.Reset + " "
		} else {
			checkbox = r.Theme.Muted + "☐" + a.Reset + " "
		}

		prefixWidth := 2*(depth+1) + 2 // indent + checkbox

		if r.Wrap && width > prefixWidth {
			wrapped := wrapText(text, width-prefixWidth)
			for j, wl := range wrapped {
				styled := r.applyInlineStyles(wl)
				if j == 0 {
					lines = append(lines, indent+checkbox+styled)
				} else {
					lines = append(lines, strings.Repeat(" ", prefixWidth)+styled)
				}
			}
		} else {
			lines = append(lines, indent+checkbox+r.applyInlineStyles(text))
		}
	}
	return lines
}

// splitTableRow splits a pipe-delimited table line into cells.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}
	cells := strings.Split(line, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

// renderTable renders a pipe-delimited table with box-drawing characters.
// Mid-table separator rows (preserved by the parser) are rendered as ├──┼──┤ borders.
func (r *Renderer) renderTable(block model.Block, width int) []string {
	if len(block.Lines) == 0 {
		return nil
	}

	// Classify each line as separator or data
	type entry struct {
		isSep bool
		cells []string
	}
	var entries []entry
	maxCols := 0
	for _, line := range block.Lines {
		if isRendererTableSeparator(line) {
			entries = append(entries, entry{isSep: true})
		} else {
			cells := splitTableRow(line)
			entries = append(entries, entry{cells: cells})
			if len(cells) > maxCols {
				maxCols = len(cells)
			}
		}
	}

	// Normalize: pad data rows with fewer cells
	for i := range entries {
		if !entries[i].isSep {
			for len(entries[i].cells) < maxCols {
				entries[i].cells = append(entries[i].cells, "")
			}
		}
	}

	// Calculate column widths based on visible content length (data rows only).
	// Inline markdown markers (**bold**, *italic*, etc.) are stripped
	// so that column widths reflect what the user actually sees.
	colWidths := make([]int, maxCols)
	for _, e := range entries {
		if e.isSep {
			continue
		}
		for c, cell := range e.cells {
			visible := stripInlineMarkdown(cell)
			cl := utf8.RuneCountInString(visible)
			if cl > colWidths[c] {
				colWidths[c] = cl
			}
		}
	}

	// Cap total width to available space
	totalWidth := maxCols + 1 // pipes
	for _, w := range colWidths {
		totalWidth += w + 2 // padding
	}
	if totalWidth > width && width > maxCols*3+maxCols+1 {
		// Shrink proportionally
		available := width - maxCols - 1 - maxCols*2
		if available < maxCols {
			available = maxCols
		}
		total := 0
		for _, w := range colWidths {
			total += w
		}
		if total > 0 {
			for c := range colWidths {
				colWidths[c] = colWidths[c] * available / total
				if colWidths[c] < 1 {
					colWidths[c] = 1
				}
			}
		}
	}

	var lines []string
	isFirstDataRow := true
	hasHeader := !block.NoHeader

	// Top border: ┌──┬──┐
	lines = append(lines, r.tableHLine("┌", "┬", "┐", colWidths, maxCols))

	for _, e := range entries {
		if e.isSep {
			// Mid-table separator: ├──┼──┤
			lines = append(lines, r.tableHLine("├", "┼", "┤", colWidths, maxCols))
			continue
		}

		// Data row: │ cell │ cell │
		line := r.Theme.Muted + "│" + a.Reset
		for c, cell := range e.cells {
			// Compute visible length and truncate if needed
			styled := r.applyInlineStyles(cell)
			visLen := a.VisibleLen(styled)
			if visLen > colWidths[c] {
				styled = a.TruncateEllipsis(styled, colWidths[c])
				visLen = a.VisibleLen(styled)
			}
			pad := strings.Repeat(" ", colWidths[c]-visLen)

			if isFirstDataRow && hasHeader {
				// Header row: bold
				line += " " + a.Bold + styled + a.Reset + pad + " " + r.Theme.Muted + "│" + a.Reset
			} else {
				line += " " + styled + pad + " " + r.Theme.Muted + "│" + a.Reset
			}
		}
		lines = append(lines, line)

		// After header row: separator ├──┼──┤
		if isFirstDataRow && hasHeader {
			lines = append(lines, r.tableHLine("├", "┼", "┤", colWidths, maxCols))
		}
		isFirstDataRow = false
	}

	// Bottom border: └──┴──┘
	lines = append(lines, r.tableHLine("└", "┴", "┘", colWidths, maxCols))

	return lines
}

// tableHLine builds a horizontal table border line: left + segments + right.
func (r *Renderer) tableHLine(left, mid, right string, colWidths []int, maxCols int) string {
	s := r.Theme.Muted + left
	for c, w := range colWidths {
		s += strings.Repeat("─", w+2)
		if c < maxCols-1 {
			s += mid
		}
	}
	s += right + a.Reset
	return s
}

// isRendererTableSeparator checks if a line is a table separator row (e.g., |---|---|).
func isRendererTableSeparator(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") {
		return false
	}
	inner := strings.Trim(line, "| ")
	for _, ch := range inner {
		if ch != '-' && ch != ':' && ch != '|' && ch != ' ' {
			return false
		}
	}
	return strings.Contains(inner, "-")
}

// renderCode renders a fenced code block.
func (r *Renderer) renderCode(block model.Block, width int) []string {
	var lines []string
	codeStyle := r.Theme.CodeFg

	for _, line := range block.Lines {
		// Expand tabs
		expanded := r.expandTabs(line)
		lines = append(lines, codeStyle+expanded+a.Reset)
	}
	return lines
}

// renderArtBlock renders an art block (ANSI, ASCII, or Braille).
func (r *Renderer) renderArtBlock(block model.Block) []string {
	content := block.Raw
	if block.Type == model.BlockANSIArt {
		content = a.ParseEscapes(content)
		if r.SafeANSI {
			content = a.StripUnsafe(content)
		}
	}

	return strings.Split(content, "\n")
}

// renderHR renders a horizontal rule.
func (r *Renderer) renderHR(width int) []string {
	hr := r.Theme.HRStyle + strings.Repeat("─", width) + a.Reset
	return []string{hr}
}

// Inline style patterns
var (
	boldRegex          = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRegex        = regexp.MustCompile(`\*(.+?)\*`)
	codeRegex          = regexp.MustCompile("`([^`]+)`")
	strikethroughRegex = regexp.MustCompile(`~~(.+?)~~`)
	linkRegex          = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
)

// stripInlineMarkdown removes markdown syntax markers to get the visible text length.
// Used for table column width calculation so padding is correct after styling.
func stripInlineMarkdown(s string) string {
	// Extract code spans first to protect their content from other transforms
	var codeContents []string
	s = codeRegex.ReplaceAllStringFunc(s, func(m string) string {
		inner := codeRegex.FindStringSubmatch(m)[1]
		idx := len(codeContents)
		codeContents = append(codeContents, inner)
		return fmt.Sprintf("\x00C%d\x00", idx)
	})

	s = boldRegex.ReplaceAllString(s, "$1")
	s = italicRegex.ReplaceAllString(s, "$1")
	s = strikethroughRegex.ReplaceAllString(s, "$1")
	s = linkRegex.ReplaceAllString(s, "$1")

	// Restore code contents (without backticks)
	for i, content := range codeContents {
		s = strings.Replace(s, fmt.Sprintf("\x00C%d\x00", i), content, 1)
	}
	return s
}

// applyInlineStyles applies bold, italic, code, and link styles.
// Code spans are extracted first so their content is protected from
// bold / italic / strikethrough transforms.
func (r *Renderer) applyInlineStyles(text string) string {
	// Extract code spans first to protect their content
	var codeSpans []string
	text = codeRegex.ReplaceAllStringFunc(text, func(m string) string {
		inner := codeRegex.FindStringSubmatch(m)[1]
		idx := len(codeSpans)
		codeSpans = append(codeSpans, r.Theme.CodeFg+inner+a.Reset+r.Theme.Fg)
		return fmt.Sprintf("\x00C%d\x00", idx)
	})

	// Bold
	boldColor := r.Theme.BoldFg
	if boldColor == "" {
		boldColor = r.Theme.Accent
	}
	text = boldRegex.ReplaceAllString(text, a.Bold+boldColor+"$1"+a.Reset+r.Theme.Fg)
	// Italic (must come after bold to avoid conflicts)
	text = italicRegex.ReplaceAllString(text, a.Italic+"$1"+a.Reset+r.Theme.Fg)
	// Strikethrough
	text = strikethroughRegex.ReplaceAllString(text, a.Strikethrough+"$1"+a.Reset+r.Theme.Fg)
	// Links (render as styled text)
	text = linkRegex.ReplaceAllString(text, a.Underline+r.Theme.Accent+"$1"+a.Reset+r.Theme.Fg)

	// Restore code spans with styling
	for i, span := range codeSpans {
		text = strings.Replace(text, fmt.Sprintf("\x00C%d\x00", i), span, 1)
	}

	return text
}

// wrapText wraps text to fit within the given width.
// After wrapping, split markdown spans are repaired: if a marker like ** was
// opened on one line but not closed, it is closed at the end and reopened
// on the next line so that applyInlineStyles works correctly on each line.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	if utf8.RuneCountInString(stripInlineMarkdown(text)) <= width {
		return []string{text}
	}

	// Simple word-wrap using visible length (excluding markdown markers)
	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder
	currentLen := 0

	for _, word := range words {
		wordLen := utf8.RuneCountInString(stripMarkdownMarkers(word))
		if currentLen > 0 && currentLen+1+wordLen > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLen = 0
		}
		if currentLen > 0 {
			currentLine.WriteByte(' ')
			currentLen++
		}
		currentLine.WriteString(word)
		currentLen += wordLen
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	// Fix markdown spans that were split across lines
	return fixSplitMarkdown(lines)
}

// fixSplitMarkdown closes and reopens inline markdown markers that were split
// across lines by wrapping. Processes ** before * to avoid ambiguity.
func fixSplitMarkdown(lines []string) []string {
	// Process markers in order: ** and ~~ first (2-char), then * and ` (1-char).
	// For *, subtract occurrences that are part of ** to get true single-star count.
	for _, marker := range []string{"**", "~~", "*", "`"} {
		carry := false
		for i := range lines {
			if carry {
				lines[i] = marker + lines[i]
			}
			n := strings.Count(lines[i], marker)
			// For single *, don't count those that are part of **
			if marker == "*" {
				nDouble := strings.Count(lines[i], "**")
				n = n - 2*nDouble
			}
			carry = (n%2 != 0)
			if carry {
				lines[i] = lines[i] + marker
			}
		}
	}
	return lines
}

// stripMarkdownMarkers removes *, ~, and ` characters used as markdown syntax.
// Unlike stripInlineMarkdown (regex-based, needs matched pairs), this works on
// individual words where markers may be split across words (e.g., "**bold" without
// a closing "**" in the same word).
func stripMarkdownMarkers(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '*' || r == '~' || r == '`' {
			return -1
		}
		return r
	}, s)
}

// padRight pads a rendered line with spaces so it fills the full width,
// preventing leftover characters from a previous wider render.
func (r *Renderer) padRight(line string, width int) string {
	vis := a.VisibleLen(line)
	if vis >= width {
		return line
	}
	return line + strings.Repeat(" ", width-vis)
}

// cropLine truncates a line if it exceeds the width, preserving ANSI styling.
func (r *Renderer) cropLine(line string, width int) string {
	visLen := a.VisibleLen(line)
	if visLen <= width {
		return line
	}
	if width <= 1 {
		return "…"
	}
	return a.Truncate(line, width-1) + "…" + a.Reset
}

// expandTabs replaces tabs with spaces.
func (r *Renderer) expandTabs(line string) string {
	return strings.ReplaceAll(line, "\t", strings.Repeat(" ", r.TabSize))
}

// RenderPresenter renders the presenter view and returns composed lines.
func (r *Renderer) RenderPresenter(slide *model.Slide, vp layout.Viewport, elapsed string, step int) []string {
	scr := newScreenBuf(vp.Width, vp.Height)

	// Filter blocks by current step
	visibleBlocks := filterBlocksByStep(slide.Blocks, step)

	// Layout: top 55% = current slide, bottom = next preview + notes
	topH := vp.Height * 55 / 100
	bottomH := vp.Height - topH - 2 // 2 lines for divider and status

	// Current slide region
	currentRegion := layout.Region{
		X: 1, Y: 0,
		Width: vp.Width - 2, Height: topH,
	}

	// Render current slide content
	lr := layout.ComputeLayout(slide, layout.Viewport{Width: currentRegion.Width, Height: currentRegion.Height}, &r.Deck.Meta)
	if len(lr.Regions) > 0 {
		renderedLines := r.renderBlocks(visibleBlocks, lr.Regions[0].Width)
		for i, line := range renderedLines {
			if i >= currentRegion.Height {
				break
			}
			scr.Set(currentRegion.Y+i, currentRegion.X, r.cropLine(line, currentRegion.Width))
		}
	}

	// Divider
	dividerRow := topH // 0-based
	scr.Set(dividerRow, 0, r.Theme.HRStyle+strings.Repeat("─", vp.Width)+a.Reset)

	// Bottom: left = next preview, right = notes
	notesWidth := vp.Width / 2
	previewWidth := vp.Width - notesWidth - 1

	// Next slide preview
	total := len(r.Deck.Slides)
	if slide.Index+1 < total {
		nextSlide := &r.Deck.Slides[slide.Index+1]
		scr.Set(dividerRow+1, 0, r.Theme.Muted+"Next:"+a.Reset)

		nextLines := r.renderBlocks(nextSlide.Blocks, previewWidth-2)
		for i, line := range nextLines {
			if i >= bottomH-1 {
				break
			}
			scr.Set(dividerRow+2+i, 1, r.Theme.Muted+r.cropLine(line, previewWidth-2)+a.Reset)
		}
	}

	// Notes
	if slide.Notes != "" {
		scr.Set(dividerRow+1, previewWidth+1, r.Theme.NotesStyle+"Notes:"+a.Reset)

		noteLines := strings.Split(slide.Notes, "\n")
		for i, line := range noteLines {
			if i >= bottomH-1 {
				break
			}
			scr.Set(dividerRow+2+i, previewWidth+1, r.Theme.NotesStyle+r.cropLine(line, notesWidth-1)+a.Reset)
		}
	}

	// Status bar into screen buffer
	statusLine := fmt.Sprintf(" %s  │  %d / %d ", elapsed, slide.Index+1, total)
	if slide.Steps > 0 {
		statusLine = fmt.Sprintf(" %s  │  %d / %d [%d/%d] ", elapsed, slide.Index+1, total, step+1, slide.Steps+1)
	}
	statusCol := vp.Width - utf8.RuneCountInString(statusLine)
	if statusCol < 0 {
		statusCol = 0
	}
	scr.Set(vp.Height-1, statusCol, r.Theme.TimerStyle+statusLine+a.Reset)

	return scr.Lines()
}

// RenderHelp renders the help overlay and returns composed lines.
func (r *Renderer) RenderHelp(vp layout.Viewport) []string {
	scr := newScreenBuf(vp.Width, vp.Height)

	helpLines := []string{
		"",
		"  mddeck – Keyboard Shortcuts",
		"",
		"  Navigation:",
		"    Space / Enter / → / PgDn / n    Next slide",
		"    Backspace / ← / PgUp / p        Previous slide",
		"    Home                             First slide",
		"    End                              Last slide",
		"",
		"  Modes:",
		"    t                                Toggle presenter mode",
		"    ?                                Toggle help",
		"    q / Ctrl+C                       Quit",
		"",
	}

	startY := (vp.Height - len(helpLines)) / 2
	if startY < 0 {
		startY = 0
	}

	for i, line := range helpLines {
		if startY+i >= vp.Height {
			break
		}
		scr.Set(startY+i, 0, r.Theme.HelpStyle+line+a.Reset)
	}

	scr.Set(vp.Height-1, 0, r.Theme.Muted+"  Press any key to dismiss"+a.Reset)

	return scr.Lines()
}
