// Package parser handles parsing .mddeck files into a Deck structure.
package parser

import (
	"fmt"
	"os"
	"strings"

	"github.com/miskun/mddeck/internal/model"
	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a .mddeck file.
func ParseFile(path string) (*model.Deck, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	deck, err := Parse(string(data))
	if err != nil {
		return nil, err
	}
	deck.Source = path
	return deck, nil
}

// Parse parses .mddeck content string into a Deck.
func Parse(content string) (*model.Deck, error) {
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	lines := strings.Split(content, "\n")

	deck := &model.Deck{
		Meta: model.DeckMetaDefaults(),
	}

	pos := 0

	// Try to parse deck frontmatter (file starts with ---)
	if pos < len(lines) && strings.TrimSpace(lines[pos]) == "---" {
		endPos := findFrontmatterEnd(lines, pos+1)
		if endPos > 0 {
			yamlContent := strings.Join(lines[pos+1:endPos], "\n")
			if err := yaml.Unmarshal([]byte(yamlContent), &deck.Meta); err != nil {
				return nil, fmt.Errorf("parsing deck frontmatter: %w", err)
			}
			// Apply defaults for unset fields
			defaults := model.DeckMetaDefaults()
			if deck.Meta.Theme == "" {
				deck.Meta.Theme = defaults.Theme
			}
			if deck.Meta.Wrap == nil {
				deck.Meta.Wrap = defaults.Wrap
			}
			if deck.Meta.TabSize == nil {
				deck.Meta.TabSize = defaults.TabSize
			}
			if deck.Meta.SafeAnsi == nil {
				deck.Meta.SafeAnsi = defaults.SafeAnsi
			}
			pos = endPos + 1
		}
	}

	// Split remaining content into slides
	slideTexts := splitSlides(lines, pos)

	for i, text := range slideTexts {
		slide, err := parseSlide(text, i)
		if err != nil {
			return nil, fmt.Errorf("parsing slide %d: %w", i+1, err)
		}
		deck.Slides = append(deck.Slides, *slide)
	}

	// Ensure at least one slide
	if len(deck.Slides) == 0 {
		deck.Slides = append(deck.Slides, model.Slide{
			Meta:  model.SlideMetaDefaults(),
			Index: 0,
		})
	}

	return deck, nil
}

// findFrontmatterEnd finds the closing --- for a YAML front matter block.
// Returns the line index of the closing ---, or -1 if not found.
func findFrontmatterEnd(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return i
		}
	}
	return -1
}

// splitSlides splits lines into slide text chunks based on slide break rules.
// A slide break is a line containing exactly "---" with at least one blank line
// before and after.
func splitSlides(lines []string, start int) []string {
	var slides []string
	var current []string

	i := start
	for i < len(lines) {
		line := lines[i]

		if strings.TrimSpace(line) == "---" && isSlideBreak(lines, i, start) {
			// This is a slide break
			text := strings.Join(current, "\n")
			text = strings.TrimRight(text, "\n ")
			if len(slides) > 0 || strings.TrimSpace(text) != "" {
				slides = append(slides, text)
			}
			current = nil
			i++
			// Skip blank lines after the break
			for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
				i++
			}
			continue
		}

		current = append(current, line)
		i++
	}

	// Last slide
	text := strings.Join(current, "\n")
	text = strings.TrimRight(text, "\n ")
	if len(slides) > 0 || strings.TrimSpace(text) != "" {
		slides = append(slides, text)
	}

	return slides
}

// isSlideBreak checks if a "---" line at position idx is a valid slide break.
// Requires at least one blank line before and after.
func isSlideBreak(lines []string, idx int, start int) bool {
	// Check blank line before
	hasBefore := false
	if idx <= start {
		// At start of content, treat as having blank line before
		hasBefore = true
	} else {
		for j := idx - 1; j >= start; j-- {
			if strings.TrimSpace(lines[j]) == "" {
				hasBefore = true
				break
			}
			break // first non-blank line above — no blank found
		}
	}

	if !hasBefore {
		return false
	}

	// Check blank line after
	hasAfter := false
	if idx >= len(lines)-1 {
		// At end of content, treat as having blank line after
		hasAfter = true
	} else {
		for j := idx + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "" {
				hasAfter = true
				break
			}
			break // first non-blank line below — no blank found
		}
	}

	return hasAfter
}

