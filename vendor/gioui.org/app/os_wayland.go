// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland
// +build linux,!android freebsd
// +build !nowayland

package app

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
	"unsafe"

	syscall "golang.org/x/sys/unix"

	"gioui.org/app/internal/xkb"
	"gioui.org/f32"
	"gioui.org/internal/fling"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/unit"
)

// Use wayland-scanner to generate glue code for the xdg-shell and xdg-decoration extensions.
//go:generate wayland-scanner client-header /usr/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wayland_xdg_shell.h
//go:generate wayland-scanner private-code /usr/share/wayland-protocols/stable/xdg-shell/xdg-shell.xml wayland_xdg_shell.c

//go:generate wayland-scanner client-header /usr/share/wayland-protocols/unstable/text-input/text-input-unstable-v3.xml wayland_text_input.h
//go:generate wayland-scanner private-code /usr/share/wayland-protocols/unstable/text-input/text-input-unstable-v3.xml wayland_text_input.c

//go:generate wayland-scanner client-header /usr/share/wayland-protocols/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml wayland_xdg_decoration.h
//go:generate wayland-scanner private-code /usr/share/wayland-protocols/unstable/xdg-decoration/xdg-decoration-unstable-v1.xml wayland_xdg_decoration.c

//go:generate sed -i "1s;^;//go:build ((linux \\&\\& !android) || freebsd) \\&\\& !nowayland\\n// +build linux,!android freebsd\\n// +build !nowayland\\n\\n;" wayland_xdg_shell.c
//go:generate sed -i "1s;^;//go:build ((linux \\&\\& !android) || freebsd) \\&\\& !nowayland\\n// +build linux,!android freebsd\\n// +build !nowayland\\n\\n;" wayland_xdg_decoration.c
//go:generate sed -i "1s;^;//go:build ((linux \\&\\& !android) || freebsd) \\&\\& !nowayland\\n// +build linux,!android freebsd\\n// +build !nowayland\\n\\n;" wayland_text_input.c

/*
#cgo linux pkg-config: wayland-client wayland-cursor
#cgo freebsd openbsd LDFLAGS: -lwayland-client -lwayland-cursor
#cgo freebsd CFLAGS: -I/usr/local/include
#cgo freebsd LDFLAGS: -L/usr/local/lib

#include <stdlib.h>
#include <wayland-client.h>
#include <wayland-cursor.h>
#include "wayland_text_input.h"
#include "wayland_xdg_shell.h"
#include "wayland_xdg_decoration.h"

extern const struct wl_registry_listener gio_registry_listener;
extern const struct wl_surface_listener gio_surface_listener;
extern const struct xdg_surface_listener gio_xdg_surface_listener;
extern const struct xdg_toplevel_listener gio_xdg_toplevel_listener;
extern const struct zxdg_toplevel_decoration_v1_listener gio_zxdg_toplevel_decoration_v1_listener;
extern const struct xdg_wm_base_listener gio_xdg_wm_base_listener;
extern const struct wl_callback_listener gio_callback_listener;
extern const struct wl_output_listener gio_output_listener;
extern const struct wl_seat_listener gio_seat_listener;
extern const struct wl_pointer_listener gio_pointer_listener;
extern const struct wl_touch_listener gio_touch_listener;
extern const struct wl_keyboard_listener gio_keyboard_listener;
extern const struct zwp_text_input_v3_listener gio_zwp_text_input_v3_listener;
extern const struct wl_data_device_listener gio_data_device_listener;
extern const struct wl_data_offer_listener gio_data_offer_listener;
extern const struct wl_data_source_listener gio_data_source_listener;
*/
import "C"

type wlDisplay struct {
	disp              *C.struct_wl_display
	reg               *C.struct_wl_registry
	compositor        *C.struct_wl_compositor
	wm                *C.struct_xdg_wm_base
	imm               *C.struct_zwp_text_input_manager_v3
	shm               *C.struct_wl_shm
	dataDeviceManager *C.struct_wl_data_device_manager
	decor             *C.struct_zxdg_decoration_manager_v1
	seat              *wlSeat
	xkb               *xkb.Context
	outputMap         map[C.uint32_t]*C.struct_wl_output
	outputConfig      map[*C.struct_wl_output]*wlOutput

	// Notification pipe fds.
	notify struct {
		read, write int
	}

	repeat repeatState
}

type wlSeat struct {
	disp     *wlDisplay
	seat     *C.struct_wl_seat
	name     C.uint32_t
	pointer  *C.struct_wl_pointer
	touch    *C.struct_wl_touch
	keyboard *C.struct_wl_keyboard
	im       *C.struct_zwp_text_input_v3

	// The most recent input serial.
	serial C.uint32_t

	pointerFocus  *window
	keyboardFocus *window
	touchFoci     map[C.int32_t]*window

	// Clipboard support.
	dataDev *C.struct_wl_data_device
	// offers is a map from active wl_data_offers to
	// the list of mime types they support.
	offers map[*C.struct_wl_data_offer][]string
	// clipboard is the wl_data_offer for the clipboard.
	clipboard *C.struct_wl_data_offer
	// mimeType is the chosen mime type of clipboard.
	mimeType string
	// source represents the clipboard content of the most recent
	// clipboard write, if any.
	source *C.struct_wl_data_source
	// content is the data belonging to source.
	content []byte
}

type repeatState struct {
	rate  int
	delay time.Duration

	key   uint32
	win   *callbacks
	stopC chan struct{}

	start time.Duration
	last  time.Duration
	mu    sync.Mutex
	now   time.Duration
}

type window struct {
	w          *callbacks
	disp       *wlDisplay
	seat       *wlSeat
	surf       *C.struct_wl_surface
	wmSurf     *C.struct_xdg_surface
	topLvl     *C.struct_xdg_toplevel
	decor      *C.struct_zxdg_toplevel_decoration_v1
	ppdp, ppsp float32
	scroll     struct {
		time  time.Duration
		steps image.Point
		dist  f32.Point
	}
	pointerBtns pointer.Buttons
	lastPos     f32.Point
	lastTouch   f32.Point

	cursor struct {
		theme  *C.struct_wl_cursor_theme
		cursor *C.struct_wl_cursor
		// system is the active cursor for system gestures
		// such as border resizes and window moves. It
		// is nil if the pointer is not in a system gesture
		// area.
		system  *C.struct_wl_cursor
		surf    *C.struct_wl_surface
		cursors struct {
			pointer         *C.struct_wl_cursor
			resizeNorth     *C.struct_wl_cursor
			resizeSouth     *C.struct_wl_cursor
			resizeWest      *C.struct_wl_cursor
			resizeEast      *C.struct_wl_cursor
			resizeNorthWest *C.struct_wl_cursor
			resizeNorthEast *C.struct_wl_cursor
			resizeSouthWest *C.struct_wl_cursor
			resizeSouthEast *C.struct_wl_cursor
		}
	}

	fling struct {
		yExtrapolation fling.Extrapolation
		xExtrapolation fling.Extrapolation
		anim           fling.Animation
		start          bool
		dir            f32.Point
	}

	stage             system.Stage
	dead              bool
	lastFrameCallback *C.struct_wl_callback

	animating bool
	redraw    bool
	// The most recent configure serial waiting to be ack'ed.
	serial C.uint32_t
	scale  int
	// size is the unscaled window size (unlike config.Size which is scaled).
	size         image.Point
	config       Config
	wsize        image.Point // window config size before going fullscreen or maximized
	inCompositor bool        // window is moving or being resized

	clipReads chan clipboard.Event

	wakeups chan struct{}
}

type poller struct {
	pollfds [2]syscall.PollFd
	// buf is scratch space for draining the notification pipe.
	buf [100]byte
}

type wlOutput struct {
	width      int
	height     int
	physWidth  int
	physHeight int
	transform  C.int32_t
	scale      int
	windows    []*window
}

// callbackMap maps Wayland native handles to corresponding Go
// references. It is necessary because the Wayland client API
// forces the use of callbacks and storing pointers to Go values
// in C is forbidden.
var callbackMap sync.Map

