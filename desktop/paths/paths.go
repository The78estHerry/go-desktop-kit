// Package paths returns per-user directories for an application:
//
//	          Config                                   Cache
//	Windows   %AppData%\<appID>                        %LocalAppData%\<appID>
//	macOS     ~/Library/Application Support/<appID>    ~/Library/Caches/<appID>
//	Linux     $XDG_CONFIG_HOME/<appID> or ~/.config    $XDG_CACHE_HOME/<appID> or ~/.cache
//
// The directories are not created; use EnsureDir before writing. Created
// directories are readable by the current user only.
package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrInvalidAppID is returned for an empty application ID or one that is not
// a single path element.
var ErrInvalidAppID = errors.New("paths: invalid application ID")

func checkAppID(appID string) error {
	if appID == "" || appID == "." || appID == ".." || strings.ContainsAny(appID, `/\:`) {
		return fmt.Errorf("%w: %q", ErrInvalidAppID, appID)
	}
	return nil
}

// ConfigDir returns the directory for appID's settings.
func ConfigDir(appID string) (string, error) {
	if err := checkAppID(appID); err != nil {
		return "", err
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("paths: %w", err)
	}
	return filepath.Join(base, appID), nil
}

// CacheDir returns the directory for appID's disposable cached data.
func CacheDir(appID string) (string, error) {
	if err := checkAppID(appID); err != nil {
		return "", err
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("paths: %w", err)
	}
	return filepath.Join(base, appID), nil
}

// EnsureDir creates dir and its parents if needed, readable by the current
// user only (mode 0700 on Unix), and returns dir.
func EnsureDir(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("paths: %w", err)
	}
	return dir, nil
}
