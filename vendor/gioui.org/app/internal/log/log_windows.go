// SPDX-License-Identifier: Unlicense OR MIT

package log

import (
	"log"
	"syscall"
	"unsafe"
)

type logger struct{}

var (
	kernel32           = syscall.NewLazyDLL("kernel32")
	outputDebugStringW = kernel32.NewProc("OutputDebugStringW")
	debugView          *logger
)

func init() {
	// Windows DebugView already includes timestamps.
	if syscall.Stderr == 0 {
		log.SetFlags(log.Flags() &^ log.LstdFlags)
		log.SetOutput(debugView)
	}
}

func (l *logger) Write(buf []byte) (int, error) {
	p, err := syscall.UTF16PtrFromString(string(buf))
	if err != nil {
		return 0, err
	}
	outputDebugStringW.Call(uintptr(unsafe.Pointer(p)))
	return len(buf), nil
}
