package gocv

import (
	"image"
	"testing"
)

func TestCascadeClassifier(t *testing.T) {
	img := IMRead("images/face.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid Mat in CascadeClassifier test")
	}
	defer img.Close()

	// load classifier to recognize faces
	classifier := NewCascadeClassifier()
	defer classifier.Close()

	classifier.Load("data/haarcascade_frontalface_default.xml")

	rects := classifier.DetectMultiScale(img)
	if len(rects) != 1 {
		t.Error("Error in TestCascadeClassifier test")
	}
}

func TestCascadeClassifierWithParams(t *testing.T) {
	img := IMRead("images/face.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid Mat in CascadeClassifierWithParams test")
	}
	defer img.Close()

	// load classifier to recognize faces
	classifier := NewCascadeClassifier()
	defer classifier.Close()

	classifier.Load("data/haarcascade_frontalface_default.xml")

	rects := classifier.DetectMultiScaleWithParams(img, 1.1, 3, 0, image.Pt(0, 0), image.Pt(0, 0))
	if len(rects) != 1 {
		t.Errorf("Error in CascadeClassifierWithParams test: %v", len(rects))
	}
}

func TestHOGDescriptor(t *testing.T) {
	img := IMRead("images/face.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid Mat in HOGDescriptor test")
	}
	defer img.Close()

	// load HOGDescriptor to recognize people
	hog := NewHOGDescriptor()
	defer hog.Close()

	hog.SetSVMDetector(HOGDefaultPeopleDetector())

	rects := hog.DetectMultiScale(img)
	if len(rects) != 1 {
		t.Errorf("Error in TestHOGDescriptor test: %d", len(rects))
	}
}

func TestHOGDescriptorWithParams(t *testing.T) {
	img := IMRead("images/face.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid Mat in HOGDescriptorWithParams test")
	}
	defer img.Close()

	// load HOGDescriptor to recognize people
	hog := NewHOGDescriptor()
	defer hog.Close()

	hog.SetSVMDetector(HOGDefaultPeopleDetector())

	rects := hog.DetectMultiScaleWithParams(img, 0, image.Pt(0, 0), image.Pt(0, 0),
		1.05, 2.0, false)
	if len(rects) != 1 {
		t.Errorf("Error in TestHOGDescriptorWithParams test: %d", len(rects))
	}
}

func TestGroupRectangles(t *testing.T) {
	rects := []image.Rectangle{
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 30, 30),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
		image.Rect(10, 10, 35, 35),
	}

	results := GroupRectangles(rects, 1, 0.2)
	if len(results) != 2 {
		t.Errorf("Error in TestGroupRectangles test: %d", len(results))
	}
}
