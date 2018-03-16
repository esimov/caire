package gocv

/*
#include <stdlib.h>
#include "highgui.h"
*/
import "C"
import (
	"image"
	"runtime"
	"unsafe"
)

// Window is a wrapper around OpenCV's "HighGUI" named windows.
// While OpenCV was designed for use in full-scale applications and can be used
// within functionally rich UI frameworks (such as Qt*, WinForms*, or Cocoa*)
// or without any UI at all, sometimes there it is required to try functionality
// quickly and visualize the results. This is what the HighGUI module has been designed for.
//
// For further details, please see:
// http://docs.opencv.org/master/d7/dfc/group__highgui.html
//
type Window struct {
	name string
	open bool
}

// NewWindow creates a new named OpenCV window
//
// For further details, please see:
// http://docs.opencv.org/master/d7/dfc/group__highgui.html#ga5afdf8410934fd099df85c75b2e0888b
//
func NewWindow(name string) *Window {
	runtime.LockOSThread()

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.Window_New(cName, 2)

	return &Window{name: name, open: true}
}

// Close closes and deletes a named OpenCV Window.
//
// For further details, please see:
// http://docs.opencv.org/master/d7/dfc/group__highgui.html#ga851ccdd6961022d1d5b4c4f255dbab34
//
func (w *Window) Close() error {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	C.Window_Close(cName)
	w.open = false

	runtime.UnlockOSThread()
	return nil
}

// IsOpen checks to see if the Window seems to be open.
func (w *Window) IsOpen() bool {
	return w.open
}

// WindowFlag value for SetWindowProperty / GetWindowProperty.
type WindowFlag float32

const (
	// WindowNormal indicates a normal window.
	WindowNormal WindowFlag = 0

	// WindowFullscreen indicates a full-screen window.
	WindowFullscreen = 1

	// WindowAutosize indicates a window sized based on the contents.
	WindowAutosize = 1

	// WindowFreeRatio indicates allow the user to resize without maintaining aspect ratio.
	WindowFreeRatio = 0x00000100

	// WindowKeepRatio indicates always maintain an aspect ratio that matches the contents.
	WindowKeepRatio = 0
)

// WindowPropertyFlag flags for SetWindowProperty / GetWindowProperty.
type WindowPropertyFlag int

const (
	// WindowPropertyFullscreen fullscreen property
	// (can be WINDOW_NORMAL or WINDOW_FULLSCREEN).
	WindowPropertyFullscreen WindowPropertyFlag = 0

	// WindowPropertyAutosize is autosize property
	// (can be WINDOW_NORMAL or WINDOW_AUTOSIZE).
	WindowPropertyAutosize = 1

	// WindowPropertyAspectRatio window's aspect ration
	// (can be set to WINDOW_FREERATIO or WINDOW_KEEPRATIO).
	WindowPropertyAspectRatio = 2

	// WindowPropertyOpenGL opengl support.
	WindowPropertyOpenGL = 3

	// WindowPropertyVisible or not.
	WindowPropertyVisible = 4
)

// GetWindowProperty returns properties of a window.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#gaaf9504b8f9cf19024d9d44a14e461656
//
func (w *Window) GetWindowProperty(flag WindowPropertyFlag) float64 {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	return float64(C.Window_GetProperty(cName, C.int(flag)))
}

// SetWindowProperty changes parameters of a window dynamically.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga66e4a6db4d4e06148bcdfe0d70a5df27
//
func (w *Window) SetWindowProperty(flag WindowPropertyFlag, value WindowFlag) {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	C.Window_SetProperty(cName, C.int(flag), C.double(value))
}

// SetWindowTitle updates window title.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga56f8849295fd10d0c319724ddb773d96
//
func (w *Window) SetWindowTitle(title string) {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	C.Window_SetTitle(cName, cTitle)
}

// IMShow displays an image Mat in the specified window.
// This function should be followed by the WaitKey function which displays
// the image for specified milliseconds. Otherwise, it won't display the image.
//
// For further details, please see:
// http://docs.opencv.org/master/d7/dfc/group__highgui.html#ga453d42fe4cb60e5723281a89973ee563
//
func (w *Window) IMShow(img Mat) {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	C.Window_IMShow(cName, img.p)
}

