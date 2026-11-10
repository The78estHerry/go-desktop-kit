// Package i18n translates UI text with a message table.
//
// Each language is a map from a message key to its text. Lookups fall back
// to the fallback language, then to the key itself, so a missing
// translation shows up as English (or as the key) instead of failing:
//
//	cat := i18n.New("en")
//	cat.Add("en", map[string]string{"greet": "Hello, %s"})
//	cat.Add("zh-TW", map[string]string{"greet": "你好，%s"})
//	cat.Add("ja", map[string]string{"greet": "こんにちは、%s"})
//	cat.SetLanguage(i18n.Match(i18n.SystemLanguages(), cat.Languages(), "en"))
//	label := cat.T("greet", name)
//
// Language tags are BCP 47 style ("en", "zh-TW", "ja"); "_" and POSIX
// suffixes such as ".UTF-8" are accepted. Where the language comes from is
// up to the application: SystemLanguages, a setting the user picked, or an
// installer option.
package i18n

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// Catalog holds the messages of every language and the current language.
// It is safe for concurrent use.
type Catalog struct {
	mu       sync.RWMutex
	fallback string
	current  string
	msgs     map[string]map[string]string
	langs    []string
}

// New returns an empty catalog whose fallback language is fallback. The
// current language starts as the fallback.
func New(fallback string) *Catalog {
	fallback = Canonical(fallback)
	return &Catalog{fallback: fallback, current: fallback, msgs: map[string]map[string]string{}}
}

// Add merges messages into a language, adding the language if needed.
func (c *Catalog) Add(lang string, msgs map[string]string) {
	lang = Canonical(lang)
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.msgs[lang]
	if !ok {
		m = map[string]string{}
		c.msgs[lang] = m
		c.langs = append(c.langs, lang)
	}
	for k, v := range msgs {
		m[k] = v
	}
}

// Languages returns the languages added so far, in the order they were added.
func (c *Catalog) Languages() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Clone(c.langs)
}

// SetLanguage switches the current language. A language without messages
// is matched to the closest one available (see Match); the language
// actually used is returned.
func (c *Catalog) SetLanguage(lang string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = Match([]string{lang}, c.langs, c.fallback)
	return c.current
}

// Language returns the current language.
func (c *Catalog) Language() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

// T returns the message for key in the current language. With args, the
// message is a fmt format string.
func (c *Catalog) T(key string, args ...any) string {
	c.mu.RLock()
	s, ok := c.msgs[c.current][key]
	if !ok {
		s, ok = c.msgs[c.fallback][key]
	}
	c.mu.RUnlock()
	if !ok {
		s = key
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

// Missing returns, sorted, the keys the fallback language has and lang
// lacks. Use it in a test to keep translations complete.
func (c *Catalog) Missing(lang string) []string {
	lang = Canonical(lang)
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []string
	for k := range c.msgs[c.fallback] {
		if _, ok := c.msgs[lang][k]; !ok {
			out = append(out, k)
		}
	}
	slices.Sort(out)
	return out
}

// Canonical normalises a language tag: "zh_tw.UTF-8" becomes "zh-TW",
// "zh-hant-tw" becomes "zh-Hant-TW". "C" and "POSIX" become "".
func Canonical(tag string) string {
	tag, _, _ = strings.Cut(tag, ".")
	tag, _, _ = strings.Cut(tag, "@")
	if tag == "C" || tag == "POSIX" {
		return ""
	}
	parts := strings.FieldsFunc(tag, func(r rune) bool { return r == '-' || r == '_' })
	for i, p := range parts {
		switch {
		case i == 0:
			parts[i] = strings.ToLower(p)
		case len(p) == 4: // script
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		case len(p) == 2 || len(p) == 3: // region
			parts[i] = strings.ToUpper(p)
		}
	}
	return strings.Join(parts, "-")
}

// split returns the base language and the script of a canonical tag. The
// script is inferred for Chinese, the one language here where it changes
// which text a reader can use.
func split(tag string) (base, script string) {
	parts := strings.Split(tag, "-")
	base = parts[0]
	region := ""
	for _, p := range parts[1:] {
		switch len(p) {
		case 4:
			script = p
		case 2, 3:
			region = p
		}
	}
	if base == "zh" && script == "" {
		switch region {
		case "TW", "HK", "MO":
			script = "Hant"
		default:
			script = "Hans"
		}
	}
	return base, script
}

// Match picks the language in have that best serves the ordered list want,
// typically SystemLanguages. For each wanted language in turn it tries an
// exact match, then the same language and script, then the same language
// (so a Simplified Chinese reader gets Traditional Chinese when that is the
// only Chinese available). Nothing matching returns fallback.
func Match(want, have []string, fallback string) string {
	canon := make([]string, len(have))
	for i, h := range have {
		canon[i] = Canonical(h)
	}
	for _, w := range want {
		w = Canonical(w)
		if w == "" {
			continue
		}
		if slices.Contains(canon, w) {
			return w
		}
		wb, ws := split(w)
		for _, h := range canon {
			if hb, hs := split(h); hb == wb && hs == ws {
				return h
			}
		}
		for _, h := range canon {
			if hb, _ := split(h); hb == wb {
				return h
			}
		}
	}
	return Canonical(fallback)
}

// SystemLanguages returns the user's preferred UI languages, most preferred
// first, as canonical tags. It may be empty.
func SystemLanguages() []string {
	var out []string
	for _, l := range systemLanguages() {
		if l = Canonical(l); l != "" && !slices.Contains(out, l) {
			out = append(out, l)
		}
	}
	return out
}
