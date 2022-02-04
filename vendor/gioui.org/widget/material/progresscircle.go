// SPDX-License-Identifier: Unlicense OR MIT

package material

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
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
		diam = gtx.Px(unit.Dp(24))
	}
	sz := gtx.Constraints.Constrain(image.Pt(diam, diam))
	radius := float32(sz.X) * .5
	defer op.Offset(f32.Pt(radius, radius)).Push(gtx.Ops).Pop()

	defer clipLoader(gtx.Ops, -math.Pi/2, -math.Pi/2+math.Pi*2*p.Progress, radius).Push(gtx.Ops).Pop()
	paint.ColorOp{
		Color: p.Color,
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{
		Size: sz,
	}
}
