package caire

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	pigo "github.com/esimov/pigo/core"
	"github.com/pkg/errors"
	"golang.org/x/image/bmp"
)

//go:embed data/facefinder
var classifier embed.FS

var (
	g      *gif.GIF
	xCount int
	yCount int
	isGif  = false
)

// SeamCarver interface defines the Resize method.
// This needs to be implemented by every struct which declares a Resize method.
type SeamCarver interface {
	Resize(*image.NRGBA) (image.Image, error)
}

// Processor options
type Processor struct {
	SobelThreshold   int
	BlurRadius       int
	NewWidth         int
	NewHeight        int
	Percentage       bool
	Square           bool
	Debug            bool
	FaceDetect       bool
	FaceAngle        float64
	PigoFaceDetector *pigo.Pigo
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
		err       error
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

	reduce := func() error {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		if err := c.ComputeSeams(img, p); err != nil {
			return err
		}
		seams := c.FindLowestEnergySeams()
		img = c.RemoveSeam(img, seams, p.Debug)

		if isGif {
			g = encodeImageToGif(img)
		}
		return nil
	}
	enlarge := func() error {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		if err := c.ComputeSeams(img, p); err != nil {
			return err
		}
		seams := c.FindLowestEnergySeams()
		img = c.AddSeam(img, seams, p.Debug)

		return nil
	}

	if p.Percentage || p.Square {
		// When square option is used the image will be resized to a square based on the shortest edge.
		pw = c.Width - c.Height
		ph = c.Height - c.Width

		// In case pw and ph is zero, it means that the target image is square.
		// In this case we don't have to apply the seam carving algorithm, we can simply resize the image.
		if pw == 0 && ph == 0 {
			return imaging.Resize(img, p.NewWidth, 0, imaging.Lanczos), nil
		}

		if p.Square {
			if p.NewWidth < p.NewHeight {
				newImg = imaging.Resize(img, p.NewWidth, 0, imaging.Lanczos)
			} else {
				newImg = imaging.Resize(img, p.NewHeight, 0, imaging.Lanczos)
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
			if err = reduce(); err != nil {
				return nil, err
			}
			xCount++
		}
		// Reduce image size vertically
		img = c.RotateImage90(img)
		for y := 0; y < ph; y++ {
			if err = reduce(); err != nil {
				return nil, err
			}
			yCount++
		}
		img = c.RotateImage270(img)
	} else if newWidth > 0 || newHeight > 0 {
		// We are trying to rescale the image proportionally prior resizing.
		// First the image is scaled down or up by preserving the image aspect ratio,
		// then the seam carving algorithm is applied only to the remaining pixels.

		// Scale the width and height by the smaller factor (i.e Min(wScaleFactor, hScaleFactor))
		// Example: input: 5000x2500, scale: 2160x1080, final target: 1920x1080
		wScaleFactor := float64(c.Width) / float64(p.NewWidth)
		hScaleFactor := float64(c.Height) / float64(p.NewHeight)
		scaleWidth := math.Round(float64(c.Width) / math.Min(wScaleFactor, hScaleFactor))
		scaleHeight := math.Round(float64(c.Height) / math.Min(wScaleFactor, hScaleFactor))

		newImg = imaging.Resize(img, int(scaleWidth), int(scaleHeight), imaging.Lanczos)

		dx0, dy0 := img.Bounds().Max.X, newImg.Bounds().Max.Y
		dx1, dy1 := newImg.Bounds().Max.X, newImg.Bounds().Max.Y

		// Rescale the image only when it's resized both horizontally and vertically
		// and the new image width or height are preserved, otherwise it might happen, that
		// the generated image size does not match with the requested image size.
		if !((p.NewWidth == 0 && dx0 == dx1) || (p.NewHeight == 0 && dy0 == dy1)) {
			// The amount needed to remove by carving. One or both of these will be 0.
			newWidth = int(scaleWidth) - p.NewWidth
			newHeight = int(scaleHeight) - p.NewHeight

			dst := image.NewNRGBA(newImg.Bounds())
			draw.Draw(dst, newImg.Bounds(), newImg, image.ZP, draw.Src)
			img = dst
		}

		// Run the carver function if the desired image width is not identical with the rescaled image width.
		if newWidth > 0 && newWidth != img.Bounds().Max.X {
			// Because of scaling horizontally and vertically at the same time it might happen
			// that the scaled image exceeds the desired image size, we need to make sure
			// that the new width and|or height is reduced and not enlarged.
			if p.NewWidth > c.Width && img.Bounds().Max.X < p.NewWidth {
				for x := 0; x < newWidth; x++ {
					if err = enlarge(); err != nil {
						return nil, err
					}
				}
			} else {
				for x := 0; x < newWidth; x++ {
					if err = reduce(); err != nil {
						return nil, err
					}
					xCount++
				}
			}
		}
		// Run the carver function if the desired image height is not identical with the rescaled image height.
		if newHeight > 0 && newHeight != img.Bounds().Max.Y {
			img = c.RotateImage90(img)
			// Check new height against the width of the image because the image is rotated 90deg.
			if p.NewHeight > c.Height && img.Bounds().Max.X < p.NewHeight {
				for y := 0; y < newHeight; y++ {
					if err = enlarge(); err != nil {
						return nil, err
					}
				}
			} else {
				for y := 0; y < newHeight; y++ {
					if err = reduce(); err != nil {
						return nil, err
					}
				}
			}
			img = c.RotateImage270(img)
		}
	}
	return img, nil
}

// Process encodes the resized image into an io.Writer interface.
// We are using the io package, because this way we can provide different input and output types,
// as long as they implement the io.Reader and io.Writer interface.
func (p *Processor) Process(r io.Reader, w io.Writer) error {
	var err error

	// Instantiate a new Pigo object in case the face detection option is used.
	p.PigoFaceDetector = pigo.NewPigo()

	if p.FaceDetect {
		cascadeFile, err := classifier.ReadFile("data/facefinder")
		if err != nil {
			return errors.New(fmt.Sprintf("error reading the cascade file: %v", err))
		}
		// Unpack the binary file. This will return the number of cascade trees,
		// the tree depth, the threshold and the prediction from tree's leaf nodes.
		p.PigoFaceDetector, err = p.PigoFaceDetector.Unpack(cascadeFile)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading the cascade file: %v\n", err))
		}
	}

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
		case ".gif":
			isGif = true
			_, err := Resize(p, img)
			if err != nil {
				return err
			}
			return writeGifToFile(w.(*os.File).Name())
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
	return nil
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

// encodeImageToGif encodes the provided image to a Gif file.
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
