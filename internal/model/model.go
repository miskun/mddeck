// Package model defines the core data types for mddeck.
package model

// Layout represents slide layout mode.
type Layout string

const (
	LayoutAuto     Layout = "auto"
	LayoutTitle    Layout = "title"
	LayoutCenter   Layout = "center"
	LayoutTwoCol   Layout = "two-col"
	LayoutSplit    Layout = "split"
	LayoutTerminal Layout = "terminal"
)

// Align represents vertical alignment within a layout region.
type Align string

const (
	AlignTop    Align = "top"
	AlignMiddle Align = "middle"
	AlignBottom Align = "bottom"
)

// DeckMeta holds deck-level frontmatter.
type DeckMeta struct {
	Title     string `yaml:"title"`
	Theme     string `yaml:"theme"`
	Wrap      *bool  `yaml:"wrap"`      // pointer so we can detect unset vs false
	TabSize   *int   `yaml:"tabSize"`   // pointer so we can detect unset vs 0
	MaxWidth  int    `yaml:"maxWidth"`
	MaxHeight int    `yaml:"maxHeight"`
	SafeAnsi  *bool  `yaml:"safeAnsi"` // pointer so we can detect unset vs false
}

// DeckMetaDefaults returns a DeckMeta with default values applied.
func DeckMetaDefaults() DeckMeta {
	t := true
	tab := 2
	return DeckMeta{
		Theme:    "default",
		Wrap:     &t,
		TabSize:  &tab,
		SafeAnsi: &t,
	}
}

// GetWrap returns the effective wrap setting.
func (d DeckMeta) GetWrap() bool {
	if d.Wrap == nil {
		return true
	}
	return *d.Wrap
}

// GetTabSize returns the effective tab size.
func (d DeckMeta) GetTabSize() int {
	if d.TabSize == nil {
		return 2
	}
	return *d.TabSize
}

// GetSafeAnsi returns the effective safe ANSI setting.
func (d DeckMeta) GetSafeAnsi() bool {
	if d.SafeAnsi == nil {
		return true
	}
	return *d.SafeAnsi
}

// SlideMeta holds per-slide frontmatter.
type SlideMeta struct {
	Layout Layout `yaml:"layout"`
	Ratio  string `yaml:"ratio"`
	Align  Align  `yaml:"align"`
	Title  string `yaml:"title"`
	Class  string `yaml:"class"`
}

// SlideMetaDefaults returns a SlideMeta with default values.
func SlideMetaDefaults() SlideMeta {
	return SlideMeta{
		Layout: LayoutAuto,
		Align:  AlignTop,
	}
}

// BlockType identifies the kind of content block in a slide.
type BlockType int

const (
	BlockParagraph BlockType = iota
	BlockHeading
	BlockUnorderedList
	BlockOrderedList
	BlockBlockquote
	BlockFencedCode
	BlockHorizontalRule
	BlockANSIArt
	BlockASCIIArt
	BlockBrailleArt
)

// Block represents a parsed content block.
type Block struct {
	Type     BlockType
	Raw      string   // raw text content
	Level    int      // heading level (1-3) or list indent
	Language string   // fenced code block language
	Lines    []string // individual lines (for lists, blockquotes)
	Children []Block  // nested blocks (for blockquotes)
}

// IsArtBlock returns true if this block is an art block (ansi, ascii, braille).
func (b Block) IsArtBlock() bool {
	return b.Type == BlockANSIArt || b.Type == BlockASCIIArt || b.Type == BlockBrailleArt
}

// IsCodeLike returns true if this is a code or art block.
func (b Block) IsCodeLike() bool {
	return b.Type == BlockFencedCode || b.IsArtBlock()
}

// Slide represents a single slide in the deck.
type Slide struct {
	Meta   SlideMeta
	Blocks []Block
	Notes  string // speaker notes (raw markdown)
	Index  int    // 0-based index in deck
}

// Deck is the top-level container for a parsed presentation.
type Deck struct {
	Meta   DeckMeta
	Slides []Slide
	Source string // file path
}

// MajorBlock represents a top-level heading and its following content,
// used for layout heuristics.
type MajorBlock struct {
	Heading Block
	Content []Block
}
