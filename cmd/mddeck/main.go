package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/miskun/mddeck/internal/model"
	"github.com/miskun/mddeck/internal/parser"
	"github.com/miskun/mddeck/internal/runtime"
)

var version = "1.0.0"

func main() {
	// Flags
	present := flag.Bool("present", false, "Start in presenter mode")
	presentShort := flag.Bool("p", false, "Start in presenter mode (short)")
	themeName := flag.String("theme", "", "Override theme")
	safeAnsi := flag.Bool("safe-ansi", false, "Force safe ANSI mode")
	unsafeAnsi := flag.Bool("unsafe-ansi", false, "Disable safe ANSI mode")
	startAt := flag.Int("start", 0, "Start at slide number (1-based)")
	watch := flag.Bool("watch", false, "Reload on file change")
	showVersion := flag.Bool("version", false, "Show version")

	// Dump mode flags
	dump := flag.Bool("dump", false, "Dump slide data to stdout and exit")
	format := flag.String("format", "text", "Dump format: text, json")
	slideN := flag.Int("slide", 0, "Dump only slide N (1-based, 0=all)")
	width := flag.Int("width", 0, "Virtual terminal width (0=auto)")
	height := flag.Int("height", 0, "Virtual terminal height (0=auto)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "mddeck – Terminal-native Markdown slide decks\n\n")
		fmt.Fprintf(os.Stderr, "Usage: mddeck [flags] <file.mddeck>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDump mode:\n")
		fmt.Fprintf(os.Stderr, "  --dump                Dump slide data to stdout and exit\n")
		fmt.Fprintf(os.Stderr, "  --format text|json    Output format (default: text)\n")
		fmt.Fprintf(os.Stderr, "  --slide N             Dump only slide N (1-based, 0=all)\n")
		fmt.Fprintf(os.Stderr, "  --width W             Virtual terminal width (0=auto)\n")
		fmt.Fprintf(os.Stderr, "  --height H            Virtual terminal height (0=auto)\n")
		fmt.Fprintf(os.Stderr, "\nKeys:\n")
		fmt.Fprintf(os.Stderr, "  Space/Enter/→/PgDn/n  Next slide\n")
		fmt.Fprintf(os.Stderr, "  Backspace/←/PgUp/p    Previous slide\n")
		fmt.Fprintf(os.Stderr, "  Home/End              First/Last slide\n")
		fmt.Fprintf(os.Stderr, "  t                     Toggle presenter mode\n")
		fmt.Fprintf(os.Stderr, "  ?                     Help\n")
		fmt.Fprintf(os.Stderr, "  q/Ctrl+C              Quit\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("mddeck v%s\n", version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: no input file specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	filePath := flag.Arg(0)

	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
		os.Exit(1)
	}

	// Parse the deck
	deck, err := parser.ParseFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Dump mode — output and exit without entering TUI
	if *dump {
		if *format != "text" && *format != "json" {
			fmt.Fprintf(os.Stderr, "Error: invalid format %q (use \"text\" or \"json\")\n", *format)
			os.Exit(1)
		}
		if *slideN < 0 || *slideN > len(deck.Slides) {
			fmt.Fprintf(os.Stderr, "Error: slide %d out of range (deck has %d slides)\n", *slideN, len(deck.Slides))
			os.Exit(1)
		}
		runDump(deck, *format, *slideN)
		os.Exit(0)
	}

	// Build runtime config
	cfg := runtime.Config{
		Presenter: *present || *presentShort,
		StartAt:   *startAt,
		Theme:     *themeName,
		Width:     *width,
		Height:    *height,
	}

	if *safeAnsi {
		t := true
		cfg.SafeANSI = &t
	}
	if *unsafeAnsi {
		f := false
		cfg.SafeANSI = &f
	}

	// Create and run the runtime
	rt := runtime.New(deck, cfg)

	// File watcher
	if *watch {
		go watchFile(filePath, rt)
	}

	if err := rt.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// watchFile polls for file changes and triggers reload.
func watchFile(path string, rt *runtime.Runtime) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	lastMod := time.Time{}
	info, err := os.Stat(absPath)
	if err == nil {
		lastMod = info.ModTime()
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}

		if info.ModTime().After(lastMod) {
			lastMod = info.ModTime()

			deck, err := parser.ParseFile(absPath)
			if err != nil {
				continue // skip reload on parse error
			}

			rt.Reload(deck)
		}
	}
}

