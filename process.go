package caire

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"github.com/fogleman/gg"
)

// Processor options
type Processor struct {
	BlurRadius      int
	SobelThreshold  int
}

// Process : Convert image
func (p *Processor) Process(file io.Reader, output string) (*os.File, error) {
	src, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	width, height := src.Bounds().Dx(), src.Bounds().Dy()
	ctx := gg.NewContext(width, height)
	ctx.DrawRectangle(0, 0, float64(width), float64(height))
	ctx.SetRGBA(1, 1, 1, 1)
	ctx.Fill()

	img := toNRGBA(src)

	blur := Stackblur(img, uint32(width), uint32(height), uint32(p.BlurRadius))
	gray := Grayscale(blur)
	sobel := SobelFilter(gray, float64(p.SobelThreshold))

	dpTable := &DPTable{width, height, make([]float64, width*height)}
	dpTable.ComputeSeams(sobel)
	dpTable.FindLowestEnergySeams()

	fq, err := os.Create(output)
	if err != nil {
		return nil, err
	}
	defer fq.Close()

	if err = png.Encode(fq, sobel); err != nil {
		return nil, err
	}

	return fq, err
}

// toNRGBA converts any image type to *image.NRGBA with min-point at (0, 0).
func toNRGBA(img image.Image) *image.NRGBA {
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
