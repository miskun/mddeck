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
	Mode        model.Layout
	Regions     []Region
	HasTitleRow bool // first row is a dedicated title region (1 column, fixed height)
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
		"title-cols-2": {
			Grid: []model.LayoutRow{
				{Height: -1, Columns: []int{100}},
				{Columns: []int{50, 50}},
			},
			PadY: intPtr(1),
		},
		"title-cols-3": {
			Grid: []model.LayoutRow{
				{Height: -1, Columns: []int{100}},
				{Columns: []int{33, 34, 33}},
			},
			PadY: intPtr(1),
		},
		"title-grid-4": {
			Grid: []model.LayoutRow{
				{Height: -1, Columns: []int{100}},
				{Columns: []int{50, 50}},
				{Columns: []int{50, 50}},
			},
			PadY: intPtr(1),
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
	if len(override.Grid) > 0 {
		result.Grid = override.Grid
	}
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

	// Read slide dimension parameters
	slideWidth := 80
	slideHeight := -1
	aspect := "16:9"
	if deckMeta != nil {
		slideWidth = deckMeta.GetSlideWidth()
		slideHeight = deckMeta.GetSlideHeight()
		if deckMeta.Aspect != "" {
			aspect = deckMeta.Aspect
		}
	}

	// Compute slide stage dimensions and centering padding
	stageW, stageH, stagePadX, stagePadY := computeSlideDimensions(vp, slideWidth, slideHeight, aspect)

	// Resolve layout definition (builtin, overridden, or custom)
	def := resolveLayout(string(layout), deckMeta)

	// For cols-2, allow per-slide ratio override
	if layout == model.LayoutCols2 && slide.Meta.Ratio != "" {
		if l, r, ok := parseRatio(slide.Meta.Ratio); ok {
			def.Columns = []int{l, r}
		}
	}

	return computeGrid(def, layout, vp, stageW, stageH, stagePadX, stagePadY)
}

