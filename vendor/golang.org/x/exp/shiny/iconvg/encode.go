// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iconvg

import (
	"errors"
	"image/color"
	"math"

	"golang.org/x/image/math/f32"
)

var (
	errCSELUsedAsBothGradientAndStop = errors.New("iconvg: CSEL used as both gradient and stop")
	errDrawingOpsUsedInStylingMode   = errors.New("iconvg: drawing ops used in styling mode")
	errInvalidSelectorAdjustment     = errors.New("iconvg: invalid selector adjustment")
	errInvalidIncrementingAdjustment = errors.New("iconvg: invalid incrementing adjustment")
	errStylingOpsUsedInDrawingMode   = errors.New("iconvg: styling ops used in drawing mode")
	errTooManyGradientStops          = errors.New("iconvg: too many gradient stops")
)

type mode uint8

const (
	modeInitial mode = iota
	modeStyling
	modeDrawing
)

// Encoder is an IconVG encoder.
//
// The zero value is usable. Calling Reset, which is optional, sets the
// Metadata for the subsequent encoded form. If Reset is not called before
// other Encoder methods, the default metadata is implied.
//
// It aims to emit byte-identical Bytes output for the same input, independent
// of the platform (and specifically its floating-point hardware).
type Encoder struct {
	// HighResolutionCoordinates is whether the encoder should encode
	// coordinate numbers for subsequent paths at the best possible resolution
	// afforded by the underlying graphic format.
	//
	// By default (false), the encoder quantizes coordinates to 1/64th of a
	// unit if possible (the default graphic size is 64 by 64 units, so
	// 1/4096th of the default width or height). Each such coordinate can
	// therefore be encoded in either 1 or 2 bytes. If true, some coordinates
	// will be encoded in 4 bytes, giving greater accuracy but larger file
	// sizes. On the Material Design icon set, the 950 or so icons take up
	// around 40% more bytes (172K vs 123K) at high resolution.
	//
	// See the package documentation for more details on the coordinate number
	// encoding format.
	HighResolutionCoordinates bool

	// highResolutionCoordinates is a local copy, copied during StartPath, to
	// avoid having to specify the semantics of modifying the exported field
	// while drawing.
	highResolutionCoordinates bool

	buf      buffer
	altBuf   buffer
	metadata Metadata
	err      error

	lod0 float32
	lod1 float32
	cSel uint8
	nSel uint8

	mode     mode
	drawOp   byte
	drawArgs []float32

	scratch [12]byte
}

// Bytes returns the encoded form.
func (e *Encoder) Bytes() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}
	if e.mode == modeInitial {
		e.appendDefaultMetadata()
	}
	return []byte(e.buf), nil
}

// Reset resets the Encoder for the given Metadata.
//
// This includes setting e.HighResolutionCoordinates to false.
func (e *Encoder) Reset(m Metadata) {
	*e = Encoder{
		buf:      append(e.buf[:0], magic...),
		metadata: m,
		mode:     modeStyling,
		lod1:     positiveInfinity,
	}

	nMetadataChunks := 0
	mcViewBox := m.ViewBox != DefaultViewBox
	if mcViewBox {
		nMetadataChunks++
	}
	mcSuggestedPalette := m.Palette != DefaultPalette
	if mcSuggestedPalette {
		nMetadataChunks++
	}
	e.buf.encodeNatural(uint32(nMetadataChunks))

	if mcViewBox {
		e.altBuf = e.altBuf[:0]
		e.altBuf.encodeNatural(midViewBox)
		e.altBuf.encodeCoordinate(m.ViewBox.Min[0])
		e.altBuf.encodeCoordinate(m.ViewBox.Min[1])
		e.altBuf.encodeCoordinate(m.ViewBox.Max[0])
		e.altBuf.encodeCoordinate(m.ViewBox.Max[1])

		e.buf.encodeNatural(uint32(len(e.altBuf)))
		e.buf = append(e.buf, e.altBuf...)
	}

	if mcSuggestedPalette {
		n := 63
		for ; n >= 0 && m.Palette[n] == (color.RGBA{0x00, 0x00, 0x00, 0xff}); n-- {
		}

		// Find the shortest encoding that can represent all of m.Palette's n+1
		// explicit colors.
		enc1, enc2, enc3 := true, true, true
		for _, c := range m.Palette[:n+1] {
			if enc1 && (!is1(c.R) || !is1(c.G) || !is1(c.B) || !is1(c.A)) {
				enc1 = false
			}
			if enc2 && (!is2(c.R) || !is2(c.G) || !is2(c.B) || !is2(c.A)) {
				enc2 = false
			}
			if enc3 && (c.A != 0xff) {
				enc3 = false
			}
		}

		e.altBuf = e.altBuf[:0]
		e.altBuf.encodeNatural(midSuggestedPalette)
		if enc1 {
			e.altBuf = append(e.altBuf, byte(n)|0x00)
			for _, c := range m.Palette[:n+1] {
				x, _ := encodeColor1(RGBAColor(c))
				e.altBuf = append(e.altBuf, x)
			}
		} else if enc2 {
			e.altBuf = append(e.altBuf, byte(n)|0x40)
			for _, c := range m.Palette[:n+1] {
				x, _ := encodeColor2(RGBAColor(c))
				e.altBuf = append(e.altBuf, x[0], x[1])
			}
		} else if enc3 {
			e.altBuf = append(e.altBuf, byte(n)|0x80)
			for _, c := range m.Palette[:n+1] {
				e.altBuf = append(e.altBuf, c.R, c.G, c.B)
			}
		} else {
			e.altBuf = append(e.altBuf, byte(n)|0xc0)
			for _, c := range m.Palette[:n+1] {
				e.altBuf = append(e.altBuf, c.R, c.G, c.B, c.A)
			}
		}

		e.buf.encodeNatural(uint32(len(e.altBuf)))
		e.buf = append(e.buf, e.altBuf...)
	}
}

