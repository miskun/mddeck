// Package model defines the core data types for mddeck.
package model

// Layout represents slide layout mode.
type Layout string

const (
	LayoutAuto      Layout = "auto"
	LayoutDefault   Layout = "default"
	LayoutTitle     Layout = "title"
	LayoutCenter    Layout = "center"
	LayoutCols2     Layout = "cols-2"
	LayoutRows2     Layout = "rows-2"
	LayoutTerminal  Layout = "terminal"
	LayoutSidebar   Layout = "sidebar"
	LayoutCols3     Layout = "cols-3"
	LayoutGrid4     Layout = "grid-4"
	LayoutTitleCols2 Layout = "title-cols-2"
	LayoutTitleCols3 Layout = "title-cols-3"
	LayoutTitleGrid4 Layout = "title-grid-4"
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
	Title            string                   `yaml:"title"`
	Theme            string                   `yaml:"theme"`
	Aspect           string                   `yaml:"aspect"`           // target aspect ratio, e.g. "16:9" or "4:3"
	Wrap             *bool                    `yaml:"wrap"`             // pointer so we can detect unset vs false
	TabSize          *int                     `yaml:"tabSize"`          // pointer so we can detect unset vs 0
	SlideWidth       *int                     `yaml:"slideWidth"`       // slide stage width in chars (default 80, 0 = fill terminal, -1 = auto from aspect)
	SlideHeight      *int                     `yaml:"slideHeight"`      // slide stage height in chars (default -1/auto, 0 = fill terminal)
	SafeAnsi         *bool                    `yaml:"safeAnsi"`         // pointer so we can detect unset vs false
	IncrementalLists *bool                    `yaml:"incrementalLists"` // auto-reveal list items one by one (default false)
	DisableReveal    *bool                    `yaml:"disableReveal"`    // disable all reveal effects (pause markers + incremental lists)
	Layouts          map[string]CustomLayout  `yaml:"layouts"`          // user-defined or overridden layouts
	Footer           Footer                   `yaml:"footer"`           // configurable footer sections
}

// Footer defines the three sections of the slide footer bar.
// Left and Center are static text from frontmatter.
// Right defaults to the slide counter ("N / M") if not set.
type Footer struct {
	Left   string `yaml:"left"`
	Center string `yaml:"center"`
	Right  string `yaml:"right"`
}

// CustomLayout defines a user-configurable grid layout.
// Columns and Rows define the grid cell sizes as percentages.
// - columns only → single-row, N columns
// - rows only → single-column, N rows
// - both → len(columns) × len(rows) grid (row-major order)
// - grid → per-row column definitions (overrides columns/rows)
type CustomLayout struct {
	Columns []int       `yaml:"columns"` // column widths as percentages, e.g. [30, 70]
	Rows    []int       `yaml:"rows"`    // row heights as percentages, e.g. [60, 40]
	Grid    []LayoutRow `yaml:"grid"`    // per-row column definitions (overrides columns/rows)
	Gutter  *int        `yaml:"gutter"`  // gap between cells in characters (default: 2)
	PadX    *int        `yaml:"padX"`    // horizontal padding override
	PadY    *int        `yaml:"padY"`    // vertical padding override
	Align   Align       `yaml:"align"`   // content alignment within cells
}

// LayoutRow defines a single row in a grid layout with its own column structure.
// Height semantics:
//   - positive: percentage of available space (after fixed rows are subtracted)
//   - negative: fixed height in rows (absolute value), e.g. -1 = 1 row
//   - zero: equal share of remaining percentage space
type LayoutRow struct {
	Height  int   `yaml:"height"`  // row height: >0 = percentage, <0 = fixed rows, 0 = equal share
	Columns []int `yaml:"columns"` // column widths for this row as percentages
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

// GetSlideWidth returns the effective slide width.
// Default is 80 (balanced for readability and multi-column layouts).
// 0 = fill terminal width, -1 = auto-calculate from slide height + aspect ratio.
func (d DeckMeta) GetSlideWidth() int {
	if d.SlideWidth == nil {
		return 80
	}
	return *d.SlideWidth
}

// GetSlideHeight returns the effective slide height.
// Default is -1 (auto-calculate from slide width + aspect ratio).
// 0 = fill terminal height.
func (d DeckMeta) GetSlideHeight() int {
	if d.SlideHeight == nil {
		return -1
	}
	return *d.SlideHeight
}

// GetIncrementalLists returns the effective incrementalLists setting (default false).
func (d DeckMeta) GetIncrementalLists() bool {
	if d.IncrementalLists == nil {
		return false
	}
	return *d.IncrementalLists
}

// GetDisableReveal returns the effective disableReveal setting (default false).
func (d DeckMeta) GetDisableReveal() bool {
	if d.DisableReveal == nil {
		return false
	}
	return *d.DisableReveal
}

// SlideMeta holds per-slide frontmatter.
type SlideMeta struct {
	Layout           Layout `yaml:"layout"`
	Ratio            string `yaml:"ratio"`
	Align            Align  `yaml:"align"`
	Title            string `yaml:"title"`
	Class            string `yaml:"class"`
	AutoSplit        *bool  `yaml:"autosplit"`        // pointer so we can detect unset vs false
	IncrementalLists *bool  `yaml:"incrementalLists"` // per-slide override for incremental lists
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
	BlockRegionBreak // region divider for multi-region layouts (not rendered)
)

// String returns a human-readable kebab-case name for a BlockType.
func (bt BlockType) String() string {
	switch bt {
	case BlockParagraph:
		return "paragraph"
	case BlockHeading:
		return "heading"
	case BlockUnorderedList:
		return "unordered-list"
	case BlockOrderedList:
		return "ordered-list"
	case BlockBlockquote:
		return "blockquote"
	case BlockFencedCode:
		return "fenced-code"
	case BlockHorizontalRule:
		return "horizontal-rule"
	case BlockANSIArt:
		return "ansi-art"
	case BlockASCIIArt:
		return "ascii-art"
	case BlockBrailleArt:
		return "braille-art"
	case BlockTaskList:
		return "task-list"
	case BlockTable:
		return "table"
	case BlockAlert:
		return "alert"
	case BlockRegionBreak:
		return "region-break"
	default:
		return "unknown"
	}
}

// Block represents a parsed content block.
type Block struct {
	Type     BlockType
	Raw      string   // raw text content
	Level    int      // heading level (1-3) or list indent
	Language string   // fenced code block language
	Lines    []string // individual lines (for lists, blockquotes)
	Children []Block  // nested blocks (for blockquotes)
	NoHeader  bool     // table without header row
	Step      int      // reveal step (0 = always visible, 1+ = revealed on click)
	ListStart int      // 1-based starting index for split list items (0 = default)
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
	Steps  int    // total number of reveal steps (0 = no progressive reveal)
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
