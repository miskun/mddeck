package render

import (
	"strings"
	"testing"

	a "github.com/miskun/mddeck/internal/ansi"
	"github.com/miskun/mddeck/internal/layout"
	"github.com/miskun/mddeck/internal/model"
	"github.com/miskun/mddeck/internal/theme"
)

func testRenderer() *Renderer {
	deck := &model.Deck{}
	return NewRenderer(deck, theme.Default)
}

// TestStripInlineMarkdown_CodeProtectsContent verifies that inline markers
// inside backtick code spans are preserved (not stripped).
func TestStripInlineMarkdown_CodeProtectsContent(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"`~~text~~`", "~~text~~"},
		{"`**bold**`", "**bold**"},
		{"`*italic*`", "*italic*"},
		{"normal `~~kept~~` text", "normal ~~kept~~ text"},
		{"~~struck~~", "struck"},
		{"**bold** and `code`", "bold and code"},
	}
	for _, tt := range tests {
		got := stripInlineMarkdown(tt.input)
		if got != tt.want {
			t.Errorf("stripInlineMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestApplyInlineStyles_CodeProtectsContent verifies that bold/italic/
// strikethrough markers inside code spans are rendered literally, not styled.
func TestApplyInlineStyles_CodeProtectsContent(t *testing.T) {
	r := testRenderer()

	input := "`~~text~~`"
	result := r.applyInlineStyles(input)

	// The output must NOT contain the strikethrough ANSI escape
	if strings.Contains(result, a.Strikethrough) {
		t.Errorf("applyInlineStyles(%q) applied strikethrough inside code span", input)
	}
	// The ~~ markers must be present in the output (rendered literally)
	if !strings.Contains(result, "~~text~~") {
		t.Errorf("applyInlineStyles(%q) did not preserve ~~ inside code span; got %q", input, result)
	}
	// The code styling must be applied
	if !strings.Contains(result, r.Theme.CodeFg) {
		t.Errorf("applyInlineStyles(%q) missing code styling", input)
	}
}

// TestApplyInlineStyles_BoldInsideCodePreserved verifies **bold** inside
// backticks is rendered as literal text.
func TestApplyInlineStyles_BoldInsideCodePreserved(t *testing.T) {
	r := testRenderer()

	result := r.applyInlineStyles("`**bold**`")

	if strings.Contains(result, a.Bold) {
		t.Error("applyInlineStyles applied bold inside code span")
	}
	if !strings.Contains(result, "**bold**") {
		t.Errorf("literal **bold** not preserved inside code span; got %q", result)
	}
}

// TestTitleStyleFallback verifies that empty TitleStyle falls back to H1Style.
func TestTitleStyleFallback(t *testing.T) {
	th := theme.Theme{
		H1Style: a.Bold + a.FgCyan,
		H2Style: a.Bold + a.FgBlue,
	}
	if got := th.GetTitleStyle(); got != th.H1Style {
		t.Errorf("GetTitleStyle() = %q, want H1Style %q", got, th.H1Style)
	}
	if got := th.GetSlideTitleStyle(); got != th.H2Style {
		t.Errorf("GetSlideTitleStyle() = %q, want H2Style %q", got, th.H2Style)
	}

	// When TitleStyle is set, it should be returned instead of H1Style
	th.TitleStyle = a.Bold + a.FgRed
	if got := th.GetTitleStyle(); got != th.TitleStyle {
		t.Errorf("GetTitleStyle() with set value = %q, want %q", got, th.TitleStyle)
	}
}

// TestRenderHeadingStyled verifies that overrides substitute correctly.
func TestRenderHeadingStyled(t *testing.T) {
	r := testRenderer()
	customStyle := a.Bold + a.FgRed
	ov := headingOverrides{
		H1:       customStyle,
		MarginH1: -1,
		MarginH2: -1,
		MarginH3: -1,
	}

	block := model.Block{Type: model.BlockHeading, Level: 1, Raw: "Title"}
	lines := r.renderHeadingStyled(block, 80, ov)

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], customStyle) {
		t.Errorf("heading should use custom style %q, got %q", customStyle, lines[0])
	}
	// Should NOT contain the theme's default H1Style (unless it happens to be the same)
	if customStyle != r.Theme.H1Style && strings.Contains(lines[0], r.Theme.H1Style) {
		t.Errorf("heading should not contain theme H1Style when overridden")
	}
}

// TestRenderCenteredUsesTitleStyle verifies title slide uses TitleStyle, not H1Style.
func TestRenderCenteredUsesTitleStyle(t *testing.T) {
	th := theme.Default
	// Make TitleStyle and H1Style distinguishable
	th.TitleStyle = a.Bold + a.FgRed
	th.H1Style = a.Bold + a.FgGreen

	deck := &model.Deck{}
	r := NewRenderer(deck, th)

	slide := &model.Slide{
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 1, Raw: "My Title"},
		},
		Meta: model.SlideMeta{Layout: "title"},
	}

	region := layout.Region{X: 0, Y: 0, Width: 80, Height: 24}
	scr := newScreenBuf(80, 24)
	r.renderCentered(slide, region, scr)

	output := strings.Join(scr.Lines(), "\n")
	if !strings.Contains(output, th.TitleStyle) {
		t.Error("renderCentered should use TitleStyle for H1")
	}
	if strings.Contains(output, th.H1Style) {
		t.Error("renderCentered should not use H1Style when TitleStyle is set")
	}
}

