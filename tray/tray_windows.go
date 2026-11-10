//go:build windows

package tray

import (
	"runtime"

	"fyne.io/systray"
)

// start runs the tray on a dedicated goroutine locked to one OS thread: the
// hidden window that owns the icon must be created and pumped on the same
// thread. systray.RunWithExternalLoop would pump on an arbitrary goroutine
// and does not work.
func start(ready, onExit func()) {
	go func() {
		runtime.LockOSThread()
		systray.Run(ready, onExit)
	}()
}

func platformIcon(png []byte) []byte { return pngToICO(png) }
