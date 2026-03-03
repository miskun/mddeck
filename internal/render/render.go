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

// RenderSlide renders a single slide into a buffer of positioned lines.
func (r *Renderer) RenderSlide(slide *model.Slide, vp layout.Viewport) string {
	lr := layout.ComputeLayout(slide, vp)
	var buf strings.Builder

	// Clear screen and set base colors
	buf.WriteString(a.ClearScreen)
	buf.WriteString(a.CursorHome)
	buf.WriteString(r.Theme.Fg)

	switch lr.Mode {
	case model.LayoutTitle:
		r.renderTitle(slide, lr.Regions[0], &buf)
	case model.LayoutTwoCol:
		r.renderTwoCol(slide, lr, &buf)
	case model.LayoutSplit:
		r.renderSplitLayout(slide, lr, &buf)
	case model.LayoutTerminal:
		r.renderSingleRegion(slide.Blocks, lr.Regions[0], &buf)
	default: // center
		r.renderSingleRegion(slide.Blocks, lr.Regions[0], &buf)
	}

	// Render status bar
	r.renderStatusBar(slide, vp, &buf)

	buf.WriteString(a.Reset)
	return buf.String()
}

// renderTitle renders a title-layout slide (centered vertically and horizontally).
func (r *Renderer) renderTitle(slide *model.Slide, region layout.Region, buf *strings.Builder) {
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
		// Center horizontally
		visLen := a.VisibleLen(line)
		padLeft := (region.Width - visLen) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		buf.WriteString(a.CursorTo(row+1, region.X+padLeft+1))
		buf.WriteString(line)
	}
}

// renderTwoCol renders a two-column layout.
func (r *Renderer) renderTwoCol(slide *model.Slide, lr layout.LayoutResult, buf *strings.Builder) {
	majors := layout.SplitBlocksIntoMajor(slide.Blocks)

	// Distribute major blocks alternately
	var leftBlocks, rightBlocks []model.Block
	for i, maj := range majors {
		blocks := []model.Block{maj.Heading}
		blocks = append(blocks, maj.Content...)
		// Remove zero-value headings
		var filtered []model.Block
		for _, b := range blocks {
			if b.Type == model.BlockHeading || b.Raw != "" || len(b.Lines) > 0 {
				filtered = append(filtered, b)
			}
		}
		if i%2 == 0 {
			leftBlocks = append(leftBlocks, filtered...)
		} else {
			rightBlocks = append(rightBlocks, filtered...)
		}
	}

	// If no right blocks, put everything on left
	if len(rightBlocks) == 0 && len(leftBlocks) > 1 {
		mid := len(leftBlocks) / 2
		rightBlocks = leftBlocks[mid:]
		leftBlocks = leftBlocks[:mid]
	}

	r.renderInRegion(leftBlocks, lr.Regions[0], buf)
	if len(lr.Regions) > 1 {
		r.renderInRegion(rightBlocks, lr.Regions[1], buf)
	}
}

// renderSplitLayout renders a split (top/bottom) layout.
func (r *Renderer) renderSplitLayout(slide *model.Slide, lr layout.LayoutResult, buf *strings.Builder) {
	majors := layout.SplitBlocksIntoMajor(slide.Blocks)

	var topBlocks, bottomBlocks []model.Block
	if len(majors) > 0 {
		topBlocks = append([]model.Block{majors[0].Heading}, majors[0].Content...)
	}
	for _, maj := range majors[1:] {
		bottomBlocks = append(bottomBlocks, maj.Heading)
		bottomBlocks = append(bottomBlocks, maj.Content...)
	}

	r.renderInRegion(topBlocks, lr.Regions[0], buf)
	if len(lr.Regions) > 1 {
		r.renderInRegion(bottomBlocks, lr.Regions[1], buf)
	}
}

// renderSingleRegion renders all blocks into a single region.
func (r *Renderer) renderSingleRegion(blocks []model.Block, region layout.Region, buf *strings.Builder) {
	r.renderInRegion(blocks, region, buf)
}

// renderInRegion renders blocks within a specific region.
func (r *Renderer) renderInRegion(blocks []model.Block, region layout.Region, buf *strings.Builder) {
	lines := r.renderBlocks(blocks, region.Width)

	// Apply vertical alignment
	startY := region.Y
	if len(lines) < region.Height {
		// Default is top-aligned, handled by startY = region.Y
	}

	for i, line := range lines {
		if i >= region.Height {
			// Vertical overflow indicator
			buf.WriteString(a.CursorTo(region.Y+region.Height, region.X+1))
			buf.WriteString(r.Theme.Muted + "↓" + a.Reset)
			break
		}
		buf.WriteString(a.CursorTo(startY+i+1, region.X+1))
		rendered := r.cropLine(line, region.Width)
		buf.WriteString(rendered)
	}
}

