package caire

import (
	_ "embed"
	"errors"
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
	"strings"

	"github.com/disintegration/imaging"
	"github.com/esimov/caire/utils"
	pigo "github.com/esimov/pigo/core"
	"golang.org/x/image/bmp"
)

//go:embed data/facefinder
var cascadeFile []byte

var (
	g      *gif.GIF
	rCount int
)

var (
	resizeXY = false // the image is resized both verticlaly and horizontally
	isGif    = false

	imgWorker = make(chan worker) // channel used to transfer the image to the GUI
	errs      = make(chan error)
)

// worker struct contains all the information needed for transfering the resized image to the Gio GUI.
type worker struct {
	carver *Carver
	img    *image.NRGBA
	debug  *image.NRGBA
	done   bool
}

// SeamCarver interface defines the Resize method.
// This needs to be implemented by every struct which declares a Resize method.
type SeamCarver interface {
	Resize(*image.NRGBA) (image.Image, error)
}

// shrinkFn is a generic function used to shrink an image.
type shrinkFn func(*Carver, *image.NRGBA) (*image.NRGBA, error)

// enlargeFn is a generic function used to enlarge an image.
type enlargeFn func(*Carver, *image.NRGBA) (*image.NRGBA, error)

// Processor options
type Processor struct {
	SobelThreshold   int
	BlurRadius       int
	NewWidth         int
	NewHeight        int
	Percentage       bool
	Square           bool
	Debug            bool
	Preview          bool
	FaceDetect       bool
	ShapeType        string
	SeamColor        string
	MaskPath         string
	RMaskPath        string
	Mask             *image.NRGBA
	RMask            *image.NRGBA
	GuiDebug         *image.NRGBA
	FaceAngle        float64
	PigoFaceDetector *pigo.Pigo
	Spinner          *utils.Spinner

	vRes bool
}

var (
	shrinkHorizFn  shrinkFn
	shrinkVertFn   shrinkFn
	enlargeHorizFn enlargeFn
	enlargeVertFn  enlargeFn
)

// resize implements the Resize method of the Carver interface.
// It returns the concrete resize operation method.
func resize(s SeamCarver, img *image.NRGBA) (image.Image, error) {
	return s.Resize(img)
}

