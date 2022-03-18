// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iconvg

import (
	"bytes"
	"errors"
	"image/color"
	"math"
)

// UpgradeToFileFormatVersion1Options are the options to the
// UpgradeToFileFormatVersion1 function.
type UpgradeToFileFormatVersion1Options struct {
	// ArcsExpandWithHighResolutionCoordinates is like the
	// Encoder.HighResolutionCoordinates field. It controls whether to favor
	// file size (false) or precision (true) when replacing File Format Version
	// 0's arcs with cubic Bézier curves.
	ArcsExpandWithHighResolutionCoordinates bool
}

// UpgradeToFileFormatVersion1 upgrades IconVG data from the 2016 experimental
// "File Format Version 0" to the 2021 "File Format Version 1".
//
// This package (golang.org/x/exp/shiny/iconvg) holds a decoder for FFV0,
// including this function to convert from FFV0 to FFV1. Different packages
// (github.com/google/iconvg/src/go/*) decode FFV1.
//
// Amongst some new features and other clean-ups, FFV1 sets up the capability
// for animated vector graphics, therefore removing some FFV0 features (such as
// arc segments) that can be hard to animate smoothly. The IconvG FFV1 format
// and its design decisions are discussed at
// https://github.com/google/iconvg/issues/4#issuecomment-874105547
func UpgradeToFileFormatVersion1(v0 []byte, opts *UpgradeToFileFormatVersion1Options) (v1 []byte, retErr error) {
	u := &upgrader{}
	if opts != nil {
		u.opts = *opts
	}
	for i := range u.creg {
		u.creg[i] = upgradeColor{
			typ:          ColorTypePaletteIndex,
			paletteIndex: uint8(i),
		}
	}

	if !bytes.HasPrefix(v0, magicBytes) {
		return nil, errInvalidMagicIdentifier
	}
	v1 = append(v1, "\x8AIVG"...)
	v0 = v0[4:]

	v1, v0, retErr = u.upgradeMetadata(v1, v0)
	if retErr != nil {
		return nil, retErr
	}

	v1, _, retErr = u.upgradeBytecode(v1, v0)
	if retErr != nil {
		return nil, retErr
	}

	return v1, nil
}

const (
	upgradeVerbMoveTo = 0
	upgradeVerbLineTo = 1
	upgradeVerbQuadTo = 2
	upgradeVerbCubeTo = 3
)

type upgrader struct {
	opts UpgradeToFileFormatVersion1Options

	// These fields hold the current path's geometry.
	verbs []uint8
	args  [][2]float32

	// These fields track most of the FFV0 virtual machine register state. The
	// FFV1 register model is different enough that we don't just translate
	// each FFV0 register-related opcode individually.
	creg [64]upgradeColor
	nreg [64]float32
	csel uint32
	nsel uint32
	fill uint32

	// These fields track the most recent color written to FFV1 register
	// REGS[SEL+7] (and SEL is kept at 56). As a file size optimization, we
	// don't have to emit the first half of "Set REGS[SEL+7] = etc; Use
	// REGS[SEL+7]" if the register already holds the "etc" value.
	regsSel7    color.RGBA
	hasRegsSel7 bool

	// calculatingJumpLOD is whether the upgrader.upgradeBytecode method is
	// being called recursively. FFV0 sets a Level-Of-Detail filter that
	// applies implicitly until the next SetLOD opcode (if any). FFV1 instead
	// explicitly gives the number of opcodes to skip if outside the LOD range.
	calculatingJumpLOD bool
}

func (u *upgrader) upgradeMetadata(v1 buffer, v0 buffer) (newV1 buffer, newV0 buffer, retErr error) {
	nMetadataChunks, n := v0.decodeNatural()
	if n == 0 {
		return nil, nil, errInvalidNumberOfMetadataChunks
	}
	v1.encodeNaturalFFV1(nMetadataChunks)
	v0 = v0[n:]

	for ; nMetadataChunks > 0; nMetadataChunks-- {
		length, n := v0.decodeNatural()
		if n == 0 {
			return nil, nil, errInvalidMetadataChunkLength
		}
		v0 = v0[n:]
		if uint64(length) > uint64(len(v0)) {
			return nil, nil, errInvalidMetadataChunkLength
		}
		upgrade, err := u.upgradeMetadataChunk(v0[:length])
		if err != nil {
			return nil, nil, err
		}
		v1.encodeNaturalFFV1(uint32(len(upgrade)))
		v1 = append(v1, upgrade...)
		v0 = v0[length:]
	}
	return v1, v0, nil
}

