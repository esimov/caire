package caire

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/esimov/caire/utils"
	"golang.org/x/image/bmp"
)

// decodeImg decodes an image file to type image.Image
func decodeImg(src string) (image.Image, error) {
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

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("could not decode the mask file: %v", err)
	}

	return img, nil
}

// encodeImg encodes an image to a destination of type io.Writer.
func encodeImg(p *Processor, w io.Writer, img *image.NRGBA) error {
	switch w := w.(type) {
	case *os.File:
		ext := filepath.Ext(w.Name())
		switch ext {
		case "", ".jpg", ".jpeg":
			res, err := Resize(p, img)
			if err != nil {
				return err
			}
			return jpeg.Encode(w, res, &jpeg.Options{Quality: 100})
		case ".png":
			res, err := Resize(p, img)
			if err != nil {
				return err
			}
			return png.Encode(w, res)
		case ".bmp":
			res, err := Resize(p, img)
			if err != nil {
				return err
			}
			return bmp.Encode(w, res)
		default:
			return errors.New("unsupported image format")
		}
	default:
		res, err := Resize(p, img)
		if err != nil {
			return err
		}
		return jpeg.Encode(w, res, &jpeg.Options{Quality: 100})
	}
}

// rotateImage90 rotate the image by 90 degree counter clockwise.
func rotateImage90(src *image.NRGBA) *image.NRGBA {
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

// rotateImage270 rotate the image by 270 degree counter clockwise.
func rotateImage270(src *image.NRGBA) *image.NRGBA {
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

// imgToNRGBA converts any image type to *image.NRGBA with min-point at (0, 0).
func imgToNRGBA(img image.Image) *image.NRGBA {
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

// imgToPix converts an image to a pixel array.
func imgToPix(src *image.NRGBA) []uint8 {
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
func pixToImage(pixels []uint8, width, height int) image.Image {
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

// rgbToGrayscale converts an image to grayscale mode and
// returns the pixel values as an one dimensional array.
func rgbToGrayscale(src *image.NRGBA) []uint8 {
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

// dither converts an image to black and white image, where the white is fully transparent.
func dither(src *image.NRGBA) *image.NRGBA {
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
