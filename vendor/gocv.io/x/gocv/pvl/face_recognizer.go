package pvl

/*
#include <stdlib.h>
#include "face_recognizer.h"
*/
import "C"

import (
	"unsafe"

	"gocv.io/x/gocv"
)

// UnknownFace is when the FaceRecognizer cannot identify a Face.
const UnknownFace = -10000

// FaceRecognizer is a wrapper around the cv::pvl::FaceRecognizer.
type FaceRecognizer struct {
	// C.FaceRecognizer
	p unsafe.Pointer
}

// NewFaceRecognizer returns a new PVL FaceRecognizer.
func NewFaceRecognizer() FaceRecognizer {
	return FaceRecognizer{p: unsafe.Pointer(C.FaceRecognizer_New())}
}

// Close FaceRecognizer.
func (f *FaceRecognizer) Close() error {
	C.FaceRecognizer_Close((C.FaceRecognizer)(f.p))
	f.p = nil
	return nil
}

// Clear FaceRecognizer data.
func (f *FaceRecognizer) Clear() {
	C.FaceRecognizer_Clear((C.FaceRecognizer)(f.p))
}

// Empty checks if FaceRecognizer has no data.
func (f *FaceRecognizer) Empty() bool {
	return bool(C.FaceRecognizer_Empty((C.FaceRecognizer)(f.p)))
}

// SetTrackingModeEnabled sets if the PVL FaceRecognizer tracking mode is enabled.
func (f *FaceRecognizer) SetTrackingModeEnabled(enabled bool) {
	C.FaceRecognizer_SetTrackingModeEnabled((C.FaceRecognizer)(f.p), C.bool(enabled))
}

// CreateNewPersonID gets the next available ID from the PVL FaceRecognizer to be added to the database.
func (f *FaceRecognizer) CreateNewPersonID() int {
	return int(C.FaceRecognizer_CreateNewPersonID((C.FaceRecognizer)(f.p)))
}

// GetNumRegisteredPersons gets the number of people in the current database.
func (f *FaceRecognizer) GetNumRegisteredPersons() int {
	return int(C.FaceRecognizer_GetNumRegisteredPersons((C.FaceRecognizer)(f.p)))
}

// Recognize recognizes faces with the given image and face information.
func (f *FaceRecognizer) Recognize(img gocv.Mat, faces []Face) (personIDs, confidences []int) {
	cFaceArray := make([]C.Face, len(faces))
	for i, r := range faces {
		cFaceArray[i] = r.p
	}
	cFaces := C.struct_Faces{
		faces:  (*C.Face)(&cFaceArray[0]),
		length: C.int(len(faces)),
	}

	pids := C.IntVector{}
	confs := C.IntVector{}

	C.FaceRecognizer_Recognize((C.FaceRecognizer)(f.p), C.Mat(img.Ptr()), cFaces, &pids, &confs)

	aPids := (*[1 << 30]C.int)(unsafe.Pointer(pids.val))
	for i := 0; i < int(pids.length); i++ {
		personIDs = append(personIDs, int(aPids[i]))
	}

	aConfs := (*[1 << 30]C.int)(unsafe.Pointer(confs.val))
	for i := 0; i < int(confs.length); i++ {
		confidences = append(confidences, int(aConfs[i]))
	}
	return
}

// RegisterFace registers face information into the internal database.
func (f *FaceRecognizer) RegisterFace(img gocv.Mat, face Face, personID int, saveToFile bool) int64 {
	return int64(C.FaceRecognizer_RegisterFace((C.FaceRecognizer)(f.p), C.Mat(img.Ptr()), C.Face(face.Ptr()),
		C.int(personID), C.bool(saveToFile)))
}

// DeregisterFace deregisters the previously registered face from the internal database.
func (f *FaceRecognizer) DeregisterFace(faceID int64) {
	C.FaceRecognizer_DeregisterFace((C.FaceRecognizer)(f.p), C.int64_t(faceID))
}

// DeregisterPerson deregisters the previously registered person from the internal database.
func (f *FaceRecognizer) DeregisterPerson(personID int) {
	C.FaceRecognizer_DeregisterPerson((C.FaceRecognizer)(f.p), C.int(personID))
}

// LoadFaceRecognizer loads data from file and returns a FaceRecognizer.
func LoadFaceRecognizer(name string) FaceRecognizer {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	return FaceRecognizer{p: unsafe.Pointer(C.FaceRecognizer_Load(cName))}
}

// Save FaceRecognizer data to file.
func (f *FaceRecognizer) Save(name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.FaceRecognizer_Save((C.FaceRecognizer)(f.p), cName)
}
