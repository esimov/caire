// SPDX-License-Identifier: Unlicense OR MIT

package layout

import (
	"image"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/unit"
)

// Constraints represent the minimum and maximum size of a widget.
//
// A widget does not have to treat its constraints as "hard". For
// example, if it's passed a constraint with a minimum size that's
// smaller than its actual minimum size, it should return its minimum
// size dimensions instead. Parent widgets should deal appropriately
// with child widgets that return dimensions that do not fit their
// constraints (for example, by clipping).
type Constraints struct {
	Min, Max image.Point
}

// Dimensions are the resolved size and baseline for a widget.
//
// Baseline is the distance from the bottom of a widget to the baseline of
// any text it contains (or 0). The purpose is to be able to align text
// that span multiple widgets.
type Dimensions struct {
	Size     image.Point
	Baseline int
}

// Axis is the Horizontal or Vertical direction.
type Axis uint8

// Alignment is the mutual alignment of a list of widgets.
type Alignment uint8

// Direction is the alignment of widgets relative to a containing
// space.
type Direction uint8

// Widget is a function scope for drawing, processing events and
// computing dimensions for a user interface element.
type Widget func(gtx Context) Dimensions

const (
	Start Alignment = iota
	End
	Middle
	Baseline
)

const (
	NW Direction = iota
	N
	NE
	E
	SE
	S
	SW
	W
	Center
)

const (
	Horizontal Axis = iota
	Vertical
)

// Exact returns the Constraints with the minimum and maximum size
// set to size.
func Exact(size image.Point) Constraints {
	return Constraints{
		Min: size, Max: size,
	}
}

// FPt converts an point to a f32.Point.
func FPt(p image.Point) f32.Point {
	return f32.Point{
		X: float32(p.X), Y: float32(p.Y),
	}
}

// FRect converts a rectangle to a f32.Rectangle.
func FRect(r image.Rectangle) f32.Rectangle {
	return f32.Rectangle{
		Min: FPt(r.Min), Max: FPt(r.Max),
	}
}

// Constrain a size so each dimension is in the range [min;max].
func (c Constraints) Constrain(size image.Point) image.Point {
	if min := c.Min.X; size.X < min {
		size.X = min
	}
	if min := c.Min.Y; size.Y < min {
		size.Y = min
	}
	if max := c.Max.X; size.X > max {
		size.X = max
	}
	if max := c.Max.Y; size.Y > max {
		size.Y = max
	}
	return size
}

// Inset adds space around a widget by decreasing its maximum
// constraints. The minimum constraints will be adjusted to ensure
// they do not exceed the maximum.
type Inset struct {
	Top, Right, Bottom, Left unit.Value
}

// Layout a widget.
func (in Inset) Layout(gtx Context, w Widget) Dimensions {
	top := gtx.Px(in.Top)
	right := gtx.Px(in.Right)
	bottom := gtx.Px(in.Bottom)
	left := gtx.Px(in.Left)
	mcs := gtx.Constraints
	mcs.Max.X -= left + right
	if mcs.Max.X < 0 {
		left = 0
		right = 0
		mcs.Max.X = 0
	}
	if mcs.Min.X > mcs.Max.X {
		mcs.Min.X = mcs.Max.X
	}
	mcs.Max.Y -= top + bottom
	if mcs.Max.Y < 0 {
		bottom = 0
		top = 0
		mcs.Max.Y = 0
	}
	if mcs.Min.Y > mcs.Max.Y {
		mcs.Min.Y = mcs.Max.Y
	}
	gtx.Constraints = mcs
	trans := op.Offset(FPt(image.Point{X: left, Y: top})).Push(gtx.Ops)
	dims := w(gtx)
	trans.Pop()
	return Dimensions{
		Size:     dims.Size.Add(image.Point{X: right + left, Y: top + bottom}),
		Baseline: dims.Baseline + bottom,
	}
}

// UniformInset returns an Inset with a single inset applied to all
// edges.
func UniformInset(v unit.Value) Inset {
	return Inset{Top: v, Right: v, Bottom: v, Left: v}
}

