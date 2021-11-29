// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"errors"
	"fmt"
	"image"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unsafe"

	syscall "golang.org/x/sys/windows"

	"gioui.org/app/internal/windows"
	"gioui.org/unit"
	gowindows "golang.org/x/sys/windows"

	"gioui.org/f32"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
)

type ViewEvent struct {
	HWND uintptr
}

type winDeltas struct {
	width  int32
	height int32
}

type window struct {
	hwnd        syscall.Handle
	hdc         syscall.Handle
	w           *callbacks
	stage       system.Stage
	pointerBtns pointer.Buttons

	// cursorIn tracks whether the cursor was inside the window according
	// to the most recent WM_SETCURSOR.
	cursorIn bool
	cursor   syscall.Handle

	// placement saves the previous window position when in full screen mode.
	placement *windows.WindowPlacement

	animating bool

	deltas winDeltas
	config Config
}

const _WM_WAKEUP = windows.WM_USER + iota

type gpuAPI struct {
	priority    int
	initializer func(w *window) (context, error)
}

// drivers is the list of potential Context implementations.
var drivers []gpuAPI

// winMap maps win32 HWNDs to *windows.
var winMap sync.Map

// iconID is the ID of the icon in the resource file.
const iconID = 1

var resources struct {
	once sync.Once
	// handle is the module handle from GetModuleHandle.
	handle syscall.Handle
	// class is the Gio window class from RegisterClassEx.
	class uint16
	// cursor is the arrow cursor resource.
	cursor syscall.Handle
}

func osMain() {
	select {}
}

func newWindow(window *callbacks, options []Option) error {
	cerr := make(chan error)
	go func() {
		// GetMessage and PeekMessage can filter on a window HWND, but
		// then thread-specific messages such as WM_QUIT are ignored.
		// Instead lock the thread so window messages arrive through
		// unfiltered GetMessage calls.
		runtime.LockOSThread()
		w, err := createNativeWindow()
		if err != nil {
			cerr <- err
			return
		}
		cerr <- nil
		winMap.Store(w.hwnd, w)
		defer winMap.Delete(w.hwnd)
		w.w = window
		w.w.SetDriver(w)
		w.w.Event(ViewEvent{HWND: uintptr(w.hwnd)})
		w.Configure(options)
		windows.ShowWindow(w.hwnd, windows.SW_SHOWDEFAULT)
		windows.SetForegroundWindow(w.hwnd)
		windows.SetFocus(w.hwnd)
		// Since the window class for the cursor is null,
		// set it here to show the cursor.
		w.SetCursor(pointer.CursorDefault)
		if err := w.loop(); err != nil {
			panic(err)
		}
	}()
	return <-cerr
}

// initResources initializes the resources global.
func initResources() error {
	windows.SetProcessDPIAware()
	hInst, err := windows.GetModuleHandle()
	if err != nil {
		return err
	}
	resources.handle = hInst
	c, err := windows.LoadCursor(windows.IDC_ARROW)
	if err != nil {
		return err
	}
	resources.cursor = c
	icon, _ := windows.LoadImage(hInst, iconID, windows.IMAGE_ICON, 0, 0, windows.LR_DEFAULTSIZE|windows.LR_SHARED)
	wcls := windows.WndClassEx{
		CbSize:        uint32(unsafe.Sizeof(windows.WndClassEx{})),
		Style:         windows.CS_HREDRAW | windows.CS_VREDRAW | windows.CS_OWNDC,
		LpfnWndProc:   syscall.NewCallback(windowProc),
		HInstance:     hInst,
		HIcon:         icon,
		LpszClassName: syscall.StringToUTF16Ptr("GioWindow"),
	}
	cls, err := windows.RegisterClassEx(&wcls)
	if err != nil {
		return err
	}
	resources.class = cls
	return nil
}

