//go:build !darwin

package nativeui

// SetApplicationIcon is a no-op outside macOS.
func SetApplicationIcon(_ []byte) {}
