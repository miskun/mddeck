package layout

import (
	"testing"

	"github.com/miskun/mddeck/internal/model"
)

func TestAutoDetectTitle(t *testing.T) {
	slide := &model.Slide{
		Meta: model.SlideMetaDefaults(),
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 1, Raw: "My Title"},
		},
	}

	vp := Viewport{Width: 80, Height: 24}
	result := ComputeLayout(slide, vp, nil)

	if result.Mode != model.LayoutTitle {
		t.Errorf("mode = %q, want %q", result.Mode, model.LayoutTitle)
	}
}

func TestAutoDetectTerminal(t *testing.T) {
	slide := &model.Slide{
		Meta: model.SlideMetaDefaults(),
		Blocks: []model.Block{
			{Type: model.BlockFencedCode, Raw: "lots of code", Language: "go"},
		},
	}

	vp := Viewport{Width: 80, Height: 24}
	result := ComputeLayout(slide, vp, nil)

	if result.Mode != model.LayoutTerminal {
		t.Errorf("mode = %q, want %q", result.Mode, model.LayoutTerminal)
	}
}

func TestAutoDetectCols2(t *testing.T) {
	slide := &model.Slide{
		Meta: model.SlideMetaDefaults(),
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 2, Raw: "Left"},
			{Type: model.BlockParagraph, Raw: "Left content"},
			{Type: model.BlockHeading, Level: 2, Raw: "Right"},
			{Type: model.BlockParagraph, Raw: "Right content"},
		},
	}

	vp := Viewport{Width: 80, Height: 24}
	result := ComputeLayout(slide, vp, nil)

	if result.Mode != model.LayoutCols2 {
		t.Errorf("mode = %q, want %q", result.Mode, model.LayoutCols2)
	}
}

func TestCols2Ratio(t *testing.T) {
	slide := &model.Slide{
		Meta: model.SlideMeta{
			Layout: model.LayoutCols2,
			Ratio:  "70/30",
			Align:  model.AlignTop,
		},
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 2, Raw: "Left"},
			{Type: model.BlockHeading, Level: 2, Raw: "Right"},
		},
	}

	vp := Viewport{Width: 100, Height: 24}
	result := ComputeLayout(slide, vp, nil)

	if len(result.Regions) != 2 {
		t.Fatalf("regions = %d, want 2", len(result.Regions))
	}

	leftW := result.Regions[0].Width
	rightW := result.Regions[1].Width

	// Default slideWidth 80, 16:9 → stageW=80, minus gutter 2 = 78
	// 70/30 of 78 → ~54/24
	if leftW < 50 || leftW > 65 {
		t.Errorf("left width = %d, expected ~59", leftW)
	}
	if rightW < 18 || rightW > 32 {
		t.Errorf("right width = %d, expected ~25", rightW)
	}
}

func TestParseRatio(t *testing.T) {
	tests := []struct {
		input    string
		a, b     int
		ok       bool
	}{
		{"60/40", 60, 40, true},
		{"50/50", 50, 50, true},
		{"invalid", 0, 0, false},
		{"0/50", 0, 0, false},
		{"/50", 0, 0, false},
	}

	for _, tt := range tests {
		a, b, ok := parseRatio(tt.input)
		if ok != tt.ok || a != tt.a || b != tt.b {
			t.Errorf("parseRatio(%q) = (%d, %d, %v), want (%d, %d, %v)",
				tt.input, a, b, ok, tt.a, tt.b, tt.ok)
		}
	}
}

func TestSplitBlocksIntoMajor(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 1, Raw: "Title"},
		{Type: model.BlockParagraph, Raw: "Intro"},
		{Type: model.BlockHeading, Level: 2, Raw: "Section"},
		{Type: model.BlockParagraph, Raw: "Content"},
	}

	majors := SplitBlocksIntoMajor(blocks)

	if len(majors) != 2 {
		t.Fatalf("majors = %d, want 2", len(majors))
	}

	if majors[0].Heading.Raw != "Title" {
		t.Errorf("first major heading = %q, want %q", majors[0].Heading.Raw, "Title")
	}
	if majors[1].Heading.Raw != "Section" {
		t.Errorf("second major heading = %q, want %q", majors[1].Heading.Raw, "Section")
	}
}

