package pigo

import (
	"image"
)

// RgbToGrayscale converts the image to grayscale mode.
func RgbToGrayscale(src image.Image) []uint8 {
	cols, rows := src.Bounds().Dx(), src.Bounds().Dy()
	gray := make([]uint8, rows*cols)

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			gray[y*cols+x] = uint8(
				(0.299*float64(r) +
					0.587*float64(g) +
					0.114*float64(b)) / 256,
			)
		}
	}
	return gray
}
