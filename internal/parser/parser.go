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

	// Apply header-based splitting (à la patat) to each chunk produced by
	// splitSlides.  This handles files that mix --- slide breaks with ##
	// header-based splitting: each ---separated chunk is further split on
	// headers when it contains more than one at the deepest level.
	var expanded []string
	for _, st := range slideTexts {
		headerSlides := splitSlidesByHeaders(st, deck.Meta.Layouts)
		if len(headerSlides) > 1 {
			expanded = append(expanded, headerSlides...)
		} else {
			expanded = append(expanded, st)
		}
	}
	slideTexts = expanded

	for i, text := range slideTexts {
		slide, err := parseSlide(text, i)
		if err != nil {
			return nil, fmt.Errorf("parsing slide %d: %w", i+1, err)
		}
		// Skip empty slides (e.g. resume markers like "autosplit: true")
		if len(slide.Blocks) == 0 && slide.Notes == "" {
			continue
		}
		deck.Slides = append(deck.Slides, *slide)
	}

	// Re-index slides after filtering
	for i := range deck.Slides {
		deck.Slides[i].Index = i
	}

	// Apply incremental lists and compute reveal steps
	for i := range deck.Slides {
		applyRevealSteps(&deck.Slides[i], &deck.Meta)
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

// splitSlidesByHeaders implements patat-style header-based slide splitting.
// When a file contains no --- slide breaks, headers are used instead:
//   - Find the most deeply nested header level (the "split level")
//   - Each occurrence of that header starts a new slide
//   - Headers above the split level also start a new slide and become title slides
//   - A YAML frontmatter block (--- / key: val / ---) also starts a new slide
//
// For example, if the deepest header is ## (h2):
//   - # Title     → starts a new title slide
//   - ## Content   → starts a new content slide
//   - ### Sub      → stays within the current slide
//   - ---\nlayout: cols-2\n--- → starts a new slide with that layout
func splitSlidesByHeaders(content string, layouts map[string]model.CustomLayout) []string {
	lines := strings.Split(content, "\n")

	// Find the deepest (most nested) header level used.
	deepest := 0
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		level := headerLevel(trimmed)
		if level > 0 && level > deepest {
			deepest = level
		}
	}

	hasFM := false
	for i := range lines {
		if _, ok := isSlideFrontmatter(lines, i); ok {
			hasFM = true
			break
		}
	}

	if deepest == 0 && !hasFM {
		// No headers or frontmatter found — can't split
		return []string{content}
	}

	// Split on headers at or above the deepest level, and on frontmatter blocks.
	var slides []string
	var current []string
	inFence = false
	headersToSkip := 0

	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			current = append(current, lines[i])
			i++
			continue
		}
		if inFence {
			current = append(current, lines[i])
			i++
			continue
		}

		// Check for slide frontmatter block (--- / yaml / ---)
		if endFM, ok := isSlideFrontmatter(lines, i); ok {
			// Flush current content as a slide
			text := strings.Join(current, "\n")
			text = strings.TrimRight(text, "\n ")
			if strings.TrimSpace(text) != "" {
				slides = append(slides, text)
			}

			fmYAML := strings.Join(lines[i+1:endFM], "\n")

			// Check if autosplit is explicitly disabled.
			// When autosplit: false, consume all lines until the
			// next frontmatter block — no header splitting within.
			noSplit := strings.Contains(fmYAML, "autosplit: false") ||
				strings.Contains(fmYAML, "autosplit:false")

			// Start new slide with the frontmatter block
			current = nil
			for j := i; j <= endFM; j++ {
				current = append(current, lines[j])
			}
			i = endFM + 1

			if noSplit {
				// Absorb all lines until the next frontmatter block or EOF.
				// Headers within this zone are NOT treated as split points.
				// To resume normal header splitting after a no-split zone,
				// start the next slide with a frontmatter block, e.g.:
				//   ---
				//   autosplit: true
				//   ---
				for i < len(lines) {
					if _, ok := isSlideFrontmatter(lines, i); ok {
						break // next frontmatter starts a new slide
					}
					current = append(current, lines[i])
					i++
				}
			} else {
				// Determine how many subsequent headers to absorb.
				// Built-in multi-region layouts need 2, single-region need 1.
				// Custom layouts: compute cols × rows for region count.
				// No layout key → resume marker, absorb nothing.
				headersToSkip = computeHeadersToSkip(fmYAML, layouts)
			}
			continue
		}

		// Check for header split
		level := headerLevel(trimmed)
		if level > 0 && deepest > 0 && level <= deepest {
			if headersToSkip > 0 {
				// This header belongs to the preceding frontmatter slide.
				headersToSkip--
				current = append(current, lines[i])
				i++
				continue
			}
			// This header starts a new slide.
			text := strings.Join(current, "\n")
			text = strings.TrimRight(text, "\n ")
			if strings.TrimSpace(text) != "" {
				slides = append(slides, text)
			}
			current = nil
		}

		current = append(current, lines[i])
		i++
	}

	// Last slide
	text := strings.Join(current, "\n")
	text = strings.TrimRight(text, "\n ")
	if strings.TrimSpace(text) != "" {
		slides = append(slides, text)
	}

	return slides
}