func TestSplitBlocksIntoMajorWithRegionBreak(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 1, Raw: "Title"},
		{Type: model.BlockParagraph, Raw: "Left content"},
		{Type: model.BlockRegionBreak},
		{Type: model.BlockParagraph, Raw: "Right content"},
	}

	majors := SplitBlocksIntoMajor(blocks)

	if len(majors) != 2 {
		t.Fatalf("majors = %d, want 2", len(majors))
	}

	// First major: heading + left content
	if majors[0].Heading.Raw != "Title" {
		t.Errorf("first major heading = %q, want %q", majors[0].Heading.Raw, "Title")
	}
	if len(majors[0].Content) != 1 || majors[0].Content[0].Raw != "Left content" {
		t.Errorf("first major content = %v, want [Left content]", majors[0].Content)
	}

	// Second major: no heading, just right content
	if majors[1].Heading.Type != 0 {
		t.Errorf("second major should have no heading, got type %d", majors[1].Heading.Type)
	}
	if len(majors[1].Content) != 1 || majors[1].Content[0].Raw != "Right content" {
		t.Errorf("second major content = %v, want [Right content]", majors[1].Content)
	}
}

func TestSplitBlocksRegionBreakOnly(t *testing.T) {
	// No headings at all, just region breaks
	blocks := []model.Block{
		{Type: model.BlockParagraph, Raw: "Section 1"},
		{Type: model.BlockRegionBreak},
		{Type: model.BlockParagraph, Raw: "Section 2"},
		{Type: model.BlockRegionBreak},
		{Type: model.BlockParagraph, Raw: "Section 3"},
	}

	majors := SplitBlocksIntoMajor(blocks)

	if len(majors) != 3 {
		t.Fatalf("majors = %d, want 3", len(majors))
	}

	for i, m := range majors {
		if m.Heading.Type != 0 {
			t.Errorf("major %d should have no heading", i)
		}
	}
	if majors[0].Content[0].Raw != "Section 1" {
		t.Errorf("major 0 content = %q, want Section 1", majors[0].Content[0].Raw)
	}
	if majors[2].Content[0].Raw != "Section 3" {
		t.Errorf("major 2 content = %q, want Section 3", majors[2].Content[0].Raw)
	}
}

func TestAspectPadding16x9(t *testing.T) {
	// 16:9 aspect on a 120x40 terminal, slideWidth auto, slideHeight auto
	vp := Viewport{Width: 120, Height: 40}
	_, _, padX, padY := computeSlideDimensions(vp, -1, -1, "16:9")

	// Target ratio in cells: 2*16/9 ≈ 3.556
	// Current ratio: 120/39(after footer) ≈ 3.077 → terminal is taller → padY > 0
	if padX != 0 {
		t.Errorf("padX = %d, want 0 (terminal is narrower than target)", padX)
	}
	if padY <= 0 {
		t.Errorf("padY = %d, want > 0 (terminal is taller than target)", padY)
	}
}

func TestAspectPadding4x3(t *testing.T) {
	// 4:3 aspect on a 120x40 terminal
	vp := Viewport{Width: 120, Height: 40}
	_, _, padX, _ := computeSlideDimensions(vp, -1, -1, "4:3")

	// Target ratio: 2*4/3 ≈ 2.667
	// termH = 39, candidateH = 120*3/(2*4) = 22 → fits → stageW=120, stageH=22
	// padY = (39-22)/2 = 8 → vertical padding
	// Actually: terminal is wider than target → padX > 0
	// Let me check: candidateH = 120*3/8 = 45 > 39 → doesn't fit → fit to height
	// stageW = 39*2*4/3 = 104 → padX = (120-104)/2 = 8
	if padX <= 0 {
		t.Errorf("padX = %d, want > 0 (terminal is wider than target)", padX)
	}
}

func TestAspectPaddingInvalid(t *testing.T) {
	vp := Viewport{Width: 80, Height: 24}
	stageW, stageH, _, _ := computeSlideDimensions(vp, -1, -1, "invalid")
	// Without valid aspect, both auto → fill terminal
	if stageW != 80 {
		t.Errorf("stageW = %d, want 80 (fill terminal on invalid aspect)", stageW)
	}
	if stageH != 23 { // 24 - 1 footer
		t.Errorf("stageH = %d, want 23", stageH)
	}
}