// renderBlocks converts blocks into styled lines.
func (r *Renderer) renderBlocks(blocks []model.Block, width int) []string {
	var lines []string

	for i, block := range blocks {
		blockLines := r.renderBlock(block, width)
		lines = append(lines, blockLines...)

		// Add spacing between blocks
		if i < len(blocks)-1 {
			lines = append(lines, "")
		}
	}

	return lines
}

// renderBlock renders a single block into styled lines.
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
func (r *Renderer) renderParagraph(block model.Block, width int) []string {
	text := r.applyInlineStyles(block.Raw)

	if r.Wrap && width > 0 {
		return wrapText(text, width)
	}
	return []string{text}
}

// renderUnorderedList renders an unordered list.
func (r *Renderer) renderUnorderedList(block model.Block, width int) []string {
	var lines []string
	bullet := r.Theme.Accent + r.Theme.BulletChar + a.Reset

	for _, item := range block.Lines {
		text := r.applyInlineStyles(item)
		itemLine := "  " + bullet + text

		if r.Wrap && width > 4 {
			wrapped := wrapText(text, width-4)
			for j, wl := range wrapped {
				if j == 0 {
					lines = append(lines, "  "+bullet+wl)
				} else {
					lines = append(lines, "    "+wl) // hanging indent
				}
			}
		} else {
			lines = append(lines, itemLine)
		}
	}
	return lines
}

// renderOrderedList renders an ordered list.
func (r *Renderer) renderOrderedList(block model.Block, width int) []string {
	var lines []string

	for i, item := range block.Lines {
		num := fmt.Sprintf("%s%d.%s ", r.Theme.Accent, i+1, a.Reset)
		text := r.applyInlineStyles(item)

		prefix := "  " + num
		prefixWidth := 5 // "  N. "

		if r.Wrap && width > prefixWidth {
			wrapped := wrapText(text, width-prefixWidth)
			for j, wl := range wrapped {
				if j == 0 {
					lines = append(lines, prefix+wl)
				} else {
					lines = append(lines, strings.Repeat(" ", prefixWidth)+wl)
				}
			}
		} else {
			lines = append(lines, prefix+text)
		}
	}
	return lines
}

// renderBlockquote renders a blockquote.
func (r *Renderer) renderBlockquote(block model.Block, width int) []string {
	var lines []string
	indicator := r.Theme.Muted + r.Theme.BlockquoteChar

	for _, line := range block.Lines {
		text := r.applyInlineStyles(line)
		styled := indicator + text + a.Reset
		lines = append(lines, styled)
	}
	return lines
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

// renderStatusBar draws the slide number at the bottom.
func (r *Renderer) renderStatusBar(slide *model.Slide, vp layout.Viewport, buf *strings.Builder) {
	total := len(r.Deck.Slides)
	status := fmt.Sprintf(" %d / %d ", slide.Index+1, total)

	buf.WriteString(a.CursorTo(vp.Height, vp.Width-len(status)))
	buf.WriteString(r.Theme.SlideNumStyle + status + a.Reset)
}

// Inline style patterns
var (
	boldRegex   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRegex = regexp.MustCompile(`\*(.+?)\*`)
	codeRegex   = regexp.MustCompile("`([^`]+)`")
	linkRegex   = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
)

// applyInlineStyles applies bold, italic, code, and link styles.
func (r *Renderer) applyInlineStyles(text string) string {
	// Bold
	text = boldRegex.ReplaceAllString(text, a.Bold+"$1"+a.Reset+r.Theme.Fg)
	// Italic (must come after bold to avoid conflicts)
	text = italicRegex.ReplaceAllString(text, a.Italic+"$1"+a.Reset+r.Theme.Fg)
	// Inline code
	text = codeRegex.ReplaceAllString(text, r.Theme.CodeFg+"$1"+a.Reset+r.Theme.Fg)
	// Links (render as styled text)
	text = linkRegex.ReplaceAllString(text, a.Underline+r.Theme.Accent+"$1"+a.Reset+r.Theme.Fg)

	return text
}

// wrapText wraps text to fit within the given width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	// Strip ANSI for measuring, but we need to preserve them in output
	clean := a.StripAll(text)
	if utf8.RuneCountInString(clean) <= width {
		return []string{text}
	}

	// Simple word-wrap on the clean text, then reconstruct with ANSI
	words := strings.Fields(clean)
	var lines []string
	var currentLine strings.Builder
	currentLen := 0

	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)
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

	// Re-apply inline styles to each wrapped line
	return lines
}

