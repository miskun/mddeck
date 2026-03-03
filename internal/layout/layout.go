// Package layout implements the slide layout system.
package layout

import (
	"strconv"
	"strings"

	"github.com/miskun/mddeck/internal/model"
)

// Region represents a rectangular area of the viewport.
type Region struct {
	X, Y          int // top-left position (0-based)
	Width, Height int
}

// LayoutResult describes how a slide should be rendered.
type LayoutResult struct {
	Mode    model.Layout
	Regions []Region
}

// Viewport describes the terminal dimensions.
type Viewport struct {
	Width  int
	Height int
}

// ComputeLayout determines the layout for a slide.
func ComputeLayout(slide *model.Slide, vp Viewport) LayoutResult {
	layout := slide.Meta.Layout
	if layout == model.LayoutAuto {
		layout = autoDetect(slide)
	}

	switch layout {
	case model.LayoutTitle:
		return layoutTitle(vp)
	case model.LayoutCenter:
		return layoutCenter(vp)
	case model.LayoutTwoCol:
		return layoutTwoCol(slide, vp)
	case model.LayoutSplit:
		return layoutSplit(vp)
	case model.LayoutTerminal:
		return layoutTerminal(vp)
	default:
		return layoutCenter(vp)
	}
}

// autoDetect implements the auto layout heuristics from §10.3.
func autoDetect(slide *model.Slide) model.Layout {
	blocks := slide.Blocks

	if len(blocks) == 0 {
		return model.LayoutCenter
	}

	// Count headings, art/code blocks, and major blocks
	h1Count := 0
	artCodeCount := 0
	totalNonBlank := 0
	majorBlocks := countMajorBlocks(blocks)

	for _, b := range blocks {
		if b.Type == model.BlockHeading && b.Level == 1 {
			h1Count++
		}
		if b.IsCodeLike() {
			artCodeCount++
		}
		if b.Type != model.BlockHorizontalRule {
			totalNonBlank++
		}
	}

	// Single H1 + minimal text → title
	if h1Count == 1 && totalNonBlank <= 3 {
		return model.LayoutTitle
	}

	// Single large art/code block → terminal
	if artCodeCount == 1 && totalNonBlank <= 2 {
		return model.LayoutTerminal
	}

	// Two major blocks → two-col
	if majorBlocks == 2 {
		return model.LayoutTwoCol
	}

	return model.LayoutCenter
}

// countMajorBlocks counts the number of major blocks (top-level heading + content).
func countMajorBlocks(blocks []model.Block) int {
	count := 0
	for _, b := range blocks {
		if b.Type == model.BlockHeading && b.Level <= 2 {
			count++
		}
	}
	if count == 0 && len(blocks) > 0 {
		return 1
	}
	return count
}

// layoutTitle creates a centered title layout (single region).
func layoutTitle(vp Viewport) LayoutResult {
	return LayoutResult{
		Mode: model.LayoutTitle,
		Regions: []Region{
			{X: 0, Y: 0, Width: vp.Width, Height: vp.Height},
		},
	}
}

// layoutCenter creates a centered content layout (single region with padding).
func layoutCenter(vp Viewport) LayoutResult {
	padX := vp.Width / 8
	if padX < 2 {
		padX = 2
	}
	padY := 1

	return LayoutResult{
		Mode: model.LayoutCenter,
		Regions: []Region{
			{
				X:      padX,
				Y:      padY,
				Width:  vp.Width - 2*padX,
				Height: vp.Height - 2*padY - 1, // leave room for status bar
			},
		},
	}
}

// layoutTwoCol splits viewport into two columns.
func layoutTwoCol(slide *model.Slide, vp Viewport) LayoutResult {
	leftPct, rightPct := 62, 38
	if slide.Meta.Ratio != "" {
		l, r, ok := parseRatio(slide.Meta.Ratio)
		if ok {
			leftPct, rightPct = l, r
		}
	}

	gutter := 2
	usable := vp.Width - gutter
	leftW := usable * leftPct / (leftPct + rightPct)
	rightW := usable - leftW

	padY := 1
	h := vp.Height - 2*padY - 1

	return LayoutResult{
		Mode: model.LayoutTwoCol,
		Regions: []Region{
			{X: 0, Y: padY, Width: leftW, Height: h},
			{X: leftW + gutter, Y: padY, Width: rightW, Height: h},
		},
	}
}

// layoutSplit splits viewport into top (60%) and bottom (40%).
func layoutSplit(vp Viewport) LayoutResult {
	padY := 1
	usable := vp.Height - 2*padY - 1
	topH := usable * 60 / 100
	bottomH := usable - topH

	padX := 2

	return LayoutResult{
		Mode: model.LayoutSplit,
		Regions: []Region{
			{X: padX, Y: padY, Width: vp.Width - 2*padX, Height: topH},
			{X: padX, Y: padY + topH + 1, Width: vp.Width - 2*padX, Height: bottomH},
		},
	}
}

// layoutTerminal creates a full-width region for terminal/code content.
func layoutTerminal(vp Viewport) LayoutResult {
	padX := 2
	padY := 1
	return LayoutResult{
		Mode: model.LayoutTerminal,
		Regions: []Region{
			{
				X:      padX,
				Y:      padY,
				Width:  vp.Width - 2*padX,
				Height: vp.Height - 2*padY - 1,
			},
		},
	}
}

// parseRatio parses a "A/B" ratio string.
func parseRatio(s string) (int, int, bool) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || a <= 0 || b <= 0 {
		return 0, 0, false
	}
	return a, b, true
}

// SplitBlocksIntoMajor groups blocks into major blocks for two-col/split layout.
func SplitBlocksIntoMajor(blocks []model.Block) []model.MajorBlock {
	var majors []model.MajorBlock
	var current *model.MajorBlock

	for _, b := range blocks {
		if b.Type == model.BlockHeading && b.Level <= 2 {
			if current != nil {
				majors = append(majors, *current)
			}
			current = &model.MajorBlock{
				Heading: b,
			}
		} else {
			if current == nil {
				current = &model.MajorBlock{}
			}
			current.Content = append(current.Content, b)
		}
	}

	if current != nil {
		majors = append(majors, *current)
	}

	return majors
}
