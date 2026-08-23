//go:build darwin

package nativeui

/*
#cgo LDFLAGS: -framework Cocoa
void jetkvmInstallSettingsMenu(void);
*/
import "C"

import "sync"

var (
	menuHandlerMu sync.RWMutex
	menuHandler   func(string)
)

// InstallSettingsMenu adds a first-class Settings menu to the macOS menu bar.
func InstallSettingsMenu(handler func(string)) {
	menuHandlerMu.Lock()
	menuHandler = handler
	menuHandlerMu.Unlock()
	C.jetkvmInstallSettingsMenu()
}

//export jetkvmNativeMenuAction
func jetkvmNativeMenuAction(action *C.char) {
	menuHandlerMu.RLock()
	handler := menuHandler
	menuHandlerMu.RUnlock()
	if handler != nil {
		handler(C.GoString(action))
	}
}
