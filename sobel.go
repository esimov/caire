package caire

import (
	"image"
	"math"
)

type kernel [][]int32

var (
	kernelX = kernel{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	kernelY = kernel{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}
)

// SobelFilter detects image edges.
// See https://en.wikipedia.org/wiki/Sobel_operator
func SobelFilter(img *image.NRGBA, threshold float64) *image.NRGBA {
	var sumX, sumY int32
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y
	dst := image.NewNRGBA(img.Bounds())

	// Get 3x3 window of pixels because image data given is just a 1D array of pixels
	maxPixelOffset := dx*2 + len(kernelX) - 1

	data := getImageData(img)
	length := len(data)*4 - maxPixelOffset
	magnitudes := make([]uint8, length)

	for i := 0; i < length; i++ {
		// Sum each pixel with the kernel value
		sumX, sumY = 0, 0
		for x := 0; x < len(kernelX); x++ {
			for y := 0; y < len(kernelY); y++ {
				if idx := i + (dx * y) + x; idx < len(data) {
					r := data[i+(dx*y)+x]
					sumX += int32(r) * kernelX[y][x]
					sumY += int32(r) * kernelY[y][x]
				}
			}
		}
		magnitude := math.Sqrt(float64(sumX*sumX) + float64(sumY*sumY))
		// Check for pixel color boundaries
		if magnitude < 0 {
			magnitude = 0
		} else if magnitude > 255 {
			magnitude = 255
		}

		// Set magnitude to 0 if doesn't exceed threshold, else set to magnitude
		if magnitude > threshold {
			magnitudes[i] = uint8(magnitude)
		} else {
			magnitudes[i] = 0
		}
	}

	dataLength := dx * dy * 4
	edges := make([]int32, dataLength)

	// Apply the kernel values
	for i := 0; i < dataLength; i++ {
		edges[i] = 0
		if i%4 != 0 {
			m := magnitudes[i/4]
			if m != 0 {
				edges[i-1] = int32(m)
			}
		}
	}

	// Generate the new image with the sobel filter applied
	for idx := 0; idx < len(edges); idx += 4 {
		dst.Pix[idx] = uint8(edges[idx])
		dst.Pix[idx+1] = uint8(edges[idx+1])
		dst.Pix[idx+2] = uint8(edges[idx+2])
		dst.Pix[idx+3] = 255
	}
	return dst
}

// getImageData returns an array of pixel grayscale brightness values
// for the image (taking the red component of each pixel).
func getImageData(img *image.NRGBA) []uint8 {
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y
	pixels := make([]uint8, dx*dy)

	for i := range pixels {
		pixels[i] = img.Pix[i*4]
	}
	return pixels
}
