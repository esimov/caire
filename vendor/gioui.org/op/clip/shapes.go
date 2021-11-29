// SPDX-License-Identifier: Unlicense OR MIT

package clip

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/internal/ops"
	"gioui.org/op"
)

// Rect represents the clip area of a pixel-aligned rectangle.
type Rect image.Rectangle

// Op returns the op for the rectangle.
func (r Rect) Op() Op {
	return Op{
		outline: true,
		path:    r.Path(),
	}
}

// Push the clip operation on the clip stack.
func (r Rect) Push(ops *op.Ops) Stack {
	return r.Op().Push(ops)
}

// Path returns the PathSpec for the rectangle.
func (r Rect) Path() PathSpec {
	return PathSpec{
		shape:  ops.Rect,
		bounds: image.Rectangle(r),
	}
}

// UniformRRect returns an RRect with all corner radii set to the
// provided radius.
func UniformRRect(rect f32.Rectangle, radius float32) RRect {
	return RRect{
		Rect: rect,
		SE:   radius,
		SW:   radius,
		NE:   radius,
		NW:   radius,
	}
}

// RRect represents the clip area of a rectangle with rounded
// corners.
//
// Specify a square with corner radii equal to half the square size to
// construct a circular clip area.
type RRect struct {
	Rect f32.Rectangle
	// The corner radii.
	SE, SW, NW, NE float32
}

// Op returns the op for the rounded rectangle.
func (rr RRect) Op(ops *op.Ops) Op {
	if rr.SE == 0 && rr.SW == 0 && rr.NW == 0 && rr.NE == 0 {
		r := image.Rectangle{
			Min: image.Point{X: int(rr.Rect.Min.X), Y: int(rr.Rect.Min.Y)},
			Max: image.Point{X: int(rr.Rect.Max.X), Y: int(rr.Rect.Max.Y)},
		}
		// Only use Rect if rr is pixel-aligned, as Rect is guaranteed to be.
		if fPt(r.Min) == rr.Rect.Min && fPt(r.Max) == rr.Rect.Max {
			return Rect(r).Op()
		}
	}
	return Outline{Path: rr.Path(ops)}.Op()
}

// Push the rectangle clip on the clip stack.
func (rr RRect) Push(ops *op.Ops) Stack {
	return rr.Op(ops).Push(ops)
}

// Path returns the PathSpec for the rounded rectangle.
func (rr RRect) Path(ops *op.Ops) PathSpec {
	var p Path
	p.Begin(ops)

	// https://pomax.github.io/bezierinfo/#circles_cubic.
	const q = 4 * (math.Sqrt2 - 1) / 3
	const iq = 1 - q

	se, sw, nw, ne := rr.SE, rr.SW, rr.NW, rr.NE
	w, n, e, s := rr.Rect.Min.X, rr.Rect.Min.Y, rr.Rect.Max.X, rr.Rect.Max.Y

	p.MoveTo(f32.Point{X: w + nw, Y: n})
	p.LineTo(f32.Point{X: e - ne, Y: n}) // N
	p.CubeTo(                            // NE
		f32.Point{X: e - ne*iq, Y: n},
		f32.Point{X: e, Y: n + ne*iq},
		f32.Point{X: e, Y: n + ne})
	p.LineTo(f32.Point{X: e, Y: s - se}) // E
	p.CubeTo(                            // SE
		f32.Point{X: e, Y: s - se*iq},
		f32.Point{X: e - se*iq, Y: s},
		f32.Point{X: e - se, Y: s})
	p.LineTo(f32.Point{X: w + sw, Y: s}) // S
	p.CubeTo(                            // SW
		f32.Point{X: w + sw*iq, Y: s},
		f32.Point{X: w, Y: s - sw*iq},
		f32.Point{X: w, Y: s - sw})
	p.LineTo(f32.Point{X: w, Y: n + nw}) // W
	p.CubeTo(                            // NW
		f32.Point{X: w, Y: n + nw*iq},
		f32.Point{X: w + nw*iq, Y: n},
		f32.Point{X: w + nw, Y: n})

	return p.End()
}

// Circle represents the clip area of a circle.
type Circle struct {
	Center f32.Point
	Radius float32
}

// Op returns the op for the filled circle.
func (c Circle) Op(ops *op.Ops) Op {
	return Outline{Path: c.Path(ops)}.Op()
}

// Push the circle clip on the clip stack.
func (c Circle) Push(ops *op.Ops) Stack {
	return c.Op(ops).Push(ops)
}

// Path returns the PathSpec for the circle.
//
// Deprecated: use Ellipse instead.
func (c Circle) Path(ops *op.Ops) PathSpec {
	b := f32.Rectangle{
		Min: f32.Pt(c.Center.X-c.Radius, c.Center.Y-c.Radius),
		Max: f32.Pt(c.Center.X+c.Radius, c.Center.Y+c.Radius),
	}
	return Ellipse(b).path(ops)
}

// Ellipse represents the largest axis-aligned ellipse that
// is contained in its bounds.
type Ellipse f32.Rectangle

// Op returns the op for the filled ellipse.
func (e Ellipse) Op(ops *op.Ops) Op {
	return Outline{Path: e.path(ops)}.Op()
}

// Push the filled ellipse clip op on the clip stack.
func (e Ellipse) Push(ops *op.Ops) Stack {
	return e.Op(ops).Push(ops)
}

// path constructs a path for the ellipse.
func (e Ellipse) path(o *op.Ops) PathSpec {
	var p Path
	p.Begin(o)

	bounds := f32.Rectangle(e)
	center := bounds.Max.Add(bounds.Min).Mul(.5)
	diam := bounds.Dx()
	r := diam * .5
	// We'll model the ellipse as a circle scaled in the Y
	// direction.
	scale := bounds.Dy() / diam

	// https://pomax.github.io/bezierinfo/#circles_cubic.
	const q = 4 * (math.Sqrt2 - 1) / 3

	curve := r * q
	top := f32.Point{X: center.X, Y: center.Y - r*scale}

	p.MoveTo(top)
	p.CubeTo(
		f32.Point{X: center.X + curve, Y: center.Y - r*scale},
		f32.Point{X: center.X + r, Y: center.Y - curve*scale},
		f32.Point{X: center.X + r, Y: center.Y},
	)
	p.CubeTo(
		f32.Point{X: center.X + r, Y: center.Y + curve*scale},
		f32.Point{X: center.X + curve, Y: center.Y + r*scale},
		f32.Point{X: center.X, Y: center.Y + r*scale},
	)
	p.CubeTo(
		f32.Point{X: center.X - curve, Y: center.Y + r*scale},
		f32.Point{X: center.X - r, Y: center.Y + curve*scale},
		f32.Point{X: center.X - r, Y: center.Y},
	)
	p.CubeTo(
		f32.Point{X: center.X - r, Y: center.Y - curve*scale},
		f32.Point{X: center.X - curve, Y: center.Y - r*scale},
		top,
	)
	ellipse := p.End()
	ellipse.shape = ops.Ellipse
	return ellipse
}

func fPt(p image.Point) f32.Point {
	return f32.Point{
		X: float32(p.X), Y: float32(p.Y),
	}
}
