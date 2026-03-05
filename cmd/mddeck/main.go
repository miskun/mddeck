package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	autoAdvance := flag.String("auto-advance", "", "Auto-advance slides after duration (e.g., 30s, 1m)")
	loop := flag.Bool("loop", false, "Loop back to first slide when auto-advance reaches the end")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "mddeck – Terminal-native Markdown slide decks\n\n")
		fmt.Fprintf(os.Stderr, "Usage: mddeck [flags] <file.mddeck>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
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

	// Parse auto-advance duration
	var autoAdvanceDuration time.Duration
	if *autoAdvance != "" {
		parsedDuration, err := time.ParseDuration(*autoAdvance)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid auto-advance duration: %s\n", *autoAdvance)
			fmt.Fprintf(os.Stderr, "Use format like: 30s, 1m, 1m30s\n")
			os.Exit(1)
		}
		if parsedDuration < time.Second {
			fmt.Fprintf(os.Stderr, "Error: auto-advance duration must be at least 1s\n")
			os.Exit(1)
		}
		autoAdvanceDuration = parsedDuration
	}

	// Build runtime config
	cfg := runtime.Config{
		Presenter:   *present || *presentShort,
		StartAt:     *startAt,
		Theme:       *themeName,
		AutoAdvance: autoAdvanceDuration,
		Loop:        *loop,
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
