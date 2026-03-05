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
layout: title-cols-2
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
	if slide.Meta.Layout != model.LayoutTitleCols2 {
		t.Errorf("layout = %q, want %q", slide.Meta.Layout, model.LayoutTitleCols2)
	}
	if slide.Meta.Ratio != "60/40" {
		t.Errorf("ratio = %q, want %q", slide.Meta.Ratio, "60/40")
	}
}

func TestHeadings(t *testing.T) {
	// Headers are split into separate slides by header-based splitting.
	// The deepest header level is ### (h3), so #, ##, ### each start a slide.
	input := `# H1

## H2

### H3
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 3 {
		for i, s := range deck.Slides {
			t.Logf("slide %d: %d blocks", i, len(s.Blocks))
			for _, b := range s.Blocks {
				t.Logf("  type=%d level=%d raw=%q", b.Type, b.Level, b.Raw)
			}
		}
		t.Fatalf("slides = %d, want 3", len(deck.Slides))
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
		block := deck.Slides[i].Blocks[0]
		if block.Type != model.BlockHeading {
			t.Errorf("slide[%d] block.Type = %v, want BlockHeading", i, block.Type)
		}
		if block.Level != tt.level {
			t.Errorf("slide[%d] block.Level = %d, want %d", i, block.Level, tt.level)
		}
		if block.Raw != tt.text {
			t.Errorf("slide[%d] block.Raw = %q, want %q", i, block.Raw, tt.text)
		}
	}
}

func TestUnorderedList(t *testing.T) {
	input := `---
incrementalLists: false
---
- Item one
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
	input := `---
incrementalLists: false
---
1. First
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

func TestHeaderSplittingWithHR(t *testing.T) {
	// Each --- chunk is independently header-split.
	// Both chunks have only H1 (deepest=1), so each H1 stays as one slide.
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

func TestMixedHRAndHeaderSplitting(t *testing.T) {
	// First chunk has # and ## headers — header splitting produces 3 slides.
	// Second chunk (after ---) has no headers — stays as 1 slide.
	// Total: 4 slides.
	input := `# Title

Intro text.

## Topic A

Content A.

## Topic B

Content B.

---

Just a plain slide.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 4 {
		for i, s := range deck.Slides {
			t.Logf("slide %d: %d blocks", i, len(s.Blocks))
			for _, b := range s.Blocks {
				t.Logf("  type=%d level=%d raw=%q", b.Type, b.Level, b.Raw)
			}
		}
		t.Fatalf("got %d slides, want 4", len(deck.Slides))
	}

	// Slide 0: # Title + paragraph
	if deck.Slides[0].Blocks[0].Raw != "Title" {
		t.Errorf("slide 0 heading = %q, want %q", deck.Slides[0].Blocks[0].Raw, "Title")
	}
	// Slide 1: ## Topic A + paragraph
	if deck.Slides[1].Blocks[0].Raw != "Topic A" {
		t.Errorf("slide 1 heading = %q, want %q", deck.Slides[1].Blocks[0].Raw, "Topic A")
	}
	// Slide 2: ## Topic B + paragraph
	if deck.Slides[2].Blocks[0].Raw != "Topic B" {
		t.Errorf("slide 2 heading = %q, want %q", deck.Slides[2].Blocks[0].Raw, "Topic B")
	}
	// Slide 3: plain text
	if deck.Slides[3].Blocks[0].Raw != "Just a plain slide." {
		t.Errorf("slide 3 text = %q, want %q", deck.Slides[3].Blocks[0].Raw, "Just a plain slide.")
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
	// title-cols-2 absorbs 3 headers (1 title + 2 columns).
	input := `## Slide One

First content.

## Slide Two

Second content.

---
layout: title-cols-2
ratio: "50/50"
---

## Title Row

## Left Column

Left text.

## Right Column

Right text.

## Slide Five

Back to normal.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect: Slide One, Slide Two, title-cols-2 slide (Title+Left+Right), Slide Five
	if len(deck.Slides) != 4 {
		for i, s := range deck.Slides {
			t.Logf("slide %d: blocks=%d layout=%v", i, len(s.Blocks), s.Meta.Layout)
			for _, b := range s.Blocks {
				t.Logf("  type=%d level=%d raw=%q", b.Type, b.Level, b.Raw)
			}
		}
		t.Fatalf("got %d slides, want 4", len(deck.Slides))
	}

	// Slide 1: normal header slide
	if len(deck.Slides[0].Blocks) == 0 {
		t.Error("slide 1 should have blocks")
	}

	// Slide 3: should have title-cols-2 layout from frontmatter
	if deck.Slides[2].Meta.Layout != model.LayoutTitleCols2 {
		t.Errorf("slide 3 layout = %v, want title-cols-2", deck.Slides[2].Meta.Layout)
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
layout: blank
---

Blank content here.

---
layout: section
---

Section content.
`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Slides) != 3 {
		t.Fatalf("got %d slides, want 3", len(deck.Slides))
	}

	if deck.Slides[1].Meta.Layout != model.LayoutBlank {
		t.Errorf("slide 2 layout = %v, want blank", deck.Slides[1].Meta.Layout)
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
	input := `---
incrementalLists: false
---
- Top
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
	input := `---
incrementalLists: false
---
- [x] Done item
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
	// Header separator is skipped, so 3 lines: header + 2 data rows
	if len(blocks[0].Lines) != 3 {
		t.Errorf("lines = %d, want 3", len(blocks[0].Lines))
	}
}

func TestTableMidSeparator(t *testing.T) {
	input := `| Name | Role |
|------|------|
| Alice| Dev  |
| Bob  | QA   |
|------|------|
| Carol| PM   |`

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
	// Header + 2 data + mid-separator + 1 data = 5 lines
	if len(blocks[0].Lines) != 5 {
		t.Errorf("lines = %d, want 5", len(blocks[0].Lines))
	}
	// The mid-table separator should be the 4th line (index 3)
	if blocks[0].Lines[3] != "|------|------|" {
		t.Errorf("line[3] = %q, want separator row", blocks[0].Lines[3])
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

func TestDeckFrontmatterCustomLayouts(t *testing.T) {
	input := `---
aspect: "16:9"
layouts:
  sidebar:
    columns: [30, 70]
    gutter: 3
    padX: 4
    padY: 2
  grid2x2:
    columns: [50, 50]
    rows: [50, 50]
---

# Slide 1

Hello`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deck.Meta.Aspect != "16:9" {
		t.Errorf("aspect = %q, want %q", deck.Meta.Aspect, "16:9")
	}
	if len(deck.Meta.Layouts) != 2 {
		t.Fatalf("layouts count = %d, want 2", len(deck.Meta.Layouts))
	}

	sb, ok := deck.Meta.Layouts["sidebar"]
	if !ok {
		t.Fatal("missing sidebar layout")
	}
	if len(sb.Columns) != 2 || sb.Columns[0] != 30 || sb.Columns[1] != 70 {
		t.Errorf("sidebar columns = %v, want [30 70]", sb.Columns)
	}
	if sb.GetGutter() != 3 {
		t.Errorf("sidebar gutter = %d, want 3", sb.GetGutter())
	}
	if sb.GetPadX() != 4 {
		t.Errorf("sidebar padx = %d, want 4", sb.GetPadX())
	}
	if sb.GetPadY() != 2 {
		t.Errorf("sidebar pady = %d, want 2", sb.GetPadY())
	}

	g, ok := deck.Meta.Layouts["grid2x2"]
	if !ok {
		t.Fatal("missing grid2x2 layout")
	}
	if len(g.Columns) != 2 || len(g.Rows) != 2 {
		t.Errorf("grid2x2 = %v cols, %v rows", g.Columns, g.Rows)
	}
	// No gutter/pad specified → defaults
	if g.GetGutter() != 2 {
		t.Errorf("default gutter = %d, want 2", g.GetGutter())
	}
	if g.GetPadX() != -1 {
		t.Errorf("default padx = %d, want -1 (unset)", g.GetPadX())
	}
}

func TestPauseMarker(t *testing.T) {
	input := `---
incrementalLists: false
---
# Title

First paragraph

. . .

Second paragraph

. . .

Third paragraph`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slide := deck.Slides[0]
	if slide.Steps != 2 {
		t.Fatalf("steps = %d, want 2", slide.Steps)
	}
	// Heading + 3 paragraphs = 4 blocks
	if len(slide.Blocks) != 4 {
		t.Fatalf("blocks = %d, want 4", len(slide.Blocks))
	}
	// Heading and first paragraph are step 0
	if slide.Blocks[0].Step != 0 {
		t.Errorf("heading step = %d, want 0", slide.Blocks[0].Step)
	}
	if slide.Blocks[1].Step != 0 {
		t.Errorf("para1 step = %d, want 0", slide.Blocks[1].Step)
	}
	// Second paragraph is step 1
	if slide.Blocks[2].Step != 1 {
		t.Errorf("para2 step = %d, want 1", slide.Blocks[2].Step)
	}
	// Third paragraph is step 2
	if slide.Blocks[3].Step != 2 {
		t.Errorf("para3 step = %d, want 2", slide.Blocks[3].Step)
	}
}

func TestIncrementalListsSplitsItems(t *testing.T) {
	input := `---
incrementalLists: true
---
- Alpha
- Beta
- Gamma`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slide := deck.Slides[0]
	// 3 items → 3 blocks (one per item)
	if len(slide.Blocks) != 3 {
		t.Fatalf("blocks = %d, want 3", len(slide.Blocks))
	}
	// Steps 0, 1, 2
	for i, b := range slide.Blocks {
		if b.Step != i {
			t.Errorf("block[%d].Step = %d, want %d", i, b.Step, i)
		}
		if b.Type != model.BlockUnorderedList {
			t.Errorf("block[%d].Type = %v, want BlockUnorderedList", i, b.Type)
		}
	}
	if slide.Steps != 2 {
		t.Errorf("slide.Steps = %d, want 2", slide.Steps)
	}
}

func TestIncrementalListsDisabled(t *testing.T) {
	input := `---
incrementalLists: false
---
- Alpha
- Beta
- Gamma`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slide := deck.Slides[0]
	// List stays as single block
	if len(slide.Blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(slide.Blocks))
	}
	if slide.Steps != 0 {
		t.Errorf("slide.Steps = %d, want 0", slide.Steps)
	}
}

func TestPauseWithIncrementalLists(t *testing.T) {
	input := `---
incrementalLists: true
---

# Title

. . .

- First
- Second

. . .

Final paragraph`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slide := deck.Slides[0]
	// Heading(0) + list item First(1) + list item Second(2) + paragraph(3)
	if len(slide.Blocks) != 4 {
		t.Fatalf("blocks = %d, want 4", len(slide.Blocks))
	}
	if slide.Blocks[0].Step != 0 {
		t.Errorf("heading step = %d, want 0", slide.Blocks[0].Step)
	}
	// List items after first ". . ." get incremental steps
	if slide.Blocks[1].Step != 1 {
		t.Errorf("list item 1 step = %d, want 1", slide.Blocks[1].Step)
	}
	if slide.Blocks[2].Step != 2 {
		t.Errorf("list item 2 step = %d, want 2", slide.Blocks[2].Step)
	}
	// Paragraph after second ". . ."
	if slide.Blocks[3].Step != 3 {
		t.Errorf("paragraph step = %d, want 3", slide.Blocks[3].Step)
	}
	if slide.Steps != 3 {
		t.Errorf("slide.Steps = %d, want 3", slide.Steps)
	}
}

func TestDisableRevealOverridesAll(t *testing.T) {
	input := `---
disableReveal: true
---
# Title

. . .

- First
- Second

. . .

Final paragraph`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slide := deck.Slides[0]
	// All blocks should have Step 0, no reveal
	for i, b := range slide.Blocks {
		if b.Step != 0 {
			t.Errorf("block %d step = %d, want 0", i, b.Step)
		}
	}
	if slide.Steps != 0 {
		t.Errorf("slide.Steps = %d, want 0", slide.Steps)
	}
	// List should NOT be expanded into individual items
	listCount := 0
	for _, b := range slide.Blocks {
		if b.Type == model.BlockUnorderedList {
			listCount++
		}
	}
	if listCount != 1 {
		t.Errorf("list blocks = %d, want 1 (should not be split)", listCount)
	}
}

func TestRegionBreakTitleCols2(t *testing.T) {
	// 1 heading + 1 region break → 3 sections (title + left + right) for title-cols-2.
	// Deck FM first, then slide FM with layout, then content with --- region break.
	input := `---
title: test
---
---
layout: title-cols-2
---
# Slide Title

- bullet one
- bullet two

---

` + "```art" + `
  +---------+
  | diagram |
  +---------+
` + "```"

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1", len(deck.Slides))
	}
	slide := deck.Slides[0]

	// Should contain: heading, list, region-break, art block
	hasBreak := false
	for _, b := range slide.Blocks {
		if b.Type == model.BlockRegionBreak {
			hasBreak = true
		}
	}
	if !hasBreak {
		t.Error("expected BlockRegionBreak in slide blocks")
	}
}

func TestRegionBreakCols2NoHeaders(t *testing.T) {
	// 0 headings + 1 region break → 3 sections for title-cols-2.
	input := `---
title: test
---
---
layout: title-cols-2
---

Left side content

---

Right side content`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deck.Slides) != 1 {
		t.Fatalf("slides = %d, want 1", len(deck.Slides))
	}
	slide := deck.Slides[0]

	hasBreak := false
	for _, b := range slide.Blocks {
		if b.Type == model.BlockRegionBreak {
			hasBreak = true
		}
	}
	if !hasBreak {
		t.Error("expected BlockRegionBreak in slide blocks")
	}
}