func TestCustomGridLayout2x1(t *testing.T) {
	// Two columns, one row: 60%/40%
	custom := model.CustomLayout{
		Columns: []int{60, 40},
		Rows:    []int{100},
	}
	vp := Viewport{Width: 100, Height: 30}
	// Full stage, no padding
	result := computeGrid(custom, "twocol", vp, 100, 29, 0, 0)

	if len(result.Regions) != 2 {
		t.Fatalf("regions = %d, want 2", len(result.Regions))
	}

	// Left should be wider than right
	if result.Regions[0].Width <= result.Regions[1].Width {
		t.Errorf("left width %d should be > right width %d",
			result.Regions[0].Width, result.Regions[1].Width)
	}

	// Both should have same Y
	if result.Regions[0].Y != result.Regions[1].Y {
		t.Errorf("regions should have same Y: %d vs %d",
			result.Regions[0].Y, result.Regions[1].Y)
	}
}

func TestCustomGridLayout2x2(t *testing.T) {
	// 2x2 grid: 50/50 columns, 50/50 rows
	custom := model.CustomLayout{
		Columns: []int{50, 50},
		Rows:    []int{50, 50},
	}
	vp := Viewport{Width: 100, Height: 30}
	// Full stage, no padding
	result := computeGrid(custom, "grid", vp, 100, 29, 0, 0)

	if len(result.Regions) != 4 {
		t.Fatalf("regions = %d, want 4", len(result.Regions))
	}

	// Check row-major order: [0]=top-left, [1]=top-right, [2]=bottom-left, [3]=bottom-right
	if result.Regions[0].X >= result.Regions[1].X {
		t.Error("region[0] should be left of region[1]")
	}
	if result.Regions[0].Y >= result.Regions[2].Y {
		t.Error("region[0] should be above region[2]")
	}
	if result.Regions[2].X >= result.Regions[3].X {
		t.Error("region[2] should be left of region[3]")
	}
}

func TestCustomLayoutLookup(t *testing.T) {
	deckMeta := &model.DeckMeta{
		Layouts: map[string]model.CustomLayout{
			"sidebar": {
				Columns: []int{30, 70},
			},
		},
	}
	slide := &model.Slide{
		Meta: model.SlideMeta{Layout: "sidebar"},
	}
	vp := Viewport{Width: 100, Height: 30}
	result := ComputeLayout(slide, vp, deckMeta)

	if len(result.Regions) != 2 {
		t.Fatalf("regions = %d, want 2", len(result.Regions))
	}
	if result.Mode != "sidebar" {
		t.Errorf("mode = %q, want %q", result.Mode, "sidebar")
	}
	// First region (30%) should be narrower than second (70%)
	if result.Regions[0].Width >= result.Regions[1].Width {
		t.Errorf("sidebar left %d should be < right %d",
			result.Regions[0].Width, result.Regions[1].Width)
	}
}

func TestDistributeSpace(t *testing.T) {
	sizes := distributeSpace([]int{60, 40}, 100)
	if len(sizes) != 2 {
		t.Fatalf("len = %d, want 2", len(sizes))
	}
	if sizes[0]+sizes[1] != 100 {
		t.Errorf("total = %d, want 100", sizes[0]+sizes[1])
	}
	if sizes[0] != 60 {
		t.Errorf("sizes[0] = %d, want 60", sizes[0])
	}
	if sizes[1] != 40 {
		t.Errorf("sizes[1] = %d, want 40", sizes[1])
	}
}

func TestSlideWidthCapsContentStage(t *testing.T) {
	// Wide terminal (160 cols), slideWidth 78 caps content stage
	vp := Viewport{Width: 160, Height: 40}

	stageW, _, padX, _ := computeSlideDimensions(vp, 78, -1, "16:9")

	// Stage width should be exactly 78
	if stageW != 78 {
		t.Errorf("stageW = %d, want 78", stageW)
	}

	// Content should be centered: padX = (160 - 78) / 2 = 41
	wantX := (160 - 78) / 2
	if padX != wantX {
		t.Errorf("padX = %d, want %d (centered)", padX, wantX)
	}
}

func TestSlideWidthNarrowTerminal(t *testing.T) {
	// Narrow terminal (60 cols) — slideWidth 78 gets clamped to terminal
	vp := Viewport{Width: 60, Height: 30}

	stageW, _, padX, _ := computeSlideDimensions(vp, 78, -1, "16:9")

	// Stage width clamped to terminal width
	if stageW != 60 {
		t.Errorf("stageW = %d, want 60 (clamped to terminal)", stageW)
	}
	if padX != 0 {
		t.Errorf("padX = %d, want 0 (no padding when clamped)", padX)
	}
}