func (u *upgrader) upgradeMetadataChunk(v0 buffer) (v1 buffer, retErr error) {
	mid, n := v0.decodeNatural()
	if n == 0 {
		return nil, errInvalidMetadataIdentifier
	}
	switch mid {
	case midViewBox:
		mid = ffv1MIDViewBox
	case midSuggestedPalette:
		mid = ffv1MIDSuggestedPalette
	default:
		return nil, errInvalidMetadataIdentifier
	}
	v1.encodeNaturalFFV1(mid)
	v0 = v0[n:]

	switch mid {
	case ffv1MIDViewBox:
		for i := 0; i < 4; i++ {
			x, n := v0.decodeNatural()
			if n == 0 {
				return nil, errInvalidViewBox
			}
			v1.encodeNaturalFFV1(x)
			v0 = v0[n:]
		}
		if len(v0) != 0 {
			return nil, errInvalidViewBox
		}

	case ffv1MIDSuggestedPalette:
		if len(v0) == 0 {
			return nil, errInvalidSuggestedPalette
		}
		numColors := 1 + int(v0[0]&0x3f)
		colorLength := 1 + int(v0[0]>>6)
		v1 = append(v1, uint8(numColors-1))
		v0 = v0[1:]
		for i := 0; i < numColors; i++ {
			c, n := Color{}, 0
			switch colorLength {
			case 1:
				c, n = v0.decodeColor1()
			case 2:
				c, n = v0.decodeColor2()
			case 3:
				c, n = v0.decodeColor3Direct()
			case 4:
				c, n = v0.decodeColor4()
			}
			if n == 0 {
				return nil, errInvalidSuggestedPalette
			} else if (c.typ == ColorTypeRGBA) && validAlphaPremulColor(c.data) {
				v1 = append(v1, c.data.R, c.data.G, c.data.B, c.data.A)
			} else {
				v1 = append(v1, 0x00, 0x00, 0x00, 0xff)
			}
			v0 = v0[n:]
		}
		if len(v0) != 0 {
			return nil, errInvalidSuggestedPalette
		}
	}
	return v1, nil
}

func (u *upgrader) upgradeBytecode(v1 buffer, v0 buffer) (newV1 buffer, newV0 buffer, retErr error) {
	uf := upgradeFunc(upgradeStyling)
	for len(v0) > 0 {
		uf, v1, v0, retErr = uf(u, v1, v0)
		if retErr != nil {
			if retErr == errCalculatingJumpLOD {
				return v1, v0, nil
			}
			return nil, nil, retErr
		}
	}
	return v1, v0, nil
}

var errCalculatingJumpLOD = errors.New("iconvg: calculating JumpLOD")

type upgradeFunc func(*upgrader, buffer, buffer) (upgradeFunc, buffer, buffer, error)

func upgradeStyling(u *upgrader, v1 buffer, v0 buffer) (uf upgradeFunc, newV1 buffer, newV0 buffer, retErr error) {
	for len(v0) > 0 {
		switch opcode := v0[0]; {
		case opcode < 0x80: // "Set CSEL/NSEL"
			if opcode < 0x40 {
				u.csel = uint32(opcode & 63)
			} else {
				u.nsel = uint32(opcode & 63)
			}
			v0 = v0[1:]

		case opcode < 0xa8: // "Set CREG[etc] to an etc color"
			adj := uint32(opcode & 7)
			if adj == 7 {
				adj = 0
			}
			index := (u.csel - adj) & 63

			v0 = v0[1:]
			c, n := Color{}, 0
			switch (opcode - 0x80) >> 3 {
			case 0:
				c, n = v0.decodeColor1()
			case 1:
				c, n = v0.decodeColor2()
			case 2:
				c, n = v0.decodeColor3Direct()
			case 3:
				c, n = v0.decodeColor4()
			case 4:
				c, n = v0.decodeColor3Indirect()
			}
			if n == 0 {
				return nil, nil, nil, errInvalidColor
			}
			u.creg[index], retErr = u.resolve(c, false)
			if retErr != nil {
				return nil, nil, nil, retErr
			}
			v0 = v0[n:]

			if (opcode & 7) == 7 {
				u.csel = (u.csel + 1) & 63
			}

		case opcode < 0xc0: // "Set NREG[etc] to a real number"
			adj := uint32(opcode & 7)
			if adj == 7 {
				adj = 0
			}
			index := (u.nsel - adj) & 63

			v0 = v0[1:]
			f, n := float32(0), 0
			switch (opcode - 0x80) >> 3 {
			case 5:
				f, n = v0.decodeReal()
			case 6:
				f, n = v0.decodeCoordinate()
			case 7:
				f, n = v0.decodeZeroToOne()
			}
			if n == 0 {
				return nil, nil, nil, errInvalidNumber
			}
			u.nreg[index] = f
			v0 = v0[n:]

			if (opcode & 7) == 7 {
				u.nsel = (u.nsel + 1) & 63
			}

		case opcode < 0xc7: // Start path.
			adj := uint32(opcode & 7)
			u.fill = (u.csel - adj) & 63
			v1 = append(v1, 0x35) // FFV1 MoveTo.
			v0 = v0[1:]
			return upgradeDrawing, v1, v0, nil

		case opcode == 0xc7: // "Set LOD"
			if u.calculatingJumpLOD {
				u.calculatingJumpLOD = false
				return nil, v1, v0, errCalculatingJumpLOD
			}

			v0 = v0[1:]
			lod := [2]float32{}
			for i := range lod {
				f, n := v0.decodeReal()
				if n == 0 {
					return nil, nil, nil, errInvalidNumber
				}
				lod[i] = f
				v0 = v0[n:]
			}
			if (lod[0] == 0) && math.IsInf(float64(lod[1]), +1) {
				break
			}

			u.calculatingJumpLOD = true
			ifTrue := []byte(nil)
			if ifTrue, v0, retErr = u.upgradeBytecode(nil, v0); retErr != nil {
				return nil, nil, nil, retErr
			}
			nInstructions := countFFV1Instructions(ifTrue)
			if nInstructions >= (1 << 30) {
				return nil, nil, nil, errUnsupportedUpgrade
			}
			v1 = append(v1, 0x3a) // FFV1 JumpLOD.
			v1.encodeNaturalFFV1(uint32(nInstructions))
			v1.encodeCoordinateFFV1(lod[0])
			v1.encodeCoordinateFFV1(lod[1])
			v1 = append(v1, ifTrue...)

		default:
			return nil, nil, nil, errUnsupportedStylingOpcode
		}
	}
	return upgradeStyling, v1, v0, nil
}

