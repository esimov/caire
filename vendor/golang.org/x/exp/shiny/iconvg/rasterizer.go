// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iconvg

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/exp/shiny/iconvg/internal/gradient"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/vector"
)

const (
	smoothTypeNone = iota
	smoothTypeQuad
	smoothTypeCube
)

// Rasterizer is a Destination that draws an IconVG graphic onto a raster
// image.
//
// The zero value is usable, in that it has no raster image to draw onto, so
// that calling Decode with this Destination is a no-op (other than checking
// the encoded form for errors in the byte code). Call SetDstImage to change
// the raster image, before calling Decode or between calls to Decode.
type Rasterizer struct {
	z vector.Rasterizer

	dst    draw.Image
	r      image.Rectangle
	drawOp draw.Op

	// scale and bias transforms the metadata.ViewBox rectangle to the (0, 0) -
	// (r.Dx(), r.Dy()) rectangle.
	scaleX float32
	biasX  float32
	scaleY float32
	biasY  float32

	metadata Metadata

	lod0 float32
	lod1 float32
	cSel uint8
	nSel uint8

	disabled bool

	firstStartPath   bool
	prevSmoothType   uint8
	prevSmoothPointX float32
	prevSmoothPointY float32

	fill      image.Image
	flatColor color.RGBA
	flatImage image.Uniform
	gradient  gradient.Gradient

	cReg  [64]color.RGBA
	nReg  [64]float32
	stops [64]gradient.Stop
}

// SetDstImage sets the Rasterizer to draw onto a destination image, given by
// dst and r, with the given compositing operator.
//
// The IconVG graphic (which does not have a fixed size in pixels) will be
// scaled in the X and Y dimensions to fit the rectangle r. The scaling factors
// may differ in the two dimensions.
func (z *Rasterizer) SetDstImage(dst draw.Image, r image.Rectangle, drawOp draw.Op) {
	z.dst = dst
	if r.Empty() {
		r = image.Rectangle{}
	}
	z.r = r
	z.drawOp = drawOp
	z.recalcTransform()
}

// Reset resets the Rasterizer for the given Metadata.
func (z *Rasterizer) Reset(m Metadata) {
	z.metadata = m
	z.lod0 = 0
	z.lod1 = positiveInfinity
	z.cSel = 0
	z.nSel = 0
	z.firstStartPath = true
	z.prevSmoothType = smoothTypeNone
	z.prevSmoothPointX = 0
	z.prevSmoothPointY = 0
	z.cReg = m.Palette
	z.nReg = [64]float32{}
	z.recalcTransform()
}

func (z *Rasterizer) recalcTransform() {
	z.scaleX = float32(z.r.Dx()) / (z.metadata.ViewBox.Max[0] - z.metadata.ViewBox.Min[0])
	z.biasX = -z.metadata.ViewBox.Min[0]
	z.scaleY = float32(z.r.Dy()) / (z.metadata.ViewBox.Max[1] - z.metadata.ViewBox.Min[1])
	z.biasY = -z.metadata.ViewBox.Min[1]
}

func (z *Rasterizer) SetCSel(cSel uint8) { z.cSel = cSel & 0x3f }
func (z *Rasterizer) SetNSel(nSel uint8) { z.nSel = nSel & 0x3f }

func (z *Rasterizer) SetCReg(adj uint8, incr bool, c Color) {
	z.cReg[(z.cSel-adj)&0x3f] = c.Resolve(&z.metadata.Palette, &z.cReg)
	if incr {
		z.cSel++
	}
}

func (z *Rasterizer) SetNReg(adj uint8, incr bool, f float32) {
	z.nReg[(z.nSel-adj)&0x3f] = f
	if incr {
		z.nSel++
	}
}

func (z *Rasterizer) SetLOD(lod0, lod1 float32) {
	z.lod0, z.lod1 = lod0, lod1
}

func (z *Rasterizer) unabsX(x float32) float32 { return x/z.scaleX - z.biasX }
func (z *Rasterizer) unabsY(y float32) float32 { return y/z.scaleY - z.biasY }

func (z *Rasterizer) absX(x float32) float32 { return z.scaleX * (x + z.biasX) }
func (z *Rasterizer) absY(y float32) float32 { return z.scaleY * (y + z.biasY) }
func (z *Rasterizer) relX(x float32) float32 { return z.scaleX * x }
func (z *Rasterizer) relY(y float32) float32 { return z.scaleY * y }

