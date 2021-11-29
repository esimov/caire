// SPDX-License-Identifier: Unlicense OR MIT

/*
Package f32 is a float32 implementation of package image's
Point and Rectangle.

The coordinate space has the origin in the top left
corner with the axes extending right and down.
*/
package f32

import "strconv"

// A Point is a two dimensional point.
type Point struct {
	X, Y float32
}

// String return a string representation of p.
func (p Point) String() string {
	return "(" + strconv.FormatFloat(float64(p.X), 'f', -1, 32) +
		"," + strconv.FormatFloat(float64(p.Y), 'f', -1, 32) + ")"
}

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
func Pt(x, y float32) Point {
	return Point{X: x, Y: y}
}

// Add return the point p+p2.
func (p Point) Add(p2 Point) Point {
	return Point{X: p.X + p2.X, Y: p.Y + p2.Y}
}

// Sub returns the vector p-p2.
func (p Point) Sub(p2 Point) Point {
	return Point{X: p.X - p2.X, Y: p.Y - p2.Y}
}

// Mul returns p scaled by s.
func (p Point) Mul(s float32) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

// Div returns the vector p/s.
func (p Point) Div(s float32) Point {
	return Point{X: p.X / s, Y: p.Y / s}
}

// In reports whether p is in r.
func (p Point) In(r Rectangle) bool {
	return r.Min.X <= p.X && p.X < r.Max.X &&
		r.Min.Y <= p.Y && p.Y < r.Max.Y
}

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