func upgradeDrawing(u *upgrader, v1 buffer, v0 buffer) (uf upgradeFunc, newV1 buffer, newV0 buffer, retErr error) {
	u.verbs = u.verbs[:0]
	u.args = u.args[:0]

	coords := [3][2]float32{}
	pen := [2]float32{}
	prevSmoothType := smoothTypeNone
	prevSmoothPoint := [2]float32{}

	// Handle the implicit M after a "Start path" styling op.
	v0, retErr = decodeCoordinates(pen[:2], nil, v0)
	if retErr != nil {
		return nil, nil, nil, retErr
	}
	u.verbs = append(u.verbs, upgradeVerbMoveTo)
	u.args = append(u.args, pen)
	startingPoint := pen

	for len(v0) > 0 {
		switch opcode := v0[0]; {
		case opcode < 0xc0: // LineTo, QuadTo, CubeTo.
			nCoordPairs, nReps, relative, smoothType := 0, 1+int(opcode&0x0f), false, smoothTypeNone
			switch opcode >> 4 {
			case 0x00, 0x01: // "L (absolute lineTo)"
				nCoordPairs = 1
				nReps = 1 + int(opcode&0x1f)
			case 0x02, 0x03: // "l (relative lineTo)"
				nCoordPairs = 1
				nReps = 1 + int(opcode&0x1f)
				relative = true
			case 0x04: // "T (absolute smooth quadTo)"
				nCoordPairs = 1
				smoothType = smoothTypeQuad
			case 0x05: // "t (relative smooth quadTo)"
				nCoordPairs = 1
				relative = true
				smoothType = smoothTypeQuad
			case 0x06: // "Q (absolute quadTo)"
				nCoordPairs = 2
			case 0x07: // "q (relative quadTo)"
				nCoordPairs = 2
				relative = true
			case 0x08: // "S (absolute smooth cubeTo)"
				nCoordPairs = 2
				smoothType = smoothTypeCube
			case 0x09: // "s (relative smooth cubeTo)"
				nCoordPairs = 2
				relative = true
				smoothType = smoothTypeCube
			case 0x0a: // "C (absolute cubeTo)"
				nCoordPairs = 3
			case 0x0b: // "c (relative cubeTo)"
				nCoordPairs = 3
				relative = true
			}
			v0 = v0[1:]

			for i := 0; i < nReps; i++ {
				smoothIndex := 0
				if smoothType != smoothTypeNone {
					smoothIndex = 1
					if smoothType != prevSmoothType {
						coords[0][0] = pen[0]
						coords[0][1] = pen[1]
					} else {
						coords[0][0] = (2 * pen[0]) - prevSmoothPoint[0]
						coords[0][1] = (2 * pen[1]) - prevSmoothPoint[1]
					}
				}
				allCoords := coords[:smoothIndex+nCoordPairs]
				explicitCoords := allCoords[smoothIndex:]

				v0, retErr = decodeCoordinatePairs(explicitCoords, nil, v0)
				if retErr != nil {
					return nil, nil, nil, retErr
				}
				if relative {
					for c := range explicitCoords {
						explicitCoords[c][0] += pen[0]
						explicitCoords[c][1] += pen[1]
					}
				}

				u.verbs = append(u.verbs, uint8(len(allCoords)))
				u.args = append(u.args, allCoords...)

				pen = allCoords[len(allCoords)-1]
				if len(allCoords) == 2 {
					prevSmoothPoint = allCoords[0]
					prevSmoothType = smoothTypeQuad
				} else if len(allCoords) == 3 {
					prevSmoothPoint = allCoords[1]
					prevSmoothType = smoothTypeCube
				} else {
					prevSmoothType = smoothTypeNone
				}
			}

		case opcode < 0xe0: // ArcTo.
			v1, v0, retErr = u.upgradeArcs(&pen, v1, v0)
			if retErr != nil {
				return nil, nil, nil, retErr
			}
			prevSmoothType = smoothTypeNone

		default: // Other drawing opcodes.
			v0 = v0[1:]
			switch opcode {
			case 0xe1: // "z (closePath); end path"
				goto endPath

			case 0xe2, 0xe3: // "z (closePath); M (absolute/relative moveTo)"
				v0, retErr = decodeCoordinatePairs(coords[:1], nil, v0)
				if retErr != nil {
					return nil, nil, nil, retErr
				}
				if opcode == 0xe2 {
					pen[0] = coords[0][0]
					pen[1] = coords[0][1]
				} else {
					pen[0] += coords[0][0]
					pen[1] += coords[0][1]
				}
				u.verbs = append(u.verbs, upgradeVerbMoveTo)
				u.args = append(u.args, pen)

			default:
				tmp := [1]float32{}
				v0, retErr = decodeCoordinates(tmp[:1], nil, v0)
				if retErr != nil {
					return nil, nil, nil, retErr
				}
				switch opcode {
				case 0xe6: // "H (absolute horizontal lineTo)"
					pen[0] = tmp[0]
				case 0xe7: // "h (relative horizontal lineTo)"
					pen[0] += tmp[0]
				case 0xe8: // "V (absolute vertical lineTo)"
					pen[1] = tmp[0]
				case 0xe9: // "v (relative vertical lineTo)"
					pen[1] += tmp[0]
				default:
					return nil, nil, nil, errUnsupportedDrawingOpcode
				}
				u.verbs = append(u.verbs, upgradeVerbLineTo)
				u.args = append(u.args, pen)
			}
			prevSmoothType = smoothTypeNone
		}
	}

endPath:
	v1, retErr = u.finishDrawing(v1, startingPoint)
	return upgradeStyling, v1, v0, retErr
}

