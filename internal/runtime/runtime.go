// Package runtime implements the terminal runtime for mddeck presentations.
package runtime

import (
	"fmt"
	"os"
	"time"

	a "github.com/miskun/mddeck/internal/ansi"
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
	step      int  // current reveal step within slide
	mode      Mode
	prevMode  Mode // mode before help overlay
	startTime time.Time
	running   bool

	// Channel for resize events
	resizeCh chan struct{}

	// Previous frame for diff-based rendering
	prevLines  []string
	prevWidth  int
	prevHeight int

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
		resizeCh:  make(chan struct{}, 1),
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

	// Start platform-specific resize watcher (writes to rt.resizeCh)
	rt.watchResize()

	// Enter alternate screen buffer (prevents reflow on resize) and hide cursor
	fmt.Print("\x1b[?1049h\x1b[?25l")

	rt.running = true
	rt.render()

	// Read input in a goroutine so we can also receive resize events
	inputCh := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 32)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				close(inputCh)
				return
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			inputCh <- data
		}
	}()

	// Debounce timer for resize events — coalesce rapid SIGWINCH bursts.
	// We wait until no new resize arrives for 50ms, then render. This avoids
	// the race where we measure terminal size, the terminal resizes again
	// mid-render, and our output wraps at the wrong width.
	var resizeTimer *time.Timer
	resizeTimerCh := make(<-chan time.Time)

	// 1-second ticker to update the elapsed timer in presenter mode.
	timerTick := time.NewTicker(1 * time.Second)
	defer timerTick.Stop()

	for rt.running {
		select {
		case data, ok := <-inputCh:
			if !ok {
				rt.running = false
				continue
			}
			rt.handleInput(data)
		case <-rt.resizeCh:
			// Reset debounce timer on each resize event
			if resizeTimer != nil {
				resizeTimer.Stop()
			}
			resizeTimer = time.NewTimer(100 * time.Millisecond)
			resizeTimerCh = resizeTimer.C
			// Invalidate previous frame so debounce-triggered render
			// does a full redraw (no stale diff state)
			rt.prevLines = nil
		case <-resizeTimerCh:
			resizeTimerCh = make(<-chan time.Time) // disarm
			rt.render()
		case <-timerTick.C:
			if rt.mode == ModePresenter {
				rt.render()
			}
		}
	}

	return nil
}

// cleanup restores terminal state.
func (rt *Runtime) cleanup() {
	// Show cursor and leave alternate screen buffer (restores original content)
	fmt.Print("\x1b[?25h\x1b[?1049l")
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
	slide := &rt.Deck.Slides[rt.current]
	if rt.step < slide.Steps {
		// Advance to next reveal step within current slide
		rt.step++
		rt.render()
	} else if rt.current < len(rt.Deck.Slides)-1 {
		rt.current++
		rt.step = 0
		rt.render()
	}
}

func (rt *Runtime) prevSlide() {
	if rt.step > 0 {
		// Go back one reveal step within current slide
		rt.step--
		rt.render()
	} else if rt.current > 0 {
		rt.current--
		// Jump to the last step of the previous slide
		rt.step = rt.Deck.Slides[rt.current].Steps
		rt.render()
	}
}

func (rt *Runtime) firstSlide() {
	rt.current = 0
	rt.step = 0
	rt.render()
}

func (rt *Runtime) lastSlide() {
	rt.current = len(rt.Deck.Slides) - 1
	rt.step = rt.Deck.Slides[rt.current].Steps
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

// render draws the current state to the terminal using diff-based updates.
// Only lines that differ from the previous frame are written, minimizing
// terminal I/O and eliminating flicker.
func (rt *Runtime) render() {
	vp := rt.getViewport()
	slide := &rt.Deck.Slides[rt.current]

	var lines []string
	var baseFg string
	switch rt.mode {
	case ModePresenter:
		elapsed := rt.formatElapsed()
		lines = rt.Renderer.RenderPresenter(slide, vp, elapsed, rt.step)
		baseFg = rt.Renderer.Theme.Fg
	case ModeHelp:
		lines = rt.Renderer.RenderHelp(vp)
	default:
		lines = rt.Renderer.RenderSlide(slide, vp, rt.step)
		baseFg = rt.Renderer.Theme.Fg
	}

	var output string
	if rt.prevWidth != vp.Width || rt.prevHeight != vp.Height || rt.prevLines == nil {
		// Viewport changed or first render: write full frame sequentially
		output = render.RenderFull(lines, baseFg)
	} else {
		// Same viewport: diff against previous frame
		output = render.RenderDiff(rt.prevLines, lines, baseFg, vp.Width)
	}

	rt.prevLines = lines
	rt.prevWidth = vp.Width
	rt.prevHeight = vp.Height

	// Wrap in synchronized output markers and write as a single syscall
	buf := a.BeginSync + output + a.EndSync
	os.Stdout.Write([]byte(buf))
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

	// Clamp step to current slide's steps
	if rt.current < len(deck.Slides) {
		if rt.step > deck.Slides[rt.current].Steps {
			rt.step = deck.Slides[rt.current].Steps
		}
	} else {
		rt.step = 0
	}

	rt.render()
}