// clipboardMimeTypes is a list of supported clipboard mime types, in
// order of preference.
var clipboardMimeTypes = []string{"text/plain;charset=utf8", "UTF8_STRING", "text/plain", "TEXT", "STRING"}

var (
	newWaylandEGLContext    func(w *window) (context, error)
	newWaylandVulkanContext func(w *window) (context, error)
)

func init() {
	wlDriver = newWLWindow
}

func newWLWindow(callbacks *callbacks, options []Option) error {
	d, err := newWLDisplay()
	if err != nil {
		return err
	}
	w, err := d.createNativeWindow(options)
	if err != nil {
		d.destroy()
		return err
	}
	w.w = callbacks
	go func() {
		defer d.destroy()
		defer w.destroy()

		w.w.SetDriver(w)

		// Finish and commit setup from createNativeWindow.
		w.Configure(options)
		C.wl_surface_commit(w.surf)

		w.w.Event(WaylandViewEvent{
			Display: unsafe.Pointer(w.display()),
			Surface: unsafe.Pointer(w.surf),
		})

		err := w.loop()
		w.w.Event(WaylandViewEvent{})
		w.w.Event(system.DestroyEvent{Err: err})
	}()
	return nil
}

func (d *wlDisplay) writeClipboard(content []byte) error {
	s := d.seat
	if s == nil {
		return nil
	}
	// Clear old offer.
	if s.source != nil {
		C.wl_data_source_destroy(s.source)
		s.source = nil
		s.content = nil
	}
	if d.dataDeviceManager == nil || s.dataDev == nil {
		return nil
	}
	s.content = content
	s.source = C.wl_data_device_manager_create_data_source(d.dataDeviceManager)
	C.wl_data_source_add_listener(s.source, &C.gio_data_source_listener, unsafe.Pointer(s.seat))
	for _, mime := range clipboardMimeTypes {
		C.wl_data_source_offer(s.source, C.CString(mime))
	}
	C.wl_data_device_set_selection(s.dataDev, s.source, s.serial)
	return nil
}

func (d *wlDisplay) readClipboard() (io.ReadCloser, error) {
	s := d.seat
	if s == nil {
		return nil, nil
	}
	if s.clipboard == nil {
		return nil, nil
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	// wl_data_offer_receive performs and implicit dup(2) of the write end
	// of the pipe. Close our version.
	defer w.Close()
	cmimeType := C.CString(s.mimeType)
	defer C.free(unsafe.Pointer(cmimeType))
	C.wl_data_offer_receive(s.clipboard, cmimeType, C.int(w.Fd()))
	return r, nil
}

func (d *wlDisplay) createNativeWindow(options []Option) (*window, error) {
	if d.compositor == nil {
		return nil, errors.New("wayland: no compositor available")
	}
	if d.wm == nil {
		return nil, errors.New("wayland: no xdg_wm_base available")
	}
	if d.shm == nil {
		return nil, errors.New("wayland: no wl_shm available")
	}
	if len(d.outputMap) == 0 {
		return nil, errors.New("wayland: no outputs available")
	}
	var scale int
	for _, conf := range d.outputConfig {
		if s := conf.scale; s > scale {
			scale = s
		}
	}
	ppdp := detectUIScale()

	w := &window{
		disp:      d,
		scale:     scale,
		ppdp:      ppdp,
		ppsp:      ppdp,
		wakeups:   make(chan struct{}, 1),
		clipReads: make(chan clipboard.Event, 1),
	}
	w.surf = C.wl_compositor_create_surface(d.compositor)
	if w.surf == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_compositor_create_surface failed")
	}
	C.wl_surface_set_buffer_scale(w.surf, C.int32_t(w.scale))
	callbackStore(unsafe.Pointer(w.surf), w)
	w.wmSurf = C.xdg_wm_base_get_xdg_surface(d.wm, w.surf)
	if w.wmSurf == nil {
		w.destroy()
		return nil, errors.New("wayland: xdg_wm_base_get_xdg_surface failed")
	}
	w.topLvl = C.xdg_surface_get_toplevel(w.wmSurf)
	if w.topLvl == nil {
		w.destroy()
		return nil, errors.New("wayland: xdg_surface_get_toplevel failed")
	}
	cursorTheme := C.CString(os.Getenv("XCURSOR_THEME"))
	defer C.free(unsafe.Pointer(cursorTheme))
	cursorSize := 32
	if envSize, ok := os.LookupEnv("XCURSOR_SIZE"); ok && envSize != "" {
		size, err := strconv.Atoi(envSize)
		if err == nil {
			cursorSize = size
		}
	}

	w.cursor.theme = C.wl_cursor_theme_load(cursorTheme, C.int(cursorSize*w.scale), d.shm)
	if w.cursor.theme == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_cursor_theme_load failed")
	}
	w.loadCursors()
	w.cursor.cursor = w.cursor.cursors.pointer
	if w.cursor.cursor == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_cursor_theme_get_cursor failed")
	}
	w.cursor.surf = C.wl_compositor_create_surface(d.compositor)
	if w.cursor.surf == nil {
		w.destroy()
		return nil, errors.New("wayland: wl_compositor_create_surface failed")
	}
	C.wl_surface_set_buffer_scale(w.cursor.surf, C.int32_t(w.scale))
	C.xdg_wm_base_add_listener(d.wm, &C.gio_xdg_wm_base_listener, unsafe.Pointer(w.surf))
	C.wl_surface_add_listener(w.surf, &C.gio_surface_listener, unsafe.Pointer(w.surf))
	C.xdg_surface_add_listener(w.wmSurf, &C.gio_xdg_surface_listener, unsafe.Pointer(w.surf))
	C.xdg_toplevel_add_listener(w.topLvl, &C.gio_xdg_toplevel_listener, unsafe.Pointer(w.surf))

	if d.decor != nil {
		w.decor = C.zxdg_decoration_manager_v1_get_toplevel_decoration(d.decor, w.topLvl)
		C.zxdg_toplevel_decoration_v1_add_listener(w.decor, &C.gio_zxdg_toplevel_decoration_v1_listener, unsafe.Pointer(w.surf))
	}
	w.updateOpaqueRegion()
	return w, nil
}

func (w *window) loadCursors() {
	w.cursor.cursors.pointer = w.loadCursor(pointer.CursorDefault)
	w.cursor.cursors.resizeNorth = w.loadCursor(pointer.CursorNorthResize)
	w.cursor.cursors.resizeSouth = w.loadCursor(pointer.CursorSouthResize)
	w.cursor.cursors.resizeWest = w.loadCursor(pointer.CursorWestResize)
	w.cursor.cursors.resizeEast = w.loadCursor(pointer.CursorEastResize)
	w.cursor.cursors.resizeSouthWest = w.loadCursor(pointer.CursorSouthWestResize)
	w.cursor.cursors.resizeSouthEast = w.loadCursor(pointer.CursorSouthEastResize)
	w.cursor.cursors.resizeNorthWest = w.loadCursor(pointer.CursorNorthWestResize)
	w.cursor.cursors.resizeNorthEast = w.loadCursor(pointer.CursorNorthEastResize)
}

func (w *window) loadCursor(name pointer.Cursor) *C.struct_wl_cursor {
	if name == pointer.CursorNone {
		return nil
	}
	xcursor := xCursor[name]
	cname := C.CString(xcursor)
	defer C.free(unsafe.Pointer(cname))
	c := C.wl_cursor_theme_get_cursor(w.cursor.theme, cname)
	if c == nil {
		// Fall back to default cursor.
		c = w.cursor.cursors.pointer
	}
	return c
}

func callbackDelete(k unsafe.Pointer) {
	callbackMap.Delete(k)
}

func callbackStore(k unsafe.Pointer, v interface{}) {
	callbackMap.Store(k, v)
}

func callbackLoad(k unsafe.Pointer) interface{} {
	v, exists := callbackMap.Load(k)
	if !exists {
		panic("missing callback entry")
	}
	return v
}

