// SPDX-License-Identifier: Unlicense OR MIT

package f32

import (
	"fmt"
	"math"
)

// Affine2D represents an affine 2D transformation. The zero value if Affine2D
// represents the identity transform.
type Affine2D struct {
	// in order to make the zero value of Affine2D represent the identity
	// transform we store it with the identity matrix subtracted, that is
	// if the actual transformation matrix is:
	// [sx, hx, ox]
	// [hy, sy, oy]
	// [ 0,  0,  1]
	// we store a = sx-1 and e = sy-1
	a, b, c float32
	d, e, f float32
}

// NewAffine2D creates a new Affine2D transform from the matrix elements
// in row major order. The rows are: [sx, hx, ox], [hy, sy, oy], [0, 0, 1].
func NewAffine2D(sx, hx, ox, hy, sy, oy float32) Affine2D {
	return Affine2D{
		a: sx - 1, b: hx, c: ox,
		d: hy, e: sy - 1, f: oy,
	}
}

// Offset the transformation.
func (a Affine2D) Offset(offset Point) Affine2D {
	return Affine2D{
		a.a, a.b, a.c + offset.X,
		a.d, a.e, a.f + offset.Y,
	}
}

// Scale the transformation around the given origin.
func (a Affine2D) Scale(origin, factor Point) Affine2D {
	if origin == (Point{}) {
		return a.scale(factor)
	}
	a = a.Offset(origin.Mul(-1))
	a = a.scale(factor)
	return a.Offset(origin)
}

// Rotate the transformation by the given angle (in radians) counter clockwise around the given origin.
func (a Affine2D) Rotate(origin Point, radians float32) Affine2D {
	if origin == (Point{}) {
		return a.rotate(radians)
	}
	a = a.Offset(origin.Mul(-1))
	a = a.rotate(radians)
	return a.Offset(origin)
}

// Shear the transformation by the given angle (in radians) around the given origin.
func (a Affine2D) Shear(origin Point, radiansX, radiansY float32) Affine2D {
	if origin == (Point{}) {
		return a.shear(radiansX, radiansY)
	}
	a = a.Offset(origin.Mul(-1))
	a = a.shear(radiansX, radiansY)
	return a.Offset(origin)
}

// Mul returns A*B.
func (A Affine2D) Mul(B Affine2D) (r Affine2D) {
	r.a = (A.a+1)*(B.a+1) + A.b*B.d - 1
	r.b = (A.a+1)*B.b + A.b*(B.e+1)
	r.c = (A.a+1)*B.c + A.b*B.f + A.c
	r.d = A.d*(B.a+1) + (A.e+1)*B.d
	r.e = A.d*B.b + (A.e+1)*(B.e+1) - 1
	r.f = A.d*B.c + (A.e+1)*B.f + A.f
	return r
}

// Invert the transformation. Note that if the matrix is close to singular
// numerical errors may become large or infinity.
func (a Affine2D) Invert() Affine2D {
	if a.a == 0 && a.b == 0 && a.d == 0 && a.e == 0 {
		return Affine2D{a: 0, b: 0, c: -a.c, d: 0, e: 0, f: -a.f}
	}
	a.a += 1
	a.e += 1
	det := a.a*a.e - a.b*a.d
	a.a, a.e = a.e/det, a.a/det
	a.b, a.d = -a.b/det, -a.d/det
	temp := a.c
	a.c = -a.a*a.c - a.b*a.f
	a.f = -a.d*temp - a.e*a.f
	a.a -= 1
	a.e -= 1
	return a
}

// Transform p by returning a*p.
func (a Affine2D) Transform(p Point) Point {
	return Point{
		X: p.X*(a.a+1) + p.Y*a.b + a.c,
		Y: p.X*a.d + p.Y*(a.e+1) + a.f,
	}
}

// Elems returns the matrix elements of the transform in row-major order. The
// rows are: [sx, hx, ox], [hy, sy, oy], [0, 0, 1].
func (a Affine2D) Elems() (sx, hx, ox, hy, sy, oy float32) {
	return a.a + 1, a.b, a.c, a.d, a.e + 1, a.f
}

func (a Affine2D) scale(factor Point) Affine2D {
	return Affine2D{
		(a.a+1)*factor.X - 1, a.b * factor.X, a.c * factor.X,
		a.d * factor.Y, (a.e+1)*factor.Y - 1, a.f * factor.Y,
	}
}

func (a Affine2D) rotate(radians float32) Affine2D {
	sin, cos := math.Sincos(float64(radians))
	s, c := float32(sin), float32(cos)
	return Affine2D{
		(a.a+1)*c - a.d*s - 1, a.b*c - (a.e+1)*s, a.c*c - a.f*s,
		(a.a+1)*s + a.d*c, a.b*s + (a.e+1)*c - 1, a.c*s + a.f*c,
	}
}

func (a Affine2D) shear(radiansX, radiansY float32) Affine2D {
	tx := float32(math.Tan(float64(radiansX)))
	ty := float32(math.Tan(float64(radiansY)))
	return Affine2D{
		(a.a + 1) + a.d*tx - 1, a.b + (a.e+1)*tx, a.c + a.f*tx,
		(a.a+1)*ty + a.d, a.b*ty + (a.e + 1) - 1, a.f*ty + a.f,
	}
}

func (a Affine2D) String() string {
	sx, hx, ox, hy, sy, oy := a.Elems()
	return fmt.Sprintf("[[%f %f %f] [%f %f %f]]", sx, hx, ox, hy, sy, oy)
}
