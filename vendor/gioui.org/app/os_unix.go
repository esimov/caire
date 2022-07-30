// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux && !android) || freebsd || openbsd
// +build linux,!android freebsd openbsd

package app

import (
	"errors"
	"unsafe"

	"gioui.org/io/pointer"
)

// ViewEvent provides handles to the underlying window objects for the
// current display protocol.
type ViewEvent interface {
	implementsViewEvent()
	ImplementsEvent()
}

type X11ViewEvent struct {
	// Display is a pointer to the X11 Display created by XOpenDisplay.
	Display unsafe.Pointer
	// Window is the X11 window ID as returned by XCreateWindow.
	Window uintptr
}

func (X11ViewEvent) implementsViewEvent() {}
func (X11ViewEvent) ImplementsEvent()     {}

type WaylandViewEvent struct {
	// Display is the *wl_display returned by wl_display_connect.
	Display unsafe.Pointer
	// Surface is the *wl_surface returned by wl_compositor_create_surface.
	Surface unsafe.Pointer
}

func (WaylandViewEvent) implementsViewEvent() {}
func (WaylandViewEvent) ImplementsEvent()     {}

func osMain() {
	select {}
}

type windowDriver func(*callbacks, []Option) error

// Instead of creating files with build tags for each combination of wayland +/- x11
// let each driver initialize these variables with their own version of createWindow.
var wlDriver, x11Driver windowDriver

func newWindow(window *callbacks, options []Option) error {
	var errFirst error
	for _, d := range []windowDriver{wlDriver, x11Driver} {
		if d == nil {
			continue
		}
		err := d(window, options)
		if err == nil {
			return nil
		}
		if errFirst == nil {
			errFirst = err
		}
	}
	if errFirst != nil {
		return errFirst
	}
	return errors.New("app: no window driver available")
}

// xCursor contains mapping from pointer.Cursor to XCursor.
var xCursor = [...]string{
	pointer.CursorDefault:                  "left_ptr",
	pointer.CursorNone:                     "",
	pointer.CursorText:                     "xterm",
	pointer.CursorVerticalText:             "vertical-text",
	pointer.CursorPointer:                  "hand2",
	pointer.CursorCrosshair:                "crosshair",
	pointer.CursorAllScroll:                "fleur",
	pointer.CursorColResize:                "sb_h_double_arrow",
	pointer.CursorRowResize:                "sb_v_double_arrow",
	pointer.CursorGrab:                     "hand1",
	pointer.CursorGrabbing:                 "move",
	pointer.CursorNotAllowed:               "crossed_circle",
	pointer.CursorWait:                     "watch",
	pointer.CursorProgress:                 "left_ptr_watch",
	pointer.CursorNorthWestResize:          "top_left_corner",
	pointer.CursorNorthEastResize:          "top_right_corner",
	pointer.CursorSouthWestResize:          "bottom_left_corner",
	pointer.CursorSouthEastResize:          "bottom_right_corner",
	pointer.CursorNorthSouthResize:         "sb_v_double_arrow",
	pointer.CursorEastWestResize:           "sb_h_double_arrow",
	pointer.CursorWestResize:               "left_side",
	pointer.CursorEastResize:               "right_side",
	pointer.CursorNorthResize:              "top_side",
	pointer.CursorSouthResize:              "bottom_side",
	pointer.CursorNorthEastSouthWestResize: "fd_double_arrow",
	pointer.CursorNorthWestSouthEastResize: "bd_double_arrow",
}
