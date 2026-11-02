// Package textdetector is used to identify and group characters from an image.
// The implementation is based on this article:
// https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/201020CVPR20TextDetection.pdf
// The scope of this package is to avoid removing seams from certain areas (like on face detection).
// After successfully detecting text areas in the provided image we can manipulate these areas by
// applying a  high threshold (bright color) on the detected rectangles.
// This package could be integrated into the seam removal logic to avoid
// removing seams from an image where the texts should be preserved.
package textdetector

import (
	"fmt"
	"math"
)

const (
	GaussianSigma = 1.5
	LowThreshold  = 90
	HighThreshold = 100
)

// Canny holds the minimum required components for the edge detection.
// Width and Height represents the width and height of the source image, which
// is passed downstream over the whole edge detection operation.
type Canny struct {
	width         int
	height        int
	gaussianSigma float64
	lowThreshold  float64
	highThreshold float64
}

type point struct {
	x int
	y int
}

type gradient struct {
	x float64
	y float64
}

type cannyPixel uint8

const (
	// Type of the resulted canny pixel.
	cannyNone cannyPixel = iota
	cannyWeak
	cannyStrong
)

var _ EdgeDetector = (*Canny)(nil)

type CannyOption func(*Canny) error

// NewCannyDetector creates a new Canny edge detector with the default values
// if no additional constructor parameters are provided,
// otherwise it creates a new constructor with the provided options.
func NewCannyDetector(width, height int, options ...CannyOption) (*Canny, error) {
	canny := &Canny{
		width:  width,
		height: height,
	}

	if options == nil {
		options = append(options, WithDefaultOptions(Canny{
			gaussianSigma: GaussianSigma,
			lowThreshold:  LowThreshold,
			highThreshold: HighThreshold,
		}))
	}

	for _, opt := range options {
		if err := opt(canny); err != nil {
			return nil, fmt.Errorf("failed to apply the detector options: %w", err)
		}
	}

	if canny.lowThreshold >= canny.highThreshold {
		canny.lowThreshold = canny.highThreshold * 0.5
	}

	return canny, nil
}

// Detect detect edges in a grayscaled converted image passed as flattened pixel buffer.
func (c *Canny) Detect(gray []uint8) ([]bool, []gradient, error) {
	blurred, err := c.GaussianBlur(gray, c.gaussianSigma)
	if err != nil {
		return nil, nil, fmt.Errorf("error on image bluring opreration: %w", err)
	}

	gradients, magnitudes := c.applyGradients(blurred)
	suppressed := c.nonMaximumSuppression(magnitudes, gradients)
	cannyPixels := c.normalize(suppressed, 255)

	edges := c.removeWeakEdges(cannyPixels)

	return edges, gradients, nil
}

func (c *Canny) applyGradients(input []uint8) ([]gradient, []float64) {
	gradients := make([]gradient, c.width*c.height)
	magnitudes := make([]float64, c.width*c.height)

	for y := 1; y < c.height-1; y++ {
		previousRow := (y - 1) * c.width
		currentRow := y * c.width
		nextRow := (y + 1) * c.width

		for x := 1; x < c.width-1; x++ {
			gx := -float64(input[previousRow+x-1]) +
				float64(input[previousRow+x+1]) -
				2*float64(input[currentRow+x-1]) +
				2*float64(input[currentRow+x+1]) -
				float64(input[nextRow+x-1]) +
				float64(input[nextRow+x+1])

			gy := -float64(input[previousRow+x-1]) -
				2*float64(input[previousRow+x]) -
				float64(input[previousRow+x+1]) +
				float64(input[nextRow+x-1]) +
				2*float64(input[nextRow+x]) +
				float64(input[nextRow+x+1])

			magnitude := math.Sqrt(gx*gx + gy*gy)
			index := currentRow + x

			magnitudes[index] = magnitude

			if magnitude > 0 {
				gradients[index] = gradient{
					x: gx / magnitude,
					y: gy / magnitude,
				}
			}
		}
	}

	return gradients, magnitudes
}

func (c *Canny) nonMaximumSuppression(magnitudes []float64, gradients []gradient) []float64 {
	result := make([]float64, c.width*c.height)

	for y := 1; y < c.height-1; y++ {
		rowOffset := y * c.width

		for x := 1; x < c.width-1; x++ {
			index := rowOffset + x
			current := magnitudes[index]

			if current <= 0 {
				continue
			}

			direction := gradients[index]

			forward := c.bilinearInterpolation(magnitudes,
				float64(x)+direction.x,
				float64(y)+direction.y,
			)

			backward := c.bilinearInterpolation(magnitudes,
				float64(x)-direction.x,
				float64(y)-direction.y,
			)

			if current >= forward && current >= backward {
				result[index] = current
			}
		}
	}

	return result
}

