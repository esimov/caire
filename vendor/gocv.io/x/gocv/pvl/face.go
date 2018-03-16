package pvl

/*
#include <stdlib.h>
#include "face.h"
*/
import "C"

import (
	"image"
)

// Face is a wrapper around cv::pvl::Face.
type Face struct {
	p C.Face
}

// NewFace returns a new PVL Face.
func NewFace() Face {
	return Face{p: C.Face_New()}
}

// Close Face.
func (f *Face) Close() error {
	C.Face_Close(f.p)
	f.p = nil
	return nil
}

// Ptr returns the Face's underlying object pointer.
func (f *Face) Ptr() C.Face {
	return f.p
}

// Rectangle returns the image.Rectangle for this Face.
func (f *Face) Rectangle() image.Rectangle {
	r := C.Face_GetRect(f.p)
	return image.Rect(int(r.x), int(r.y), int(r.x+r.width), int(r.y+r.height))
}

// RIPAngle of Face.
func (f *Face) RIPAngle() int {
	return int(C.Face_RIPAngle(f.p))
}

// ROPAngle of Face.
func (f *Face) ROPAngle() int {
	return int(C.Face_ROPAngle(f.p))
}

// LeftEyePosition of Face.
func (f *Face) LeftEyePosition() image.Point {
	pt := C.Face_LeftEyePosition(f.p)
	return image.Pt(int(pt.y), int(pt.y))
}

// IsLeftEyeClosed checks if the right sys is closed or not.
func (f *Face) IsLeftEyeClosed() bool {
	return bool(C.Face_LeftEyeClosed(f.p))
}

// RightEyePosition of Face.
func (f *Face) RightEyePosition() image.Point {
	pt := C.Face_RightEyePosition(f.p)
	return image.Pt(int(pt.y), int(pt.y))
}

// IsRightEyeClosed checks if the right sys is closed or not.
func (f *Face) IsRightEyeClosed() bool {
	return bool(C.Face_RightEyeClosed(f.p))
}

// IsSmiling Face? :)
// You must call FaceDetector's DetectEye() and DetectSmile() with this Face
// first, or this function will throw an exception.
func (f *Face) IsSmiling() bool {
	return bool(C.Face_IsSmiling(f.p))
}

// MouthPosition of Face.
func (f *Face) MouthPosition() image.Point {
	pt := C.Face_MouthPosition(f.p)
	return image.Pt(int(pt.y), int(pt.y))
}
