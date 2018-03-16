package contrib

/*
#include <stdlib.h>
#include "xfeatures2d.h"
*/
import "C"

import (
	"reflect"
	"unsafe"

	"gocv.io/x/gocv"
)

// SIFT is a wrapper around the cv::SIFT algorithm.
type SIFT struct {
	// C.SIFT
	p unsafe.Pointer
}

// NewSIFT returns a new SIFT algorithm.
//
// For further details, please see:
// https://docs.opencv.org/master/d5/d3c/classcv_1_1xfeatures2d_1_1SIFT.html
//
func NewSIFT() SIFT {
	return SIFT{p: unsafe.Pointer(C.SIFT_Create())}
}

// Close SIFT.
func (d *SIFT) Close() error {
	C.SIFT_Close((C.SIFT)(d.p))
	d.p = nil
	return nil
}

// Detect keypoints in an image using SIFT.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (d *SIFT) Detect(src gocv.Mat) []gocv.KeyPoint {
	ret := C.SIFT_Detect((C.SIFT)(d.p), C.Mat(src.Ptr()))
	//defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// DetectAndCompute detects and computes keypoints in an image using SIFT.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#a8be0d1c20b08eb867184b8d74c15a677
//
func (d *SIFT) DetectAndCompute(src gocv.Mat, mask gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	desc := gocv.NewMat()
	ret := C.SIFT_DetectAndCompute((C.SIFT)(d.p), C.Mat(src.Ptr()), C.Mat(mask.Ptr()),
		C.Mat(desc.Ptr()))
	//defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret), desc
}

// SURF is a wrapper around the cv::SURF algorithm.
type SURF struct {
	// C.SURF
	p unsafe.Pointer
}

// NewSURF returns a new SURF algorithm.
//
// For further details, please see:
// https://docs.opencv.org/master/d5/df7/classcv_1_1xfeatures2d_1_1SURF.html
//
func NewSURF() SURF {
	return SURF{p: unsafe.Pointer(C.SURF_Create())}
}

// Close SURF.
func (d *SURF) Close() error {
	C.SURF_Close((C.SURF)(d.p))
	d.p = nil
	return nil
}

// Detect keypoints in an image using SURF.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#aa4e9a7082ec61ebc108806704fbd7887
//
func (d *SURF) Detect(src gocv.Mat) []gocv.KeyPoint {
	ret := C.SURF_Detect((C.SURF)(d.p), C.Mat(src.Ptr()))
	//defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret)
}

// DetectAndCompute detects and computes keypoints in an image using SURF.
//
// For further details, please see:
// https://docs.opencv.org/master/d0/d13/classcv_1_1Feature2D.html#a8be0d1c20b08eb867184b8d74c15a677
//
func (d *SURF) DetectAndCompute(src gocv.Mat, mask gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	desc := gocv.NewMat()
	ret := C.SURF_DetectAndCompute((C.SURF)(d.p), C.Mat(src.Ptr()), C.Mat(mask.Ptr()),
		C.Mat(desc.Ptr()))
	//defer C.KeyPoints_Close(ret)

	return getKeyPoints(ret), desc
}

func getKeyPoints(ret C.KeyPoints) []gocv.KeyPoint {
	cArray := ret.keypoints
	length := int(ret.length)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cArray)),
		Len:  length,
		Cap:  length,
	}
	s := *(*[]C.KeyPoint)(unsafe.Pointer(&hdr))

	keys := make([]gocv.KeyPoint, length)
	for i, r := range s {
		keys[i] = gocv.KeyPoint{X: float64(r.x), Y: float64(r.y), Size: float64(r.size), Angle: float64(r.angle),
			Response: float64(r.response), Octave: int(r.octave), ClassID: int(r.classID)}
	}
	return keys
}