// WaitKey waits for a pressed key.
// This function is the only method in OpenCV's HighGUI that can fetch
// and handle events, so it needs to be called periodically
// for normal event processing
//
// For further details, please see:
// http://docs.opencv.org/master/d7/dfc/group__highgui.html#ga5628525ad33f52eab17feebcfba38bd7
//
func (w *Window) WaitKey(delay int) int {
	return int(C.Window_WaitKey(C.int(delay)))
}

// MoveWindow moves window to the specified position.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga8d86b207f7211250dbe6e28f76307ffb
//
func (w *Window) MoveWindow(x, y int) {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	C.Window_Move(cName, C.int(x), C.int(y))
}

// ResizeWindow resizes window to the specified size.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga9e80e080f7ef33f897e415358aee7f7e
//
func (w *Window) ResizeWindow(width, height int) {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	C.Window_Resize(cName, C.int(width), C.int(height))
}

// SelectROI selects a Region Of Interest (ROI) on the given image.
// It creates a window and allows user to select a ROI using mouse.
//
// Controls:
// use space or enter to finish selection,
// use key c to cancel selection (function will return a zero Rect).
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga8daf4730d3adf7035b6de9be4c469af5
//
func SelectROI(name string, img Mat) image.Rectangle {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	r := C.Window_SelectROI(cName, img.p)
	rect := image.Rect(int(r.x), int(r.y), int(r.x+r.width), int(r.y+r.height))
	return rect
}

// SelectROIs selects multiple Regions Of Interest (ROI) on the given image.
// It creates a window and allows user to select ROIs using mouse.
//
// Controls:
// use space or enter to finish current selection and start a new one
// use esc to terminate multiple ROI selection process
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga0f11fad74a6432b8055fb21621a0f893
//
func SelectROIs(name string, img Mat) []image.Rectangle {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.Window_SelectROIs(cName, img.p)
	defer C.Rects_Close(ret)

	return toRectangles(ret)
}

// WaitKey that is not attached to a specific Window.
// Only use when no Window exists in your application, e.g. command line app.
//
func WaitKey(delay int) int {
	return int(C.Window_WaitKey(C.int(delay)))
}

// Trackbar is a wrapper around OpenCV's "HighGUI" window Trackbars.
type Trackbar struct {
	name   string
	parent *Window
}

// CreateTrackbar creates a trackbar and attaches it to the specified window.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#gaf78d2155d30b728fc413803745b67a9b
//
func (w *Window) CreateTrackbar(name string, max int) *Trackbar {
	cName := C.CString(w.name)
	defer C.free(unsafe.Pointer(cName))

	tName := C.CString(name)
	defer C.free(unsafe.Pointer(tName))

	C.Trackbar_Create(cName, tName, C.int(max))
	return &Trackbar{name: name, parent: w}
}

// GetPos returns the trackbar position.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga122632e9e91b9ec06943472c55d9cda8
//
func (t *Trackbar) GetPos() int {
	cName := C.CString(t.parent.name)
	defer C.free(unsafe.Pointer(cName))

	tName := C.CString(t.name)
	defer C.free(unsafe.Pointer(tName))

	return int(C.Trackbar_GetPos(cName, tName))
}

// SetPos sets the trackbar position.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga67d73c4c9430f13481fd58410d01bd8d
//
func (t *Trackbar) SetPos(pos int) {
	cName := C.CString(t.parent.name)
	defer C.free(unsafe.Pointer(cName))

	tName := C.CString(t.name)
	defer C.free(unsafe.Pointer(tName))

	C.Trackbar_SetPos(cName, tName, C.int(pos))
}

// SetMin sets the trackbar minimum position.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#gabe26ffe8d2b60cc678895595a581b7aa
//
func (t *Trackbar) SetMin(pos int) {
	cName := C.CString(t.parent.name)
	defer C.free(unsafe.Pointer(cName))

	tName := C.CString(t.name)
	defer C.free(unsafe.Pointer(tName))

	C.Trackbar_SetMin(cName, tName, C.int(pos))
}

// SetMax sets the trackbar maximum position.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/dfc/group__highgui.html#ga7e5437ccba37f1154b65210902fc4480
//
func (t *Trackbar) SetMax(pos int) {
	cName := C.CString(t.parent.name)
	defer C.free(unsafe.Pointer(cName))

	tName := C.CString(t.name)
	defer C.free(unsafe.Pointer(tName))

	C.Trackbar_SetMax(cName, tName, C.int(pos))
}
