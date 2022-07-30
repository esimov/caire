// SPDX-License-Identifier: Unlicense OR MIT

package material

import (
	"image"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

type ProgressCircleStyle struct {
	Color    color.NRGBA
	Progress float32
}

func ProgressCircle(th *Theme, progress float32) ProgressCircleStyle {
	return ProgressCircleStyle{
		Color:    th.Palette.ContrastBg,
		Progress: progress,
	}
}

func (p ProgressCircleStyle) Layout(gtx layout.Context) layout.Dimensions {
	diam := gtx.Constraints.Min.X
	if minY := gtx.Constraints.Min.Y; minY > diam {
		diam = minY
	}
	if diam == 0 {
		diam = gtx.Dp(24)
	}
	sz := gtx.Constraints.Constrain(image.Pt(diam, diam))
	radius := sz.X / 2
	defer op.Offset(image.Pt(radius, radius)).Push(gtx.Ops).Pop()

	defer clipLoader(gtx.Ops, -math.Pi/2, -math.Pi/2+math.Pi*2*p.Progress, float32(radius)).Push(gtx.Ops).Pop()
	paint.ColorOp{
		Color: p.Color,
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{
		Size: sz,
	}
}