func (e *Encoder) appendDefaultMetadata() {
	e.buf = append(e.buf[:0], magic...)
	e.buf = append(e.buf, 0x00) // There are zero metadata chunks.
	e.mode = modeStyling
}

func (e *Encoder) CSel() uint8 {
	if e.mode == modeInitial {
		e.appendDefaultMetadata()
	}
	return e.cSel
}

func (e *Encoder) NSel() uint8 {
	if e.mode == modeInitial {
		e.appendDefaultMetadata()
	}
	return e.nSel
}

func (e *Encoder) LOD() (lod0, lod1 float32) {
	if e.mode == modeInitial {
		e.appendDefaultMetadata()
	}
	return e.lod0, e.lod1
}

func (e *Encoder) checkModeStyling() {
	if e.mode == modeStyling {
		return
	}
	if e.mode == modeInitial {
		e.appendDefaultMetadata()
		return
	}
	e.err = errStylingOpsUsedInDrawingMode
}

func (e *Encoder) SetCSel(cSel uint8) {
	e.checkModeStyling()
	if e.err != nil {
		return
	}
	e.cSel = cSel & 0x3f
	e.buf = append(e.buf, e.cSel)
}

func (e *Encoder) SetNSel(nSel uint8) {
	e.checkModeStyling()
	if e.err != nil {
		return
	}
	e.nSel = nSel & 0x3f
	e.buf = append(e.buf, e.nSel|0x40)
}

func (e *Encoder) SetCReg(adj uint8, incr bool, c Color) {
	e.checkModeStyling()
	if e.err != nil {
		return
	}
	if adj > 6 {
		e.err = errInvalidSelectorAdjustment
		return
	}
	if incr {
		if adj != 0 {
			e.err = errInvalidIncrementingAdjustment
		}
		adj = 7
	}

	if x, ok := encodeColor1(c); ok {
		e.buf = append(e.buf, adj|0x80, x)
		return
	}
	if x, ok := encodeColor2(c); ok {
		e.buf = append(e.buf, adj|0x88, x[0], x[1])
		return
	}
	if x, ok := encodeColor3Direct(c); ok {
		e.buf = append(e.buf, adj|0x90, x[0], x[1], x[2])
		return
	}
	if x, ok := encodeColor4(c); ok {
		e.buf = append(e.buf, adj|0x98, x[0], x[1], x[2], x[3])
		return
	}
	if x, ok := encodeColor3Indirect(c); ok {
		e.buf = append(e.buf, adj|0xa0, x[0], x[1], x[2])
		return
	}
	panic("unreachable")
}

func (e *Encoder) SetNReg(adj uint8, incr bool, f float32) {
	e.checkModeStyling()
	if e.err != nil {
		return
	}
	if adj > 6 {
		e.err = errInvalidSelectorAdjustment
		return
	}
	if incr {
		if adj != 0 {
			e.err = errInvalidIncrementingAdjustment
		}
		adj = 7
	}

	// Try three different encodings and pick the shortest.
	b := buffer(e.scratch[0:0])
	opcode, iBest, nBest := uint8(0xa8), 0, b.encodeReal(f)

	b = buffer(e.scratch[4:4])
	if n := b.encodeCoordinate(f); n < nBest {
		opcode, iBest, nBest = 0xb0, 4, n
	}

	b = buffer(e.scratch[8:8])
	if n := b.encodeZeroToOne(f); n < nBest {
		opcode, iBest, nBest = 0xb8, 8, n
	}

	e.buf = append(e.buf, adj|opcode)
	e.buf = append(e.buf, e.scratch[iBest:iBest+nBest]...)
}