func (u *upgrader) finishDrawing(v1 buffer, startingPoint [2]float32) (newV1 buffer, retErr error) {
	v1.encodeCoordinatePairFFV1(u.args[0])

	for i, j := 1, 1; i < len(u.verbs); {
		curr := u.args[j-1]
		runLength := u.computeRunLength(u.verbs[i:])
		verb := u.verbs[i]

		if verb == upgradeVerbMoveTo {
			v1 = append(v1, 0x35) // FFV1 MoveTo.
			v1.encodeCoordinatePairFFV1(u.args[j])
			i += 1
			j += 1
			continue
		}

		switch verb {
		case upgradeVerbLineTo:
			if ((runLength == 3) && ((j + 3) == len(u.args)) && u.looksLikeParallelogram3(&curr, u.args[j:], &startingPoint)) ||
				((runLength == 4) && u.looksLikeParallelogram4(&curr, u.args[j:j+4])) {
				v1 = append(v1, 0x34) // FFV1 Parallelogram.
				v1.encodeCoordinatePairFFV1(u.args[j+0])
				v1.encodeCoordinatePairFFV1(u.args[j+1])
				i += 4
				j += 4 * 1
				continue
			}
		case upgradeVerbCubeTo:
			if (runLength == 4) && u.looksLikeEllipse(&curr, u.args[j:j+(4*3)]) {
				v1 = append(v1, 0x33) // FFV1 Ellipse (4 quarters).
				v1.encodeCoordinatePairFFV1(u.args[j+2])
				v1.encodeCoordinatePairFFV1(u.args[j+5])
				i += 4
				j += 4 * 3
				continue
			}
		}

		opcodeBase := 0x10 * (verb - 1) // FFV1 LineTo / QuadTo / CubeTo.
		if runLength < 16 {
			v1 = append(v1, opcodeBase|uint8(runLength))
		} else {
			v1 = append(v1, opcodeBase)
			v1.encodeNaturalFFV1(uint32(runLength) - 16)
		}
		args := u.args[j : j+(runLength*int(verb))]
		for _, arg := range args {
			v1.encodeCoordinatePairFFV1(arg)
		}
		i += runLength
		j += len(args)
	}

	return u.emitFill(v1)
}

