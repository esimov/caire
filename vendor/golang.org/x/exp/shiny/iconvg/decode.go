// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iconvg

import (
	"bytes"
	"errors"
	"image/color"
)

var (
	errInconsistentMetadataChunkLength = errors.New("iconvg: inconsistent metadata chunk length")
	errInvalidColor                    = errors.New("iconvg: invalid color")
	errInvalidMagicIdentifier          = errors.New("iconvg: invalid magic identifier")
	errInvalidMetadataChunkLength      = errors.New("iconvg: invalid metadata chunk length")
	errInvalidMetadataIdentifier       = errors.New("iconvg: invalid metadata identifier")
	errInvalidNumber                   = errors.New("iconvg: invalid number")
	errInvalidNumberOfMetadataChunks   = errors.New("iconvg: invalid number of metadata chunks")
	errInvalidSuggestedPalette         = errors.New("iconvg: invalid suggested palette")
	errInvalidViewBox                  = errors.New("iconvg: invalid view box")
	errUnsupportedDrawingOpcode        = errors.New("iconvg: unsupported drawing opcode")
	errUnsupportedMetadataIdentifier   = errors.New("iconvg: unsupported metadata identifier")
	errUnsupportedStylingOpcode        = errors.New("iconvg: unsupported styling opcode")
)

var midDescriptions = [...]string{
	midViewBox:          "viewBox",
	midSuggestedPalette: "suggested palette",
}

// Destination handles the actions decoded from an IconVG graphic's opcodes.
//
// When passed to Decode, the first method called (if any) will be Reset. No
// methods will be called at all if an error is encountered in the encoded form
// before the metadata is fully decoded.
type Destination interface {
	Reset(m Metadata)

	SetCSel(cSel uint8)
	SetNSel(nSel uint8)
	SetCReg(adj uint8, incr bool, c Color)
	SetNReg(adj uint8, incr bool, f float32)
	SetLOD(lod0, lod1 float32)

	StartPath(adj uint8, x, y float32)
	ClosePathEndPath()
	ClosePathAbsMoveTo(x, y float32)
	ClosePathRelMoveTo(x, y float32)

	AbsHLineTo(x float32)
	RelHLineTo(x float32)
	AbsVLineTo(y float32)
	RelVLineTo(y float32)
	AbsLineTo(x, y float32)
	RelLineTo(x, y float32)
	AbsSmoothQuadTo(x, y float32)
	RelSmoothQuadTo(x, y float32)
	AbsQuadTo(x1, y1, x, y float32)
	RelQuadTo(x1, y1, x, y float32)
	AbsSmoothCubeTo(x2, y2, x, y float32)
	RelSmoothCubeTo(x2, y2, x, y float32)
	AbsCubeTo(x1, y1, x2, y2, x, y float32)
	RelCubeTo(x1, y1, x2, y2, x, y float32)
	AbsArcTo(rx, ry, xAxisRotation float32, largeArc, sweep bool, x, y float32)
	RelArcTo(rx, ry, xAxisRotation float32, largeArc, sweep bool, x, y float32)
}

type printer func(b []byte, format string, args ...interface{})

// DecodeOptions are the optional parameters to the Decode function.
type DecodeOptions struct {
	// Palette is an optional 64 color palette. If one isn't provided, the
	// IconVG graphic's suggested palette will be used.
	Palette *Palette
}

// DecodeMetadata decodes only the metadata in an IconVG graphic.
func DecodeMetadata(src []byte) (m Metadata, err error) {
	m.ViewBox = DefaultViewBox
	m.Palette = DefaultPalette
	if err = decode(nil, nil, &m, true, src, nil); err != nil {
		return Metadata{}, err
	}
	return m, nil
}

// Decode decodes an IconVG graphic.
func Decode(dst Destination, src []byte, opts *DecodeOptions) error {
	m := Metadata{
		ViewBox: DefaultViewBox,
		Palette: DefaultPalette,
	}
	if opts != nil && opts.Palette != nil {
		m.Palette = *opts.Palette
	}
	return decode(dst, nil, &m, false, src, opts)
}

