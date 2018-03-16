// What it does:
//
// This example uses the Tensorflow (https://www.tensorflow.org/) deep learning framework
// to classify whatever is in front of the camera.
//
// Download the Tensorflow "Inception" model and descriptions file from:
// https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip
//
// Extract the tensorflow_inception_graph.pb model file from the .zip file.
//
// Also extract the imagenet_comp_graph_label_strings.txt file with the descriptions.
//
// How to run:
//
// 		go run ./cmd/tf-classifier/main.go 0 [modelfile] [descriptionsfile]
//
// +build example

package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"

	"gocv.io/x/gocv"
)

// readDescriptions reads the descriptions from a file
// and returns a slice of its lines.
func readDescriptions(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("How to run:\ntf-classifier [camera ID] [modelfile] [descriptionsfile]")
		return
	}

	// parse args
	deviceID, _ := strconv.Atoi(os.Args[1])
	model := os.Args[2]
	descr := os.Args[3]
	descriptions, err := readDescriptions(descr)
	if err != nil {
		fmt.Printf("Error reading descriptions file: %v\n", descr)
		return
	}

	// open capture device
	webcam, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	window := gocv.NewWindow("Tensorflow Classifier")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	// open DNN classifier
	net := gocv.ReadNetFromTensorflow(model)
	if net.Empty() {
		fmt.Printf("Error reading network model : %v\n", model)
		return
	}
	defer net.Close()

	status := "Ready"
	statusColor := color.RGBA{0, 255, 0, 0}
	fmt.Printf("Start reading camera device: %v\n", deviceID)

	for {
		if ok := webcam.Read(img); !ok {
			fmt.Printf("Error cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// convert image Mat to 224x224 blob that the classifier can analyze
		blob := gocv.BlobFromImage(img, 1.0, image.Pt(224, 224), gocv.NewScalar(0, 0, 0, 0), true, false)

		// feed the blob into the classifier
		net.SetInput(blob, "input")

		// run a forward pass thru the network
		prob := net.Forward("softmax2")

		// reshape the results into a 1x1000 matrix
		probMat := prob.Reshape(1, 1)

		// determine the most probable classification
		_, maxVal, _, maxLoc := gocv.MinMaxLoc(probMat)

		// display classification
		desc := "Unknown"
		if maxLoc.X < 1000 {
			desc = descriptions[maxLoc.X]
		}
		status = fmt.Sprintf("description: %v, maxVal: %v\n", desc, maxVal)
		gocv.PutText(img, status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		blob.Close()
		prob.Close()
		probMat.Close()

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
