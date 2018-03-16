package gocv

import (
	"image"
	"testing"
)

func TestMOG2(t *testing.T) {
	img := IMRead("images/face.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid Mat in MOG2 test")
	}
	defer img.Close()

	dst := NewMat()
	defer dst.Close()

	mog2 := NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	mog2.Apply(img, dst)

	if dst.Empty() {
		t.Error("Error in TestMOG2 test")
	}
}

func TestKNN(t *testing.T) {
	img := IMRead("images/face.jpg", IMReadColor)
	if img.Empty() {
		t.Error("Invalid Mat in KNN test")
	}
	defer img.Close()

	dst := NewMat()
	defer dst.Close()

	knn := NewBackgroundSubtractorKNN()
	defer knn.Close()

	knn.Apply(img, dst)

	if dst.Empty() {
		t.Error("Error in TestKNN test")
	}
}

func TestCalcOpticalFlowFarneback(t *testing.T) {
	img1 := IMRead("images/face.jpg", IMReadColor)
	if img1.Empty() {
		t.Error("Invalid Mat in CalcOpticalFlowFarneback test")
	}
	defer img1.Close()

	dest := NewMat()
	defer dest.Close()

	CvtColor(img1, dest, ColorBGRAToGray)

	img2 := dest.Clone()

	flow := NewMat()
	defer flow.Close()

	CalcOpticalFlowFarneback(dest, img2, flow, 0.4, 1, 12, 2, 8, 1.2, 0)

	if flow.Empty() {
		t.Error("Error in CalcOpticalFlowFarneback test")
	}
	if flow.Rows() != 480 {
		t.Errorf("Invalid CalcOpticalFlowFarneback test rows: %v", flow.Rows())
	}
	if flow.Cols() != 640 {
		t.Errorf("Invalid CalcOpticalFlowFarneback test cols: %v", flow.Cols())
	}
}

func TestCalcOpticalFlowPyrLK(t *testing.T) {
	img1 := IMRead("images/face.jpg", IMReadColor)
	if img1.Empty() {
		t.Error("Invalid Mat in CalcOpticalFlowPyrLK test")
	}
	defer img1.Close()

	dest := NewMat()
	defer dest.Close()

	CvtColor(img1, dest, ColorBGRAToGray)

	img2 := dest.Clone()

	prevPts := NewMat()
	defer prevPts.Close()

	nextPts := NewMat()
	defer nextPts.Close()

	status := NewMat()
	defer status.Close()

	err := NewMat()
	defer err.Close()

	corners := NewMat()
	defer corners.Close()

	GoodFeaturesToTrack(dest, corners, 500, 0.01, 10)
	tc := NewTermCriteria(Count|EPS, 20, 0.03)
	CornerSubPix(dest, corners, image.Pt(10, 10), image.Pt(-1, -1), tc)

	CalcOpticalFlowPyrLK(dest, img2, corners, nextPts, status, err)

	if status.Empty() {
		t.Error("Error in CalcOpticalFlowPyrLK test")
	}
	if status.Rows() != 323 {
		t.Errorf("Invalid CalcOpticalFlowPyrLK test rows: %v", status.Rows())
	}
	if status.Cols() != 1 {
		t.Errorf("Invalid CalcOpticalFlowPyrLK test cols: %v", status.Cols())
	}
}
