package caire

import (
	"image"
	"fmt"
	"math"
)

type DPTable struct {
	width int
	height int
	table []float64
}

type Seam []Point

type Point struct {
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
func (dpt *DPTable) ComputeSeams(img *image.NRGBA) []float64 {
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			r, _, _, a := img.At(x, y).RGBA()
			dpt.set(x, y, float64(r) / float64(a))
		}
	}

	computeEnergyLevel := func(x, y int) {
		var left, middle, right float64
		left, right = dpt.get(x, y-1), dpt.get(x, y-1)

		// Do not compute edge cases: pixels are far left.
		if x > 0 {
			left = dpt.get(x-1, y-1)
		}
		middle = dpt.get(x, y-1)
		// Do not compute edge cases: pixels are far right.
		if x < width-1 {
			right = dpt.get(x+1, y-1)
		}
		// Obtain the minimum pixel value
		min := math.Min(math.Min(left, middle), right)
		dpt.set(x, y, dpt.get(x, y) + min)
	}

	for x := 0; x < width; x++ {
		for y := 1; y < height; y++ {
			computeEnergyLevel(x, y)
		}
	}

	//fmt.Println(dpt.table)
	return dpt.table
}

// Find the lowest vertical energy seam.
func (dpt *DPTable) FindLowestEnergySeams() []Point {
	// Find the lowest cost seam from the energy matrix starting from the last row.
	var min float64 = math.MaxFloat64
	var px int
	seams := make([]Point, 0)

	// Find the lowest seam from the bottom row
	for x := 0; x < dpt.width; x++ {
		seam := dpt.get(x, dpt.height-1)
		if seam < min {
			min = seam
			px = x
		}
	}
	seams = append(seams, Point{X: px, Y: dpt.height-1})

	// Walk up in the matrix table,
	// check the immediate three top pixel seam level and
	// add add the one which has the lowest cumulative energy.
	for y := dpt.height-2; y >= 0; y-- {
		left, center, right := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
		// Leftmost seam, no child to the left
		if px == 0 {
			right = dpt.get(px+1, y)
			center = dpt.get(px, y)
			if right < center {
				px += 1
			}
		// Rightmost seam, no child to the right
		} else if px == dpt.width-1 {
			left = dpt.get(px-1, y)
			center = dpt.get(px, y)
			if left < center {
				px -= 1
			}
		} else {
			left = dpt.get(px-1, y)
			center = dpt.get(px, y)
			right = dpt.get(px+1, y)
			min := math.Min(math.Min(left, center), right)

			if min == left {
				px -= 1
			} else if min == right {
				px += 1
			}
		}
		seams = append(seams, Point{X: px, Y: y})
	}
	fmt.Println(seams)
	return seams
}

// Remove image pixels based on energy seams level
func (dpt *DPTable) RemoveSeams(seams Seam) {

}