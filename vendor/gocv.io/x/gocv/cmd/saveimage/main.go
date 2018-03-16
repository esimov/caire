// What it does:
//
// This example uses the VideoCapture class to capture a frame from a connected webcam,
// then save it to an image file on disk.
//
// How to run:
//
// saveimage [camera ID] [image file]
//
// 		go run ./cmd/saveimage/main.go filename.jpg
//
// +build example

package main

import (
	"fmt"
	"os"
	"strconv"

	"gocv.io/x/gocv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("How to run:\n\tsaveimage [camera ID] [image file]")
		return
	}

	deviceID, _ := strconv.Atoi(os.Args[1])
	saveFile := os.Args[2]

	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(img); !ok {
		fmt.Printf("cannot read device %d\n", deviceID)
		return
	}
	if img.Empty() {
		fmt.Printf("no image on device %d\n", deviceID)
		return
	}

	gocv.IMWrite(saveFile, img)
}
