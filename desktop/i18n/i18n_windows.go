//go:build windows

package i18n

import "golang.org/x/sys/windows"

// systemLanguages asks Windows for the display languages directly; the
// culture that PowerShell's Get-UICulture reports can differ.
func systemLanguages() []string {
	l, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil {
		return nil
	}
	return l
}