func createNativeWindow() (*window, error) {
	var resErr error
	resources.once.Do(func() {
		resErr = initResources()
	})
	if resErr != nil {
		return nil, resErr
	}
	dwStyle := uint32(windows.WS_OVERLAPPEDWINDOW)
	dwExStyle := uint32(windows.WS_EX_APPWINDOW | windows.WS_EX_WINDOWEDGE)

	hwnd, err := windows.CreateWindowEx(dwExStyle,
		resources.class,
		"",
		dwStyle|windows.WS_CLIPSIBLINGS|windows.WS_CLIPCHILDREN,
		windows.CW_USEDEFAULT, windows.CW_USEDEFAULT,
		windows.CW_USEDEFAULT, windows.CW_USEDEFAULT,
		0,
		0,
		resources.handle,
		0)
	if err != nil {
		return nil, err
	}
	w := &window{
		hwnd: hwnd,
	}
	w.hdc, err = windows.GetDC(hwnd)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func windowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	win, exists := winMap.Load(hwnd)
	if !exists {
		return windows.DefWindowProc(hwnd, msg, wParam, lParam)
	}

	w := win.(*window)

	switch msg {
	case windows.WM_UNICHAR:
		if wParam == windows.UNICODE_NOCHAR {
			// Tell the system that we accept WM_UNICHAR messages.
			return windows.TRUE
		}
		fallthrough
	case windows.WM_CHAR:
		if r := rune(wParam); unicode.IsPrint(r) {
			w.w.Event(key.EditEvent{Text: string(r)})
		}
		// The message is processed.
		return windows.TRUE
	case windows.WM_DPICHANGED:
		// Let Windows know we're prepared for runtime DPI changes.
		return windows.TRUE
	case windows.WM_ERASEBKGND:
		// Avoid flickering between GPU content and background color.
		return windows.TRUE
	case windows.WM_KEYDOWN, windows.WM_KEYUP, windows.WM_SYSKEYDOWN, windows.WM_SYSKEYUP:
		if n, ok := convertKeyCode(wParam); ok {
			e := key.Event{
				Name:      n,
				Modifiers: getModifiers(),
				State:     key.Press,
			}
			if msg == windows.WM_KEYUP || msg == windows.WM_SYSKEYUP {
				e.State = key.Release
			}

			w.w.Event(e)

			if (wParam == windows.VK_F10) && (msg == windows.WM_SYSKEYDOWN || msg == windows.WM_SYSKEYUP) {
				// Reserve F10 for ourselves, and don't let it open the system menu. Other Windows programs
				// such as cmd.exe and graphical debuggers also reserve F10.
				return 0
			}
		}
	case windows.WM_LBUTTONDOWN:
		w.pointerButton(pointer.ButtonPrimary, true, lParam, getModifiers())
	case windows.WM_LBUTTONUP:
		w.pointerButton(pointer.ButtonPrimary, false, lParam, getModifiers())
	case windows.WM_RBUTTONDOWN:
		w.pointerButton(pointer.ButtonSecondary, true, lParam, getModifiers())
	case windows.WM_RBUTTONUP:
		w.pointerButton(pointer.ButtonSecondary, false, lParam, getModifiers())
	case windows.WM_MBUTTONDOWN:
		w.pointerButton(pointer.ButtonTertiary, true, lParam, getModifiers())
	case windows.WM_MBUTTONUP:
		w.pointerButton(pointer.ButtonTertiary, false, lParam, getModifiers())
	case windows.WM_CANCELMODE:
		w.w.Event(pointer.Event{
			Type: pointer.Cancel,
		})
	case windows.WM_SETFOCUS:
		w.w.Event(key.FocusEvent{Focus: true})
	case windows.WM_KILLFOCUS:
		w.w.Event(key.FocusEvent{Focus: false})
	case windows.WM_MOUSEMOVE:
		x, y := coordsFromlParam(lParam)
		p := f32.Point{X: float32(x), Y: float32(y)}
		w.w.Event(pointer.Event{
			Type:     pointer.Move,
			Source:   pointer.Mouse,
			Position: p,
			Buttons:  w.pointerBtns,
			Time:     windows.GetMessageTime(),
		})
	case windows.WM_MOUSEWHEEL:
		w.scrollEvent(wParam, lParam, false)
	case windows.WM_MOUSEHWHEEL:
		w.scrollEvent(wParam, lParam, true)
	case windows.WM_DESTROY:
		w.w.Event(ViewEvent{})
		w.w.Event(system.DestroyEvent{})
		if w.hdc != 0 {
			windows.ReleaseDC(w.hdc)
			w.hdc = 0
		}
		// The system destroys the HWND for us.
		w.hwnd = 0
		windows.PostQuitMessage(0)
	case windows.WM_PAINT:
		w.draw(true)
	case windows.WM_SIZE:
		switch wParam {
		case windows.SIZE_MINIMIZED:
			w.setStage(system.StagePaused)
		case windows.SIZE_MAXIMIZED, windows.SIZE_RESTORED:
			w.setStage(system.StageRunning)
		}
	case windows.WM_GETMINMAXINFO:
		mm := (*windows.MinMaxInfo)(unsafe.Pointer(uintptr(lParam)))
		if p := w.config.MinSize; p.X > 0 || p.Y > 0 {
			mm.PtMinTrackSize = windows.Point{
				X: int32(p.X) + w.deltas.width,
				Y: int32(p.Y) + w.deltas.height,
			}
		}
		if p := w.config.MaxSize; p.X > 0 || p.Y > 0 {
			mm.PtMaxTrackSize = windows.Point{
				X: int32(p.X) + w.deltas.width,
				Y: int32(p.Y) + w.deltas.height,
			}
		}
	case windows.WM_SETCURSOR:
		w.cursorIn = (lParam & 0xffff) == windows.HTCLIENT
		if w.cursorIn {
			windows.SetCursor(w.cursor)
			return windows.TRUE
		}
	case _WM_WAKEUP:
		w.w.Event(wakeupEvent{})
	}

	return windows.DefWindowProc(hwnd, msg, wParam, lParam)
}