func (u *upgrader) emitFill(v1 buffer) (newV1 buffer, retErr error) {
	switch c := u.creg[u.fill]; c.typ {
	case ColorTypeRGBA:
		if validAlphaPremulColor(c.rgba) {
			if !u.hasRegsSel7 || (u.regsSel7 != c.rgba) {
				u.hasRegsSel7, u.regsSel7 = true, c.rgba
				v1 = append(v1, 0x57, // FFV1 Set REGS[SEL+7].hi32.
					c.rgba.R, c.rgba.G, c.rgba.B, c.rgba.A)
			}
			v1 = append(v1, 0x87) // FFV1 Fill (flat color) with REGS[SEL+7].

		} else if (c.rgba.A == 0) && (c.rgba.B&0x80 != 0) {
			nStops := int(c.rgba.R & 63)
			cBase := int(c.rgba.G & 63)
			nBase := int(c.rgba.B & 63)
			if nStops < 2 {
				return nil, errInvalidColor
			} else if nStops > 17 {
				return nil, errUnsupportedUpgrade
			}

			v1 = append(v1, 0x70|uint8(nStops-2)) // FFV1 SEL -= N; Set REGS[SEL+1 .. SEL+1+N].
			for i := 0; i < nStops; i++ {
				if stopOffset := u.nreg[(nBase+i)&63]; stopOffset <= 0 {
					v1 = append(v1, 0x00, 0x00, 0x00, 0x00)
				} else if stopOffset < 1 {
					u := uint32(stopOffset * 0x10000)
					v1 = append(v1, uint8(u>>0), uint8(u>>8), uint8(u>>16), uint8(u>>24))
				} else {
					v1 = append(v1, 0x00, 0x00, 0x01, 0x00)
				}

				if stopColor := u.creg[(cBase+i)&63]; stopColor.typ != ColorTypeRGBA {
					return nil, errUnsupportedUpgrade
				} else {
					v1 = append(v1,
						stopColor.rgba.R,
						stopColor.rgba.G,
						stopColor.rgba.B,
						stopColor.rgba.A,
					)
				}
			}

			nMatrixElements := 0
			if c.rgba.B&0x40 == 0 {
				v1 = append(v1, 0x91, // FFV1 Fill (linear gradient) with REGS[SEL+1 .. SEL+1+N].
					(c.rgba.G&0xc0)|uint8(nStops-2))
				nMatrixElements = 3
			} else {
				v1 = append(v1, 0xa1, // FFV1 Fill (radial gradient) with REGS[SEL+1 .. SEL+1+N].
					(c.rgba.G&0xc0)|uint8(nStops-2))
				nMatrixElements = 6
			}
			for i := 0; i < nMatrixElements; i++ {
				u := math.Float32bits(u.nreg[(nBase+i-6)&63])
				v1 = append(v1, uint8(u>>0), uint8(u>>8), uint8(u>>16), uint8(u>>24))
			}

			v1 = append(v1, 0x36, // FFV1 SEL += N.
				uint8(nStops))
		} else {
			return nil, errInvalidColor
		}

	case ColorTypePaletteIndex:
		if c.paletteIndex < 7 {
			v1 = append(v1, 0x88+c.paletteIndex) // FFV1 Fill (flat color) with REGS[SEL+8+N].
		} else {
			v1 = append(v1, 0x56, // FFV1 Set REGS[SEL+6].hi32.
				0x80|c.paletteIndex, 0, 0, 0,
				0x86) // FFV1 Fill (flat color) with REGS[SEL+6].
		}

	case ColorTypeBlend:
		if c.color0.typ == ColorTypeRGBA {
			v1 = append(v1, 0x53, // FFV1 Set REGS[SEL+3].hi32.
				c.color0.rgba.R, c.color0.rgba.G, c.color0.rgba.B, c.color0.rgba.A)
		}
		if c.color1.typ == ColorTypeRGBA {
			v1 = append(v1, 0x54, // FFV1 Set REGS[SEL+4].hi32.
				c.color1.rgba.R, c.color1.rgba.G, c.color1.rgba.B, c.color1.rgba.A)
		}
		v1 = append(v1, 0x55, // FFV1 Set REGS[SEL+5].hi32.
			c.blend)
		if c.color0.typ == ColorTypeRGBA {
			v1 = append(v1, 0xfe)
		} else {
			v1 = append(v1, 0x80|c.color0.paletteIndex)
		}
		if c.color1.typ == ColorTypeRGBA {
			v1 = append(v1, 0xff)
		} else {
			v1 = append(v1, 0x80|c.color1.paletteIndex)
		}
		v1 = append(v1, 0, 0x85) // FFV1 Fill (flat color) with REGS[SEL+5].
	}

	return v1, nil
}