// Resize is the main entry point for the image resize operation.
// The new image can be resized either horizontally or vertically (or both).
// Depending on the provided options the image can be either reduced or enlarged.
func (p *Processor) Resize(img *image.NRGBA) (image.Image, error) {
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	var (
		newImg    image.Image
		newWidth  int
		newHeight int
		pw, ph    int
		err       error
	)
	rCount = 0

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

	// shrinkHorizFn calls itself recursively to shrink the image horizontally.
	// If the image is resized on both X and Y axis it calls the shrink and enlarge
	// function intermitently up until the desired dimension is reached.
	// We are opting for this solution instead of resizing the image secventially,
	// because this way the horizontal and vertical seams are merged together seamlessly.
	shrinkHorizFn = func(c *Carver, img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = false
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
		if dx > p.NewWidth {
			img, err = p.shrink(c, img)
			if err != nil {
				return nil, err
			}
			if p.NewHeight > 0 && p.NewHeight != dy {
				if p.NewHeight <= dy {
					img, _ = shrinkVertFn(c, img)
				} else {
					img, _ = enlargeVertFn(c, img)
				}
			} else {
				img, _ = shrinkHorizFn(c, img)
			}
		}
		rCount++
		return img, nil
	}

	// enlargeHorizFn calls itself recursively to enlarge the image horizontally.
	enlargeHorizFn = func(c *Carver, img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = false
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
		if dx < p.NewWidth {
			img, err = p.enlarge(c, img)
			if err != nil {
				return nil, err
			}
			if p.NewHeight > 0 && p.NewHeight != dy {
				if p.NewHeight <= dy {
					img, _ = shrinkVertFn(c, img)
				} else {
					img, _ = enlargeVertFn(c, img)
				}
			} else {
				img, _ = enlargeHorizFn(c, img)
			}
		}
		rCount++
		return img, nil
	}

	// shrinkVertFn calls itself recursively to shrink the image vertically.
	shrinkVertFn = func(c *Carver, img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = true
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()

		// If the image is resized both horizontally and vertically we need
		// to rotate the image each time we are invoking the shrink function.
		// Otherwise we rotate the image only once, right before calling this function.
		if resizeXY {
			dx, dy = img.Bounds().Dy(), img.Bounds().Dx()
			img = c.RotateImage90(img)
		}
		if dx > p.NewHeight {
			img, err = p.shrink(c, img)
			if err != nil {
				return nil, err
			}
			if resizeXY {
				img = c.RotateImage270(img)
			}
			if p.NewWidth > 0 && p.NewWidth != dy {
				if p.NewWidth <= dy {
					img, _ = shrinkHorizFn(c, img)
				} else {
					img, _ = enlargeHorizFn(c, img)
				}
			} else {
				img, _ = shrinkVertFn(c, img)
			}
		} else {
			if resizeXY {
				img = c.RotateImage270(img)
			}
		}
		rCount++
		return img, nil
	}

	// enlargeVertFn calls itself recursively to enlarge the image vertically.
	enlargeVertFn = func(c *Carver, img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = true
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()

		if resizeXY {
			dx, dy = img.Bounds().Dy(), img.Bounds().Dx()
			img = c.RotateImage90(img)
		}
		if dx < p.NewHeight {
			img, err = p.enlarge(c, img)
			if err != nil {
				return nil, err
			}
			if resizeXY {
				img = c.RotateImage270(img)
			}
			if p.NewWidth > 0 && p.NewWidth != dy {
				if p.NewWidth <= dy {
					img, _ = shrinkHorizFn(c, img)
				} else {
					img, _ = enlargeHorizFn(c, img)
				}
			} else {
				img, _ = enlargeVertFn(c, img)
			}
		} else {
			if resizeXY {
				img = c.RotateImage270(img)
			}
		}
		rCount++
		return img, nil
	}

	if p.Percentage || p.Square {
		pw = c.Width - c.Height
		ph = c.Height - c.Width

		// In case pw and ph is zero, it means that the target image is square.
		// In this case we can simply resize the image without running the carving operation.
		if p.Percentage && pw == 0 && ph == 0 {
			pw = c.Width - int(float64(c.Width)-(float64(p.NewWidth)/100*float64(c.Width)))
			ph = c.Height - int(float64(c.Height)-(float64(p.NewHeight)/100*float64(c.Height)))

			p.NewWidth = utils.Abs(c.Width - pw)
			p.NewHeight = utils.Abs(c.Height - ph)

			resImgSize := utils.Min(p.NewWidth, p.NewHeight)
			return imaging.Resize(img, resImgSize, 0, imaging.Lanczos), nil
		}

		// When the square option is used the image will be resized to a square based on the shortest edge.
		if p.Square {
			// Calling the image rescale method only when both a new width and height is provided.
			if p.NewWidth != 0 && p.NewHeight != 0 {
				p.NewWidth = utils.Min(p.NewWidth, p.NewHeight)
				p.NewHeight = p.NewWidth

				newImg = p.calculateFitness(img, c)
				dst := image.NewNRGBA(newImg.Bounds())
				draw.Draw(dst, newImg.Bounds(), newImg, image.Point{}, draw.Src)
				img = dst

				nw, nh := img.Bounds().Dx(), img.Bounds().Dy()

				p.NewWidth = utils.Min(nw, nh)
				p.NewHeight = p.NewWidth
			} else {
				return nil, errors.New("please provide a new WIDTH and HEIGHT when using the square option")
			}
		}

		// Use the Percentage flag only for shrinking the image.
		if p.Percentage {
			// Calculate the new image size based on the provided percentage.
			pw = c.Width - int(float64(c.Width)-(float64(p.NewWidth)/100*float64(c.Width)))
			ph = c.Height - int(float64(c.Height)-(float64(p.NewHeight)/100*float64(c.Height)))

			if p.NewWidth != 0 {
				p.NewWidth = utils.Abs(c.Width - pw)
			}
			if p.NewHeight != 0 {
				p.NewHeight = utils.Abs(c.Height - ph)
			}
			if pw >= c.Width || ph >= c.Height {
				return nil, errors.New("cannot use the percentage flag for image enlargement")
			}
		}
	}

	// Rescale the image when it is resized both horizontally and vertically.
	// First the image is scaled down or up by preserving the image aspect ratio,
	// then the seam carving algorithm is applied only to the remaining pixels.

	// Scale the width and height by the smaller factor (i.e Min(wScaleFactor, hScaleFactor))
	// Example: input: 5000x2500, scale: 2160x1080, final target: 1920x1080
	if (c.Width > p.NewWidth && c.Height > p.NewHeight) &&
		(p.NewWidth != 0 && p.NewHeight != 0) {

		newImg = p.calculateFitness(img, c)

		dx0, dy0 := img.Bounds().Max.X, img.Bounds().Max.Y
		dx1, dy1 := newImg.Bounds().Max.X, newImg.Bounds().Max.Y

		// Rescale the image when the new image width or height are preserved, otherwise
		// it might happen, that the generated image size does not match with the requested image size.
		if !((p.NewWidth == 0 && dx0 == dx1) || (p.NewHeight == 0 && dy0 == dy1)) {
			dst := image.NewNRGBA(newImg.Bounds())
			draw.Draw(dst, newImg.Bounds(), newImg, image.Point{}, draw.Src)
			img = dst
		}
	}

	// Run the carver function if the desired image width is not identical with the rescaled image width.
	if newWidth > 0 && p.NewWidth != c.Width {
		if p.NewWidth > c.Width {
			img, err = enlargeHorizFn(c, img)
			if err != nil {
				return nil, err
			}
		} else {
			img, err = shrinkHorizFn(c, img)
			if err != nil {
				return nil, err
			}
		}
	}

	// Run the carver function if the desired image height is not identical with the rescaled image height.
	if newHeight > 0 && p.NewHeight != c.Height {
		if !resizeXY {
			img = c.RotateImage90(img)

			if len(p.MaskPath) > 0 {
				p.Mask = c.RotateImage90(p.Mask)
			}
			if len(p.RMaskPath) > 0 {
				p.RMask = c.RotateImage90(p.RMask)
			}
		}
		if p.NewHeight > c.Height {
			img, err = enlargeVertFn(c, img)
			if err != nil {
				return nil, err
			}
		} else {
			img, err = shrinkVertFn(c, img)
			if err != nil {
				return nil, err
			}
		}
		if !resizeXY {
			img = c.RotateImage270(img)

			if len(p.MaskPath) > 0 {
				p.Mask = c.RotateImage270(p.Mask)
			}
			if len(p.RMaskPath) > 0 {
				p.RMask = c.RotateImage270(p.RMask)
			}
		}
	}
	// Signal that the process is done and no more data is sent through the channel.
	go func() {
		imgWorker <- worker{
			carver: nil,
			img:    nil,
			done:   true,
		}
	}()

	return img, nil
}

