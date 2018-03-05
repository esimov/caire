package caire

import (
	"image"
	"image/color"
	"testing"
)

const ImgWidth = 10
const ImgHeight = 10

func TestGrayscale(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	for i := 0; i < img.Bounds().Dx(); i++ {
		for j := 0; j < img.Bounds().Dy(); j++ {
			img.Set(i, j, color.RGBA{177, 177, 177, 255})
		}
	}

	for i := 0; i < img.Bounds().Dx(); i++ {
		for j := 0; j < img.Bounds().Dy(); j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8

			if r != g || r != b || g != b {
				t.Errorf("R, G, B value expected to be equal. Got %v, %v, %v", r, g, b)
			}
		}
	}
}
