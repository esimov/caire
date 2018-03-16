package contrib

/*
#include <stdlib.h>
#include "face.h"
*/
import "C"
import (
	"unsafe"

	"gocv.io/x/gocv"
)

type PredictResponse struct {
	Label      int32   `json:"label"`
	Confidence float32 `json:"confidence"`
}

type LBPHFaceRecognizer struct {
	p C.LBPHFaceRecognizer
}

// Create new LBPH Recognizer model
//
// see https://docs.opencv.org/master/df/d25/classcv_1_1face_1_1LBPHFaceRecognizer.html
//
func NewLBPHFaceRecognizer() *LBPHFaceRecognizer {
	return &LBPHFaceRecognizer{p: C.CreateLBPHFaceRecognizer()}
}

// Train loaded model with images and their labels
//
// see https://docs.opencv.org/master/dd/d65/classcv_1_1face_1_1FaceRecognizer.html#ac8680c2aa9649ad3f55e27761165c0d6
//
func (fr *LBPHFaceRecognizer) Train(images []gocv.Mat, labels []int) {
	cparams := []C.int{}
	for _, v := range labels {
		cparams = append(cparams, C.int(v))
	}
	labelsVector := C.struct_IntVector{}
	labelsVector.val = (*C.int)(&cparams[0])
	labelsVector.length = (C.int)(len(cparams))

	cMatArray := make([]C.Mat, len(images))
	for i, r := range images {
		cMatArray[i] = (C.Mat)(r.Ptr())
	}
	matsVector := C.struct_Mats{
		mats:   (*C.Mat)(&cMatArray[0]),
		length: C.int(len(images)),
	}

	C.LBPHFaceRecognizer_Train(fr.p, matsVector, labelsVector)
}

// update existing trained model with new images and labels
//
// see https://docs.opencv.org/master/dd/d65/classcv_1_1face_1_1FaceRecognizer.html#a8a4e73ea878dcd0c235d0487189d25f3
//
func (fr *LBPHFaceRecognizer) Update(newImages []gocv.Mat, newLabels []int) {
	cparams := []C.int{}
	for _, v := range newLabels {
		cparams = append(cparams, C.int(v))
	}
	labelsVector := C.struct_IntVector{}
	labelsVector.val = (*C.int)(&cparams[0])
	labelsVector.length = (C.int)(len(cparams))

	cMatArray := make([]C.Mat, len(newImages))
	for i, r := range newImages {
		cMatArray[i] = (C.Mat)(r.Ptr())
	}
	matsVector := C.struct_Mats{
		mats:   (*C.Mat)(&cMatArray[0]),
		length: C.int(len(newImages)),
	}

	C.LBPHFaceRecognizer_Update(fr.p, matsVector, labelsVector)
}

// predict image for trained model, retun label for correctly predicted image, return -1 if not found
//
// see https://docs.opencv.org/master/dd/d65/classcv_1_1face_1_1FaceRecognizer.html#aa2d2f02faffab1bf01317ae6502fb631
//
func (fr *LBPHFaceRecognizer) Predict(sample gocv.Mat) int {
	label := C.LBPHFaceRecognizer_Predict(fr.p, (C.Mat)(sample.Ptr()))

	return int(label)
}

// the same as above but returns some more info
//
// see https://docs.opencv.org/master/dd/d65/classcv_1_1face_1_1FaceRecognizer.html#ab0d593e53ebd9a0f350c989fcac7f251
//
func (fr *LBPHFaceRecognizer) PredictExtendedResponse(sample gocv.Mat) PredictResponse {
	respp := C.LBPHFaceRecognizer_PredictExtended(fr.p, (C.Mat)(sample.Ptr()))
	resp := PredictResponse{
		Label:      int32(respp.label),
		Confidence: float32(respp.confidence),
	}

	return resp
}

// set Threshold value
//
// see https://docs.opencv.org/master/dd/d65/classcv_1_1face_1_1FaceRecognizer.html#a3182081e5f8023e658ad8ab96656dd63
//
func (fr *LBPHFaceRecognizer) SetThreshold(threshold float32) {
	C.LBPHFaceRecognizer_SetThreshold(fr.p, (C.double)(threshold))
}

// set Neighbors
//
// see https://docs.opencv.org/master/df/d25/classcv_1_1face_1_1LBPHFaceRecognizer.html#ab225f7bf353ce8697a506eda10124a92
// wrong neighbors can raise opencv exception!
//
func (fr *LBPHFaceRecognizer) SetNeighbors(neighbors int) {
	C.LBPHFaceRecognizer_SetNeighbors(fr.p, (C.int)(neighbors))
}

// get Neighbors
//
// see https://docs.opencv.org/master/df/d25/classcv_1_1face_1_1LBPHFaceRecognizer.html#a50a3e2ca6e8464166e153c9df84b0a77
//
func (fr *LBPHFaceRecognizer) GetNeighbors() int {
	n := C.LBPHFaceRecognizer_GetNeighbors(fr.p)

	return int(n)
}

// set Radius
//
// see https://docs.opencv.org/master/df/d25/classcv_1_1face_1_1LBPHFaceRecognizer.html#a62d94c75cade902fd3b487b1ef9883fc
//
func (fr *LBPHFaceRecognizer) SetRadius(radius int) {
	C.LBPHFaceRecognizer_SetRadius(fr.p, (C.int)(radius))
}

// save trained model data to file
//
// see https://docs.opencv.org/master/dd/d65/classcv_1_1face_1_1FaceRecognizer.html#a2adf2d555550194244b05c91fefcb4d6
//
func (fr *LBPHFaceRecognizer) SaveFile(fname string) {
	cName := C.CString(fname)
	defer C.free(unsafe.Pointer(cName))
	C.LBPHFaceRecognizer_SaveFile(fr.p, cName)
}

// load traned model data from file
//
// see https://docs.opencv.org/master/dd/d65/classcv_1_1face_1_1FaceRecognizer.html#acc42e5b04595dba71f0777c7179af8c3
//
func (fr *LBPHFaceRecognizer) LoadFile(fname string) {
	cName := C.CString(fname)
	defer C.free(unsafe.Pointer(cName))
	C.LBPHFaceRecognizer_LoadFile(fr.p, cName)
}
