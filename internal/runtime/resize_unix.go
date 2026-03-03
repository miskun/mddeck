//go:build !windows

package runtime

import (
	"os"
	"os/signal"
	"syscall"
)

// watchResize listens for SIGWINCH (terminal resize) and sends on rt.resizeCh.
// Also handles SIGINT/SIGTERM for clean shutdown.
func (rt *Runtime) watchResize() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigCh {
			switch sig {
			case syscall.SIGWINCH:
				// Non-blocking send to avoid deadlock if resize fires rapidly
				select {
				case rt.resizeCh <- struct{}{}:
				default:
				}
			case syscall.SIGINT, syscall.SIGTERM:
				rt.running = false
				// Unblock the input read by closing stdin
				os.Stdin.Close()
				return
			}
		}
	}()
}
