// SPDX-License-Identifier: Unlicense OR MIT

package app

/*
#include <Foundation/Foundation.h>

__attribute__ ((visibility ("hidden"))) void gio_wakeupMainThread(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createDisplayLink(void);
__attribute__ ((visibility ("hidden"))) void gio_releaseDisplayLink(CFTypeRef dl);
__attribute__ ((visibility ("hidden"))) int gio_startDisplayLink(CFTypeRef dl);
__attribute__ ((visibility ("hidden"))) int gio_stopDisplayLink(CFTypeRef dl);
__attribute__ ((visibility ("hidden"))) void gio_setDisplayLinkDisplay(CFTypeRef dl, uint64_t did);
__attribute__ ((visibility ("hidden"))) void gio_hideCursor();
__attribute__ ((visibility ("hidden"))) void gio_showCursor();
__attribute__ ((visibility ("hidden"))) void gio_setCursor(NSUInteger curID);

static bool isMainThread() {
	return [NSThread isMainThread];
}

static NSUInteger nsstringLength(CFTypeRef cstr) {
	NSString *str = (__bridge NSString *)cstr;
	return [str length];
}

static void nsstringGetCharacters(CFTypeRef cstr, unichar *chars, NSUInteger loc, NSUInteger length) {
	NSString *str = (__bridge NSString *)cstr;
	[str getCharacters:chars range:NSMakeRange(loc, length)];
}
*/
import "C"
import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf16"
	"unsafe"

	"gioui.org/io/pointer"
)

// displayLink is the state for a display link (CVDisplayLinkRef on macOS,
// CADisplayLink on iOS). It runs a state-machine goroutine that keeps the
// display link running for a while after being stopped to avoid the thread
// start/stop overhead and because the CVDisplayLink sometimes fails to
// start, stop and start again within a short duration.
type displayLink struct {
	callback func()
	// states is for starting or stopping the display link.
	states chan bool
	// done is closed when the display link is destroyed.
	done chan struct{}
	// dids receives the display id when the callback owner is moved
	// to a different screen.
	dids chan uint64
	// running tracks the desired state of the link. running is accessed
	// with atomic.
	running uint32
}

// displayLinks maps CFTypeRefs to *displayLinks.
var displayLinks sync.Map

var mainFuncs = make(chan func(), 1)

// runOnMain runs the function on the main thread.
func runOnMain(f func()) {
	if C.isMainThread() {
		f()
		return
	}
	go func() {
		mainFuncs <- f
		C.gio_wakeupMainThread()
	}()
}

//export gio_dispatchMainFuncs
func gio_dispatchMainFuncs() {
	for {
		select {
		case f := <-mainFuncs:
			f()
		default:
			return
		}
	}
}

// nsstringToString converts a NSString to a Go string, and
// releases the original string.
func nsstringToString(str C.CFTypeRef) string {
	if str == 0 {
		return ""
	}
	defer C.CFRelease(str)
	n := C.nsstringLength(str)
	if n == 0 {
		return ""
	}
	chars := make([]uint16, n)
	C.nsstringGetCharacters(str, (*C.unichar)(unsafe.Pointer(&chars[0])), 0, n)
	utf8 := utf16.Decode(chars)
	return string(utf8)
}

func NewDisplayLink(callback func()) (*displayLink, error) {
	d := &displayLink{
		callback: callback,
		done:     make(chan struct{}),
		states:   make(chan bool),
		dids:     make(chan uint64),
	}
	dl := C.gio_createDisplayLink()
	if dl == 0 {
		return nil, errors.New("app: failed to create display link")
	}
	go d.run(dl)
	return d, nil
}

func (d *displayLink) run(dl C.CFTypeRef) {
	defer C.gio_releaseDisplayLink(dl)
	displayLinks.Store(dl, d)
	defer displayLinks.Delete(dl)
	var stopTimer *time.Timer
	var tchan <-chan time.Time
	started := false
	for {
		select {
		case <-tchan:
			tchan = nil
			started = false
			C.gio_stopDisplayLink(dl)
		case start := <-d.states:
			switch {
			case !start && tchan == nil:
				// stopTimeout is the delay before stopping the display link to
				// avoid the overhead of frequently starting and stopping the
				// link thread.
				const stopTimeout = 500 * time.Millisecond
				if stopTimer == nil {
					stopTimer = time.NewTimer(stopTimeout)
				} else {
					// stopTimer is always drained when tchan == nil.
					stopTimer.Reset(stopTimeout)
				}
				tchan = stopTimer.C
				atomic.StoreUint32(&d.running, 0)
			case start:
				if tchan != nil && !stopTimer.Stop() {
					<-tchan
				}
				tchan = nil
				atomic.StoreUint32(&d.running, 1)
				if !started {
					started = true
					C.gio_startDisplayLink(dl)
				}
			}
		case did := <-d.dids:
			C.gio_setDisplayLinkDisplay(dl, C.uint64_t(did))
		case <-d.done:
			return
		}
	}
}

func (d *displayLink) Start() {
	d.states <- true
}

func (d *displayLink) Stop() {
	d.states <- false
}

func (d *displayLink) Close() {
	close(d.done)
}

func (d *displayLink) SetDisplayID(did uint64) {
	d.dids <- did
}

//export gio_onFrameCallback
func gio_onFrameCallback(dl C.CFTypeRef) {
	if d, exists := displayLinks.Load(dl); exists {
		d := d.(*displayLink)
		if atomic.LoadUint32(&d.running) != 0 {
			d.callback()
		}
	}
}

// windowSetCursor updates the cursor from the current one to a new one
// and returns the new one.
func windowSetCursor(from, to pointer.CursorName) pointer.CursorName {
	if from == to {
		return to
	}
	var curID int
	switch to {
	default:
		to = pointer.CursorDefault
		fallthrough
	case pointer.CursorDefault:
		curID = 1
	case pointer.CursorText:
		curID = 2
	case pointer.CursorPointer:
		curID = 3
	case pointer.CursorCrossHair:
		curID = 4
	case pointer.CursorColResize:
		curID = 5
	case pointer.CursorRowResize:
		curID = 6
	case pointer.CursorGrab:
		curID = 7
	case pointer.CursorNone:
		C.gio_hideCursor()
		return to
	}
	if from == pointer.CursorNone {
		C.gio_showCursor()
	}
	C.gio_setCursor(C.NSUInteger(curID))
	return to
}

func (w *window) Wakeup() {
	runOnMain(func() {
		w.w.Event(wakeupEvent{})
	})
}
