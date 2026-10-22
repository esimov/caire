package caire

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io"
	"math"

	"github.com/disintegration/imaging"
	"github.com/esimov/caire/utils"
	pigo "github.com/esimov/pigo/core"
)

//go:embed data/facefinder
var cascadeFile []byte

var _ SeamCarver = (*Processor)(nil)

// worker struct contains all the information needed for transferring the resized image to the Gio GUI.
type worker struct {
	img   *image.NRGBA
	mask  *image.NRGBA
	seams []Seam
	done  bool
}

// shrinkFn is a generic function used to shrink an image.
type shrinkFn func(*image.NRGBA) (*image.NRGBA, error)

// enlargeFn is a generic function used to enlarge an image.
type enlargeFn func(*image.NRGBA) (*image.NRGBA, error)

// Processor options
type Processor struct {
	FaceAngle      float64
	SeamColor      string
	MaskPath       string
	RMaskPath      string
	ShapeType      ShapeType
	SobelThreshold int
	BlurRadius     int
	NewWidth       int
	NewHeight      int
	FaceDetector   *pigo.Pigo
	Spinner        *utils.Spinner
	Mask           *image.NRGBA
	RMask          *image.NRGBA
	DebugMask      *image.NRGBA
	Percentage     bool
	Square         bool
	Debug          bool
	Preview        bool
	FaceDetect     bool
	vRes           bool

	// imgWorker transfers the intermediate images to the GUI preview.
	// Both channels are nil when the preview window is not running.
	imgWorker chan worker
	errs      chan error

	// State of the current resize operation. Resize resets it on every call.
	resizeXY       bool     // the image is resized both vertically and horizontally
	energySeams    [][]Seam // seams already added while enlarging, so the next pass picks a different one
	detAttempts    int
	isFaceDetected bool
}