// TestRenderGridTitleRowUsesSlideTitleStyle verifies grid title region uses SlideTitleStyle.
func TestRenderGridTitleRowUsesSlideTitleStyle(t *testing.T) {
	th := theme.Default
	// Make SlideTitleStyle distinguishable
	th.SlideTitleStyle = a.Bold + a.FgMagenta
	th.H1Style = a.Bold + a.FgGreen
	th.H2Style = a.Bold + a.FgGreen

	deck := &model.Deck{}
	r := NewRenderer(deck, th)

	slide := &model.Slide{
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 2, Raw: "Slide Title"},
			{Type: model.BlockParagraph, Raw: "Body content"},
		},
		Meta: model.SlideMeta{Layout: "title-body"},
	}

	lr := layout.LayoutResult{
		Mode: model.LayoutTitleBody,
		Regions: []layout.Region{
			{X: 0, Y: 0, Width: 80, Height: 3},
			{X: 0, Y: 4, Width: 80, Height: 20},
		},
		HasTitleRow: true,
	}

	scr := newScreenBuf(80, 24)
	r.renderGrid(slide, lr, scr)

	output := strings.Join(scr.Lines(), "\n")
	if !strings.Contains(output, th.SlideTitleStyle) {
		t.Error("renderGrid title row should use SlideTitleStyle")
	}
}

// TestHeadingLevel4Plus verifies H4+ uses body foreground + bold.
func TestHeadingLevel4Plus(t *testing.T) {
	r := testRenderer()
	expected := a.Bold + r.Theme.Fg

	for level := 4; level <= 6; level++ {
		block := model.Block{Type: model.BlockHeading, Level: level, Raw: "Heading"}
		lines := r.renderHeading(block, 80)
		if len(lines) != 1 {
			t.Fatalf("H%d: expected 1 line, got %d", level, len(lines))
		}
		if !strings.HasPrefix(lines[0], expected) {
			t.Errorf("H%d should start with Bold+Fg (%q), got %q", level, expected, lines[0])
		}
		// Should not contain H3Style
		if strings.Contains(lines[0], r.Theme.H3Style) && r.Theme.H3Style != expected {
			t.Errorf("H%d should not use H3Style", level)
		}
	}
}

// TestHeadingMarginDefault verifies default margin produces 1 blank line after heading.
func TestHeadingMarginDefault(t *testing.T) {
	r := testRenderer()
	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 1, Raw: "Title"},
		{Type: model.BlockParagraph, Raw: "Body text"},
	}

	lines := r.renderBlocks(blocks, 80)

	// Expect: heading line, blank line, body line
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[1] != "" {
		t.Errorf("expected blank line after heading, got %q", lines[1])
	}
}

// TestHeadingMarginZero verifies margin 0 produces no blank line after heading.
func TestHeadingMarginZero(t *testing.T) {
	th := theme.Default
	th.H1Margin = 0

	deck := &model.Deck{}
	r := NewRenderer(deck, th)

	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 1, Raw: "Title"},
		{Type: model.BlockParagraph, Raw: "Body text"},
	}

	lines := r.renderBlocks(blocks, 80)

	// With margin 0: heading line, body line (no blank line between)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines with margin 0, got %d: %v", len(lines), lines)
	}
	// Second line should be the body, not blank
	stripped := a.StripAll(lines[1])
	if stripped != "Body text" {
		t.Errorf("second line should be body text, got %q", stripped)
	}
}

