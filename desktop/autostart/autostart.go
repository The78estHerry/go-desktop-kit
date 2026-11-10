// Package autostart starts an application when the current user logs in.
//
//	Windows  value <ID> under HKCU\Software\Microsoft\Windows\CurrentVersion\Run
//	macOS    ~/Library/LaunchAgents/<ID>.plist (RunAtLoad)
//	Linux    $XDG_CONFIG_HOME/autostart/<ID>.desktop (or ~/.config/autostart)
//
// Nothing needs administrator rights. Example:
//
//	err := autostart.Enable(autostart.Entry{ID: "my-app", Name: "My App", Args: []string{"--tray"}})
package autostart

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrInvalidID is returned for an ID that is empty, too long or contains
// characters other than ASCII letters, digits, '.', '_' and '-'.
var ErrInvalidID = errors.New("autostart: invalid application ID")

// Entry describes what to start at login.
type Entry struct {
	// ID is the application's stable identifier; it names the registry
	// value or file. On macOS a reverse domain name such as
	// "com.example.app" is customary.
	ID string
	// Name is shown by the desktop (Linux) and the login items list.
	// It defaults to ID.
	Name string
	// Exe is the absolute path of the program. It defaults to the running
	// executable.
	Exe string
	// Args are passed to the program.
	Args []string
}

func checkID(id string) error {
	if id == "" || len(id) > 64 {
		return ErrInvalidID
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '_', r == '-':
		default:
			return ErrInvalidID
		}
	}
	return nil
}

// resolve checks the entry and fills in the defaults.
func (e Entry) resolve() (Entry, error) {
	if err := checkID(e.ID); err != nil {
		return e, err
	}
	if e.Name == "" {
		e.Name = e.ID
	}
	if e.Exe == "" {
		exe, err := os.Executable()
		if err != nil {
			return e, fmt.Errorf("autostart: %w", err)
		}
		e.Exe = exe
	}
	return e, nil
}

// Enable makes the entry start at login, replacing an earlier entry with the
// same ID.
func Enable(e Entry) error {
	e, err := e.resolve()
	if err != nil {
		return err
	}
	return enable(e)
}

// Disable removes the entry. Removing an entry that does not exist is not an
// error.
func Disable(id string) error {
	if err := checkID(id); err != nil {
		return err
	}
	return disable(id)
}

// Enabled reports whether an entry with this ID exists.
func Enabled(id string) (bool, error) {
	if err := checkID(id); err != nil {
		return false, err
	}
	return enabled(id)
}

// desktopEntry returns a freedesktop.org .desktop file. Exec arguments are
// quoted as the Desktop Entry Specification requires.
func desktopEntry(e Entry) string {
	args := append([]string{e.Exe}, e.Args...)
	for i, a := range args {
		args[i] = desktopQuote(a)
	}
	name := strings.NewReplacer("\n", " ", "\r", " ").Replace(e.Name)
	return "[Desktop Entry]\nType=Application\nName=" + name +
		"\nExec=" + strings.Join(args, " ") +
		"\nX-GNOME-Autostart-enabled=true\n"
}

// The string-level escape (\\ for \) is undone before the quoting rule, so a
// reserved character is written as \\" and a backslash as four backslashes.
func desktopQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '`', '$':
			b.WriteString(`\\`)
		case '\\':
			b.WriteString(`\\\`)
		case '%': // field codes
			b.WriteByte('%')
		case '\n', '\r':
			r = ' '
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// launchAgent returns a launchd property list that runs the program once at
// login.
func launchAgent(e Entry) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>` + xmlEscape(e.ID) + `</string>
	<key>ProgramArguments</key>
	<array>
`)
	for _, a := range append([]string{e.Exe}, e.Args...) {
		b.WriteString("\t\t<string>" + xmlEscape(a) + "</string>\n")
	}
	b.WriteString(`	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`)
	return b.String()
}

func xmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;").Replace(s)
}
