// SPDX-License-Identifier: Unlicense OR MIT

//go:build windows
// +build windows

package windows

import (
	"fmt"
	"runtime"
	"time"
	"unicode/utf16"
	"unsafe"

	syscall "golang.org/x/sys/windows"
)

type CompositionForm struct {
	dwStyle      uint32
	ptCurrentPos Point
	rcArea       Rect
}

type CandidateForm struct {
	dwIndex      uint32
	dwStyle      uint32
	ptCurrentPos Point
	rcArea       Rect
}

type Rect struct {
	Left, Top, Right, Bottom int32
}

type WndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CnClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type Margins struct {
	CxLeftWidth    int32
	CxRightWidth   int32
	CyTopHeight    int32
	CyBottomHeight int32
}

type Msg struct {
	Hwnd     syscall.Handle
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       Point
	LPrivate uint32
}

type Point struct {
	X, Y int32
}

type MinMaxInfo struct {
	PtReserved     Point
	PtMaxSize      Point
	PtMaxPosition  Point
	PtMinTrackSize Point
	PtMaxTrackSize Point
}

type NCCalcSizeParams struct {
	Rgrc  [3]Rect
	LpPos *WindowPos
}

type WindowPos struct {
	HWND            syscall.Handle
	HWNDInsertAfter syscall.Handle
	x               int32
	y               int32
	cx              int32
	cy              int32
	flags           uint32
}

type WindowPlacement struct {
	length           uint32
	flags            uint32
	showCmd          uint32
	ptMinPosition    Point
	ptMaxPosition    Point
	rcNormalPosition Rect
	rcDevice         Rect
}

type MonitorInfo struct {
	cbSize   uint32
	Monitor  Rect
	WorkArea Rect
	Flags    uint32
}

