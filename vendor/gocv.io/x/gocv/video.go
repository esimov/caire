package gocv

/*
#include <stdlib.h>
#include "video.h"
*/
import "C"
import "unsafe"

// BackgroundSubtractorMOG2 is a wrapper around the cv::BackgroundSubtractorMOG2.
type BackgroundSubtractorMOG2 struct {
	// C.BackgroundSubtractorMOG2
	p unsafe.Pointer
}

// NewBackgroundSubtractorMOG2 returns a new BackgroundSubtractor algorithm
// of type MOG2. MOG2 is a Gaussian Mixture-based Background/Foreground
// Segmentation Algorithm.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/d7b/classcv_1_1BackgroundSubtractorMOG2.html
//
func NewBackgroundSubtractorMOG2() BackgroundSubtractorMOG2 {
	return BackgroundSubtractorMOG2{p: unsafe.Pointer(C.BackgroundSubtractorMOG2_Create())}
}

// Close BackgroundSubtractorMOG2.
func (b *BackgroundSubtractorMOG2) Close() error {
	C.BackgroundSubtractorMOG2_Close((C.BackgroundSubtractorMOG2)(b.p))
	b.p = nil
	return nil
}

// Apply computes a foreground mask using the current BackgroundSubtractorMOG2.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/df6/classcv_1_1BackgroundSubtractor.html#aa735e76f7069b3fa9c3f32395f9ccd21
//
func (b *BackgroundSubtractorMOG2) Apply(src Mat, dst Mat) {
	C.BackgroundSubtractorMOG2_Apply((C.BackgroundSubtractorMOG2)(b.p), src.p, dst.p)
	return
}

// BackgroundSubtractorKNN is a wrapper around the cv::BackgroundSubtractorKNN.
type BackgroundSubtractorKNN struct {
	// C.BackgroundSubtractorKNN
	p unsafe.Pointer
}

// NewBackgroundSubtractorKNN returns a new BackgroundSubtractor algorithm
// of type KNN. K-Nearest Neighbors (KNN) uses a Background/Foreground
// Segmentation Algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/db/d88/classcv_1_1BackgroundSubtractorKNN.html
//
func NewBackgroundSubtractorKNN() BackgroundSubtractorKNN {
	return BackgroundSubtractorKNN{p: unsafe.Pointer(C.BackgroundSubtractorKNN_Create())}
}

// Close BackgroundSubtractorKNN.
func (k *BackgroundSubtractorKNN) Close() error {
	C.BackgroundSubtractorKNN_Close((C.BackgroundSubtractorKNN)(k.p))
	k.p = nil
	return nil
}

// Apply computes a foreground mask using the current BackgroundSubtractorKNN.
//
// For further details, please see:
// https://docs.opencv.org/master/d7/df6/classcv_1_1BackgroundSubtractor.html#aa735e76f7069b3fa9c3f32395f9ccd21
//
func (k *BackgroundSubtractorKNN) Apply(src Mat, dst Mat) {
	C.BackgroundSubtractorKNN_Apply((C.BackgroundSubtractorKNN)(k.p), src.p, dst.p)
	return
}

// CalcOpticalFlowFarneback computes a dense optical flow using
// Gunnar Farneback's algorithm.
//
// For further details, please see:
// https://docs.opencv.org/master/dc/d6b/group__video__track.html#ga5d10ebbd59fe09c5f650289ec0ece5af
//
func CalcOpticalFlowFarneback(prevImg Mat, nextImg Mat, flow Mat, pyrScale float64, levels int, winsize int,
	iterations int, polyN int, polySigma float64, flags int) {
	C.CalcOpticalFlowFarneback(prevImg.p, nextImg.p, flow.p, C.double(pyrScale), C.int(levels), C.int(winsize),
		C.int(iterations), C.int(polyN), C.double(polySigma), C.int(flags))
	return
}

// CalcOpticalFlowPyrLK calculates an optical flow for a sparse feature set using
// the iterative Lucas-Kanade method with pyramids.
//
// For further details, please see:
// https://docs.opencv.org/master/dc/d6b/group__video__track.html#ga473e4b886d0bcc6b65831eb88ed93323
//
func CalcOpticalFlowPyrLK(prevImg Mat, nextImg Mat, prevPts Mat, nextPts Mat, status Mat, err Mat) {
	C.CalcOpticalFlowPyrLK(prevImg.p, nextImg.p, prevPts.p, nextPts.p, status.p, err.p)
	return
}