//export gio_onSeatCapabilities
func gio_onSeatCapabilities(data unsafe.Pointer, seat *C.struct_wl_seat, caps C.uint32_t) {
	s := callbackLoad(data).(*wlSeat)
	s.updateCaps(caps)
}

// flushOffers remove all wl_data_offers that isn't the clipboard
// content.
func (s *wlSeat) flushOffers() {
	for o := range s.offers {
		if o == s.clipboard {
			continue
		}
		// We're only interested in clipboard offers.
		delete(s.offers, o)
		callbackDelete(unsafe.Pointer(o))
		C.wl_data_offer_destroy(o)
	}
}

func (s *wlSeat) destroy() {
	if s.source != nil {
		C.wl_data_source_destroy(s.source)
		s.source = nil
	}
	if s.im != nil {
		C.zwp_text_input_v3_destroy(s.im)
		s.im = nil
	}
	if s.pointer != nil {
		C.wl_pointer_release(s.pointer)
	}
	if s.touch != nil {
		C.wl_touch_release(s.touch)
	}
	if s.keyboard != nil {
		C.wl_keyboard_release(s.keyboard)
	}
	s.clipboard = nil
	s.flushOffers()
	if s.dataDev != nil {
		C.wl_data_device_release(s.dataDev)
	}
	if s.seat != nil {
		callbackDelete(unsafe.Pointer(s.seat))
		C.wl_seat_release(s.seat)
	}
}

func (s *wlSeat) updateCaps(caps C.uint32_t) {
	if s.im == nil && s.disp.imm != nil {
		s.im = C.zwp_text_input_manager_v3_get_text_input(s.disp.imm, s.seat)
		C.zwp_text_input_v3_add_listener(s.im, &C.gio_zwp_text_input_v3_listener, unsafe.Pointer(s.seat))
	}
	switch {
	case s.pointer == nil && caps&C.WL_SEAT_CAPABILITY_POINTER != 0:
		s.pointer = C.wl_seat_get_pointer(s.seat)
		C.wl_pointer_add_listener(s.pointer, &C.gio_pointer_listener, unsafe.Pointer(s.seat))
	case s.pointer != nil && caps&C.WL_SEAT_CAPABILITY_POINTER == 0:
		C.wl_pointer_release(s.pointer)
		s.pointer = nil
	}
	switch {
	case s.touch == nil && caps&C.WL_SEAT_CAPABILITY_TOUCH != 0:
		s.touch = C.wl_seat_get_touch(s.seat)
		C.wl_touch_add_listener(s.touch, &C.gio_touch_listener, unsafe.Pointer(s.seat))
	case s.touch != nil && caps&C.WL_SEAT_CAPABILITY_TOUCH == 0:
		C.wl_touch_release(s.touch)
		s.touch = nil
	}
	switch {
	case s.keyboard == nil && caps&C.WL_SEAT_CAPABILITY_KEYBOARD != 0:
		s.keyboard = C.wl_seat_get_keyboard(s.seat)
		C.wl_keyboard_add_listener(s.keyboard, &C.gio_keyboard_listener, unsafe.Pointer(s.seat))
	case s.keyboard != nil && caps&C.WL_SEAT_CAPABILITY_KEYBOARD == 0:
		C.wl_keyboard_release(s.keyboard)
		s.keyboard = nil
	}
}

//export gio_onSeatName
func gio_onSeatName(data unsafe.Pointer, seat *C.struct_wl_seat, name *C.char) {
}

//export gio_onXdgSurfaceConfigure
func gio_onXdgSurfaceConfigure(data unsafe.Pointer, wmSurf *C.struct_xdg_surface, serial C.uint32_t) {
	w := callbackLoad(data).(*window)
	w.serial = serial
	w.redraw = true
	C.xdg_surface_ack_configure(wmSurf, serial)
	w.setStage(system.StageRunning)
}

//export gio_onToplevelClose
func gio_onToplevelClose(data unsafe.Pointer, topLvl *C.struct_xdg_toplevel) {
	w := callbackLoad(data).(*window)
	w.dead = true
}

//export gio_onToplevelConfigure
func gio_onToplevelConfigure(data unsafe.Pointer, topLvl *C.struct_xdg_toplevel, width, height C.int32_t, states *C.struct_wl_array) {
	w := callbackLoad(data).(*window)
	if width != 0 && height != 0 {
		w.size = image.Pt(int(width), int(height))
		w.updateOpaqueRegion()
	}
}

//export gio_onToplevelDecorationConfigure
func gio_onToplevelDecorationConfigure(data unsafe.Pointer, deco *C.struct_zxdg_toplevel_decoration_v1, mode C.uint32_t) {
	w := callbackLoad(data).(*window)
	decorated := w.config.Decorated
	switch mode {
	case C.ZXDG_TOPLEVEL_DECORATION_V1_MODE_CLIENT_SIDE:
		w.config.Decorated = false
	case C.ZXDG_TOPLEVEL_DECORATION_V1_MODE_SERVER_SIDE:
		w.config.Decorated = true
	}
	if decorated != w.config.Decorated {
		w.setWindowConstraints()
		if w.config.Decorated {
			w.size.Y -= int(w.config.decoHeight)
		} else {
			w.size.Y += int(w.config.decoHeight)
		}
		w.w.Event(ConfigEvent{Config: w.config})
		w.redraw = true
	}
}

//export gio_onOutputMode
func gio_onOutputMode(data unsafe.Pointer, output *C.struct_wl_output, flags C.uint32_t, width, height, refresh C.int32_t) {
	if flags&C.WL_OUTPUT_MODE_CURRENT == 0 {
		return
	}
	d := callbackLoad(data).(*wlDisplay)
	c := d.outputConfig[output]
	c.width = int(width)
	c.height = int(height)
}

//export gio_onOutputGeometry
func gio_onOutputGeometry(data unsafe.Pointer, output *C.struct_wl_output, x, y, physWidth, physHeight, subpixel C.int32_t, make, model *C.char, transform C.int32_t) {
	d := callbackLoad(data).(*wlDisplay)
	c := d.outputConfig[output]
	c.transform = transform
	c.physWidth = int(physWidth)
	c.physHeight = int(physHeight)
}

//export gio_onOutputScale
func gio_onOutputScale(data unsafe.Pointer, output *C.struct_wl_output, scale C.int32_t) {
	d := callbackLoad(data).(*wlDisplay)
	c := d.outputConfig[output]
	c.scale = int(scale)
}

//export gio_onOutputDone
func gio_onOutputDone(data unsafe.Pointer, output *C.struct_wl_output) {
	d := callbackLoad(data).(*wlDisplay)
	conf := d.outputConfig[output]
	for _, w := range conf.windows {
		w.updateOutputs()
	}
}

//export gio_onSurfaceEnter
func gio_onSurfaceEnter(data unsafe.Pointer, surf *C.struct_wl_surface, output *C.struct_wl_output) {
	w := callbackLoad(data).(*window)
	conf := w.disp.outputConfig[output]
	var found bool
	for _, w2 := range conf.windows {
		if w2 == w {
			found = true
			break
		}
	}
	if !found {
		conf.windows = append(conf.windows, w)
	}
	w.updateOutputs()
	if w.config.Mode == Minimized {
		// Minimized window got brought back up: it is no longer so.
		w.config.Mode = Windowed
		w.w.Event(ConfigEvent{Config: w.config})
	}
}

//export gio_onSurfaceLeave
func gio_onSurfaceLeave(data unsafe.Pointer, surf *C.struct_wl_surface, output *C.struct_wl_output) {
	w := callbackLoad(data).(*window)
	conf := w.disp.outputConfig[output]
	for i, w2 := range conf.windows {
		if w2 == w {
			conf.windows = append(conf.windows[:i], conf.windows[i+1:]...)
			break
		}
	}
	w.updateOutputs()
}

