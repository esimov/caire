// SPDX-License-Identifier: Unlicense OR MIT

// Package scene encodes and decodes graphics commands in the format used by the
// compute renderer.
package scene

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"unsafe"

	"gioui.org/internal/f32"
)

type Op uint32

type Command [sceneElemSize / 4]uint32

// GPU commands from piet/scene.h in package gioui.org/shaders.
const (
	OpNop Op = iota
	OpLine
	OpQuad
	OpCubic
	OpFillColor
	OpLineWidth
	OpTransform
	OpBeginClip
	OpEndClip
	OpFillImage
	OpSetFillMode
	OpGap
)

// FillModes, from setup.h.
type FillMode uint32

const (
	FillModeNonzero = 0
	FillModeStroke  = 1
)

const CommandSize = int(unsafe.Sizeof(Command{}))

const sceneElemSize = 36

func (c Command) Op() Op {
	return Op(c[0])
}

func (c Command) String() string {
	switch Op(c[0]) {
	case OpNop:
		return "nop"
	case OpLine:
		from, to := DecodeLine(c)
		return fmt.Sprintf("line(%v, %v)", from, to)
	case OpGap:
		from, to := DecodeLine(c)
		return fmt.Sprintf("gap(%v, %v)", from, to)
	case OpQuad:
		from, ctrl, to := DecodeQuad(c)
		return fmt.Sprintf("quad(%v, %v, %v)", from, ctrl, to)
	case OpCubic:
		from, ctrl0, ctrl1, to := DecodeCubic(c)
		return fmt.Sprintf("cubic(%v, %v, %v, %v)", from, ctrl0, ctrl1, to)
	case OpFillColor:
		return fmt.Sprintf("fillcolor %#.8x", c[1])
	case OpLineWidth:
		return "linewidth"
	case OpTransform:
		t := f32.NewAffine2D(
			math.Float32frombits(c[1]),
			math.Float32frombits(c[3]),
			math.Float32frombits(c[5]),
			math.Float32frombits(c[2]),
			math.Float32frombits(c[4]),
			math.Float32frombits(c[6]),
		)
		return fmt.Sprintf("transform (%v)", t)
	case OpBeginClip:
		bounds := f32.Rectangle{
			Min: f32.Pt(math.Float32frombits(c[1]), math.Float32frombits(c[2])),
			Max: f32.Pt(math.Float32frombits(c[3]), math.Float32frombits(c[4])),
		}
		return fmt.Sprintf("beginclip (%v)", bounds)
	case OpEndClip:
		bounds := f32.Rectangle{
			Min: f32.Pt(math.Float32frombits(c[1]), math.Float32frombits(c[2])),
			Max: f32.Pt(math.Float32frombits(c[3]), math.Float32frombits(c[4])),
		}
		return fmt.Sprintf("endclip (%v)", bounds)
	case OpFillImage:
		return "fillimage"
	case OpSetFillMode:
		return "setfillmode"
	default:
		panic("unreachable")
	}
}

func Line(start, end f32.Point) Command {
	return Command{
		0: uint32(OpLine),
		1: math.Float32bits(start.X),
		2: math.Float32bits(start.Y),
		3: math.Float32bits(end.X),
		4: math.Float32bits(end.Y),
	}
}

func Gap(start, end f32.Point) Command {
	return Command{
		0: uint32(OpGap),
		1: math.Float32bits(start.X),
		2: math.Float32bits(start.Y),
		3: math.Float32bits(end.X),
		4: math.Float32bits(end.Y),
	}
}

func Cubic(start, ctrl0, ctrl1, end f32.Point) Command {
	return Command{
		0: uint32(OpCubic),
		1: math.Float32bits(start.X),
		2: math.Float32bits(start.Y),
		3: math.Float32bits(ctrl0.X),
		4: math.Float32bits(ctrl0.Y),
		5: math.Float32bits(ctrl1.X),
		6: math.Float32bits(ctrl1.Y),
		7: math.Float32bits(end.X),
		8: math.Float32bits(end.Y),
	}
}