// TestTitleMarginInCentered verifies TitleMargin controls spacing on title slides.
func TestTitleMarginInCentered(t *testing.T) {
	th := theme.Default
	th.TitleMargin = 2

	deck := &model.Deck{}
	r := NewRenderer(deck, th)

	slide := &model.Slide{
		Blocks: []model.Block{
			{Type: model.BlockHeading, Level: 1, Raw: "Title"},
			{Type: model.BlockParagraph, Raw: "Subtitle text"},
		},
		Meta: model.SlideMeta{Layout: "title"},
	}

	region := layout.Region{X: 0, Y: 0, Width: 80, Height: 24}
	scr := newScreenBuf(80, 24)
	r.renderCentered(slide, region, scr)

	// Collect non-empty lines from screen
	allLines := scr.Lines()
	var contentLines []string
	for _, l := range allLines {
		contentLines = append(contentLines, l)
	}

	// Find the heading and count blank lines after it
	headingIdx := -1
	for i, l := range contentLines {
		if strings.Contains(l, "Title") && !strings.Contains(l, "Subtitle") {
			headingIdx = i
			break
		}
	}
	if headingIdx < 0 {
		t.Fatal("could not find heading line")
	}

	// Count consecutive blank/empty lines after heading
	blankCount := 0
	for i := headingIdx + 1; i < len(contentLines); i++ {
		if strings.TrimSpace(contentLines[i]) == "" {
			blankCount++
		} else {
			break
		}
	}
	if blankCount != 2 {
		t.Errorf("expected 2 blank lines after title (TitleMargin=2), got %d", blankCount)
	}
}

// TestTitleBodySlideTitleMargin1 verifies that SlideTitleMargin=1 produces
// exactly 1 empty line between a heading and the next block when rendered
// with SlideTitleStyle overrides (as happens in the title row of grid layouts).
func TestTitleBodySlideTitleMargin1(t *testing.T) {
	th := theme.Default
	th.SlideTitleMargin = 1

	deck := &model.Deck{}
	r := NewRenderer(deck, th)

	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 2, Raw: "Grid Title"},
		{Type: model.BlockParagraph, Raw: "Body paragraph"},
	}

	ov := headingOverrides{
		H1:       r.Theme.GetSlideTitleStyle(),
		H2:       r.Theme.GetSlideTitleStyle(),
		H3:       r.Theme.GetSlideTitleStyle(),
		MarginH1: th.SlideTitleMargin,
		MarginH2: th.SlideTitleMargin,
		MarginH3: th.SlideTitleMargin,
	}
	lines := r.renderBlocksStyled(blocks, 80, ov)

	// Expect: heading, 1 blank line, body
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (heading + 1 blank + body), got %d: %v", len(lines), lines)
	}
	if lines[1] != "" {
		t.Errorf("expected blank line at index 1, got %q", lines[1])
	}
}

// TestTitleBodySlideTitleMargin2 verifies that SlideTitleMargin=2 produces
// exactly 2 empty lines between a heading and the next block when rendered
// with SlideTitleStyle overrides.
func TestTitleBodySlideTitleMargin2(t *testing.T) {
	th := theme.Default
	th.SlideTitleMargin = 2

	deck := &model.Deck{}
	r := NewRenderer(deck, th)

	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 2, Raw: "Grid Title"},
		{Type: model.BlockParagraph, Raw: "Body paragraph"},
	}

	ov := headingOverrides{
		H1:       r.Theme.GetSlideTitleStyle(),
		H2:       r.Theme.GetSlideTitleStyle(),
		H3:       r.Theme.GetSlideTitleStyle(),
		MarginH1: th.SlideTitleMargin,
		MarginH2: th.SlideTitleMargin,
		MarginH3: th.SlideTitleMargin,
	}
	lines := r.renderBlocksStyled(blocks, 80, ov)

	// Expect: heading, 2 blank lines, body
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (heading + 2 blanks + body), got %d: %v", len(lines), lines)
	}
	if lines[1] != "" || lines[2] != "" {
		t.Errorf("expected blank lines at indices 1-2, got %q and %q", lines[1], lines[2])
	}
}