func (z *Rasterizer) absVec2(x, y float32) (zx, zy float32) {
	return z.absX(x), z.absY(y)
}

func (z *Rasterizer) relVec2(x, y float32) (zx, zy float32) {
	px, py := z.z.Pen()
	return px + z.relX(x), py + z.relY(y)
}

// implicitSmoothPoint returns the implicit control point for smooth-quadratic
// and smooth-cubic Bézier curves.
//
// https://www.w3.org/TR/SVG/paths.html#PathDataCurveCommands says, "The first
// control point is assumed to be the reflection of the second control point on
// the previous command relative to the current point. (If there is no previous
// command or if the previous command was not [a quadratic or cubic command],
// assume the first control point is coincident with the current point.)"
func (z *Rasterizer) implicitSmoothPoint(thisSmoothType uint8) (zx, zy float32) {
	px, py := z.z.Pen()
	if z.prevSmoothType != thisSmoothType {
		return px, py
	}
	return 2*px - z.prevSmoothPointX, 2*py - z.prevSmoothPointY
}

func (z *Rasterizer) initGradient(rgba color.RGBA) (ok bool) {
	nStops := int(rgba.R & 0x3f)
	cBase := int(rgba.G & 0x3f)
	nBase := int(rgba.B & 0x3f)
	prevN := negativeInfinity
	for i := 0; i < nStops; i++ {
		c := z.cReg[(cBase+i)&0x3f]
		if !validAlphaPremulColor(c) {
			return false
		}
		n := z.nReg[(nBase+i)&0x3f]
		if !(0 <= n && n <= 1) || !(n > prevN) {
			return false
		}
		prevN = n
		z.stops[i] = gradient.Stop{
			Offset: float64(n),
			RGBA64: color.RGBA64{
				R: uint16(c.R) * 0x101,
				G: uint16(c.G) * 0x101,
				B: uint16(c.B) * 0x101,
				A: uint16(c.A) * 0x101,
			},
		}
	}

	// The affine transformation matrix in the IconVG graphic, stored in 6
	// contiguous NREG registers, goes from graphic coordinate space (i.e. the
	// metadata viewBox) to the gradient coordinate space. We need it to start
	// in pixel space, not graphic coordinate space.

	invZSX := 1 / float64(z.scaleX)
	invZSY := 1 / float64(z.scaleY)
	zBX := float64(z.biasX)
	zBY := float64(z.biasY)

	a := float64(z.nReg[(nBase-6)&0x3f])
	b := float64(z.nReg[(nBase-5)&0x3f])
	c := float64(z.nReg[(nBase-4)&0x3f])
	d := float64(z.nReg[(nBase-3)&0x3f])
	e := float64(z.nReg[(nBase-2)&0x3f])
	f := float64(z.nReg[(nBase-1)&0x3f])

	pix2Grad := f64.Aff3{
		a * invZSX,
		b * invZSY,
		c - a*zBX - b*zBY,
		d * invZSX,
		e * invZSY,
		f - d*zBX - e*zBY,
	}

	shape := gradient.ShapeLinear
	if (rgba.B>>6)&0x01 != 0 {
		shape = gradient.ShapeRadial
	}
	z.gradient.Init(
		shape,
		gradient.Spread(rgba.G>>6),
		pix2Grad,
		z.stops[:nStops],
	)

	return true
}

func (z *Rasterizer) StartPath(adj uint8, x, y float32) {
	z.flatColor = z.cReg[(z.cSel-adj)&0x3f]
	if validAlphaPremulColor(z.flatColor) {
		z.flatImage.C = &z.flatColor
		z.fill = &z.flatImage
		z.disabled = z.flatColor.A == 0
	} else if z.flatColor.A == 0x00 && z.flatColor.B&0x80 != 0 {
		z.fill = &z.gradient
		z.disabled = !z.initGradient(z.flatColor)
	} else {
		z.fill = nil
		z.disabled = true
	}

	width, height := z.r.Dx(), z.r.Dy()
	h := float32(height)
	z.disabled = z.disabled || !(z.lod0 <= h && h < z.lod1)
	if z.disabled {
		return
	}

	z.z.Reset(width, height)
	if z.firstStartPath {
		z.firstStartPath = false
		z.z.DrawOp = z.drawOp
	}
	z.prevSmoothType = smoothTypeNone
	z.z.MoveTo(z.absVec2(x, y))
}