const (
	TRUE = 1

	CPS_CANCEL = 0x0004

	CS_HREDRAW     = 0x0002
	CS_INSERTCHAR  = 0x2000
	CS_NOMOVECARET = 0x4000
	CS_VREDRAW     = 0x0001
	CS_OWNDC       = 0x0020

	CW_USEDEFAULT = -2147483648

	GWL_STYLE = ^(uintptr(16) - 1) // -16

	GCS_COMPSTR       = 0x0008
	GCS_COMPREADSTR   = 0x0001
	GCS_CURSORPOS     = 0x0080
	GCS_DELTASTART    = 0x0100
	GCS_RESULTREADSTR = 0x0200
	GCS_RESULTSTR     = 0x0800

	CFS_POINT        = 0x0002
	CFS_CANDIDATEPOS = 0x0040

	HWND_TOPMOST = ^(uint32(1) - 1) // -1

	HTCAPTION     = 2
	HTCLIENT      = 1
	HTLEFT        = 10
	HTRIGHT       = 11
	HTTOP         = 12
	HTTOPLEFT     = 13
	HTTOPRIGHT    = 14
	HTBOTTOM      = 15
	HTBOTTOMLEFT  = 16
	HTBOTTOMRIGHT = 17

	IDC_APPSTARTING = 32650 // Standard arrow and small hourglass
	IDC_ARROW       = 32512 // Standard arrow
	IDC_CROSS       = 32515 // Crosshair
	IDC_HAND        = 32649 // Hand
	IDC_HELP        = 32651 // Arrow and question mark
	IDC_IBEAM       = 32513 // I-beam
	IDC_NO          = 32648 // Slashed circle
	IDC_SIZEALL     = 32646 // Four-pointed arrow pointing north, south, east, and west
	IDC_SIZENESW    = 32643 // Double-pointed arrow pointing northeast and southwest
	IDC_SIZENS      = 32645 // Double-pointed arrow pointing north and south
	IDC_SIZENWSE    = 32642 // Double-pointed arrow pointing northwest and southeast
	IDC_SIZEWE      = 32644 // Double-pointed arrow pointing west and east
	IDC_UPARROW     = 32516 // Vertical arrow
	IDC_WAIT        = 32514 // Hour

	INFINITE = 0xFFFFFFFF

	LOGPIXELSX = 88

	MDT_EFFECTIVE_DPI = 0

	MONITOR_DEFAULTTOPRIMARY = 1

	NI_COMPOSITIONSTR = 0x0015

	SIZE_MAXIMIZED = 2
	SIZE_MINIMIZED = 1
	SIZE_RESTORED  = 0

	SCS_SETSTR = GCS_COMPREADSTR | GCS_COMPSTR

	SM_CXSIZEFRAME = 32
	SM_CYSIZEFRAME = 33

	SW_SHOWDEFAULT   = 10
	SW_SHOWMINIMIZED = 2
	SW_SHOWMAXIMIZED = 3
	SW_SHOWNORMAL    = 1
	SW_SHOW          = 5

	SWP_FRAMECHANGED  = 0x0020
	SWP_NOMOVE        = 0x0002
	SWP_NOOWNERZORDER = 0x0200
	SWP_NOSIZE        = 0x0001
	SWP_NOZORDER      = 0x0004
	SWP_SHOWWINDOW    = 0x0040

	USER_TIMER_MINIMUM = 0x0000000A

	VK_CONTROL = 0x11
	VK_LWIN    = 0x5B
	VK_MENU    = 0x12
	VK_RWIN    = 0x5C
	VK_SHIFT   = 0x10

	VK_BACK   = 0x08
	VK_DELETE = 0x2e
	VK_DOWN   = 0x28
	VK_END    = 0x23
	VK_ESCAPE = 0x1b
	VK_HOME   = 0x24
	VK_LEFT   = 0x25
	VK_NEXT   = 0x22
	VK_PRIOR  = 0x21
	VK_RIGHT  = 0x27
	VK_RETURN = 0x0d
	VK_SPACE  = 0x20
	VK_TAB    = 0x09
	VK_UP     = 0x26

	VK_F1  = 0x70
	VK_F2  = 0x71
	VK_F3  = 0x72
	VK_F4  = 0x73
	VK_F5  = 0x74
	VK_F6  = 0x75
	VK_F7  = 0x76
	VK_F8  = 0x77
	VK_F9  = 0x78
	VK_F10 = 0x79
	VK_F11 = 0x7A
	VK_F12 = 0x7B

	VK_OEM_1      = 0xba
	VK_OEM_PLUS   = 0xbb
	VK_OEM_COMMA  = 0xbc
	VK_OEM_MINUS  = 0xbd
	VK_OEM_PERIOD = 0xbe
	VK_OEM_2      = 0xbf
	VK_OEM_3      = 0xc0
	VK_OEM_4      = 0xdb
	VK_OEM_5      = 0xdc
	VK_OEM_6      = 0xdd
	VK_OEM_7      = 0xde
	VK_OEM_102    = 0xe2

	UNICODE_NOCHAR = 65535

	WM_CANCELMODE           = 0x001F
	WM_CHAR                 = 0x0102
	WM_CLOSE                = 0x0010
	WM_CREATE               = 0x0001
	WM_DPICHANGED           = 0x02E0
	WM_DESTROY              = 0x0002
	WM_ERASEBKGND           = 0x0014
	WM_GETMINMAXINFO        = 0x0024
	WM_IME_COMPOSITION      = 0x010F
	WM_IME_ENDCOMPOSITION   = 0x010E
	WM_IME_STARTCOMPOSITION = 0x010D
	WM_KEYDOWN              = 0x0100
	WM_KEYUP                = 0x0101
	WM_KILLFOCUS            = 0x0008
	WM_LBUTTONDOWN          = 0x0201
	WM_LBUTTONUP            = 0x0202
	WM_MBUTTONDOWN          = 0x0207
	WM_MBUTTONUP            = 0x0208
	WM_MOUSEMOVE            = 0x0200
	WM_MOUSEWHEEL           = 0x020A
	WM_MOUSEHWHEEL          = 0x020E
	WM_NCACTIVATE           = 0x0086
	WM_NCHITTEST            = 0x0084
	WM_NCCALCSIZE           = 0x0083
	WM_PAINT                = 0x000F
	WM_QUIT                 = 0x0012
	WM_SETCURSOR            = 0x0020
	WM_SETFOCUS             = 0x0007
	WM_SHOWWINDOW           = 0x0018
	WM_SIZE                 = 0x0005
	WM_SYSKEYDOWN           = 0x0104
	WM_SYSKEYUP             = 0x0105
	WM_RBUTTONDOWN          = 0x0204
	WM_RBUTTONUP            = 0x0205
	WM_TIMER                = 0x0113
	WM_UNICHAR              = 0x0109
	WM_USER                 = 0x0400
	WM_WINDOWPOSCHANGED     = 0x0047

	WS_CLIPCHILDREN     = 0x02000000
	WS_CLIPSIBLINGS     = 0x04000000
	WS_MAXIMIZE         = 0x01000000
	WS_ICONIC           = 0x20000000
	WS_VISIBLE          = 0x10000000
	WS_OVERLAPPED       = 0x00000000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME |
		WS_MINIMIZEBOX | WS_MAXIMIZEBOX
	WS_CAPTION     = 0x00C00000
	WS_SYSMENU     = 0x00080000
	WS_THICKFRAME  = 0x00040000
	WS_MINIMIZEBOX = 0x00020000
	WS_MAXIMIZEBOX = 0x00010000

	WS_EX_APPWINDOW  = 0x00040000
	WS_EX_WINDOWEDGE = 0x00000100

	QS_ALLINPUT = 0x04FF

	MWMO_WAITALL        = 0x0001
	MWMO_INPUTAVAILABLE = 0x0004

	WAIT_OBJECT_0 = 0

	PM_REMOVE   = 0x0001
	PM_NOREMOVE = 0x0000

	GHND = 0x0042

	CF_UNICODETEXT = 13
	IMAGE_BITMAP   = 0
	IMAGE_ICON     = 1
	IMAGE_CURSOR   = 2

	LR_CREATEDIBSECTION = 0x00002000
	LR_DEFAULTCOLOR     = 0x00000000
	LR_DEFAULTSIZE      = 0x00000040
	LR_LOADFROMFILE     = 0x00000010
	LR_LOADMAP3DCOLORS  = 0x00001000
	LR_LOADTRANSPARENT  = 0x00000020
	LR_MONOCHROME       = 0x00000001
	LR_SHARED           = 0x00008000
	LR_VGACOLOR         = 0x00000080
)

