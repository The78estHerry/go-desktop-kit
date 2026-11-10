//go:build !windows && !darwin

// Status: implemented, not yet tested on real hardware.

package tray

import (
	"runtime"

	"fyne.io/systray"
)

// start talks to the StatusNotifierItem host over D-Bus; desktops without
// one (GNOME without the AppIndicator extension) show nothing.
func start(ready, onExit func()) {
	go func() {
		runtime.LockOSThread()
		systray.Run(ready, onExit)
	}()
}

func platformIcon(png []byte) []byte { return png }