// Carve is the main entry point for the image resize operation.
// The new image can be resized either horizontally or vertically (or both).
// Depending on the provided options the image can be either reduced or enlarged.
func (p *Processor) Resize(img *image.NRGBA) (image.Image, error) {
	var (
		newImg    image.Image
		newWidth  int
		newHeight int
		pw, ph    int
		err       error

		shrinkHorizFn  shrinkFn
		shrinkVertFn   shrinkFn
		enlargeHorizFn enlargeFn
		enlargeVertFn  enlargeFn
	)
	c := NewCarver(img.Bounds().Dx(), img.Bounds().Dy())

	// Reset the state left behind by a previous run of the same Processor.
	p.resizeXY = p.NewWidth != 0 && p.NewHeight != 0
	p.energySeams = nil
	p.detAttempts = 0
	p.isFaceDetected = false

	// The percentage and square options are converted to pixel sizes by
	// updating NewWidth and NewHeight. Restore the requested values on return,
	// otherwise the next image processed with the same Processor (as in
	// directory mode) takes the pixel size for a percentage.
	reqWidth, reqHeight := p.NewWidth, p.NewHeight
	defer func() { p.NewWidth, p.NewHeight = reqWidth, reqHeight }()

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
	// function intermittently up until the desired dimension is reached.
	// I opted for this solution instead of resizing the image sequentially,
	// to avoid resulting visual artefacts, since this way
	// the horizontal and vertical seams are merged together seamlessly.
	shrinkHorizFn = func(img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = false
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
		if dx > p.NewWidth {
			img, err = p.shrink(img)
			if err != nil {
				return nil, err
			}
			if p.NewHeight > 0 && p.NewHeight != dy {
				if p.NewHeight <= dy {
					img, err = shrinkVertFn(img)
					if err != nil {
						return nil, err
					}
				} else {
					img, err = enlargeVertFn(img)
					if err != nil {
						return nil, err
					}
				}
			} else {
				img, err = shrinkHorizFn(img)
				if err != nil {
					return nil, err
				}
			}
		}
		return img, nil
	}

	// enlargeHorizFn calls itself recursively to enlarge the image horizontally.
	enlargeHorizFn = func(img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = false
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
		if dx < p.NewWidth {
			img, err = p.enlarge(img)
			if err != nil {
				return nil, err
			}
			if p.NewHeight > 0 && p.NewHeight != dy {
				if p.NewHeight <= dy {
					img, err = shrinkVertFn(img)
					if err != nil {
						return nil, err
					}
				} else {
					img, err = enlargeVertFn(img)
					if err != nil {
						return nil, err
					}
				}
			} else {
				img, err = enlargeHorizFn(img)
				if err != nil {
					return nil, err
				}
			}
		}
		return img, nil
	}

	// shrinkVertFn calls itself recursively to shrink the image vertically.
	shrinkVertFn = func(img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = true
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()

		// If the image is resized both horizontally and vertically we need
		// to rotate the image each time when invoking the shrink function.
		// Otherwise we rotate the image only once, right before calling this function.
		if p.resizeXY {
			dx, dy = img.Bounds().Dy(), img.Bounds().Dx()
			img = rotateImage90(img)
		}
		if dx > p.NewHeight {
			img, err = p.shrink(img)
			if err != nil {
				return nil, err
			}
			if p.resizeXY {
				img = rotateImage270(img)
			}
			if p.NewWidth > 0 && p.NewWidth != dy {
				if p.NewWidth <= dy {
					img, err = shrinkHorizFn(img)
					if err != nil {
						return nil, err
					}
				} else {
					img, err = enlargeHorizFn(img)
					if err != nil {
						return nil, err
					}
				}
			} else {
				img, err = shrinkVertFn(img)
				if err != nil {
					return nil, err
				}
			}
		} else {
			if p.resizeXY {
				img = rotateImage270(img)
			}
		}
		return img, nil
	}

	// enlargeVertFn calls itself recursively to enlarge the image vertically.
	enlargeVertFn = func(img *image.NRGBA) (*image.NRGBA, error) {
		p.vRes = true
		dx, dy := img.Bounds().Dx(), img.Bounds().Dy()

		if p.resizeXY {
			dx, dy = img.Bounds().Dy(), img.Bounds().Dx()
			img = rotateImage90(img)
		}
		if dx < p.NewHeight {
			img, err = p.enlarge(img)
			if err != nil {
				return nil, err
			}
			if p.resizeXY {
				img = rotateImage270(img)
			}
			if p.NewWidth > 0 && p.NewWidth != dy {
				if p.NewWidth <= dy {
					img, err = shrinkHorizFn(img)
					if err != nil {
						return nil, err
					}
				} else {
					img, err = enlargeHorizFn(img)
					if err != nil {
						return nil, err
					}
				}
			} else {
				img, err = enlargeVertFn(img)
				if err != nil {
					return nil, err
				}
			}
		} else {
			if p.resizeXY {
				img = rotateImage270(img)
			}
		}
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
			img, err = enlargeHorizFn(img)
			if err != nil {
				return nil, err
			}
		} else {
			img, err = shrinkHorizFn(img)
			if err != nil {
				return nil, err
			}
		}
	}

	// Run the carver function if the desired image height is not identical with the rescaled image height.
	if newHeight > 0 && p.NewHeight != c.Height {
		if !p.resizeXY {
			img = rotateImage90(img)

			if p.Mask != nil {
				p.Mask = rotateImage90(p.Mask)
			}
			if p.RMask != nil {
				p.RMask = rotateImage90(p.RMask)
			}
		}
		if p.NewHeight > c.Height {
			img, err = enlargeVertFn(img)
			if err != nil {
				return nil, err
			}
		} else {
			img, err = shrinkVertFn(img)
			if err != nil {
				return nil, err
			}
		}
		if !p.resizeXY {
			img = rotateImage270(img)

			if p.Mask != nil {
				p.Mask = rotateImage270(p.Mask)
			}
			if p.RMask != nil {
				p.RMask = rotateImage270(p.RMask)
			}
		}
	}

	// Signal that the process is done and no more data is sent through the channel.
	p.notifyPreview(worker{done: true})

	return img, nil
}