func getModifiers() key.Modifiers {
	var kmods key.Modifiers
	if windows.GetKeyState(windows.VK_LWIN)&0x1000 != 0 || windows.GetKeyState(windows.VK_RWIN)&0x1000 != 0 {
		kmods |= key.ModSuper
	}
	if windows.GetKeyState(windows.VK_MENU)&0x1000 != 0 {
		kmods |= key.ModAlt
	}
	if windows.GetKeyState(windows.VK_CONTROL)&0x1000 != 0 {
		kmods |= key.ModCtrl
	}
	if windows.GetKeyState(windows.VK_SHIFT)&0x1000 != 0 {
		kmods |= key.ModShift
	}
	return kmods
}

func (w *window) pointerButton(btn pointer.Buttons, press bool, lParam uintptr, kmods key.Modifiers) {
	var typ pointer.Type
	if press {
		typ = pointer.Press
		if w.pointerBtns == 0 {
			windows.SetCapture(w.hwnd)
		}
		w.pointerBtns |= btn
	} else {
		typ = pointer.Release
		w.pointerBtns &^= btn
		if w.pointerBtns == 0 {
			windows.ReleaseCapture()
		}
	}
	x, y := coordsFromlParam(lParam)
	p := f32.Point{X: float32(x), Y: float32(y)}
	w.w.Event(pointer.Event{
		Type:      typ,
		Source:    pointer.Mouse,
		Position:  p,
		Buttons:   w.pointerBtns,
		Time:      windows.GetMessageTime(),
		Modifiers: kmods,
	})
}

func coordsFromlParam(lParam uintptr) (int, int) {
	x := int(int16(lParam & 0xffff))
	y := int(int16((lParam >> 16) & 0xffff))
	return x, y
}

func (w *window) scrollEvent(wParam, lParam uintptr, horizontal bool) {
	x, y := coordsFromlParam(lParam)
	// The WM_MOUSEWHEEL coordinates are in screen coordinates, in contrast
	// to other mouse events.
	np := windows.Point{X: int32(x), Y: int32(y)}
	windows.ScreenToClient(w.hwnd, &np)
	p := f32.Point{X: float32(np.X), Y: float32(np.Y)}
	dist := float32(int16(wParam >> 16))
	var sp f32.Point
	if horizontal {
		sp.X = dist
	} else {
		sp.Y = -dist
	}
	w.w.Event(pointer.Event{
		Type:     pointer.Scroll,
		Source:   pointer.Mouse,
		Position: p,
		Buttons:  w.pointerBtns,
		Scroll:   sp,
		Time:     windows.GetMessageTime(),
	})
}

// Adapted from https://blogs.msdn.microsoft.com/oldnewthing/20060126-00/?p=32513/
func (w *window) loop() error {
	msg := new(windows.Msg)
loop:
	for {
		anim := w.animating
		if anim && !windows.PeekMessage(msg, 0, 0, 0, windows.PM_NOREMOVE) {
			w.draw(false)
			continue
		}
		switch ret := windows.GetMessage(msg, 0, 0, 0); ret {
		case -1:
			return errors.New("GetMessage failed")
		case 0:
			// WM_QUIT received.
			break loop
		}
		windows.TranslateMessage(msg)
		windows.DispatchMessage(msg)
	}
	return nil
}

