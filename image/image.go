// Package image can be perceived as an utilitary library for all kind of
// image processing operations, like image conversion, image decoding and encoding,
// pixel manipulation, color conversion etc.
package image

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"strings"

	"github.com/esimov/caire/utils"
)

func DecodeContent(src string) (image.Image, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("could not open the mask file: %v", err)
	}

	ctype, err := utils.DetectContentType(file.Name())
	if err != nil {
		return nil, err
	}

	if !strings.Contains(ctype.(string), "image") {
		return nil, fmt.Errorf("the mask should be an image file")
	}

	img, err := Decode(file)
	if err != nil {
		return nil, fmt.Errorf("cannot decode content: %w", err)
	}

	return img, nil
}

// Decode decodes an image file to type image.Image
func Decode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("could not decode the mask file: %v", err)
	}

	return img, nil
}

// ImgToNRGBA converts any image type to *image.NRGBA with min-point at (0, 0).
func ImgToNRGBA(img image.Image) *image.NRGBA {
	srcBounds := img.Bounds()
	if srcBounds.Min.X == 0 && srcBounds.Min.Y == 0 {
		if src0, ok := img.(*image.NRGBA); ok {
			return src0
		}
	}
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y

	dstBounds := srcBounds.Sub(srcBounds.Min)
	dstW := dstBounds.Dx()
	dstH := dstBounds.Dy()
	dst := image.NewNRGBA(dstBounds)

	switch src := img.(type) {
	case *image.NRGBA:
		rowSize := srcBounds.Dx() * 4
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			si := src.PixOffset(srcMinX, srcMinY+dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				copy(dst.Pix[di:di+rowSize], src.Pix[si:si+rowSize])
			}
		}
	case *image.YCbCr:
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				srcX := srcMinX + dstX
				srcY := srcMinY + dstY
				siy := src.YOffset(srcX, srcY)
				sic := src.COffset(srcX, srcY)
				r, g, b := color.YCbCrToRGB(src.Y[siy], src.Cb[sic], src.Cr[sic])
				dst.Pix[di+0] = r
				dst.Pix[di+1] = g
				dst.Pix[di+2] = b
				dst.Pix[di+3] = 0xff
				di += 4
			}
		}
	default:
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				c := color.NRGBAModel.Convert(img.At(srcMinX+dstX, srcMinY+dstY)).(color.NRGBA)
				dst.Pix[di+0] = c.R
				dst.Pix[di+1] = c.G
				dst.Pix[di+2] = c.B
				dst.Pix[di+3] = c.A
				di += 4
			}
		}
	}

	return dst
}

// RgbToGrayscale converts an image to grayscale mode and
// returns the pixel values as an one dimensional array.
func RgbToGrayscale(img *image.NRGBA) []uint8 {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	gray := make([]uint8, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray[y*width+x] = uint8(
				(0.299*float64(r) +
					0.587*float64(g) +
					0.114*float64(b)) / 256,
			)
		}
	}

	return gray
}

// Dither converts an image to black and white image, where the white is fully transparent.
func Dither(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	dithered := image.NewNRGBA(bounds)
	dx, dy := bounds.Dx(), bounds.Dy()

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
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

// ImageToPix converts an image to a pixel array.
func ImageToPix(img *image.NRGBA) []uint8 {
	bounds := img.Bounds()
	pixels := make([]uint8, 0, bounds.Max.X*bounds.Max.Y*4)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := img.At(y, x).RGBA()
			pixels = append(pixels, uint8(r>>8), uint8(g>>8), uint8(b>>8), 255)
		}
	}

	return pixels
}

// PixToImage converts an array buffer to an image.
func PixToImage(pixels []uint8, width, height int) image.Image {
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
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

// RotateImage90 rotate the image by 90 degree counter clockwise.
func RotateImage90(img *image.NRGBA) *image.NRGBA {
	bounds := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Max.Y, bounds.Max.X))

	for dstY := 0; dstY < bounds.Max.X; dstY++ {
		for dstX := 0; dstX < bounds.Max.Y; dstX++ {
			srcX := bounds.Max.X - dstY - 1
			srcY := dstX

			srcOff := srcY*img.Stride + srcX*4
			dstOff := dstY*dst.Stride + dstX*4
			copy(dst.Pix[dstOff:dstOff+4], img.Pix[srcOff:srcOff+4])
		}
	}

	return dst
}

// RotateImage270 rotate the image by 270 degree counter clockwise.
func RotateImage270(img *image.NRGBA) *image.NRGBA {
	bounds := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Max.Y, bounds.Max.X))

	for dstY := 0; dstY < bounds.Max.X; dstY++ {
		for dstX := 0; dstX < bounds.Max.Y; dstX++ {
			srcX := dstY
			srcY := bounds.Max.Y - dstX - 1

			srcOff := srcY*img.Stride + srcX*4
			dstOff := dstY*dst.Stride + dstX*4
			copy(dst.Pix[dstOff:dstOff+4], img.Pix[srcOff:srcOff+4])
		}
	}

	return dst
}
