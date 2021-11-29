// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios
// +build darwin,!ios

package app

import (
	"errors"
	"image"
	"runtime"
	"time"
	"unicode"
	"unicode/utf16"
	"unsafe"

	"gioui.org/f32"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/unit"

	_ "gioui.org/internal/cocoainit"
)

/*
#cgo CFLAGS: -DGL_SILENCE_DEPRECATION -Werror -Wno-deprecated-declarations -fmodules -fobjc-arc -x objective-c

#include <AppKit/AppKit.h>

#define MOUSE_MOVE 1
#define MOUSE_UP 2
#define MOUSE_DOWN 3
#define MOUSE_SCROLL 4

__attribute__ ((visibility ("hidden"))) void gio_main(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createView(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createWindow(CFTypeRef viewRef, const char *title, CGFloat width, CGFloat height, CGFloat minWidth, CGFloat minHeight, CGFloat maxWidth, CGFloat maxHeight);

static void writeClipboard(unichar *chars, NSUInteger length) {
	@autoreleasepool {
		NSString *s = [NSString string];
		if (length > 0) {
			s = [NSString stringWithCharacters:chars length:length];
		}
		NSPasteboard *p = NSPasteboard.generalPasteboard;
		[p declareTypes:@[NSPasteboardTypeString] owner:nil];
		[p setString:s forType:NSPasteboardTypeString];
	}
}

static CFTypeRef readClipboard(void) {
	@autoreleasepool {
		NSPasteboard *p = NSPasteboard.generalPasteboard;
		NSString *content = [p stringForType:NSPasteboardTypeString];
		return (__bridge_retained CFTypeRef)content;
	}
}

static CGFloat viewHeight(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return [view bounds].size.height;
}

static CGFloat viewWidth(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return [view bounds].size.width;
}

static CGFloat getScreenBackingScale(void) {
	return [NSScreen.mainScreen backingScaleFactor];
}

static CGFloat getViewBackingScale(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return [view.window backingScaleFactor];
}

static void setNeedsDisplay(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	[view setNeedsDisplay:YES];
}

static NSPoint cascadeTopLeftFromPoint(CFTypeRef windowRef, NSPoint topLeft) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	return [window cascadeTopLeftFromPoint:topLeft];
}

static void makeKeyAndOrderFront(CFTypeRef windowRef) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	[window makeKeyAndOrderFront:nil];
}

static void toggleFullScreen(CFTypeRef windowRef) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	[window toggleFullScreen:nil];
}

static void closeWindow(CFTypeRef windowRef) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	[window performClose:nil];
}

static void setSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	NSSize size = NSMakeSize(width, height);
	[window setContentSize:size];
}

static void setMinSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	window.contentMinSize = NSMakeSize(width, height);
}

static void setMaxSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	window.contentMaxSize = NSMakeSize(width, height);
}

static void setTitle(CFTypeRef windowRef, const char *title) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	window.title = [NSString stringWithUTF8String: title];
}

static CFTypeRef layerForView(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return (__bridge CFTypeRef)view.layer;
}

static void raiseWindow(CFTypeRef windowRef) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	[window makeKeyAndOrderFront:nil];
}

*/
import "C"

func init() {
	// Darwin requires that UI operations happen on the main thread only.
	runtime.LockOSThread()
}

// ViewEvent notified the client of changes to the window AppKit handles.
// The handles are retained until another ViewEvent is sent.
type ViewEvent struct {
	// View is a CFTypeRef for the NSView for the window.
	View uintptr
	// Layer is a CFTypeRef of the CALayer of View.
	Layer uintptr
}

type window struct {
	view        C.CFTypeRef
	window      C.CFTypeRef
	w           *callbacks
	stage       system.Stage
	displayLink *displayLink
	cursor      pointer.CursorName

	scale  float32
	config Config
}

// viewMap is the mapping from Cocoa NSViews to Go windows.
var viewMap = make(map[C.CFTypeRef]*window)

// launched is closed when applicationDidFinishLaunching is called.
var launched = make(chan struct{})

// nextTopLeft is the offset to use for the next window's call to
// cascadeTopLeftFromPoint.
var nextTopLeft C.NSPoint

