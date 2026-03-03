package parser

import (
	"strings"
	"testing"

	"github.com/miskun/mddeck/internal/model"
)

func TestParseDeckFrontmatter(t *testing.T) {
	input := `---
title: "My Talk"
theme: "dark"
wrap: false
tabSize: 4
---

# First Slide
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deck.Meta.Title != "My Talk" {
		t.Errorf("title = %q, want %q", deck.Meta.Title, "My Talk")
	}
	if deck.Meta.Theme != "dark" {
		t.Errorf("theme = %q, want %q", deck.Meta.Theme, "dark")
	}
	if deck.Meta.GetWrap() != false {
		t.Errorf("wrap = %v, want false", deck.Meta.GetWrap())
	}
	if deck.Meta.GetTabSize() != 4 {
		t.Errorf("tabSize = %d, want 4", deck.Meta.GetTabSize())
	}
}

func TestParseDeckFrontmatterDefaults(t *testing.T) {
	input := `# Just a slide`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deck.Meta.Theme != "default" {
		t.Errorf("theme = %q, want %q", deck.Meta.Theme, "default")
	}
	if deck.Meta.GetWrap() != true {
		t.Errorf("wrap = %v, want true", deck.Meta.GetWrap())
	}
	if deck.Meta.GetTabSize() != 2 {
		t.Errorf("tabSize = %d, want 2", deck.Meta.GetTabSize())
	}
	if deck.Meta.GetSafeAnsi() != true {
		t.Errorf("safeAnsi = %v, want true", deck.Meta.GetSafeAnsi())
	}
}

func TestSlideBreaks(t *testing.T) {
	input := `# Slide 1

---

# Slide 2

---

# Slide 3`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 3 {
		t.Fatalf("slides = %d, want 3", len(deck.Slides))
	}
}

func TestNotASlideBreak(t *testing.T) {
	// --- without blank lines before/after is NOT a slide break
	input := `# Slide 1
---
More text`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1 (--- should not be a slide break)", len(deck.Slides))
	}
}

func TestSpeakerNotes(t *testing.T) {
	input := `# Architecture

- Access boundary
- Authorization-aware AI

???
Mention Datadog comparison here.`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1", len(deck.Slides))
	}

	notes := deck.Slides[0].Notes
	if !strings.Contains(notes, "Mention Datadog") {
		t.Errorf("notes = %q, expected to contain 'Mention Datadog'", notes)
	}
}

func TestSlideFrontmatter(t *testing.T) {
	input := `# Intro

---

---
layout: two-col
ratio: "60/40"
align: top
---

## Left

## Right`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 2 {
		t.Fatalf("slides = %d, want 2", len(deck.Slides))
	}

	slide := deck.Slides[1]
	if slide.Meta.Layout != model.LayoutTwoCol {
		t.Errorf("layout = %q, want %q", slide.Meta.Layout, model.LayoutTwoCol)
	}
	if slide.Meta.Ratio != "60/40" {
		t.Errorf("ratio = %q, want %q", slide.Meta.Ratio, "60/40")
	}
}

func TestHeadings(t *testing.T) {
	input := `# H1

## H2

### H3`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 3 {
		t.Fatalf("blocks = %d, want 3", len(blocks))
	}

	tests := []struct {
		level int
		text  string
	}{
		{1, "H1"},
		{2, "H2"},
		{3, "H3"},
	}

	for i, tt := range tests {
		if blocks[i].Type != model.BlockHeading {
			t.Errorf("block[%d].Type = %v, want BlockHeading", i, blocks[i].Type)
		}
		if blocks[i].Level != tt.level {
			t.Errorf("block[%d].Level = %d, want %d", i, blocks[i].Level, tt.level)
		}
		if blocks[i].Raw != tt.text {
			t.Errorf("block[%d].Raw = %q, want %q", i, blocks[i].Raw, tt.text)
		}
	}
}