// --- Dump mode ---

func runDump(deck *model.Deck, format string, slideNum int) {
	var slides []model.Slide
	if slideNum > 0 {
		slides = []model.Slide{deck.Slides[slideNum-1]}
	} else {
		slides = deck.Slides
	}

	switch format {
	case "json":
		dumpJSON(os.Stdout, deck, slides)
	default:
		dumpText(os.Stdout, deck, slides)
	}
}

// --- Text dump ---

func dumpText(w io.Writer, deck *model.Deck, slides []model.Slide) {
	for i, slide := range slides {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "=== Slide %d ===\n", slide.Index+1)
		fmt.Fprintf(w, "[layout: %s, align: %s, steps: %d]\n", slide.Meta.Layout, slide.Meta.Align, slide.Steps)

		lastStep := 0
		for _, block := range slide.Blocks {
			if block.Step > 0 && block.Step != lastStep {
				fmt.Fprintf(w, "\n[step %d]\n", block.Step)
				lastStep = block.Step
			}
			text := blockToText(block)
			if text != "" {
				fmt.Fprintf(w, "\n%s\n", text)
			}
		}

		if slide.Notes != "" {
			fmt.Fprintf(w, "\n--- Notes ---\n%s\n", slide.Notes)
		}
	}
}

func blockToText(block model.Block) string {
	switch block.Type {
	case model.BlockHeading:
		prefix := strings.Repeat("#", block.Level)
		return prefix + " " + block.Raw

	case model.BlockParagraph:
		return block.Raw

	case model.BlockFencedCode:
		lang := ""
		if block.Language != "" {
			lang = block.Language
		}
		return "```" + lang + "\n" + block.Raw + "\n```"

	case model.BlockUnorderedList:
		var lines []string
		for _, item := range block.Lines {
			depth, text := parseTextListItem(item)
			indent := strings.Repeat("  ", depth)
			lines = append(lines, indent+"- "+text)
		}
		return strings.Join(lines, "\n")

	case model.BlockOrderedList:
		var lines []string
		num := block.ListStart
		if num == 0 {
			num = 1
		}
		for _, item := range block.Lines {
			depth, text := parseTextListItem(item)
			indent := strings.Repeat("  ", depth)
			lines = append(lines, fmt.Sprintf("%s%d. %s", indent, num, text))
			if depth == 0 {
				num++
			}
		}
		return strings.Join(lines, "\n")

	case model.BlockTaskList:
		var lines []string
		for _, item := range block.Lines {
			depth, checked, text := parseTextTaskItem(item)
			indent := strings.Repeat("  ", depth)
			check := "[ ]"
			if checked {
				check = "[x]"
			}
			lines = append(lines, indent+"- "+check+" "+text)
		}
		return strings.Join(lines, "\n")

	case model.BlockBlockquote:
		var lines []string
		for _, line := range block.Lines {
			lines = append(lines, "> "+line)
		}
		return strings.Join(lines, "\n")

	case model.BlockHorizontalRule:
		return "---"

	case model.BlockANSIArt, model.BlockASCIIArt, model.BlockBrailleArt:
		lang := block.Type.String()
		return "```" + lang + "\n" + block.Raw + "\n```"

	case model.BlockTable:
		return block.Raw

	case model.BlockAlert:
		return block.Raw

	case model.BlockRegionBreak:
		return "<!-- region-break -->"

	default:
		return block.Raw
	}
}

// parseTextListItem decodes "DEPTH:text" list encoding.
func parseTextListItem(item string) (int, string) {
	if len(item) >= 2 && item[0] >= '0' && item[0] <= '9' && item[1] == ':' {
		return int(item[0] - '0'), item[2:]
	}
	return 0, item
}

// parseTextTaskItem decodes "DEPTH:C:text" task list encoding.
func parseTextTaskItem(item string) (int, bool, string) {
	if len(item) >= 4 && item[0] >= '0' && item[0] <= '9' && item[1] == ':' && item[3] == ':' {
		depth := int(item[0] - '0')
		checked := item[2] == '1'
		return depth, checked, item[4:]
	}
	return 0, false, item
}

