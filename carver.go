package caire

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	pigo "github.com/esimov/pigo/core"
)

var usedSeams []UsedSeams

// TempImage temporary image file.
var TempImage = fmt.Sprintf("%d.jpg", time.Now().Unix())

// Carver is the main entry struct having as parameters the newly generated image width, height and seam points.
type Carver struct {
	Width  int
	Height int
	Points []float64
}

// UsedSeams contains the already generated seams.
type UsedSeams struct {
	ActiveSeam []ActiveSeam
}

// ActiveSeam contains the current seam position and color.
type ActiveSeam struct {
	Seam
	Pix color.Color
}

// Seam struct contains the seam pixel coordinates.
type Seam struct {
	X int
	Y int
}

// NewCarver returns an initialized Carver structure.
func NewCarver(width, height int) *Carver {
	return &Carver{
		width,
		height,
		make([]float64, width*height),
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
// 	- traverse the image from the second row to the last row
// 	  and compute the cumulative minimum energy M for all possible
//	  connected seams for each entry (i, j).
//
//	- the minimum energy level is calculated by summing up the current pixel value
// 	  with the minimum pixel value of the neighboring pixels from the previous row.
func (c *Carver) ComputeSeams(img *image.NRGBA, p *Processor) {
	var srcImg *image.NRGBA
	newImg := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(newImg, newImg.Bounds(), img, image.ZP, draw.Src)

	// Replace the energy map seam values with the stored pixel values each time we add a new seam.
	for _, seam := range usedSeams {
		for _, as := range seam.ActiveSeam {
			newImg.Set(as.X, as.Y, as.Pix)
		}
	}
	sobel := SobelFilter(Grayscale(newImg), float64(p.SobelThreshold))

	if p.FaceDetect {
		if len(p.Classifier) == 0 {
			log.Fatal("Please provide a face classifier!")
		}

		cascadeFile, err := ioutil.ReadFile(p.Classifier)
		if err != nil {
			log.Fatalf("Error reading the cascade file: %v", err)
		}

		tmpImg, err := os.OpenFile(TempImage, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatalf("Cannot access temporary image file: %v", err)
		}

		if err := jpeg.Encode(tmpImg, img, &jpeg.Options{Quality: 100}); err != nil {
			log.Fatalf("Cannot encode temporary image file: %v", err)
		}

		src, err := pigo.GetImage(TempImage)
		if err != nil {
			log.Fatalf("Cannot open the image file: %v", err)
		}

		pixels := pigo.RgbToGrayscale(src)
		cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

		cParams := pigo.CascadeParams{
			MinSize:     100,
			MaxSize:     int(math.Max(float64(cols), float64(rows))),
			ShiftFactor: 0.1,
			ScaleFactor: 1.1,

			ImageParams: pigo.ImageParams{
				Pixels: pixels,
				Rows:   rows,
				Cols:   cols,
				Dim:    cols,
			},
		}

		pigo := pigo.NewPigo()
		// Unpack the binary file. This will return the number of cascade trees,
		// the tree depth, the threshold and the prediction from tree's leaf nodes.
		classifier, err := pigo.Unpack(cascadeFile)
		if err != nil {
			log.Fatalf("Error reading the cascade file: %v\n", err)
		}

		// Run the classifier over the obtained leaf nodes and return the detection results.
		// The result contains quadruplets representing the row, column, scale and detection score.
		faces := classifier.RunCascade(cParams, p.FaceAngle)

		// Calculate the intersection over union (IoU) of two clusters.
		faces = classifier.ClusterDetections(faces, 0.2)

		// Range over all the detected faces and draw a white rectangle mask over each of them.
		// We need to trick the sobel detector to consider them as important image parts.
		for _, face := range faces {
			if face.Q > 5.0 {
				rect := image.Rect(
					face.Col-face.Scale/2,
					face.Row-face.Scale/2,
					face.Col+face.Scale/2,
					face.Row+face.Scale/2,
				)
				draw.Draw(sobel, rect, &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)
			}
		}

		// Capture CTRL-C signal and remove the generated temporary image.
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			for range c {
				RemoveTempImage(TempImage)
				os.Exit(1)
			}
		}()
	}

	if p.BlurRadius > 0 {
		srcImg = StackBlur(sobel, uint32(p.BlurRadius))
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
}

// FindLowestEnergySeams find the lowest vertical energy seam.
func (c *Carver) FindLowestEnergySeams() []Seam {
	// Find the lowest cost seam from the energy matrix starting from the last row.
	var min = math.MaxFloat64
	var px int
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

	// Walk up in the matrix table, check the immediate three top pixel seam level
	// and add the one which has the lowest cumulative energy.
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
					dst.Set(x-1, y, color.RGBA{255, 0, 0, 255})
				}
				continue
			} else if seam.X < x {
				dst.Set(x-1, y, img.At(x, y))
			} else {
				dst.Set(x, y, img.At(x, y))
			}
		}
	}
	return dst
}

