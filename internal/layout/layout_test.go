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

	// Default aspect 16:9 with 100x24 → aspectPadX ≈ 7 each side → 86 usable
	// minus gutter 2 = 84, 70/30 of 84 → ~59/25
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

func TestAspectPadding16x9(t *testing.T) {
	// 16:9 aspect on a 120x40 terminal
	vp := Viewport{Width: 120, Height: 40}
	padX, padY := computeAspectPadding("16:9", vp)

	// Target ratio in cells: 2*16/9 ≈ 3.556
	// Current ratio: 120/40 = 3.0 → terminal is taller than target → padY > 0
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
	padX, padY := computeAspectPadding("4:3", vp)

	// Target ratio: 2*4/3 ≈ 2.667
	// Current: 120/40 = 3.0 → terminal wider than target → padX > 0
	if padX <= 0 {
		t.Errorf("padX = %d, want > 0 (terminal is wider than target)", padX)
	}
	if padY != 0 {
		t.Errorf("padY = %d, want 0", padY)
	}
}

func TestAspectPaddingInvalid(t *testing.T) {
	vp := Viewport{Width: 80, Height: 24}
	padX, padY := computeAspectPadding("invalid", vp)
	if padX != 0 || padY != 0 {
		t.Errorf("invalid aspect should return (0,0), got (%d,%d)", padX, padY)
	}
}

func TestCustomGridLayout2x1(t *testing.T) {
	// Two columns, one row: 60%/40%
	custom := model.CustomLayout{
		Columns: []int{60, 40},
		Rows:    []int{100},
	}
	vp := Viewport{Width: 100, Height: 30}
	result := computeGrid(custom, "twocol", vp, 0, 0, 0, "")

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
	result := computeGrid(custom, "grid", vp, 0, 0, 0, "")

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

func TestLineWidthCapsContentStage(t *testing.T) {
	// Wide terminal (160 cols), lineWidth caps content to 78
	custom := model.CustomLayout{
		Columns: []int{100},
		Rows:    []int{100},
	}
	vp := Viewport{Width: 160, Height: 40}

	// Without cap
	uncapped := computeGrid(custom, "auto", vp, 0, 0, 0, "")
	// With 78-char cap
	capped := computeGrid(custom, "auto", vp, 0, 0, 78, "16:9")

	// Uncapped region should be wider than capped
	if uncapped.Regions[0].Width <= capped.Regions[0].Width {
		t.Errorf("uncapped width %d should be > capped width %d",
			uncapped.Regions[0].Width, capped.Regions[0].Width)
	}

	// Capped region width should be exactly 78
	if capped.Regions[0].Width != 78 {
		t.Errorf("capped width = %d, want 78", capped.Regions[0].Width)
	}

	// Content should be centered: X offset = (160 - 78) / 2
	wantX := (160 - 78) / 2
	if capped.Regions[0].X != wantX {
		t.Errorf("capped X = %d, want %d (centered)", capped.Regions[0].X, wantX)
	}
}

func TestLineWidthNarrowTerminal(t *testing.T) {
	// Narrow terminal (60 cols) — lineWidth 78 should have no effect
	custom := model.CustomLayout{
		Columns: []int{100},
	}
	vp := Viewport{Width: 60, Height: 30}

	result := computeGrid(custom, "auto", vp, 0, 0, 78, "16:9")
	// With default padX=2, usableW = 60 - 4 = 56, which is < 78 → no capping
	if result.Regions[0].Width != 56 {
		t.Errorf("width = %d, want 56 (no capping on narrow terminal)", result.Regions[0].Width)
	}
}

func TestLineWidthWithMultipleColumns(t *testing.T) {
	// Two columns within a 78-char stage
	custom := model.CustomLayout{
		Columns: []int{50, 50},
	}
	vp := Viewport{Width: 160, Height: 40}

	result := computeGrid(custom, "cols-2", vp, 0, 0, 78, "16:9")

	// Both regions should be within the 78-char stage
	leftEnd := result.Regions[0].X + result.Regions[0].Width
	rightStart := result.Regions[1].X
	rightEnd := result.Regions[1].X + result.Regions[1].Width

	stageLeft := result.Regions[0].X
	stageRight := rightEnd

	stageWidth := stageRight - stageLeft
	// Stage should be ≤ 78 (may be less due to gutter)
	if stageWidth > 78 {
		t.Errorf("total stage width %d exceeds lineWidth 78", stageWidth)
	}

	// Right column should start after left + gutter
	if rightStart <= leftEnd {
		t.Errorf("right column X %d should be > left end %d", rightStart, leftEnd)
	}
}

func TestAspectRatioHeightCap(t *testing.T) {
	// With lineWidth 78 and 16:9, height should be capped.
	// maxH = 78 * 9 / (2 * 16) = 21
	custom := model.CustomLayout{
		Columns: []int{100},
		Rows:    []int{100},
	}
	vp := Viewport{Width: 160, Height: 40}

	result := computeGrid(custom, "auto", vp, 0, 0, 78, "16:9")

	if result.Regions[0].Width != 78 {
		t.Errorf("width = %d, want 78", result.Regions[0].Width)
	}

	wantH := 78 * 9 / (2 * 16) // = 21
	if result.Regions[0].Height != wantH {
		t.Errorf("height = %d, want %d (aspect-ratio capped)", result.Regions[0].Height, wantH)
	}

	// Content should be vertically centered
	// Original: padY=1, usableH=40-2-1=37, maxH=21, extra=16, padY=1+8=9
	if result.Regions[0].Y != 9 {
		t.Errorf("Y = %d, want 9 (vertically centered)", result.Regions[0].Y)
	}
}

func TestNoHeightCapWithoutLineWidth(t *testing.T) {
	// Without lineWidth cap, aspect ratio height capping should not reduce
	// height from what the viewport provides (aspect padding handles it).
	custom := model.CustomLayout{
		Columns: []int{100},
	}
	vp := Viewport{Width: 100, Height: 30}

	// No lineWidth cap, no aspect padding
	result := computeGrid(custom, "auto", vp, 0, 0, 0, "16:9")

	// usableW = 100-4 = 96, maxH = 96*9/32 = 27
	// usableH = 30-2-1 = 27 → exactly at the cap, so no extra padding
	if result.Regions[0].Height != 27 {
		t.Errorf("height = %d, want 27", result.Regions[0].Height)
	}
}