func decode(dst Destination, p printer, m *Metadata, metadataOnly bool, src buffer, opts *DecodeOptions) (err error) {
	if !bytes.HasPrefix(src, magicBytes) {
		return errInvalidMagicIdentifier
	}
	if p != nil {
		p(src[:len(magic)], "IconVG Magic identifier\n")
	}
	src = src[len(magic):]

	nMetadataChunks, n := src.decodeNatural()
	if n == 0 {
		return errInvalidNumberOfMetadataChunks
	}
	if p != nil {
		p(src[:n], "Number of metadata chunks: %d\n", nMetadataChunks)
	}
	src = src[n:]

	for ; nMetadataChunks > 0; nMetadataChunks-- {
		src, err = decodeMetadataChunk(p, m, src, opts)
		if err != nil {
			return err
		}
	}
	if metadataOnly {
		return nil
	}
	if dst != nil {
		dst.Reset(*m)
	}

	mf := modeFunc(decodeStyling)
	for len(src) > 0 {
		mf, src, err = mf(dst, p, src)
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeMetadataChunk(p printer, m *Metadata, src buffer, opts *DecodeOptions) (src1 buffer, err error) {
	length, n := src.decodeNatural()
	if n == 0 {
		return nil, errInvalidMetadataChunkLength
	}
	if p != nil {
		p(src[:n], "Metadata chunk length: %d\n", length)
	}
	src = src[n:]
	lenSrcWant := int64(len(src)) - int64(length)

	mid, n := src.decodeNatural()
	if n == 0 {
		return nil, errInvalidMetadataIdentifier
	}
	if mid >= uint32(len(midDescriptions)) {
		return nil, errUnsupportedMetadataIdentifier
	}
	if p != nil {
		p(src[:n], "Metadata Identifier: %d (%s)\n", mid, midDescriptions[mid])
	}
	src = src[n:]

	switch mid {
	case midViewBox:
		if m.ViewBox.Min[0], src, err = decodeNumber(p, src, buffer.decodeCoordinate); err != nil {
			return nil, errInvalidViewBox
		}
		if m.ViewBox.Min[1], src, err = decodeNumber(p, src, buffer.decodeCoordinate); err != nil {
			return nil, errInvalidViewBox
		}
		if m.ViewBox.Max[0], src, err = decodeNumber(p, src, buffer.decodeCoordinate); err != nil {
			return nil, errInvalidViewBox
		}
		if m.ViewBox.Max[1], src, err = decodeNumber(p, src, buffer.decodeCoordinate); err != nil {
			return nil, errInvalidViewBox
		}
		if m.ViewBox.Min[0] > m.ViewBox.Max[0] || m.ViewBox.Min[1] > m.ViewBox.Max[1] ||
			isNaNOrInfinity(m.ViewBox.Min[0]) || isNaNOrInfinity(m.ViewBox.Min[1]) ||
			isNaNOrInfinity(m.ViewBox.Max[0]) || isNaNOrInfinity(m.ViewBox.Max[1]) {
			return nil, errInvalidViewBox
		}

	case midSuggestedPalette:
		if len(src) == 0 {
			return nil, errInvalidSuggestedPalette
		}
		length, format := 1+int(src[0]&0x3f), src[0]>>6
		decode := buffer.decodeColor4
		switch format {
		case 0:
			decode = buffer.decodeColor1
		case 1:
			decode = buffer.decodeColor2
		case 2:
			decode = buffer.decodeColor3Direct
		}
		if p != nil {
			p(src[:1], "    %d palette colors, %d bytes per color\n", length, 1+format)
		}
		src = src[1:]

		for i := 0; i < length; i++ {
			c, n := decode(src)
			if n == 0 {
				return nil, errInvalidSuggestedPalette
			}
			rgba := c.rgba()
			if c.typ != ColorTypeRGBA || !validAlphaPremulColor(rgba) {
				rgba = color.RGBA{0x00, 0x00, 0x00, 0xff}
			}
			if p != nil {
				p(src[:n], "    RGBA %02x%02x%02x%02x\n", rgba.R, rgba.G, rgba.B, rgba.A)
			}
			src = src[n:]
			if opts == nil || opts.Palette == nil {
				m.Palette[i] = rgba
			}
		}

	default:
		return nil, errUnsupportedMetadataIdentifier
	}

	if int64(len(src)) != lenSrcWant {
		return nil, errInconsistentMetadataChunkLength
	}
	return src, nil
}

// modeFunc is the decoding mode: whether we are decoding styling or drawing
// opcodes.
//
// It is a function type. The decoding loop calls this function to decode and
// execute the next opcode from the src buffer, returning the subsequent mode
// and the remaining source bytes.
type modeFunc func(dst Destination, p printer, src buffer) (modeFunc, buffer, error)

func decodeStyling(dst Destination, p printer, src buffer) (modeFunc, buffer, error) {
	switch opcode := src[0]; {
	case opcode < 0x80:
		if opcode < 0x40 {
			opcode &= 0x3f
			if p != nil {
				p(src[:1], "Set CSEL = %d\n", opcode)
			}
			src = src[1:]
			if dst != nil {
				dst.SetCSel(opcode)
			}
		} else {
			opcode &= 0x3f
			if p != nil {
				p(src[:1], "Set NSEL = %d\n", opcode)
			}
			src = src[1:]
			if dst != nil {
				dst.SetNSel(opcode)
			}
		}
		return decodeStyling, src, nil
	case opcode < 0xa8:
		return decodeSetCReg(dst, p, src, opcode)
	case opcode < 0xc0:
		return decodeSetNReg(dst, p, src, opcode)
	case opcode < 0xc7:
		return decodeStartPath(dst, p, src, opcode)
	case opcode == 0xc7:
		return decodeSetLOD(dst, p, src)
	}
	return nil, nil, errUnsupportedStylingOpcode
}

func decodeSetCReg(dst Destination, p printer, src buffer, opcode byte) (modeFunc, buffer, error) {
	nBytes, directness, adj := 0, "", opcode&0x07
	var decode func(buffer) (Color, int)
	incr := adj == 7
	if incr {
		adj = 0
	}

	switch (opcode - 0x80) >> 3 {
	case 0:
		nBytes, directness, decode = 1, "", buffer.decodeColor1
	case 1:
		nBytes, directness, decode = 2, "", buffer.decodeColor2
	case 2:
		nBytes, directness, decode = 3, " (direct)", buffer.decodeColor3Direct
	case 3:
		nBytes, directness, decode = 4, "", buffer.decodeColor4
	case 4:
		nBytes, directness, decode = 3, " (indirect)", buffer.decodeColor3Indirect
	}
	if p != nil {
		if incr {
			p(src[:1], "Set CREG[CSEL-0] to a %d byte%s color; CSEL++\n", nBytes, directness)
		} else {
			p(src[:1], "Set CREG[CSEL-%d] to a %d byte%s color\n", adj, nBytes, directness)
		}
	}
	src = src[1:]

	c, n := decode(src)
	if n == 0 {
		return nil, nil, errInvalidColor
	}

	if p != nil {
		printColor(src[:n], p, c, "")
	}
	src = src[n:]

	if dst != nil {
		dst.SetCReg(adj, incr, c)
	}

	return decodeStyling, src, nil
}

func printColor(src []byte, p printer, c Color, prefix string) {
	switch c.typ {
	case ColorTypeRGBA:
		if rgba := c.rgba(); validAlphaPremulColor(rgba) {
			p(src, "    %sRGBA %02x%02x%02x%02x\n", prefix, rgba.R, rgba.G, rgba.B, rgba.A)
		} else if rgba.A == 0 && rgba.B&0x80 != 0 {
			p(src, "    %sgradient (NSTOPS=%d, CBASE=%d, NBASE=%d, %s, %s)\n",
				prefix,
				rgba.R&0x3f,
				rgba.G&0x3f,
				rgba.B&0x3f,
				gradientShapeNames[(rgba.B>>6)&0x01],
				gradientSpreadNames[rgba.G>>6],
			)
		} else {
			p(src, "    %snonsensical color\n", prefix)
		}
	case ColorTypePaletteIndex:
		p(src, "    %scustomPalette[%d]\n", prefix, c.paletteIndex())
	case ColorTypeCReg:
		p(src, "    %sCREG[%d]\n", prefix, c.cReg())
	case ColorTypeBlend:
		t, c0, c1 := c.blend()
		p(src[:1], "    blend %d:%d c0:c1\n", 0xff-t, t)
		printColor(src[1:2], p, decodeColor1(c0), "    c0: ")
		printColor(src[2:3], p, decodeColor1(c1), "    c1: ")
	}
}

func decodeSetNReg(dst Destination, p printer, src buffer, opcode byte) (modeFunc, buffer, error) {
	decode, typ, adj := buffer.decodeZeroToOne, "zero-to-one", opcode&0x07
	incr := adj == 7
	if incr {
		adj = 0
	}

	switch (opcode - 0xa8) >> 3 {
	case 0:
		decode, typ = buffer.decodeReal, "real"
	case 1:
		decode, typ = buffer.decodeCoordinate, "coordinate"
	}
	if p != nil {
		if incr {
			p(src[:1], "Set NREG[NSEL-0] to a %s number; NSEL++\n", typ)
		} else {
			p(src[:1], "Set NREG[NSEL-%d] to a %s number\n", adj, typ)
		}
	}
	src = src[1:]

	f, n := decode(src)
	if n == 0 {
		return nil, nil, errInvalidNumber
	}
	if p != nil {
		p(src[:n], "    %g\n", f)
	}
	src = src[n:]

	if dst != nil {
		dst.SetNReg(adj, incr, f)
	}

	return decodeStyling, src, nil
}

func decodeStartPath(dst Destination, p printer, src buffer, opcode byte) (modeFunc, buffer, error) {
	adj := opcode & 0x07
	if p != nil {
		p(src[:1], "Start path, filled with CREG[CSEL-%d]; M (absolute moveTo)\n", adj)
	}
	src = src[1:]

	x, src, err := decodeNumber(p, src, buffer.decodeCoordinate)
	if err != nil {
		return nil, nil, err
	}
	y, src, err := decodeNumber(p, src, buffer.decodeCoordinate)
	if err != nil {
		return nil, nil, err
	}

	if dst != nil {
		dst.StartPath(adj, x, y)
	}

	return decodeDrawing, src, nil
}

func decodeSetLOD(dst Destination, p printer, src buffer) (modeFunc, buffer, error) {
	if p != nil {
		p(src[:1], "Set LOD\n")
	}
	src = src[1:]

	lod0, src, err := decodeNumber(p, src, buffer.decodeReal)
	if err != nil {
		return nil, nil, err
	}
	lod1, src, err := decodeNumber(p, src, buffer.decodeReal)
	if err != nil {
		return nil, nil, err
	}

	if dst != nil {
		dst.SetLOD(lod0, lod1)
	}
	return decodeStyling, src, nil
}

func decodeDrawing(dst Destination, p printer, src buffer) (mf modeFunc, src1 buffer, err error) {
	var coords [6]float32

	switch opcode := src[0]; {
	case opcode < 0xe0:
		op, nCoords, nReps := "", 0, 1+int(opcode&0x0f)
		switch opcode >> 4 {
		case 0x00, 0x01:
			op = "L (absolute lineTo)"
			nCoords = 2
			nReps = 1 + int(opcode&0x1f)
		case 0x02, 0x03:
			op = "l (relative lineTo)"
			nCoords = 2
			nReps = 1 + int(opcode&0x1f)
		case 0x04:
			op = "T (absolute smooth quadTo)"
			nCoords = 2
		case 0x05:
			op = "t (relative smooth quadTo)"
			nCoords = 2
		case 0x06:
			op = "Q (absolute quadTo)"
			nCoords = 4
		case 0x07:
			op = "q (relative quadTo)"
			nCoords = 4
		case 0x08:
			op = "S (absolute smooth cubeTo)"
			nCoords = 4
		case 0x09:
			op = "s (relative smooth cubeTo)"
			nCoords = 4
		case 0x0a:
			op = "C (absolute cubeTo)"
			nCoords = 6
		case 0x0b:
			op = "c (relative cubeTo)"
			nCoords = 6
		case 0x0c:
			op = "A (absolute arcTo)"
			nCoords = 0
		case 0x0d:
			op = "a (relative arcTo)"
			nCoords = 0
		}

		if p != nil {
			p(src[:1], "%s, %d reps\n", op, nReps)
		}
		src = src[1:]

		for i := 0; i < nReps; i++ {
			if p != nil && i != 0 {
				p(nil, "%s, implicit\n", op)
			}
			var largeArc, sweep bool
			if op[0] != 'A' && op[0] != 'a' {
				src, err = decodeCoordinates(coords[:nCoords], p, src)
				if err != nil {
					return nil, nil, err
				}
			} else {
				// We have an absolute or relative arcTo.
				src, err = decodeCoordinates(coords[:2], p, src)
				if err != nil {
					return nil, nil, err
				}
				coords[2], src, err = decodeAngle(p, src)
				if err != nil {
					return nil, nil, err
				}
				largeArc, sweep, src, err = decodeArcToFlags(p, src)
				if err != nil {
					return nil, nil, err
				}
				src, err = decodeCoordinates(coords[4:6], p, src)
				if err != nil {
					return nil, nil, err
				}
			}

			if dst == nil {
				continue
			}
			switch op[0] {
			case 'L':
				dst.AbsLineTo(coords[0], coords[1])
			case 'l':
				dst.RelLineTo(coords[0], coords[1])
			case 'T':
				dst.AbsSmoothQuadTo(coords[0], coords[1])
			case 't':
				dst.RelSmoothQuadTo(coords[0], coords[1])
			case 'Q':
				dst.AbsQuadTo(coords[0], coords[1], coords[2], coords[3])
			case 'q':
				dst.RelQuadTo(coords[0], coords[1], coords[2], coords[3])
			case 'S':
				dst.AbsSmoothCubeTo(coords[0], coords[1], coords[2], coords[3])
			case 's':
				dst.RelSmoothCubeTo(coords[0], coords[1], coords[2], coords[3])
			case 'C':
				dst.AbsCubeTo(coords[0], coords[1], coords[2], coords[3], coords[4], coords[5])
			case 'c':
				dst.RelCubeTo(coords[0], coords[1], coords[2], coords[3], coords[4], coords[5])
			case 'A':
				dst.AbsArcTo(coords[0], coords[1], coords[2], largeArc, sweep, coords[4], coords[5])
			case 'a':
				dst.RelArcTo(coords[0], coords[1], coords[2], largeArc, sweep, coords[4], coords[5])
			}
		}

	case opcode == 0xe1:
		if p != nil {
			p(src[:1], "z (closePath); end path\n")
		}
		src = src[1:]
		if dst != nil {
			dst.ClosePathEndPath()
		}
		return decodeStyling, src, nil

	case opcode == 0xe2:
		if p != nil {
			p(src[:1], "z (closePath); M (absolute moveTo)\n")
		}
		src = src[1:]
		src, err = decodeCoordinates(coords[:2], p, src)
		if err != nil {
			return nil, nil, err
		}
		if dst != nil {
			dst.ClosePathAbsMoveTo(coords[0], coords[1])
		}

	case opcode == 0xe3:
		if p != nil {
			p(src[:1], "z (closePath); m (relative moveTo)\n")
		}
		src = src[1:]
		src, err = decodeCoordinates(coords[:2], p, src)
		if err != nil {
			return nil, nil, err
		}
		if dst != nil {
			dst.ClosePathRelMoveTo(coords[0], coords[1])
		}

	case opcode == 0xe6:
		if p != nil {
			p(src[:1], "H (absolute horizontal lineTo)\n")
		}
		src = src[1:]
		src, err = decodeCoordinates(coords[:1], p, src)
		if err != nil {
			return nil, nil, err
		}
		if dst != nil {
			dst.AbsHLineTo(coords[0])
		}

	case opcode == 0xe7:
		if p != nil {
			p(src[:1], "h (relative horizontal lineTo)\n")
		}
		src = src[1:]
		src, err = decodeCoordinates(coords[:1], p, src)
		if err != nil {
			return nil, nil, err
		}
		if dst != nil {
			dst.RelHLineTo(coords[0])
		}

	case opcode == 0xe8:
		if p != nil {
			p(src[:1], "V (absolute vertical lineTo)\n")
		}
		src = src[1:]
		src, err = decodeCoordinates(coords[:1], p, src)
		if err != nil {
			return nil, nil, err
		}
		if dst != nil {
			dst.AbsVLineTo(coords[0])
		}

	case opcode == 0xe9:
		if p != nil {
			p(src[:1], "v (relative vertical lineTo)\n")
		}
		src = src[1:]
		src, err = decodeCoordinates(coords[:1], p, src)
		if err != nil {
			return nil, nil, err
		}
		if dst != nil {
			dst.RelVLineTo(coords[0])
		}

	default:
		return nil, nil, errUnsupportedDrawingOpcode
	}
	return decodeDrawing, src, nil
}

type decodeNumberFunc func(buffer) (float32, int)

func decodeNumber(p printer, src buffer, dnf decodeNumberFunc) (float32, buffer, error) {
	x, n := dnf(src)
	if n == 0 {
		return 0, nil, errInvalidNumber
	}
	if p != nil {
		p(src[:n], "    %+g\n", x)
	}
	return x, src[n:], nil
}

func decodeCoordinates(coords []float32, p printer, src buffer) (src1 buffer, err error) {
	for i := range coords {
		coords[i], src, err = decodeNumber(p, src, buffer.decodeCoordinate)
		if err != nil {
			return nil, err
		}
	}
	return src, nil
}

func decodeAngle(p printer, src buffer) (float32, buffer, error) {
	x, n := src.decodeZeroToOne()
	if n == 0 {
		return 0, nil, errInvalidNumber
	}
	if p != nil {
		p(src[:n], "    %v × 360 degrees (%v degrees)\n", x, x*360)
	}
	return x, src[n:], nil
}

func decodeArcToFlags(p printer, src buffer) (bool, bool, buffer, error) {
	x, n := src.decodeNatural()
	if n == 0 {
		return false, false, nil, errInvalidNumber
	}
	if p != nil {
		p(src[:n], "    %#x (largeArc=%d, sweep=%d)\n", x, (x>>0)&0x01, (x>>1)&0x01)
	}
	return (x>>0)&0x01 != 0, (x>>1)&0x01 != 0, src[n:], nil
}