// --- JSON dump ---

type dumpDeck struct {
	Source     string      `json:"source"`
	SlideCount int         `json:"slideCount"`
	Meta       dumpDeckMeta `json:"meta"`
	Slides     []dumpSlide `json:"slides"`
}

type dumpDeckMeta struct {
	Title            string     `json:"title"`
	Theme            string     `json:"theme"`
	Wrap             bool       `json:"wrap"`
	TabSize          int        `json:"tabSize"`
	SlideWidth       int        `json:"slideWidth"`
	SlideHeight      int        `json:"slideHeight"`
	SafeAnsi         bool       `json:"safeAnsi"`
	IncrementalLists bool       `json:"incrementalLists"`
	DisableReveal    bool       `json:"disableReveal"`
	Footer           dumpFooter `json:"footer"`
}

type dumpFooter struct {
	Left   string `json:"left"`
	Center string `json:"center"`
	Right  string `json:"right"`
}

type dumpSlide struct {
	Index      int           `json:"index"`
	Meta       dumpSlideMeta `json:"meta"`
	Steps      int           `json:"steps"`
	BlockCount int           `json:"blockCount"`
	Blocks     []dumpBlock   `json:"blocks"`
	Notes      string        `json:"notes"`
	HasNotes   bool          `json:"hasNotes"`
}

type dumpSlideMeta struct {
	Layout           string `json:"layout"`
	Align            string `json:"align"`
	Title            string `json:"title"`
	Class            string `json:"class"`
	AutoSplit        bool   `json:"autoSplit"`
	IncrementalLists *bool  `json:"incrementalLists,omitempty"`
}

type dumpBlock struct {
	Type      string   `json:"type"`
	Level     int      `json:"level,omitempty"`
	Raw       string   `json:"raw"`
	Language  string   `json:"language,omitempty"`
	Step      int      `json:"step,omitempty"`
	Lines     []string `json:"lines,omitempty"`
	NoHeader  bool     `json:"noHeader,omitempty"`
	ListStart int      `json:"listStart,omitempty"`
}

func dumpJSON(w io.Writer, deck *model.Deck, slides []model.Slide) {
	out := dumpDeck{
		Source:     deck.Source,
		SlideCount: len(deck.Slides),
		Meta: dumpDeckMeta{
			Title:            deck.Meta.Title,
			Theme:            deck.Meta.Theme,
			Wrap:             deck.Meta.GetWrap(),
			TabSize:          deck.Meta.GetTabSize(),
			SlideWidth:       deck.Meta.GetSlideWidth(),
			SlideHeight:      deck.Meta.GetSlideHeight(),
			SafeAnsi:         deck.Meta.GetSafeAnsi(),
			IncrementalLists: deck.Meta.GetIncrementalLists(),
			DisableReveal:    deck.Meta.GetDisableReveal(),
			Footer: dumpFooter{
				Left:   deck.Meta.Footer.Left,
				Center: deck.Meta.Footer.Center,
				Right:  deck.Meta.Footer.Right,
			},
		},
	}

	for _, slide := range slides {
		ds := dumpSlide{
			Index:      slide.Index,
			Meta: dumpSlideMeta{
				Layout:           string(slide.Meta.Layout),
				Align:            string(slide.Meta.Align),
				Title:            slide.Meta.Title,
				Class:            slide.Meta.Class,
				AutoSplit:        slide.Meta.GetAutoSplit(),
				IncrementalLists: slide.Meta.IncrementalLists,
			},
			Steps:      slide.Steps,
			BlockCount: len(slide.Blocks),
			Notes:      slide.Notes,
			HasNotes:   slide.Notes != "",
		}

		for _, block := range slide.Blocks {
			db := dumpBlock{
				Type:      block.Type.String(),
				Level:     block.Level,
				Raw:       block.Raw,
				Language:  block.Language,
				Step:      block.Step,
				Lines:     block.Lines,
				NoHeader:  block.NoHeader,
				ListStart: block.ListStart,
			}
			ds.Blocks = append(ds.Blocks, db)
		}
		if ds.Blocks == nil {
			ds.Blocks = []dumpBlock{}
		}

		out.Slides = append(out.Slides, ds)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	enc.Encode(out)
}