func (w *window) SetAnimating(anim bool) {
	w.animating = anim
}

func (w *window) Wakeup() {
	if err := windows.PostMessage(w.hwnd, _WM_WAKEUP, 0, 0); err != nil {
		panic(err)
	}
}

func (w *window) setStage(s system.Stage) {
	if s != w.stage {
		w.stage = s
		w.w.Event(system.StageEvent{Stage: s})
	}
}

func (w *window) draw(sync bool) {
	var r windows.Rect
	windows.GetClientRect(w.hwnd, &r)
	size := image.Point{
		X: int(r.Right - r.Left),
		Y: int(r.Bottom - r.Top),
	}
	if size != w.config.Size {
		w.config.Size = size
		w.w.Event(ConfigEvent{Config: w.config})
	}
	if w.config.Size.X == 0 || w.config.Size.Y == 0 {
		return
	}
	dpi := windows.GetWindowDPI(w.hwnd)
	cfg := configForDPI(dpi)
	w.w.Event(frameEvent{
		FrameEvent: system.FrameEvent{
			Now:    time.Now(),
			Size:   w.config.Size,
			Metric: cfg,
		},
		Sync: sync,
	})
}

func (w *window) NewContext() (context, error) {
	sort.Slice(drivers, func(i, j int) bool {
		return drivers[i].priority < drivers[j].priority
	})
	var errs []string
	for _, b := range drivers {
		ctx, err := b.initializer(w)
		if err == nil {
			return ctx, nil
		}
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("NewContext: failed to create a GPU device, tried: %s", strings.Join(errs, ", "))
	}
	return nil, errors.New("NewContext: no available GPU drivers")
}

func (w *window) ReadClipboard() {
	w.readClipboard()
}

func (w *window) readClipboard() error {
	if err := windows.OpenClipboard(w.hwnd); err != nil {
		return err
	}
	defer windows.CloseClipboard()
	mem, err := windows.GetClipboardData(windows.CF_UNICODETEXT)
	if err != nil {
		return err
	}
	ptr, err := windows.GlobalLock(mem)
	if err != nil {
		return err
	}
	defer windows.GlobalUnlock(mem)
	content := gowindows.UTF16PtrToString((*uint16)(unsafe.Pointer(ptr)))
	w.w.Event(clipboard.Event{Text: content})
	return nil
}

func (w *window) Configure(options []Option) {
	dpi := windows.GetSystemDPI()
	cfg := configForDPI(dpi)
	prev := w.config
	cnf := w.config
	cnf.apply(cfg, options)

	if cnf.Mode != Fullscreen && prev.Size != cnf.Size {
		width := int32(cnf.Size.X)
		height := int32(cnf.Size.Y)
		w.config.Size = cnf.Size

		// Include the window decorations.
		wr := windows.Rect{
			Right:  width,
			Bottom: height,
		}
		dwStyle := uint32(windows.WS_OVERLAPPEDWINDOW)
		dwExStyle := uint32(windows.WS_EX_APPWINDOW | windows.WS_EX_WINDOWEDGE)
		windows.AdjustWindowRectEx(&wr, dwStyle, 0, dwExStyle)

		dw, dh := width, height
		width = wr.Right - wr.Left
		height = wr.Bottom - wr.Top
		w.deltas.width = width - dw
		w.deltas.height = height - dh

		windows.MoveWindow(w.hwnd, 0, 0, width, height, true)
	}
	if prev.MinSize != cnf.MinSize {
		w.config.MinSize = cnf.MinSize
	}
	if prev.MaxSize != cnf.MaxSize {
		w.config.MaxSize = cnf.MaxSize
	}
	if prev.Title != cnf.Title {
		w.config.Title = cnf.Title
		windows.SetWindowText(w.hwnd, cnf.Title)
	}
	if prev.Mode != cnf.Mode {
		w.SetWindowMode(cnf.Mode)
	}
	if w.config != prev {
		w.w.Event(ConfigEvent{Config: w.config})
	}
}

// Maximize the window. It will have no effect when in fullscreen mode.
func (w *window) Maximize() {
	if w.config.Mode == Fullscreen {
		return
	}
	style := windows.GetWindowLong(w.hwnd)
	windows.SetWindowLong(w.hwnd, windows.GWL_STYLE, style|windows.WS_OVERLAPPEDWINDOW|windows.WS_MAXIMIZE)
	mi := windows.GetMonitorInfo(w.hwnd)
	windows.SetWindowPos(w.hwnd, 0,
		mi.Monitor.Left, mi.Monitor.Top,
		mi.Monitor.Right-mi.Monitor.Left,
		mi.Monitor.Bottom-mi.Monitor.Top,
		windows.SWP_NOOWNERZORDER|windows.SWP_FRAMECHANGED,
	)
}

