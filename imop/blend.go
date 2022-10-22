// Package imop implements the Porter-Duff composition operations
// used for mixing a graphic element with its backdrop.
// Porter and Duff presented in their paper 12 different composition operation,
// but the image/draw core package implements only the source-over-destination and source.
// This package is aimed to overcome the missing composite operations.

// It is mainly used to debug the seam carving operation correctness
// with face detection and image mask enabled.
// When the GUI mode and the debugging option is activated it will show
// the image mask and the detected faces rectangles in a distinct color.
package imop

import (
	"github.com/esimov/caire/utils"
)

const (
	Darken   = "darken"
	Lighten  = "lighten"
	Multiply = "multiply"
	Screen   = "screen"
	Overlay  = "overlay"
)

// Blend holds the currently active blend mode.
type Blend struct {
	OpType string
}

// NewBlend initializes a new Blend.
func NewBlend() *Blend {
	return &Blend{}
}

// Set activate one of the supported blend mode.
func (o *Blend) Set(opType string) {
	bModes := []string{Darken, Lighten, Multiply, Screen, Overlay}

	if utils.Contains(bModes, opType) {
		o.OpType = opType
	}
}

// Get returns the currently active blend mode.
func (o *Blend) Get() string {
	if len(o.OpType) > 0 {
		return o.OpType
	}
	return ""
}