// parseSlide parses a single slide's text into a Slide.
func parseSlide(text string, index int) (*model.Slide, error) {
	slide := &model.Slide{
		Meta:  model.SlideMetaDefaults(),
		Index: index,
	}

	lines := strings.Split(text, "\n")
	pos := 0

	// Skip leading blank lines
	for pos < len(lines) && strings.TrimSpace(lines[pos]) == "" {
		pos++
	}

	// Try to parse slide frontmatter
	if pos < len(lines) && strings.TrimSpace(lines[pos]) == "---" {
		endPos := findFrontmatterEnd(lines, pos+1)
		if endPos > 0 {
			yamlContent := strings.Join(lines[pos+1:endPos], "\n")
			if err := yaml.Unmarshal([]byte(yamlContent), &slide.Meta); err != nil {
				return nil, fmt.Errorf("parsing slide frontmatter: %w", err)
			}
			// Apply defaults for unset fields
			if slide.Meta.Layout == "" {
				slide.Meta.Layout = model.LayoutAuto
			}
			if slide.Meta.Align == "" {
				slide.Meta.Align = model.AlignTop
			}
			pos = endPos + 1
		}
	}

	// Split body from notes at ???
	bodyLines, notes := splitNotes(lines, pos)

	slide.Notes = notes
	slide.Blocks = parseBlocks(bodyLines)

	return slide, nil
}

// splitNotes splits lines into body and speaker notes at the first "???" line.
func splitNotes(lines []string, start int) (body []string, notes string) {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "???" {
			body = lines[start:i]
			if i+1 < len(lines) {
				notes = strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
			}
			return
		}
	}
	return lines[start:], ""
}

// parseBlocks parses markdown lines into content blocks.
func parseBlocks(lines []string) []model.Block {
	var blocks []model.Block
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip blank lines
		if trimmed == "" {
			i++
			continue
		}

		// Fenced code block
		if strings.HasPrefix(trimmed, "```") {
			block, end := parseFencedBlock(lines, i)
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Heading
		if strings.HasPrefix(trimmed, "#") {
			block, ok := parseHeading(trimmed)
			if ok {
				blocks = append(blocks, block)
				i++
				continue
			}
		}

		// Horizontal rule (--- that's not a slide break, already handled)
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			blocks = append(blocks, model.Block{Type: model.BlockHorizontalRule})
			i++
			continue
		}

		// Blockquote
		if strings.HasPrefix(trimmed, "> ") || trimmed == ">" {
			block, end := parseBlockquote(lines, i)
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Unordered list
		if isUnorderedListItem(trimmed) {
			block, end := parseUnorderedList(lines, i)
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Ordered list
		if isOrderedListItem(trimmed) {
			block, end := parseOrderedList(lines, i)
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Paragraph (default)
		block, end := parseParagraph(lines, i)
		blocks = append(blocks, block)
		i = end
	}

	return blocks
}

// parseFencedBlock parses a fenced code block starting at line i.
func parseFencedBlock(lines []string, i int) (model.Block, int) {
	trimmed := strings.TrimSpace(lines[i])
	lang := strings.TrimPrefix(trimmed, "```")
	lang = strings.TrimSpace(lang)

	var codeLines []string
	j := i + 1
	for j < len(lines) {
		if strings.TrimSpace(lines[j]) == "```" {
			j++
			break
		}
		codeLines = append(codeLines, lines[j])
		j++
	}

	raw := strings.Join(codeLines, "\n")

	blockType := model.BlockFencedCode
	switch strings.ToLower(lang) {
	case "ansi":
		blockType = model.BlockANSIArt
	case "ascii":
		blockType = model.BlockASCIIArt
	case "braille":
		blockType = model.BlockBrailleArt
	}

	return model.Block{
		Type:     blockType,
		Raw:      raw,
		Language: lang,
		Lines:    codeLines,
	}, j
}

// parseHeading parses a heading line.
func parseHeading(line string) (model.Block, bool) {
	level := 0
	for _, c := range line {
		if c == '#' {
			level++
		} else {
			break
		}
	}
	if level < 1 || level > 6 {
		return model.Block{}, false
	}
	text := strings.TrimSpace(line[level:])
	if text == "" && level > 0 {
		// A line of just ### is valid
	}
	return model.Block{
		Type:  model.BlockHeading,
		Raw:   text,
		Level: level,
	}, true
}

// parseBlockquote parses a blockquote block.
func parseBlockquote(lines []string, i int) (model.Block, int) {
	var quoteLines []string
	j := i
	for j < len(lines) {
		trimmed := strings.TrimSpace(lines[j])
		if strings.HasPrefix(trimmed, "> ") {
			quoteLines = append(quoteLines, strings.TrimPrefix(trimmed, "> "))
		} else if trimmed == ">" {
			quoteLines = append(quoteLines, "")
		} else if trimmed == "" {
			// Check if next line continues the blockquote
			if j+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[j+1]), ">") {
				quoteLines = append(quoteLines, "")
			} else {
				break
			}
		} else {
			break
		}
		j++
	}

	return model.Block{
		Type:  model.BlockBlockquote,
		Raw:   strings.Join(quoteLines, "\n"),
		Lines: quoteLines,
	}, j
}

