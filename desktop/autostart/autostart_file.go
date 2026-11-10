//go:build darwin || linux || freebsd || netbsd || openbsd || dragonfly

// Status: implemented, not yet tested on real hardware.

package autostart

import (
	"errors"
	"os"
	"path/filepath"
)

// macOS and the freedesktop systems both use one file per entry; only the
// location and the content differ (entryPath and entryContent).

func enable(e Entry) error {
	p, err := entryPath(e.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(entryContent(e)), 0o644)
}

func disable(id string) error {
	p, err := entryPath(id)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func enabled(id string) (bool, error) {
	p, err := entryPath(id)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}