func (u *upgrader) computeRunLength(verbs []uint8) int {
	firstVerb := verbs[0]
	if firstVerb == 0 {
		return 1
	}
	n := 1
	for ; (n < len(verbs)) && (verbs[n] == firstVerb); n++ {
	}
	return n
}

// looksLikeParallelogram3 is like looksLikeParallelogram4 but the final point
// (implied by the ClosePath op) is separate from the middle 3 args.
func (u *upgrader) looksLikeParallelogram3(curr *[2]float32, args [][2]float32, final *[2]float32) bool {
	if len(args) != 3 {
		panic("unreachable")
	}
	return (*curr == *final) &&
		(curr[0] == (args[0][0] - args[1][0] + args[2][0])) &&
		(curr[1] == (args[0][1] - args[1][1] + args[2][1]))
}

// looksLikeParallelogram4 returns whether the 5 coordinate pairs (A, B, C, D,
// E) form a parallelogram:
//
// E=A           B
//    o---------o
//     \         \
//      \         \
//       \         \
//        o---------o
//       D           C
//
// Specifically, it checks that (A == E) and ((A - B) == (D - C)). That last
// equation can be rearranged as (A == (B - C + D)).
//
// The motivation is that, if looksLikeParallelogram4 is true, then the 5 input
// coordinate pairs can then be compressed to 3: A, B and C. Or, if the current
// point A is implied by context then 4 input pairs can be compressed to 2.
func (u *upgrader) looksLikeParallelogram4(curr *[2]float32, args [][2]float32) bool {
	if len(args) != 4 {
		panic("unreachable")
	}
	return (*curr == args[3]) &&
		(curr[0] == (args[0][0] - args[1][0] + args[2][0])) &&
		(curr[1] == (args[0][1] - args[1][1] + args[2][1]))
}