var (
	kernel32          = syscall.NewLazySystemDLL("kernel32.dll")
	_GetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	_GlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	_GlobalFree       = kernel32.NewProc("GlobalFree")
	_GlobalLock       = kernel32.NewProc("GlobalLock")
	_GlobalUnlock     = kernel32.NewProc("GlobalUnlock")

	user32                       = syscall.NewLazySystemDLL("user32.dll")
	_AdjustWindowRectEx          = user32.NewProc("AdjustWindowRectEx")
	_CallMsgFilter               = user32.NewProc("CallMsgFilterW")
	_CloseClipboard              = user32.NewProc("CloseClipboard")
	_CreateWindowEx              = user32.NewProc("CreateWindowExW")
	_DefWindowProc               = user32.NewProc("DefWindowProcW")
	_DestroyWindow               = user32.NewProc("DestroyWindow")
	_DispatchMessage             = user32.NewProc("DispatchMessageW")
	_EmptyClipboard              = user32.NewProc("EmptyClipboard")
	_GetWindowRect               = user32.NewProc("GetWindowRect")
	_GetClientRect               = user32.NewProc("GetClientRect")
	_GetClipboardData            = user32.NewProc("GetClipboardData")
	_GetDC                       = user32.NewProc("GetDC")
	_GetDpiForWindow             = user32.NewProc("GetDpiForWindow")
	_GetKeyState                 = user32.NewProc("GetKeyState")
	_GetMessage                  = user32.NewProc("GetMessageW")
	_GetMessageTime              = user32.NewProc("GetMessageTime")
	_GetMonitorInfo              = user32.NewProc("GetMonitorInfoW")
	_GetSystemMetrics            = user32.NewProc("GetSystemMetrics")
	_GetWindowLong               = user32.NewProc("GetWindowLongPtrW")
	_GetWindowLong32             = user32.NewProc("GetWindowLongW")
	_GetWindowPlacement          = user32.NewProc("GetWindowPlacement")
	_KillTimer                   = user32.NewProc("KillTimer")
	_LoadCursor                  = user32.NewProc("LoadCursorW")
	_LoadImage                   = user32.NewProc("LoadImageW")
	_MonitorFromPoint            = user32.NewProc("MonitorFromPoint")
	_MonitorFromWindow           = user32.NewProc("MonitorFromWindow")
	_MoveWindow                  = user32.NewProc("MoveWindow")
	_MsgWaitForMultipleObjectsEx = user32.NewProc("MsgWaitForMultipleObjectsEx")
	_OpenClipboard               = user32.NewProc("OpenClipboard")
	_PeekMessage                 = user32.NewProc("PeekMessageW")
	_PostMessage                 = user32.NewProc("PostMessageW")
	_PostQuitMessage             = user32.NewProc("PostQuitMessage")
	_ReleaseCapture              = user32.NewProc("ReleaseCapture")
	_RegisterClassExW            = user32.NewProc("RegisterClassExW")
	_ReleaseDC                   = user32.NewProc("ReleaseDC")
	_ScreenToClient              = user32.NewProc("ScreenToClient")
	_ShowWindow                  = user32.NewProc("ShowWindow")
	_SetCapture                  = user32.NewProc("SetCapture")
	_SetCursor                   = user32.NewProc("SetCursor")
	_SetClipboardData            = user32.NewProc("SetClipboardData")
	_SetForegroundWindow         = user32.NewProc("SetForegroundWindow")
	_SetFocus                    = user32.NewProc("SetFocus")
	_SetProcessDPIAware          = user32.NewProc("SetProcessDPIAware")
	_SetTimer                    = user32.NewProc("SetTimer")
	_SetWindowLong               = user32.NewProc("SetWindowLongPtrW")
	_SetWindowLong32             = user32.NewProc("SetWindowLongW")
	_SetWindowPlacement          = user32.NewProc("SetWindowPlacement")
	_SetWindowPos                = user32.NewProc("SetWindowPos")
	_SetWindowText               = user32.NewProc("SetWindowTextW")
	_TranslateMessage            = user32.NewProc("TranslateMessage")
	_UnregisterClass             = user32.NewProc("UnregisterClassW")
	_UpdateWindow                = user32.NewProc("UpdateWindow")

	shcore            = syscall.NewLazySystemDLL("shcore")
	_GetDpiForMonitor = shcore.NewProc("GetDpiForMonitor")

	gdi32          = syscall.NewLazySystemDLL("gdi32")
	_GetDeviceCaps = gdi32.NewProc("GetDeviceCaps")

	imm32                    = syscall.NewLazySystemDLL("imm32")
	_ImmGetContext           = imm32.NewProc("ImmGetContext")
	_ImmGetCompositionString = imm32.NewProc("ImmGetCompositionStringW")
	_ImmNotifyIME            = imm32.NewProc("ImmNotifyIME")
	_ImmReleaseContext       = imm32.NewProc("ImmReleaseContext")
	_ImmSetCandidateWindow   = imm32.NewProc("ImmSetCandidateWindow")
	_ImmSetCompositionWindow = imm32.NewProc("ImmSetCompositionWindow")

	dwmapi                        = syscall.NewLazySystemDLL("dwmapi")
	_DwmExtendFrameIntoClientArea = dwmapi.NewProc("DwmExtendFrameIntoClientArea")
)

