package gocv

/*
#include <stdlib.h>
#include "videoio.h"
*/
import "C"
import (
	"sync"
	"unsafe"
)

// VideoCaptureProperties are the properties used for VideoCapture operations.
type VideoCaptureProperties int

const (
	// VideoCapturePosMsec contains current position of the
	// video file in milliseconds.
	VideoCapturePosMsec VideoCaptureProperties = 0

	// VideoCapturePosFrames 0-based index of the frame to be
	// decoded/captured next.
	VideoCapturePosFrames = 1

	// VideoCapturePosAVIRatio relative position of the video file:
	// 0=start of the film, 1=end of the film.
	VideoCapturePosAVIRatio = 2

	// VideoCaptureFrameWidth is width of the frames in the video stream.
	VideoCaptureFrameWidth = 3

	// VideoCaptureFrameHeight controls height of frames in the video stream.
	VideoCaptureFrameHeight = 4

	// VideoCaptureFPS controls capture frame rate.
	VideoCaptureFPS = 5

	// VideoCaptureFOURCC contains the 4-character code of codec.
	// see VideoWriter::fourcc for details.
	VideoCaptureFOURCC = 6

	// VideoCaptureFrameCount contains number of frames in the video file.
	VideoCaptureFrameCount = 7

	// VideoCaptureFormat format of the Mat objects returned by
	// VideoCapture::retrieve().
	VideoCaptureFormat = 8

	// VideoCaptureMode contains backend-specific value indicating
	// the current capture mode.
	VideoCaptureMode = 9

	// VideoCaptureBrightness is brightness of the image
	// (only for those cameras that support).
	VideoCaptureBrightness = 10

	// VideoCaptureContrast is contrast of the image
	// (only for cameras that support it).
	VideoCaptureContrast = 11

	// VideoCaptureSaturation saturation of the image
	// (only for cameras that support).
	VideoCaptureSaturation = 12

	// VideoCaptureHue hue of the image (only for cameras that support).
	VideoCaptureHue = 13

	// VideoCaptureGain is the gain of the capture image.
	// (only for those cameras that support).
	VideoCaptureGain = 14

	// VideoCaptureExposure is the exposure of the capture image.
	// (only for those cameras that support).
	VideoCaptureExposure = 15

	// VideoCaptureConvertRGB is a boolean flags indicating whether
	// images should be converted to RGB.
	VideoCaptureConvertRGB = 16

	// VideoCaptureWhiteBalanceBlueU is currently unsupported.
	VideoCaptureWhiteBalanceBlueU = 17

	// VideoCaptureRectification is the rectification flag for stereo cameras.
	// Note: only supported by DC1394 v 2.x backend currently.
	VideoCaptureRectification = 18

	// VideoCaptureMonochrome indicates whether images should be
	// converted to monochrome.
	VideoCaptureMonochrome = 19

	// VideoCaptureSharpness controls image capture sharpness.
	VideoCaptureSharpness = 20

	// VideoCaptureAutoExposure controls the DC1394 exposure control
	// done by camera, user can adjust reference level using this feature.
	VideoCaptureAutoExposure = 21

	// VideoCaptureGamma controls video capture gamma.
	VideoCaptureGamma = 22

	// VideoCaptureTemperature controls video capture temperature.
	VideoCaptureTemperature = 23

	// VideoCaptureTrigger controls video capture trigger.
	VideoCaptureTrigger = 24

	// VideoCaptureTriggerDelay controls video capture trigger delay.
	VideoCaptureTriggerDelay = 25

	// VideoCaptureWhiteBalanceRedV controls video capture setting for
	// white balance.
	VideoCaptureWhiteBalanceRedV = 26

	// VideoCaptureZoom controls video capture zoom.
	VideoCaptureZoom = 27

	// VideoCaptureFocus controls video capture focus.
	VideoCaptureFocus = 28

	// VideoCaptureGUID controls video capture GUID.
	VideoCaptureGUID = 29

	// VideoCaptureISOSpeed controls video capture ISO speed.
	VideoCaptureISOSpeed = 30

	// VideoCaptureBacklight controls video capture backlight.
	VideoCaptureBacklight = 32

	// VideoCapturePan controls video capture pan.
	VideoCapturePan = 33

	// VideoCaptureTilt controls video capture tilt.
	VideoCaptureTilt = 34

	// VideoCaptureRoll controls video capture roll.
	VideoCaptureRoll = 35

	// VideoCaptureIris controls video capture iris.
	VideoCaptureIris = 36

	// VideoCaptureSettings is the pop up video/camera filter dialog. Note:
	// only supported by DSHOW backend currently. The property value is ignored.
	VideoCaptureSettings = 37

	// VideoCaptureBufferSize controls video capture buffer size.
	VideoCaptureBufferSize = 38

	// VideoCaptureAutoFocus controls video capture auto focus..
	VideoCaptureAutoFocus = 39
)