// mustView is like lookupView, except that it panics
// if the view isn't mapped.
func mustView(view C.CFTypeRef) *window {
	w, ok := lookupView(view)
	if !ok {
		panic("no window for view")
	}
	return w
}

func lookupView(view C.CFTypeRef) (*window, bool) {
	w, exists := viewMap[view]
	if !exists {
		return nil, false
	}
	return w, true
}

func deleteView(view C.CFTypeRef) {
	delete(viewMap, view)
}

func insertView(view C.CFTypeRef, w *window) {
	viewMap[view] = w
}

func (w *window) contextView() C.CFTypeRef {
	return w.view
}

func (w *window) ReadClipboard() {
	content := nsstringToString(C.readClipboard())
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

func (w *window) Configure(options []Option) {
	screenScale := float32(C.getScreenBackingScale())
	cfg := configFor(screenScale)
	prev := w.config
	cnf := w.config
	cnf.apply(cfg, options)
	cnf.Size = cnf.Size.Div(int(screenScale))
	cnf.MinSize = cnf.MinSize.Div(int(screenScale))
	cnf.MaxSize = cnf.MaxSize.Div(int(screenScale))

	if cnf.Mode != Fullscreen && prev.Size != cnf.Size {
		w.config.Size = cnf.Size
		C.setSize(w.window, C.CGFloat(cnf.Size.X), C.CGFloat(cnf.Size.Y))
	}
	if prev.MinSize != cnf.MinSize {
		w.config.MinSize = cnf.MinSize
		C.setMinSize(w.window, C.CGFloat(cnf.MinSize.X), C.CGFloat(cnf.MinSize.Y))
	}
	if prev.MaxSize != cnf.MaxSize {
		w.config.MaxSize = cnf.MaxSize
		C.setMaxSize(w.window, C.CGFloat(cnf.MaxSize.X), C.CGFloat(cnf.MaxSize.Y))
	}

	if prev.Title != cnf.Title {
		w.config.Title = cnf.Title
		title := C.CString(cnf.Title)
		defer C.free(unsafe.Pointer(title))
		C.setTitle(w.window, title)
	}
	if prev.Mode != cnf.Mode {
		switch cnf.Mode {
		case Windowed, Fullscreen:
			w.config.Mode = cnf.Mode
			C.toggleFullScreen(w.window)
		}
	}
	if w.config != prev {
		w.w.Event(ConfigEvent{Config: w.config})
	}
}

func (w *window) SetCursor(name pointer.CursorName) {
	w.cursor = windowSetCursor(w.cursor, name)
}

func (w *window) ShowTextInput(show bool) {}

func (w *window) SetInputHint(_ key.InputHint) {}

func (w *window) SetAnimating(anim bool) {
	if anim {
		w.displayLink.Start()
	} else {
		w.displayLink.Stop()
	}
}

func (w *window) Raise() {
	C.raiseWindow(w.window)
}

func (w *window) runOnMain(f func()) {
	runOnMain(func() {
		// Make sure the view is still valid. The window might've been closed
		// during the switch to the main thread.
		if w.view != 0 {
			f()
		}
	})
}

func (w *window) Close() {
	C.closeWindow(w.window)
}

// Maximize the window. Not implemented for macos.
func (w *window) Maximize() {}

// Center the window. Not implemented for macos.
func (w *window) Center() {}

func (w *window) setStage(stage system.Stage) {
	if stage == w.stage {
		return
	}
	w.stage = stage
	w.w.Event(system.StageEvent{Stage: stage})
}

//export gio_onKeys
func gio_onKeys(view C.CFTypeRef, cstr *C.char, ti C.double, mods C.NSUInteger, keyDown C.bool) {
	str := C.GoString(cstr)
	kmods := convertMods(mods)
	ks := key.Release
	if keyDown {
		ks = key.Press
	}
	w := mustView(view)
	for _, k := range str {
		if n, ok := convertKey(k); ok {
			w.w.Event(key.Event{
				Name:      n,
				Modifiers: kmods,
				State:     ks,
			})
		}
	}
}

//export gio_onText
func gio_onText(view C.CFTypeRef, cstr *C.char) {
	str := C.GoString(cstr)
	w := mustView(view)
	w.w.Event(key.EditEvent{Text: str})
}

//export gio_onMouse
func gio_onMouse(view C.CFTypeRef, cdir C.int, cbtns C.NSUInteger, x, y, dx, dy C.CGFloat, ti C.double, mods C.NSUInteger) {
	var typ pointer.Type
	switch cdir {
	case C.MOUSE_MOVE:
		typ = pointer.Move
	case C.MOUSE_UP:
		typ = pointer.Release
	case C.MOUSE_DOWN:
		typ = pointer.Press
	case C.MOUSE_SCROLL:
		typ = pointer.Scroll
	default:
		panic("invalid direction")
	}
	var btns pointer.Buttons
	if cbtns&(1<<0) != 0 {
		btns |= pointer.ButtonPrimary
	}
	if cbtns&(1<<1) != 0 {
		btns |= pointer.ButtonSecondary
	}
	if cbtns&(1<<2) != 0 {
		btns |= pointer.ButtonTertiary
	}
	t := time.Duration(float64(ti)*float64(time.Second) + .5)
	w := mustView(view)
	xf, yf := float32(x)*w.scale, float32(y)*w.scale
	dxf, dyf := float32(dx)*w.scale, float32(dy)*w.scale
	w.w.Event(pointer.Event{
		Type:      typ,
		Source:    pointer.Mouse,
		Time:      t,
		Buttons:   btns,
		Position:  f32.Point{X: xf, Y: yf},
		Scroll:    f32.Point{X: dxf, Y: dyf},
		Modifiers: convertMods(mods),
	})
}

//export gio_onDraw
func gio_onDraw(view C.CFTypeRef) {
	w := mustView(view)
	w.draw()
}

//export gio_onFocus
func gio_onFocus(view C.CFTypeRef, focus C.int) {
	w := mustView(view)
	w.w.Event(key.FocusEvent{Focus: focus == 1})
	w.SetCursor(w.cursor)
}

//export gio_onChangeScreen
func gio_onChangeScreen(view C.CFTypeRef, did uint64) {
	w := mustView(view)
	w.displayLink.SetDisplayID(did)
}

func (w *window) draw() {
	w.scale = float32(C.getViewBackingScale(w.view))
	wf, hf := float32(C.viewWidth(w.view)), float32(C.viewHeight(w.view))
	sz := image.Point{
		X: int(wf*w.scale + .5),
		Y: int(hf*w.scale + .5),
	}
	if sz != w.config.Size {
		w.config.Size = sz
		w.w.Event(ConfigEvent{Config: w.config})
	}
	if sz.X == 0 || sz.Y == 0 {
		return
	}
	cfg := configFor(w.scale)
	w.setStage(system.StageRunning)
	w.w.Event(frameEvent{
		FrameEvent: system.FrameEvent{
			Now:    time.Now(),
			Size:   w.config.Size,
			Metric: cfg,
		},
		Sync: true,
	})
}

func configFor(scale float32) unit.Metric {
	return unit.Metric{
		PxPerDp: scale,
		PxPerSp: scale,
	}
}

//export gio_onClose
func gio_onClose(view C.CFTypeRef) {
	w := mustView(view)
	w.w.Event(ViewEvent{})
	deleteView(view)
	w.w.Event(system.DestroyEvent{})
	w.displayLink.Close()
	C.CFRelease(w.view)
	C.CFRelease(w.window)
	w.view = 0
	w.window = 0
	w.displayLink = nil
}

//export gio_onHide
func gio_onHide(view C.CFTypeRef) {
	w := mustView(view)
	w.setStage(system.StagePaused)
}

//export gio_onShow
func gio_onShow(view C.CFTypeRef) {
	w := mustView(view)
	w.setStage(system.StageRunning)
}

//export gio_onAppHide
func gio_onAppHide() {
	for _, w := range viewMap {
		w.setStage(system.StagePaused)
	}
}

//export gio_onAppShow
func gio_onAppShow() {
	for _, w := range viewMap {
		w.setStage(system.StageRunning)
	}
}

//export gio_onFinishLaunching
func gio_onFinishLaunching() {
	close(launched)
}

func newWindow(win *callbacks, options []Option) error {
	<-launched
	errch := make(chan error)
	runOnMain(func() {
		w, err := newOSWindow()
		if err != nil {
			errch <- err
			return
		}
		errch <- nil
		w.w = win
		w.window = C.gio_createWindow(w.view, nil, 0, 0, 0, 0, 0, 0)
		win.SetDriver(w)
		w.Configure(options)
		if nextTopLeft.x == 0 && nextTopLeft.y == 0 {
			// cascadeTopLeftFromPoint treats (0, 0) as a no-op,
			// and just returns the offset we need for the first window.
			nextTopLeft = C.cascadeTopLeftFromPoint(w.window, nextTopLeft)
		}
		nextTopLeft = C.cascadeTopLeftFromPoint(w.window, nextTopLeft)
		C.makeKeyAndOrderFront(w.window)
		layer := C.layerForView(w.view)
		w.w.Event(ViewEvent{View: uintptr(w.view), Layer: uintptr(layer)})
	})
	return <-errch
}

func newOSWindow() (*window, error) {
	view := C.gio_createView()
	if view == 0 {
		return nil, errors.New("CreateWindow: failed to create view")
	}
	scale := float32(C.getViewBackingScale(view))
	w := &window{
		view:  view,
		scale: scale,
	}
	dl, err := NewDisplayLink(func() {
		w.runOnMain(func() {
			C.setNeedsDisplay(w.view)
		})
	})
	w.displayLink = dl
	if err != nil {
		C.CFRelease(view)
		return nil, err
	}
	insertView(view, w)
	return w, nil
}

func osMain() {
	C.gio_main()
}

func convertKey(k rune) (string, bool) {
	var n string
	switch k {
	case 0x1b:
		n = key.NameEscape
	case C.NSLeftArrowFunctionKey:
		n = key.NameLeftArrow
	case C.NSRightArrowFunctionKey:
		n = key.NameRightArrow
	case C.NSUpArrowFunctionKey:
		n = key.NameUpArrow
	case C.NSDownArrowFunctionKey:
		n = key.NameDownArrow
	case 0xd:
		n = key.NameReturn
	case 0x3:
		n = key.NameEnter
	case C.NSHomeFunctionKey:
		n = key.NameHome
	case C.NSEndFunctionKey:
		n = key.NameEnd
	case 0x7f:
		n = key.NameDeleteBackward
	case C.NSDeleteFunctionKey:
		n = key.NameDeleteForward
	case C.NSPageUpFunctionKey:
		n = key.NamePageUp
	case C.NSPageDownFunctionKey:
		n = key.NamePageDown
	case C.NSF1FunctionKey:
		n = "F1"
	case C.NSF2FunctionKey:
		n = "F2"
	case C.NSF3FunctionKey:
		n = "F3"
	case C.NSF4FunctionKey:
		n = "F4"
	case C.NSF5FunctionKey:
		n = "F5"
	case C.NSF6FunctionKey:
		n = "F6"
	case C.NSF7FunctionKey:
		n = "F7"
	case C.NSF8FunctionKey:
		n = "F8"
	case C.NSF9FunctionKey:
		n = "F9"
	case C.NSF10FunctionKey:
		n = "F10"
	case C.NSF11FunctionKey:
		n = "F11"
	case C.NSF12FunctionKey:
		n = "F12"
	case 0x09, 0x19:
		n = key.NameTab
	case 0x20:
		n = key.NameSpace
	default:
		k = unicode.ToUpper(k)
		if !unicode.IsPrint(k) {
			return "", false
		}
		n = string(k)
	}
	return n, true
}

func convertMods(mods C.NSUInteger) key.Modifiers {
	var kmods key.Modifiers
	if mods&C.NSAlternateKeyMask != 0 {
		kmods |= key.ModAlt
	}
	if mods&C.NSControlKeyMask != 0 {
		kmods |= key.ModCtrl
	}
	if mods&C.NSCommandKeyMask != 0 {
		kmods |= key.ModCommand
	}
	if mods&C.NSShiftKeyMask != 0 {
		kmods |= key.ModShift
	}
	return kmods
}

func (_ ViewEvent) ImplementsEvent() {}