func AdjustWindowRectEx(r *Rect, dwStyle uint32, bMenu int, dwExStyle uint32) {
	_AdjustWindowRectEx.Call(uintptr(unsafe.Pointer(r)), uintptr(dwStyle), uintptr(bMenu), uintptr(dwExStyle))
}

func CallMsgFilter(m *Msg, nCode uintptr) bool {
	r, _, _ := _CallMsgFilter.Call(uintptr(unsafe.Pointer(m)), nCode)
	return r != 0
}

func CloseClipboard() error {
	r, _, err := _CloseClipboard.Call()
	if r == 0 {
		return fmt.Errorf("CloseClipboard: %v", err)
	}
	return nil
}

func CreateWindowEx(dwExStyle uint32, lpClassName uint16, lpWindowName string, dwStyle uint32, x, y, w, h int32, hWndParent, hMenu, hInstance syscall.Handle, lpParam uintptr) (syscall.Handle, error) {
	wname := syscall.StringToUTF16Ptr(lpWindowName)
	hwnd, _, err := _CreateWindowEx.Call(
		uintptr(dwExStyle),
		uintptr(lpClassName),
		uintptr(unsafe.Pointer(wname)),
		uintptr(dwStyle),
		uintptr(x), uintptr(y),
		uintptr(w), uintptr(h),
		uintptr(hWndParent),
		uintptr(hMenu),
		uintptr(hInstance),
		uintptr(lpParam))
	if hwnd == 0 {
		return 0, fmt.Errorf("CreateWindowEx failed: %v", err)
	}
	return syscall.Handle(hwnd), nil
}

func DefWindowProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	r, _, _ := _DefWindowProc.Call(uintptr(hwnd), uintptr(msg), wparam, lparam)
	return r
}

func DestroyWindow(hwnd syscall.Handle) {
	_DestroyWindow.Call(uintptr(hwnd))
}

func DispatchMessage(m *Msg) {
	_DispatchMessage.Call(uintptr(unsafe.Pointer(m)))
}

func DwmExtendFrameIntoClientArea(hwnd syscall.Handle, margins Margins) error {
	r, _, _ := _DwmExtendFrameIntoClientArea.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&margins)))
	if r != 0 {
		return fmt.Errorf("DwmExtendFrameIntoClientArea: %#x", r)
	}
	return nil
}

func EmptyClipboard() error {
	r, _, err := _EmptyClipboard.Call()
	if r == 0 {
		return fmt.Errorf("EmptyClipboard: %v", err)
	}
	return nil
}

func GetWindowRect(hwnd syscall.Handle) Rect {
	var r Rect
	_GetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	return r
}

func GetClientRect(hwnd syscall.Handle) Rect {
	var r Rect
	_GetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	return r
}

func GetClipboardData(format uint32) (syscall.Handle, error) {
	r, _, err := _GetClipboardData.Call(uintptr(format))
	if r == 0 {
		return 0, fmt.Errorf("GetClipboardData: %v", err)
	}
	return syscall.Handle(r), nil
}

func GetDC(hwnd syscall.Handle) (syscall.Handle, error) {
	hdc, _, err := _GetDC.Call(uintptr(hwnd))
	if hdc == 0 {
		return 0, fmt.Errorf("GetDC failed: %v", err)
	}
	return syscall.Handle(hdc), nil
}

func GetModuleHandle() (syscall.Handle, error) {
	h, _, err := _GetModuleHandleW.Call(uintptr(0))
	if h == 0 {
		return 0, fmt.Errorf("GetModuleHandleW failed: %v", err)
	}
	return syscall.Handle(h), nil
}

func getDeviceCaps(hdc syscall.Handle, index int32) int {
	c, _, _ := _GetDeviceCaps.Call(uintptr(hdc), uintptr(index))
	return int(c)
}

func getDpiForMonitor(hmonitor syscall.Handle, dpiType uint32) int {
	var dpiX, dpiY uintptr
	_GetDpiForMonitor.Call(uintptr(hmonitor), uintptr(dpiType), uintptr(unsafe.Pointer(&dpiX)), uintptr(unsafe.Pointer(&dpiY)))
	return int(dpiX)
}

// GetSystemDPI returns the effective DPI of the system.
func GetSystemDPI() int {
	// Check for GetDpiForMonitor, introduced in Windows 8.1.
	if _GetDpiForMonitor.Find() == nil {
		hmon := monitorFromPoint(Point{}, MONITOR_DEFAULTTOPRIMARY)
		return getDpiForMonitor(hmon, MDT_EFFECTIVE_DPI)
	} else {
		// Fall back to the physical device DPI.
		screenDC, err := GetDC(0)
		if err != nil {
			return 96
		}
		defer ReleaseDC(screenDC)
		return getDeviceCaps(screenDC, LOGPIXELSX)
	}
}

