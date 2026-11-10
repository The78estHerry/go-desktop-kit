package autostart

import (
	"testing"

	"golang.org/x/sys/windows/registry"
)

// useScratchLocation points the package at a test key so a failed test can
// never leave a real login item behind.
func useScratchLocation(t *testing.T) {
	const scratch = `Software\go-desktop-kit-test\Run`
	old := runKey
	runKey = scratch
	t.Cleanup(func() {
		runKey = old
		registry.DeleteKey(registry.CURRENT_USER, scratch)
		registry.DeleteKey(registry.CURRENT_USER, `Software\go-desktop-kit-test`)
	})
}

func TestCommand(t *testing.T) {
	got := command(Entry{Exe: `C:\Program Files\My App\app.exe`, Args: []string{"--tray", "a b"}})
	want := `"C:\Program Files\My App\app.exe" --tray "a b"`
	if got != want {
		t.Errorf("command = %s, want %s", got, want)
	}
}