// AddSeam add new seam.
func (c *Carver) AddSeam(img *image.NRGBA, seams []Seam, debug bool) *image.NRGBA {
	var currentSeam []ActiveSeam
	var lr, lg, lb uint32
	var rr, rg, rb uint32
	var py int

	bounds := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx()+1, bounds.Dy()))

	for _, seam := range seams {
		y := seam.Y
		for x := 0; x < bounds.Max.X; x++ {
			if seam.X == x {
				if debug == true {
					dst.Set(x, y, color.RGBA{255, 0, 0, 255})
					continue
				}
				// Calculate the current seam pixel color by averaging the neighboring pixels color.
				if y > 0 {
					py = y - 1
				} else {
					py = y
				}

				if x > 0 {
					lr, lg, lb, _ = img.At(x-1, py).RGBA()
				} else {
					lr, lg, lb, _ = img.At(x, y).RGBA()
				}

				if y < bounds.Max.Y-1 {
					py = y + 1
				} else {
					py = y
				}

				if x < bounds.Max.X-1 {
					rr, rg, rb, _ = img.At(x+1, py).RGBA()
				} else {
					rr, rg, rb, _ = img.At(x, y).RGBA()
				}
				alr, alg, alb := (lr+rr)/2, (lg+rg)/2, (lb+rb)/2
				dst.Set(x, y, color.RGBA{uint8(alr >> 8), uint8(alg >> 8), uint8(alb >> 8), 255})

				// Append the current seam position and color to the existing seams.
				// To avoid picking the same optimal seam over and over again,
				// each time we detect an optimal seam we assign a large positive value
				// to the corresponding pixels in the energy map.
				// We will increase the seams weight by duplicating the pixel value.
				currentSeam = append(currentSeam,
					ActiveSeam{Seam{x + 1, y},
						color.RGBA{
							R: uint8((alr + alr) >> 8),
							G: uint8((alg + alg) >> 8),
							B: uint8((alb + alb) >> 8),
							A: 255,
						},
					})
			} else if seam.X < x {
				dst.Set(x, y, img.At(x-1, y))
				dst.Set(x+1, y, img.At(x, y))
			} else {
				dst.Set(x, y, img.At(x, y))
			}
		}
	}
	usedSeams = append(usedSeams, UsedSeams{currentSeam})
	return dst
}

// RotateImage90 rotate the image by 90 degree counter clockwise.
func (c *Carver) RotateImage90(src *image.NRGBA) *image.NRGBA {
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

// RotateImage270 rotate the image by 270 degree counter clockwise.
func (c *Carver) RotateImage270(src *image.NRGBA) *image.NRGBA {
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

// RemoveTempImage removes the temporary image generated during face detection process.
func RemoveTempImage(tmpImage string) {
	// Remove temporary image file.
	if _, err := os.Stat(tmpImage); err == nil {
		os.Remove(tmpImage)
	}
}