// looksLikeEllipse returns whether the 13 coordinate pairs (A, A+, B-, B, B+,
// C- C, C+, D-, D, D+, A-, E) form a cubic Bézier approximation to an ellipse.
// Let A± denote the two tangent vectors (A+ - A) and (A - A-) and likewise for
// B±, C± and D±.
//
//     A+     B-
// E=A  o    o   B
// A- o---------o   B+
//  o  \         \ o
//      \    X    \
//     o \         \  o
//    D+  o---------o  C-
//       D   o    o  C
//          D-     C+
//
// See https://nigeltao.github.io/blog/2021/three-points-define-ellipse.html
// for a better version of that ASCII art.
//
// Specifically, it checks that (A, B, C, D, E), also known as (*curr, args[2],
// args[5], args[8] and args[11]), forms a parallelogram. If so, let X be the
// parallelogram center and define two axis vectors: r = B-X and s = C-X.
//
// These axes define the parallelogram's or ellipse's shape but they are not
// necessarily orthogonal and hence not necessarily the ellipse's major
// (longest) and minor (shortest) axes. If s is a 90 degree rotation of r then
// the parallelogram is a square and the ellipse is a circle.
//
// This function further checks that the A±, B± C± and D± tangents are
// approximately equal to +λ×r, +λ×s, -λ×r and -λ×s, where λ = ((math.Sqrt2 -
// 1) × 4 / 3) comes from the cubic Bézier approximation to a quarter-circle.
//
// The motivation is that, if looksLikeEllipse is true, then the 13 input
// coordinate pairs can then be compressed to 3: A, B and C. Or, if the current
// point A is implied by context then 12 input pairs can be compressed to 2.
func (u *upgrader) looksLikeEllipse(curr *[2]float32, args [][2]float32) bool {
	if len(args) != 12 {
		panic("unreachable")
	}
	if (*curr != args[11]) ||
		(curr[0] != (args[2][0] - args[5][0] + args[8][0])) ||
		(curr[1] != (args[2][1] - args[5][1] + args[8][1])) {
		return false
	}
	center := [2]float32{
		(args[2][0] + args[8][0]) / 2,
		(args[2][1] + args[8][1]) / 2,
	}

	// 0.5522847498307933984022516322796 ≈ ((math.Sqrt2 - 1) × 4 / 3), the
	// tangent lengths (as a fraction of the radius) for a commonly used cubic
	// Bézier approximation to a circle. Multiplying that by 0.98 and 1.02
	// checks that we're within 2% of that fraction.
	//
	// This also covers the slightly different 0.551784777779014 constant,
	// recommended by https://pomax.github.io/bezierinfo/#circles_cubic
	const λMin = 0.98 * 0.5522847498307933984022516322796
	const λMax = 1.02 * 0.5522847498307933984022516322796

	// Check the first axis.
	r := [2]float32{
		args[2][0] - center[0],
		args[2][1] - center[1],
	}
	rMin := [2]float32{r[0] * λMin, r[1] * λMin}
	rMax := [2]float32{r[0] * λMax, r[1] * λMax}
	if rMin[0] > rMax[0] {
		rMin[0], rMax[0] = rMax[0], rMin[0]
	}
	if rMin[1] > rMax[1] {
		rMin[1], rMax[1] = rMax[1], rMin[1]
	}
	if !within(args[0][0]-curr[0], args[0][1]-curr[1], rMin, rMax) ||
		!within(args[4][0]-args[5][0], args[4][1]-args[5][1], rMin, rMax) ||
		!within(args[5][0]-args[6][0], args[5][1]-args[6][1], rMin, rMax) ||
		!within(args[11][0]-args[10][0], args[11][1]-args[10][1], rMin, rMax) {
		return false
	}

	// Check the second axis.
	s := [2]float32{
		args[5][0] - center[0],
		args[5][1] - center[1],
	}
	sMin := [2]float32{s[0] * λMin, s[1] * λMin}
	sMax := [2]float32{s[0] * λMax, s[1] * λMax}
	if sMin[0] > sMax[0] {
		sMin[0], sMax[0] = sMax[0], sMin[0]
	}
	if sMin[1] > sMax[1] {
		sMin[1], sMax[1] = sMax[1], sMin[1]
	}
	if !within(args[2][0]-args[1][0], args[2][1]-args[1][1], sMin, sMax) ||
		!within(args[3][0]-args[2][0], args[3][1]-args[2][1], sMin, sMax) ||
		!within(args[7][0]-args[8][0], args[7][1]-args[8][1], sMin, sMax) ||
		!within(args[8][0]-args[9][0], args[8][1]-args[9][1], sMin, sMax) {
		return false
	}

	return true
}

func within(v0 float32, v1 float32, min [2]float32, max [2]float32) bool {
	return (min[0] <= v0) && (v0 <= max[0]) && (min[1] <= v1) && (v1 <= max[1])
}

func (u *upgrader) upgradeArcs(pen *[2]float32, v1 buffer, v0 buffer) (newV1 buffer, newV0 buffer, retErr error) {
	coords := [6]float32{}
	largeArc, sweep := false, false
	opcode := v0[0]
	v0 = v0[1:]
	nReps := 1 + int(opcode&0x0f)
	for i := 0; i < nReps; i++ {
		v0, retErr = decodeCoordinates(coords[:2], nil, v0)
		if retErr != nil {
			return nil, nil, retErr
		}
		coords[2], v0, retErr = decodeAngle(nil, v0)
		if retErr != nil {
			return nil, nil, retErr
		}
		largeArc, sweep, v0, retErr = decodeArcToFlags(nil, v0)
		if retErr != nil {
			return nil, nil, retErr
		}
		v0, retErr = decodeCoordinates(coords[4:6], nil, v0)
		if retErr != nil {
			return nil, nil, retErr
		}
		if (opcode >> 4) == 0x0d {
			coords[4] += pen[0]
			coords[5] += pen[1]
		}
		u.upgradeArc(pen, coords[0], coords[1], coords[2], largeArc, sweep, coords[4], coords[5])
		pen[0] = coords[4]
		pen[1] = coords[5]
	}
	return v1, v0, nil
}

func (u *upgrader) upgradeArc(pen *[2]float32, rx, ry, xAxisRotation float32, largeArc, sweep bool, finalX, finalY float32) {
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
		u.verbs = append(u.verbs, upgradeVerbLineTo)
		u.args = append(u.args, [2]float32{finalX, finalY})
		return
	}

	x1 := float64(pen[0])
	y1 := float64(pen[1])
	x2 := float64(finalX)
	y2 := float64(finalY)

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
		u.arcSegmentTo(cx, cy,
			theta1+deltaTheta*float64(i+0)/float64(n),
			theta1+deltaTheta*float64(i+1)/float64(n),
			Rx, Ry, cosPhi, sinPhi,
		)
	}
}

