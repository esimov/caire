package pigo

import (
	"image"
)

// RgbToGrayscale converts the image to grayscale mode.
func RgbToGrayscale(src *image.NRGBA) []uint8 {
	rows, cols := src.Bounds().Dx(), src.Bounds().Dy()
	gray := make([]uint8, rows*cols)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			gray[r*cols+c] = uint8(
				0.299*float64(src.Pix[r*4*cols+4*c+0]) +
				0.587*float64(src.Pix[r*4*cols+4*c+1]) +
				0.114*float64(src.Pix[r*4*cols+4*c+2]),
			)
		}
	}
	return gray
}