func GetKeyState(nVirtKey int32) int16 {
	c, _, _ := _GetKeyState.Call(uintptr(nVirtKey))
	return int16(c)
}

func GetMessage(m *Msg, hwnd syscall.Handle, wMsgFilterMin, wMsgFilterMax uint32) int32 {
	r, _, _ := _GetMessage.Call(uintptr(unsafe.Pointer(m)),
		uintptr(hwnd),
		uintptr(wMsgFilterMin),
		uintptr(wMsgFilterMax))
	return int32(r)
}

func GetMessageTime() time.Duration {
	r, _, _ := _GetMessageTime.Call()
	return time.Duration(r) * time.Millisecond
}

func GetSystemMetrics(nIndex int) int {
	r, _, _ := _GetSystemMetrics.Call(uintptr(nIndex))
	return int(r)
}

// GetWindowDPI returns the effective DPI of the window.
func GetWindowDPI(hwnd syscall.Handle) int {
	// Check for GetDpiForWindow, introduced in Windows 10.
	if _GetDpiForWindow.Find() == nil {
		dpi, _, _ := _GetDpiForWindow.Call(uintptr(hwnd))
		return int(dpi)
	} else {
		return GetSystemDPI()
	}
}

func GetWindowPlacement(hwnd syscall.Handle) *WindowPlacement {
	var wp WindowPlacement
	wp.length = uint32(unsafe.Sizeof(wp))
	_GetWindowPlacement.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&wp)))
	return &wp
}

func GetMonitorInfo(hwnd syscall.Handle) MonitorInfo {
	var mi MonitorInfo
	mi.cbSize = uint32(unsafe.Sizeof(mi))
	v, _, _ := _MonitorFromWindow.Call(uintptr(hwnd), MONITOR_DEFAULTTOPRIMARY)
	_GetMonitorInfo.Call(v, uintptr(unsafe.Pointer(&mi)))
	return mi
}

func GetWindowLong(hwnd syscall.Handle, index uintptr) (val uintptr) {
	if runtime.GOARCH == "386" {
		val, _, _ = _GetWindowLong32.Call(uintptr(hwnd), index)
	} else {
		val, _, _ = _GetWindowLong.Call(uintptr(hwnd), index)
	}
	return
}

func ImmGetContext(hwnd syscall.Handle) syscall.Handle {
	h, _, _ := _ImmGetContext.Call(uintptr(hwnd))
	return syscall.Handle(h)
}

func ImmReleaseContext(hwnd, imc syscall.Handle) {
	_ImmReleaseContext.Call(uintptr(hwnd), uintptr(imc))
}

func ImmNotifyIME(imc syscall.Handle, action, index, value int) {
	_ImmNotifyIME.Call(uintptr(imc), uintptr(action), uintptr(index), uintptr(value))
}

func ImmGetCompositionString(imc syscall.Handle, key int) string {
	size, _, _ := _ImmGetCompositionString.Call(uintptr(imc), uintptr(key), 0, 0)
	if int32(size) <= 0 {
		return ""
	}
	u16 := make([]uint16, size/unsafe.Sizeof(uint16(0)))
	_ImmGetCompositionString.Call(uintptr(imc), uintptr(key), uintptr(unsafe.Pointer(&u16[0])), size)
	return string(utf16.Decode(u16))
}

func ImmGetCompositionValue(imc syscall.Handle, key int) int {
	val, _, _ := _ImmGetCompositionString.Call(uintptr(imc), uintptr(key), 0, 0)
	return int(int32(val))
}

func ImmSetCompositionWindow(imc syscall.Handle, x, y int) {
	f := CompositionForm{
		dwStyle: CFS_POINT,
		ptCurrentPos: Point{
			X: int32(x), Y: int32(y),
		},
	}
	_ImmSetCompositionWindow.Call(uintptr(imc), uintptr(unsafe.Pointer(&f)))
}

