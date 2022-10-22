package caire

import (
	"image"
	"image/color"
)

// Grayscale converts the image to grayscale mode.
func (p *Processor) Grayscale(src *image.NRGBA) *image.NRGBA {
	dx, dy := src.Bounds().Max.X, src.Bounds().Max.Y
	dst := image.NewNRGBA(src.Bounds())

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r, g, b, _ := src.At(x, y).RGBA()
			lum := float32(r)*0.299 + float32(g)*0.587 + float32(b)*0.114
			pixel := color.Gray{Y: uint8(lum / 256)}
			dst.Set(x, y, pixel)
		}
	}
	return dst
}

// Dither converts an image to black and white image, where the white is fully transparent.
func (p *Processor) Dither(src *image.NRGBA) *image.NRGBA {
	var (
		bounds   = src.Bounds()
		dithered = image.NewNRGBA(bounds)
		dx       = bounds.Dx()
		dy       = bounds.Dy()
	)

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r, g, b, _ := src.At(x, y).RGBA()
			threshold := func() color.Color {
				if r > 127 && g > 127 && b > 127 {
					return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
				}
				return color.NRGBA{A: 0x00}
			}
			dithered.Set(x, y, threshold())
		}
	}

	return dithered
}