// isSlideFrontmatter checks if lines[idx] starts a YAML frontmatter block.
// Returns the index of the closing --- and true, or (0, false).
// A frontmatter block is: ---, followed by at least one YAML line with ":",
// followed by a closing ---.
func isSlideFrontmatter(lines []string, idx int) (endIdx int, ok bool) {
	if idx >= len(lines) || strings.TrimSpace(lines[idx]) != "---" {
		return 0, false
	}
	// Look for closing ---
	hasYAML := false
	for j := idx + 1; j < len(lines) && j-idx <= 20; j++ {
		trimmed := strings.TrimSpace(lines[j])
		if trimmed == "---" {
			if hasYAML {
				return j, true
			}
			return 0, false // empty block, not frontmatter
		}
		if strings.Contains(trimmed, ":") {
			hasYAML = true
		}
	}
	return 0, false
}

// headerLevel returns the heading level (1-6) for a markdown heading line,
// computeHeadersToSkip determines how many subsequent headers a slide
// frontmatter should absorb based on its layout.
// For custom layouts, it computes cols × rows from the layout definition.
// For built-in multi-region layouts (cols-2, rows-2, sidebar), returns 2.
// For other built-in layouts, returns 1.
// If no layout is specified, returns 0 (resume marker).
func computeHeadersToSkip(fmYAML string, layouts map[string]model.CustomLayout) int {
	// Extract layout value from the YAML
	layoutName := ""
	for _, line := range strings.Split(fmYAML, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "layout:") {
			val := strings.TrimPrefix(line, "layout:")
			layoutName = strings.TrimSpace(val)
			// Strip optional quotes
			layoutName = strings.Trim(layoutName, "\"'")
			break
		}
	}

	if layoutName == "" {
		return 0 // no layout → resume marker
	}

	// Check custom layouts first
	if layouts != nil {
		if custom, ok := layouts[layoutName]; ok {
			cols := len(custom.Columns)
			rows := len(custom.Rows)
			if cols == 0 {
				cols = 1
			}
			if rows == 0 {
				rows = 1
			}
			return cols * rows
		}
	}

	// Built-in multi-region layouts
	switch layoutName {
	case "cols-2", "rows-2", "sidebar":
		return 2
	case "cols-3":
		return 3
	case "grid-4":
		return 4
	}

	// Other built-in layouts (title, default, center, terminal)
	return 1
}