func (z *Rasterizer) ClosePathEndPath() {
	if z.disabled {
		return
	}
	z.z.ClosePath()
	if z.dst == nil {
		return
	}
	z.z.Draw(z.dst, z.r, z.fill, image.Point{})
}

func (z *Rasterizer) ClosePathAbsMoveTo(x, y float32) {
	if z.disabled {
		return
	}
	z.prevSmoothType = smoothTypeNone
	z.z.ClosePath()
	z.z.MoveTo(z.absVec2(x, y))
}

func (z *Rasterizer) ClosePathRelMoveTo(x, y float32) {
	if z.disabled {
		return
	}
	z.prevSmoothType = smoothTypeNone
	z.z.ClosePath()
	z.z.MoveTo(z.relVec2(x, y))
}

func (z *Rasterizer) AbsHLineTo(x float32) {
	if z.disabled {
		return
	}
	_, py := z.z.Pen()
	z.prevSmoothType = smoothTypeNone
	z.z.LineTo(z.absX(x), py)
}

func (z *Rasterizer) RelHLineTo(x float32) {
	if z.disabled {
		return
	}
	px, py := z.z.Pen()
	z.prevSmoothType = smoothTypeNone
	z.z.LineTo(px+z.relX(x), py)
}

func (z *Rasterizer) AbsVLineTo(y float32) {
	if z.disabled {
		return
	}
	px, _ := z.z.Pen()
	z.prevSmoothType = smoothTypeNone
	z.z.LineTo(px, z.absY(y))
}

func (z *Rasterizer) RelVLineTo(y float32) {
	if z.disabled {
		return
	}
	px, py := z.z.Pen()
	z.prevSmoothType = smoothTypeNone
	z.z.LineTo(px, py+z.relY(y))
}

func (z *Rasterizer) AbsLineTo(x, y float32) {
	if z.disabled {
		return
	}
	z.prevSmoothType = smoothTypeNone
	z.z.LineTo(z.absVec2(x, y))
}

func (z *Rasterizer) RelLineTo(x, y float32) {
	if z.disabled {
		return
	}
	z.prevSmoothType = smoothTypeNone
	z.z.LineTo(z.relVec2(x, y))
}

func (z *Rasterizer) AbsSmoothQuadTo(x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 := z.implicitSmoothPoint(smoothTypeQuad)
	x, y = z.absVec2(x, y)
	z.prevSmoothType = smoothTypeQuad
	z.prevSmoothPointX, z.prevSmoothPointY = x1, y1
	z.z.QuadTo(x1, y1, x, y)
}

func (z *Rasterizer) RelSmoothQuadTo(x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 := z.implicitSmoothPoint(smoothTypeQuad)
	x, y = z.relVec2(x, y)
	z.prevSmoothType = smoothTypeQuad
	z.prevSmoothPointX, z.prevSmoothPointY = x1, y1
	z.z.QuadTo(x1, y1, x, y)
}

func (z *Rasterizer) AbsQuadTo(x1, y1, x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 = z.absVec2(x1, y1)
	x, y = z.absVec2(x, y)
	z.prevSmoothType = smoothTypeQuad
	z.prevSmoothPointX, z.prevSmoothPointY = x1, y1
	z.z.QuadTo(x1, y1, x, y)
}

func (z *Rasterizer) RelQuadTo(x1, y1, x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 = z.relVec2(x1, y1)
	x, y = z.relVec2(x, y)
	z.prevSmoothType = smoothTypeQuad
	z.prevSmoothPointX, z.prevSmoothPointY = x1, y1
	z.z.QuadTo(x1, y1, x, y)
}

func (z *Rasterizer) AbsSmoothCubeTo(x2, y2, x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 := z.implicitSmoothPoint(smoothTypeCube)
	x2, y2 = z.absVec2(x2, y2)
	x, y = z.absVec2(x, y)
	z.prevSmoothType = smoothTypeCube
	z.prevSmoothPointX, z.prevSmoothPointY = x2, y2
	z.z.CubeTo(x1, y1, x2, y2, x, y)
}