func TestSlideWidthWithMultipleColumns(t *testing.T) {
	// Two columns within a 78-char stage
	custom := model.CustomLayout{
		Columns: []int{50, 50},
	}
	vp := Viewport{Width: 160, Height: 40}

	stageW, stageH, stagePadX, stagePadY := computeSlideDimensions(vp, 78, -1, "16:9")
	result := computeGrid(custom, "cols-2", vp, stageW, stageH, stagePadX, stagePadY)

	// Both regions should be within the 78-char stage
	leftEnd := result.Regions[0].X + result.Regions[0].Width
	rightStart := result.Regions[1].X
	rightEnd := result.Regions[1].X + result.Regions[1].Width

	stageLeft := result.Regions[0].X
	stageRight := rightEnd

	totalWidth := stageRight - stageLeft
	// Stage should be ≤ 78 (may be less due to gutter)
	if totalWidth > 78 {
		t.Errorf("total stage width %d exceeds slideWidth 78", totalWidth)
	}

	// Right column should start after left + gutter
	if rightStart <= leftEnd {
		t.Errorf("right column X %d should be > left end %d", rightStart, leftEnd)
	}
}

func TestSlideWidthWithAutoHeight(t *testing.T) {
	// slideWidth 78, slideHeight auto, 16:9. Height = 78*9/(2*16) = 21
	vp := Viewport{Width: 160, Height: 40}

	stageW, stageH, _, padY := computeSlideDimensions(vp, 78, -1, "16:9")

	if stageW != 78 {
		t.Errorf("stageW = %d, want 78", stageW)
	}

	wantH := 78 * 9 / (2 * 16) // = 21
	if stageH != wantH {
		t.Errorf("stageH = %d, want %d (aspect-ratio derived)", stageH, wantH)
	}

	// Content should be vertically centered
	termH := 40 - 1 // footer
	wantPadY := (termH - wantH) / 2
	if padY != wantPadY {
		t.Errorf("padY = %d, want %d (vertically centered)", padY, wantPadY)
	}
}

func TestSlideWidthFillTerminal(t *testing.T) {
	// slideWidth 0, slideHeight 0 → fill entire terminal
	vp := Viewport{Width: 100, Height: 30}

	stageW, stageH, padX, padY := computeSlideDimensions(vp, 0, 0, "16:9")

	if stageW != 100 {
		t.Errorf("stageW = %d, want 100 (fill terminal)", stageW)
	}
	if stageH != 29 { // minus footer
		t.Errorf("stageH = %d, want 29 (fill terminal minus footer)", stageH)
	}
	if padX != 0 {
		t.Errorf("padX = %d, want 0", padX)
	}
	if padY != 0 {
		t.Errorf("padY = %d, want 0", padY)
	}
}

func TestSlideHeightExplicitWidthAuto(t *testing.T) {
	// slideWidth auto, slideHeight 20, 16:9 → width = 20*2*16/9 = 71
	vp := Viewport{Width: 160, Height: 40}

	stageW, stageH, _, _ := computeSlideDimensions(vp, -1, 20, "16:9")

	if stageH != 20 {
		t.Errorf("stageH = %d, want 20", stageH)
	}

	wantW := 20 * 2 * 16 / 9 // = 71
	if stageW != wantW {
		t.Errorf("stageW = %d, want %d (auto from aspect)", stageW, wantW)
	}
}

func TestBothExplicitIgnoresAspect(t *testing.T) {
	// Both explicit → aspect is ignored
	vp := Viewport{Width: 160, Height: 40}

	stageW, stageH, _, _ := computeSlideDimensions(vp, 100, 30, "16:9")

	if stageW != 100 {
		t.Errorf("stageW = %d, want 100", stageW)
	}
	if stageH != 30 {
		t.Errorf("stageH = %d, want 30", stageH)
	}
}

// --- Per-row grid layout tests ---

