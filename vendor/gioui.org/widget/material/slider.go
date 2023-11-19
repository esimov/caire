// SPDX-License-Identifier: Unlicense OR MIT

package material

import (
	"image"
	"image/color"

	"gioui.org/internal/f32color"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Slider is for selecting a value in a range.
func Slider(th *Theme, float *widget.Float, min, max float32) SliderStyle {
	return SliderStyle{
		Min:        min,
		Max:        max,
		Color:      th.Palette.ContrastBg,
		Float:      float,
		FingerSize: th.FingerSize,
	}
}

type SliderStyle struct {
	Axis     layout.Axis
	Min, Max float32
	Invert   bool
	Color    color.NRGBA
	Float    *widget.Float

	FingerSize unit.Dp
}

func (s SliderStyle) Layout(gtx layout.Context) layout.Dimensions {
	thumbRadius := gtx.Dp(6)
	trackWidth := gtx.Dp(2)

	axis := s.Axis
	// Keep a minimum length so that the track is always visible.
	minLength := thumbRadius + 3*thumbRadius + thumbRadius
	// Try to expand to finger size, but only if the constraints
	// allow for it.
	touchSizePx := min(gtx.Dp(s.FingerSize), axis.Convert(gtx.Constraints.Max).Y)
	sizeMain := max(axis.Convert(gtx.Constraints.Min).X, minLength)
	sizeCross := max(2*thumbRadius, touchSizePx)
	size := axis.Convert(image.Pt(sizeMain, sizeCross))

	o := axis.Convert(image.Pt(thumbRadius, 0))
	trans := op.Offset(o).Push(gtx.Ops)
	gtx.Constraints.Min = axis.Convert(image.Pt(sizeMain-2*thumbRadius, sizeCross))
	s.Float.Layout(gtx, axis, s.Min, s.Max, s.Invert, thumbRadius)
	gtx.Constraints.Min = gtx.Constraints.Min.Add(axis.Convert(image.Pt(0, sizeCross)))
	thumbPos := thumbRadius + int(s.Float.Pos())
	trans.Pop()

	color := s.Color
	if gtx.Queue == nil {
		color = f32color.Disabled(color)
	}

	rect := func(minx, miny, maxx, maxy int) image.Rectangle {
		r := image.Rect(minx, miny, maxx, maxy)
		if s.Invert != (axis == layout.Vertical) {
			r.Max.X, r.Min.X = sizeMain-r.Min.X, sizeMain-r.Max.X
		}
		r.Min = axis.Convert(r.Min)
		r.Max = axis.Convert(r.Max)
		return r
	}

	// Draw track before thumb.
	track := rect(
		thumbRadius, sizeCross/2-trackWidth/2,
		thumbPos, sizeCross/2+trackWidth/2,
	)
	paint.FillShape(gtx.Ops, color, clip.Rect(track).Op())

	// Draw track after thumb.
	track = rect(
		thumbPos, axis.Convert(track.Min).Y,
		sizeMain-thumbRadius, axis.Convert(track.Max).Y,
	)
	paint.FillShape(gtx.Ops, f32color.MulAlpha(color, 96), clip.Rect(track).Op())

	// Draw thumb.
	pt := image.Pt(thumbPos, sizeCross/2)
	thumb := rect(
		pt.X-thumbRadius, pt.Y-thumbRadius,
		pt.X+thumbRadius, pt.Y+thumbRadius,
	)
	paint.FillShape(gtx.Ops, color, clip.Ellipse(thumb).Op(gtx.Ops))

	return layout.Dimensions{Size: size}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