//export gio_onRegistryGlobal
func gio_onRegistryGlobal(data unsafe.Pointer, reg *C.struct_wl_registry, name C.uint32_t, cintf *C.char, version C.uint32_t) {
	d := callbackLoad(data).(*wlDisplay)
	switch C.GoString(cintf) {
	case "wl_compositor":
		d.compositor = (*C.struct_wl_compositor)(C.wl_registry_bind(reg, name, &C.wl_compositor_interface, 3))
	case "wl_output":
		output := (*C.struct_wl_output)(C.wl_registry_bind(reg, name, &C.wl_output_interface, 2))
		C.wl_output_add_listener(output, &C.gio_output_listener, unsafe.Pointer(d.disp))
		d.outputMap[name] = output
		d.outputConfig[output] = new(wlOutput)
	case "wl_seat":
		if d.seat != nil {
			break
		}
		s := (*C.struct_wl_seat)(C.wl_registry_bind(reg, name, &C.wl_seat_interface, 5))
		if s == nil {
			// No support for v5 protocol.
			break
		}
		d.seat = &wlSeat{
			disp:      d,
			name:      name,
			seat:      s,
			offers:    make(map[*C.struct_wl_data_offer][]string),
			touchFoci: make(map[C.int32_t]*window),
		}
		callbackStore(unsafe.Pointer(s), d.seat)
		C.wl_seat_add_listener(s, &C.gio_seat_listener, unsafe.Pointer(s))
		d.bindDataDevice()
	case "wl_shm":
		d.shm = (*C.struct_wl_shm)(C.wl_registry_bind(reg, name, &C.wl_shm_interface, 1))
	case "xdg_wm_base":
		d.wm = (*C.struct_xdg_wm_base)(C.wl_registry_bind(reg, name, &C.xdg_wm_base_interface, 1))
	case "zxdg_decoration_manager_v1":
		d.decor = (*C.struct_zxdg_decoration_manager_v1)(C.wl_registry_bind(reg, name, &C.zxdg_decoration_manager_v1_interface, 1))
		// TODO: Implement and test text-input support.
		/*case "zwp_text_input_manager_v3":
		d.imm = (*C.struct_zwp_text_input_manager_v3)(C.wl_registry_bind(reg, name, &C.zwp_text_input_manager_v3_interface, 1))*/
	case "wl_data_device_manager":
		d.dataDeviceManager = (*C.struct_wl_data_device_manager)(C.wl_registry_bind(reg, name, &C.wl_data_device_manager_interface, 3))
		d.bindDataDevice()
	}
}

//export gio_onDataOfferOffer
func gio_onDataOfferOffer(data unsafe.Pointer, offer *C.struct_wl_data_offer, mime *C.char) {
	s := callbackLoad(data).(*wlSeat)
	s.offers[offer] = append(s.offers[offer], C.GoString(mime))
}

//export gio_onDataOfferSourceActions
func gio_onDataOfferSourceActions(data unsafe.Pointer, offer *C.struct_wl_data_offer, acts C.uint32_t) {
}

//export gio_onDataOfferAction
func gio_onDataOfferAction(data unsafe.Pointer, offer *C.struct_wl_data_offer, act C.uint32_t) {
}

//export gio_onDataDeviceOffer
func gio_onDataDeviceOffer(data unsafe.Pointer, dataDev *C.struct_wl_data_device, id *C.struct_wl_data_offer) {
	s := callbackLoad(data).(*wlSeat)
	callbackStore(unsafe.Pointer(id), s)
	C.wl_data_offer_add_listener(id, &C.gio_data_offer_listener, unsafe.Pointer(id))
	s.offers[id] = nil
}

//export gio_onDataDeviceEnter
func gio_onDataDeviceEnter(data unsafe.Pointer, dataDev *C.struct_wl_data_device, serial C.uint32_t, surf *C.struct_wl_surface, x, y C.wl_fixed_t, id *C.struct_wl_data_offer) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	s.flushOffers()
}

//export gio_onDataDeviceLeave
func gio_onDataDeviceLeave(data unsafe.Pointer, dataDev *C.struct_wl_data_device) {
}

//export gio_onDataDeviceMotion
func gio_onDataDeviceMotion(data unsafe.Pointer, dataDev *C.struct_wl_data_device, t C.uint32_t, x, y C.wl_fixed_t) {
}

//export gio_onDataDeviceDrop
func gio_onDataDeviceDrop(data unsafe.Pointer, dataDev *C.struct_wl_data_device) {
}

//export gio_onDataDeviceSelection
func gio_onDataDeviceSelection(data unsafe.Pointer, dataDev *C.struct_wl_data_device, id *C.struct_wl_data_offer) {
	s := callbackLoad(data).(*wlSeat)
	defer s.flushOffers()
	s.clipboard = nil
loop:
	for _, want := range clipboardMimeTypes {
		for _, got := range s.offers[id] {
			if want != got {
				continue
			}
			s.clipboard = id
			s.mimeType = got
			break loop
		}
	}
}

//export gio_onRegistryGlobalRemove
func gio_onRegistryGlobalRemove(data unsafe.Pointer, reg *C.struct_wl_registry, name C.uint32_t) {
	d := callbackLoad(data).(*wlDisplay)
	if s := d.seat; s != nil && name == s.name {
		s.destroy()
		d.seat = nil
	}
	if output, exists := d.outputMap[name]; exists {
		C.wl_output_destroy(output)
		delete(d.outputMap, name)
		delete(d.outputConfig, output)
	}
}

