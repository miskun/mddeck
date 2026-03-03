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
	result := ComputeLayout(slide, vp)

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
	result := ComputeLayout(slide, vp)

	if result.Mode != model.LayoutTerminal {
		t.Errorf("mode = %q, want %q", result.Mode, model.LayoutTerminal)
	}
}

func TestAutoDetectTwoCol(t *testing.T) {
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
	result := ComputeLayout(slide, vp)

	if result.Mode != model.LayoutTwoCol {
		t.Errorf("mode = %q, want %q", result.Mode, model.LayoutTwoCol)
	}
}

func TestTwoColRatio(t *testing.T) {
	slide := &model.Slide{
		Meta: model.SlideMeta{
			Layout: model.LayoutTwoCol,
			Ratio:  "70/30",
			Align:  model.AlignTop,
		},
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 2, Raw: "Left"},
			{Type: model.BlockHeading, Level: 2, Raw: "Right"},
		},
	}

	vp := Viewport{Width: 100, Height: 24}
	result := ComputeLayout(slide, vp)

	if len(result.Regions) != 2 {
		t.Fatalf("regions = %d, want 2", len(result.Regions))
	}

	leftW := result.Regions[0].Width
	rightW := result.Regions[1].Width

	// 70/30 of (100-2 gutter) = 98 usable → ~68/30
	if leftW < 60 || leftW > 75 {
		t.Errorf("left width = %d, expected ~68", leftW)
	}
	if rightW < 20 || rightW > 35 {
		t.Errorf("right width = %d, expected ~30", rightW)
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
