package caire

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/esimov/caire/utils"
	pigo "github.com/esimov/pigo/core"
)

// SeamCarver defines the Carve interface method, which have to be
// implemented by the Processor struct.
type SeamCarver interface {
	Carve(*image.NRGBA) (image.Image, error)
}

// maxFaceDetAttempts defines the maximum number of attempts of face detections
const maxFaceDetAttempts = 20

var (
	detAttempts    int
	isFaceDetected bool
)

var (
	sobel       *image.NRGBA
	energySeams = make([][]Seam, 0)
)

// Carver is the main entry struct having as parameters the newly generated image width, height and seam points.
type Carver struct {
	Points []float64
	Seams  []Seam
	Width  int
	Height int
}

// Seam struct contains the seam pixel coordinates.
type Seam struct {
	X int
	Y int
}

// NewCarver returns an initialized Carver structure.
func NewCarver(width, height int) *Carver {
	return &Carver{
		Points: make([]float64, width*height),
		Seams:  []Seam{},
		Width:  width,
		Height: height,
	}
}

// Get energy pixel value.
func (c *Carver) get(x, y int) float64 {
	px := x + y*c.Width
	return c.Points[px]
}

// Set energy pixel value.
func (c *Carver) set(x, y int, px float64) {
	idx := x + y*c.Width
	c.Points[idx] = px
}

