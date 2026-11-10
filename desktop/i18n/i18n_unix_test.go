//go:build !windows

package i18n

import (
	"slices"
	"testing"
)

func TestEnvLanguages(t *testing.T) {
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}
	for _, c := range []struct {
		env  map[string]string
		want []string
	}{
		{map[string]string{"LANG": "zh_TW.UTF-8"}, []string{"zh_TW.UTF-8"}},
		{map[string]string{"LANG": "en_US.UTF-8", "LC_ALL": "ja_JP.UTF-8"}, []string{"ja_JP.UTF-8"}},
		{map[string]string{"LANGUAGE": "zh_TW:en", "LANG": "en_US.UTF-8"}, []string{"zh_TW", "en", "en_US.UTF-8"}},
		{map[string]string{"LANGUAGE": "zh_TW", "LANG": "C"}, nil},
		{map[string]string{}, nil},
	} {
		if got := envLanguages(env(c.env)); !slices.Equal(got, c.want) {
			t.Errorf("envLanguages(%v) = %v, want %v", c.env, got, c.want)
		}
	}
}

func TestParseAppleLanguages(t *testing.T) {
	got := parseAppleLanguages("(\n    \"zh-Hant-TW\",\n    \"en-US\",\n    ja\n)\n")
	if want := []string{"zh-Hant-TW", "en-US", "ja"}; !slices.Equal(got, want) {
		t.Errorf("parseAppleLanguages = %v, want %v", got, want)
	}
}
