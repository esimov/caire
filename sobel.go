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

// SobelDetector uses the sobel filter operator for detecting image edges.
// See https://en.wikipedia.org/wiki/Sobel_operator
func (c *Carver) SobelDetector(img *image.NRGBA, threshold float64) *image.NRGBA {
	var sumX, sumY int32
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y
	dst := image.NewNRGBA(img.Bounds())
	data := c.getImageData(img)

	// at returns the brightness of the pixel at (x, y). Coordinates outside
	// the image are clamped to the nearest border pixel, so the borders
	// of a uniform image produce no edge.
	at := func(x, y int) int32 {
		x = min(max(x, 0), dx-1)
		y = min(max(y, 0), dy-1)
		return int32(data[y*dx+x])
	}

	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			// Apply the 3x3 kernels centered on the current pixel.
			sumX, sumY = 0, 0
			for ky := 0; ky < len(kernelY); ky++ {
				for kx := 0; kx < len(kernelX); kx++ {
					px := at(x+kx-1, y+ky-1)
					sumX += px * kernelX[ky][kx]
					sumY += px * kernelY[ky][kx]
				}
			}
			magnitude := math.Sqrt(float64(sumX*sumX) + float64(sumY*sumY))
			if magnitude > 255 {
				magnitude = 255
			}
			// Set magnitude to 0 if doesn't exceed threshold, else set to magnitude
			if magnitude <= threshold {
				magnitude = 0
			}

			idx := (y*dx + x) * 4
			m := uint8(magnitude)
			dst.Pix[idx] = m
			dst.Pix[idx+1] = m
			dst.Pix[idx+2] = m
			dst.Pix[idx+3] = 0xff
		}
	}
	return dst
}

// getImageData converts the image to grayscale and returns an array of pixel brightness values.
func (c *Carver) getImageData(img *image.NRGBA) []uint8 {
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y
	pixels := make([]uint8, dx*dy)

	for i := range pixels {
		r, g, b := uint32(img.Pix[i*4]), uint32(img.Pix[i*4+1]), uint32(img.Pix[i*4+2])
		pixels[i] = uint8((299*r + 587*g + 114*b) / 1000)
	}
	return pixels
}