//export gio_onTouchDown
func gio_onTouchDown(data unsafe.Pointer, touch *C.struct_wl_touch, serial, t C.uint32_t, surf *C.struct_wl_surface, id C.int32_t, x, y C.wl_fixed_t) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	s.touchFoci[id] = w
	w.lastTouch = f32.Point{
		X: fromFixed(x) * float32(w.scale),
		Y: fromFixed(y) * float32(w.scale),
	}
	w.w.Event(pointer.Event{
		Type:      pointer.Press,
		Source:    pointer.Touch,
		Position:  w.lastTouch,
		PointerID: pointer.ID(id),
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

//export gio_onTouchUp
func gio_onTouchUp(data unsafe.Pointer, touch *C.struct_wl_touch, serial, t C.uint32_t, id C.int32_t) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := s.touchFoci[id]
	delete(s.touchFoci, id)
	w.w.Event(pointer.Event{
		Type:      pointer.Release,
		Source:    pointer.Touch,
		Position:  w.lastTouch,
		PointerID: pointer.ID(id),
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

//export gio_onTouchMotion
func gio_onTouchMotion(data unsafe.Pointer, touch *C.struct_wl_touch, t C.uint32_t, id C.int32_t, x, y C.wl_fixed_t) {
	s := callbackLoad(data).(*wlSeat)
	w := s.touchFoci[id]
	w.lastTouch = f32.Point{
		X: fromFixed(x) * float32(w.scale),
		Y: fromFixed(y) * float32(w.scale),
	}
	w.w.Event(pointer.Event{
		Type:      pointer.Move,
		Position:  w.lastTouch,
		Source:    pointer.Touch,
		PointerID: pointer.ID(id),
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

//export gio_onTouchFrame
func gio_onTouchFrame(data unsafe.Pointer, touch *C.struct_wl_touch) {
}

//export gio_onTouchCancel
func gio_onTouchCancel(data unsafe.Pointer, touch *C.struct_wl_touch) {
	s := callbackLoad(data).(*wlSeat)
	for id, w := range s.touchFoci {
		delete(s.touchFoci, id)
		w.w.Event(pointer.Event{
			Type:   pointer.Cancel,
			Source: pointer.Touch,
		})
	}
}

//export gio_onPointerEnter
func gio_onPointerEnter(data unsafe.Pointer, pointer *C.struct_wl_pointer, serial C.uint32_t, surf *C.struct_wl_surface, x, y C.wl_fixed_t) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	w.seat = s
	s.pointerFocus = w
	w.setCursor(pointer, serial)
	w.lastPos = f32.Point{X: fromFixed(x), Y: fromFixed(y)}
}

//export gio_onPointerLeave
func gio_onPointerLeave(data unsafe.Pointer, p *C.struct_wl_pointer, serial C.uint32_t, surf *C.struct_wl_surface) {
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	w.seat = nil
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	if w.inCompositor {
		w.inCompositor = false
		w.w.Event(pointer.Event{Type: pointer.Cancel})
	}
}

//export gio_onPointerMotion
func gio_onPointerMotion(data unsafe.Pointer, p *C.struct_wl_pointer, t C.uint32_t, x, y C.wl_fixed_t) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	w.resetFling()
	w.onPointerMotion(x, y, t)
}

//export gio_onPointerButton
func gio_onPointerButton(data unsafe.Pointer, p *C.struct_wl_pointer, serial, t, wbtn, state C.uint32_t) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := s.pointerFocus
	// From linux-event-codes.h.
	const (
		BTN_LEFT   = 0x110
		BTN_RIGHT  = 0x111
		BTN_MIDDLE = 0x112
	)
	var btn pointer.Buttons
	switch wbtn {
	case BTN_LEFT:
		btn = pointer.ButtonPrimary
	case BTN_RIGHT:
		btn = pointer.ButtonSecondary
	case BTN_MIDDLE:
		btn = pointer.ButtonTertiary
	default:
		return
	}
	var typ pointer.Type
	switch state {
	case 0:
		w.pointerBtns &^= btn
		typ = pointer.Release
		// Move or resize gestures no longer applies.
		w.inCompositor = false
	case 1:
		w.pointerBtns |= btn
		typ = pointer.Press
	}
	if typ == pointer.Press && btn == pointer.ButtonPrimary {
		if _, edge := w.systemGesture(); edge != 0 {
			w.resize(serial, edge)
			return
		}
		act, ok := w.w.ActionAt(w.lastPos)
		if ok && w.config.Mode == Windowed {
			switch act {
			case system.ActionMove:
				w.move(serial)
				return
			}
		}
	}
	w.flushScroll()
	w.resetFling()
	w.w.Event(pointer.Event{
		Type:      typ,
		Source:    pointer.Mouse,
		Buttons:   w.pointerBtns,
		Position:  w.lastPos,
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
}

//export gio_onPointerAxis
func gio_onPointerAxis(data unsafe.Pointer, p *C.struct_wl_pointer, t, axis C.uint32_t, value C.wl_fixed_t) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	v := fromFixed(value)
	w.resetFling()
	if w.scroll.dist == (f32.Point{}) {
		w.scroll.time = time.Duration(t) * time.Millisecond
	}
	switch axis {
	case C.WL_POINTER_AXIS_HORIZONTAL_SCROLL:
		w.scroll.dist.X += v
	case C.WL_POINTER_AXIS_VERTICAL_SCROLL:
		w.scroll.dist.Y += v
	}
}

//export gio_onPointerFrame
func gio_onPointerFrame(data unsafe.Pointer, p *C.struct_wl_pointer) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	w.flushScroll()
	w.flushFling()
}

func (w *window) flushFling() {
	if !w.fling.start {
		return
	}
	w.fling.start = false
	estx, esty := w.fling.xExtrapolation.Estimate(), w.fling.yExtrapolation.Estimate()
	w.fling.xExtrapolation = fling.Extrapolation{}
	w.fling.yExtrapolation = fling.Extrapolation{}
	vel := float32(math.Sqrt(float64(estx.Velocity*estx.Velocity + esty.Velocity*esty.Velocity)))
	_, c := w.getConfig()
	if !w.fling.anim.Start(c, time.Now(), vel) {
		return
	}
	invDist := 1 / vel
	w.fling.dir.X = estx.Velocity * invDist
	w.fling.dir.Y = esty.Velocity * invDist
}

//export gio_onPointerAxisSource
func gio_onPointerAxisSource(data unsafe.Pointer, pointer *C.struct_wl_pointer, source C.uint32_t) {
}

//export gio_onPointerAxisStop
func gio_onPointerAxisStop(data unsafe.Pointer, p *C.struct_wl_pointer, t, axis C.uint32_t) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	w.fling.start = true
}

//export gio_onPointerAxisDiscrete
func gio_onPointerAxisDiscrete(data unsafe.Pointer, p *C.struct_wl_pointer, axis C.uint32_t, discrete C.int32_t) {
	s := callbackLoad(data).(*wlSeat)
	w := s.pointerFocus
	w.resetFling()
	switch axis {
	case C.WL_POINTER_AXIS_HORIZONTAL_SCROLL:
		w.scroll.steps.X += int(discrete)
	case C.WL_POINTER_AXIS_VERTICAL_SCROLL:
		w.scroll.steps.Y += int(discrete)
	}
}

func (w *window) ReadClipboard() {
	r, err := w.disp.readClipboard()
	// Send empty responses on unavailable clipboards or errors.
	if r == nil || err != nil {
		w.w.Event(clipboard.Event{})
		return
	}
	// Don't let slow clipboard transfers block event loop.
	go func() {
		defer r.Close()
		data, _ := ioutil.ReadAll(r)
		w.clipReads <- clipboard.Event{Text: string(data)}
		w.Wakeup()
	}()
}

func (w *window) WriteClipboard(s string) {
	w.disp.writeClipboard([]byte(s))
}

func (w *window) Configure(options []Option) {
	_, cfg := w.getConfig()
	prev := w.config
	cnf := w.config
	cnf.apply(cfg, options)
	w.config.decoHeight = cnf.decoHeight

	switch cnf.Mode {
	case Fullscreen:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Fullscreen
			w.wsize = w.config.Size
			C.xdg_toplevel_set_fullscreen(w.topLvl, nil)
		}
	case Minimized:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Minimized
			C.xdg_toplevel_set_minimized(w.topLvl)
		}
	case Maximized:
		switch prev.Mode {
		case Minimized, Fullscreen:
		default:
			w.config.Mode = Maximized
			w.wsize = w.config.Size
			C.xdg_toplevel_set_maximized(w.topLvl)
			w.setTitle(prev, cnf)
		}
	case Windowed:
		switch prev.Mode {
		case Fullscreen:
			w.config.Mode = Windowed
			w.size = w.wsize.Div(w.scale)
			C.xdg_toplevel_unset_fullscreen(w.topLvl)
		case Minimized:
			w.config.Mode = Windowed
		case Maximized:
			w.config.Mode = Windowed
			w.size = w.wsize.Div(w.scale)
			C.xdg_toplevel_unset_maximized(w.topLvl)
		}
		w.setTitle(prev, cnf)
		if prev.Size != cnf.Size {
			w.config.Size = cnf.Size
			w.config.Size.Y += int(w.decoHeight()) * w.scale
			w.size = w.config.Size.Div(w.scale)
		}
		w.config.MinSize = cnf.MinSize
		w.config.MaxSize = cnf.MaxSize
		w.setWindowConstraints()
	}
	w.w.Event(ConfigEvent{Config: w.config})
}

func (w *window) setWindowConstraints() {
	decoHeight := w.decoHeight()
	if scaled := w.config.MinSize.Div(w.scale); scaled != (image.Point{}) {
		C.xdg_toplevel_set_min_size(w.topLvl, C.int32_t(scaled.X), C.int32_t(scaled.Y+decoHeight))
	}
	if scaled := w.config.MaxSize.Div(w.scale); scaled != (image.Point{}) {
		C.xdg_toplevel_set_max_size(w.topLvl, C.int32_t(scaled.X), C.int32_t(scaled.Y+decoHeight))
	}
}

// decoHeight returns the adjustment for client-side decorations, if applicable.
// The unit is in surface-local coordinates.
func (w *window) decoHeight() int {
	if !w.config.Decorated {
		return int(w.config.decoHeight)
	}
	return 0
}

func (w *window) setTitle(prev, cnf Config) {
	if prev.Title != cnf.Title {
		w.config.Title = cnf.Title
		title := C.CString(cnf.Title)
		C.xdg_toplevel_set_title(w.topLvl, title)
		C.free(unsafe.Pointer(title))
	}
}

func (w *window) Perform(actions system.Action) {
	// NB. there is no way for a minimized window to be unminimized.
	// https://wayland.app/protocols/xdg-shell#xdg_toplevel:request:set_minimized
	walkActions(actions, func(action system.Action) {
		switch action {
		case system.ActionClose:
			w.dead = true
		}
	})
}

func (w *window) move(serial C.uint32_t) {
	s := w.seat
	if !w.inCompositor && s != nil {
		w.inCompositor = true
		C.xdg_toplevel_move(w.topLvl, s.seat, serial)
	}
}

func (w *window) resize(serial, edge C.uint32_t) {
	s := w.seat
	if w.inCompositor || s == nil {
		return
	}
	w.inCompositor = true
	C.xdg_toplevel_resize(w.topLvl, s.seat, serial, edge)
}

func (w *window) SetCursor(cursor pointer.Cursor) {
	w.cursor.cursor = w.loadCursor(cursor)
	w.updateCursor()
}

func (w *window) updateCursor() {
	ptr := w.disp.seat.pointer
	if ptr == nil {
		return
	}
	w.setCursor(ptr, w.serial)
}

func (w *window) setCursor(pointer *C.struct_wl_pointer, serial C.uint32_t) {
	c := w.cursor.system
	if c == nil {
		c = w.cursor.cursor
	}
	if c == nil {
		C.wl_pointer_set_cursor(pointer, w.serial, nil, 0, 0)
		return
	}
	// Get images[0].
	img := *c.images
	buf := C.wl_cursor_image_get_buffer(img)
	if buf == nil {
		return
	}
	C.wl_pointer_set_cursor(pointer, serial, w.cursor.surf, C.int32_t(img.hotspot_x/C.uint(w.scale)), C.int32_t(img.hotspot_y/C.uint(w.scale)))
	C.wl_surface_attach(w.cursor.surf, buf, 0, 0)
	C.wl_surface_damage(w.cursor.surf, 0, 0, C.int32_t(img.width), C.int32_t(img.height))
	C.wl_surface_commit(w.cursor.surf)
}

func (w *window) resetFling() {
	w.fling.start = false
	w.fling.anim = fling.Animation{}
}

//export gio_onKeyboardKeymap
func gio_onKeyboardKeymap(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, format C.uint32_t, fd C.int32_t, size C.uint32_t) {
	defer syscall.Close(int(fd))
	s := callbackLoad(data).(*wlSeat)
	s.disp.repeat.Stop(0)
	s.disp.xkb.DestroyKeymapState()
	if format != C.WL_KEYBOARD_KEYMAP_FORMAT_XKB_V1 {
		return
	}
	if err := s.disp.xkb.LoadKeymap(int(format), int(fd), int(size)); err != nil {
		// TODO: Do better.
		panic(err)
	}
}

//export gio_onKeyboardEnter
func gio_onKeyboardEnter(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial C.uint32_t, surf *C.struct_wl_surface, keys *C.struct_wl_array) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := callbackLoad(unsafe.Pointer(surf)).(*window)
	s.keyboardFocus = w
	s.disp.repeat.Stop(0)
	w.w.Event(key.FocusEvent{Focus: true})
}