func Quad(start, ctrl, end f32.Point) Command {
	return Command{
		0: uint32(OpQuad),
		1: math.Float32bits(start.X),
		2: math.Float32bits(start.Y),
		3: math.Float32bits(ctrl.X),
		4: math.Float32bits(ctrl.Y),
		5: math.Float32bits(end.X),
		6: math.Float32bits(end.Y),
	}
}

func Transform(m f32.Affine2D) Command {
	sx, hx, ox, hy, sy, oy := m.Elems()
	return Command{
		0: uint32(OpTransform),
		1: math.Float32bits(sx),
		2: math.Float32bits(hy),
		3: math.Float32bits(hx),
		4: math.Float32bits(sy),
		5: math.Float32bits(ox),
		6: math.Float32bits(oy),
	}
}

func SetLineWidth(width float32) Command {
	return Command{
		0: uint32(OpLineWidth),
		1: math.Float32bits(width),
	}
}

func BeginClip(bbox f32.Rectangle) Command {
	return Command{
		0: uint32(OpBeginClip),
		1: math.Float32bits(bbox.Min.X),
		2: math.Float32bits(bbox.Min.Y),
		3: math.Float32bits(bbox.Max.X),
		4: math.Float32bits(bbox.Max.Y),
	}
}

func EndClip(bbox f32.Rectangle) Command {
	return Command{
		0: uint32(OpEndClip),
		1: math.Float32bits(bbox.Min.X),
		2: math.Float32bits(bbox.Min.Y),
		3: math.Float32bits(bbox.Max.X),
		4: math.Float32bits(bbox.Max.Y),
	}
}

func FillColor(col color.RGBA) Command {
	return Command{
		0: uint32(OpFillColor),
		1: uint32(col.R)<<24 | uint32(col.G)<<16 | uint32(col.B)<<8 | uint32(col.A),
	}
}

func FillImage(index int, offset image.Point) Command {
	x := int16(offset.X)
	y := int16(offset.Y)
	return Command{
		0: uint32(OpFillImage),
		1: uint32(index),
		2: uint32(uint16(x)) | uint32(uint16(y))<<16,
	}
}

func SetFillMode(mode FillMode) Command {
	return Command{
		0: uint32(OpSetFillMode),
		1: uint32(mode),
	}
}

func DecodeLine(cmd Command) (from, to f32.Point) {
	if cmd[0] != uint32(OpLine) {
		panic("invalid command")
	}
	from = f32.Pt(math.Float32frombits(cmd[1]), math.Float32frombits(cmd[2]))
	to = f32.Pt(math.Float32frombits(cmd[3]), math.Float32frombits(cmd[4]))
	return
}

func DecodeGap(cmd Command) (from, to f32.Point) {
	if cmd[0] != uint32(OpGap) {
		panic("invalid command")
	}
	from = f32.Pt(math.Float32frombits(cmd[1]), math.Float32frombits(cmd[2]))
	to = f32.Pt(math.Float32frombits(cmd[3]), math.Float32frombits(cmd[4]))
	return
}

func DecodeQuad(cmd Command) (from, ctrl, to f32.Point) {
	if cmd[0] != uint32(OpQuad) {
		panic("invalid command")
	}
	from = f32.Pt(math.Float32frombits(cmd[1]), math.Float32frombits(cmd[2]))
	ctrl = f32.Pt(math.Float32frombits(cmd[3]), math.Float32frombits(cmd[4]))
	to = f32.Pt(math.Float32frombits(cmd[5]), math.Float32frombits(cmd[6]))
	return
}

func DecodeCubic(cmd Command) (from, ctrl0, ctrl1, to f32.Point) {
	if cmd[0] != uint32(OpCubic) {
		panic("invalid command")
	}
	from = f32.Pt(math.Float32frombits(cmd[1]), math.Float32frombits(cmd[2]))
	ctrl0 = f32.Pt(math.Float32frombits(cmd[3]), math.Float32frombits(cmd[4]))
	ctrl1 = f32.Pt(math.Float32frombits(cmd[5]), math.Float32frombits(cmd[6]))
	to = f32.Pt(math.Float32frombits(cmd[7]), math.Float32frombits(cmd[8]))
	return
}
