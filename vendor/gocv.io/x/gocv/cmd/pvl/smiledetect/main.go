// What it does:
//
// This example uses the Intel CV SDK PVL FaceDetect class to detect smiles!
// It first detects faces, then detects the smiles on each. Based on if the person is
// smiling, it draws a green or blue rectangle around each of them,
// before displaying them within a Window.
//
// How to run:
//
// smiledetect [camera ID]
//
// 		go run ./cmd/pvl/smiledetect/main.go 0
//
// +build example

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/pvl"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\tsmiledetect [camera ID]")
		return
	}

	// parse args
	deviceID, _ := strconv.Atoi(os.Args[1])

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("PVL Smile Detect")
	defer window.Close()

	// prepare input image matrix
	img := gocv.NewMat()
	defer img.Close()

	// prepare grayscale image matrix
	imgGray := gocv.NewMat()
	defer imgGray.Close()

	// colors to draw the rect for detected faces
	blue := color.RGBA{0, 0, 255, 0}
	green := color.RGBA{0, 255, 0, 0}

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

		// convert image Mat to grayscale Mat for detection
		gocv.CvtColor(img, imgGray, gocv.ColorBGRToGray)

		// detect faces
		faces := fd.DetectFaceRect(imgGray)
		fmt.Printf("found %d faces\n", len(faces))

		// draw a rectangle around each face on the original image,
		// along with text identifing as "Human"
		for _, face := range faces {
			// detect smile
			fd.DetectEye(imgGray, face)
			fd.DetectSmile(imgGray, face)

			// set the color of the box based on if the human is smiling
			color := blue
			if face.IsSmiling() {
				color = green
			}

			rect := face.Rectangle()
			gocv.Rectangle(img, rect, color, 3)

			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(rect.Min.X+(rect.Min.X/2)-(size.X/2), rect.Min.Y-2)
			gocv.PutText(img, "Human", pt, gocv.FontHersheyPlain, 1.2, color, 2)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