func (e *Encoder) SetLOD(lod0, lod1 float32) {
	e.checkModeStyling()
	if e.err != nil {
		return
	}
	e.lod0 = lod0
	e.lod1 = lod1
	e.buf = append(e.buf, 0xc7)
	e.buf.encodeReal(lod0)
	e.buf.encodeReal(lod1)
}

// SetGradient sets CREG[CSEL] to encode the gradient whose colors defined by
// spread and stops. Its geometry is either linear or radial, depending on the
// radial argument, and the given affine transformation matrix maps from
// graphic coordinate space defined by the metadata's viewBox (e.g. from (-32,
// -32) to (+32, +32)) to gradient coordinate space. Gradient coordinate space
// is where a linear gradient ranges from x=0 to x=1, and a radial gradient has
// center (0, 0) and radius 1.
//
// The colors of the n stops are encoded at CREG[cBase+0], CREG[cBase+1], ...,
// CREG[cBase+n-1]. Similarly, the offsets of the n stops are encoded at
// NREG[nBase+0], NREG[nBase+1], ..., NREG[nBase+n-1]. Additional parameters
// are stored at NREG[nBase-4], NREG[nBase-3], NREG[nBase-2] and NREG[nBase-1].
//
// The CSEL and NSEL selector registers maintain the same values after the
// method returns as they had when the method was called.
//
// See the package documentation for more details on the gradient encoding
// format and the derivation of common transformation matrices.
func (e *Encoder) SetGradient(cBase, nBase uint8, radial bool, transform f32.Aff3, spread GradientSpread, stops []GradientStop) {
	e.checkModeStyling()
	if e.err != nil {
		return
	}
	if len(stops) > 64-len(transform) {
		e.err = errTooManyGradientStops
		return
	}
	if x, y := e.cSel, e.cSel+64; (cBase <= x && x < cBase+uint8(len(stops))) ||
		(cBase <= y && y < cBase+uint8(len(stops))) {
		e.err = errCSELUsedAsBothGradientAndStop
		return
	}

	oldCSel := e.cSel
	oldNSel := e.nSel
	cBase &= 0x3f
	nBase &= 0x3f
	bFlags := uint8(0x80)
	if radial {
		bFlags = 0xc0
	}
	e.SetCReg(0, false, RGBAColor(color.RGBA{
		R: uint8(len(stops)),
		G: cBase | uint8(spread<<6),
		B: nBase | bFlags,
		A: 0x00,
	}))
	e.SetCSel(cBase)
	e.SetNSel(nBase)
	for i, v := range transform {
		e.SetNReg(uint8(len(transform)-i), false, v)
	}
	for _, s := range stops {
		r, g, b, a := s.Color.RGBA()
		e.SetCReg(0, true, RGBAColor(color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}))
		e.SetNReg(0, true, s.Offset)
	}
	e.SetCSel(oldCSel)
	e.SetNSel(oldNSel)
}

// SetLinearGradient is like SetGradient with radial=false except that the
// transformation matrix is implicitly defined by two boundary points (x1, y1)
// and (x2, y2).
func (e *Encoder) SetLinearGradient(cBase, nBase uint8, x1, y1, x2, y2 float32, spread GradientSpread, stops []GradientStop) {
	// See the package documentation's appendix for a derivation of the
	// transformation matrix.
	dx, dy := x2-x1, y2-y1
	d := dx*dx + dy*dy
	ma := dx / d
	mb := dy / d
	e.SetGradient(cBase, nBase, false, f32.Aff3{
		ma, mb, -ma*x1 - mb*y1,
		0, 0, 0,
	}, spread, stops)
}

// SetCircularGradient is like SetGradient with radial=true except that the
// transformation matrix is implicitly defined by a center (cx, cy) and a
// radius vector (rx, ry) such that (cx+rx, cy+ry) is on the circle.
func (e *Encoder) SetCircularGradient(cBase, nBase uint8, cx, cy, rx, ry float32, spread GradientSpread, stops []GradientStop) {
	// See the package documentation's appendix for a derivation of the
	// transformation matrix.
	invR := float32(1 / math.Sqrt(float64(rx*rx+ry*ry)))
	e.SetGradient(cBase, nBase, true, f32.Aff3{
		invR, 0, -cx * invR,
		0, invR, -cy * invR,
	}, spread, stops)
}

