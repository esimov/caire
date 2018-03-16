// What it does:
//
// This example uses the Intel CV SDK PVL FaceRecognizer to recognize people by their face.
// It first detects faces using the FaceDetector, then uses that information to try to identify
// them in the database.
//
// Pressing the 't' key will train the database based on the currently
// detected face. If the data file does not exist, it will be created.
//
// Pressing the 'Esc' key will exit the program.
//
// How to run:
//
// facerecognizer [camera ID] [face data file]
//
// 		go run ./cmd/pvl/facerecognizer/main.go 0 ./myfacedata.xml
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

const (
	T_KEY   = 116
	ESC_KEY = 27
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\facerecognizer [camera ID] [face data file]")
		return
	}

	// parse args
	deviceID, _ := strconv.Atoi(os.Args[1])
	fileName := os.Args[2]

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("PVL Face Recognizer")
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

	// load PVL FaceDetector to detect faces
	fd := pvl.NewFaceDetector()
	defer fd.Close()

	// enable tracking mode for more efficient tracking of video source
	fd.SetTrackingModeEnabled(true)

	// only track 1 face at a time for recognizer
	fd.SetMaxDetectableFaces(1)

	// load PVL FaceRecognizer to recognize faces
	var fr pvl.FaceRecognizer

	// if the database file already exists, read it in
	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("Reading in file", fileName)
		fr = pvl.LoadFaceRecognizer(fileName)
		fmt.Println("registered:", fr.GetNumRegisteredPersons())
	} else {
		fr = pvl.NewFaceRecognizer()
	}
	defer fr.Close()

	// enable tracking mode for more efficient tracking of video source
	fr.SetTrackingModeEnabled(true)

	var personIDs []int

	fmt.Printf("start reading camera device: %v\n", deviceID)
MainLoop:
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
		if len(faces) > 0 {
			fmt.Println("found faces")

			// try to recognize the face
			personIDs, _ = fr.Recognize(imgGray, faces)

			// set the color of the box based on if the face is recognized
			color := blue
			msg := "Unknown"
			if len(personIDs) > 0 && personIDs[0] != pvl.UnknownFace {
				color = green
				msg = strconv.Itoa(personIDs[0])
			}

			// draw a rectangle the face on the original image,
			// along with text identifing if recognized
			rect := faces[0].Rectangle()
			gocv.Rectangle(img, rect, color, 3)

			size := gocv.GetTextSize(msg, gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(rect.Min.X+(rect.Min.X/2)-(size.X/2), rect.Min.Y-2)
			gocv.PutText(img, msg, pt, gocv.FontHersheyPlain, 1.2, color, 2)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		key := window.WaitKey(1)
		switch key {
		case T_KEY:
			// train
			if len(faces) > 0 && len(personIDs) > 0 && personIDs[0] == pvl.UnknownFace {
				pid := fr.CreateNewPersonID()
				fr.RegisterFace(imgGray, faces[0], pid, true)
				fr.Save(fileName)
			}
		case ESC_KEY:
			break MainLoop
		}
	}
}
