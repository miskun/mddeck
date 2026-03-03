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
	// Use --- slide breaks to prevent header-based splitting
	input := `# H1

## H2

### H3

---

Slide two`

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

func TestHeaderBasedSplitting(t *testing.T) {
	input := `# Introduction

Welcome to the talk.

## First Topic

Some content here.

## Second Topic

More content here.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Deepest header is ## (h2), so splits happen on # and ##.
	// # Introduction → slide 1
	// ## First Topic → slide 2
	// ## Second Topic → slide 3
	if len(deck.Slides) != 3 {
		for i, s := range deck.Slides {
			t.Logf("slide %d: %d blocks", i, len(s.Blocks))
			for _, b := range s.Blocks {
				t.Logf("  type=%d level=%d raw=%q", b.Type, b.Level, b.Raw)
			}
		}
		t.Fatalf("got %d slides, want 3", len(deck.Slides))
	}

	// First slide should contain the H1 heading
	if deck.Slides[0].Blocks[0].Type != model.BlockHeading || deck.Slides[0].Blocks[0].Level != 1 {
		t.Errorf("slide 1 should start with H1, got type=%d level=%d", deck.Slides[0].Blocks[0].Type, deck.Slides[0].Blocks[0].Level)
	}

	// Second slide should start with H2
	if deck.Slides[1].Blocks[0].Type != model.BlockHeading || deck.Slides[1].Blocks[0].Level != 2 {
		t.Errorf("slide 2 should start with H2, got type=%d level=%d", deck.Slides[1].Blocks[0].Type, deck.Slides[1].Blocks[0].Level)
	}

	// Third slide should contain H2 and H3
	if deck.Slides[2].Blocks[0].Type != model.BlockHeading || deck.Slides[2].Blocks[0].Level != 2 {
		t.Errorf("slide 3 should start with H2, got type=%d level=%d", deck.Slides[2].Blocks[0].Type, deck.Slides[2].Blocks[0].Level)
	}
}

func TestHeaderSplittingNotUsedWithHR(t *testing.T) {
	// When --- slide breaks exist, header splitting should NOT activate
	input := `# Slide One

Some text.

---

# Slide Two

More text.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 2 {
		t.Fatalf("got %d slides, want 2", len(deck.Slides))
	}
}

func TestHeaderSplittingOnlyH1(t *testing.T) {
	// Only H1 headers — deepest is 1, so each H1 is a slide
	input := `# First

Hello.

# Second

World.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 2 {
		t.Fatalf("got %d slides, want 2", len(deck.Slides))
	}
}

func TestHeaderSplittingNoHeaders(t *testing.T) {
	// No headers at all — should remain a single slide
	input := `Just some text.

More text here.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(deck.Slides))
	}
}

func TestFrontmatterAsSlideBreak(t *testing.T) {
	// Mix of header-based splitting and frontmatter blocks.
	// No --- slide breaks, so header splitting activates.
	// Frontmatter blocks should also start new slides.
	input := `## Slide One

First content.

## Slide Two

Second content.

---
layout: two-col
ratio: "50/50"
---

## Left Column

Left text.

## Right Column

Right text.

## Slide Four

Back to normal.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect: Slide One, Slide Two, two-col slide (Left+Right), Slide Four
	if len(deck.Slides) != 4 {
		t.Fatalf("got %d slides, want 4", len(deck.Slides))
	}

	// Slide 1: normal header slide
	if len(deck.Slides[0].Blocks) == 0 {
		t.Error("slide 1 should have blocks")
	}

	// Slide 3: should have two-col layout from frontmatter
	if deck.Slides[2].Meta.Layout != model.LayoutTwoCol {
		t.Errorf("slide 3 layout = %v, want two-col", deck.Slides[2].Meta.Layout)
	}

	// Slide 4: back to normal
	if len(deck.Slides[3].Blocks) == 0 {
		t.Error("slide 4 should have blocks")
	}
}

func TestFrontmatterOnlyNoHeaders(t *testing.T) {
	// File with frontmatter blocks but no headers at all.
	// Should still split on frontmatter.
	input := `Some introductory text.

---
layout: terminal
---

Terminal content here.

---
layout: center
---

Centered content.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 3 {
		t.Fatalf("got %d slides, want 3", len(deck.Slides))
	}

	if deck.Slides[1].Meta.Layout != model.LayoutTerminal {
		t.Errorf("slide 2 layout = %v, want terminal", deck.Slides[1].Meta.Layout)
	}
}