func ImmSetCandidateWindow(imc syscall.Handle, x, y int) {
	f := CandidateForm{
		dwStyle: CFS_CANDIDATEPOS,
		ptCurrentPos: Point{
			X: int32(x), Y: int32(y),
		},
	}
	_ImmSetCandidateWindow.Call(uintptr(imc), uintptr(unsafe.Pointer(&f)))
}

func SetWindowLong(hwnd syscall.Handle, idx uintptr, style uintptr) {
	if runtime.GOARCH == "386" {
		_SetWindowLong32.Call(uintptr(hwnd), idx, style)
	} else {
		_SetWindowLong.Call(uintptr(hwnd), idx, style)
	}
}

func SetWindowPlacement(hwnd syscall.Handle, wp *WindowPlacement) {
	_SetWindowPlacement.Call(uintptr(hwnd), uintptr(unsafe.Pointer(wp)))
}

func SetWindowPos(hwnd syscall.Handle, hwndInsertAfter uint32, x, y, dx, dy int32, style uintptr) {
	_SetWindowPos.Call(uintptr(hwnd), uintptr(hwndInsertAfter),
		uintptr(x), uintptr(y),
		uintptr(dx), uintptr(dy),
		style,
	)
}

func SetWindowText(hwnd syscall.Handle, title string) {
	wname := syscall.StringToUTF16Ptr(title)
	_SetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(wname)))
}

func GlobalAlloc(size int) (syscall.Handle, error) {
	r, _, err := _GlobalAlloc.Call(GHND, uintptr(size))
	if r == 0 {
		return 0, fmt.Errorf("GlobalAlloc: %v", err)
	}
	return syscall.Handle(r), nil
}

func GlobalFree(h syscall.Handle) {
	_GlobalFree.Call(uintptr(h))
}

func GlobalLock(h syscall.Handle) (unsafe.Pointer, error) {
	r, _, err := _GlobalLock.Call(uintptr(h))
	if r == 0 {
		return nil, fmt.Errorf("GlobalLock: %v", err)
	}
	return unsafe.Pointer(r), nil
}

func GlobalUnlock(h syscall.Handle) {
	_GlobalUnlock.Call(uintptr(h))
}

func KillTimer(hwnd syscall.Handle, nIDEvent uintptr) error {
	r, _, err := _SetTimer.Call(uintptr(hwnd), uintptr(nIDEvent), 0, 0)
	if r == 0 {
		return fmt.Errorf("KillTimer failed: %v", err)
	}
	return nil
}

func LoadCursor(curID uint16) (syscall.Handle, error) {
	h, _, err := _LoadCursor.Call(0, uintptr(curID))
	if h == 0 {
		return 0, fmt.Errorf("LoadCursorW failed: %v", err)
	}
	return syscall.Handle(h), nil
}

func LoadImage(hInst syscall.Handle, res uint32, typ uint32, cx, cy int, fuload uint32) (syscall.Handle, error) {
	h, _, err := _LoadImage.Call(uintptr(hInst), uintptr(res), uintptr(typ), uintptr(cx), uintptr(cy), uintptr(fuload))
	if h == 0 {
		return 0, fmt.Errorf("LoadImageW failed: %v", err)
	}
	return syscall.Handle(h), nil
}

func MoveWindow(hwnd syscall.Handle, x, y, width, height int32, repaint bool) {
	var paint uintptr
	if repaint {
		paint = TRUE
	}
	_MoveWindow.Call(uintptr(hwnd), uintptr(x), uintptr(y), uintptr(width), uintptr(height), paint)
}

func monitorFromPoint(pt Point, flags uint32) syscall.Handle {
	r, _, _ := _MonitorFromPoint.Call(uintptr(pt.X), uintptr(pt.Y), uintptr(flags))
	return syscall.Handle(r)
}

func MsgWaitForMultipleObjectsEx(nCount uint32, pHandles uintptr, millis, mask, flags uint32) (uint32, error) {
	r, _, err := _MsgWaitForMultipleObjectsEx.Call(uintptr(nCount), pHandles, uintptr(millis), uintptr(mask), uintptr(flags))
	res := uint32(r)
	if res == 0xFFFFFFFF {
		return 0, fmt.Errorf("MsgWaitForMultipleObjectsEx failed: %v", err)
	}
	return res, nil
}

