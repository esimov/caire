// SPDX-License-Identifier: Unlicense OR MIT

package material

import (
	"image"
	"image/color"

	"gioui.org/internal/f32color"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

type SwitchStyle struct {
	Description string
	Color       struct {
		Enabled  color.NRGBA
		Disabled color.NRGBA
		Track    color.NRGBA
	}
	Switch *widget.Bool
}

// Switch is for selecting a boolean value.
func Switch(th *Theme, swtch *widget.Bool, description string) SwitchStyle {
	sw := SwitchStyle{
		Switch:      swtch,
		Description: description,
	}
	sw.Color.Enabled = th.Palette.ContrastBg
	sw.Color.Disabled = th.Palette.Bg
	sw.Color.Track = f32color.MulAlpha(th.Palette.Fg, 0x88)
	return sw
}

// Layout updates the switch and displays it.
func (s SwitchStyle) Layout(gtx layout.Context) layout.Dimensions {
	trackWidth := gtx.Dp(36)
	trackHeight := gtx.Dp(16)
	thumbSize := gtx.Dp(20)
	trackOff := (thumbSize - trackHeight) / 2

	// Draw track.
	trackCorner := trackHeight / 2
	trackRect := image.Rectangle{Max: image.Point{
		X: trackWidth,
		Y: trackHeight,
	}}
	col := s.Color.Disabled
	if s.Switch.Value {
		col = s.Color.Enabled
	}
	if gtx.Queue == nil {
		col = f32color.Disabled(col)
	}
	trackColor := s.Color.Track
	t := op.Offset(image.Point{Y: trackOff}).Push(gtx.Ops)
	cl := clip.UniformRRect(trackRect, trackCorner).Push(gtx.Ops)
	paint.ColorOp{Color: trackColor}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	cl.Pop()
	t.Pop()

	// Draw thumb ink.
	inkSize := gtx.Dp(44)
	rr := inkSize / 2
	inkOff := image.Point{
		X: trackWidth/2 - rr,
		Y: -rr + trackHeight/2 + trackOff,
	}
	t = op.Offset(inkOff).Push(gtx.Ops)
	gtx.Constraints.Min = image.Pt(inkSize, inkSize)
	cl = clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, rr).Push(gtx.Ops)
	for _, p := range s.Switch.History() {
		drawInk(gtx, p)
	}
	cl.Pop()
	t.Pop()

	// Compute thumb offset.
	if s.Switch.Value {
		xoff := trackWidth - thumbSize
		defer op.Offset(image.Point{X: xoff}).Push(gtx.Ops).Pop()
	}

	thumbRadius := thumbSize / 2

	circle := func(x, y, r int) clip.Op {
		b := image.Rectangle{
			Min: image.Pt(x-r, y-r),
			Max: image.Pt(x+r, y+r),
		}
		return clip.Ellipse(b).Op(gtx.Ops)
	}
	// Draw hover.
	if s.Switch.Hovered() || s.Switch.Focused() {
		r := thumbRadius * 10 / 17
		background := f32color.MulAlpha(s.Color.Enabled, 70)
		paint.FillShape(gtx.Ops, background, circle(thumbRadius, thumbRadius, r))
	}

	// Draw thumb shadow, a translucent disc slightly larger than the
	// thumb itself.
	// Center shadow horizontally and slightly adjust its Y.
	paint.FillShape(gtx.Ops, argb(0x55000000), circle(thumbRadius, thumbRadius+gtx.Dp(.25), thumbRadius+1))

	// Draw thumb.
	paint.FillShape(gtx.Ops, col, circle(thumbRadius, thumbRadius, thumbRadius))

	// Set up click area.
	clickSize := gtx.Dp(40)
	clickOff := image.Point{
		X: (thumbSize - clickSize) / 2,
		Y: (trackHeight-clickSize)/2 + trackOff,
	}
	defer op.Offset(clickOff).Push(gtx.Ops).Pop()
	sz := image.Pt(clickSize, clickSize)
	defer clip.Ellipse(image.Rectangle{Max: sz}).Push(gtx.Ops).Pop()
	s.Switch.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if d := s.Description; d != "" {
			semantic.DescriptionOp(d).Add(gtx.Ops)
		}
		semantic.Switch.Add(gtx.Ops)
		return layout.Dimensions{Size: sz}
	})

	dims := image.Point{X: trackWidth, Y: thumbSize}
	return layout.Dimensions{Size: dims}
}
