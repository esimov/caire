package textdetector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSWT_TransformDetectsHorizontalStroke(t *testing.T) {
	s := &SWT{Component: &Component{MaxRayLength: 10}}

	width, height := 15, 5
	edges := make([]bool, width*height)
	gradients := make([]gradient, width*height)

	edges[2*width+4] = true
	edges[2*width+8] = true
	gradients[2*width+4] = gradient{x: 1, y: 0}
	gradients[2*width+8] = gradient{x: -1, y: 0}

	strokes := s.Transform(width, height, edges, gradients, 1)

	for x := 4; x <= 8; x++ {
		assert.InDelta(t, 4.0, strokes[2*width+x], 0.0001)
	}
	assert.Equal(t, unsetStroke, strokes[2*width+3])
	assert.Equal(t, unsetStroke, strokes[2*width+9])
}

func TestSWT_TransformIgnoresNonOpposingGradients(t *testing.T) {
	s := &SWT{Component: &Component{MaxRayLength: 10}}

	width, height := 15, 5
	edges := make([]bool, width*height)
	gradients := make([]gradient, width*height)

	edges[2*width+4] = true
	edges[2*width+8] = true
	gradients[2*width+4] = gradient{x: 1, y: 0}
	gradients[2*width+8] = gradient{x: 1, y: 0}

	strokes := s.Transform(width, height, edges, gradients, 1)

	for _, stroke := range strokes {
		assert.Equal(t, unsetStroke, stroke)
	}
}

func TestSWT_TransformSkipsZeroGradient(t *testing.T) {
	s := &SWT{Component: &Component{MaxRayLength: 10}}

	width, height := 5, 5
	edges := make([]bool, width*height)
	gradients := make([]gradient, width*height)

	edges[2*width+2] = true

	strokes := s.Transform(width, height, edges, gradients, 1)

	for _, stroke := range strokes {
		assert.Equal(t, unsetStroke, stroke)
	}
}

func TestSWT_AreOpposingGradients(t *testing.T) {
	s := &SWT{}

	assert.True(t, s.areOpposingGradients(gradient{x: 1}, gradient{x: -1}, 1))
	assert.True(t, s.areOpposingGradients(gradient{x: 1}, gradient{x: -1}, -1))
	assert.False(t, s.areOpposingGradients(gradient{x: 1}, gradient{x: 1}, 1))
	assert.False(t, s.areOpposingGradients(gradient{x: 1}, gradient{y: 1}, 1))
}

func TestSWT_AreOpposingGradientsAcceptsAngleBoundary(t *testing.T) {
	s := &SWT{}
	angle := 150.0 * math.Pi / 180

	assert.True(t, s.areOpposingGradients(gradient{x: 1}, gradient{x: math.Cos(angle), y: math.Sin(angle)}, 1))
}
