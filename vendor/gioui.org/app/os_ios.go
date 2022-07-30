// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && ios
// +build darwin,ios

package app

/*
#cgo CFLAGS: -DGLES_SILENCE_DEPRECATION -Werror -Wno-deprecated-declarations -fmodules -fobjc-arc -x objective-c

#include <CoreGraphics/CoreGraphics.h>
#include <UIKit/UIKit.h>
#include <stdint.h>

struct drawParams {
	CGFloat dpi, sdpi;
	CGFloat width, height;
	CGFloat top, right, bottom, left;
};

static void writeClipboard(unichar *chars, NSUInteger length) {
	@autoreleasepool {
		NSString *s = [NSString string];
		if (length > 0) {
			s = [NSString stringWithCharacters:chars length:length];
		}
		UIPasteboard *p = UIPasteboard.generalPasteboard;
		p.string = s;
	}
}

static CFTypeRef readClipboard(void) {
	@autoreleasepool {
		UIPasteboard *p = UIPasteboard.generalPasteboard;
		return (__bridge_retained CFTypeRef)p.string;
	}
}

static void showTextInput(CFTypeRef viewRef) {
	UIView *view = (__bridge UIView *)viewRef;
	[view becomeFirstResponder];
}

static void hideTextInput(CFTypeRef viewRef) {
	UIView *view = (__bridge UIView *)viewRef;
	[view resignFirstResponder];
}

static struct drawParams viewDrawParams(CFTypeRef viewRef) {
	UIView *v = (__bridge UIView *)viewRef;
	struct drawParams params;
	CGFloat scale = v.layer.contentsScale;
	// Use 163 as the standard ppi on iOS.
	params.dpi = 163*scale;
	params.sdpi = params.dpi;
	UIEdgeInsets insets = v.layoutMargins;
	if (@available(iOS 11.0, tvOS 11.0, *)) {
		UIFontMetrics *metrics = [UIFontMetrics defaultMetrics];
		params.sdpi = [metrics scaledValueForValue:params.sdpi];
		insets = v.safeAreaInsets;
	}
	params.width = v.bounds.size.width*scale;
	params.height = v.bounds.size.height*scale;
	params.top = insets.top*scale;
	params.right = insets.right*scale;
	params.bottom = insets.bottom*scale;
	params.left = insets.left*scale;
	return params;
}
*/
import "C"

import (
	"image"
	"runtime"
	"runtime/debug"
	"time"
	"unicode/utf16"
	"unsafe"

	"gioui.org/f32"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/unit"
)

type ViewEvent struct {
	// ViewController is a CFTypeRef for the UIViewController backing a Window.
	ViewController uintptr
}

type window struct {
	view        C.CFTypeRef
	w           *callbacks
	displayLink *displayLink

	visible bool
	cursor  pointer.Cursor
	config  Config

	pointerMap []C.CFTypeRef
}

var mainWindow = newWindowRendezvous()

var views = make(map[C.CFTypeRef]*window)

func init() {
	// Darwin requires UI operations happen on the main thread only.
	runtime.LockOSThread()
}

//export onCreate
func onCreate(view, controller C.CFTypeRef) {
	w := &window{
		view: view,
	}
	dl, err := NewDisplayLink(func() {
		w.draw(false)
	})
	if err != nil {
		panic(err)
	}
	w.displayLink = dl
	wopts := <-mainWindow.out
	w.w = wopts.window
	w.w.SetDriver(w)
	views[view] = w
	w.Configure(wopts.options)
	w.w.Event(system.StageEvent{Stage: system.StagePaused})
	w.w.Event(ViewEvent{ViewController: uintptr(controller)})
}

//export gio_onDraw
func gio_onDraw(view C.CFTypeRef) {
	w := views[view]
	w.draw(true)
}

func (w *window) draw(sync bool) {
	params := C.viewDrawParams(w.view)
	if params.width == 0 || params.height == 0 {
		return
	}
	wasVisible := w.visible
	w.visible = true
	if !wasVisible {
		w.w.Event(system.StageEvent{Stage: system.StageRunning})
	}
	const inchPrDp = 1.0 / 163
	m := unit.Metric{
		PxPerDp: float32(params.dpi) * inchPrDp,
		PxPerSp: float32(params.sdpi) * inchPrDp,
	}
	dppp := unit.Dp(1. / m.PxPerDp)
	w.w.Event(frameEvent{
		FrameEvent: system.FrameEvent{
			Now: time.Now(),
			Size: image.Point{
				X: int(params.width + .5),
				Y: int(params.height + .5),
			},
			Insets: system.Insets{
				Top:    unit.Dp(params.top) * dppp,
				Bottom: unit.Dp(params.bottom) * dppp,
				Left:   unit.Dp(params.left) * dppp,
				Right:  unit.Dp(params.right) * dppp,
			},
			Metric: m,
		},
		Sync: sync,
	})
}