func TestUnorderedList(t *testing.T) {
	input := `- Item one
- Item two
- Item three`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockUnorderedList {
		t.Errorf("type = %v, want BlockUnorderedList", blocks[0].Type)
	}
	if len(blocks[0].Lines) != 3 {
		t.Errorf("items = %d, want 3", len(blocks[0].Lines))
	}
}

func TestOrderedList(t *testing.T) {
	input := `1. First
2. Second
3. Third`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockOrderedList {
		t.Errorf("type = %v, want BlockOrderedList", blocks[0].Type)
	}
	if len(blocks[0].Lines) != 3 {
		t.Errorf("items = %d, want 3", len(blocks[0].Lines))
	}
}

func TestFencedCodeBlock(t *testing.T) {
	input := "# Code\n\n```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```"

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	found := false
	for _, b := range blocks {
		if b.Type == model.BlockFencedCode {
			found = true
			if b.Language != "go" {
				t.Errorf("language = %q, want %q", b.Language, "go")
			}
			if !strings.Contains(b.Raw, "func main()") {
				t.Errorf("code block should contain 'func main()'")
			}
		}
	}
	if !found {
		t.Error("no fenced code block found")
	}
}

func TestArtBlocks(t *testing.T) {
	tests := []struct {
		lang     string
		expected model.BlockType
	}{
		{"ansi", model.BlockANSIArt},
		{"ascii", model.BlockASCIIArt},
		{"braille", model.BlockBrailleArt},
	}

	for _, tt := range tests {
		input := "```" + tt.lang + "\nart content\n```"
		deck, err := Parse(input)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tt.lang, err)
		}
		blocks := deck.Slides[0].Blocks
		if len(blocks) == 0 {
			t.Fatalf("no blocks for %s", tt.lang)
		}
		if blocks[0].Type != tt.expected {
			t.Errorf("%s: type = %v, want %v", tt.lang, blocks[0].Type, tt.expected)
		}
	}
}

func TestBlockquote(t *testing.T) {
	input := `> This is a quote
> Second line`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockBlockquote {
		t.Errorf("type = %v, want BlockBlockquote", blocks[0].Type)
	}
}

func TestMultipleSlidesWithNotes(t *testing.T) {
	input := `---
title: "Test Deck"
---

# Slide 1

Content here.

???
Note for slide 1.

---

# Slide 2

More content.

???
Note for slide 2.`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 2 {
		t.Fatalf("slides = %d, want 2", len(deck.Slides))
	}

	if !strings.Contains(deck.Slides[0].Notes, "Note for slide 1") {
		t.Errorf("slide 1 notes = %q", deck.Slides[0].Notes)
	}
	if !strings.Contains(deck.Slides[1].Notes, "Note for slide 2") {
		t.Errorf("slide 2 notes = %q", deck.Slides[1].Notes)
	}
}

func TestCRLFNormalization(t *testing.T) {
	input := "# Slide 1\r\n\r\n---\r\n\r\n# Slide 2"

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 2 {
		t.Fatalf("slides = %d, want 2", len(deck.Slides))
	}
}

func TestEmptyDeck(t *testing.T) {
	input := ""

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1 (empty deck should have 1 empty slide)", len(deck.Slides))
	}
}

func TestUnknownFrontmatterKeysIgnored(t *testing.T) {
	input := `---
title: "Test"
unknownKey: "should be ignored"
anotherUnknown: 42
---

# Slide`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v (unknown keys should be ignored)", err)
	}

	if deck.Meta.Title != "Test" {
		t.Errorf("title = %q, want %q", deck.Meta.Title, "Test")
	}
}

func TestHorizontalRuleInsideSlide(t *testing.T) {
	input := `# Heading
Some text
---
More text`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The --- without blank lines should remain in the same slide
	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1", len(deck.Slides))
	}

	// Should have a horizontal rule block somewhere
	found := false
	for _, b := range deck.Slides[0].Blocks {
		if b.Type == model.BlockHorizontalRule {
			found = true
		}
	}
	if !found {
		t.Error("expected horizontal rule block inside slide")
	}
}
