//go:build !windows

package autostart

import "testing"

// useScratchLocation points HOME and XDG_CONFIG_HOME at a temporary
// directory so the test never touches the real login items.
func useScratchLocation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
}
