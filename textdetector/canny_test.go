package textdetector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanny_ApplyGradientsUniformImage(t *testing.T) {
	c := &CannyDetector{Width: 5, Height: 5}
	input := []uint8{
		100, 100, 100, 100, 100,
		100, 100, 100, 100, 100,
		100, 100, 100, 100, 100,
		100, 100, 100, 100, 100,
		100, 100, 100, 100, 100,
	}

	gradients, magnitudes := c.applyGradients(input)

	for i := range magnitudes {
		assert.Equal(t, float64(0), magnitudes[i])
		assert.Equal(t, gradient{}, gradients[i])
	}
}

func TestCanny_ApplyGradientsVerticalEdge(t *testing.T) {
	c := &CannyDetector{Width: 5, Height: 5}
	input := []uint8{
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
	}

	gradients, magnitudes := c.applyGradients(input)
	index := 2*c.Width + 2

	assert.Greater(t, magnitudes[index], float64(0))
	assert.InDelta(t, 1.0, gradients[index].x, 0.0001)
	assert.InDelta(t, 0.0, gradients[index].y, 0.0001)
}

func TestCanny_BilinearInterpolationExactPixel(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 3}
	input := []float64{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	}

	result := c.bilinearInterpolation(input, 1, 1)

	assert.Equal(t, float64(5), result)
}

func TestCanny_BilinearInterpolationBetweenPixels(t *testing.T) {
	c := &CannyDetector{Width: 2, Height: 2}
	input := []float64{
		0, 10,
		20, 30,
	}

	result := c.bilinearInterpolation(input, 0.5, 0.5)

	assert.InDelta(t, 15.0, result, 0.0001)
}

func TestCanny_BilinearInterpolationOutsideImage(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 3}
	input := make([]float64, 9)

	assert.Equal(t, float64(0), c.bilinearInterpolation(input, -1, 1))
	assert.Equal(t, float64(0), c.bilinearInterpolation(input, 1, -1))
	assert.Equal(t, float64(0), c.bilinearInterpolation(input, 3, 1))
	assert.Equal(t, float64(0), c.bilinearInterpolation(input, 1, 3))
}

func TestCanny_NonMaximumSuppressionKeepsMaximum(t *testing.T) {
	c := &CannyDetector{Width: 5, Height: 3}
	magnitudes := []float64{
		0, 0, 0, 0, 0,
		0, 5, 10, 5, 0,
		0, 0, 0, 0, 0,
	}
	gradients := make([]gradient, len(magnitudes))
	gradients[6] = gradient{x: 1}
	gradients[7] = gradient{x: 1}
	gradients[8] = gradient{x: 1}

	result := c.nonMaximumSuppression(magnitudes, gradients)

	assert.Equal(t, float64(0), result[6])
	assert.Equal(t, float64(10), result[7])
	assert.Equal(t, float64(0), result[8])
}

func TestCanny_NormalizeClassifiesPixels(t *testing.T) {
	c := &CannyDetector{LowThreshold: 50, HighThreshold: 150}
	input := []float64{0, 10, 50, 100}

	result := c.normalize(input, 255)

	assert.Equal(t, cannyNone, result[0])
	assert.Equal(t, cannyNone, result[1])
	assert.Equal(t, cannyWeak, result[2])
	assert.Equal(t, cannyStrong, result[3])
}

func TestCanny_RemoveWeakEdgesKeepsConnectedWeakEdges(t *testing.T) {
	c := &CannyDetector{Width: 5, Height: 3}
	pixels := []cannyPixel{
		cannyNone, cannyNone, cannyNone, cannyNone, cannyNone,
		cannyNone, cannyStrong, cannyWeak, cannyWeak, cannyNone,
		cannyNone, cannyNone, cannyNone, cannyNone, cannyNone,
	}

	result := c.removeWeakEdges(pixels)

	assert.True(t, result[6])
	assert.True(t, result[7])
	assert.True(t, result[8])
}

func TestCanny_RemoveWeakEdgesRejectsDisconnectedWeakEdges(t *testing.T) {
	c := &CannyDetector{Width: 5, Height: 3}
	pixels := []cannyPixel{
		cannyNone, cannyNone, cannyNone, cannyNone, cannyNone,
		cannyNone, cannyStrong, cannyNone, cannyWeak, cannyNone,
		cannyNone, cannyNone, cannyNone, cannyNone, cannyNone,
	}

	result := c.removeWeakEdges(pixels)

	assert.True(t, result[6])
	assert.False(t, result[8])
}

func TestCanny_RemoveWeakEdgesUsesDiagonalConnectivity(t *testing.T) {
	c := &CannyDetector{Width: 3, Height: 3}
	pixels := []cannyPixel{
		cannyStrong, cannyNone, cannyNone,
		cannyNone, cannyWeak, cannyNone,
		cannyNone, cannyNone, cannyWeak,
	}

	result := c.removeWeakEdges(pixels)

	assert.True(t, result[0])
	assert.True(t, result[4])
	assert.True(t, result[8])
}

func TestCanny_ContainsEdges(t *testing.T) {
	c := &CannyDetector{}

	assert.False(t, c.ContainsEdges([]bool{false, false, false}))
	assert.True(t, c.ContainsEdges([]bool{false, true, false}))
	assert.False(t, c.ContainsEdges(nil))
}

func TestCanny_DetectUniformImage(t *testing.T) {
	c := &CannyDetector{Width: 10, Height: 10, GaussianSigma: 1.4, LowThreshold: 50, HighThreshold: 100}
	input := make([]uint8, c.Width*c.Height)
	for i := range input {
		input[i] = 128
	}

	edges, gradients, err := c.Detect(input)

	assert.NoError(t, err)
	assert.Len(t, edges, c.Width*c.Height)
	assert.Len(t, gradients, c.Width*c.Height)
	assert.False(t, c.ContainsEdges(edges))
}

func TestCanny_GradientIsNormalized(t *testing.T) {
	c := &CannyDetector{Width: 5, Height: 5}
	input := []uint8{
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
		0, 0, 0, 255, 255,
	}

	gradients, magnitudes := c.applyGradients(input)

	for i, magnitude := range magnitudes {
		if magnitude == 0 {
			continue
		}

		length := math.Hypot(gradients[i].x, gradients[i].y)
		assert.InDelta(t, 1.0, length, 0.0001)
	}
}
