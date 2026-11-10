//go:build windows

package autostart

import (
	"errors"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// runKey is a variable so tests can use a scratch key instead of the real
// Run key.
var runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// command quotes the program and arguments the way the Run key expects.
func command(e Entry) string {
	parts := []string{`"` + e.Exe + `"`}
	for _, a := range e.Args {
		parts = append(parts, syscall.EscapeArg(a))
	}
	return strings.Join(parts, " ")
}

func enable(e Entry) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(e.ID, command(e))
}

func disable(id string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.DeleteValue(id); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}

func enabled(id string) (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer k.Close()
	_, _, err = k.GetStringValue(id)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}
