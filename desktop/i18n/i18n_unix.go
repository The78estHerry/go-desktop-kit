//go:build !windows

// Status: implemented, not yet tested on real hardware.

package i18n

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func systemLanguages() []string {
	var out []string
	if runtime.GOOS == "darwin" {
		out = appleLanguages()
	}
	return append(out, envLanguages(os.Getenv)...)
}

// envLanguages follows gettext: LANGUAGE (a colon-separated list) is used
// only when a locale is set at all, then the first of LC_ALL, LC_MESSAGES
// and LANG.
func envLanguages(getenv func(string) string) []string {
	var locale string
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := getenv(k); v != "" {
			locale = v
			break
		}
	}
	if locale == "" || locale == "C" || locale == "POSIX" {
		return nil
	}
	var out []string
	if l := getenv("LANGUAGE"); l != "" {
		out = strings.Split(l, ":")
	}
	return append(out, locale)
}

// appleLanguages reads the macOS language order, printed as
// ( "zh-Hant-TW", "en-US" ).
func appleLanguages() []string {
	b, err := exec.Command("defaults", "read", "-g", "AppleLanguages").Output()
	if err != nil {
		return nil
	}
	return parseAppleLanguages(string(b))
}

func parseAppleLanguages(s string) []string {
	var out []string
	for _, f := range strings.FieldsFunc(s, func(r rune) bool {
		return strings.ContainsRune("(),\" \t\r\n", r)
	}) {
		out = append(out, f)
	}
	return out
}
