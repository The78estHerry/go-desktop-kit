//go:build linux || freebsd || netbsd || openbsd || dragonfly

// Status: implemented, not yet tested on real hardware.

package autostart

import (
	"os"
	"path/filepath"
)

// entryPath follows the XDG Autostart Specification. os.UserConfigDir
// honours $XDG_CONFIG_HOME and falls back to ~/.config.
func entryPath(id string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "autostart", id+".desktop"), nil
}

func entryContent(e Entry) string { return desktopEntry(e) }
