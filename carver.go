package caire

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"math"
)

var usedSeams []UsedSeams

type Carver struct {
	Width  int
	Height int
	Points []float64
}

// Struct containing the generated seams.
type UsedSeams struct {
	ActiveSeam []ActiveSeam
}

// Struct containing the current seam color and position.
type ActiveSeam struct {
	Seam
	Pix color.Color
}

// Seam struct containing the pixel coordinate values.
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

// Compute the minimum energy level based on the following logic:
// 	- traverse the image from the second row to the last row
// 	  and compute the cumulative minimum energy M for all possible
//	  connected seams for each entry (i, j).
//
//	- the minimum energy level is calculated by summing up the current pixel value
// 	  with the minimum pixel value of the neighboring pixels from the previous row.
func (c *Carver) ComputeSeams(img *image.NRGBA, p *Processor) []float64 {
	var src *image.NRGBA
	bounds := img.Bounds()
	iw, ih := bounds.Dx(), bounds.Dy()
	newImg := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(newImg, newImg.Bounds(), img, image.ZP, draw.Src)

	// Replace the energy map seam values with the stored pixel values each time we add a new seam.
	for _, seam := range usedSeams {
		for _, as := range seam.ActiveSeam {
			newImg.Set(as.X, as.Y, as.Pix)
		}
	}
	sobel := SobelFilter(Grayscale(newImg), float64(p.SobelThreshold))

	if p.BlurRadius > 0 {
		src = Stackblur(sobel, uint32(iw), uint32(ih), uint32(p.BlurRadius))
	} else {
		src = sobel
	}
	for x := 0; x < c.Width; x++ {
		for y := 0; y < c.Height; y++ {
			r, _, _, a := src.At(x, y).RGBA()
			c.set(x, y, float64(r)/float64(a))
		}
	}

	left, middle, right := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64

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
	return c.Points
}

// Find the lowest vertical energy seam.
func (c *Carver) FindLowestEnergySeams() []Seam {
	// Find the lowest cost seam from the energy matrix starting from the last row.
	var min float64 = math.MaxFloat64
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

	// Walk up in the matrix table,
	// check the immediate three top pixel seam level and
	// add add the one which has the lowest cumulative energy.
	for y := c.Height - 2; y >= 0; y-- {
		left, right = math.MaxFloat64, math.MaxFloat64
		middle = c.get(px, y)
		// Leftmost seam, no child to the left
		if px == 0 {
			right = c.get(px+1, y)
			middle = c.get(px, y)
			if right < middle {
				px += 1
			}
			// Rightmost seam, no child to the right
		} else if px == c.Width-1 {
			left = c.get(px-1, y)
			middle = c.get(px, y)
			if left < middle {
				px -= 1
			}
		} else {
			left = c.get(px-1, y)
			middle = c.get(px, y)
			right = c.get(px+1, y)
			min := math.Min(math.Min(left, middle), right)

			if min == left {
				px -= 1
			} else if min == right {
				px += 1
			}
		}
		seams = append(seams, Seam{X: px, Y: y})
	}
	return seams
}

// Remove the least important columns based on the stored energy seams level.
func (c *Carver) RemoveSeam(img *image.NRGBA, seams []Seam, debug bool) *image.NRGBA {
	bounds := img.Bounds()
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

// Add new seam.
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

// Rotate image by 90 degree counter clockwise
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

// Rotate image by 270 degree counter clockwise
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
