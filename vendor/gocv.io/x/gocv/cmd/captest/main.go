// What it does:
//
// This example uses the VideoCapture class to test if you can capture video
// from a connected webcam, by trying to read 100 frames.
//
// How to run:
//
// 		go run ./cmd/captest/main.go
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
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\tcaptest [camera ID]")
		return
	}

	// parse args
	deviceID, _ := strconv.Atoi(os.Args[1])

	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// streaming, capture from webcam
	buf := gocv.NewMat()
	defer buf.Close()

	fmt.Printf("Start reading camera device: %v\n", deviceID)
	for i := 0; i < 100; i++ {
		if ok := webcam.Read(buf); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if buf.Empty() {
			continue
		}

		fmt.Printf("Read frame %d\n", i+1)
	}

	fmt.Println("Done.")
}