// bilinearInterpolation performs a bilinear interpolation over the pixel magnitudes.
// This produces more accurate and therefore thinner and better-localized edges.
func (c *Canny) bilinearInterpolation(input []float64, x float64, y float64) float64 {
	if x < 0 || y < 0 || x > float64(c.width-1) || y > float64(c.height-1) {
		return 0
	}

	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))

	x1 := min(x0+1, c.width-1)
	y1 := min(y0+1, c.height-1)

	dx := x - float64(x0)
	dy := y - float64(y0)

	rowOffset1 := y0 * c.width
	rowOffset2 := y1 * c.width
	top := input[rowOffset1+x0]*(1-dx) + input[rowOffset1+x1]*dx
	bottom := input[rowOffset2+x0]*(1-dx) + input[rowOffset2+x1]*dx

	return top*(1-dy) + bottom*dy
}

func (c *Canny) normalize(input []float64, targetMaximum float64) []cannyPixel {
	var maximum float64

	for _, value := range input {
		if value > maximum {
			maximum = value
		}
	}

	if maximum <= 0 {
		return nil
	}

	scale := targetMaximum / maximum
	result := make([]cannyPixel, len(input))

	for i := range input {
		switch {
		case input[i]*scale >= c.highThreshold:
			result[i] = cannyStrong
		case input[i]*scale >= c.lowThreshold:
			result[i] = cannyWeak
		default:
			result[i] = cannyNone
		}
	}

	return result
}

// removeWeakEdges removes the weak edges from the provided pixel buffer
// and returns the final edges represented as boolean (TRUE) values.
func (c *Canny) removeWeakEdges(pixels []cannyPixel) []bool {
	result := make([]bool, c.width*c.height)
	queue := make([]point, 0)

	for y := 0; y < c.height; y++ {
		rowOffset := y * c.width

		for x := 0; x < c.width; x++ {
			index := rowOffset + x

			if pixels[index] != cannyStrong {
				continue
			}

			result[index] = true
			queue = append(queue, point{
				x: x,
				y: y,
			})
		}
	}

	directions := [8]point{
		{x: -1, y: -1},
		{x: 0, y: -1},
		{x: 1, y: -1},
		{x: -1, y: 0},
		{x: 1, y: 0},
		{x: -1, y: 1},
		{x: 0, y: 1},
		{x: 1, y: 1},
	}

	head := 0

	for head < len(queue) {
		current := queue[head]
		head++

		for _, direction := range directions {
			nx := current.x + direction.x
			ny := current.y + direction.y

			if nx < 0 || nx >= c.width || ny < 0 || ny >= c.height {
				continue
			}

			index := ny*c.width + nx

			if result[index] {
				continue
			}

			if pixels[index] != cannyWeak {
				continue
			}

			result[index] = true
			queue = append(queue, point{
				x: nx,
				y: ny,
			})
		}
	}

	return result
}

// ContainsEdges returns true if an edge is encountered, otherwise returns false.
func (c *Canny) ContainsEdges(edges []bool) bool {
	for _, edge := range edges {
		if edge {
			return true
		}
	}

	return false
}

// WithDefaultOptions instantiates the edge detector with the default values.
func WithDefaultOptions(opt Canny) CannyOption {
	return func(c *Canny) error {
		c.gaussianSigma = opt.gaussianSigma
		c.lowThreshold = opt.lowThreshold
		c.highThreshold = opt.highThreshold

		return nil
	}
}

// WithGaussianSigma applies a gaussian sigma different from the default value to the edge detector.
func WithGaussianSigma(sigma float64) CannyOption {
	return func(c *Canny) error {
		c.gaussianSigma = sigma
		return nil
	}
}

// WithLowThreshold applies the lower threshold value to the edge detector.
func WithLowThreshold(lowThreshold float64) CannyOption {
	return func(c *Canny) error {
		c.lowThreshold = lowThreshold
		return nil
	}
}

// WithHighThreshold applies the higher threshold value to the edge detector.
func WithHighThreshold(highThreshold float64) CannyOption {
	return func(c *Canny) error {
		c.highThreshold = highThreshold
		return nil
	}
}
