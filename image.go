package caire

import (
	"image"
	"image/color"

	"github.com/esimov/caire/utils"
)

const (
	Darken = iota
	Lighten
	Multiply
	Screen
	Overlay
)

// Grayscale converts the source image to grayscale mode.
func (c *Carver) Grayscale(src *image.NRGBA) *image.NRGBA {
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

func (c *Carver) BlendingOp(src, msk *image.NRGBA, opType int) *image.NRGBA {
	dx, dy := src.Bounds().Dx(), src.Bounds().Dy()
	img := image.NewNRGBA(src.Bounds())

	var (
		r, g, b, a     uint32
		rn, gn, bn, an float64
	)

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r1, g1, b1, a1 := src.At(x, y).RGBA()
			r2, g2, b2, a2 := msk.At(x, y).RGBA()

			rs, gs, bs, as := r1>>8, g1>>8, b1>>8, a1>>8
			rb, gb, bb, ab := r2>>8, g2>>8, b2>>8, a2>>8

			rsn := float64(rs) / 255
			gsn := float64(gs) / 255
			bsn := float64(bs) / 255
			asn := float64(as) / 255

			rbn := float64(rb) / 255
			gbn := float64(gb) / 255
			bbn := float64(bb) / 255
			abn := float64(ab) / 255

			// applying the blending mode
			switch opType {
			case Darken:
				rn = utils.Min(rsn, rbn)
				gn = utils.Min(gsn, gbn)
				bn = utils.Min(bsn, bbn)
				an = utils.Min(asn, abn)
			case Lighten:
				rn = utils.Max(rsn, rbn)
				gn = utils.Max(gsn, gbn)
				bn = utils.Max(bsn, bbn)
				an = utils.Max(asn, abn)
			case Screen:
				rn = 1 - (1-rsn)*(1-rbn)
				gn = 1 - (1-gsn)*(1-gbn)
				bn = 1 - (1-bsn)*(1-bbn)
				an = 1 - (1-asn)*(1-abn)
			case Multiply:
				rn = rsn * rbn
				gn = gsn * gbn
				bn = bsn * bbn
				an = asn * abn
			case Overlay:
				if rsn <= 0.5 {
					rn = 2 * rsn * rbn
				} else {
					rn = 1 - 2*(1-rsn)*(1-rbn)
				}
				if gsn <= 0.5 {
					gn = 2 * gsn * gbn
				} else {
					gn = 1 - 2*(1-gsn)*(1-gbn)
				}
				if bsn <= 0.5 {
					bn = 2 * bsn * bbn
				} else {
					bn = 1 - 2*(1-bsn)*(1-bbn)
				}
				if asn <= 0.5 {
					an = 2 * asn * abn
				} else {
					an = 1 - 2*(1-asn)*(1-abn)
				}
			}

			r = uint32(rn * 255)
			g = uint32(gn * 255)
			b = uint32(bn * 255)
			a = uint32(an * 255)

			img.Set(x, y, color.NRGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			})
		}
	}
	return img
}

// RotateImage90 rotate the image by 90 degree counter clockwise.
func (c *Carver) RotateImage90(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Max.Y, b.Max.X))
	for dstY := 0; dstY < b.Max.X; dstY++ {
		for dstX := 0; dstX < b.Max.Y; dstX++ {
			srcX := b.Max.X - dstY - 1
			srcY := dstX

			srcOff := srcY*src.Stride + srcX*4
			dstOff := dstY*dst.Stride + dstX*4
			copy(dst.Pix[dstOff:dstOff+4], src.Pix[srcOff:srcOff+4])
		}
	}
	return dst
}

// RotateImage270 rotate the image by 270 degree counter clockwise.
func (c *Carver) RotateImage270(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Max.Y, b.Max.X))
	for dstY := 0; dstY < b.Max.X; dstY++ {
		for dstX := 0; dstX < b.Max.Y; dstX++ {
			srcX := dstY
			srcY := b.Max.Y - dstX - 1

			srcOff := srcY*src.Stride + srcX*4
			dstOff := dstY*dst.Stride + dstX*4
			copy(dst.Pix[dstOff:dstOff+4], src.Pix[srcOff:srcOff+4])
		}
	}
	return dst
}

// imgToPix converts an image to a pixel array.
func (c *Carver) imgToPix(src *image.NRGBA) []uint8 {
	bounds := src.Bounds()
	pixels := make([]uint8, 0, bounds.Max.X*bounds.Max.Y*4)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := src.At(y, x).RGBA()
			pixels = append(pixels, uint8(r>>8), uint8(g>>8), uint8(b>>8), 255)
		}
	}
	return pixels
}

// pixToImage converts an array buffer to an image.
func (c *Carver) pixToImage(pixels []uint8) image.Image {
	dst := image.NewNRGBA(image.Rect(0, 0, c.Width, c.Height))
	bounds := dst.Bounds()
	dx, dy := bounds.Max.X, bounds.Max.Y
	col := color.NRGBA{
		R: uint8(0),
		G: uint8(0),
		B: uint8(0),
		A: uint8(255),
	}

	for x := bounds.Min.X; x < dx; x++ {
		for y := bounds.Min.Y; y < dy*4; y += 4 {
			col.R = uint8(pixels[y+x*dy*4])
			col.G = uint8(pixels[y+x*dy*4+1])
			col.B = uint8(pixels[y+x*dy*4+2])
			col.A = uint8(pixels[y+x*dy*4+3])

			dst.SetNRGBA(x, int(y/4), col)
		}
	}
	return dst
}

// rgbToGrayscale converts an image to grayscale mode and
// returns the pixel values as an one dimensional array.
func (c *Carver) rgbToGrayscale(src *image.NRGBA) []uint8 {
	width, height := src.Bounds().Dx(), src.Bounds().Dy()
	gray := make([]uint8, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			gray[y*width+x] = uint8(
				(0.299*float64(r) +
					0.587*float64(g) +
					0.114*float64(b)) / 256,
			)
		}
	}
	return gray
}
