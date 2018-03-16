package gocv

/*
#include <stdlib.h>
#include "dnn.h"
*/
import "C"
import (
	"image"
	"unsafe"
)

// Net allows you to create and manipulate comprehensive artificial neural networks.
//
// For further details, please see:
// https://docs.opencv.org/master/db/d30/classcv_1_1dnn_1_1Net.html
//
type Net struct {
	// C.Net
	p unsafe.Pointer
}

// Close Net
func (net *Net) Close() error {
	C.Net_Close((C.Net)(net.p))
	net.p = nil
	return nil
}

// Empty returns true if there are no layers in the network.
//
// For further details, please see:
// https://docs.opencv.org/master/db/d30/classcv_1_1dnn_1_1Net.html#a6a5778787d5b8770deab5eda6968e66c
//
func (net *Net) Empty() bool {
	return bool(C.Net_Empty((C.Net)(net.p)))
}

// SetInput sets the new value for the layer output blob.
//
// For further details, please see:
// https://docs.opencv.org/trunk/db/d30/classcv_1_1dnn_1_1Net.html#a672a08ae76444d75d05d7bfea3e4a328
//
func (net *Net) SetInput(blob Mat, name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.Net_SetInput((C.Net)(net.p), blob.p, cName)
}

// Forward runs forward pass to compute output of layer with name outputName.
//
// For further details, please see:
// https://docs.opencv.org/trunk/db/d30/classcv_1_1dnn_1_1Net.html#a98ed94cb6ef7063d3697259566da310b
//
func (net *Net) Forward(outputName string) Mat {
	cName := C.CString(outputName)
	defer C.free(unsafe.Pointer(cName))

	return Mat{p: C.Net_Forward((C.Net)(net.p), cName)}
}

// ReadNetFromCaffe reads a network model stored in Caffe framework's format.
//
// For further details, please see:
// https://docs.opencv.org/master/d6/d0f/group__dnn.html#ga946b342af1355185a7107640f868b64a
//
func ReadNetFromCaffe(prototxt string, caffeModel string) Net {
	cprototxt := C.CString(prototxt)
	defer C.free(unsafe.Pointer(cprototxt))

	cmodel := C.CString(caffeModel)
	defer C.free(unsafe.Pointer(cmodel))
	return Net{p: unsafe.Pointer(C.Net_ReadNetFromCaffe(cprototxt, cmodel))}
}

// ReadNetFromTensorflow reads a network model stored in Tensorflow framework's format.
//
// For further details, please see:
// https://docs.opencv.org/master/d6/d0f/group__dnn.html#gad820b280978d06773234ba6841e77e8d
//
func ReadNetFromTensorflow(model string) Net {
	cmodel := C.CString(model)
	defer C.free(unsafe.Pointer(cmodel))
	return Net{p: unsafe.Pointer(C.Net_ReadNetFromTensorflow(cmodel))}
}

// BlobFromImage creates 4-dimensional blob from image. Optionally resizes and crops
// image from center, subtract mean values, scales values by scalefactor,
// swap Blue and Red channels.
//
// For further details, please see:
// https://docs.opencv.org/trunk/d6/d0f/group__dnn.html#ga152367f253c81b53fe6862b299f5c5cd
//
func BlobFromImage(img Mat, scaleFactor float64, size image.Point, mean Scalar,
	swapRB bool, crop bool) Mat {

	sz := C.struct_Size{
		width:  C.int(size.X),
		height: C.int(size.Y),
	}

	sMean := C.struct_Scalar{
		val1: C.double(mean.Val1),
		val2: C.double(mean.Val2),
		val3: C.double(mean.Val3),
		val4: C.double(mean.Val4),
	}

	return Mat{p: C.Net_BlobFromImage(img.p, C.double(scaleFactor), sz, sMean, C.bool(swapRB), C.bool(crop))}
}
