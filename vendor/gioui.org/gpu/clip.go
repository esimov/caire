package gpu

import (
	"gioui.org/f32"
	"gioui.org/internal/stroke"
)

type quadSplitter struct {
	bounds  f32.Rectangle
	contour uint32
	d       *drawOps
}

func encodeQuadTo(data []byte, meta uint32, from, ctrl, to f32.Point) {
	// NW.
	encodeVertex(data, meta, -1, 1, from, ctrl, to)
	// NE.
	encodeVertex(data[vertStride:], meta, 1, 1, from, ctrl, to)
	// SW.
	encodeVertex(data[vertStride*2:], meta, -1, -1, from, ctrl, to)
	// SE.
	encodeVertex(data[vertStride*3:], meta, 1, -1, from, ctrl, to)
}

func encodeVertex(data []byte, meta uint32, cornerx, cornery int16, from, ctrl, to f32.Point) {
	var corner float32
	if cornerx == 1 {
		corner += .5
	}
	if cornery == 1 {
		corner += .25
	}
	v := vertex{
		Corner: corner,
		FromX:  from.X,
		FromY:  from.Y,
		CtrlX:  ctrl.X,
		CtrlY:  ctrl.Y,
		ToX:    to.X,
		ToY:    to.Y,
	}
	v.encode(data, meta)
}

func (qs *quadSplitter) encodeQuadTo(from, ctrl, to f32.Point) {
	data := qs.d.writeVertCache(vertStride * 4)
	encodeQuadTo(data, qs.contour, from, ctrl, to)
}

func (qs *quadSplitter) splitAndEncode(quad stroke.QuadSegment) {
	cbnd := f32.Rectangle{
		Min: quad.From,
		Max: quad.To,
	}.Canon()
	from, ctrl, to := quad.From, quad.Ctrl, quad.To

	// If the curve contain areas where a vertical line
	// intersects it twice, split the curve in two x monotone
	// lower and upper curves. The stencil fragment program
	// expects only one intersection per curve.

	// Find the t where the derivative in x is 0.
	v0 := ctrl.Sub(from)
	v1 := to.Sub(ctrl)
	d := v0.X - v1.X
	// t = v0 / d. Split if t is in ]0;1[.
	if v0.X > 0 && d > v0.X || v0.X < 0 && d < v0.X {
		t := v0.X / d
		ctrl0 := from.Mul(1 - t).Add(ctrl.Mul(t))
		ctrl1 := ctrl.Mul(1 - t).Add(to.Mul(t))
		mid := ctrl0.Mul(1 - t).Add(ctrl1.Mul(t))
		qs.encodeQuadTo(from, ctrl0, mid)
		qs.encodeQuadTo(mid, ctrl1, to)
		if mid.X > cbnd.Max.X {
			cbnd.Max.X = mid.X
		}
		if mid.X < cbnd.Min.X {
			cbnd.Min.X = mid.X
		}
	} else {
		qs.encodeQuadTo(from, ctrl, to)
	}
	// Find the y extremum, if any.
	d = v0.Y - v1.Y
	if v0.Y > 0 && d > v0.Y || v0.Y < 0 && d < v0.Y {
		t := v0.Y / d
		y := (1-t)*(1-t)*from.Y + 2*(1-t)*t*ctrl.Y + t*t*to.Y
		if y > cbnd.Max.Y {
			cbnd.Max.Y = y
		}
		if y < cbnd.Min.Y {
			cbnd.Min.Y = y
		}
	}

	qs.bounds = unionRect(qs.bounds, cbnd)
}

// Union is like f32.Rectangle.Union but ignores empty rectangles.
func unionRect(r, s f32.Rectangle) f32.Rectangle {
	if r.Min.X > s.Min.X {
		r.Min.X = s.Min.X
	}
	if r.Min.Y > s.Min.Y {
		r.Min.Y = s.Min.Y
	}
	if r.Max.X < s.Max.X {
		r.Max.X = s.Max.X
	}
	if r.Max.Y < s.Max.Y {
		r.Max.Y = s.Max.Y
	}
	return r
}