//export gio_onKeyboardLeave
func gio_onKeyboardLeave(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial C.uint32_t, surf *C.struct_wl_surface) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	s.disp.repeat.Stop(0)
	w := s.keyboardFocus
	w.w.Event(key.FocusEvent{Focus: false})
}

//export gio_onKeyboardKey
func gio_onKeyboardKey(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial, timestamp, keyCode, state C.uint32_t) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	w := s.keyboardFocus
	t := time.Duration(timestamp) * time.Millisecond
	s.disp.repeat.Stop(t)
	w.resetFling()
	kc := mapXKBKeycode(uint32(keyCode))
	ks := mapXKBKeyState(uint32(state))
	for _, e := range w.disp.xkb.DispatchKey(kc, ks) {
		if ee, ok := e.(key.EditEvent); ok {
			// There's no support for IME yet.
			w.w.EditorInsert(ee.Text)
		} else {
			w.w.Event(e)
		}
	}
	if state != C.WL_KEYBOARD_KEY_STATE_PRESSED {
		return
	}
	if w.disp.xkb.IsRepeatKey(kc) {
		w.disp.repeat.Start(w, kc, t)
	}
}

func mapXKBKeycode(keyCode uint32) uint32 {
	// According to the xkb_v1 spec: "to determine the xkb keycode, clients must add 8 to the key event keycode."
	return keyCode + 8
}

func mapXKBKeyState(state uint32) key.State {
	switch state {
	case C.WL_KEYBOARD_KEY_STATE_RELEASED:
		return key.Release
	default:
		return key.Press
	}
}

func (r *repeatState) Start(w *window, keyCode uint32, t time.Duration) {
	if r.rate <= 0 {
		return
	}
	stopC := make(chan struct{})
	r.start = t
	r.last = 0
	r.now = 0
	r.stopC = stopC
	r.key = keyCode
	r.win = w.w
	rate, delay := r.rate, r.delay
	go func() {
		timer := time.NewTimer(delay)
		for {
			select {
			case <-timer.C:
			case <-stopC:
				close(stopC)
				return
			}
			r.Advance(delay)
			w.disp.wakeup()
			delay = time.Second / time.Duration(rate)
			timer.Reset(delay)
		}
	}()
}

func (r *repeatState) Stop(t time.Duration) {
	if r.stopC == nil {
		return
	}
	r.stopC <- struct{}{}
	<-r.stopC
	r.stopC = nil
	t -= r.start
	if r.now > t {
		r.now = t
	}
}

func (r *repeatState) Advance(dt time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.now += dt
}

func (r *repeatState) Repeat(d *wlDisplay) {
	if r.rate <= 0 {
		return
	}
	r.mu.Lock()
	now := r.now
	r.mu.Unlock()
	for {
		var delay time.Duration
		if r.last < r.delay {
			delay = r.delay
		} else {
			delay = time.Second / time.Duration(r.rate)
		}
		if r.last+delay > now {
			break
		}
		for _, e := range d.xkb.DispatchKey(r.key, key.Press) {
			if ee, ok := e.(key.EditEvent); ok {
				// There's no support for IME yet.
				r.win.EditorInsert(ee.Text)
			} else {
				r.win.Event(e)
			}
		}
		r.last += delay
	}
}

//export gio_onFrameDone
func gio_onFrameDone(data unsafe.Pointer, callback *C.struct_wl_callback, t C.uint32_t) {
	C.wl_callback_destroy(callback)
	w := callbackLoad(data).(*window)
	if w.lastFrameCallback == callback {
		w.lastFrameCallback = nil
	}
}

func (w *window) loop() error {
	var p poller
	for {
		if err := w.disp.dispatch(&p); err != nil {
			return err
		}
		select {
		case e := <-w.clipReads:
			w.w.Event(e)
		case <-w.wakeups:
			w.w.Event(wakeupEvent{})
		default:
		}
		if w.dead {
			break
		}
		w.draw()
	}
	return nil
}

// bindDataDevice initializes the dataDev field if and only if both
// the seat and dataDeviceManager fields are initialized.
func (d *wlDisplay) bindDataDevice() {
	if d.seat != nil && d.dataDeviceManager != nil {
		d.seat.dataDev = C.wl_data_device_manager_get_data_device(d.dataDeviceManager, d.seat.seat)
		if d.seat.dataDev == nil {
			return
		}
		callbackStore(unsafe.Pointer(d.seat.dataDev), d.seat)
		C.wl_data_device_add_listener(d.seat.dataDev, &C.gio_data_device_listener, unsafe.Pointer(d.seat.dataDev))
	}
}

