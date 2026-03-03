// Package model defines the core data types for mddeck.
package model

// Layout represents slide layout mode.
type Layout string

const (
	LayoutAuto     Layout = "auto"
	LayoutDefault  Layout = "default"
	LayoutTitle    Layout = "title"
	LayoutCenter   Layout = "center"
	LayoutTwoCol   Layout = "two-col"
	LayoutSplit    Layout = "split"
	LayoutTerminal Layout = "terminal"
	LayoutSidebar  Layout = "sidebar"
	LayoutThirds   Layout = "thirds"
	LayoutQuad     Layout = "quad"
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
	Title     string                   `yaml:"title"`
	Theme     string                   `yaml:"theme"`
	Aspect    string                   `yaml:"aspect"`    // target aspect ratio, e.g. "16:9" or "4:3"
	Wrap      *bool                    `yaml:"wrap"`      // pointer so we can detect unset vs false
	TabSize   *int                     `yaml:"tabSize"`   // pointer so we can detect unset vs 0
	MaxWidth  int                      `yaml:"maxWidth"`
	MaxHeight int                      `yaml:"maxHeight"`
	SafeAnsi  *bool                    `yaml:"safeAnsi"`  // pointer so we can detect unset vs false
	Layouts   map[string]CustomLayout  `yaml:"layouts"`   // user-defined or overridden layouts
}

// CustomLayout defines a user-configurable grid layout.
// Columns and Rows define the grid cell sizes as percentages.
// - columns only → single-row, N columns
// - rows only → single-column, N rows
// - both → len(columns) × len(rows) grid (row-major order)
type CustomLayout struct {
	Columns []int  `yaml:"columns"` // column widths as percentages, e.g. [30, 70]
	Rows    []int  `yaml:"rows"`    // row heights as percentages, e.g. [60, 40]
	Gutter  *int   `yaml:"gutter"`  // gap between cells in characters (default: 2)
	PadX    *int   `yaml:"padX"`    // horizontal padding override
	PadY    *int   `yaml:"padY"`    // vertical padding override
	Align   Align  `yaml:"align"`   // content alignment within cells
}

// GetGutter returns the effective gutter value (default 2).
func (cl CustomLayout) GetGutter() int {
	if cl.Gutter == nil {
		return 2
	}
	return *cl.Gutter
}

// GetPadX returns the padX override, or -1 if unset.
func (cl CustomLayout) GetPadX() int {
	if cl.PadX == nil {
		return -1
	}
	return *cl.PadX
}

// GetPadY returns the padY override, or -1 if unset.
func (cl CustomLayout) GetPadY() int {
	if cl.PadY == nil {
		return -1
	}
	return *cl.PadY
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
	Layout    Layout `yaml:"layout"`
	Ratio     string `yaml:"ratio"`
	Align     Align  `yaml:"align"`
	Title     string `yaml:"title"`
	Class     string `yaml:"class"`
	AutoSplit *bool  `yaml:"autosplit"` // pointer so we can detect unset vs false
}

// GetAutoSplit returns the effective autosplit setting (default true).
func (s SlideMeta) GetAutoSplit() bool {
	if s.AutoSplit == nil {
		return true
	}
	return *s.AutoSplit
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
	BlockTaskList
	BlockTable
	BlockAlert
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