func (z *Rasterizer) RelSmoothCubeTo(x2, y2, x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 := z.implicitSmoothPoint(smoothTypeCube)
	x2, y2 = z.relVec2(x2, y2)
	x, y = z.relVec2(x, y)
	z.prevSmoothType = smoothTypeCube
	z.prevSmoothPointX, z.prevSmoothPointY = x2, y2
	z.z.CubeTo(x1, y1, x2, y2, x, y)
}

func (z *Rasterizer) AbsCubeTo(x1, y1, x2, y2, x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 = z.absVec2(x1, y1)
	x2, y2 = z.absVec2(x2, y2)
	x, y = z.absVec2(x, y)
	z.prevSmoothType = smoothTypeCube
	z.prevSmoothPointX, z.prevSmoothPointY = x2, y2
	z.z.CubeTo(x1, y1, x2, y2, x, y)
}

func (z *Rasterizer) RelCubeTo(x1, y1, x2, y2, x, y float32) {
	if z.disabled {
		return
	}
	x1, y1 = z.relVec2(x1, y1)
	x2, y2 = z.relVec2(x2, y2)
	x, y = z.relVec2(x, y)
	z.prevSmoothType = smoothTypeCube
	z.prevSmoothPointX, z.prevSmoothPointY = x2, y2
	z.z.CubeTo(x1, y1, x2, y2, x, y)
}

func (z *Rasterizer) AbsArcTo(rx, ry, xAxisRotation float32, largeArc, sweep bool, x, y float32) {
	if z.disabled {
		return
	}
	z.prevSmoothType = smoothTypeNone

	// We follow the "Conversion from endpoint to center parameterization"
	// algorithm as per
	// https://www.w3.org/TR/SVG/implnote.html#ArcConversionEndpointToCenter

	// There seems to be a bug in the spec's "implementation notes".
	//
	// Actual implementations, such as
	//	- https://git.gnome.org/browse/librsvg/tree/rsvg-path.c
	//	- http://svn.apache.org/repos/asf/xmlgraphics/batik/branches/svg11/sources/org/apache/batik/ext/awt/geom/ExtendedGeneralPath.java
	//	- https://java.net/projects/svgsalamander/sources/svn/content/trunk/svg-core/src/main/java/com/kitfox/svg/pathcmd/Arc.java
	//	- https://github.com/millermedeiros/SVGParser/blob/master/com/millermedeiros/geom/SVGArc.as
	// do something slightly different (marked with a †).

	// (†) The Abs isn't part of the spec. Neither is checking that Rx and Ry
	// are non-zero (and non-NaN).
	Rx := math.Abs(float64(rx))
	Ry := math.Abs(float64(ry))
	if !(Rx > 0 && Ry > 0) {
		z.z.LineTo(x, y)
		return
	}

	// We work in IconVG coordinates (e.g. from -32 to +32 by default), rather
	// than destination image coordinates (e.g. the width of the dst image),
	// since the rx and ry radii also need to be scaled, but their scaling
	// factors can be different, and aren't trivial to calculate due to
	// xAxisRotation.
	//
	// We convert back to destination image coordinates via absX and absY calls
	// later, during arcSegmentTo.
	penX, penY := z.z.Pen()
	x1 := float64(z.unabsX(penX))
	y1 := float64(z.unabsY(penY))
	x2 := float64(x)
	y2 := float64(y)

	phi := 2 * math.Pi * float64(xAxisRotation)

	// Step 1: Compute (x1′, y1′)
	halfDx := (x1 - x2) / 2
	halfDy := (y1 - y2) / 2
	cosPhi := math.Cos(phi)
	sinPhi := math.Sin(phi)
	x1Prime := +cosPhi*halfDx + sinPhi*halfDy
	y1Prime := -sinPhi*halfDx + cosPhi*halfDy

	// Step 2: Compute (cx′, cy′)
	rxSq := Rx * Rx
	rySq := Ry * Ry
	x1PrimeSq := x1Prime * x1Prime
	y1PrimeSq := y1Prime * y1Prime

	// (†) Check that the radii are large enough.
	radiiCheck := x1PrimeSq/rxSq + y1PrimeSq/rySq
	if radiiCheck > 1 {
		c := math.Sqrt(radiiCheck)
		Rx *= c
		Ry *= c
		rxSq = Rx * Rx
		rySq = Ry * Ry
	}

	denom := rxSq*y1PrimeSq + rySq*x1PrimeSq
	step2 := 0.0
	if a := rxSq*rySq/denom - 1; a > 0 {
		step2 = math.Sqrt(a)
	}
	if largeArc == sweep {
		step2 = -step2
	}
	cxPrime := +step2 * Rx * y1Prime / Ry
	cyPrime := -step2 * Ry * x1Prime / Rx

	// Step 3: Compute (cx, cy) from (cx′, cy′)
	cx := +cosPhi*cxPrime - sinPhi*cyPrime + (x1+x2)/2
	cy := +sinPhi*cxPrime + cosPhi*cyPrime + (y1+y2)/2

	// Step 4: Compute θ1 and Δθ
	ax := (+x1Prime - cxPrime) / Rx
	ay := (+y1Prime - cyPrime) / Ry
	bx := (-x1Prime - cxPrime) / Rx
	by := (-y1Prime - cyPrime) / Ry
	theta1 := angle(1, 0, ax, ay)
	deltaTheta := angle(ax, ay, bx, by)
	if sweep {
		if deltaTheta < 0 {
			deltaTheta += 2 * math.Pi
		}
	} else {
		if deltaTheta > 0 {
			deltaTheta -= 2 * math.Pi
		}
	}

	// This ends the
	// https://www.w3.org/TR/SVG/implnote.html#ArcConversionEndpointToCenter
	// algorithm. What follows below is specific to this implementation.

	// We approximate an arc by one or more cubic Bézier curves.
	n := int(math.Ceil(math.Abs(deltaTheta) / (math.Pi/2 + 0.001)))
	for i := 0; i < n; i++ {
		z.arcSegmentTo(cx, cy,
			theta1+deltaTheta*float64(i+0)/float64(n),
			theta1+deltaTheta*float64(i+1)/float64(n),
			Rx, Ry, cosPhi, sinPhi,
		)
	}
}

