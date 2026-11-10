//go:build darwin

// Status: implemented, not yet tested on real hardware.

package autostart

import (
	"os"
	"path/filepath"
)

// The agent is loaded at the next login; it is not started now.
func entryPath(id string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", id+".plist"), nil
}

func entryContent(e Entry) string { return launchAgent(e) }