// computeGrid creates a grid layout from a CustomLayout definition.
// This is the single layout engine used by all layouts — built-in and custom.
// stageW and stageH are the pre-computed content stage dimensions.
// stagePadX and stagePadY position the stage within the viewport.
func computeGrid(def model.CustomLayout, name model.Layout, vp Viewport, stageW, stageH, stagePadX, stagePadY int) LayoutResult {
	gutter := def.GetGutter()

	// Layout-level padding overrides (additive on top of stage padding)
	padX := stagePadX
	padY := stagePadY
	usableW := stageW
	usableH := stageH

	// If layout defines explicit padX/padY, they add inner padding to the stage
	if px := def.GetPadX(); px >= 0 {
		padX += px
		usableW -= 2 * px
	}
	if py := def.GetPadY(); py >= 0 {
		padY += py
		usableH -= 2 * py
	}

	if usableW < 1 {
		usableW = 1
	}
	if usableH < 1 {
		usableH = 1
	}

	// Per-row grid mode: each row defines its own columns
	if len(def.Grid) > 0 {
		return computePerRowGrid(def.Grid, gutter, name, usableW, usableH, padX, padY)
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

// computePerRowGrid builds regions for a per-row grid layout where each row
// has its own column definitions.
// Height semantics: positive = percentage, negative = fixed rows, zero = equal share.
func computePerRowGrid(grid []model.LayoutRow, gutter int, name model.Layout, usableW, usableH, padX, padY int) LayoutResult {
	// Compute row heights, handling fixed vs percentage rows
	totalGutterY := gutter * (len(grid) - 1)
	availH := usableH - totalGutterY
	if availH < len(grid) {
		availH = len(grid)
	}

	// First pass: subtract fixed-height rows from available space
	fixedTotal := 0
	pctCount := 0
	for _, row := range grid {
		if row.Height < 0 {
			fixedTotal += -row.Height
		} else {
			pctCount++
		}
	}
	remainingH := availH - fixedTotal
	if remainingH < pctCount {
		remainingH = pctCount
	}

	// Second pass: distribute remaining space among percentage rows
	var pctValues []int
	for _, row := range grid {
		if row.Height >= 0 {
			h := row.Height
			if h == 0 && pctCount > 0 {
				h = 100 / pctCount // equal share
			}
			pctValues = append(pctValues, h)
		}
	}
	var pctHeights []int
	if len(pctValues) > 0 {
		pctHeights = distributeSpace(pctValues, remainingH)
	}

	// Build final rowHeights array
	rowHeights := make([]int, len(grid))
	pi := 0
	for i, row := range grid {
		if row.Height < 0 {
			rowHeights[i] = -row.Height
		} else {
			rowHeights[i] = pctHeights[pi]
			pi++
		}
	}

	// Detect title row: first row is single-column with fixed height of 1.
	hasTitleRow := len(grid) > 1 && len(grid[0].Columns) <= 1 && grid[0].Height < 0

	// Build regions row by row, each with its own column widths
	var regions []Region
	curY := padY
	for ri, row := range grid {
		cols := row.Columns
		if len(cols) == 0 {
			cols = []int{100}
		}

		totalGutterX := gutter * (len(cols) - 1)
		availW := usableW - totalGutterX
		if availW < len(cols) {
			availW = len(cols)
		}
		colWidths := distributeSpace(cols, availW)

		curX := padX
		for ci, cw := range colWidths {
			regions = append(regions, Region{
				X: curX, Y: curY,
				Width: cw, Height: rowHeights[ri],
			})
			curX += cw
			if ci < len(colWidths)-1 {
				curX += gutter
			}
		}

		curY += rowHeights[ri]
		if ri < len(grid)-1 {
			curY += gutter
		}
	}

	return LayoutResult{
		Mode:        name,
		Regions:     regions,
		HasTitleRow: hasTitleRow,
	}
}

// computeSlideDimensions calculates stage width, height, and centering padding.
//
// slideWidth/slideHeight semantics:
//
//	> 0  = explicit size in characters
//	  0  = fill terminal (no padding on that axis)
//	 -1  = auto-calculate from the other dimension + aspect ratio
//
// When both are auto (-1), the slide fills the terminal constrained by aspect.
// When both are explicit (> 0), aspect is ignored.
// The footer row is always reserved (1 row subtracted from available height).
func computeSlideDimensions(vp Viewport, slideWidth, slideHeight int, aspect string) (stageW, stageH, padX, padY int) {
	termW := vp.Width
	termH := vp.Height - 1 // reserve footer row

	if termW < 1 {
		termW = 1
	}
	if termH < 1 {
		termH = 1
	}

	num, den, hasAspect := parseAspect(aspect)

	switch {
	case slideWidth > 0 && slideHeight > 0:
		// Both explicit → use as-is, ignore aspect
		stageW = slideWidth
		stageH = slideHeight

	case slideWidth > 0 && slideHeight == 0:
		// Explicit width, fill height
		stageW = slideWidth
		stageH = termH

	case slideWidth == 0 && slideHeight > 0:
		// Fill width, explicit height
		stageW = termW
		stageH = slideHeight

	case slideWidth == 0 && slideHeight == 0:
		// Fill both axes — no padding
		stageW = termW
		stageH = termH

	case slideWidth > 0 && slideHeight < 0:
		// Explicit width, auto height from aspect
		stageW = slideWidth
		if hasAspect {
			stageH = stageW * den / (2 * num)
		} else {
			stageH = termH
		}

	case slideWidth < 0 && slideHeight > 0:
		// Auto width from aspect, explicit height
		stageH = slideHeight
		if hasAspect {
			stageW = stageH * 2 * num / den
		} else {
			stageW = termW
		}

	case slideWidth == 0 && slideHeight < 0:
		// Fill width, auto height from aspect
		stageW = termW
		if hasAspect {
			stageH = stageW * den / (2 * num)
		} else {
			stageH = termH
		}

	case slideWidth < 0 && slideHeight == 0:
		// Auto width from aspect, fill height
		stageH = termH
		if hasAspect {
			stageW = stageH * 2 * num / den
		} else {
			stageW = termW
		}

	default:
		// Both auto (-1, -1): maximize within terminal constrained by aspect
		if hasAspect {
			// Try fitting to terminal width first
			candidateH := termW * den / (2 * num)
			if candidateH <= termH {
				stageW = termW
				stageH = candidateH
			} else {
				// Fit to terminal height
				stageH = termH
				stageW = termH * 2 * num / den
			}
		} else {
			stageW = termW
			stageH = termH
		}
	}

	// Clamp to terminal bounds
	if stageW > termW {
		stageW = termW
	}
	if stageH > termH {
		stageH = termH
	}
	if stageW < 1 {
		stageW = 1
	}
	if stageH < 1 {
		stageH = 1
	}

	// Center the stage in the terminal
	padX = (termW - stageW) / 2
	padY = (termH - stageH) / 2

	return stageW, stageH, padX, padY
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

	n := len(pcts)
	sizes := make([]int, n)

	// Compute base sizes using integer division and track remainders
	type remainder struct {
		idx  int
		frac int // fractional part (numerator) for largest-remainder sorting
	}
	remainders := make([]remainder, n)
	used := 0
	for i, p := range pcts {
		// exact = available * p / total, but we need the fractional part
		// sizes[i] = floor(available * p / total)
		sizes[i] = available * p / total
		if sizes[i] < 1 {
			sizes[i] = 1
		}
		// fractional remainder: (available * p) mod total
		remainders[i] = remainder{idx: i, frac: (available * p) % total}
		used += sizes[i]
	}

	// Distribute leftover rows one at a time to entries with largest remainders
	leftover := available - used
	if leftover > 0 {
		// Simple selection: give extra to entries with largest fractional remainder
		for leftover > 0 {
			bestIdx := 0
			bestFrac := -1
			for _, r := range remainders {
				if r.frac > bestFrac {
					bestFrac = r.frac
					bestIdx = r.idx
				}
			}
			sizes[bestIdx]++
			// Zero out this entry's remainder so it doesn't get another extra
			for j := range remainders {
				if remainders[j].idx == bestIdx {
					remainders[j].frac = -1
					break
				}
			}
			leftover--
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

// countMajorBlocks counts the number of major blocks (top-level heading + content
// or region break boundaries).
func countMajorBlocks(blocks []model.Block) int {
	count := 0
	for _, b := range blocks {
		if b.Type == model.BlockHeading && b.Level <= 2 {
			count++
		}
		if b.Type == model.BlockRegionBreak {
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
// Splits occur on headings and on region break markers (BlockRegionBreak).
func SplitBlocksIntoMajor(blocks []model.Block) []model.MajorBlock {
	var majors []model.MajorBlock
	var current *model.MajorBlock

	for _, b := range blocks {
		if b.Type == model.BlockRegionBreak {
			// Region break: flush current major block and start a new
			// headingless one.  The break itself is consumed.
			if current != nil {
				majors = append(majors, *current)
			}
			current = nil
			continue
		}
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
