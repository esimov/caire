// SPDX-License-Identifier: Unlicense OR MIT

package paint

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"math"

	"gioui.org/f32"
	"gioui.org/internal/ops"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// ImageOp sets the brush to an image.
type ImageOp struct {
	uniform bool
	color   color.NRGBA
	src     *image.RGBA

	// handle is a key to uniquely identify this ImageOp
	// in a map of cached textures.
	handle interface{}
}

// ColorOp sets the brush to a constant color.
type ColorOp struct {
	Color color.NRGBA
}

// LinearGradientOp sets the brush to a gradient starting at stop1 with color1 and
// ending at stop2 with color2.
type LinearGradientOp struct {
	Stop1  f32.Point
	Color1 color.NRGBA
	Stop2  f32.Point
	Color2 color.NRGBA
}

// PaintOp fills the current clip area with the current brush.
type PaintOp struct {
}

// NewImageOp creates an ImageOp backed by src.
//
// NewImageOp assumes the backing image is immutable, and may cache a
// copy of its contents in a GPU-friendly way. Create new ImageOps to
// ensure that changes to an image is reflected in the display of
// it.
func NewImageOp(src image.Image) ImageOp {
	switch src := src.(type) {
	case *image.Uniform:
		col := color.NRGBAModel.Convert(src.C).(color.NRGBA)
		return ImageOp{
			uniform: true,
			color:   col,
		}
	case *image.RGBA:
		return ImageOp{
			src:    src,
			handle: new(int),
		}
	}

	sz := src.Bounds().Size()
	// Copy the image into a GPU friendly format.
	dst := image.NewRGBA(image.Rectangle{
		Max: sz,
	})
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	return ImageOp{
		src:    dst,
		handle: new(int),
	}
}

func (i ImageOp) Size() image.Point {
	if i.src == nil {
		return image.Point{}
	}
	return i.src.Bounds().Size()
}

func (i ImageOp) Add(o *op.Ops) {
	if i.uniform {
		ColorOp{
			Color: i.color,
		}.Add(o)
		return
	} else if i.src == nil || i.src.Bounds().Empty() {
		return
	}
	data := ops.Write2(&o.Internal, ops.TypeImageLen, i.src, i.handle)
	data[0] = byte(ops.TypeImage)
}

func (c ColorOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeColorLen)
	data[0] = byte(ops.TypeColor)
	data[1] = c.Color.R
	data[2] = c.Color.G
	data[3] = c.Color.B
	data[4] = c.Color.A
}

func (c LinearGradientOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeLinearGradientLen)
	data[0] = byte(ops.TypeLinearGradient)

	bo := binary.LittleEndian
	bo.PutUint32(data[1:], math.Float32bits(c.Stop1.X))
	bo.PutUint32(data[5:], math.Float32bits(c.Stop1.Y))
	bo.PutUint32(data[9:], math.Float32bits(c.Stop2.X))
	bo.PutUint32(data[13:], math.Float32bits(c.Stop2.Y))

	data[17+0] = c.Color1.R
	data[17+1] = c.Color1.G
	data[17+2] = c.Color1.B
	data[17+3] = c.Color1.A
	data[21+0] = c.Color2.R
	data[21+1] = c.Color2.G
	data[21+2] = c.Color2.B
	data[21+3] = c.Color2.A
}

func (d PaintOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypePaintLen)
	data[0] = byte(ops.TypePaint)
}

// FillShape fills the clip shape with a color.
func FillShape(ops *op.Ops, c color.NRGBA, shape clip.Op) {
	defer shape.Push(ops).Pop()
	Fill(ops, c)
}

// Fill paints an infinitely large plane with the provided color. It
// is intended to be used with a clip.Op already in place to limit
// the painted area. Use FillShape unless you need to paint several
// times within the same clip.Op.
func Fill(ops *op.Ops, c color.NRGBA) {
	ColorOp{Color: c}.Add(ops)
	PaintOp{}.Add(ops)
}