// or 0 if the line is not a heading.
func headerLevel(line string) int {
	if !strings.HasPrefix(line, "#") {
		return 0
	}
	level := 0
	for _, ch := range line {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level > 6 {
		return 0
	}
	// Must be followed by a space or be just "#"s
	if len(line) > level && line[level] != ' ' {
		return 0
	}
	return level
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
	step := 0 // current reveal step

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip blank lines
		if trimmed == "" {
			i++
			continue
		}

		// Pause marker: ". . ." advances the step counter
		if trimmed == ". . ." {
			step++
			i++
			continue
		}

		// Fenced code block
		if strings.HasPrefix(trimmed, "```") {
			block, end := parseFencedBlock(lines, i)
			block.Step = step
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Heading
		if strings.HasPrefix(trimmed, "#") {
			block, ok := parseHeading(trimmed)
			if ok {
				block.Step = step
				blocks = append(blocks, block)
				i++
				continue
			}
		}

		// Horizontal rule (--- that's not a slide break, already handled)
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			blocks = append(blocks, model.Block{Type: model.BlockHorizontalRule, Step: step})
			i++
			continue
		}

		// Blockquote
		if strings.HasPrefix(trimmed, "> ") || trimmed == ">" {
			block, end := parseBlockquote(lines, i)
			block.Step = step
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Task list (must be checked before unordered list)
		if isTaskListItem(trimmed) {
			block, end := parseTaskList(lines, i)
			block.Step = step
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Table (pipe-delimited)
		if isTableLine(trimmed) {
			block, end := parseTable(lines, i)
			block.Step = step
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Unordered list
		if isUnorderedListItem(trimmed) {
			block, end := parseUnorderedList(lines, i)
			block.Step = step
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Ordered list
		if isOrderedListItem(trimmed) {
			block, end := parseOrderedList(lines, i)
			block.Step = step
			blocks = append(blocks, block)
			i = end
			continue
		}

		// Paragraph (default)
		block, end := parseParagraph(lines, i)
		block.Step = step
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

// parseBlockquote parses a blockquote or alert/callout block.
func parseBlockquote(lines []string, i int) (model.Block, int) {
	var quoteLines []string
	j := i
	for j < len(lines) {
		trimmed := strings.TrimSpace(lines[j])
		if strings.HasPrefix(trimmed, "> ") {
			quoteLines = append(quoteLines, strings.TrimPrefix(trimmed, "> "))
		} else if trimmed == ">" {
			quoteLines = append(quoteLines, "")
		} else {
			// Blank or non-blockquote line ends this block
			break
		}
		j++
	}

	// Check if this is an alert/callout: first line matches [!TYPE]
	if len(quoteLines) > 0 {
		alertType := parseAlertType(quoteLines[0])
		if alertType != "" {
			// Remove the [!TYPE] line from content
			contentLines := quoteLines[1:]
			// Remove leading empty line if present
			if len(contentLines) > 0 && contentLines[0] == "" {
				contentLines = contentLines[1:]
			}
			return model.Block{
				Type:     model.BlockAlert,
				Raw:      alertType,
				Lines:    contentLines,
				Language: alertType, // store alert type in Language field
			}, j
		}
	}

	return model.Block{
		Type:  model.BlockBlockquote,
		Raw:   strings.Join(quoteLines, "\n"),
		Lines: quoteLines,
	}, j
}

// parseAlertType extracts the alert type from a [!TYPE] marker.
// Returns the type string (e.g., "NOTE", "WARNING") or "" if not an alert.
func parseAlertType(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "[!") || !strings.HasSuffix(line, "]") {
		return ""
	}
	alertType := strings.ToUpper(line[2 : len(line)-1])
	switch alertType {
	case "NOTE", "TIP", "IMPORTANT", "WARNING", "CAUTION":
		return alertType
	}
	return ""
}

// isUnorderedListItem checks if a line starts an unordered list.
func isUnorderedListItem(line string) bool {
	if len(line) < 2 {
		return false
	}
	return (line[0] == '-' || line[0] == '*') && line[1] == ' '
}

// indentedListItem checks if a line is a list item at any indentation level.
// Returns (depth, itemText, true) if it's a list item, or (0, "", false) otherwise.
// depth is measured in units of 2 spaces.
func indentedListItem(line string) (int, string, bool) {
	indent := 0
	for _, ch := range line {
		if ch == ' ' {
			indent++
		} else if ch == '\t' {
			indent += 2
		} else {
			break
		}
	}
	trimmed := strings.TrimLeft(line, " \t")
	if isUnorderedListItem(trimmed) {
		depth := indent / 2
		return depth, trimmed[2:], true
	}
	return 0, "", false
}

// parseUnorderedList parses an unordered list block with nesting support.
// Items are stored in Lines with a depth prefix: each entry is "DEPTH:text"
// where DEPTH is a single digit (0-9).
func parseUnorderedList(lines []string, i int) (model.Block, int) {
	var items []string
	j := i
	currentItem := ""
	currentDepth := 0

	for j < len(lines) {
		line := lines[j]
		trimmed := strings.TrimSpace(line)

		if depth, text, ok := indentedListItem(line); ok {
			if currentItem != "" {
				items = append(items, fmt.Sprintf("%d:%s", currentDepth, currentItem))
			}
			currentDepth = depth
			currentItem = text
		} else if trimmed == "" {
			// Check if list continues
			if j+1 < len(lines) {
				if _, _, ok := indentedListItem(lines[j+1]); ok {
					j++
					continue
				}
			}
			break
		} else if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
			// Continuation of current item (indented)
			if strings.HasSuffix(currentItem, "\\") {
				// Trailing backslash = hard line break
				currentItem = strings.TrimSuffix(currentItem, "\\") + "\n" + trimmed
			} else {
				currentItem += " " + trimmed
			}
		} else if currentItem != "" && strings.HasSuffix(currentItem, "\\") {
			// Trailing backslash absorbs next line as continuation regardless of indent
			currentItem = strings.TrimSuffix(currentItem, "\\") + "\n" + trimmed
		} else {
			break
		}
		j++
	}

	if currentItem != "" {
		items = append(items, fmt.Sprintf("%d:%s", currentDepth, currentItem))
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

// indentedOrderedListItem checks if a line is an ordered list item at any indentation.
// Returns (depth, itemText, true) if it's an ordered list item.
func indentedOrderedListItem(line string) (int, string, bool) {
	indent := 0
	for _, ch := range line {
		if ch == ' ' {
			indent++
		} else if ch == '\t' {
			indent += 2
		} else {
			break
		}
	}
	trimmed := strings.TrimLeft(line, " \t")
	if isOrderedListItem(trimmed) {
		depth := indent / 2
		dotIdx := strings.Index(trimmed, ". ")
		if dotIdx >= 0 {
			return depth, trimmed[dotIdx+2:], true
		}
	}
	return 0, "", false
}

// parseOrderedList parses an ordered list block with nesting support.
func parseOrderedList(lines []string, i int) (model.Block, int) {
	var items []string
	j := i
	currentItem := ""
	currentDepth := 0

	for j < len(lines) {
		line := lines[j]
		trimmed := strings.TrimSpace(line)

		if depth, text, ok := indentedOrderedListItem(line); ok {
			if currentItem != "" {
				items = append(items, fmt.Sprintf("%d:%s", currentDepth, currentItem))
			}
			currentDepth = depth
			currentItem = text
		} else if trimmed == "" {
			if j+1 < len(lines) {
				if _, _, ok := indentedOrderedListItem(lines[j+1]); ok {
					j++
					continue
				}
			}
			break
		} else if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
			if strings.HasSuffix(currentItem, "\\") {
				currentItem = strings.TrimSuffix(currentItem, "\\") + "\n" + trimmed
			} else {
				currentItem += " " + trimmed
			}
		} else if currentItem != "" && strings.HasSuffix(currentItem, "\\") {
			// Trailing backslash absorbs next line as continuation regardless of indent
			currentItem = strings.TrimSuffix(currentItem, "\\") + "\n" + trimmed
		} else {
			break
		}
		j++
	}

	if currentItem != "" {
		items = append(items, fmt.Sprintf("%d:%s", currentDepth, currentItem))
	}

	return model.Block{
		Type:  model.BlockOrderedList,
		Raw:   strings.Join(items, "\n"),
		Lines: items,
	}, j
}

// isTableLine checks if a line looks like a pipe-delimited table row.
func isTableLine(line string) bool {
	return strings.HasPrefix(line, "|") && strings.Contains(line[1:], "|")
}

// isTableSeparator checks if a line is a table separator (e.g., |---|---|).
func isTableSeparator(line string) bool {
	if !strings.HasPrefix(line, "|") {
		return false
	}
	inner := strings.Trim(line, "| ")
	// Should contain only dashes, colons, pipes, and spaces
	for _, ch := range inner {
		if ch != '-' && ch != ':' && ch != '|' && ch != ' ' {
			return false
		}
	}
	return strings.Contains(inner, "-")
}

// parseTableRow splits a pipe-delimited row into cells.
func parseTableRow(line string) []string {
	// Trim leading/trailing pipes
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}
	cells := strings.Split(line, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

// parseTable parses a pipe-delimited table block.
// Lines[0] = header row, Lines[1:] = data rows. Each line is pipe-separated.
// The first separator row after the header is skipped (standard markdown).
// Subsequent separator rows are preserved to render mid-table borders.
func parseTable(lines []string, i int) (model.Block, int) {
	var tableLines []string
	j := i
	headerSepSeen := false
	noHeader := false

	for j < len(lines) {
		trimmed := strings.TrimSpace(lines[j])
		if !isTableLine(trimmed) {
			break
		}
		if isTableSeparator(trimmed) {
			if !headerSepSeen {
				if len(tableLines) == 0 {
					// Separator before any data row — headerless table
					noHeader = true
				}
				// Skip the first separator (standard header separator)
				headerSepSeen = true
			} else {
				// Keep subsequent separators for mid-table borders
				tableLines = append(tableLines, trimmed)
			}
		} else {
			tableLines = append(tableLines, trimmed)
		}
		j++
	}

	return model.Block{
		Type:     model.BlockTable,
		Raw:      strings.Join(tableLines, "\n"),
		Lines:    tableLines,
		NoHeader: noHeader,
	}, j
}

// isTaskListItem checks if a line is a task list item (- [ ] or - [x]).
func isTaskListItem(line string) bool {
	return strings.HasPrefix(line, "- [ ] ") || strings.HasPrefix(line, "- [x] ") ||
		strings.HasPrefix(line, "- [X] ") || line == "- [ ]" || line == "- [x]" || line == "- [X]"
}

// indentedTaskListItem checks if a line is a task list item at any indentation.
func indentedTaskListItem(line string) (int, bool, string, bool) {
	indent := 0
	for _, ch := range line {
		if ch == ' ' {
			indent++
		} else if ch == '\t' {
			indent += 2
		} else {
			break
		}
	}
	trimmed := strings.TrimLeft(line, " \t")
	if isTaskListItem(trimmed) {
		depth := indent / 2
		checked := strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]")
		text := ""
		if len(trimmed) > 6 {
			text = trimmed[6:]
		}
		return depth, checked, text, true
	}
	return 0, false, "", false
}

// parseTaskList parses a task list block.
// Items in Lines are stored as "DEPTH:C:text" where C is 1 (checked) or 0 (unchecked).
func parseTaskList(lines []string, i int) (model.Block, int) {
	var items []string
	j := i

	for j < len(lines) {
		line := lines[j]
		trimmed := strings.TrimSpace(line)

		if depth, checked, text, ok := indentedTaskListItem(line); ok {
			c := "0"
			if checked {
				c = "1"
			}
			items = append(items, fmt.Sprintf("%d:%s:%s", depth, c, text))
		} else if trimmed == "" {
			// Check if list continues
			if j+1 < len(lines) {
				if _, _, _, ok := indentedTaskListItem(lines[j+1]); ok {
					j++
					continue
				}
			}
			break
		} else {
			break
		}
		j++
	}

	return model.Block{
		Type:  model.BlockTaskList,
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
			isTaskListItem(trimmed) ||
			isTableLine(trimmed) ||
			trimmed == "---" || trimmed == "***" || trimmed == "___" ||
			trimmed == "???" {
			break
		}
		paraLines = append(paraLines, trimmed)
		j++
	}

	return model.Block{
		Type: model.BlockParagraph,
		Raw:  joinParagraphLines(paraLines),
	}, j
}

// joinParagraphLines joins paragraph lines, preserving hard line breaks.
// A trailing backslash (\) or two trailing spaces indicate a hard break,
// which is preserved as a newline in the raw text.
func joinParagraphLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			// Check if previous line ended with a hard break
			prev := lines[i-1]
			if strings.HasSuffix(prev, "\\") || strings.HasSuffix(prev, "  ") {
				// Already added newline below
			} else {
				result.WriteByte(' ')
			}
		}
		// Strip trailing backslash (the break marker)
		if strings.HasSuffix(line, "\\") {
			result.WriteString(strings.TrimSuffix(line, "\\"))
			if i < len(lines)-1 {
				result.WriteByte('\n')
			}
		} else if strings.HasSuffix(line, "  ") {
			// Two trailing spaces = hard break
			result.WriteString(strings.TrimRight(line, " "))
			if i < len(lines)-1 {
				result.WriteByte('\n')
			}
		} else {
			result.WriteString(line)
		}
	}
	return result.String()
}

// applyRevealSteps processes a slide's blocks for progressive reveal.
// If incrementalLists is enabled, each list block is split into individual items,
// each getting an incrementing step value. Then Slide.Steps is set to the max step.
func applyRevealSteps(slide *model.Slide, deckMeta *model.DeckMeta) {
	// Determine if incremental lists are enabled for this slide.
	// Slide-level setting overrides deck-level.
	incremental := deckMeta.GetIncrementalLists()
	if slide.Meta.IncrementalLists != nil {
		incremental = *slide.Meta.IncrementalLists
	}

	if incremental {
		slide.Blocks = expandIncrementalLists(slide.Blocks)
	}

	// Compute total steps for this slide
	maxStep := 0
	for _, b := range slide.Blocks {
		if b.Step > maxStep {
			maxStep = b.Step
		}
	}
	slide.Steps = maxStep
}

// expandIncrementalLists splits list blocks into individual single-item list
// blocks, each with an incrementing Step value. The first item inherits the
// block's original step; subsequent items get step+1, step+2, etc.
// Non-list blocks are passed through unchanged. When a list expansion adds
// extra steps, all subsequent blocks are shifted forward accordingly.
func expandIncrementalLists(blocks []model.Block) []model.Block {
	var result []model.Block
	offset := 0

	for _, b := range blocks {
		adjustedStep := b.Step + offset

		if !isListBlock(b) || len(b.Lines) <= 1 {
			b.Step = adjustedStep
			result = append(result, b)
			continue
		}

		// Split list into individual items
		for j, line := range b.Lines {
			item := model.Block{
				Type:  b.Type,
				Raw:   line,
				Lines: []string{line},
				Step:  adjustedStep + j,
			}
			result = append(result, item)
		}
		// A list originally occupied 1 step slot; expansion adds len-1 extra
		offset += len(b.Lines) - 1
	}

	return result
}

// isListBlock returns true if the block is a list type.
func isListBlock(b model.Block) bool {
	return b.Type == model.BlockUnorderedList || b.Type == model.BlockOrderedList || b.Type == model.BlockTaskList
}