// arcSegmentTo approximates an arc by a cubic Bézier curve. The mathematical
// formulae for the control points are the same as that used by librsvg.
func (u *upgrader) arcSegmentTo(cx, cy, theta1, theta2, rx, ry, cosPhi, sinPhi float64) {
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
	highResolutionCoordinates := u.opts.ArcsExpandWithHighResolutionCoordinates
	u.verbs = append(u.verbs, upgradeVerbCubeTo)
	u.args = append(u.args,
		[2]float32{
			quantize(float32(cx+cosPhi*x1-sinPhi*y1), highResolutionCoordinates),
			quantize(float32(cy+sinPhi*x1+cosPhi*y1), highResolutionCoordinates),
		},
		[2]float32{
			quantize(float32(cx+cosPhi*x2-sinPhi*y2), highResolutionCoordinates),
			quantize(float32(cy+sinPhi*x2+cosPhi*y2), highResolutionCoordinates),
		},
		[2]float32{
			quantize(float32(cx+cosPhi*x3-sinPhi*y3), highResolutionCoordinates),
			quantize(float32(cy+sinPhi*x3+cosPhi*y3), highResolutionCoordinates),
		},
	)
}

func countFFV1Instructions(src buffer) (ret uint64) {
	for len(src) > 0 {
		ret++
		opcode := src[0]
		src = src[1:]

		switch {
		case opcode < 0x40:
			switch {
			case opcode < 0x30:
				nReps := uint32(opcode & 15)
				if nReps == 0 {
					n := 0
					nReps, n = src.decodeNaturalFFV1()
					src = src[n:]
					nReps += 16
				}
				nCoords := 2 * (1 + int(opcode>>4))
				for ; nReps > 0; nReps-- {
					for i := 0; i < nCoords; i++ {
						_, n := src.decodeNaturalFFV1()
						src = src[n:]
					}
				}
			case opcode < 0x35:
				for i := 0; i < 4; i++ {
					_, n := src.decodeNaturalFFV1()
					src = src[n:]
				}
			case opcode == 0x35:
				for i := 0; i < 2; i++ {
					_, n := src.decodeNaturalFFV1()
					src = src[n:]
				}
			case opcode == 0x36:
				src = src[1:]
			case opcode == 0x37:
				// No-op.
			default:
				// upgradeBytecode (with calculatingJumpLOD set) will not emit
				// jump or call instructions.
				panic("unexpected FFV1 instruction")
			}

		case opcode < 0x80:
			switch (opcode >> 4) & 3 {
			case 0, 1:
				src = src[4:]
			case 2:
				src = src[8:]
			default:
				src = src[8*(2+int(opcode&15)):]
			}

		case opcode < 0xc0:
			switch (opcode >> 4) & 3 {
			case 0:
				// No-op.
			case 1:
				src = src[13:]
			case 2:
				src = src[25:]
			default:
				// upgradeBytecode (with calculatingJumpLOD set) will not emit
				// reserved instructions.
				panic("unexpected FFV1 instruction")
			}

		default:
			// upgradeBytecode (with calculatingJumpLOD set) will not emit
			// reserved instructions.
			panic("unexpected FFV1 instruction")
		}
	}
	return ret
}

type upgradeColor struct {
	typ          ColorType
	paletteIndex uint8
	blend        uint8
	rgba         color.RGBA
	color0       *upgradeColor
	color1       *upgradeColor
}

func (u *upgrader) resolve(c Color, denyBlend bool) (upgradeColor, error) {
	switch c.typ {
	case ColorTypeRGBA:
		return upgradeColor{
			typ:  ColorTypeRGBA,
			rgba: c.data,
		}, nil
	case ColorTypePaletteIndex:
		return upgradeColor{
			typ:          ColorTypePaletteIndex,
			paletteIndex: c.paletteIndex(),
		}, nil
	case ColorTypeCReg:
		upgrade := u.creg[c.cReg()]
		if denyBlend && (upgrade.typ == ColorTypeBlend) {
			return upgradeColor{}, errUnsupportedUpgrade
		}
		return upgrade, nil
	}

	if denyBlend {
		return upgradeColor{}, errUnsupportedUpgrade
	}
	t, c0, c1 := c.blend()
	color0, err := u.resolve(decodeColor1(c0), true)
	if err != nil {
		return upgradeColor{}, err
	}
	color1, err := u.resolve(decodeColor1(c1), true)
	if err != nil {
		return upgradeColor{}, err
	}
	return upgradeColor{
		typ:    ColorTypeBlend,
		blend:  t,
		color0: &color0,
		color1: &color1,
	}, nil
}
