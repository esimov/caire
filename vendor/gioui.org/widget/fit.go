// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
)

// Fit scales a widget to fit and clip to the constraints.
type Fit uint8

const (
	// Unscaled does not alter the scale of a widget.
	Unscaled Fit = iota
	// Contain scales widget as large as possible without cropping
	// and it preserves aspect-ratio.
	Contain
	// Cover scales the widget to cover the constraint area and
	// preserves aspect-ratio.
	Cover
	// ScaleDown scales the widget smaller without cropping,
	// when it exceeds the constraint area.
	// It preserves aspect-ratio.
	ScaleDown
	// Fill stretches the widget to the constraints and does not
	// preserve aspect-ratio.
	Fill
)

// scale computes the new dimensions and transformation required to fit dims to cs, given the position.
func (fit Fit) scale(cs layout.Constraints, pos layout.Direction, dims layout.Dimensions) (layout.Dimensions, f32.Affine2D) {
	widgetSize := dims.Size

	if fit == Unscaled || dims.Size.X == 0 || dims.Size.Y == 0 {
		dims.Size = cs.Constrain(dims.Size)

		offset := pos.Position(widgetSize, dims.Size)
		dims.Baseline += offset.Y
		return dims, f32.Affine2D{}.Offset(layout.FPt(offset))
	}

	scale := f32.Point{
		X: float32(cs.Max.X) / float32(dims.Size.X),
		Y: float32(cs.Max.Y) / float32(dims.Size.Y),
	}

	switch fit {
	case Contain:
		if scale.Y < scale.X {
			scale.X = scale.Y
		} else {
			scale.Y = scale.X
		}
	case Cover:
		if scale.Y > scale.X {
			scale.X = scale.Y
		} else {
			scale.Y = scale.X
		}
	case ScaleDown:
		if scale.Y < scale.X {
			scale.X = scale.Y
		} else {
			scale.Y = scale.X
		}

		// The widget would need to be scaled up, no change needed.
		if scale.X >= 1 {
			dims.Size = cs.Constrain(dims.Size)

			offset := pos.Position(widgetSize, dims.Size)
			dims.Baseline += offset.Y
			return dims, f32.Affine2D{}.Offset(layout.FPt(offset))
		}
	case Fill:
	}

	var scaledSize image.Point
	scaledSize.X = int(float32(widgetSize.X) * scale.X)
	scaledSize.Y = int(float32(widgetSize.Y) * scale.Y)
	dims.Size = cs.Constrain(scaledSize)
	dims.Baseline = int(float32(dims.Baseline) * scale.Y)

	offset := pos.Position(scaledSize, dims.Size)
	trans := f32.Affine2D{}.
		Scale(f32.Point{}, scale).
		Offset(layout.FPt(offset))

	dims.Baseline += offset.Y

	return dims, trans
}