// isUnorderedListItem checks if a line starts an unordered list.
func isUnorderedListItem(line string) bool {
	if len(line) < 2 {
		return false
	}
	return (line[0] == '-' || line[0] == '*') && line[1] == ' '
}

// parseUnorderedList parses an unordered list block.
func parseUnorderedList(lines []string, i int) (model.Block, int) {
	var items []string
	j := i
	currentItem := ""

	for j < len(lines) {
		trimmed := strings.TrimSpace(lines[j])
		if isUnorderedListItem(trimmed) {
			if currentItem != "" {
				items = append(items, currentItem)
			}
			currentItem = trimmed[2:] // skip "- " or "* "
		} else if trimmed == "" {
			// Check if list continues
			if j+1 < len(lines) && isUnorderedListItem(strings.TrimSpace(lines[j+1])) {
				j++
				continue
			}
			break
		} else if strings.HasPrefix(lines[j], "  ") || strings.HasPrefix(lines[j], "\t") {
			// Continuation of current item
			currentItem += " " + trimmed
		} else {
			break
		}
		j++
	}

	if currentItem != "" {
		items = append(items, currentItem)
	}

	return model.Block{
		Type:  model.BlockUnorderedList,
		Raw:   strings.Join(items, "\n"),
		Lines: items,
	}, j
}

// isOrderedListItem checks if a line starts an ordered list item.
func isOrderedListItem(line string) bool {
	for i, c := range line {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' && i > 0 && i+1 < len(line) && line[i+1] == ' ' {
			return true
		}
		return false
	}
	return false
}

// parseOrderedList parses an ordered list block.
func parseOrderedList(lines []string, i int) (model.Block, int) {
	var items []string
	j := i
	currentItem := ""

	for j < len(lines) {
		trimmed := strings.TrimSpace(lines[j])
		if isOrderedListItem(trimmed) {
			if currentItem != "" {
				items = append(items, currentItem)
			}
			// Find the ". " and take text after it
			dotIdx := strings.Index(trimmed, ". ")
			if dotIdx >= 0 {
				currentItem = trimmed[dotIdx+2:]
			}
		} else if trimmed == "" {
			if j+1 < len(lines) && isOrderedListItem(strings.TrimSpace(lines[j+1])) {
				j++
				continue
			}
			break
		} else if strings.HasPrefix(lines[j], "  ") || strings.HasPrefix(lines[j], "\t") {
			currentItem += " " + trimmed
		} else {
			break
		}
		j++
	}

	if currentItem != "" {
		items = append(items, currentItem)
	}

	return model.Block{
		Type:  model.BlockOrderedList,
		Raw:   strings.Join(items, "\n"),
		Lines: items,
	}, j
}

// parseParagraph parses a paragraph (consecutive non-blank, non-special lines).
func parseParagraph(lines []string, i int) (model.Block, int) {
	var paraLines []string
	j := i

	for j < len(lines) {
		trimmed := strings.TrimSpace(lines[j])
		if trimmed == "" {
			break
		}
		// Stop if we hit a block-level element
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, "> ") ||
			trimmed == ">" ||
			isUnorderedListItem(trimmed) ||
			isOrderedListItem(trimmed) ||
			trimmed == "---" || trimmed == "***" || trimmed == "___" ||
			trimmed == "???" {
			break
		}
		paraLines = append(paraLines, trimmed)
		j++
	}

	return model.Block{
		Type: model.BlockParagraph,
		Raw:  strings.Join(paraLines, " "),
	}, j
}
