package textdetector

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEdgeDetector struct {
	edges     []bool
	gradients []gradient
}

func (d *testEdgeDetector) Detect(_ []uint8) ([]bool, []gradient, error) {
	return d.edges, d.gradients, nil
}

func (d *testEdgeDetector) ContainsEdges(edges []bool) bool {
	for _, edge := range edges {
		if edge {
			return true
		}
	}
	return false
}

func TestSWT_DetectReturnsNilWhenNoEdgesAreDetected(t *testing.T) {
	width, height := 15, 10
	detector := &testEdgeDetector{
		edges:     make([]bool, width*height),
		gradients: make([]gradient, width*height),
	}
	swt := NewSWT(detector, DefaultOptions())
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	regions := swt.Detect(img)

	assert.Nil(t, regions)
}

func TestSWT_DetectFindsTextRegion(t *testing.T) {
	const width, height = 140, 60

	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	black := color.NRGBA{A: 255}

	// Simulate a text like "UUU".
	// The detector should detect as one grouped text component.
	for _, offset := range []int{10, 50, 90} {
		// Left vertical stroke.
		for y := 10; y < 42; y++ {
			for x := offset; x < offset+4; x++ {
				img.SetNRGBA(x, y, black)
			}
		}

		// Right vertical stroke.
		for y := 10; y < 42; y++ {
			for x := offset + 16; x < offset+20; x++ {
				img.SetNRGBA(x, y, black)
			}
		}

		// Bottom stroke.
		for y := 38; y < 45; y++ {
			for x := offset; x < offset+20; x++ {
				img.SetNRGBA(x, y, black)
			}
		}
	}

	edgeDetector, err := NewCannyDetector(width, height,
		WithDefaultOptions(CannyDetector{
			GaussianSigma: 1.0,
			LowThreshold:  30,
			HighThreshold: 80,
		}))
	assert.NoError(t, err)

	swt := NewSWT(edgeDetector, DefaultOptions())
	swt.Direction = Horizontal
	swt.MaxRayLength = 30

	regions := swt.Detect(img)

	assert.NotEmpty(t, regions)
}