// SetEllipticalGradient is like SetGradient with radial=true except that the
// transformation matrix is implicitly defined by a center (cx, cy) and two
// axis vectors (rx, ry) and (sx, sy) such that (cx+rx, cy+ry) and (cx+sx,
// cy+sy) are on the ellipse.
func (e *Encoder) SetEllipticalGradient(cBase, nBase uint8, cx, cy, rx, ry, sx, sy float32, spread GradientSpread, stops []GradientStop) {
	// Explicitly disable FMA in the floating-point calculations below
	// to get consistent results on all platforms, and in turn produce
	// a byte-identical encoding.
	// See https://golang.org/ref/spec#Floating_point_operators and issue 43219.

	// See the package documentation's appendix for a derivation of the
	// transformation matrix.
	invRSSR := 1 / (float32(rx*sy) - float32(sx*ry))

	ma := +sy * invRSSR
	mb := -sx * invRSSR
	mc := -float32(ma*cx) - float32(mb*cy)
	md := -ry * invRSSR
	me := +rx * invRSSR
	mf := -float32(md*cx) - float32(me*cy)

	e.SetGradient(cBase, nBase, true, f32.Aff3{
		ma, mb, mc,
		md, me, mf,
	}, spread, stops)
}

func (e *Encoder) StartPath(adj uint8, x, y float32) {
	e.checkModeStyling()
	if e.err != nil {
		return
	}
	if adj > 6 {
		e.err = errInvalidSelectorAdjustment
		return
	}
	e.highResolutionCoordinates = e.HighResolutionCoordinates
	e.buf = append(e.buf, uint8(0xc0+adj))
	e.buf.encodeCoordinate(e.quantize(x))
	e.buf.encodeCoordinate(e.quantize(y))
	e.mode = modeDrawing
}

func (e *Encoder) AbsHLineTo(x float32)                   { e.draw('H', x, 0, 0, 0, 0, 0) }
func (e *Encoder) RelHLineTo(x float32)                   { e.draw('h', x, 0, 0, 0, 0, 0) }
func (e *Encoder) AbsVLineTo(y float32)                   { e.draw('V', y, 0, 0, 0, 0, 0) }
func (e *Encoder) RelVLineTo(y float32)                   { e.draw('v', y, 0, 0, 0, 0, 0) }
func (e *Encoder) AbsLineTo(x, y float32)                 { e.draw('L', x, y, 0, 0, 0, 0) }
func (e *Encoder) RelLineTo(x, y float32)                 { e.draw('l', x, y, 0, 0, 0, 0) }
func (e *Encoder) AbsSmoothQuadTo(x, y float32)           { e.draw('T', x, y, 0, 0, 0, 0) }
func (e *Encoder) RelSmoothQuadTo(x, y float32)           { e.draw('t', x, y, 0, 0, 0, 0) }
func (e *Encoder) AbsQuadTo(x1, y1, x, y float32)         { e.draw('Q', x1, y1, x, y, 0, 0) }
func (e *Encoder) RelQuadTo(x1, y1, x, y float32)         { e.draw('q', x1, y1, x, y, 0, 0) }
func (e *Encoder) AbsSmoothCubeTo(x2, y2, x, y float32)   { e.draw('S', x2, y2, x, y, 0, 0) }
func (e *Encoder) RelSmoothCubeTo(x2, y2, x, y float32)   { e.draw('s', x2, y2, x, y, 0, 0) }
func (e *Encoder) AbsCubeTo(x1, y1, x2, y2, x, y float32) { e.draw('C', x1, y1, x2, y2, x, y) }
func (e *Encoder) RelCubeTo(x1, y1, x2, y2, x, y float32) { e.draw('c', x1, y1, x2, y2, x, y) }
func (e *Encoder) ClosePathEndPath()                      { e.draw('Z', 0, 0, 0, 0, 0, 0) }
func (e *Encoder) ClosePathAbsMoveTo(x, y float32)        { e.draw('Y', x, y, 0, 0, 0, 0) }
func (e *Encoder) ClosePathRelMoveTo(x, y float32)        { e.draw('y', x, y, 0, 0, 0, 0) }