// VideoCapture is a wrapper around the OpenCV VideoCapture class.
//
// For further details, please see:
// http://docs.opencv.org/master/d8/dfe/classcv_1_1VideoCapture.html
//
type VideoCapture struct {
	p C.VideoCapture
}

// VideoCaptureFile opens a VideoCapture from a file and prepares
// to start capturing.
func VideoCaptureFile(uri string) (vc *VideoCapture, err error) {
	vc = &VideoCapture{p: C.VideoCapture_New()}

	cURI := C.CString(uri)
	defer C.free(unsafe.Pointer(cURI))

	C.VideoCapture_Open(vc.p, cURI)
	return
}

// VideoCaptureDevice opens a VideoCapture from a device and prepares
// to start capturing.
func VideoCaptureDevice(device int) (vc *VideoCapture, err error) {
	vc = &VideoCapture{p: C.VideoCapture_New()}
	C.VideoCapture_OpenDevice(vc.p, C.int(device))
	return
}

// Close VideoCapture object.
func (v *VideoCapture) Close() error {
	C.VideoCapture_Close(v.p)
	v.p = nil
	return nil
}

// Set parameter with property (=key).
func (v *VideoCapture) Set(prop VideoCaptureProperties, param float64) {
	C.VideoCapture_Set(v.p, C.int(prop), C.double(param))
}

// Get parameter with property (=key).
func (v VideoCapture) Get(prop VideoCaptureProperties) float64 {
	return float64(C.VideoCapture_Get(v.p, C.int(prop)))
}

// IsOpened returns if the VideoCapture has been opened to read from
// a file or capture device.
func (v *VideoCapture) IsOpened() bool {
	isOpened := C.VideoCapture_IsOpened(v.p)
	return isOpened != 0
}

// Read read the next frame from the VideoCapture to the Mat passed in
// as the parem. It returns false if the VideoCapture cannot read frame.
func (v *VideoCapture) Read(m Mat) bool {
	return C.VideoCapture_Read(v.p, m.p) != 0
}

// Grab skips a specific number of frames.
func (v *VideoCapture) Grab(skip int) {
	C.VideoCapture_Grab(v.p, C.int(skip))
}

// VideoWriter is a wrapper around the OpenCV VideoWriter`class.
//
// For further details, please see:
// http://docs.opencv.org/master/dd/d9e/classcv_1_1VideoWriter.html
//
type VideoWriter struct {
	mu *sync.RWMutex
	p  C.VideoWriter
}

// VideoWriterFile opens a VideoWriter with a specific output file.
// The "codec" param should be the four-letter code for the desired output
// codec, for example "MJPG".
//
// For further details, please see:
// http://docs.opencv.org/master/dd/d9e/classcv_1_1VideoWriter.html#a0901c353cd5ea05bba455317dab81130
//
func VideoWriterFile(name string, codec string, fps float64, width int, height int) (vw *VideoWriter, err error) {
	vw = &VideoWriter{
		p:  C.VideoWriter_New(),
		mu: &sync.RWMutex{},
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cCodec := C.CString(codec)
	defer C.free(unsafe.Pointer(cCodec))

	C.VideoWriter_Open(vw.p, cName, cCodec, C.double(fps), C.int(width), C.int(height))
	return
}

// Close VideoWriter object.
func (vw *VideoWriter) Close() error {
	C.VideoWriter_Close(vw.p)
	vw.p = nil
	return nil
}

// IsOpened checks if the VideoWriter is open and ready to be written to.
//
// For further details, please see:
// http://docs.opencv.org/master/dd/d9e/classcv_1_1VideoWriter.html#a9a40803e5f671968ac9efa877c984d75
//
func (vw *VideoWriter) IsOpened() bool {
	isOpend := C.VideoWriter_IsOpened(vw.p)
	return isOpend != 0
}

// Write the next video frame from the Mat image to the open VideoWriter.
//
// For further details, please see:
// http://docs.opencv.org/master/dd/d9e/classcv_1_1VideoWriter.html#a3115b679d612a6a0b5864a0c88ed4b39
//
func (vw *VideoWriter) Write(img Mat) error {
	vw.mu.Lock()
	defer vw.mu.Unlock()
	C.VideoWriter_Write(vw.p, img.p)
	return nil
}
