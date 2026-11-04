package textdetector

import (
	"fmt"
	"math"

	"github.com/esimov/caire/common"
)

// GaussianBlur applies the gaussian formula to the pixel buffer to reduce the image noisiness.
// Used as a very efficient and common solution for edge detection.
func (c *CannyDetector) GaussianBlur(input []uint8, sigma float64) ([]uint8, error) {
	if c.Width <= 0 || c.Height <= 0 || len(input) < c.Width*c.Height {
		return nil, fmt.Errorf("the pixel buffer size does not match with the image size")
	}

	kernel := c.gaussianKernel(sigma)
	radius := len(kernel) / 2

	horizontal := make([]float64, c.Width*c.Height)
	result := make([]uint8, c.Width*c.Height)

	// Apply the gaussian blur formula in 2d dimension.
	for y := 0; y < c.Height; y++ {
		rowOffset := y * c.Width

		for x := 0; x < c.Width; x++ {
			var sum float64

			for k := -radius; k <= radius; k++ {
				sampleX := common.Clamp(x+k, 0, c.Width-1)

				sum += float64(input[rowOffset+sampleX]) * kernel[k+radius]
			}

			horizontal[rowOffset+x] = sum
		}
	}

	for y := 0; y < c.Height; y++ {
		rowOffset := y * c.Width

		for x := 0; x < c.Width; x++ {
			var sum float64

			for k := -radius; k <= radius; k++ {
				sampleY := common.Clamp(y+k, 0, c.Height-1)

				sum += horizontal[sampleY*c.Width+x] * kernel[k+radius]
			}

			result[rowOffset+x] = uint8(math.Round(math.Max(0, math.Min(255, sum))))
		}
	}

	return result, nil
}

func (c *CannyDetector) gaussianKernel(sigma float64) []float64 {
	if sigma <= 0 {
		return []float64{1}
	}

	radius := int(math.Ceil(3 * sigma))
	size := radius*2 + 1

	kernel := make([]float64, size)

	var sum float64

	for i := -radius; i <= radius; i++ {
		value := math.Exp(
			-float64(i*i) / (2 * sigma * sigma),
		)

		kernel[i+radius] = value
		sum += value
	}

	for i := range kernel {
		kernel[i] /= sum
	}

	return kernel
}