func (d *wlDisplay) dispatch(p *poller) error {
	dispfd := C.wl_display_get_fd(d.disp)
	// Poll for events and notifications.
	pollfds := append(p.pollfds[:0],
		syscall.PollFd{Fd: int32(dispfd), Events: syscall.POLLIN | syscall.POLLERR},
		syscall.PollFd{Fd: int32(d.notify.read), Events: syscall.POLLIN | syscall.POLLERR},
	)
	dispFd := &pollfds[0]
	if ret, err := C.wl_display_flush(d.disp); ret < 0 {
		if err != syscall.EAGAIN {
			return fmt.Errorf("wayland: wl_display_flush failed: %v", err)
		}
		// EAGAIN means the output buffer was full. Poll for
		// POLLOUT to know when we can write again.
		dispFd.Events |= syscall.POLLOUT
	}
	if _, err := syscall.Poll(pollfds, -1); err != nil && err != syscall.EINTR {
		return fmt.Errorf("wayland: poll failed: %v", err)
	}
	// Clear notifications.
	for {
		_, err := syscall.Read(d.notify.read, p.buf[:])
		if err == syscall.EAGAIN {
			break
		}
		if err != nil {
			return fmt.Errorf("wayland: read from notify pipe failed: %v", err)
		}
	}
	// Handle events
	switch {
	case dispFd.Revents&syscall.POLLIN != 0:
		if ret, err := C.wl_display_dispatch(d.disp); ret < 0 {
			return fmt.Errorf("wayland: wl_display_dispatch failed: %v", err)
		}
	case dispFd.Revents&(syscall.POLLERR|syscall.POLLHUP) != 0:
		return errors.New("wayland: display file descriptor gone")
	}
	d.repeat.Repeat(d)
	return nil
}

func (w *window) Wakeup() {
	select {
	case w.wakeups <- struct{}{}:
	default:
	}
	w.disp.wakeup()
}

func (w *window) SetAnimating(anim bool) {
	w.animating = anim
}

// Wakeup wakes up the event loop through the notification pipe.
func (d *wlDisplay) wakeup() {
	oneByte := make([]byte, 1)
	if _, err := syscall.Write(d.notify.write, oneByte); err != nil && err != syscall.EAGAIN {
		panic(fmt.Errorf("failed to write to pipe: %v", err))
	}
}

func (w *window) destroy() {
	if w.cursor.surf != nil {
		C.wl_surface_destroy(w.cursor.surf)
	}
	if w.cursor.theme != nil {
		C.wl_cursor_theme_destroy(w.cursor.theme)
	}
	if w.topLvl != nil {
		C.xdg_toplevel_destroy(w.topLvl)
	}
	if w.surf != nil {
		C.wl_surface_destroy(w.surf)
	}
	if w.wmSurf != nil {
		C.xdg_surface_destroy(w.wmSurf)
	}
	if w.decor != nil {
		C.zxdg_toplevel_decoration_v1_destroy(w.decor)
	}
	callbackDelete(unsafe.Pointer(w.surf))
}

//export gio_onKeyboardModifiers
func gio_onKeyboardModifiers(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, serial, depressed, latched, locked, group C.uint32_t) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
	d := s.disp
	d.repeat.Stop(0)
	if d.xkb == nil {
		return
	}
	d.xkb.UpdateMask(uint32(depressed), uint32(latched), uint32(locked), uint32(group), uint32(group), uint32(group))
}

//export gio_onKeyboardRepeatInfo
func gio_onKeyboardRepeatInfo(data unsafe.Pointer, keyboard *C.struct_wl_keyboard, rate, delay C.int32_t) {
	s := callbackLoad(data).(*wlSeat)
	d := s.disp
	d.repeat.Stop(0)
	d.repeat.rate = int(rate)
	d.repeat.delay = time.Duration(delay) * time.Millisecond
}

//export gio_onTextInputEnter
func gio_onTextInputEnter(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, surf *C.struct_wl_surface) {
}

//export gio_onTextInputLeave
func gio_onTextInputLeave(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, surf *C.struct_wl_surface) {
}

//export gio_onTextInputPreeditString
func gio_onTextInputPreeditString(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, ctxt *C.char, begin, end C.int32_t) {
}

//export gio_onTextInputCommitString
func gio_onTextInputCommitString(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, ctxt *C.char) {
}

//export gio_onTextInputDeleteSurroundingText
func gio_onTextInputDeleteSurroundingText(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, before, after C.uint32_t) {
}

//export gio_onTextInputDone
func gio_onTextInputDone(data unsafe.Pointer, im *C.struct_zwp_text_input_v3, serial C.uint32_t) {
	s := callbackLoad(data).(*wlSeat)
	s.serial = serial
}

//export gio_onDataSourceTarget
func gio_onDataSourceTarget(data unsafe.Pointer, source *C.struct_wl_data_source, mime *C.char) {
}

//export gio_onDataSourceSend
func gio_onDataSourceSend(data unsafe.Pointer, source *C.struct_wl_data_source, mime *C.char, fd C.int32_t) {
	s := callbackLoad(data).(*wlSeat)
	content := s.content
	go func() {
		defer syscall.Close(int(fd))
		syscall.Write(int(fd), content)
	}()
}

//export gio_onDataSourceCancelled
func gio_onDataSourceCancelled(data unsafe.Pointer, source *C.struct_wl_data_source) {
	s := callbackLoad(data).(*wlSeat)
	if s.source == source {
		s.content = nil
		s.source = nil
	}
	C.wl_data_source_destroy(source)
}

//export gio_onDataSourceDNDDropPerformed
func gio_onDataSourceDNDDropPerformed(data unsafe.Pointer, source *C.struct_wl_data_source) {
}

//export gio_onDataSourceDNDFinished
func gio_onDataSourceDNDFinished(data unsafe.Pointer, source *C.struct_wl_data_source) {
}

//export gio_onDataSourceAction
func gio_onDataSourceAction(data unsafe.Pointer, source *C.struct_wl_data_source, act C.uint32_t) {
}

func (w *window) flushScroll() {
	var fling f32.Point
	if w.fling.anim.Active() {
		dist := float32(w.fling.anim.Tick(time.Now()))
		fling = w.fling.dir.Mul(dist)
	}
	// The Wayland reported scroll distance for
	// discrete scroll axes is only 10 pixels, where
	// 100 seems more appropriate.
	const discreteScale = 10
	if w.scroll.steps.X != 0 {
		w.scroll.dist.X *= discreteScale
	}
	if w.scroll.steps.Y != 0 {
		w.scroll.dist.Y *= discreteScale
	}
	total := w.scroll.dist.Add(fling)
	if total == (f32.Point{}) {
		return
	}
	w.w.Event(pointer.Event{
		Type:      pointer.Scroll,
		Source:    pointer.Mouse,
		Buttons:   w.pointerBtns,
		Position:  w.lastPos,
		Scroll:    total,
		Time:      w.scroll.time,
		Modifiers: w.disp.xkb.Modifiers(),
	})
	if w.scroll.steps == (image.Point{}) {
		w.fling.xExtrapolation.SampleDelta(w.scroll.time, -w.scroll.dist.X)
		w.fling.yExtrapolation.SampleDelta(w.scroll.time, -w.scroll.dist.Y)
	}
	w.scroll.dist = f32.Point{}
	w.scroll.steps = image.Point{}
}

func (w *window) onPointerMotion(x, y C.wl_fixed_t, t C.uint32_t) {
	w.flushScroll()
	w.lastPos = f32.Point{
		X: fromFixed(x) * float32(w.scale),
		Y: fromFixed(y) * float32(w.scale),
	}
	w.w.Event(pointer.Event{
		Type:      pointer.Move,
		Position:  w.lastPos,
		Buttons:   w.pointerBtns,
		Source:    pointer.Mouse,
		Time:      time.Duration(t) * time.Millisecond,
		Modifiers: w.disp.xkb.Modifiers(),
	})
	c, _ := w.systemGesture()
	if c != w.cursor.system {
		w.cursor.system = c
		w.updateCursor()
	}
}

