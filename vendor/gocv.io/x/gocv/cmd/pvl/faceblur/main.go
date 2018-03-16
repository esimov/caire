// What it does:
//
// This example uses the Intel CV SDK PVL FaceDetect to detect faces,
// then blurs them using a Gaussian blur before displaying in a window.
//
// How to run:
//
// faceblur [camera ID]
//
// 		go run ./cmd/pvl/faceblur/main.go 0
//
// +build example

package main

import (
	"fmt"
	"image"
	"os"
	"strconv"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/pvl"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\tfaceblur [camera ID]")
		return
	}

	// parse args
	deviceID, _ := strconv.Atoi(os.Args[1])

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("PVL Faceblur")
	defer window.Close()

	// prepare input image matrix
	img := gocv.NewMat()
	defer img.Close()

	// prepare grayscale image matrix
	imgGray := gocv.NewMat()
	defer imgGray.Close()

	// load PVL FaceDetector to recognize faces
	fd := pvl.NewFaceDetector()
	defer fd.Close()

	// enable tracking mode for more efficient tracking of video source
	fd.SetTrackingModeEnabled(true)

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// convert image to grayscale for detection
		gocv.CvtColor(img, imgGray, gocv.ColorBGRAToGray)

		// detect faces
		faces := fd.DetectFaceRect(imgGray)
		fmt.Printf("found %d faces\n", len(faces))

		// blur each face on the original image
		for _, face := range faces {
			imgFace := img.Region(face.Rectangle())

			// blur face
			gocv.GaussianBlur(imgFace, imgFace, image.Pt(23, 23), 30, 50, 4)
			imgFace.Close()
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
