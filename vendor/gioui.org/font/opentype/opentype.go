// SPDX-License-Identifier: Unlicense OR MIT

// Package opentype implements text layout and shaping for OpenType
// files.
package opentype

import (
	"bytes"
	"fmt"
	"image"
	"io"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"gioui.org/f32"
	"gioui.org/font/opentype/internal"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
)

// Font implements the text.Shaper interface using a rich text
// shaping engine.
type Font struct {
	font *truetype.Font
}

// Parse constructs a Font from source bytes.
func Parse(src []byte) (*Font, error) {
	face, err := truetype.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("failed parsing truetype font: %w", err)
	}
	return &Font{
		font: face,
	}, nil
}

func (f *Font) Layout(ppem fixed.Int26_6, maxWidth int, lc system.Locale, txt io.RuneReader) ([]text.Line, error) {
	return internal.Document(shaping.Shape, f.font, ppem, maxWidth, lc, txt), nil
}

func (f *Font) Shape(ppem fixed.Int26_6, str text.Layout) clip.PathSpec {
	return textPath(ppem, f, str)
}

func (f *Font) Metrics(ppem fixed.Int26_6) font.Metrics {
	metrics := font.Metrics{}
	font := harfbuzz.NewFont(f.font)
	font.XScale = int32(ppem.Ceil()) << 6
	font.YScale = font.XScale
	// Use any horizontal direction.
	fontExtents := font.ExtentsForDirection(harfbuzz.LeftToRight)
	ascender := fixed.I(int(fontExtents.Ascender * 64))
	descender := fixed.I(int(fontExtents.Descender * 64))
	gap := fixed.I(int(fontExtents.LineGap * 64))
	metrics.Height = ascender + descender + gap
	metrics.Ascent = ascender
	metrics.Descent = descender
	// These three are not readily available.
	// TODO(whereswaldon): figure out how to get these values.
	metrics.XHeight = ascender
	metrics.CapHeight = ascender
	metrics.CaretSlope = image.Pt(0, 1)

	return metrics
}

func textPath(ppem fixed.Int26_6, font *Font, str text.Layout) clip.PathSpec {
	var lastPos f32.Point
	var builder clip.Path
	ops := new(op.Ops)
	var x fixed.Int26_6
	builder.Begin(ops)
	rune := 0
	ppemInt := ppem.Round()
	ppem16 := uint16(ppemInt)
	scaleFactor := float32(ppemInt) / float32(font.font.Upem())
	for _, g := range str.Glyphs {
		advance := g.XAdvance
		outline, ok := font.font.GlyphData(g.ID, ppem16, ppem16).(fonts.GlyphOutline)
		if !ok {
			continue
		}
		// Move to glyph position.
		pos := f32.Point{
			X: float32(x)/64 - float32(g.XOffset)/64,
			Y: -float32(g.YOffset) / 64,
		}
		builder.Move(pos.Sub(lastPos))
		lastPos = pos
		var lastArg f32.Point

		// Convert sfnt.Segments to relative segments.
		for _, fseg := range outline.Segments {
			nargs := 1
			switch fseg.Op {
			case fonts.SegmentOpQuadTo:
				nargs = 2
			case fonts.SegmentOpCubeTo:
				nargs = 3
			}
			var args [3]f32.Point
			for i := 0; i < nargs; i++ {
				a := f32.Point{
					X: fseg.Args[i].X * scaleFactor,
					Y: -fseg.Args[i].Y * scaleFactor,
				}
				args[i] = a.Sub(lastArg)
				if i == nargs-1 {
					lastArg = a
				}
			}
			switch fseg.Op {
			case fonts.SegmentOpMoveTo:
				builder.Move(args[0])
			case fonts.SegmentOpLineTo:
				builder.Line(args[0])
			case fonts.SegmentOpQuadTo:
				builder.Quad(args[0], args[1])
			case fonts.SegmentOpCubeTo:
				builder.Cube(args[0], args[1], args[2])
			default:
				panic("unsupported segment op")
			}
		}
		lastPos = lastPos.Add(lastArg)
		x += advance
		rune++
	}
	return builder.End()
}
