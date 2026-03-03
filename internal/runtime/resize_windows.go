//go:build windows

package runtime

import (
	"os"
	"time"

	"golang.org/x/term"
)

// watchResize polls terminal size on Windows (no SIGWINCH signal available).
// Redraws when a size change is detected.
func (rt *Runtime) watchResize() {
	go func() {
		lastW, lastH, _ := term.GetSize(int(os.Stdout.Fd()))

		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if !rt.running {
				return
			}
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				continue
			}
			if w != lastW || h != lastH {
				lastW, lastH = w, h
				select {
				case rt.resizeCh <- struct{}{}:
				default:
				}
			}
		}
	}()
}
