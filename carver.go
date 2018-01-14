package caire

import (
	"image"
	"math"
	"os"
	"image/png"
//	"image/color"
)

type DPTable struct {
	width int
	height int
	table []float64
}

type Seam struct {
	X int
	Y int
}

// Get energy pixel value
func (dpt *DPTable) get(x, y int) float64 {
	px := x + y * dpt.width
	return dpt.table[px]
}

// Set energy pixel value
func (dpt *DPTable) set(x, y int, px float64) {
	idx := x + y * dpt.width
	dpt.table[idx] = px
}

// Compute the minimum energy level based on the following logic:
// 	- traverse the image from the second row to the last row
// 	  and compute the cumulative minimum energy M for all possible
//	  connected seams for each entry (i, j).
//
//	- the minimum energy level is calculated by summing up the current pixel value
// 	  with the minimum pixel value of the neighboring pixels from the previous row.
func (dpt *DPTable) computeSeams(img *image.NRGBA, threshold, blur int) []float64 {
	var src *image.NRGBA
	bounds := img.Bounds()
	iw, ih := bounds.Dx(), bounds.Dy()
	sobel := SobelFilter(Grayscale(img), float64(threshold))

	if blur > 0 {
		src = Stackblur(sobel, uint32(iw), uint32(ih), uint32(blur))
	} else {
		src = sobel
	}
	for x := 0; x < dpt.width; x++ {
		for y := 0; y < dpt.height; y++ {
			r, _, _, a := src.At(x, y).RGBA()
			dpt.set(x, y, float64(r) / float64(a))
		}
	}

	// Compute the minimum energy level and set the resulting value into dpt table.
	for x := 0; x < dpt.width; x++ {
		for y := 1; y < dpt.height; y++ {
			var left, middle, right float64
			left, right = math.MaxFloat64, math.MaxFloat64

			// Do not compute edge cases: pixels are far left.
			if x > 0 {
				left = dpt.get(x-1, y-1)
			}
			middle = dpt.get(x, y-1)
			// Do not compute edge cases: pixels are far right.
			if x < dpt.width-1 {
				right = dpt.get(x+1, y-1)
			}
			// Obtain the minimum pixel value
			min := math.Min(math.Min(left, middle), right)
			dpt.set(x, y, dpt.get(x, y) + min)
		}
	}
	return dpt.table
}

// Find the lowest vertical energy seam.
func (dpt *DPTable) findLowestEnergySeams() []Seam {
	// Find the lowest cost seam from the energy matrix starting from the last row.
	var min float64 = math.MaxFloat64
	var px int
	seams := make([]Seam, 0)

	// Find the pixel on the last row with the minimum cumulative energy and use this as the starting pixel
	for x := 0; x < dpt.width; x++ {
		seam := dpt.get(x, dpt.height-1)
		if seam < min && seam > 0 {
			min = seam
			px = x
		}
	}
	seams = append(seams, Seam{X: px, Y: dpt.height-1})
	var left, middle, right float64

	// Walk up in the matrix table,
	// check the immediate three top pixel seam level and
	// add add the one which has the lowest cumulative energy.
	for y := dpt.height-2; y >= 0; y-- {
		left, right = math.MaxFloat64, math.MaxFloat64
		middle = dpt.get(px, y)
		// Leftmost seam, no child to the left
		if px == 0 {
			right = dpt.get(px+1, y)
			middle = dpt.get(px, y)
			if right < middle {
				px += 1
			}
		// Rightmost seam, no child to the right
		} else if px == dpt.width-1 {
			left = dpt.get(px-1, y)
			middle = dpt.get(px, y)
			if left < middle {
				px -= 1
			}
		} else {
			left = dpt.get(px-1, y)
			middle = dpt.get(px, y)
			right = dpt.get(px+1, y)
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

// Remove image pixels based on energy seams level
func (dpt *DPTable) removeSeam(img *image.NRGBA, seams []Seam) *image.NRGBA {
	bounds := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx()-1, bounds.Dy()))

	for _, seam := range seams {
		y := seam.Y
		for x := 0; x < bounds.Max.X; x++ {
			if seam.X == x {
				continue
				//dst.Set(x-1, y, color.RGBA{255, 0, 0, 255})
			} else if seam.X < x {
				dst.Set(x-1, y, img.At(x, y))
			} else {
				dst.Set(x, y, img.At(x, y))
			}
		}
	}
	return dst
}


func Process(src *image.NRGBA, sobel *image.NRGBA, output string, threshold, blur int) (*os.File, error) {
	for x := 0; x < 180; x++ {
		width, height := src.Bounds().Max.X, src.Bounds().Max.Y
		dpt := &DPTable{
			width,
			height,
			make([]float64, width*height),
		}
		dpt.computeSeams(src, threshold, blur)
		seams := dpt.findLowestEnergySeams()
		src = dpt.removeSeam(src, seams)
	}

	fq, err := os.Create(output)
	if err != nil {
		return nil, err
	}
	defer fq.Close()

	if err = png.Encode(fq, src); err != nil {
		return nil, err
	}
	return fq, nil
}