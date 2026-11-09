package paths

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDirs(t *testing.T) {
	base := t.TempDir()
	var wantConfig, wantCache string
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", filepath.Join(base, "Roaming"))
		t.Setenv("LocalAppData", filepath.Join(base, "Local"))
		wantConfig, wantCache = filepath.Join(base, "Roaming", "app"), filepath.Join(base, "Local", "app")
	case "darwin":
		t.Setenv("HOME", base)
		wantConfig = filepath.Join(base, "Library", "Application Support", "app")
		wantCache = filepath.Join(base, "Library", "Caches", "app")
	default:
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "config"))
		t.Setenv("XDG_CACHE_HOME", filepath.Join(base, "cache"))
		wantConfig, wantCache = filepath.Join(base, "config", "app"), filepath.Join(base, "cache", "app")
	}
	if got, err := ConfigDir("app"); err != nil || got != wantConfig {
		t.Errorf("ConfigDir = %q, %v; want %q", got, err, wantConfig)
	}
	if got, err := CacheDir("app"); err != nil || got != wantCache {
		t.Errorf("CacheDir = %q, %v; want %q", got, err, wantCache)
	}
}

func TestInvalidAppID(t *testing.T) {
	for _, id := range []string{"", ".", "..", "a/b", `a\b`, "c:"} {
		if _, err := ConfigDir(id); !errors.Is(err, ErrInvalidAppID) {
			t.Errorf("ConfigDir(%q): got %v", id, err)
		}
	}
}

func TestEnsureDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b")
	got, err := EnsureDir(dir)
	if err != nil || got != dir {
		t.Fatalf("EnsureDir = %q, %v", got, err)
	}
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		t.Fatalf("not a directory: %v", err)
	}
	if runtime.GOOS != "windows" && fi.Mode().Perm() != 0o700 {
		t.Errorf("mode = %v, want 0700", fi.Mode().Perm())
	}
}
