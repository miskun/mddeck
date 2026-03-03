// Package runtime implements the terminal runtime for mddeck presentations.
package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miskun/mddeck/internal/layout"
	"github.com/miskun/mddeck/internal/model"
	"github.com/miskun/mddeck/internal/render"
	"github.com/miskun/mddeck/internal/theme"
	"golang.org/x/term"
)

// Mode represents the runtime display mode.
type Mode int

const (
	ModeAudience  Mode = iota
	ModePresenter
	ModeHelp
)

// Runtime manages the presentation lifecycle.
type Runtime struct {
	Deck     *model.Deck
	Renderer *render.Renderer
	Theme    theme.Theme

	current   int  // current slide index
	mode      Mode
	prevMode  Mode // mode before help overlay
	startTime time.Time
	running   bool

	// Terminal state
	oldState *term.State
}

// Config holds runtime configuration.
type Config struct {
	Presenter bool
	StartAt   int // 1-based slide number
	Theme     string
	SafeANSI  *bool
}

// New creates a new runtime.
func New(deck *model.Deck, cfg Config) *Runtime {
	// Apply config overrides
	if cfg.SafeANSI != nil {
		deck.Meta.SafeAnsi = cfg.SafeANSI
	}

	themeName := deck.Meta.Theme
	if cfg.Theme != "" {
		themeName = cfg.Theme
	}
	th := theme.Get(themeName)

	r := &Runtime{
		Deck:      deck,
		Theme:     th,
		Renderer:  render.NewRenderer(deck, th),
		startTime: time.Now(),
	}

	if cfg.Presenter {
		r.mode = ModePresenter
	}

	if cfg.StartAt > 0 && cfg.StartAt <= len(deck.Slides) {
		r.current = cfg.StartAt - 1
	}

	return r
}

// Run starts the presentation event loop.
func (rt *Runtime) Run() error {
	// Enter raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("entering raw mode: %w", err)
	}
	rt.oldState = oldState

	defer rt.cleanup()

	// Handle SIGINT, SIGTERM, SIGWINCH
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH)

	// Hide cursor
	fmt.Print("\x1b[?25l")

	rt.running = true
	rt.render()

	// Input buffer for escape sequences
	buf := make([]byte, 32)

	for rt.running {
		// Check for signals (non-blocking)
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				rt.running = false
				continue
			case syscall.SIGWINCH:
				rt.render()
				continue
			}
		default:
		}

		// Read input
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		rt.handleInput(buf[:n])
	}

	return nil
}

// cleanup restores terminal state.
func (rt *Runtime) cleanup() {
	// Show cursor
	fmt.Print("\x1b[?25h")
	// Clear screen
	fmt.Print("\x1b[2J\x1b[H")
	// Restore terminal
	if rt.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), rt.oldState)
	}
}

// handleInput processes keyboard input.
func (rt *Runtime) handleInput(data []byte) {
	if len(data) == 0 {
		return
	}

	// If in help mode, any key dismisses
	if rt.mode == ModeHelp {
		rt.mode = rt.prevMode
		rt.render()
		return
	}

	switch {
	// Ctrl+C
	case data[0] == 3:
		rt.running = false

	// 'q'
	case data[0] == 'q' || data[0] == 'Q':
		rt.running = false

	// Space, Enter
	case data[0] == ' ' || data[0] == '\r' || data[0] == '\n':
		rt.nextSlide()

	// 'n'
	case data[0] == 'n' || data[0] == 'N':
		rt.nextSlide()

	// 'p'
	case data[0] == 'p' || data[0] == 'P':
		rt.prevSlide()

	// Backspace
	case data[0] == 127 || data[0] == 8:
		rt.prevSlide()

	// '?'
	case data[0] == '?':
		rt.prevMode = rt.mode
		rt.mode = ModeHelp
		rt.render()

	// 't'
	case data[0] == 't' || data[0] == 'T':
		rt.togglePresenter()

	// Escape sequences
	case data[0] == 27 && len(data) >= 3:
		rt.handleEscapeSequence(data)

	default:
		// Ignore unknown input
	}
}

// handleEscapeSequence processes arrow keys, Home, End, PageUp, PageDown.
func (rt *Runtime) handleEscapeSequence(data []byte) {
	if len(data) < 3 {
		return
	}

	if data[1] == '[' {
		switch data[2] {
		case 'C': // Right arrow
			rt.nextSlide()
		case 'D': // Left arrow
			rt.prevSlide()
		case '5': // Page Up
			if len(data) >= 4 && data[3] == '~' {
				rt.prevSlide()
			}
		case '6': // Page Down
			if len(data) >= 4 && data[3] == '~' {
				rt.nextSlide()
			}
		case 'H': // Home
			rt.firstSlide()
		case 'F': // End
			rt.lastSlide()
		case '1': // Home (alternate)
			if len(data) >= 4 && data[3] == '~' {
				rt.firstSlide()
			}
		case '4': // End (alternate)
			if len(data) >= 4 && data[3] == '~' {
				rt.lastSlide()
			}
		}
	}
}

// Navigation methods
func (rt *Runtime) nextSlide() {
	if rt.current < len(rt.Deck.Slides)-1 {
		rt.current++
		rt.render()
	}
}

func (rt *Runtime) prevSlide() {
	if rt.current > 0 {
		rt.current--
		rt.render()
	}
}

func (rt *Runtime) firstSlide() {
	rt.current = 0
	rt.render()
}

func (rt *Runtime) lastSlide() {
	rt.current = len(rt.Deck.Slides) - 1
	rt.render()
}

func (rt *Runtime) togglePresenter() {
	if rt.mode == ModePresenter {
		rt.mode = ModeAudience
	} else {
		rt.mode = ModePresenter
	}
	rt.render()
}

// render draws the current state to the terminal.
func (rt *Runtime) render() {
	vp := rt.getViewport()
	slide := &rt.Deck.Slides[rt.current]

	var output string
	switch rt.mode {
	case ModePresenter:
		elapsed := rt.formatElapsed()
		output = rt.Renderer.RenderPresenter(slide, vp, elapsed)
	case ModeHelp:
		output = rt.Renderer.RenderHelp(vp)
	default:
		output = rt.Renderer.RenderSlide(slide, vp)
	}

	fmt.Print(output)
}

// getViewport returns the current terminal dimensions.
func (rt *Runtime) getViewport() layout.Viewport {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w, h = 80, 24
	}

	// Apply maxWidth/maxHeight overrides
	if rt.Deck.Meta.MaxWidth > 0 && w > rt.Deck.Meta.MaxWidth {
		w = rt.Deck.Meta.MaxWidth
	}
	if rt.Deck.Meta.MaxHeight > 0 && h > rt.Deck.Meta.MaxHeight {
		h = rt.Deck.Meta.MaxHeight
	}

	return layout.Viewport{Width: w, Height: h}
}

// formatElapsed returns a formatted elapsed time string.
func (rt *Runtime) formatElapsed() string {
	d := time.Since(rt.startTime)
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

// Reload reloads the deck from the given parsed deck.
func (rt *Runtime) Reload(deck *model.Deck) {
	rt.Deck = deck
	rt.Renderer = render.NewRenderer(deck, rt.Theme)

	// Clamp current slide index
	if rt.current >= len(deck.Slides) {
		rt.current = len(deck.Slides) - 1
	}
	if rt.current < 0 {
		rt.current = 0
	}

	rt.render()
}