// Layout a widget according to the direction.
// The widget is called with the context constraints minimum cleared.
func (d Direction) Layout(gtx Context, w Widget) Dimensions {
	macro := op.Record(gtx.Ops)
	csn := gtx.Constraints.Min
	switch d {
	case N, S:
		gtx.Constraints.Min.Y = 0
	case E, W:
		gtx.Constraints.Min.X = 0
	default:
		gtx.Constraints.Min = image.Point{}
	}
	dims := w(gtx)
	call := macro.Stop()
	sz := dims.Size
	if sz.X < csn.X {
		sz.X = csn.X
	}
	if sz.Y < csn.Y {
		sz.Y = csn.Y
	}

	p := d.Position(dims.Size, sz)
	defer op.Offset(FPt(p)).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)

	return Dimensions{
		Size:     sz,
		Baseline: dims.Baseline + sz.Y - dims.Size.Y - p.Y,
	}
}

// Position calculates widget position according to the direction.
func (d Direction) Position(widget, bounds image.Point) image.Point {
	var p image.Point

	switch d {
	case N, S, Center:
		p.X = (bounds.X - widget.X) / 2
	case NE, SE, E:
		p.X = bounds.X - widget.X
	}

	switch d {
	case W, Center, E:
		p.Y = (bounds.Y - widget.Y) / 2
	case SW, S, SE:
		p.Y = bounds.Y - widget.Y
	}

	return p
}

// Spacer adds space between widgets.
type Spacer struct {
	Width, Height unit.Value
}

func (s Spacer) Layout(gtx Context) Dimensions {
	return Dimensions{
		Size: image.Point{
			X: gtx.Px(s.Width),
			Y: gtx.Px(s.Height),
		},
	}
}

func (a Alignment) String() string {
	switch a {
	case Start:
		return "Start"
	case End:
		return "End"
	case Middle:
		return "Middle"
	case Baseline:
		return "Baseline"
	default:
		panic("unreachable")
	}
}

// Convert a point in (x, y) coordinates to (main, cross) coordinates,
// or vice versa. Specifically, Convert((x, y)) returns (x, y) unchanged
// for the horizontal axis, or (y, x) for the vertical axis.
func (a Axis) Convert(pt image.Point) image.Point {
	if a == Horizontal {
		return pt
	}
	return image.Pt(pt.Y, pt.X)
}

// FConvert a point in (x, y) coordinates to (main, cross) coordinates,
// or vice versa. Specifically, FConvert((x, y)) returns (x, y) unchanged
// for the horizontal axis, or (y, x) for the vertical axis.
func (a Axis) FConvert(pt f32.Point) f32.Point {
	if a == Horizontal {
		return pt
	}
	return f32.Pt(pt.Y, pt.X)
}

// mainConstraint returns the min and max main constraints for axis a.
func (a Axis) mainConstraint(cs Constraints) (int, int) {
	if a == Horizontal {
		return cs.Min.X, cs.Max.X
	}
	return cs.Min.Y, cs.Max.Y
}

// crossConstraint returns the min and max cross constraints for axis a.
func (a Axis) crossConstraint(cs Constraints) (int, int) {
	if a == Horizontal {
		return cs.Min.Y, cs.Max.Y
	}
	return cs.Min.X, cs.Max.X
}

// constraints returns the constraints for axis a.
func (a Axis) constraints(mainMin, mainMax, crossMin, crossMax int) Constraints {
	if a == Horizontal {
		return Constraints{Min: image.Pt(mainMin, crossMin), Max: image.Pt(mainMax, crossMax)}
	}
	return Constraints{Min: image.Pt(crossMin, mainMin), Max: image.Pt(crossMax, mainMax)}
}

func (a Axis) String() string {
	switch a {
	case Horizontal:
		return "Horizontal"
	case Vertical:
		return "Vertical"
	default:
		panic("unreachable")
	}
}

func (d Direction) String() string {
	switch d {
	case NW:
		return "NW"
	case N:
		return "N"
	case NE:
		return "NE"
	case E:
		return "E"
	case SE:
		return "SE"
	case S:
		return "S"
	case SW:
		return "SW"
	case W:
		return "W"
	case Center:
		return "Center"
	default:
		panic("unreachable")
	}
}