func TestAutoSplitFalse(t *testing.T) {
	// autosplit: false should absorb all headers until the next
	// frontmatter block (which acts as a resume marker).
	input := `## Intro

Welcome!

---
autosplit: false
---

## Heading Examples

# H1

## H2

### H3

All headings on one slide.

---
autosplit: true
---

## Next Slide

Content after the no-split zone.

## Another Slide

More content.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect: Intro, Heading Examples (autosplit:false zone),
	//         Next Slide, Another Slide
	// The autosplit:true resume marker should be an empty slide that gets filtered.
	if len(deck.Slides) != 4 {
		for i, s := range deck.Slides {
			title := ""
			for _, b := range s.Blocks {
				if b.Type == model.BlockHeading {
					title = b.Raw
					break
				}
			}
			t.Logf("  slide %d: blocks=%d autosplit=%v title=%q", i+1, len(s.Blocks), s.Meta.GetAutoSplit(), title)
		}
		t.Fatalf("got %d slides, want 4", len(deck.Slides))
	}

	// Slide 2: autosplit=false, should contain 5 blocks (H2, H1, H2, H3, P)
	if deck.Slides[1].Meta.GetAutoSplit() != false {
		t.Errorf("slide 2 autosplit = %v, want false", deck.Slides[1].Meta.GetAutoSplit())
	}
	if len(deck.Slides[1].Blocks) != 5 {
		t.Errorf("slide 2 blocks = %d, want 5", len(deck.Slides[1].Blocks))
	}

	// Slide 3: should be "Next Slide"
	if deck.Slides[2].Blocks[0].Raw != "Next Slide" {
		t.Errorf("slide 3 title = %q, want %q", deck.Slides[2].Blocks[0].Raw, "Next Slide")
	}

	// Slide 4: should be "Another Slide"
	if deck.Slides[3].Blocks[0].Raw != "Another Slide" {
		t.Errorf("slide 4 title = %q, want %q", deck.Slides[3].Blocks[0].Raw, "Another Slide")
	}
}

func TestNestedUnorderedList(t *testing.T) {
	input := `- Top
  - Middle
    - Deep
- Another top`

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
	if len(blocks[0].Lines) != 4 {
		t.Errorf("items = %d, want 4", len(blocks[0].Lines))
	}
	// Check depths
	expected := []string{"0:Top", "1:Middle", "2:Deep", "0:Another top"}
	for i, want := range expected {
		if blocks[0].Lines[i] != want {
			t.Errorf("item[%d] = %q, want %q", i, blocks[0].Lines[i], want)
		}
	}
}

func TestTaskList(t *testing.T) {
	input := `- [x] Done item
- [ ] Todo item
- [X] Also done`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockTaskList {
		t.Errorf("type = %v, want BlockTaskList", blocks[0].Type)
	}
	if len(blocks[0].Lines) != 3 {
		t.Errorf("items = %d, want 3", len(blocks[0].Lines))
	}
	expected := []string{"0:1:Done item", "0:0:Todo item", "0:1:Also done"}
	for i, want := range expected {
		if blocks[0].Lines[i] != want {
			t.Errorf("item[%d] = %q, want %q", i, blocks[0].Lines[i], want)
		}
	}
}

func TestTable(t *testing.T) {
	input := `| Name | Value |
|------|-------|
| foo  | 42    |
| bar  | 99    |`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockTable {
		t.Errorf("type = %v, want BlockTable", blocks[0].Type)
	}
	// Separator row is skipped, so 3 lines: header + 2 data rows
	if len(blocks[0].Lines) != 3 {
		t.Errorf("lines = %d, want 3", len(blocks[0].Lines))
	}
}

func TestAlert(t *testing.T) {
	input := `> [!WARNING]
> Be careful with this.
> It could break things.`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockAlert {
		t.Errorf("type = %v, want BlockAlert", blocks[0].Type)
	}
	if blocks[0].Language != "WARNING" {
		t.Errorf("alert type = %q, want %q", blocks[0].Language, "WARNING")
	}
	if len(blocks[0].Lines) != 2 {
		t.Errorf("lines = %d, want 2", len(blocks[0].Lines))
	}
}

func TestLineBreaks(t *testing.T) {
	input := `Line one\
Line two`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks := deck.Slides[0].Blocks
	if len(blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockParagraph {
		t.Errorf("type = %v, want BlockParagraph", blocks[0].Type)
	}
	// Should contain a newline (hard break)
	if !strings.Contains(blocks[0].Raw, "\n") {
		t.Errorf("raw = %q, want embedded newline", blocks[0].Raw)
	}
}
