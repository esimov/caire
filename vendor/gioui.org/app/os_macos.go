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
	"unicode/utf8"

	"gioui.org/internal/f32"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/unit"

	_ "gioui.org/internal/cocoainit"
)

/*
#cgo CFLAGS: -Werror -Wno-deprecated-declarations -fobjc-arc -x objective-c
#cgo LDFLAGS: -framework AppKit -framework QuartzCore

#include <AppKit/AppKit.h>

#define MOUSE_MOVE 1
#define MOUSE_UP 2
#define MOUSE_DOWN 3
#define MOUSE_SCROLL 4

__attribute__ ((visibility ("hidden"))) void gio_main(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createView(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createWindow(CFTypeRef viewRef, CGFloat width, CGFloat height, CGFloat minWidth, CGFloat minHeight, CGFloat maxWidth, CGFloat maxHeight);

static void writeClipboard(CFTypeRef str) {
	@autoreleasepool {
		NSString *s = (__bridge NSString *)str;
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

static NSWindowStyleMask getWindowStyleMask(CFTypeRef windowRef) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	return [window styleMask];
}

static void setWindowStyleMask(CFTypeRef windowRef, NSWindowStyleMask mask) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	window.styleMask = mask;
}

static void setWindowTitleVisibility(CFTypeRef windowRef, NSWindowTitleVisibility state) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	window.titleVisibility = state;
}

static void setWindowTitlebarAppearsTransparent(CFTypeRef windowRef, int transparent) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	window.titlebarAppearsTransparent = (BOOL)transparent;
}

static void setWindowStandardButtonHidden(CFTypeRef windowRef, NSWindowButton btn, int hide) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	[window standardWindowButton:btn].hidden = (BOOL)hide;
}

static void performWindowDragWithEvent(CFTypeRef windowRef, CFTypeRef evt) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	[window performWindowDragWithEvent:(__bridge NSEvent*)evt];
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

static void setScreenFrame(CFTypeRef windowRef, CGFloat x, CGFloat y, CGFloat w, CGFloat h) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	NSRect r = NSMakeRect(x, y, w, h);
	[window setFrame:r display:YES];
}

static void hideWindow(CFTypeRef windowRef) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	[window miniaturize:window];
}

static void unhideWindow(CFTypeRef windowRef) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	[window deminiaturize:window];
}

static NSRect getScreenFrame(CFTypeRef windowRef) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	return [[window screen] frame];
}

static void setTitle(CFTypeRef windowRef, CFTypeRef titleRef) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	window.title = (__bridge NSString *)titleRef;
}

static int isWindowZoomed(CFTypeRef windowRef) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	return window.zoomed ? 1 : 0;
}

static void zoomWindow(CFTypeRef windowRef) {
	NSWindow *window = (__bridge NSWindow *)windowRef;
	[window zoom:nil];
}

static CFTypeRef layerForView(CFTypeRef viewRef) {
	NSView *view = (__bridge NSView *)viewRef;
	return (__bridge CFTypeRef)view.layer;
}

static void raiseWindow(CFTypeRef windowRef) {
	NSWindow* window = (__bridge NSWindow *)windowRef;
	[window makeKeyAndOrderFront:nil];
}

static CFTypeRef createInputContext(CFTypeRef clientRef) {
	@autoreleasepool {
		id<NSTextInputClient> client = (__bridge id<NSTextInputClient>)clientRef;
		NSTextInputContext *ctx = [[NSTextInputContext alloc] initWithClient:client];
		return CFBridgingRetain(ctx);
	}
}

static void discardMarkedText(CFTypeRef viewRef) {
    @autoreleasepool {
		id<NSTextInputClient> view = (__bridge id<NSTextInputClient>)viewRef;
		NSTextInputContext *ctx = [NSTextInputContext currentInputContext];
		if (view == [ctx client]) {
			[ctx discardMarkedText];
		}
    }
}

static void invalidateCharacterCoordinates(CFTypeRef viewRef) {
    @autoreleasepool {
		id<NSTextInputClient> view = (__bridge id<NSTextInputClient>)viewRef;
		NSTextInputContext *ctx = [NSTextInputContext currentInputContext];
		if (view == [ctx client]) {
			[ctx invalidateCharacterCoordinates];
		}
    }
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
	// redraw is a single entry channel for making sure only one
	// display link redraw request is in flight.
	redraw      chan struct{}
	cursor      pointer.Cursor
	pointerBtns pointer.Buttons

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
	cstr := C.readClipboard()
	defer C.CFRelease(cstr)
	content := nsstringToString(cstr)
	w.w.Event(clipboard.Event{Text: content})
}

func (w *window) WriteClipboard(s string) {
	cstr := stringToNSString(s)
	defer C.CFRelease(cstr)
	C.writeClipboard(cstr)
}

func (w *window) updateWindowMode() {
	style := int(C.getWindowStyleMask(w.window))
	if style&C.NSWindowStyleMaskFullScreen != 0 {
		w.config.Mode = Fullscreen
	} else {
		w.config.Mode = Windowed
	}
	w.config.Decorated = style&C.NSWindowStyleMaskFullSizeContentView == 0
}

func (w *window) Configure(options []Option) {
	screenScale := float32(C.getScreenBackingScale())
	cfg := configFor(screenScale)
	prev := w.config
	w.updateWindowMode()
	cnf := w.config
	cnf.apply(cfg, options)

	switch cnf.Mode {
	case Fullscreen:
		switch prev.Mode {
		case Fullscreen:
		case Minimized:
			C.unhideWindow(w.window)
			fallthrough
		default:
			w.config.Mode = Fullscreen
			C.toggleFullScreen(w.window)
		}
	case Minimized:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Minimized
			C.hideWindow(w.window)
		}
	case Maximized:
		switch prev.Mode {
		case Fullscreen:
		case Minimized:
			C.unhideWindow(w.window)
			fallthrough
		default:
			w.config.Mode = Maximized
			w.setTitle(prev, cnf)
			if C.isWindowZoomed(w.window) == 0 {
				C.zoomWindow(w.window)
			}
		}
	case Windowed:
		switch prev.Mode {
		case Fullscreen:
			C.toggleFullScreen(w.window)
		case Minimized:
			C.unhideWindow(w.window)
		case Maximized:
		}
		w.config.Mode = Windowed
		if C.isWindowZoomed(w.window) != 0 {
			C.zoomWindow(w.window)
		}
		w.setTitle(prev, cnf)
		if prev.Size != cnf.Size {
			w.config.Size = cnf.Size
			cnf.Size = cnf.Size.Div(int(screenScale))
			C.setSize(w.window, C.CGFloat(cnf.Size.X), C.CGFloat(cnf.Size.Y))
		}
		if prev.MinSize != cnf.MinSize {
			w.config.MinSize = cnf.MinSize
			cnf.MinSize = cnf.MinSize.Div(int(screenScale))
			C.setMinSize(w.window, C.CGFloat(cnf.MinSize.X), C.CGFloat(cnf.MinSize.Y))
		}
		if prev.MaxSize != cnf.MaxSize {
			w.config.MaxSize = cnf.MaxSize
			cnf.MaxSize = cnf.MaxSize.Div(int(screenScale))
			C.setMaxSize(w.window, C.CGFloat(cnf.MaxSize.X), C.CGFloat(cnf.MaxSize.Y))
		}
	}
	if cnf.Decorated != prev.Decorated {
		w.config.Decorated = cnf.Decorated
		mask := C.getWindowStyleMask(w.window)
		style := C.NSWindowStyleMask(C.NSWindowStyleMaskTitled | C.NSWindowStyleMaskResizable | C.NSWindowStyleMaskMiniaturizable | C.NSWindowStyleMaskClosable)
		style = C.NSWindowStyleMaskFullSizeContentView
		mask &^= style
		barTrans := C.int(C.NO)
		titleVis := C.NSWindowTitleVisibility(C.NSWindowTitleVisible)
		if !cnf.Decorated {
			mask |= style
			barTrans = C.YES
			titleVis = C.NSWindowTitleHidden
		}
		C.setWindowTitlebarAppearsTransparent(w.window, barTrans)
		C.setWindowTitleVisibility(w.window, titleVis)
		C.setWindowStyleMask(w.window, mask)
		C.setWindowStandardButtonHidden(w.window, C.NSWindowCloseButton, barTrans)
		C.setWindowStandardButtonHidden(w.window, C.NSWindowMiniaturizeButton, barTrans)
		C.setWindowStandardButtonHidden(w.window, C.NSWindowZoomButton, barTrans)
	}
	w.w.Event(ConfigEvent{Config: w.config})
}

func (w *window) setTitle(prev, cnf Config) {
	if prev.Title != cnf.Title {
		w.config.Title = cnf.Title
		title := stringToNSString(cnf.Title)
		defer C.CFRelease(title)
		C.setTitle(w.window, title)
	}
}

func (w *window) Perform(acts system.Action) {
	walkActions(acts, func(a system.Action) {
		switch a {
		case system.ActionCenter:
			r := C.getScreenFrame(w.window) // the screen size of the window
			sz := w.config.Size
			x := (int(r.size.width) - sz.X) / 2
			y := (int(r.size.height) - sz.Y) / 2
			C.setScreenFrame(w.window, C.CGFloat(x), C.CGFloat(y), C.CGFloat(sz.X), C.CGFloat(sz.Y))
		case system.ActionRaise:
			C.raiseWindow(w.window)
		}
	})
	if acts&system.ActionClose != 0 {
		C.closeWindow(w.window)
	}
}

func (w *window) SetCursor(cursor pointer.Cursor) {
	w.cursor = windowSetCursor(w.cursor, cursor)
}

func (w *window) EditorStateChanged(old, new editorState) {
	if old.Selection.Range != new.Selection.Range || old.Snippet != new.Snippet {
		C.discardMarkedText(w.view)
		w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
	}
	if old.Selection.Caret != new.Selection.Caret || old.Selection.Transform != new.Selection.Transform {
		C.invalidateCharacterCoordinates(w.view)
	}
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

func (w *window) runOnMain(f func()) {
	runOnMain(func() {
		// Make sure the view is still valid. The window might've been closed
		// during the switch to the main thread.
		if w.view != 0 {
			f()
		}
	})
}

func (w *window) setStage(stage system.Stage) {
	if stage == w.stage {
		return
	}
	w.stage = stage
	w.w.Event(system.StageEvent{Stage: stage})
}

//export gio_onKeys
func gio_onKeys(view, cstr C.CFTypeRef, ti C.double, mods C.NSUInteger, keyDown C.bool) {
	str := nsstringToString(cstr)
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
func gio_onText(view, cstr C.CFTypeRef) {
	str := nsstringToString(cstr)
	w := mustView(view)
	w.w.EditorInsert(str)
}

//export gio_onMouse
func gio_onMouse(view, evt C.CFTypeRef, cdir C.int, cbtn C.NSInteger, x, y, dx, dy C.CGFloat, ti C.double, mods C.NSUInteger) {
	w := mustView(view)
	t := time.Duration(float64(ti)*float64(time.Second) + .5)
	xf, yf := float32(x)*w.scale, float32(y)*w.scale
	dxf, dyf := float32(dx)*w.scale, float32(dy)*w.scale
	pos := f32.Point{X: xf, Y: yf}
	var btn pointer.Buttons
	switch cbtn {
	case 0:
		btn = pointer.ButtonPrimary
	case 1:
		btn = pointer.ButtonSecondary
	}
	var typ pointer.Type
	switch cdir {
	case C.MOUSE_MOVE:
		typ = pointer.Move
	case C.MOUSE_UP:
		typ = pointer.Release
		w.pointerBtns &^= btn
	case C.MOUSE_DOWN:
		typ = pointer.Press
		w.pointerBtns |= btn
		act, ok := w.w.ActionAt(pos)
		if ok && w.config.Mode != Fullscreen {
			switch act {
			case system.ActionMove:
				C.performWindowDragWithEvent(w.window, evt)
				return
			}
		}
	case C.MOUSE_SCROLL:
		typ = pointer.Scroll
	default:
		panic("invalid direction")
	}
	w.w.Event(pointer.Event{
		Type:      typ,
		Source:    pointer.Mouse,
		Time:      t,
		Buttons:   w.pointerBtns,
		Position:  pos,
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
	C.setNeedsDisplay(w.view)
}

//export gio_hasMarkedText
func gio_hasMarkedText(view C.CFTypeRef) C.int {
	w := mustView(view)
	state := w.w.EditorState()
	if state.compose.Start != -1 {
		return 1
	}
	return 0
}

//export gio_markedRange
func gio_markedRange(view C.CFTypeRef) C.NSRange {
	w := mustView(view)
	state := w.w.EditorState()
	rng := state.compose
	start, end := rng.Start, rng.End
	if start == -1 {
		return C.NSMakeRange(C.NSNotFound, 0)
	}
	u16start := state.UTF16Index(start)
	return C.NSMakeRange(
		C.NSUInteger(u16start),
		C.NSUInteger(state.UTF16Index(end)-u16start),
	)
}

//export gio_selectedRange
func gio_selectedRange(view C.CFTypeRef) C.NSRange {
	w := mustView(view)
	state := w.w.EditorState()
	rng := state.Selection
	start, end := rng.Start, rng.End
	if start > end {
		start, end = end, start
	}
	u16start := state.UTF16Index(start)
	return C.NSMakeRange(
		C.NSUInteger(u16start),
		C.NSUInteger(state.UTF16Index(end)-u16start),
	)
}

//export gio_unmarkText
func gio_unmarkText(view C.CFTypeRef) {
	w := mustView(view)
	w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
}

//export gio_setMarkedText
func gio_setMarkedText(view, cstr C.CFTypeRef, selRange C.NSRange, replaceRange C.NSRange) {
	w := mustView(view)
	str := nsstringToString(cstr)
	state := w.w.EditorState()
	rng := state.compose
	if rng.Start == -1 {
		rng = state.Selection.Range
	}
	if replaceRange.location != C.NSNotFound {
		// replaceRange is relative to marked (or selected) text.
		offset := state.UTF16Index(rng.Start)
		start := state.RunesIndex(int(replaceRange.location) + offset)
		end := state.RunesIndex(int(replaceRange.location+replaceRange.length) + offset)
		rng = key.Range{
			Start: start,
			End:   end,
		}
	}
	w.w.EditorReplace(rng, str)
	comp := key.Range{
		Start: rng.Start,
		End:   rng.Start + utf8.RuneCountInString(str),
	}
	w.w.SetComposingRegion(comp)

	sel := key.Range{Start: comp.End, End: comp.End}
	if selRange.location != C.NSNotFound {
		// selRange is relative to inserted text.
		offset := state.UTF16Index(rng.Start)
		start := state.RunesIndex(int(selRange.location) + offset)
		end := state.RunesIndex(int(selRange.location+selRange.length) + offset)
		sel = key.Range{
			Start: start,
			End:   end,
		}
	}
	w.w.SetEditorSelection(sel)
}

//export gio_substringForProposedRange
func gio_substringForProposedRange(view C.CFTypeRef, crng C.NSRange, actual C.NSRangePointer) C.CFTypeRef {
	w := mustView(view)
	state := w.w.EditorState()
	start, end := state.Snippet.Start, state.Snippet.End
	if start > end {
		start, end = end, start
	}
	rng := key.Range{
		Start: state.RunesIndex(int(crng.location)),
		End:   state.RunesIndex(int(crng.location + crng.length)),
	}
	if rng.Start < start || end < rng.End {
		w.w.SetEditorSnippet(rng)
	}
	u16start := state.UTF16Index(start)
	actual.location = C.NSUInteger(u16start)
	actual.length = C.NSUInteger(state.UTF16Index(end) - u16start)
	return stringToNSString(state.Snippet.Text)
}

//export gio_insertText
func gio_insertText(view, cstr C.CFTypeRef, crng C.NSRange) {
	w := mustView(view)
	state := w.w.EditorState()
	rng := state.compose
	if rng.Start == -1 {
		rng = state.Selection.Range
	}
	if crng.location != C.NSNotFound {
		rng = key.Range{
			Start: state.RunesIndex(int(crng.location)),
			End:   state.RunesIndex(int(crng.location + crng.length)),
		}
	}
	str := nsstringToString(cstr)
	w.w.EditorReplace(rng, str)
	w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
	start := rng.Start
	if rng.End < start {
		start = rng.End
	}
	pos := start + utf8.RuneCountInString(str)
	w.w.SetEditorSelection(key.Range{Start: pos, End: pos})
}

//export gio_characterIndexForPoint
func gio_characterIndexForPoint(view C.CFTypeRef, p C.NSPoint) C.NSUInteger {
	return C.NSNotFound
}

//export gio_firstRectForCharacterRange
func gio_firstRectForCharacterRange(view C.CFTypeRef, crng C.NSRange, actual C.NSRangePointer) C.NSRect {
	w := mustView(view)
	state := w.w.EditorState()
	sel := state.Selection
	u16start := state.UTF16Index(sel.Start)
	actual.location = C.NSUInteger(u16start)
	actual.length = 0
	// Transform to NSView local coordinates (lower left origin, undo backing scale).
	scale := 1. / float32(C.getViewBackingScale(w.view))
	height := float32(C.viewHeight(w.view))
	local := f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(scale, -scale)).Offset(f32.Pt(0, height))
	t := local.Mul(sel.Transform)
	bounds := f32.Rectangle{
		Min: t.Transform(sel.Pos.Sub(f32.Pt(0, sel.Ascent))),
		Max: t.Transform(sel.Pos.Add(f32.Pt(0, sel.Descent))),
	}.Canon()
	sz := bounds.Size()
	return C.NSMakeRect(
		C.CGFloat(bounds.Min.X), C.CGFloat(bounds.Min.Y),
		C.CGFloat(sz.X), C.CGFloat(sz.Y),
	)
}

func (w *window) draw() {
	select {
	case <-w.redraw:
	default:
	}
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

//export gio_onFullscreen
func gio_onFullscreen(view C.CFTypeRef) {
	w := mustView(view)
	w.config.Mode = Fullscreen
	w.w.Event(ConfigEvent{Config: w.config})
}

//export gio_onWindowed
func gio_onWindowed(view C.CFTypeRef) {
	w := mustView(view)
	w.config.Mode = Windowed
	w.w.Event(ConfigEvent{Config: w.config})
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
		w.window = C.gio_createWindow(w.view, 0, 0, 0, 0, 0, 0)
		w.updateWindowMode()
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
		return nil, errors.New("newOSWindows: failed to create view")
	}
	scale := float32(C.getViewBackingScale(view))
	w := &window{
		view:   view,
		scale:  scale,
		redraw: make(chan struct{}, 1),
	}
	dl, err := NewDisplayLink(func() {
		select {
		case w.redraw <- struct{}{}:
		default:
			return
		}
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
		n = key.NameF1
	case C.NSF2FunctionKey:
		n = key.NameF2
	case C.NSF3FunctionKey:
		n = key.NameF3
	case C.NSF4FunctionKey:
		n = key.NameF4
	case C.NSF5FunctionKey:
		n = key.NameF5
	case C.NSF6FunctionKey:
		n = key.NameF6
	case C.NSF7FunctionKey:
		n = key.NameF7
	case C.NSF8FunctionKey:
		n = key.NameF8
	case C.NSF9FunctionKey:
		n = key.NameF9
	case C.NSF10FunctionKey:
		n = key.NameF10
	case C.NSF11FunctionKey:
		n = key.NameF11
	case C.NSF12FunctionKey:
		n = key.NameF12
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