// updateCursor updates the system gesture cursor according to the pointer
// position.
func (w *window) systemGesture() (*C.struct_wl_cursor, C.uint32_t) {
	if w.config.Mode != Windowed || w.config.Decorated {
		return nil, 0
	}
	border := w.w.w.metric.Dp(3)
	x, y, size := int(w.lastPos.X), int(w.lastPos.Y), w.config.Size
	north := y <= border
	south := y >= size.Y-border
	west := x <= border
	east := x >= size.X-border

	switch {
	default:
		fallthrough
	case !north && !south && !west && !east:
		return nil, 0
	case north && west:
		return w.cursor.cursors.resizeNorthWest, C.XDG_TOPLEVEL_RESIZE_EDGE_TOP_LEFT
	case north && east:
		return w.cursor.cursors.resizeNorthEast, C.XDG_TOPLEVEL_RESIZE_EDGE_TOP_RIGHT
	case south && west:
		return w.cursor.cursors.resizeSouthWest, C.XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_LEFT
	case south && east:
		return w.cursor.cursors.resizeSouthEast, C.XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM_RIGHT
	case north:
		return w.cursor.cursors.resizeNorth, C.XDG_TOPLEVEL_RESIZE_EDGE_TOP
	case south:
		return w.cursor.cursors.resizeSouth, C.XDG_TOPLEVEL_RESIZE_EDGE_BOTTOM
	case west:
		return w.cursor.cursors.resizeWest, C.XDG_TOPLEVEL_RESIZE_EDGE_LEFT
	case east:
		return w.cursor.cursors.resizeEast, C.XDG_TOPLEVEL_RESIZE_EDGE_RIGHT
	}
}

func (w *window) updateOpaqueRegion() {
	reg := C.wl_compositor_create_region(w.disp.compositor)
	C.wl_region_add(reg, 0, 0, C.int32_t(w.size.X), C.int32_t(w.size.Y))
	C.wl_surface_set_opaque_region(w.surf, reg)
	C.wl_region_destroy(reg)
}

func (w *window) updateOutputs() {
	scale := 1
	var found bool
	for _, conf := range w.disp.outputConfig {
		for _, w2 := range conf.windows {
			if w2 == w {
				found = true
				if conf.scale > scale {
					scale = conf.scale
				}
			}
		}
	}
	if found && scale != w.scale {
		w.scale = scale
		C.wl_surface_set_buffer_scale(w.surf, C.int32_t(w.scale))
		w.redraw = true
	}
	if !found {
		w.setStage(system.StagePaused)
	} else {
		w.setStage(system.StageRunning)
		w.redraw = true
	}
}

func (w *window) getConfig() (image.Point, unit.Metric) {
	size := w.size.Mul(w.scale)
	return size, unit.Metric{
		PxPerDp: w.ppdp * float32(w.scale),
		PxPerSp: w.ppsp * float32(w.scale),
	}
}

func (w *window) draw() {
	w.flushScroll()
	size, cfg := w.getConfig()
	if cfg == (unit.Metric{}) {
		return
	}
	if size != w.config.Size {
		w.config.Size = size
		w.w.Event(ConfigEvent{Config: w.config})
	}
	anim := w.animating || w.fling.anim.Active()
	sync := w.redraw
	w.redraw = false
	// Draw animation only when not waiting for frame callback.
	redrawAnim := anim && w.lastFrameCallback == nil
	if !redrawAnim && !sync {
		return
	}
	if anim {
		w.lastFrameCallback = C.wl_surface_frame(w.surf)
		// Use the surface as listener data for gio_onFrameDone.
		C.wl_callback_add_listener(w.lastFrameCallback, &C.gio_callback_listener, unsafe.Pointer(w.surf))
	}
	w.w.Event(frameEvent{
		FrameEvent: system.FrameEvent{
			Now:    time.Now(),
			Size:   w.config.Size,
			Metric: cfg,
		},
		Sync: sync,
	})
}

func (w *window) setStage(s system.Stage) {
	if s == w.stage {
		return
	}
	w.stage = s
	w.w.Event(system.StageEvent{Stage: s})
}

func (w *window) display() *C.struct_wl_display {
	return w.disp.disp
}

func (w *window) surface() (*C.struct_wl_surface, int, int) {
	sz, _ := w.getConfig()
	return w.surf, sz.X, sz.Y
}

func (w *window) ShowTextInput(show bool) {}

func (w *window) SetInputHint(_ key.InputHint) {}

func (w *window) EditorStateChanged(old, new editorState) {}

func (w *window) NewContext() (context, error) {
	var firstErr error
	if f := newWaylandVulkanContext; f != nil {
		c, err := f(w)
		if err == nil {
			return c, nil
		}
		firstErr = err
	}
	if f := newWaylandEGLContext; f != nil {
		c, err := f(w)
		if err == nil {
			return c, nil
		}
		firstErr = err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, errors.New("wayland: no available GPU backends")
}

// detectUIScale reports the system UI scale, or 1.0 if it fails.
func detectUIScale() float32 {
	// TODO: What about other window environments?
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "text-scaling-factor").Output()
	if err != nil {
		return 1.0
	}
	scale, err := strconv.ParseFloat(string(bytes.TrimSpace(out)), 32)
	if err != nil {
		return 1.0
	}
	return float32(scale)
}

func newWLDisplay() (*wlDisplay, error) {
	d := &wlDisplay{
		outputMap:    make(map[C.uint32_t]*C.struct_wl_output),
		outputConfig: make(map[*C.struct_wl_output]*wlOutput),
	}
	pipe := make([]int, 2)
	if err := syscall.Pipe2(pipe, syscall.O_NONBLOCK|syscall.O_CLOEXEC); err != nil {
		return nil, fmt.Errorf("wayland: failed to create pipe: %v", err)
	}
	d.notify.read = pipe[0]
	d.notify.write = pipe[1]
	xkb, err := xkb.New()
	if err != nil {
		d.destroy()
		return nil, fmt.Errorf("wayland: %v", err)
	}
	d.xkb = xkb
	d.disp, err = C.wl_display_connect(nil)
	if d.disp == nil {
		d.destroy()
		return nil, fmt.Errorf("wayland: wl_display_connect failed: %v", err)
	}
	callbackMap.Store(unsafe.Pointer(d.disp), d)
	d.reg = C.wl_display_get_registry(d.disp)
	if d.reg == nil {
		d.destroy()
		return nil, errors.New("wayland: wl_display_get_registry failed")
	}
	C.wl_registry_add_listener(d.reg, &C.gio_registry_listener, unsafe.Pointer(d.disp))
	// Wait for the server to register all its globals to the
	// registry listener (gio_onRegistryGlobal).
	C.wl_display_roundtrip(d.disp)
	// Configuration listeners are added to outputs by gio_onRegistryGlobal.
	// We need another roundtrip to get the initial output configurations
	// through the gio_onOutput* callbacks.
	C.wl_display_roundtrip(d.disp)
	return d, nil
}

func (d *wlDisplay) destroy() {
	if d.notify.write != 0 {
		syscall.Close(d.notify.write)
		d.notify.write = 0
	}
	if d.notify.read != 0 {
		syscall.Close(d.notify.read)
		d.notify.read = 0
	}
	d.repeat.Stop(0)
	if d.xkb != nil {
		d.xkb.Destroy()
		d.xkb = nil
	}
	if d.seat != nil {
		d.seat.destroy()
		d.seat = nil
	}
	if d.imm != nil {
		C.zwp_text_input_manager_v3_destroy(d.imm)
	}
	if d.decor != nil {
		C.zxdg_decoration_manager_v1_destroy(d.decor)
	}
	if d.shm != nil {
		C.wl_shm_destroy(d.shm)
	}
	if d.compositor != nil {
		C.wl_compositor_destroy(d.compositor)
	}
	if d.wm != nil {
		C.xdg_wm_base_destroy(d.wm)
	}
	for _, output := range d.outputMap {
		C.wl_output_destroy(output)
	}
	if d.reg != nil {
		C.wl_registry_destroy(d.reg)
	}
	if d.disp != nil {
		C.wl_display_disconnect(d.disp)
		callbackDelete(unsafe.Pointer(d.disp))
	}
}

// fromFixed converts a Wayland wl_fixed_t 23.8 number to float32.
func fromFixed(v C.wl_fixed_t) float32 {
	// Convert to float64 to avoid overflow.
	// From wayland-util.h.
	b := ((1023 + 44) << 52) + (1 << 51) + uint64(v)
	f := math.Float64frombits(b) - (3 << 43)
	return float32(f)
}
