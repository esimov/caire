package caire

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"image/jpeg"
	"github.com/pkg/errors"
)

// SeamCarver is an interface that Carver uses to implement the Resize function.
// It takes an image and the output as parameters and returns the resized image.
type SeamCarver interface {
	Resize(*image.NRGBA) (image.Image, error)
}

// Processor options
type Processor struct {
	SobelThreshold int
	BlurRadius     int
	NewWidth       int
	NewHeight      int
	Percentage     bool
}


// Implement the Resize method of the Carver interface.
func Resize(s SeamCarver, img *image.NRGBA) (image.Image, error) {
	return s.Resize(img)
}

// This is the main entry point which takes the source image
// and encodes the new, rescaled image into the output file.
func (p *Processor) Resize(img *image.NRGBA) (image.Image, error) {
	var c *Carver = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())

	resize := func() {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.RemoveSeam(img, seams)
	}

	if p.Percentage {
		// Calculate new sizes based on provided percentage.
		nw := c.Width - int(float64(c.Width) - (float64(p.NewWidth)/100 * float64(c.Width)))
		nh := c.Height - int(float64(c.Height) - (float64(p.NewHeight)/100 * float64(c.Height)))
		// Resize image horizontally
		for x := 0; x < nw; x++ {
			resize()
		}
		// Resize image vertically
		img = c.RotateImage90(img)

		// Needs to update the slice width & height because of image rotation.
		c.Width = img.Bounds().Dx()
		c.Height = img.Bounds().Dy()
		for y := 0; y < nh; y++ {
			resize()
		}
		img = c.RotateImage270(img)
	} else if p.NewWidth > 0 || p.NewHeight > 0 {
		if p.NewWidth > 0 {
			if p.NewWidth > c.Width {
				err := errors.New("new width should be less than image width.")
				return nil, err
			}
			for x := 0; x < p.NewWidth; x++ {
				resize()
			}
		}

		if p.NewHeight > 0 {
			if p.NewHeight > c.Height {
				err := errors.New("new height should be less than image height.")
				return nil, err
			}
			img = c.RotateImage90(img)

			// Needs to update the slice width & height because of image rotation
			// otherwise the new image will be cut off.
			c.Width = img.Bounds().Dx()
			c.Height = img.Bounds().Dy()
			for y := 0; y < p.NewHeight; y++ {
				resize()
			}
			img = c.RotateImage270(img)
		}
	}
	return img, nil
}


// Process is the main entry point for the image resize operation.
func (p *Processor) Process(file io.Reader, output string) (*os.File, error) {
	src, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	img := imgToNRGBA(src)
	//sobel := SobelFilter(Grayscale(img), float64(p.SobelThreshold))
	res, err := Resize(p, img)
	if err != nil {
		return nil, err
	}

	fq, err := os.Create(output)
	if err != nil {
		return nil, err
	}
	defer fq.Close()

	if err = jpeg.Encode(fq, res, &jpeg.Options{100}); err != nil {
		return nil, err
	}
	return fq, nil
}

// Converts any image type to *image.NRGBA with min-point at (0, 0).
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