// notifyPreview hands the intermediate image over to the GUI preview.
// It returns immediately when no preview window is running, otherwise the
// goroutine would wait forever for a reader and keep the image in memory.
func (p *Processor) notifyPreview(w worker) {
	if p.imgWorker == nil {
		return
	}
	go func() {
		select {
		case p.imgWorker <- w:
		case <-p.errs:
		}
	}()
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
		if p.Mask != nil {
			p.Mask = imaging.Resize(p.Mask, 0, int(sw), imaging.Lanczos)
		}
		if p.RMask != nil {
			p.RMask = imaging.Resize(p.RMask, 0, int(sw), imaging.Lanczos)
		}
	} else {
		newImg = imaging.Resize(img, 0, int(sh), imaging.Lanczos)
		if p.Mask != nil {
			p.Mask = imaging.Resize(p.Mask, 0, int(sh), imaging.Lanczos)
		}
		if p.RMask != nil {
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

	if p.FaceDetect {
		// Instantiate a new Pigo object in case the face detection option is used.
		p.FaceDetector = pigo.NewPigo()

		// Unpack the binary file. This will return the number of cascade trees,
		// the tree depth, the threshold and the prediction from tree's leaf nodes.
		p.FaceDetector, err = p.FaceDetector.Unpack(cascadeFile)
		if err != nil {
			return fmt.Errorf("error unpacking the cascade file: %v", err)
		}
	}

	// Resize sets this flag as well, but the preview window size depends on it.
	p.resizeXY = p.NewWidth != 0 && p.NewHeight != 0

	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	img := imgToNRGBA(src)

	if p.Preview {
		guiWidth := img.Bounds().Max.X
		guiHeight := img.Bounds().Max.Y

		if p.NewWidth > guiWidth {
			guiWidth = p.NewWidth
		}
		if p.NewHeight > guiHeight {
			guiHeight = p.NewHeight
		}
		if p.resizeXY {
			guiWidth = int(maxScreenX)
			guiHeight = int(maxScreenY)
		}

		guiWindow := struct {
			width  int
			height int
		}{
			width:  guiWidth,
			height: guiHeight,
		}
		p.imgWorker = make(chan worker)
		p.errs = make(chan error)

		// Lunch Gio GUI thread.
		go p.showPreview(p.imgWorker, p.errs, guiWindow)
	}

	return encodeImg(p, w, img)
}

// shrink reduces the image dimension either horizontally or vertically.
func (p *Processor) shrink(img *image.NRGBA) (*image.NRGBA, error) {
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	c := NewCarver(width, height)

	if _, err := c.ComputeSeams(p, img); err != nil {
		return nil, err
	}
	seams := c.FindLowestEnergySeams(p)
	img = c.RemoveSeam(img, seams, p.Debug)

	if p.Mask != nil {
		p.Mask = c.RemoveSeam(p.Mask, seams, false)
		draw.Draw(p.DebugMask, img.Bounds(), p.Mask, image.Point{}, draw.Over)
	}

	if p.RMask != nil {
		p.RMask = c.RemoveSeam(p.RMask, seams, false)
		draw.Draw(p.DebugMask, img.Bounds(), p.RMask, image.Point{}, draw.Over)
	}

	p.notifyPreview(worker{
		img:   img,
		mask:  p.DebugMask,
		seams: c.Seams,
		done:  false,
	})
	return img, nil
}

// enlarge increases the image dimension either horizontally or vertically.
func (p *Processor) enlarge(img *image.NRGBA) (*image.NRGBA, error) {
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	c := NewCarver(width, height)

	if _, err := c.ComputeSeams(p, img); err != nil {
		return nil, err
	}
	seams := c.FindLowestEnergySeams(p)
	img = c.AddSeam(img, seams, p.Debug)

	if p.Mask != nil {
		p.Mask = c.AddSeam(p.Mask, seams, false)
		p.DebugMask = p.Mask
	}

	if p.RMask != nil {
		p.RMask = c.AddSeam(p.RMask, seams, false)
		p.DebugMask = p.RMask
	}

	p.notifyPreview(worker{
		img:   img,
		mask:  p.DebugMask,
		seams: c.Seams,
		done:  false,
	})
	return img, nil
}
