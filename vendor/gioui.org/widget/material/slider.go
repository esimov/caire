// SPDX-License-Identifier: Unlicense OR MIT

package material

import (
	"image"
	"image/color"

	"gioui.org/f32"
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
	Min, Max float32
	Color    color.NRGBA
	Float    *widget.Float

	FingerSize unit.Value
}

func (s SliderStyle) Layout(gtx layout.Context) layout.Dimensions {
	thumbRadius := gtx.Px(unit.Dp(6))
	trackWidth := gtx.Px(unit.Dp(2))

	axis := s.Float.Axis
	// Keep a minimum length so that the track is always visible.
	minLength := thumbRadius + 3*thumbRadius + thumbRadius
	// Try to expand to finger size, but only if the constraints
	// allow for it.
	touchSizePx := min(gtx.Px(s.FingerSize), axis.Convert(gtx.Constraints.Max).Y)
	sizeMain := max(axis.Convert(gtx.Constraints.Min).X, minLength)
	sizeCross := max(2*thumbRadius, touchSizePx)
	size := axis.Convert(image.Pt(sizeMain, sizeCross))

	o := axis.Convert(image.Pt(thumbRadius, 0))
	trans := op.Offset(layout.FPt(o)).Push(gtx.Ops)
	gtx.Constraints.Min = axis.Convert(image.Pt(sizeMain-2*thumbRadius, sizeCross))
	s.Float.Layout(gtx, thumbRadius, s.Min, s.Max)
	gtx.Constraints.Min = gtx.Constraints.Min.Add(axis.Convert(image.Pt(0, sizeCross)))
	thumbPos := thumbRadius + int(s.Float.Pos())
	trans.Pop()

	color := s.Color
	if gtx.Queue == nil {
		color = f32color.Disabled(color)
	}

	// Draw track before thumb.
	track := image.Rectangle{
		Min: axis.Convert(image.Pt(thumbRadius, sizeCross/2-trackWidth/2)),
		Max: axis.Convert(image.Pt(thumbPos, sizeCross/2+trackWidth/2)),
	}
	paint.FillShape(gtx.Ops, color, clip.Rect(track).Op())

	// Draw track after thumb.
	track = image.Rectangle{
		Min: axis.Convert(image.Pt(thumbPos, axis.Convert(track.Min).Y)),
		Max: axis.Convert(image.Pt(sizeMain-thumbRadius, axis.Convert(track.Max).Y)),
	}
	paint.FillShape(gtx.Ops, f32color.MulAlpha(color, 96), clip.Rect(track).Op())

	// Draw thumb.
	pt := axis.Convert(image.Pt(thumbPos, sizeCross/2))
	paint.FillShape(gtx.Ops, color,
		clip.Circle{
			Center: f32.Point{X: float32(pt.X), Y: float32(pt.Y)},
			Radius: float32(thumbRadius),
		}.Op(gtx.Ops))

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
