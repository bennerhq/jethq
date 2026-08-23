//go:build darwin

package nativeui

/*
#cgo LDFLAGS: -framework Cocoa
#include <stddef.h>
void jetkvmSetApplicationIcon(const unsigned char *data, size_t length);
*/
import "C"

import "unsafe"

// SetApplicationIcon sets the icon that macOS displays for this app in the Dock.
func SetApplicationIcon(data []byte) {
	if len(data) == 0 {
		return
	}
	C.jetkvmSetApplicationIcon((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
}
