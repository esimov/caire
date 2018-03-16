package gocv

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestVideoCaptureFile(t *testing.T) {
	vc, _ := VideoCaptureFile("images/small.mp4")
	defer vc.Close()

	if !vc.IsOpened() {
		t.Error("Unable to open VideoCaptureFile")
	}

	if fw := vc.Get(VideoCaptureFrameWidth); int(fw) != 560 {
		t.Errorf("Expected frame width property of 560.0 got %f", fw)
	}
	if fh := vc.Get(VideoCaptureFrameHeight); int(fh) != 320 {
		t.Errorf("Expected frame height property of 320.0 got %f", fh)
	}

	vc.Set(VideoCaptureBrightness, 100.0)

	vc.Grab(10)

	img := NewMat()
	defer img.Close()

	vc.Read(img)
	if img.Empty() {
		t.Error("Unable to read VideoCaptureFile")
	}
}

func TestVideoWriterFile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "gocvtests")
	tmpfn := filepath.Join(dir, "test.avi")

	img := IMRead("images/face-detect.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid read of Mat in VideoWriterFile test")
	}
	defer img.Close()

	vw, _ := VideoWriterFile(tmpfn, "MJPG", 25, img.Cols(), img.Rows())
	defer vw.Close()

	if !vw.IsOpened() {
		t.Error("Unable to open VideoWriterFile")
	}

	err := vw.Write(img)
	if err != nil {
		t.Error("Invalid Write() in VideoWriter")
	}
}
