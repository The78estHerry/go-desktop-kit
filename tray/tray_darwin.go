//go:build darwin

// Status: implemented, not yet tested on real hardware.

package tray

import "fyne.io/systray"

// start registers the tray with the application's existing main loop
// (AppKit runs on the main thread, which the GUI toolkit already owns).
// Without a GUI toolkit, the caller's main goroutine must run one, for
// example by blocking in the toolkit's main function.
func start(ready, onExit func()) {
	run, _ := systray.RunWithExternalLoop(ready, onExit)
	run()
}

func platformIcon(png []byte) []byte { return png }
