package gocv

/*
#include <stdlib.h>
#include "features2d.h"
*/
import "C"
import (
	"reflect"
	"unsafe"
)

// AKAZE is a wrapper around the cv::AKAZE algorithm.
type AKAZE struct {
	// C.AKAZE
	p unsafe.Pointer
}

// NewAKAZE returns a new AKAZE algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/d8/d30/classcv_1_1AKAZE.html
//
func NewAKAZE() AKAZE {
	return AKAZE{p: unsafe.Pointer(C.AKAZE_Create())}
}

// Close AKAZE.
func (a *AKAZE) Close() error {
	C.AKAZE_Close((C.AKAZE)(a.p))
	a.p = nil
	return nil
}

// Detect keypoints in an image using AKAZE.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (a *AKAZE) Detect(src Mat) []KeyPoint {
	ret := C.AKAZE_Detect((C.AKAZE)(a.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// DetectAndCompute keypoints and compute in an image using AKAZE.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#a8be0d1c20b08eb867184b8d74c15a677
//
func (a *AKAZE) DetectAndCompute(src Mat, mask Mat) ([]KeyPoint, Mat) {
	desc := NewMat()
	ret := C.AKAZE_DetectAndCompute((C.AKAZE)(a.p), src.p, mask.p, desc.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret), desc
}

// AgastFeatureDetector is a wrapper around the cv::AgastFeatureDetector.
type AgastFeatureDetector struct {
	// C.AgastFeatureDetector
	p unsafe.Pointer
}

// NewAgastFeatureDetector returns a new AgastFeatureDetector algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/d7/d19/classcv_1_1AgastFeatureDetector.html
//
func NewAgastFeatureDetector() AgastFeatureDetector {
	return AgastFeatureDetector{p: unsafe.Pointer(C.AgastFeatureDetector_Create())}
}

// Close AgastFeatureDetector.
func (a *AgastFeatureDetector) Close() error {
	C.AgastFeatureDetector_Close((C.AgastFeatureDetector)(a.p))
	a.p = nil
	return nil
}

// Detect keypoints in an image using AgastFeatureDetector.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (a *AgastFeatureDetector) Detect(src Mat) []KeyPoint {
	ret := C.AgastFeatureDetector_Detect((C.AgastFeatureDetector)(a.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// BRISK is a wrapper around the cv::BRISK algorithm.
type BRISK struct {
	// C.BRISK
	p unsafe.Pointer
}

// NewBRISK returns a new BRISK algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/d8/d30/classcv_1_1AKAZE.html
//
func NewBRISK() BRISK {
	return BRISK{p: unsafe.Pointer(C.BRISK_Create())}
}

// Close BRISK.
func (b *BRISK) Close() error {
	C.BRISK_Close((C.BRISK)(b.p))
	b.p = nil
	return nil
}

// Detect keypoints in an image using BRISK.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (b *BRISK) Detect(src Mat) []KeyPoint {
	ret := C.BRISK_Detect((C.BRISK)(b.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// DetectAndCompute keypoints and compute in an image using BRISK.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#a8be0d1c20b08eb867184b8d74c15a677
//
func (b *BRISK) DetectAndCompute(src Mat, mask Mat) ([]KeyPoint, Mat) {
	desc := NewMat()
	ret := C.BRISK_DetectAndCompute((C.BRISK)(b.p), src.p, mask.p, desc.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret), desc
}

// FastFeatureDetector is a wrapper around the cv::FastFeatureDetector.
type FastFeatureDetector struct {
	// C.FastFeatureDetector
	p unsafe.Pointer
}

// NewFastFeatureDetector returns a new FastFeatureDetector algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/df/d74/classcv_1_1FastFeatureDetector.html
//
func NewFastFeatureDetector() FastFeatureDetector {
	return FastFeatureDetector{p: unsafe.Pointer(C.FastFeatureDetector_Create())}
}

// Close FastFeatureDetector.
func (f *FastFeatureDetector) Close() error {
	C.FastFeatureDetector_Close((C.FastFeatureDetector)(f.p))
	f.p = nil
	return nil
}

// Detect keypoints in an image using FastFeatureDetector.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (f *FastFeatureDetector) Detect(src Mat) []KeyPoint {
	ret := C.FastFeatureDetector_Detect((C.FastFeatureDetector)(f.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// GFTTDetector is a wrapper around the cv::GFTTDetector algorithm.
type GFTTDetector struct {
	// C.GFTTDetector
	p unsafe.Pointer
}

// NewGFTTDetector returns a new GFTTDetector algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/df/d21/classcv_1_1GFTTDetector.html
//
func NewGFTTDetector() GFTTDetector {
	return GFTTDetector{p: unsafe.Pointer(C.GFTTDetector_Create())}
}

// Close GFTTDetector.
func (a *GFTTDetector) Close() error {
	C.GFTTDetector_Close((C.GFTTDetector)(a.p))
	a.p = nil
	return nil
}

// Detect keypoints in an image using GFTTDetector.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (a *GFTTDetector) Detect(src Mat) []KeyPoint {
	ret := C.GFTTDetector_Detect((C.GFTTDetector)(a.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// KAZE is a wrapper around the cv::KAZE algorithm.
type KAZE struct {
	// C.KAZE
	p unsafe.Pointer
}

// NewKAZE returns a new KAZE algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/d3/d61/classcv_1_1KAZE.html
//
func NewKAZE() KAZE {
	return KAZE{p: unsafe.Pointer(C.KAZE_Create())}
}

// Close KAZE.
func (a *KAZE) Close() error {
	C.KAZE_Close((C.KAZE)(a.p))
	a.p = nil
	return nil
}

// Detect keypoints in an image using KAZE.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (a *KAZE) Detect(src Mat) []KeyPoint {
	ret := C.KAZE_Detect((C.KAZE)(a.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// DetectAndCompute keypoints and compute in an image using KAZE.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#a8be0d1c20b08eb867184b8d74c15a677
//
func (a *KAZE) DetectAndCompute(src Mat, mask Mat) ([]KeyPoint, Mat) {
	desc := NewMat()
	ret := C.KAZE_DetectAndCompute((C.KAZE)(a.p), src.p, mask.p, desc.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret), desc
}

// MSER is a wrapper around the cv::MSER algorithm.
type MSER struct {
	// C.MSER
	p unsafe.Pointer
}

// NewMSER returns a new MSER algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/d3/d28/classcv_1_1MSER.html
//
func NewMSER() MSER {
	return MSER{p: unsafe.Pointer(C.MSER_Create())}
}

// Close MSER.
func (a *MSER) Close() error {
	C.MSER_Close((C.MSER)(a.p))
	a.p = nil
	return nil
}

// Detect keypoints in an image using MSER.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (a *MSER) Detect(src Mat) []KeyPoint {
	ret := C.MSER_Detect((C.MSER)(a.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// ORB is a wrapper around the cv::ORB.
type ORB struct {
	// C.ORB
	p unsafe.Pointer
}

// NewORB returns a new ORB algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/d7/d19/classcv_1_1AgastFeatureDetector.html
//
func NewORB() ORB {
	return ORB{p: unsafe.Pointer(C.ORB_Create())}
}

// Close ORB.
func (o *ORB) Close() error {
	C.ORB_Close((C.ORB)(o.p))
	o.p = nil
	return nil
}

// Detect keypoints in an image using ORB.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (o *ORB) Detect(src Mat) []KeyPoint {
	ret := C.ORB_Detect((C.ORB)(o.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// DetectAndCompute detects keypoints and computes from an image using ORB.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#a8be0d1c20b08eb867184b8d74c15a677
//
func (o *ORB) DetectAndCompute(src Mat, mask Mat) ([]KeyPoint, Mat) {
	desc := NewMat()
	ret := C.ORB_DetectAndCompute((C.ORB)(o.p), src.p, mask.p, desc.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret), desc
}

// SimpleBlobDetector is a wrapper around the cv::SimpleBlobDetector.
type SimpleBlobDetector struct {
	// C.SimpleBlobDetector
	p unsafe.Pointer
}

// NewSimpleBlobDetector returns a new SimpleBlobDetector algorithm
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d7a/classcv_1_1SimpleBlobDetector.html
//
func NewSimpleBlobDetector() SimpleBlobDetector {
	return SimpleBlobDetector{p: unsafe.Pointer(C.SimpleBlobDetector_Create())}
}

// Close SimpleBlobDetector.
func (b *SimpleBlobDetector) Close() error {
	C.SimpleBlobDetector_Close((C.SimpleBlobDetector)(b.p))
	b.p = nil
	return nil
}

// Detect keypoints in an image using SimpleBlobDetector.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (b *SimpleBlobDetector) Detect(src Mat) []KeyPoint {
	ret := C.SimpleBlobDetector_Detect((C.SimpleBlobDetector)(b.p), src.p)
	defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

func getKeyPoints(ret C.KeyPoints) []KeyPoint {
	cArray := ret.keypoints
	length := int(ret.length)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cArray)),
		Len:  length,
		Cap:  length,
	}
	s := *(*[]C.KeyPoint)(unsafe.Pointer(&hdr))

	keys := make([]KeyPoint, length)
	for i, r := range s {
		keys[i] = KeyPoint{float64(r.x), float64(r.y), float64(r.size), float64(r.angle), float64(r.response),
			int(r.octave), int(r.classID)}
	}
	return keys
}