func (e *Encoder) AbsArcTo(rx, ry, xAxisRotation float32, largeArc, sweep bool, x, y float32) {
	e.arcTo('A', rx, ry, xAxisRotation, largeArc, sweep, x, y)
}

func (e *Encoder) RelArcTo(rx, ry, xAxisRotation float32, largeArc, sweep bool, x, y float32) {
	e.arcTo('a', rx, ry, xAxisRotation, largeArc, sweep, x, y)
}

func (e *Encoder) arcTo(drawOp byte, rx, ry, xAxisRotation float32, largeArc, sweep bool, x, y float32) {
	flags := uint32(0)
	if largeArc {
		flags |= 0x01
	}
	if sweep {
		flags |= 0x02
	}
	e.draw(drawOp, rx, ry, xAxisRotation, float32(flags), x, y)
}

func (e *Encoder) draw(drawOp byte, arg0, arg1, arg2, arg3, arg4, arg5 float32) {
	if e.err != nil {
		return
	}
	if e.mode != modeDrawing {
		e.err = errDrawingOpsUsedInStylingMode
		return
	}
	if e.drawOp != drawOp {
		e.flushDrawOps()
	}
	e.drawOp = drawOp
	switch drawOps[drawOp].nArgs {
	case 0:
		// No-op.
	case 1:
		e.drawArgs = append(e.drawArgs, arg0)
	case 2:
		e.drawArgs = append(e.drawArgs, arg0, arg1)
	case 4:
		e.drawArgs = append(e.drawArgs, arg0, arg1, arg2, arg3)
	case 6:
		e.drawArgs = append(e.drawArgs, arg0, arg1, arg2, arg3, arg4, arg5)
	default:
		panic("unreachable")
	}

	switch drawOp {
	case 'Z':
		e.mode = modeStyling
		fallthrough
	case 'Y', 'y':
		e.flushDrawOps()
	}
}

func (e *Encoder) flushDrawOps() {
	if e.drawOp == 0x00 {
		return
	}

	if op := drawOps[e.drawOp]; op.nArgs == 0 {
		e.buf = append(e.buf, op.opcodeBase)
	} else {
		n := len(e.drawArgs) / int(op.nArgs)
		for i := 0; n > 0; {
			m := n
			if m > int(op.maxRepCount) {
				m = int(op.maxRepCount)
			}
			e.buf = append(e.buf, op.opcodeBase+uint8(m)-1)

			switch e.drawOp {
			default:
				for j := m * int(op.nArgs); j > 0; j-- {
					e.buf.encodeCoordinate(e.quantize(e.drawArgs[i]))
					i++
				}
			case 'A', 'a':
				for j := m; j > 0; j-- {
					e.buf.encodeCoordinate(e.quantize(e.drawArgs[i+0]))
					e.buf.encodeCoordinate(e.quantize(e.drawArgs[i+1]))
					e.buf.encodeAngle(e.drawArgs[i+2])
					e.buf.encodeNatural(uint32(e.drawArgs[i+3]))
					e.buf.encodeCoordinate(e.quantize(e.drawArgs[i+4]))
					e.buf.encodeCoordinate(e.quantize(e.drawArgs[i+5]))
					i += 6
				}
			}

			n -= m
		}
	}

	e.drawOp = 0x00
	e.drawArgs = e.drawArgs[:0]
}

func (e *Encoder) quantize(coord float32) float32 {
	if !e.highResolutionCoordinates && (-128 <= coord && coord < 128) {
		x := math.Floor(float64(coord*64 + 0.5))
		return float32(x) / 64
	}
	return coord
}

var drawOps = [256]struct {
	opcodeBase  byte
	maxRepCount uint8
	nArgs       uint8
}{
	'L': {0x00, 32, 2},
	'l': {0x20, 32, 2},
	'T': {0x40, 16, 2},
	't': {0x50, 16, 2},
	'Q': {0x60, 16, 4},
	'q': {0x70, 16, 4},
	'S': {0x80, 16, 4},
	's': {0x90, 16, 4},
	'C': {0xa0, 16, 6},
	'c': {0xb0, 16, 6},
	'A': {0xc0, 16, 6},
	'a': {0xd0, 16, 6},

	// Z means close path and then end path.
	'Z': {0xe1, 1, 0},
	// Y/y means close path and then open a new path (with a MoveTo/moveTo).
	'Y': {0xe2, 1, 2},
	'y': {0xe3, 1, 2},

	'H': {0xe6, 1, 1},
	'h': {0xe7, 1, 1},
	'V': {0xe8, 1, 1},
	'v': {0xe9, 1, 1},
}