// ComputeSeams compute the minimum energy level based on the following logic:
//
//   - traverse the image from the second row to the last row
//     and compute the cumulative minimum energy M for all possible
//     connected seams for each entry (i, j).
//
//   - the minimum energy level is calculated by summing up the current pixel value
//     with the minimum pixel value of the neighboring pixels from the previous row.
func (c *Carver) ComputeSeams(p *Processor, img *image.NRGBA) (*image.NRGBA, error) {
	var srcImg *image.NRGBA
	p.DebugMask = image.NewNRGBA(img.Bounds())

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	sobel = c.SobelDetector(img, float64(p.SobelThreshold))

	dets := []pigo.Detection{}

	if p.FaceDetector != nil && p.FaceDetect && detAttempts < maxFaceDetAttempts {
		var ratio float64

		if width < height {
			ratio = float64(width) / float64(height)
		} else {
			ratio = float64(height) / float64(width)
		}
		minSize := float64(utils.Min(width, height)) * ratio / 3

		// Transform the image to pixel array.
		pixels := rgbToGrayscale(img)

		cParams := pigo.CascadeParams{
			MinSize:     int(minSize),
			MaxSize:     utils.Min(width, height),
			ShiftFactor: 0.1,
			ScaleFactor: 1.1,

			ImageParams: pigo.ImageParams{
				Pixels: pixels,
				Rows:   height,
				Cols:   width,
				Dim:    width,
			},
		}
		if p.vRes {
			p.FaceAngle = 0.2
		}
		// Run the classifier over the obtained leaf nodes and return the detection results.
		// The result contains quadruplets representing the row, column, scale and detection score.
		dets = p.FaceDetector.RunCascade(cParams, p.FaceAngle)

		// Calculate the intersection over union (IoU) of two clusters.
		dets = p.FaceDetector.ClusterDetections(dets, 0.1)

		if len(dets) == 0 {
			// Retry detecting faces for a certain amount of time.
			if detAttempts < maxFaceDetAttempts {
				detAttempts++
			}
		} else {
			detAttempts = 0
			isFaceDetected = true
		}
	}

	// Traverse the pixel data of the binary file used for protecting the regions
	// which we do not want to be altered by the seam carver,
	// obtain the white patches and apply it to the sobel image.
	if len(p.MaskPath) > 0 && p.Mask != nil {
		for i := 0; i < width*height; i++ {
			x := i % width
			y := (i - x) / width

			r, g, b, _ := p.Mask.At(x, y).RGBA()
			if r>>8 == 0xff && g>>8 == 0xff && b>>8 == 0xff {
				if isFaceDetected {
					// Reduce the brightness of the mask with a small factor if human faces are detected.
					// This way we can avoid the seam carver to remove
					// the pixels inside the detected human faces.
					sobel.Set(x, y, color.RGBA{R: 225, G: 225, B: 225, A: 255})
				} else {
					sobel.Set(x, y, color.White)
				}
			}
		}
	}

	// Traverse the pixel data of the binary file used to remove the image regions
	// we do not want to be retained in the final image, obtain the white patches,
	// but this time inverse the colors to black and merge it back to the sobel image.
	if len(p.RMaskPath) > 0 && p.RMask != nil {
		for i := 0; i < width*height; i++ {
			x := i % width
			y := (i - x) / width

			r, g, b, _ := p.RMask.At(x, y).RGBA()
			// Replace the white pixels with black.
			if r>>8 == 0xff && g>>8 == 0xff && b>>8 == 0xff {
				if isFaceDetected {
					// Reduce the brightness of the mask with a small factor if human faces are detected.
					// This way we can avoid the seam carver to remove
					// the pixels inside the detected human faces.
					sobel.Set(x, y, color.RGBA{R: 25, G: 25, B: 25, A: 255})
				} else {
					sobel.Set(x, y, color.Black)
				}
				p.DebugMask.Set(x, y, color.Black)
			} else {
				p.DebugMask.Set(x, y, color.Transparent)
			}
		}
	}

	// Iterate over the detected faces and fill out the rectangles with white.
	// We need to trick the sobel detector to consider them as important image parts.
	for _, face := range dets {
		if (p.NewHeight != 0 && p.NewHeight < face.Scale) ||
			(p.NewWidth != 0 && p.NewWidth < face.Scale) {
			return nil, fmt.Errorf("%s %s",
				"cannot resize the image to the specified dimension without face deformation.\n",
				"\tRemove the face detection option in case you still wish to resize the image.")
		}
		if face.Q > 5.0 {
			scale := int(float64(face.Scale) / 1.7)
			rect := image.Rect(
				face.Col-scale,
				face.Row-scale,
				face.Col+scale,
				face.Row+scale,
			)
			draw.Draw(sobel, rect, &image.Uniform{color.White}, image.Point{}, draw.Src)
			draw.Draw(p.DebugMask, rect, &image.Uniform{color.White}, image.Point{}, draw.Src)
		}
	}

	// Increase the energy value for each of the selected seam from the seams table
	// in order to avoid picking the same seam over and over again.
	// We expand the energy level of the selected seams to have a better redistribution.
	if len(energySeams) > 0 {
		for i := 0; i < len(energySeams); i++ {
			for _, seam := range energySeams[i] {
				sobel.Set(seam.X, seam.Y, &image.Uniform{color.White})
			}
		}
	}

	if p.BlurRadius > 0 {
		srcImg = c.StackBlur(sobel, uint32(p.BlurRadius))
	} else {
		srcImg = sobel
	}

	for x := 0; x < c.Width; x++ {
		for y := 0; y < c.Height; y++ {
			r, _, _, a := srcImg.At(x, y).RGBA()
			c.set(x, y, float64(r)/float64(a))
		}
	}

	var left, middle, right float64

	// Traverse the image from top to bottom and compute the minimum energy level.
	// For each pixel in a row we compute the energy of the current pixel
	// plus the energy of one of the three possible pixels above it.
	for y := 1; y < c.Height; y++ {
		for x := 1; x < c.Width-1; x++ {
			left = c.get(x-1, y-1)
			middle = c.get(x, y-1)
			right = c.get(x+1, y-1)
			min := math.Min(math.Min(left, middle), right)
			// Set the minimum energy level.
			c.set(x, y, c.get(x, y)+min)
		}
		// Special cases: pixels are far left or far right
		left := c.get(0, y) + math.Min(c.get(0, y-1), c.get(1, y-1))
		c.set(0, y, left)
		right := c.get(0, y) + math.Min(c.get(c.Width-1, y-1), c.get(c.Width-2, y-1))
		c.set(c.Width-1, y, right)
	}
	return srcImg, nil
}