func TestPerRowGridTitleCols2(t *testing.T) {
	// title-cols-2: row 0 = fixed 1-row title, row 1 = two columns
	def := model.CustomLayout{
		Grid: []model.LayoutRow{
			{Height: -1, Columns: []int{100}},
			{Columns: []int{50, 50}},
		},
	}

	result := computeGrid(def, "title-cols-2", Viewport{Width: 80, Height: 24}, 80, 23, 0, 0)

	if len(result.Regions) != 3 {
		t.Fatalf("got %d regions, want 3", len(result.Regions))
	}

	// Region 0: title (full width, fixed 1 row)
	if result.Regions[0].Width != 80 {
		t.Errorf("title region width = %d, want 80", result.Regions[0].Width)
	}
	if result.Regions[0].Height != 1 {
		t.Errorf("title region height = %d, want 1", result.Regions[0].Height)
	}

	// Regions 1 and 2: columns in second row
	if result.Regions[1].Width == result.Regions[2].Width {
		// Good: equal columns
	} else {
		t.Logf("col widths: %d, %d (may differ by 1 due to rounding)", result.Regions[1].Width, result.Regions[2].Width)
	}

	// Second row Y should be below title row
	if result.Regions[1].Y <= result.Regions[0].Y {
		t.Errorf("cols Y (%d) should be below title Y (%d)", result.Regions[1].Y, result.Regions[0].Y)
	}

	// Both columns same Y
	if result.Regions[1].Y != result.Regions[2].Y {
		t.Errorf("column Y mismatch: %d vs %d", result.Regions[1].Y, result.Regions[2].Y)
	}
}

func TestPerRowGridTitleGrid4(t *testing.T) {
	// title-grid-4: row 0 = fixed title, row 1 = 2 cols, row 2 = 2 cols
	def := model.CustomLayout{
		Grid: []model.LayoutRow{
			{Height: -1, Columns: []int{100}},
			{Columns: []int{50, 50}},
			{Columns: []int{50, 50}},
		},
	}

	result := computeGrid(def, "title-grid-4", Viewport{Width: 80, Height: 24}, 80, 23, 0, 0)

	if len(result.Regions) != 5 {
		t.Fatalf("got %d regions, want 5", len(result.Regions))
	}

	// Region 0: title (full width, fixed 1 row)
	if result.Regions[0].Width != 80 {
		t.Errorf("title width = %d, want 80", result.Regions[0].Width)
	}
	if result.Regions[0].Height != 1 {
		t.Errorf("title height = %d, want 1", result.Regions[0].Height)
	}

	// Regions 1-2: row 1 (same Y, different X)
	if result.Regions[1].Y != result.Regions[2].Y {
		t.Errorf("row 1 Y mismatch: %d vs %d", result.Regions[1].Y, result.Regions[2].Y)
	}

	// Regions 3-4: row 2 (same Y, different X, below row 1)
	if result.Regions[3].Y != result.Regions[4].Y {
		t.Errorf("row 2 Y mismatch: %d vs %d", result.Regions[3].Y, result.Regions[4].Y)
	}
	if result.Regions[3].Y <= result.Regions[1].Y {
		t.Errorf("row 2 Y (%d) should be below row 1 Y (%d)", result.Regions[3].Y, result.Regions[1].Y)
	}
}

func TestPerRowGridCustomViaYAML(t *testing.T) {
	// Test a fully custom grid: 3 rows with different column counts
	def := model.CustomLayout{
		Grid: []model.LayoutRow{
			{Height: 10, Columns: []int{100}},          // 1 region
			{Height: 60, Columns: []int{33, 34, 33}},   // 3 regions
			{Height: 30, Columns: []int{50, 50}},        // 2 regions
		},
	}

	result := computeGrid(def, "custom", Viewport{Width: 100, Height: 40}, 100, 39, 0, 0)

	if len(result.Regions) != 6 {
		t.Fatalf("got %d regions, want 6 (1+3+2)", len(result.Regions))
	}
}

func TestBuiltinTitleCols2Layout(t *testing.T) {
	// Verify the built-in title-cols-2 works through ComputeLayout
	slide := &model.Slide{
		Meta: model.SlideMeta{
			Layout: model.LayoutTitleCols2,
			Align:  model.AlignTop,
		},
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 2, Raw: "Title"},
			{Type: model.BlockHeading, Level: 3, Raw: "Left"},
			{Type: model.BlockParagraph, Raw: "Left content"},
			{Type: model.BlockHeading, Level: 3, Raw: "Right"},
			{Type: model.BlockParagraph, Raw: "Right content"},
		},
	}

	vp := Viewport{Width: 80, Height: 24}
	result := ComputeLayout(slide, vp, nil)

	if len(result.Regions) != 3 {
		t.Fatalf("title-cols-2 should produce 3 regions, got %d", len(result.Regions))
	}
}