// arcSegmentTo approximates an arc by a cubic Bézier curve. The mathematical
// formulae for the control points are the same as that used by librsvg.
func (z *Rasterizer) arcSegmentTo(cx, cy, theta1, theta2, rx, ry, cosPhi, sinPhi float64) {
	halfDeltaTheta := (theta2 - theta1) * 0.5
	q := math.Sin(halfDeltaTheta * 0.5)
	t := (8 * q * q) / (3 * math.Sin(halfDeltaTheta))
	cos1 := math.Cos(theta1)
	sin1 := math.Sin(theta1)
	cos2 := math.Cos(theta2)
	sin2 := math.Sin(theta2)
	x1 := rx * (+cos1 - t*sin1)
	y1 := ry * (+sin1 + t*cos1)
	x2 := rx * (+cos2 + t*sin2)
	y2 := ry * (+sin2 - t*cos2)
	x3 := rx * (+cos2)
	y3 := ry * (+sin2)
	z.z.CubeTo(
		z.absX(float32(cx+cosPhi*x1-sinPhi*y1)),
		z.absY(float32(cy+sinPhi*x1+cosPhi*y1)),
		z.absX(float32(cx+cosPhi*x2-sinPhi*y2)),
		z.absY(float32(cy+sinPhi*x2+cosPhi*y2)),
		z.absX(float32(cx+cosPhi*x3-sinPhi*y3)),
		z.absY(float32(cy+sinPhi*x3+cosPhi*y3)),
	)
}

func (z *Rasterizer) RelArcTo(rx, ry, xAxisRotation float32, largeArc, sweep bool, x, y float32) {
	ax, ay := z.relVec2(x, y)
	z.AbsArcTo(rx, ry, xAxisRotation, largeArc, sweep, z.unabsX(ax), z.unabsY(ay))
}

// angle returns the angle between the u and v vectors.
func angle(ux, uy, vx, vy float64) float64 {
	uNorm := math.Sqrt(ux*ux + uy*uy)
	vNorm := math.Sqrt(vx*vx + vy*vy)
	norm := uNorm * vNorm
	cos := (ux*vx + uy*vy) / norm
	ret := 0.0
	if cos <= -1 {
		ret = math.Pi
	} else if cos >= +1 {
		ret = 0
	} else {
		ret = math.Acos(cos)
	}
	if ux*vy < uy*vx {
		return -ret
	}
	return +ret
}
