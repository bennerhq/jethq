//go:build !darwin

package nativeui

// InstallSettingsMenu is a no-op outside macOS. The in-window control strip
// remains available on every supported platform.
func InstallSettingsMenu(_ func(string)) {}