// Center will place window at monitor center.
func (w *window) Center() {
	// Make sure that the window is sizeable
	style := windows.GetWindowLong(w.hwnd) & (^uintptr(windows.WS_MAXIMIZE))
	windows.SetWindowLong(w.hwnd, windows.GWL_STYLE, style|windows.WS_OVERLAPPEDWINDOW)

	// Find with/height including the window decorations.
	wr := windows.Rect{
		Right:  int32(w.config.Size.X),
		Bottom: int32(w.config.Size.Y),
	}
	dwStyle := uint32(windows.WS_OVERLAPPEDWINDOW)
	dwExStyle := uint32(windows.WS_EX_APPWINDOW | windows.WS_EX_WINDOWEDGE)
	windows.AdjustWindowRectEx(&wr, dwStyle, 0, dwExStyle)
	width := wr.Right - wr.Left
	height := wr.Bottom - wr.Top

	// Move to center of current monitor
	mi := windows.GetMonitorInfo(w.hwnd).Monitor
	x := mi.Left + (mi.Right-mi.Left-width)/2
	y := mi.Top + (mi.Bottom-mi.Top-height)/2
	windows.MoveWindow(w.hwnd, x, y, width, height, true)
}

func (w *window) SetWindowMode(mode WindowMode) {
	// https://devblogs.microsoft.com/oldnewthing/20100412-00/?p=14353
	switch mode {
	case Windowed:
		if w.placement == nil {
			return
		}
		w.config.Mode = Windowed
		windows.SetWindowPlacement(w.hwnd, w.placement)
		w.placement = nil
		style := windows.GetWindowLong(w.hwnd)
		windows.SetWindowLong(w.hwnd, windows.GWL_STYLE, style|windows.WS_OVERLAPPEDWINDOW)
		windows.SetWindowPos(w.hwnd, windows.HWND_TOPMOST,
			0, 0, 0, 0,
			windows.SWP_NOOWNERZORDER|windows.SWP_FRAMECHANGED,
		)
	case Fullscreen:
		if w.placement != nil {
			return
		}
		w.config.Mode = Fullscreen
		w.placement = windows.GetWindowPlacement(w.hwnd)
		style := windows.GetWindowLong(w.hwnd)
		windows.SetWindowLong(w.hwnd, windows.GWL_STYLE, style&^windows.WS_OVERLAPPEDWINDOW)
		mi := windows.GetMonitorInfo(w.hwnd)
		windows.SetWindowPos(w.hwnd, 0,
			mi.Monitor.Left, mi.Monitor.Top,
			mi.Monitor.Right-mi.Monitor.Left,
			mi.Monitor.Bottom-mi.Monitor.Top,
			windows.SWP_NOOWNERZORDER|windows.SWP_FRAMECHANGED,
		)
	}
}

func (w *window) WriteClipboard(s string) {
	w.writeClipboard(s)
}

func (w *window) writeClipboard(s string) error {
	if err := windows.OpenClipboard(w.hwnd); err != nil {
		return err
	}
	defer windows.CloseClipboard()
	if err := windows.EmptyClipboard(); err != nil {
		return err
	}
	u16, err := gowindows.UTF16FromString(s)
	if err != nil {
		return err
	}
	n := len(u16) * int(unsafe.Sizeof(u16[0]))
	mem, err := windows.GlobalAlloc(n)
	if err != nil {
		return err
	}
	ptr, err := windows.GlobalLock(mem)
	if err != nil {
		windows.GlobalFree(mem)
		return err
	}
	var u16v []uint16
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&u16v))
	hdr.Data = ptr
	hdr.Cap = len(u16)
	hdr.Len = len(u16)
	copy(u16v, u16)
	windows.GlobalUnlock(mem)
	if err := windows.SetClipboardData(windows.CF_UNICODETEXT, mem); err != nil {
		windows.GlobalFree(mem)
		return err
	}
	return nil
}

func (w *window) SetCursor(name pointer.CursorName) {
	c, err := loadCursor(name)
	if err != nil {
		c = resources.cursor
	}
	w.cursor = c
	if w.cursorIn {
		windows.SetCursor(w.cursor)
	}
}