// calculateFitness iteratively try to find the best image aspect ratio for the rescale.
func (p *Processor) calculateFitness(img *image.NRGBA, c *Carver) *image.NRGBA {
	var (
		w      = float64(c.Width)
		h      = float64(c.Height)
		nw     = float64(p.NewWidth)
		nh     = float64(p.NewHeight)
		newImg *image.NRGBA
	)
	wsf := w / nw
	hsf := h / nh
	sw := math.Round(w / math.Min(wsf, hsf))
	sh := math.Round(h / math.Min(wsf, hsf))

	if sw <= sh {
		newImg = imaging.Resize(img, 0, int(sw), imaging.Lanczos)
		if len(p.MaskPath) > 0 {
			p.Mask = imaging.Resize(p.Mask, 0, int(sw), imaging.Lanczos)
		}
		if len(p.RMaskPath) > 0 {
			p.RMask = imaging.Resize(p.RMask, 0, int(sw), imaging.Lanczos)
		}
	} else {
		newImg = imaging.Resize(img, 0, int(sh), imaging.Lanczos)
		if len(p.MaskPath) > 0 {
			p.Mask = imaging.Resize(p.Mask, 0, int(sh), imaging.Lanczos)
		}
		if len(p.RMaskPath) > 0 {
			p.RMask = imaging.Resize(p.RMask, 0, int(sh), imaging.Lanczos)
		}
	}
	dx, dy := newImg.Bounds().Max.X, newImg.Bounds().Max.Y
	c.Width = dx
	c.Height = dy

	if int(sw) < p.NewWidth || int(sh) < p.NewHeight {
		newImg = p.calculateFitness(newImg, c)
	}
	return newImg
}

