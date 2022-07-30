// SPDX-License-Identifier: Unlicense OR MIT

package clip

import (
	"encoding/binary"
	"hash/maphash"
	"image"
	"math"

	"gioui.org/f32"
	f32internal "gioui.org/internal/f32"
	"gioui.org/internal/ops"
	"gioui.org/internal/scene"
	"gioui.org/internal/stroke"
	"gioui.org/op"
)

// Op represents a clip area. Op intersects the current clip area with
// itself.
type Op struct {
	path PathSpec

	outline bool
	width   float32
}

// Stack represents an Op pushed on the clip stack.
type Stack struct {
	ops     *ops.Ops
	id      ops.StackID
	macroID int
}

var pathSeed maphash.Seed

func init() {
	pathSeed = maphash.MakeSeed()
}

// Push saves the current clip state on the stack and updates the current
// state to the intersection of the current p.
func (p Op) Push(o *op.Ops) Stack {
	id, macroID := ops.PushOp(&o.Internal, ops.ClipStack)
	p.add(o)
	return Stack{ops: &o.Internal, id: id, macroID: macroID}
}

func (p Op) add(o *op.Ops) {
	path := p.path

	if !path.hasSegments && p.width > 0 {
		switch p.path.shape {
		case ops.Rect:
			b := f32internal.FRect(path.bounds)
			var rect Path
			rect.Begin(o)
			rect.MoveTo(b.Min)
			rect.LineTo(f32.Pt(b.Max.X, b.Min.Y))
			rect.LineTo(b.Max)
			rect.LineTo(f32.Pt(b.Min.X, b.Max.Y))
			rect.Close()
			path = rect.End()
		case ops.Path:
			// Nothing to do.
		default:
			panic("invalid empty path for shape")
		}
	}
	bo := binary.LittleEndian
	if path.hasSegments {
		data := ops.Write(&o.Internal, ops.TypePathLen)
		data[0] = byte(ops.TypePath)
		bo.PutUint64(data[1:], path.hash)
		path.spec.Add(o)
	}

	bounds := path.bounds
	if p.width > 0 {
		// Expand bounds to cover stroke.
		half := int(p.width*.5 + .5)
		bounds.Min.X -= half
		bounds.Min.Y -= half
		bounds.Max.X += half
		bounds.Max.Y += half
		data := ops.Write(&o.Internal, ops.TypeStrokeLen)
		data[0] = byte(ops.TypeStroke)
		bo := binary.LittleEndian
		bo.PutUint32(data[1:], math.Float32bits(p.width))
	}

	data := ops.Write(&o.Internal, ops.TypeClipLen)
	data[0] = byte(ops.TypeClip)
	bo.PutUint32(data[1:], uint32(bounds.Min.X))
	bo.PutUint32(data[5:], uint32(bounds.Min.Y))
	bo.PutUint32(data[9:], uint32(bounds.Max.X))
	bo.PutUint32(data[13:], uint32(bounds.Max.Y))
	if p.outline {
		data[17] = byte(1)
	}
	data[18] = byte(path.shape)
}

func (s Stack) Pop() {
	ops.PopOp(s.ops, ops.ClipStack, s.id, s.macroID)
	data := ops.Write(s.ops, ops.TypePopClipLen)
	data[0] = byte(ops.TypePopClip)
}

type PathSpec struct {
	spec op.CallOp
	// hasSegments tracks whether there are any segments in the path.
	hasSegments bool
	bounds      image.Rectangle
	shape       ops.Shape
	hash        uint64
}

// Path constructs a Op clip path described by lines and
// Bézier curves, where drawing outside the Path is discarded.
// The inside-ness of a pixel is determines by the non-zero winding rule,
// similar to the SVG rule of the same name.
//
// Path generates no garbage and can be used for dynamic paths; path
// data is stored directly in the Ops list supplied to Begin.
type Path struct {
	ops         *ops.Ops
	contour     int
	pen         f32.Point
	macro       op.MacroOp
	start       f32.Point
	hasSegments bool
	bounds      f32internal.Rectangle
	hash        maphash.Hash
}

// Pos returns the current pen position.
func (p *Path) Pos() f32.Point { return p.pen }

// Begin the path, storing the path data and final Op into ops.
func (p *Path) Begin(o *op.Ops) {
	*p = Path{
		ops:     &o.Internal,
		macro:   op.Record(o),
		contour: 1,
	}
	p.hash.SetSeed(pathSeed)
	ops.BeginMulti(p.ops)
	data := ops.WriteMulti(p.ops, ops.TypeAuxLen)
	data[0] = byte(ops.TypeAux)
}

// End returns a PathSpec ready to use in clipping operations.
func (p *Path) End() PathSpec {
	p.gap()
	c := p.macro.Stop()
	ops.EndMulti(p.ops)
	return PathSpec{
		spec:        c,
		hasSegments: p.hasSegments,
		bounds:      p.bounds.Round(),
		hash:        p.hash.Sum64(),
	}
}

// Move moves the pen by the amount specified by delta.
func (p *Path) Move(delta f32.Point) {
	to := delta.Add(p.pen)
	p.MoveTo(to)
}

// MoveTo moves the pen to the specified absolute coordinate.
func (p *Path) MoveTo(to f32.Point) {
	if p.pen == to {
		return
	}
	p.gap()
	p.end()
	p.pen = to
	p.start = to
}

