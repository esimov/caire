// SPDX-License-Identifier: Unlicense OR MIT

/*
Package f32 is an internal version of the public package f32 with
extra types for internal use.
*/
package f32

import (
	"image"
	"math"

	"gioui.org/f32"
)

type Point = f32.Point

type Affine2D = f32.Affine2D

var NewAffine2D = f32.NewAffine2D

// A Rectangle contains the points (X, Y) where Min.X <= X < Max.X,
// Min.Y <= Y < Max.Y.
type Rectangle struct {
	Min, Max Point
}

// String return a string representation of r.
func (r Rectangle) String() string {
	return r.Min.String() + "-" + r.Max.String()
}

// Rect is a shorthand for Rectangle{Point{x0, y0}, Point{x1, y1}}.
// The returned Rectangle has x0 and y0 swapped if necessary so that
// it's correctly formed.
func Rect(x0, y0, x1, y1 float32) Rectangle {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return Rectangle{Point{x0, y0}, Point{x1, y1}}
}

// Pt is shorthand for Point{X: x, Y: y}.
var Pt = f32.Pt

// Size returns r's width and height.
func (r Rectangle) Size() Point {
	return Point{X: r.Dx(), Y: r.Dy()}
}

// Dx returns r's width.
func (r Rectangle) Dx() float32 {
	return r.Max.X - r.Min.X
}

// Dy returns r's Height.
func (r Rectangle) Dy() float32 {
	return r.Max.Y - r.Min.Y
}

// Intersect returns the intersection of r and s.
func (r Rectangle) Intersect(s Rectangle) Rectangle {
	if r.Min.X < s.Min.X {
		r.Min.X = s.Min.X
	}
	if r.Min.Y < s.Min.Y {
		r.Min.Y = s.Min.Y
	}
	if r.Max.X > s.Max.X {
		r.Max.X = s.Max.X
	}
	if r.Max.Y > s.Max.Y {
		r.Max.Y = s.Max.Y
	}
	if r.Empty() {
		return Rectangle{}
	}
	return r
}

// Union returns the union of r and s.
func (r Rectangle) Union(s Rectangle) Rectangle {
	if r.Empty() {
		return s
	}
	if s.Empty() {
		return r
	}
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

// Canon returns the canonical version of r, where Min is to
// the upper left of Max.
func (r Rectangle) Canon() Rectangle {
	if r.Max.X < r.Min.X {
		r.Min.X, r.Max.X = r.Max.X, r.Min.X
	}
	if r.Max.Y < r.Min.Y {
		r.Min.Y, r.Max.Y = r.Max.Y, r.Min.Y
	}
	return r
}

// Empty reports whether r represents the empty area.
func (r Rectangle) Empty() bool {
	return r.Min.X >= r.Max.X || r.Min.Y >= r.Max.Y
}

// Add offsets r with the vector p.
func (r Rectangle) Add(p Point) Rectangle {
	return Rectangle{
		Point{r.Min.X + p.X, r.Min.Y + p.Y},
		Point{r.Max.X + p.X, r.Max.Y + p.Y},
	}
}

// Sub offsets r with the vector -p.
func (r Rectangle) Sub(p Point) Rectangle {
	return Rectangle{
		Point{r.Min.X - p.X, r.Min.Y - p.Y},
		Point{r.Max.X - p.X, r.Max.Y - p.Y},
	}
}

// Round returns the smallest integer rectangle that
// contains r.
func (r Rectangle) Round() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: int(floor(r.Min.X)),
			Y: int(floor(r.Min.Y)),
		},
		Max: image.Point{
			X: int(ceil(r.Max.X)),
			Y: int(ceil(r.Max.Y)),
		},
	}
}

// fRect converts a rectangle to a f32internal.Rectangle.
func FRect(r image.Rectangle) Rectangle {
	return Rectangle{
		Min: FPt(r.Min), Max: FPt(r.Max),
	}
}

// Fpt converts an point to a f32.Point.
func FPt(p image.Point) Point {
	return Point{
		X: float32(p.X), Y: float32(p.Y),
	}
}

func ceil(v float32) int {
	return int(math.Ceil(float64(v)))
}

func floor(v float32) int {
	return int(math.Floor(float64(v)))
}
