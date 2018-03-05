package caire

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	_ "golang.org/x/image/bmp"
)

// SeamCarver interface defines the Resize method.
// This has to be implemented by every struct which declares a Resize method.
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
	Square         bool
	Debug          bool
	Scale          bool
}

// Resize implements the Resize method of the Carver interface.
// It returns the concrete resize operation method.
func Resize(s SeamCarver, img *image.NRGBA) (image.Image, error) {
	return s.Resize(img)
}

// Resize method takes the source image and rescales it using the parameters provided.
// The new image can be rescaled either horizontally or vertically (or both).
// Depending on the provided parameters the image can be either reduced or enlarged.
func (p *Processor) Resize(img *image.NRGBA) (image.Image, error) {
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	var newImg image.Image
	var newWidth, newHeight int
	var pw, ph int

	if p.NewWidth > c.Width {
		newWidth = p.NewWidth - (p.NewWidth - (p.NewWidth - c.Width))
	} else {
		newWidth = c.Width - (c.Width - (c.Width - p.NewWidth))
	}

	if p.NewHeight > c.Height {
		newHeight = p.NewHeight - (p.NewHeight - (p.NewHeight - c.Height))
	} else {
		newHeight = c.Height - (c.Height - (c.Height - p.NewHeight))
	}

	if p.NewWidth == 0 {
		newWidth = p.NewWidth
	}
	if p.NewHeight == 0 {
		newHeight = p.NewHeight
	}
	reduce := func() {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	enlarge := func() {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.AddSeam(img, seams, p.Debug)
	}

	if p.Percentage || p.Square {
		// When square option is used the image will be resized to a square based on the shortest edge.
		pw = c.Width - c.Height
		ph = c.Height - c.Width

		if p.Percentage {
			// Calculate new sizes based on provided percentage.
			pw = c.Width - int(float64(c.Width)-(float64(p.NewWidth)/100*float64(c.Width)))
			ph = c.Height - int(float64(c.Height)-(float64(p.NewHeight)/100*float64(c.Height)))

			if pw > newWidth || ph > newHeight {
				return nil, errors.New("the generated image size should be less than original image size")
			}
		}
		// Reduce image size horizontally
		for x := 0; x < pw; x++ {
			reduce()
		}
		// Reduce image size vertically
		img = c.RotateImage90(img)
		for y := 0; y < ph; y++ {
			reduce()
		}
		img = c.RotateImage270(img)
	} else if newWidth > 0 || newHeight > 0 {
		// p.Scale will the scale the image proportionally.
		// First the image is scaled down preserving the image aspect ratio,
		// then the seam carving algorithm is applied only to remaining points.
		// Ex. : given an image of dimensions 2048x1536 if we want to resize to the 1024x500,
		// the tool first rescale the image to 1024x768, then it will remove the remaining 268px.
		if p.Scale {
			// Preserve the aspect ratio on horizontal or vertical axes.
			if p.NewWidth > p.NewHeight {
				newWidth = 0
				newImg = resize.Resize(uint(p.NewWidth), 0, img, resize.Lanczos3)
				if p.NewHeight < newImg.Bounds().Dy() {
					newHeight = newImg.Bounds().Dy() - p.NewHeight
				} else {
					return nil, errors.New("cannot rescale to this size preserving the image aspect ratio")
				}
			} else {
				newHeight = 0
				newImg = resize.Resize(0, uint(p.NewHeight), img, resize.Lanczos3)
				if p.NewWidth < newImg.Bounds().Dx() {
					newWidth = newImg.Bounds().Dx() - p.NewWidth
				} else {
					return nil, errors.New("cannot rescale to this size preserving the image aspect ratio")
				}
			}
			dst := image.NewNRGBA(image.Rect(0, 0, newImg.Bounds().Max.X, newImg.Bounds().Max.Y))
			draw.Draw(dst, image.Rect(0, 0, newImg.Bounds().Dx(), newImg.Bounds().Dy()), newImg, image.ZP, draw.Src)
			img = dst
		}

		if newWidth > 0 {
			if p.NewWidth > c.Width {
				for x := 0; x < newWidth; x++ {
					enlarge()
				}
			} else {
				for x := 0; x < newWidth; x++ {
					reduce()
				}
			}
		}
		if newHeight > 0 {
			img = c.RotateImage90(img)
			if p.NewHeight > c.Height {
				for y := 0; y < newHeight; y++ {
					enlarge()
				}
			} else {
				for y := 0; y < newHeight; y++ {
					reduce()
				}
			}
			img = c.RotateImage270(img)
		}
	}
	return img, nil
}

// Process is the main function having as parameters an input reader and an output writer.
// We are using the `io` package, because this way we can provide different types of input and output source,
// as long as they implement the `io.Reader` and `io.Writer` interface.
func (p *Processor) Process(r io.Reader, w io.Writer) error {
	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	img := imgToNRGBA(src)
	res, err := Resize(p, img)
	if err != nil {
		return err
	}

	return jpeg.Encode(w, res, &jpeg.Options{Quality: 100})
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
