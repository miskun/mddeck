package render

import (
	"strings"
	"testing"

	a "github.com/miskun/mddeck/internal/ansi"
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