func loadCursor(name pointer.CursorName) (syscall.Handle, error) {
	var curID uint16
	switch name {
	default:
		fallthrough
	case pointer.CursorDefault:
		return resources.cursor, nil
	case pointer.CursorText:
		curID = windows.IDC_IBEAM
	case pointer.CursorPointer:
		curID = windows.IDC_HAND
	case pointer.CursorCrossHair:
		curID = windows.IDC_CROSS
	case pointer.CursorColResize:
		curID = windows.IDC_SIZEWE
	case pointer.CursorRowResize:
		curID = windows.IDC_SIZENS
	case pointer.CursorGrab:
		curID = windows.IDC_SIZEALL
	case pointer.CursorNone:
		return 0, nil
	}
	return windows.LoadCursor(curID)
}

func (w *window) ShowTextInput(show bool) {}

func (w *window) SetInputHint(_ key.InputHint) {}

func (w *window) HDC() syscall.Handle {
	return w.hdc
}

func (w *window) HWND() (syscall.Handle, int, int) {
	return w.hwnd, w.config.Size.X, w.config.Size.Y
}

func (w *window) Close() {
	windows.PostMessage(w.hwnd, windows.WM_CLOSE, 0, 0)
}

func (w *window) Raise() {
	windows.SetForegroundWindow(w.hwnd)
	windows.SetWindowPos(w.hwnd, windows.HWND_TOPMOST, 0, 0, 0, 0,
		windows.SWP_NOMOVE|windows.SWP_NOSIZE|windows.SWP_SHOWWINDOW)
}

func convertKeyCode(code uintptr) (string, bool) {
	if '0' <= code && code <= '9' || 'A' <= code && code <= 'Z' {
		return string(rune(code)), true
	}
	var r string
	switch code {
	case windows.VK_ESCAPE:
		r = key.NameEscape
	case windows.VK_LEFT:
		r = key.NameLeftArrow
	case windows.VK_RIGHT:
		r = key.NameRightArrow
	case windows.VK_RETURN:
		r = key.NameReturn
	case windows.VK_UP:
		r = key.NameUpArrow
	case windows.VK_DOWN:
		r = key.NameDownArrow
	case windows.VK_HOME:
		r = key.NameHome
	case windows.VK_END:
		r = key.NameEnd
	case windows.VK_BACK:
		r = key.NameDeleteBackward
	case windows.VK_DELETE:
		r = key.NameDeleteForward
	case windows.VK_PRIOR:
		r = key.NamePageUp
	case windows.VK_NEXT:
		r = key.NamePageDown
	case windows.VK_F1:
		r = "F1"
	case windows.VK_F2:
		r = "F2"
	case windows.VK_F3:
		r = "F3"
	case windows.VK_F4:
		r = "F4"
	case windows.VK_F5:
		r = "F5"
	case windows.VK_F6:
		r = "F6"
	case windows.VK_F7:
		r = "F7"
	case windows.VK_F8:
		r = "F8"
	case windows.VK_F9:
		r = "F9"
	case windows.VK_F10:
		r = "F10"
	case windows.VK_F11:
		r = "F11"
	case windows.VK_F12:
		r = "F12"
	case windows.VK_TAB:
		r = key.NameTab
	case windows.VK_SPACE:
		r = key.NameSpace
	case windows.VK_OEM_1:
		r = ";"
	case windows.VK_OEM_PLUS:
		r = "+"
	case windows.VK_OEM_COMMA:
		r = ","
	case windows.VK_OEM_MINUS:
		r = "-"
	case windows.VK_OEM_PERIOD:
		r = "."
	case windows.VK_OEM_2:
		r = "/"
	case windows.VK_OEM_3:
		r = "`"
	case windows.VK_OEM_4:
		r = "["
	case windows.VK_OEM_5, windows.VK_OEM_102:
		r = "\\"
	case windows.VK_OEM_6:
		r = "]"
	case windows.VK_OEM_7:
		r = "'"
	default:
		return "", false
	}
	return r, true
}

func configForDPI(dpi int) unit.Metric {
	const inchPrDp = 1.0 / 96.0
	ppdp := float32(dpi) * inchPrDp
	return unit.Metric{
		PxPerDp: ppdp,
		PxPerSp: ppdp,
	}
}

func (_ ViewEvent) ImplementsEvent() {}