// Process encodes the resized image into an io.Writer interface.
// We are using the io package, since we can provide different input and output types,
// as long as they implement the io.Reader and io.Writer interface.
func (p *Processor) Process(r io.Reader, w io.Writer) error {
	var err error

	// Instantiate a new Pigo object in case the face detection option is used.
	p.PigoFaceDetector = pigo.NewPigo()

	if p.FaceDetect {
		// Unpack the binary file. This will return the number of cascade trees,
		// the tree depth, the threshold and the prediction from tree's leaf nodes.
		p.PigoFaceDetector, err = p.PigoFaceDetector.Unpack(cascadeFile)
		if err != nil {
			return fmt.Errorf("error unpacking the cascade file: %v", err)
		}
	}

	if p.NewWidth != 0 && p.NewHeight != 0 {
		resizeXY = true
	}

	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	img := p.imgToNRGBA(src)
	p.GuiDebug = image.NewNRGBA(img.Bounds())

	if len(p.MaskPath) > 0 {
		mf, err := os.Open(p.MaskPath)
		if err != nil {
			return fmt.Errorf("could not open the mask file: %v", err)
		}

		ctype, err := utils.DetectContentType(mf.Name())
		if err != nil {
			return err
		}
		if !strings.Contains(ctype.(string), "image") {
			return fmt.Errorf("the mask should be an image file")
		}

		mask, _, err := image.Decode(mf)
		if err != nil {
			return fmt.Errorf("could not decode the mask file: %v", err)
		}
		p.Mask = p.Dither(p.imgToNRGBA(mask))
		p.GuiDebug = p.Mask
	}

	if len(p.RMaskPath) > 0 {
		rmf, err := os.Open(p.RMaskPath)
		if err != nil {
			return fmt.Errorf("could not open the mask file: %v", err)
		}

		ctype, err := utils.DetectContentType(rmf.Name())
		if err != nil {
			return err
		}
		if !strings.Contains(ctype.(string), "image") {
			return fmt.Errorf("the mask should be an image file")
		}

		rmask, _, err := image.Decode(rmf)
		if err != nil {
			return fmt.Errorf("could not decode the mask file: %v", err)
		}
		p.RMask = p.Dither(p.imgToNRGBA(rmask))
		p.GuiDebug = p.RMask
	}

	if p.Preview {
		guiWidth := img.Bounds().Max.X
		guiHeight := img.Bounds().Max.Y

		if p.NewWidth > guiWidth {
			guiWidth = p.NewWidth
		}
		if p.NewHeight > guiHeight {
			guiHeight = p.NewHeight
		}
		if resizeXY {
			guiWidth = 1024
			guiHeight = 640
		}

		guiParams := struct {
			width  int
			height int
		}{
			width:  guiWidth,
			height: guiHeight,
		}
		// Lunch Gio GUI thread.
		go p.showPreview(imgWorker, errs, guiParams)
	}

	switch w := w.(type) {
	case *os.File:
		ext := filepath.Ext(w.Name())
		switch ext {
		case "", ".jpg", ".jpeg":
			res, err := resize(p, img)
			if err != nil {
				return err
			}
			return jpeg.Encode(w, res, &jpeg.Options{Quality: 100})
		case ".png":
			res, err := resize(p, img)
			if err != nil {
				return err
			}
			return png.Encode(w, res)
		case ".bmp":
			res, err := resize(p, img)
			if err != nil {
				return err
			}
			return bmp.Encode(w, res)
		case ".gif":
			g = new(gif.GIF)
			isGif = true
			_, err := resize(p, img)
			if err != nil {
				return err
			}
			return writeGifToFile(w.Name(), g)
		default:
			return errors.New("unsupported image format")
		}
	default:
		res, err := resize(p, img)
		if err != nil {
			return err
		}
		return jpeg.Encode(w, res, &jpeg.Options{Quality: 100})
	}
}