func OpenClipboard(hwnd syscall.Handle) error {
	r, _, err := _OpenClipboard.Call(uintptr(hwnd))
	if r == 0 {
		return fmt.Errorf("OpenClipboard: %v", err)
	}
	return nil
}

func PeekMessage(m *Msg, hwnd syscall.Handle, wMsgFilterMin, wMsgFilterMax, wRemoveMsg uint32) bool {
	r, _, _ := _PeekMessage.Call(uintptr(unsafe.Pointer(m)), uintptr(hwnd), uintptr(wMsgFilterMin), uintptr(wMsgFilterMax), uintptr(wRemoveMsg))
	return r != 0
}

func PostQuitMessage(exitCode uintptr) {
	_PostQuitMessage.Call(exitCode)
}

func PostMessage(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) error {
	r, _, err := _PostMessage.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	if r == 0 {
		return fmt.Errorf("PostMessage failed: %v", err)
	}
	return nil
}

func ReleaseCapture() bool {
	r, _, _ := _ReleaseCapture.Call()
	return r != 0
}

func RegisterClassEx(cls *WndClassEx) (uint16, error) {
	a, _, err := _RegisterClassExW.Call(uintptr(unsafe.Pointer(cls)))
	if a == 0 {
		return 0, fmt.Errorf("RegisterClassExW failed: %v", err)
	}
	return uint16(a), nil
}

func ReleaseDC(hdc syscall.Handle) {
	_ReleaseDC.Call(uintptr(hdc))
}

func SetForegroundWindow(hwnd syscall.Handle) {
	_SetForegroundWindow.Call(uintptr(hwnd))
}

func SetFocus(hwnd syscall.Handle) {
	_SetFocus.Call(uintptr(hwnd))
}

func SetProcessDPIAware() {
	_SetProcessDPIAware.Call()
}

func SetCapture(hwnd syscall.Handle) syscall.Handle {
	r, _, _ := _SetCapture.Call(uintptr(hwnd))
	return syscall.Handle(r)
}

func SetClipboardData(format uint32, mem syscall.Handle) error {
	r, _, err := _SetClipboardData.Call(uintptr(format), uintptr(mem))
	if r == 0 {
		return fmt.Errorf("SetClipboardData: %v", err)
	}
	return nil
}

func SetCursor(h syscall.Handle) {
	_SetCursor.Call(uintptr(h))
}

func SetTimer(hwnd syscall.Handle, nIDEvent uintptr, uElapse uint32, timerProc uintptr) error {
	r, _, err := _SetTimer.Call(uintptr(hwnd), uintptr(nIDEvent), uintptr(uElapse), timerProc)
	if r == 0 {
		return fmt.Errorf("SetTimer failed: %v", err)
	}
	return nil
}

func ScreenToClient(hwnd syscall.Handle, p *Point) {
	_ScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(p)))
}

func ShowWindow(hwnd syscall.Handle, nCmdShow int32) {
	_ShowWindow.Call(uintptr(hwnd), uintptr(nCmdShow))
}

func TranslateMessage(m *Msg) {
	_TranslateMessage.Call(uintptr(unsafe.Pointer(m)))
}

func UnregisterClass(cls uint16, hInst syscall.Handle) {
	_UnregisterClass.Call(uintptr(cls), uintptr(hInst))
}

func UpdateWindow(hwnd syscall.Handle) {
	_UpdateWindow.Call(uintptr(hwnd))
}

func (p WindowPlacement) Rect() Rect {
	return p.rcNormalPosition
}

func (p WindowPlacement) IsMinimized() bool {
	return p.showCmd == SW_SHOWMINIMIZED
}

func (p WindowPlacement) IsMaximized() bool {
	return p.showCmd == SW_SHOWMAXIMIZED
}

func (p *WindowPlacement) Set(Left, Top, Right, Bottom int) {
	p.rcNormalPosition.Left = int32(Left)
	p.rcNormalPosition.Top = int32(Top)
	p.rcNormalPosition.Right = int32(Right)
	p.rcNormalPosition.Bottom = int32(Bottom)
}
