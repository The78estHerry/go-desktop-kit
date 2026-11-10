package i18n

import (
	"slices"
	"testing"
)

func TestCanonical(t *testing.T) {
	for in, want := range map[string]string{
		"zh_tw.UTF-8":    "zh-TW",
		"zh-hant-tw":     "zh-Hant-TW",
		"EN-us":          "en-US",
		"ja_JP.eucJP@x":  "ja-JP",
		"C":              "",
		"POSIX":          "",
		"es-419":         "es-419",
		"":               "",
		"de_DE@euro":     "de-DE",
		"zh-Hans-CN.gbk": "zh-Hans-CN",
	} {
		if got := Canonical(in); got != want {
			t.Errorf("Canonical(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMatch(t *testing.T) {
	have := []string{"en", "zh-TW", "ja"}
	for _, c := range []struct {
		want []string
		out  string
	}{
		{[]string{"zh-TW"}, "zh-TW"},
		{[]string{"zh-Hant-TW"}, "zh-TW"},
		{[]string{"zh-HK"}, "zh-TW"},
		{[]string{"zh-CN"}, "zh-TW"}, // only Chinese available
		{[]string{"ja-JP"}, "ja"},
		{[]string{"en-GB"}, "en"},
		{[]string{"fr-FR", "ja-JP"}, "ja"},
		{[]string{"fr-FR"}, "en"},
		{nil, "en"},
	} {
		if got := Match(c.want, have, "en"); got != c.out {
			t.Errorf("Match(%v) = %q, want %q", c.want, got, c.out)
		}
	}
	// With both scripts available, each reader gets their own.
	both := []string{"en", "zh-TW", "zh-CN"}
	if got := Match([]string{"zh-SG"}, both, "en"); got != "zh-CN" {
		t.Errorf("zh-SG matched %q, want zh-CN", got)
	}
	if got := Match([]string{"zh-Hant"}, both, "en"); got != "zh-TW" {
		t.Errorf("zh-Hant matched %q, want zh-TW", got)
	}
}

func TestCatalog(t *testing.T) {
	c := New("en")
	c.Add("en", map[string]string{"hello": "Hello, %s", "bye": "Bye"})
	c.Add("zh_TW", map[string]string{"hello": "你好，%s"})
	c.Add("ja", map[string]string{"hello": "こんにちは、%s", "bye": "さようなら"})

	if got := c.Languages(); !slices.Equal(got, []string{"en", "zh-TW", "ja"}) {
		t.Errorf("Languages = %v", got)
	}
	if got := c.T("hello", "Ann"); got != "Hello, Ann" {
		t.Errorf("default language: %q", got)
	}
	if got := c.SetLanguage("zh-Hant-TW"); got != "zh-TW" {
		t.Errorf("SetLanguage = %q", got)
	}
	if got := c.T("hello", "Ann"); got != "你好，Ann" {
		t.Errorf("zh-TW: %q", got)
	}
	if got := c.T("bye"); got != "Bye" {
		t.Errorf("missing translation should fall back to English: %q", got)
	}
	if got := c.T("no.such.key"); got != "no.such.key" {
		t.Errorf("unknown key: %q", got)
	}
	if got := c.Missing("zh-TW"); !slices.Equal(got, []string{"bye"}) {
		t.Errorf("Missing(zh-TW) = %v", got)
	}
	if got := c.Missing("ja"); len(got) != 0 {
		t.Errorf("Missing(ja) = %v", got)
	}
	if got := c.SetLanguage("fr"); got != "en" {
		t.Errorf("SetLanguage(fr) = %q, want en", got)
	}
}

func TestSystemLanguages(t *testing.T) {
	// Only checks the result is clean; the content depends on the machine.
	for _, l := range SystemLanguages() {
		if l == "" || l != Canonical(l) {
			t.Errorf("SystemLanguages returned %q", l)
		}
	}
	t.Logf("system languages: %v", SystemLanguages())
}