// shrink reduces the image dimension either horizontally or vertically.
func (p *Processor) shrink(c *Carver, img *image.NRGBA) (*image.NRGBA, error) {
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	c = NewCarver(width, height)

	if _, err := c.ComputeSeams(p, img); err != nil {
		fmt.Println(err)
		return nil, err
	}
	seams := c.FindLowestEnergySeams(p)
	img = c.RemoveSeam(img, seams, p.Debug)

	if len(p.MaskPath) > 0 {
		p.Mask = c.RemoveSeam(p.Mask, seams, false)
		draw.Draw(p.GuiDebug, img.Bounds(), p.Mask, image.Point{}, draw.Over)
	}
	if len(p.RMaskPath) > 0 {
		p.RMask = c.RemoveSeam(p.RMask, seams, false)
		draw.Draw(p.GuiDebug, img.Bounds(), p.RMask, image.Point{}, draw.Over)
	}

	if isGif {
		p.encodeImgToGif(c, img, g)
	}

	go func() {
		select {
		case imgWorker <- worker{
			carver: c,
			img:    img,
			debug:  p.GuiDebug,
			done:   false,
		}:
		case <-errs:
			return
		}
	}()
	return img, nil
}

// enlarge increases the image dimension either horizontally or vertically.
func (p *Processor) enlarge(c *Carver, img *image.NRGBA) (*image.NRGBA, error) {
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	c = NewCarver(width, height)

	if _, err := c.ComputeSeams(p, img); err != nil {
		return nil, err
	}
	seams := c.FindLowestEnergySeams(p)
	img = c.AddSeam(img, seams, p.Debug)

	if len(p.MaskPath) > 0 {
		p.Mask = c.AddSeam(p.Mask, seams, false)
		p.GuiDebug = p.Mask
	}
	if len(p.RMaskPath) > 0 {
		p.RMask = c.AddSeam(p.RMask, seams, false)
		p.GuiDebug = p.RMask
	}

	if isGif {
		p.encodeImgToGif(c, img, g)
	}

	go func() {
		select {
		case imgWorker <- worker{
			carver: c,
			img:    img,
			debug:  p.GuiDebug,
			done:   false,
		}:
		case <-errs:
			return
		}
	}()
	return img, nil
}

// imgToNRGBA converts any image type to *image.NRGBA with min-point at (0, 0).
func (p *Processor) imgToNRGBA(img image.Image) *image.NRGBA {
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

// encodeImgToGif encodes the provided image to a Gif file.
func (p *Processor) encodeImgToGif(c *Carver, src image.Image, g *gif.GIF) {
	dx, dy := src.Bounds().Max.X, src.Bounds().Max.Y
	dst := image.NewPaletted(image.Rect(0, 0, dx, dy), palette.Plan9)
	if p.NewHeight != 0 {
		dst = image.NewPaletted(image.Rect(0, 0, dy, dx), palette.Plan9)
	}

	if p.NewWidth > dx {
		dx += rCount
		g.Config.Width = dst.Bounds().Max.X + 1
		g.Config.Height = dst.Bounds().Max.Y + 1
	} else {
		dx -= rCount
	}
	if p.NewHeight > dx {
		dx += rCount
		g.Config.Width = dst.Bounds().Max.X + 1
		g.Config.Height = dst.Bounds().Max.Y + 1
	} else {
		dx -= rCount
	}

	if p.NewHeight != 0 {
		src = c.RotateImage270(src.(*image.NRGBA))
	}
	draw.Draw(dst, src.Bounds(), src, image.Point{}, draw.Src)
	g.Image = append(g.Image, dst)
	g.Delay = append(g.Delay, 0)
}

// writeGifToFile writes the encoded Gif file to the destination file.
func writeGifToFile(path string, g *gif.GIF) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, g)
}