// cropLine truncates a line if it exceeds the width.
func (r *Renderer) cropLine(line string, width int) string {
	visLen := a.VisibleLen(line)
	if visLen <= width {
		return line
	}

	// Need to truncate while keeping ANSI sequences intact
	clean := a.StripAll(line)
	runes := []rune(clean)
	if len(runes) > width-1 {
		return string(runes[:width-1]) + "…"
	}
	return line
}

// expandTabs replaces tabs with spaces.
func (r *Renderer) expandTabs(line string) string {
	return strings.ReplaceAll(line, "\t", strings.Repeat(" ", r.TabSize))
}

// RenderPresenter renders the presenter view with current slide, next preview,
// notes, timer, and slide index.
func (r *Renderer) RenderPresenter(slide *model.Slide, vp layout.Viewport, elapsed string) string {
	var buf strings.Builder

	buf.WriteString(a.ClearScreen)
	buf.WriteString(a.CursorHome)
	buf.WriteString(r.Theme.Fg)

	// Layout: top 60% = current slide, bottom 40% split into next preview + notes
	topH := vp.Height * 55 / 100
	bottomH := vp.Height - topH - 2 // 2 lines for divider and status

	// Current slide region
	currentRegion := layout.Region{
		X: 1, Y: 0,
		Width: vp.Width - 2, Height: topH,
	}

	// Render current slide content
	lr := layout.ComputeLayout(slide, layout.Viewport{Width: currentRegion.Width, Height: currentRegion.Height})
	if len(lr.Regions) > 0 {
		renderedLines := r.renderBlocks(slide.Blocks, lr.Regions[0].Width)
		for i, line := range renderedLines {
			if i >= currentRegion.Height {
				break
			}
			buf.WriteString(a.CursorTo(currentRegion.Y+i+1, currentRegion.X+1))
			buf.WriteString(r.cropLine(line, currentRegion.Width))
		}
	}

	// Divider
	dividerY := topH + 1
	buf.WriteString(a.CursorTo(dividerY, 1))
	buf.WriteString(r.Theme.HRStyle + strings.Repeat("─", vp.Width) + a.Reset)

	// Bottom: left = next preview, right = notes
	notesWidth := vp.Width / 2
	previewWidth := vp.Width - notesWidth - 1

	// Next slide preview
	total := len(r.Deck.Slides)
	if slide.Index+1 < total {
		nextSlide := &r.Deck.Slides[slide.Index+1]
		buf.WriteString(a.CursorTo(dividerY+1, 1))
		buf.WriteString(r.Theme.Muted + "Next:" + a.Reset)

		nextLines := r.renderBlocks(nextSlide.Blocks, previewWidth-2)
		for i, line := range nextLines {
			if i >= bottomH-1 {
				break
			}
			buf.WriteString(a.CursorTo(dividerY+2+i, 2))
			buf.WriteString(r.Theme.Muted + r.cropLine(line, previewWidth-2) + a.Reset)
		}
	}

	// Notes
	if slide.Notes != "" {
		buf.WriteString(a.CursorTo(dividerY+1, previewWidth+2))
		buf.WriteString(r.Theme.NotesStyle + "Notes:" + a.Reset)

		noteLines := strings.Split(slide.Notes, "\n")
		for i, line := range noteLines {
			if i >= bottomH-1 {
				break
			}
			buf.WriteString(a.CursorTo(dividerY+2+i, previewWidth+2))
			buf.WriteString(r.Theme.NotesStyle + r.cropLine(line, notesWidth-1) + a.Reset)
		}
	}

	// Status bar: timer + slide index
	statusLine := fmt.Sprintf(" %s  │  %d / %d ", elapsed, slide.Index+1, total)
	buf.WriteString(a.CursorTo(vp.Height, vp.Width-utf8.RuneCountInString(statusLine)))
	buf.WriteString(r.Theme.TimerStyle + statusLine + a.Reset)

	return buf.String()
}

// RenderHelp renders the help overlay.
func (r *Renderer) RenderHelp(vp layout.Viewport) string {
	var buf strings.Builder

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

	buf.WriteString(a.ClearScreen)
	buf.WriteString(a.CursorHome)

	startY := (vp.Height - len(helpLines)) / 2
	if startY < 0 {
		startY = 0
	}

	for i, line := range helpLines {
		buf.WriteString(a.CursorTo(startY+i+1, 1))
		buf.WriteString(r.Theme.HelpStyle + line + a.Reset)
	}

	buf.WriteString(a.CursorTo(vp.Height, 1))
	buf.WriteString(r.Theme.Muted + "  Press any key to dismiss" + a.Reset)

	return buf.String()
}
