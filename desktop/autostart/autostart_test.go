package autostart

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCheckID(t *testing.T) {
	for _, id := range []string{"", "a/b", "a b", "../x", strings.Repeat("a", 65)} {
		if err := Disable(id); !errors.Is(err, ErrInvalidID) {
			t.Errorf("Disable(%q) = %v, want ErrInvalidID", id, err)
		}
	}
}

func TestDesktopEntry(t *testing.T) {
	got := desktopEntry(Entry{ID: "my-app", Name: "My\nApp", Exe: "/opt/my app/run", Args: []string{"--tray", `a"b$c\d`, "50%"}})
	want := "[Desktop Entry]\nType=Application\nName=My App\n" +
		`Exec="/opt/my app/run" "--tray" "a\\"b\\$c\\\\d" "50%%"` +
		"\nX-GNOME-Autostart-enabled=true\n"
	if got != want {
		t.Errorf("desktopEntry =\n%s\nwant\n%s", got, want)
	}
}

func TestLaunchAgent(t *testing.T) {
	got := launchAgent(Entry{ID: "com.example.app", Exe: "/Applications/A&B.app/Contents/MacOS/ab", Args: []string{"<x>"}})
	for _, s := range []string{
		"<string>com.example.app</string>",
		"<string>/Applications/A&amp;B.app/Contents/MacOS/ab</string>",
		"<string>&lt;x&gt;</string>",
		"<key>RunAtLoad</key>\n\t<true/>",
	} {
		if !strings.Contains(got, s) {
			t.Errorf("launchAgent output lacks %q:\n%s", s, got)
		}
	}
}

// TestEnableDisable writes a real entry: in a scratch registry key on
// Windows, in a temporary home directory elsewhere.
func TestEnableDisable(t *testing.T) {
	useScratchLocation(t)
	id := fmt.Sprintf("go-desktop-kit-test-%d", os.Getpid())

	if on, err := Enabled(id); err != nil || on {
		t.Fatalf("Enabled before Enable = %v, %v", on, err)
	}
	if err := Enable(Entry{ID: id, Exe: "/nonexistent/test-program", Args: []string{"--tray"}}); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if on, err := Enabled(id); err != nil || !on {
		t.Fatalf("Enabled after Enable = %v, %v", on, err)
	}
	if err := Disable(id); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if err := Disable(id); err != nil {
		t.Fatalf("second Disable: %v", err)
	}
	if on, err := Enabled(id); err != nil || on {
		t.Fatalf("Enabled after Disable = %v, %v", on, err)
	}
}
