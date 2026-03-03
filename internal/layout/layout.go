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

// intPtr is a helper to create *int values for layout definitions.
func intPtr(v int) *int { return &v }

// builtinLayouts returns the default CustomLayout definitions for all built-in layouts.
// These use the exact same parameters as user-defined custom layouts.
func builtinLayouts() map[string]model.CustomLayout {
	return map[string]model.CustomLayout{
		"title": {
			Columns: []int{100},
		},
		"default": {
			Columns: []int{100},
			PadY:    intPtr(1),
		},
		"center": {
			Columns: []int{100},
			PadY:    intPtr(0),
		},
		"cols-2": {
			Columns: []int{50, 50},
			PadY:    intPtr(1),
		},
		"rows-2": {
			Rows: []int{60, 40},
			PadY: intPtr(1),
		},
		"terminal": {
			Columns: []int{100},
			PadY:    intPtr(1),
		},
		"sidebar": {
			Columns: []int{30, 70},
			PadY:    intPtr(1),
		},
		"cols-3": {
			Columns: []int{33, 34, 33},
			PadY:    intPtr(1),
		},
		"grid-4": {
			Columns: []int{50, 50},
			Rows:    []int{50, 50},
			PadY:    intPtr(1),
		},
	}
}

// resolveLayout looks up the layout definition by name.
// Priority: deck-level override (merged with builtin) → builtin → "default" builtin.
func resolveLayout(name string, deckMeta *model.DeckMeta) model.CustomLayout {
	builtins := builtinLayouts()
	base, isBuiltin := builtins[name]

	// Check deck-level layouts
	if deckMeta != nil && deckMeta.Layouts != nil {
		if custom, ok := deckMeta.Layouts[name]; ok {
			if isBuiltin {
				// Merge: user overrides on top of builtin defaults
				return mergeCustomLayout(base, custom)
			}
			// Pure custom layout
			return custom
		}
	}

	if isBuiltin {
		return base
	}

	// Unknown layout name → fall back to default
	return builtins[string(model.LayoutDefault)]
}

// mergeCustomLayout overlays non-zero fields from override onto base.
func mergeCustomLayout(base, override model.CustomLayout) model.CustomLayout {
	result := base
	if len(override.Columns) > 0 {
		result.Columns = override.Columns
	}
	if len(override.Rows) > 0 {
		result.Rows = override.Rows
	}
	if override.Gutter != nil {
		result.Gutter = override.Gutter
	}
	if override.PadX != nil {
		result.PadX = override.PadX
	}
	if override.PadY != nil {
		result.PadY = override.PadY
	}
	if override.Align != "" {
		result.Align = override.Align
	}
	return result
}

// ComputeLayout determines the layout for a slide.
// All layouts — built-in and custom — go through the same grid engine.
func ComputeLayout(slide *model.Slide, vp Viewport, deckMeta *model.DeckMeta) LayoutResult {
	layout := slide.Meta.Layout
	if layout == model.LayoutAuto {
		layout = autoDetect(slide)
	}

	// Compute aspect-ratio-based padding (default: 16:9)
	aspect := "16:9"
	if deckMeta != nil && deckMeta.Aspect != "" {
		aspect = deckMeta.Aspect
	}
	aspectPadX, aspectPadY := computeAspectPadding(aspect, vp)

	// Resolve layout definition (builtin, overridden, or custom)
	def := resolveLayout(string(layout), deckMeta)

	// For cols-2, allow per-slide ratio override
	if layout == model.LayoutCols2 && slide.Meta.Ratio != "" {
		if l, r, ok := parseRatio(slide.Meta.Ratio); ok {
			def.Columns = []int{l, r}
		}
	}

	return computeGrid(def, layout, vp, aspectPadX, aspectPadY)
}

// computeGrid creates a grid layout from a CustomLayout definition.
// This is the single layout engine used by all layouts — built-in and custom.
func computeGrid(def model.CustomLayout, name model.Layout, vp Viewport, aspectPadX, aspectPadY int) LayoutResult {
	gutter := def.GetGutter()

	// Determine padding.
	// If padX is explicitly set, use it (aspect padding as minimum).
	// If unset, use aspect padding or small fixed minimum (2).
	padX := resolvePadX(def, vp, aspectPadX)
	padY := resolvePadY(def, aspectPadY)

	usableW := vp.Width - 2*padX
	usableH := vp.Height - 2*padY - 1 // reserve status bar row

	if usableW < 1 {
		usableW = 1
	}
	if usableH < 1 {
		usableH = 1
	}

	cols := def.Columns
	rows := def.Rows

	// Default: 1 column or 1 row if not specified
	if len(cols) == 0 {
		cols = []int{100}
	}
	if len(rows) == 0 {
		rows = []int{100}
	}

	// Compute column widths from percentages
	totalGutterX := gutter * (len(cols) - 1)
	availW := usableW - totalGutterX
	if availW < len(cols) {
		availW = len(cols)
	}
	colWidths := distributeSpace(cols, availW)

	// Compute row heights from percentages
	totalGutterY := gutter * (len(rows) - 1)
	availH := usableH - totalGutterY
	if availH < len(rows) {
		availH = len(rows)
	}
	rowHeights := distributeSpace(rows, availH)

	// Build regions in row-major order
	var regions []Region
	curY := padY
	for ri, rh := range rowHeights {
		curX := padX
		for ci, cw := range colWidths {
			regions = append(regions, Region{
				X: curX, Y: curY,
				Width: cw, Height: rh,
			})
			curX += cw
			if ci < len(colWidths)-1 {
				curX += gutter
			}
		}
		curY += rh
		if ri < len(rowHeights)-1 {
			curY += gutter
		}
	}

	return LayoutResult{
		Mode:    name,
		Regions: regions,
	}
}