func TestRegionBreakNotAbsorbedWhenHeadersSuffice(t *testing.T) {
	// 3 headings for title-cols-2 (needs 3) → no merge, next slide stays separate.
	input := `---
title: test
---
---
layout: title-cols-2
---
# Title

## Left Column

Left content

## Right Column

Right content

---

This is a separate slide`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deck.Slides) != 2 {
		t.Fatalf("slides = %d, want 2", len(deck.Slides))
	}
	// Second slide should just have paragraph content
	s2 := deck.Slides[1]
	if len(s2.Blocks) != 1 {
		t.Fatalf("second slide blocks = %d, want 1", len(s2.Blocks))
	}
	if s2.Blocks[0].Type != model.BlockParagraph {
		t.Errorf("second slide block type = %d, want BlockParagraph", s2.Blocks[0].Type)
	}
}

func TestRegionBreakMixedWithHeaders(t *testing.T) {
	// 1 heading + 1 region break for title-cols-2 (needs 3 sections).
	// We have 1 heading (= 1 section) + title auto-split = 2, so we absorb 1 chunk.
	// The third chunk stays separate.
	input := `---
title: test
---
---
layout: title-cols-2
---
## Left Side

Left content

---

Right content

---

Separate slide`

	deck, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The layout needs 3 regions. We have 1 heading with content (= 2 majors via title auto-split)
	// so we absorb 1 more chunk. The third chunk is a separate slide.
	if len(deck.Slides) != 2 {
		t.Fatalf("slides = %d, want 2", len(deck.Slides))
	}
}