//export onStop
func onStop(view C.CFTypeRef) {
	w := views[view]
	w.visible = false
	w.w.Event(system.StageEvent{Stage: system.StagePaused})
}

//export onDestroy
func onDestroy(view C.CFTypeRef) {
	w := views[view]
	delete(views, view)
	w.w.Event(ViewEvent{})
	w.w.Event(system.DestroyEvent{})
	w.displayLink.Close()
	w.view = 0
}

//export onFocus
func onFocus(view C.CFTypeRef, focus int) {
	w := views[view]
	w.w.Event(key.FocusEvent{Focus: focus != 0})
}

//export onLowMemory
func onLowMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}

//export onUpArrow
func onUpArrow(view C.CFTypeRef) {
	views[view].onKeyCommand(key.NameUpArrow)
}

//export onDownArrow
func onDownArrow(view C.CFTypeRef) {
	views[view].onKeyCommand(key.NameDownArrow)
}

//export onLeftArrow
func onLeftArrow(view C.CFTypeRef) {
	views[view].onKeyCommand(key.NameLeftArrow)
}

//export onRightArrow
func onRightArrow(view C.CFTypeRef) {
	views[view].onKeyCommand(key.NameRightArrow)
}

//export onDeleteBackward
func onDeleteBackward(view C.CFTypeRef) {
	views[view].onKeyCommand(key.NameDeleteBackward)
}

//export onText
func onText(view, str C.CFTypeRef) {
	w := views[view]
	w.w.EditorInsert(nsstringToString(str))
}

//export onTouch
func onTouch(last C.int, view, touchRef C.CFTypeRef, phase C.NSInteger, x, y C.CGFloat, ti C.double) {
	var typ pointer.Type
	switch phase {
	case C.UITouchPhaseBegan:
		typ = pointer.Press
	case C.UITouchPhaseMoved:
		typ = pointer.Move
	case C.UITouchPhaseEnded:
		typ = pointer.Release
	case C.UITouchPhaseCancelled:
		typ = pointer.Cancel
	default:
		return
	}
	w := views[view]
	t := time.Duration(float64(ti) * float64(time.Second))
	p := f32.Point{X: float32(x), Y: float32(y)}
	w.w.Event(pointer.Event{
		Type:      typ,
		Source:    pointer.Touch,
		PointerID: w.lookupTouch(last != 0, touchRef),
		Position:  p,
		Time:      t,
	})
}

func (w *window) ReadClipboard() {
	cstr := C.readClipboard()
	defer C.CFRelease(cstr)
	content := nsstringToString(cstr)
	w.w.Event(clipboard.Event{Text: content})
}

func (w *window) WriteClipboard(s string) {
	u16 := utf16.Encode([]rune(s))
	var chars *C.unichar
	if len(u16) > 0 {
		chars = (*C.unichar)(unsafe.Pointer(&u16[0]))
	}
	C.writeClipboard(chars, C.NSUInteger(len(u16)))
}

func (w *window) Configure([]Option) {
	// Decorations are never disabled.
	w.config.Decorated = true
	w.w.Event(ConfigEvent{Config: w.config})
}

func (w *window) EditorStateChanged(old, new editorState) {}

func (w *window) Perform(system.Action) {}

func (w *window) SetAnimating(anim bool) {
	v := w.view
	if v == 0 {
		return
	}
	if anim {
		w.displayLink.Start()
	} else {
		w.displayLink.Stop()
	}
}

func (w *window) SetCursor(cursor pointer.Cursor) {
	w.cursor = windowSetCursor(w.cursor, cursor)
}

func (w *window) onKeyCommand(name string) {
	w.w.Event(key.Event{
		Name: name,
	})
}

// lookupTouch maps an UITouch pointer value to an index. If
// last is set, the map is cleared.
func (w *window) lookupTouch(last bool, touch C.CFTypeRef) pointer.ID {
	id := -1
	for i, ref := range w.pointerMap {
		if ref == touch {
			id = i
			break
		}
	}
	if id == -1 {
		id = len(w.pointerMap)
		w.pointerMap = append(w.pointerMap, touch)
	}
	if last {
		w.pointerMap = w.pointerMap[:0]
	}
	return pointer.ID(id)
}

func (w *window) contextView() C.CFTypeRef {
	return w.view
}

func (w *window) ShowTextInput(show bool) {
	if show {
		C.showTextInput(w.view)
	} else {
		C.hideTextInput(w.view)
	}
}

func (w *window) SetInputHint(_ key.InputHint) {}

func newWindow(win *callbacks, options []Option) error {
	mainWindow.in <- windowAndConfig{win, options}
	return <-mainWindow.errs
}

func osMain() {
}

//export gio_runMain
func gio_runMain() {
	runMain()
}

func (_ ViewEvent) ImplementsEvent() {}