// resolvePadX determines horizontal padding.
// Explicit padX → use it (aspect as minimum).
// Unset → aspect padding if set, otherwise small fixed minimum (2).
func resolvePadX(def model.CustomLayout, vp Viewport, aspectPadX int) int {
	if px := def.GetPadX(); px >= 0 {
		if aspectPadX > px {
			return aspectPadX
		}
		return px
	}
	// No explicit padX → use aspect padding, or minimal default
	if aspectPadX > 2 {
		return aspectPadX
	}
	return 2
}

// resolvePadY determines vertical padding.
// Explicit padY → use it (aspect as minimum).
// Unset → aspect padding if set, otherwise small fixed minimum (1).
func resolvePadY(def model.CustomLayout, aspectPadY int) int {
	if py := def.GetPadY(); py >= 0 {
		if aspectPadY > py {
			return aspectPadY
		}
		return py
	}
	// No explicit padY → use aspect padding, or minimal default
	if aspectPadY > 1 {
		return aspectPadY
	}
	return 1
}

// computeAspectPadding calculates horizontal and vertical padding to achieve
// the target aspect ratio. Terminal characters are roughly 1:2 (width:height),
// so a character cell that is 1 col wide is about 2 units tall.
func computeAspectPadding(aspect string, vp Viewport) (padX, padY int) {
	num, den, ok := parseAspect(aspect)
	if !ok {
		return 0, 0
	}

	targetRatio := 2.0 * float64(num) / float64(den) // target W/H in character cells
	currentRatio := float64(vp.Width) / float64(vp.Height)

	if currentRatio > targetRatio {
		// Terminal is wider than target → add horizontal padding
		targetW := int(targetRatio * float64(vp.Height))
		padX = (vp.Width - targetW) / 2
		if padX < 0 {
			padX = 0
		}
	} else if currentRatio < targetRatio {
		// Terminal is taller than target → add vertical padding
		targetH := int(float64(vp.Width) / targetRatio)
		padY = (vp.Height - targetH) / 2
		if padY < 0 {
			padY = 0
		}
	}

	return padX, padY
}

// parseAspect parses an aspect ratio string like "16:9" or "4:3".
func parseAspect(s string) (int, int, bool) {
	parts := strings.SplitN(s, ":", 2)
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

// distributeSpace distributes available space according to percentage weights.
func distributeSpace(pcts []int, available int) []int {
	total := 0
	for _, p := range pcts {
		total += p
	}
	if total == 0 {
		total = 100
	}

	sizes := make([]int, len(pcts))
	used := 0
	for i, p := range pcts {
		if i == len(pcts)-1 {
			sizes[i] = available - used
		} else {
			sizes[i] = available * p / total
			used += sizes[i]
		}
		if sizes[i] < 1 {
			sizes[i] = 1
		}
	}
	return sizes
}

// autoDetect implements the auto layout heuristics.
func autoDetect(slide *model.Slide) model.Layout {
	blocks := slide.Blocks

	if len(blocks) == 0 {
		return model.LayoutDefault
	}

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

	if h1Count == 1 && totalNonBlank <= 3 {
		return model.LayoutTitle
	}
	// A single heading (any level) with minimal content → title slide.
	// This catches section dividers produced by header-based splitting
	// (e.g. "## 01 / THE FINANCIAL REALITY" as a standalone slide).
	if totalNonBlank <= 2 {
		headingOnly := true
		for _, b := range blocks {
			if b.Type != model.BlockHeading && b.Type != model.BlockHorizontalRule {
				headingOnly = false
				break
			}
		}
		if headingOnly {
			return model.LayoutTitle
		}
	}
	if artCodeCount == 1 && totalNonBlank <= 2 {
		return model.LayoutTerminal
	}
	if majorBlocks == 2 {
		return model.LayoutCols2
	}

	return model.LayoutDefault
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

// SplitBlocksIntoMajor groups blocks into major blocks for multi-region layouts.
func SplitBlocksIntoMajor(blocks []model.Block) []model.MajorBlock {
	var majors []model.MajorBlock
	var current *model.MajorBlock

	for _, b := range blocks {
		if b.Type == model.BlockHeading {
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
