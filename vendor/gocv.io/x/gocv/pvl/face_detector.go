package pvl

/*
#include <stdlib.h>
#include "face_detector.h"
*/
import "C"

import (
	"reflect"
	"unsafe"

	"gocv.io/x/gocv"
)

// FaceDetector is a wrapper around the cv::pvl::FaceDetector.
type FaceDetector struct {
	// C.FaceDetector
	p unsafe.Pointer
}

// NewFaceDetector returns a new PVL FaceDetector.
func NewFaceDetector() FaceDetector {
	return FaceDetector{p: unsafe.Pointer(C.FaceDetector_New())}
}

// Close FaceDetector.
func (f *FaceDetector) Close() error {
	C.FaceDetector_Close((C.FaceDetector)(f.p))
	f.p = nil
	return nil
}

// GetRIPAngleRange returns RIP(Rotation-In-Plane) angle range for face detection.
func (f *FaceDetector) GetRIPAngleRange() int {
	return int(C.FaceDetector_GetRIPAngleRange((C.FaceDetector)(f.p)))
}

// SetRIPAngleRange sets RIP(Rotation-In-Plane) angle range for face detection.
// Rotated faces within this angle range can be detected when detect method is invoked.
// If you specify small value for the range, Detection takes lesser time since it doesn't need to find
// much rotated faces. Default value is 135.
func (f *FaceDetector) SetRIPAngleRange(rip int) {
	C.FaceDetector_SetRIPAngleRange((C.FaceDetector)(f.p), C.int(rip))
}

// GetROPAngleRange returns ROP(Rotation-Out-Of-Plane) angle range for face detection.
func (f *FaceDetector) GetROPAngleRange() int {
	return int(C.FaceDetector_GetROPAngleRange((C.FaceDetector)(f.p)))
}

// SetROPAngleRange sets ROP(Rotation-Out-Of-Plane) angle range for face detection.
// Rotated faces within this angle range can be detected when detect method is invoked.
// If you specify small value for the range, Detection takes lesser time since it doesn't need to find
// much rotated faces. Default value is 90.
func (f *FaceDetector) SetROPAngleRange(rop int) {
	C.FaceDetector_SetROPAngleRange((C.FaceDetector)(f.p), C.int(rop))
}

// GetMaxDetectableFaces Returns the maximum number of detected faces.
func (f *FaceDetector) GetMaxDetectableFaces() int {
	return int(C.FaceDetector_GetMaxDetectableFaces((C.FaceDetector)(f.p)))
}

// SetMaxDetectableFaces sets the maximum number of detected faces.
func (f *FaceDetector) SetMaxDetectableFaces(max int) {
	C.FaceDetector_SetMaxDetectableFaces((C.FaceDetector)(f.p), C.int(max))
}

// GetMinFaceSize gets the minimum face size in pixel.
func (f *FaceDetector) GetMinFaceSize() int {
	return int(C.FaceDetector_GetMinFaceSize((C.FaceDetector)(f.p)))
}

// SetMinFaceSize sets the minimum face size in pixel.
func (f *FaceDetector) SetMinFaceSize(min int) {
	C.FaceDetector_SetMinFaceSize((C.FaceDetector)(f.p), C.int(min))
}

// GetBlinkThreshold gets the threshold value used for evaluating blink.
func (f *FaceDetector) GetBlinkThreshold() int {
	return int(C.FaceDetector_GetBlinkThreshold((C.FaceDetector)(f.p)))
}

// SetBlinkThreshold sets the threshold value used for evaluating blink.
// When the blink score is equal or greater than this threshold, the eye is considered closing.
// Default value is 50.
func (f *FaceDetector) SetBlinkThreshold(thresh int) {
	C.FaceDetector_SetBlinkThreshold((C.FaceDetector)(f.p), C.int(thresh))
}

// GetSmileThreshold gets the threshold value used for evaluating smiles.
func (f *FaceDetector) GetSmileThreshold() int {
	return int(C.FaceDetector_GetSmileThreshold((C.FaceDetector)(f.p)))
}

// SetSmileThreshold sets the threshold value used for evaluating smiles.
// When the blink score is equal or greater than this threshold, the eye is considered smiling.
// Default value is 48.
func (f *FaceDetector) SetSmileThreshold(thresh int) {
	C.FaceDetector_SetSmileThreshold((C.FaceDetector)(f.p), C.int(thresh))
}

// SetTrackingModeEnabled sets if the PVL FaceDetector tracking mode is enabled.
func (f *FaceDetector) SetTrackingModeEnabled(enabled bool) {
	C.FaceDetector_SetTrackingModeEnabled((C.FaceDetector)(f.p), C.bool(enabled))
}

// IsTrackingModeEnabled checks if the PVL FaceDetector tracking mode is enabled.
func (f *FaceDetector) IsTrackingModeEnabled() bool {
	return bool(C.FaceDetector_IsTrackingModeEnabled((C.FaceDetector)(f.p)))
}

// DetectFaceRect tries to detect Faces from the image Mat passed in as the param.
// The Mat must be a grayed image that has only one channel and 8-bit depth.
func (f *FaceDetector) DetectFaceRect(img gocv.Mat) []Face {
	ret := C.FaceDetector_DetectFaceRect((C.FaceDetector)(f.p), C.Mat(img.Ptr()))
	fArray := ret.faces
	length := int(ret.length)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(fArray)),
		Len:  length,
		Cap:  length,
	}
	s := *(*[]C.Face)(unsafe.Pointer(&hdr))

	faces := make([]Face, length)
	for i, r := range s {
		faces[i] = Face{p: r}
	}
	return faces
}

// DetectEye uses PVL FaceDetector to detect eyes on a Face.
func (f *FaceDetector) DetectEye(img gocv.Mat, face Face) {
	C.FaceDetector_DetectEye((C.FaceDetector)(f.p), C.Mat(img.Ptr()), C.Face(face.Ptr()))
	return
}

// DetectMouth uses PVL FaceDetector to detect mouth on a Face.
func (f *FaceDetector) DetectMouth(img gocv.Mat, face Face) {
	C.FaceDetector_DetectMouth((C.FaceDetector)(f.p), C.Mat(img.Ptr()), C.Face(face.Ptr()))
	return
}

// DetectSmile uses PVL FaceDetector to detect smile on a Face.
func (f *FaceDetector) DetectSmile(img gocv.Mat, face Face) {
	C.FaceDetector_DetectSmile((C.FaceDetector)(f.p), C.Mat(img.Ptr()), C.Face(face.Ptr()))
	return
}

// DetectBlink uses PVL FaceDetector to detect blink on a Face.
func (f *FaceDetector) DetectBlink(img gocv.Mat, face Face) {
	C.FaceDetector_DetectBlink((C.FaceDetector)(f.p), C.Mat(img.Ptr()), C.Face(face.Ptr()))
	return
}
