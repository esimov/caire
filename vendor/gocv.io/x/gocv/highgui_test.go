package gocv

import (
	"testing"
)

func TestWindow(t *testing.T) {
	window := NewWindow("test")
	if window == nil {
		t.Error("Unable to create Window")
	}
	if window.name != "test" {
		t.Error("Invalid Window name")
	}
	val := window.WaitKey(1)
	if val != -1 {
		t.Error("Invalid WaitKey")
	}
	if !window.IsOpen() {
		t.Error("Window should have been open")
	}

	window.SetWindowProperty(WindowPropertyFullscreen, WindowFullscreen)

	prop := int(window.GetWindowProperty(WindowPropertyFullscreen))
	if prop != WindowFullscreen {
		t.Error("Window property should have been fullscreen")
	}

	window.SetWindowTitle("My new title")

	window.MoveWindow(100, 100)

	window.ResizeWindow(100, 100)

	window.Close()
	if window.IsOpen() {
		t.Error("Window should have been closed")
	}
}

func TestIMShow(t *testing.T) {
	window := NewWindow("imshow")
	if window == nil {
		t.Error("Unable to create IMShow Window")
	}

	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid Mat in IMShow")
	}
	defer img.Close()

	// TODO: some way to determine if the call succeeded
	window.IMShow(img)

	val := WaitKey(1)
	if val != -1 {
		t.Error("Invalid for IMShow")
	}

	window.Close()
	if window.IsOpen() {
		t.Error("IMShow window should have been closed")
	}
}

func TestSelectROI(t *testing.T) {
	t.Skip("TODO: figure out how to implement a test that can exercise the GUI")
}

func TestSelectROIs(t *testing.T) {
	t.Skip("TODO: figure out how to implement a test that can exercise the GUI")
}

func TestTrackbar(t *testing.T) {
	window := NewWindow("trackbar")
	defer window.Close()

	tracker := window.CreateTrackbar("trackme", 100)
	if tracker.GetPos() != 0 {
		t.Error("Trackbar pos should have been 0")
	}

	tracker.SetMin(10)
	tracker.SetMax(150)
	tracker.SetPos(50)

	if tracker.GetPos() != 50 {
		t.Error("Trackbar pos should have been 50")
	}
}
