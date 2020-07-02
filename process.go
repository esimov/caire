package caire

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"golang.org/x/image/bmp"
)

const maxResizeWithoutScaling = 2000

var (
	g      *gif.GIF
	xCount int
	yCount int
	isGif  = false
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
	FaceDetect     bool
	FaceAngle      float64
	Classifier     string
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
	var (
		newImg    image.Image
		newWidth  int
		newHeight int
		pw, ph    int
	)
	xCount, yCount = 0, 0

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

		if isGif {
			g = encodeImageToGif(img)
		}
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

		// In case pw and ph is zero, it means that the target image is square.
		// In this case we don't have to apply the seam carving algorithm, we can simply resize the image.
		if pw == 0 && ph == 0 {
			return resize.Resize(uint(p.NewWidth), 0, img, resize.Lanczos3), nil
		}

		if p.Square {
			if p.NewWidth < p.NewHeight {
				newImg = resize.Resize(uint(p.NewWidth), 0, img, resize.Lanczos3)
			} else {
				newImg = resize.Resize(uint(p.NewHeight), 0, img, resize.Lanczos3)
			}
			dst := image.NewNRGBA(newImg.Bounds())
			draw.Draw(dst, newImg.Bounds(), newImg, image.ZP, draw.Src)
			img = dst

			nw, nh := img.Bounds().Dx(), img.Bounds().Dy()
			if nw > nh {
				pw = nw - nh
				ph = 0
			} else {
				ph = nh - nw
				pw = 0
			}
		}

		if p.Percentage {
			// Calculate new sizes based on provided percentage.
			pw = c.Width - int(float64(c.Width)-(float64(p.NewWidth)/100*float64(c.Width)))
			ph = c.Height - int(float64(c.Height)-(float64(p.NewHeight)/100*float64(c.Height)))

			if pw > newWidth || ph > newHeight {
				return nil, errors.New("the generated image size should be less than the original image size")
			}
		}
		// Reduce image size horizontally
		for x := 0; x < pw; x++ {
			reduce()
			xCount++
		}
		// Reduce image size vertically
		img = c.RotateImage90(img)
		for y := 0; y < ph; y++ {
			reduce()
			yCount++
		}
		img = c.RotateImage270(img)
	} else if newWidth > 0 || newHeight > 0 {
		// Use this option to rescale the image proportionally prior resizing.
		// First the image is scaled down preserving the image aspect ratio,
		// then the seam carving algorithm is applied only to the remaining pixels.
		// Ex. : given an image of dimensions 2048x1536 if we want to resize to the 1024x500,
		// the tool first rescale the image to 1024x768, then it will remove the remaining 268px.

		// Prevent memory overflow issue in case of huge images by switching to scaling first option
		if img.Bounds().Dx() > maxResizeWithoutScaling ||
			img.Bounds().Dy() > maxResizeWithoutScaling {
			p.Scale = true
		}

		if p.Scale {
			if p.NewWidth > img.Bounds().Max.X || p.NewHeight > img.Bounds().Max.Y {
				return nil, errors.New("scale option can not be used on image enlargement")
			}
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
			dst := image.NewNRGBA(newImg.Bounds())
			draw.Draw(dst, newImg.Bounds(), newImg, image.ZP, draw.Src)
			img = dst
		}

		// Check if the new width does not match with the rescaled image width.
		// We only need to run the carver function if the desired image width is less than the rescaled image width.
		if newWidth > 0 && newWidth != img.Bounds().Max.X {
			if p.NewWidth > c.Width {
				for x := 0; x < newWidth; x++ {
					enlarge()
				}
			} else {
				for x := 0; x < newWidth; x++ {
					reduce()
					xCount++
				}
			}
		}
		// Check if the new height does not match with the rescaled image height.
		// We only need to run the carver function if the desired image height is less than the rescaled image height.
		if newHeight > 0 && newHeight != img.Bounds().Max.Y {
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
// We are using the io package, because this way we can provide different types of input and output source,
// as long as they implement the io.Reader and io.Writer interface.
func (p *Processor) Process(r io.Reader, w io.Writer) error {
	g = new(gif.GIF)
	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	img := imgToNRGBA(src)

	switch w.(type) {
	case *os.File:
		ext := filepath.Ext(w.(*os.File).Name())
		switch ext {
		case "", ".jpg", ".jpeg":
			res, err := Resize(p, img)
			if err != nil {
				return err
			}
			err = jpeg.Encode(w, res, &jpeg.Options{Quality: 100})
		case ".png":
			res, err := Resize(p, img)
			if err != nil {
				return err
			}
			err = png.Encode(w, res)
		case ".bmp":
			res, err := Resize(p, img)
			if err != nil {
				return err
			}
			err = bmp.Encode(w, res)
		case ".gif":
			isGif = true
			_, err := Resize(p, img)
			if err != nil {
				return err
			}
			err = writeGifToFile(w.(*os.File).Name())
		default:
			err = errors.New("unsupported image format")
		}
	default:
		res, err := Resize(p, img)
		if err != nil {
			return err
		}
		err = jpeg.Encode(w, res, &jpeg.Options{Quality: 100})
	}
	return err
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

// encodeImageToGif encode the provided image file to a Gif image.
func encodeImageToGif(src image.Image) *gif.GIF {
	bounds := src.Bounds()
	dst := image.NewPaletted(image.Rect(0, 0, bounds.Dx()-xCount, bounds.Dy()-yCount), palette.Plan9)
	draw.Draw(dst, src.Bounds(), src, image.Point{}, draw.Src)
	g.Image = append(g.Image, dst)
	g.Delay = append(g.Delay, 0)

	return g
}

// writeGifToFile writes the encoded Gif file to the destination file.
func writeGifToFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, g)
}
