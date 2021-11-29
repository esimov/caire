// SPDX-License-Identifier: Unlicense OR MIT

package log

/*
#cgo LDFLAGS: -llog

#include <stdlib.h>
#include <android/log.h>
*/
import "C"

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// 1024 is the truncation limit from android/log.h, plus a \n.
const logLineLimit = 1024

var logTag = C.CString(appID)

func init() {
	// Android's logcat already includes timestamps.
	log.SetFlags(log.Flags() &^ log.LstdFlags)
	log.SetOutput(new(androidLogWriter))

	// Redirect stdout and stderr to the Android logger.
	logFd(os.Stdout.Fd())
	logFd(os.Stderr.Fd())
}

type androidLogWriter struct {
	// buf has room for the maximum log line, plus a terminating '\0'.
	buf [logLineLimit + 1]byte
}

func (w *androidLogWriter) Write(data []byte) (int, error) {
	n := 0
	for len(data) > 0 {
		msg := data
		// Truncate the buffer, leaving space for the '\0'.
		if max := len(w.buf) - 1; len(msg) > max {
			msg = msg[:max]
		}
		buf := w.buf[:len(msg)+1]
		copy(buf, msg)
		// Terminating '\0'.
		buf[len(msg)] = 0
		C.__android_log_write(C.ANDROID_LOG_INFO, logTag, (*C.char)(unsafe.Pointer(&buf[0])))
		n += len(msg)
		data = data[len(msg):]
	}
	return n, nil
}

func logFd(fd uintptr) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	if err := syscall.Dup3(int(w.Fd()), int(fd), syscall.O_CLOEXEC); err != nil {
		panic(err)
	}
	go func() {
		lineBuf := bufio.NewReaderSize(r, logLineLimit)
		// The buffer to pass to C, including the terminating '\0'.
		buf := make([]byte, lineBuf.Size()+1)
		cbuf := (*C.char)(unsafe.Pointer(&buf[0]))
		for {
			line, _, err := lineBuf.ReadLine()
			if err != nil {
				break
			}
			copy(buf, line)
			buf[len(line)] = 0
			C.__android_log_write(C.ANDROID_LOG_INFO, logTag, cbuf)
		}
		// The garbage collector doesn't know that w's fd was dup'ed.
		// Avoid finalizing w, and thereby avoid its finalizer closing its fd.
		runtime.KeepAlive(w)
	}()
}