// FindLowestEnergySeams find the lowest vertical energy seam.
func (c *Carver) FindLowestEnergySeams(p *Processor) []Seam {
	// Find the lowest cost seam from the energy matrix starting from the last row.
	var (
		min = math.MaxFloat64
		px  int
	)
	seams := make([]Seam, 0)

	// Find the pixel on the last row with the minimum cumulative energy and use this as the starting pixel
	for x := 0; x < c.Width; x++ {
		seam := c.get(x, c.Height-1)
		if seam < min {
			min = seam
			px = x
		}
	}

	seams = append(seams, Seam{X: px, Y: c.Height - 1})
	var left, middle, right float64

	// Walk up in the matrix table, check the immediate three top pixels seam level
	// and add that one which has the lowest cumulative energy.
	for y := c.Height - 2; y >= 0; y-- {
		middle = c.get(px, y)
		// Leftmost seam, no child to the left
		if px == 0 {
			right = c.get(px+1, y)
			if right < middle {
				px++
			}
			// Rightmost seam, no child to the right
		} else if px == c.Width-1 {
			left = c.get(px-1, y)
			if left < middle {
				px--
			}
		} else {
			left = c.get(px-1, y)
			right = c.get(px+1, y)
			min := math.Min(math.Min(left, middle), right)

			if min == left {
				px--
			} else if min == right {
				px++
			}
		}
		seams = append(seams, Seam{X: px, Y: y})
	}

	// compare against c.Width and NOT c.Height, because the image is rotated.
	if p.NewWidth > c.Width || (p.NewHeight > 0 && p.NewHeight > c.Width) {
		// Include the currently processed energy seam into the seams table,
		// but only when an image enlargement operation is commenced.
		// We need to take this approach in order to avoid picking the same seam each time.
		energySeams = append(energySeams, seams)
	}
	return seams
}

// RemoveSeam remove the least important columns based on the stored energy (seams) level.
func (c *Carver) RemoveSeam(img *image.NRGBA, seams []Seam, debug bool) *image.NRGBA {
	bounds := img.Bounds()
	// Reduce the image width with one pixel on each iteration.
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx()-1, bounds.Dy()))

	for _, seam := range seams {
		y := seam.Y
		for x := 0; x < bounds.Max.X; x++ {
			if seam.X == x {
				if debug {
					c.Seams = append(c.Seams, Seam{X: x, Y: y})
				}
			} else if seam.X < x {
				dst.Set(x-1, y, img.At(x, y))
			} else {
				dst.Set(x, y, img.At(x, y))
			}
		}
	}
	return dst
}

// AddSeam add a new seam.
func (c *Carver) AddSeam(img *image.NRGBA, seams []Seam, debug bool) *image.NRGBA {
	var (
		lr, lg, lb uint32
		rr, rg, rb uint32
	)

	bounds := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx()+1, bounds.Dy()))

	for _, seam := range seams {
		y := seam.Y
		for x := 0; x < bounds.Max.X; x++ {
			if seam.X == x {
				if debug {
					c.Seams = append(c.Seams, Seam{X: x, Y: y})
				}
				if x > 0 && x != bounds.Max.X {
					lr, lg, lb, _ = img.At(x-1, y).RGBA()
				} else {
					lr, lg, lb, _ = img.At(x, y).RGBA()
				}

				if x < bounds.Max.X-1 {
					rr, rg, rb, _ = img.At(x+1, y).RGBA()
				} else if x == bounds.Max.X {
					rr, rg, rb, _ = img.At(x, y).RGBA()
				}

				// calculate the average color of the neighboring pixels
				avr, avg, avb := (lr+rr)>>1, (lg+rg)>>1, (lb+rb)>>1
				dst.Set(x, y, color.RGBA{uint8(avr >> 8), uint8(avg >> 8), uint8(avb >> 8), 0xff})
				dst.Set(x+1, y, img.At(x, y))
			} else if seam.X < x {
				dst.Set(x, y, img.At(x-1, y))
				dst.Set(x+1, y, img.At(x, y))
			} else {
				dst.Set(x, y, img.At(x, y))
			}
		}
	}

	return dst
}