func (p *Path) gap() {
	if p.pen != p.start {
		// A closed contour starts and ends in the same point.
		// This move creates a gap in the contour, register it.
		data := ops.WriteMulti(p.ops, scene.CommandSize+4)
		bo := binary.LittleEndian
		bo.PutUint32(data[0:], uint32(p.contour))
		p.cmd(data[4:], scene.Gap(p.pen, p.start))
	}
}

// end completes the current contour.
func (p *Path) end() {
	p.contour++
}

// Line moves the pen by the amount specified by delta, recording a line.
func (p *Path) Line(delta f32.Point) {
	to := delta.Add(p.pen)
	p.LineTo(to)
}

// LineTo moves the pen to the absolute point specified, recording a line.
func (p *Path) LineTo(to f32.Point) {
	data := ops.WriteMulti(p.ops, scene.CommandSize+4)
	bo := binary.LittleEndian
	bo.PutUint32(data[0:], uint32(p.contour))
	p.cmd(data[4:], scene.Line(p.pen, to))
	p.pen = to
	p.expand(to)
}

func (p *Path) cmd(data []byte, c scene.Command) {
	ops.EncodeCommand(data, c)
	p.hash.Write(data)
}

func (p *Path) expand(pt f32.Point) {
	if !p.hasSegments {
		p.hasSegments = true
		p.bounds = f32internal.Rectangle{Min: pt, Max: pt}
	} else {
		b := p.bounds
		if pt.X < b.Min.X {
			b.Min.X = pt.X
		}
		if pt.Y < b.Min.Y {
			b.Min.Y = pt.Y
		}
		if pt.X > b.Max.X {
			b.Max.X = pt.X
		}
		if pt.Y > b.Max.Y {
			b.Max.Y = pt.Y
		}
		p.bounds = b
	}
}

// Quad records a quadratic Bézier from the pen to end
// with the control point ctrl.
func (p *Path) Quad(ctrl, to f32.Point) {
	ctrl = ctrl.Add(p.pen)
	to = to.Add(p.pen)
	p.QuadTo(ctrl, to)
}

// QuadTo records a quadratic Bézier from the pen to end
// with the control point ctrl, with absolute coordinates.
func (p *Path) QuadTo(ctrl, to f32.Point) {
	data := ops.WriteMulti(p.ops, scene.CommandSize+4)
	bo := binary.LittleEndian
	bo.PutUint32(data[0:], uint32(p.contour))
	p.cmd(data[4:], scene.Quad(p.pen, ctrl, to))
	p.pen = to
	p.expand(ctrl)
	p.expand(to)
}

// ArcTo adds an elliptical arc to the path. The implied ellipse is defined
// by its focus points f1 and f2.
// The arc starts in the current point and ends angle radians along the ellipse boundary.
// The sign of angle determines the direction; positive being counter-clockwise,
// negative clockwise.
func (p *Path) ArcTo(f1, f2 f32.Point, angle float32) {
	m, segments := stroke.ArcTransform(p.pen, f1, f2, angle)
	for i := 0; i < segments; i++ {
		p0 := p.pen
		p1 := m.Transform(p0)
		p2 := m.Transform(p1)
		ctl := p1.Mul(2).Sub(p0.Add(p2).Mul(.5))
		p.QuadTo(ctl, p2)
	}
}

// Arc is like ArcTo where f1 and f2 are relative to the current position.
func (p *Path) Arc(f1, f2 f32.Point, angle float32) {
	f1 = f1.Add(p.pen)
	f2 = f2.Add(p.pen)
	p.ArcTo(f1, f2, angle)
}

// Cube records a cubic Bézier from the pen through
// two control points ending in to.
func (p *Path) Cube(ctrl0, ctrl1, to f32.Point) {
	p.CubeTo(p.pen.Add(ctrl0), p.pen.Add(ctrl1), p.pen.Add(to))
}

// CubeTo records a cubic Bézier from the pen through
// two control points ending in to, with absolute coordinates.
func (p *Path) CubeTo(ctrl0, ctrl1, to f32.Point) {
	if ctrl0 == p.pen && ctrl1 == p.pen && to == p.pen {
		return
	}
	data := ops.WriteMulti(p.ops, scene.CommandSize+4)
	bo := binary.LittleEndian
	bo.PutUint32(data[0:], uint32(p.contour))
	p.cmd(data[4:], scene.Cubic(p.pen, ctrl0, ctrl1, to))
	p.pen = to
	p.expand(ctrl0)
	p.expand(ctrl1)
	p.expand(to)
}

// Close closes the path contour.
func (p *Path) Close() {
	if p.pen != p.start {
		p.LineTo(p.start)
	}
	p.end()
}

// Stroke represents a stroked path.
type Stroke struct {
	Path PathSpec
	// Width of the stroked path.
	Width float32
}

// Op returns a clip operation representing the stroke.
func (s Stroke) Op() Op {
	return Op{
		path:  s.Path,
		width: s.Width,
	}
}

// Outline represents the area inside of a path, according to the
// non-zero winding rule.
type Outline struct {
	Path PathSpec
}

// Op returns a clip operation representing the outline.
func (o Outline) Op() Op {
	return Op{
		path:    o.Path,
		outline: true,
	}
}
