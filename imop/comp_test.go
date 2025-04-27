package imop

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComp_Basic(t *testing.T) {
	assert := assert.New(t)

	op := InitOp()

	op.Set(Clear)
	assert.Equal(Clear, op.Get())
	assert.NotEqual("unsupported_composite_operation", op.Get())

	op.Set(Dst)
	assert.Equal(Dst, op.Get())
}

func TestComp_Ops(t *testing.T) {
	assert := assert.New(t)
	op := InitOp()

	transparent := color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	cyan := color.NRGBA{R: 33, G: 150, B: 243, A: 255}
	magenta := color.NRGBA{R: 233, G: 30, B: 99, A: 255}

	rect := image.Rect(0, 0, 10, 10)
	bmp := NewBitmap(rect)
	source := image.NewNRGBA(rect)
	backdrop := image.NewNRGBA(rect)

	// No composition operation applied. The SrcOver is the default one.
	draw.Draw(source, image.Rect(0, 4, 6, 10), &image.Uniform{cyan}, image.Point{}, draw.Src)
	draw.Draw(backdrop, image.Rect(4, 0, 10, 6), &image.Uniform{magenta}, image.Point{}, draw.Src)
	op.Draw(bmp, source, backdrop, nil)

	// Pick three representative points/pixels from the generated image output.
	// Depending on the applied composition operation the colors of the
	// selected pixels should be the source color, the destination color or transparent.
	topRight := bmp.Img.At(9, 0)
	bottomLeft := bmp.Img.At(0, 9)
	center := bmp.Img.At(5, 5)

	assert.EqualValues(topRight, magenta)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, cyan)

	// Clear
	op.Set(Clear)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, transparent)
	assert.EqualValues(bottomLeft, transparent)
	assert.EqualValues(center, transparent)

	// Copy
	op.Set(Copy)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, transparent)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, cyan)

	// Dst
	op.Set(Dst)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, magenta)
	assert.EqualValues(bottomLeft, transparent)
	assert.EqualValues(center, magenta)

	// SrcOver
	op.Set(SrcOver)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, magenta)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, cyan)

	// DstOver
	op.Set(DstOver)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, magenta)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, magenta)

	// SrcIn
	op.Set(SrcIn)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, transparent)
	assert.EqualValues(bottomLeft, transparent)
	assert.EqualValues(center, cyan)

	// DstIn
	op.Set(DstIn)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, transparent)
	assert.EqualValues(bottomLeft, transparent)
	assert.EqualValues(center, magenta)

	// SrcOut
	op.Set(SrcOut)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, transparent)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, transparent)

	// DstOut
	op.Set(DstOut)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, magenta)
	assert.EqualValues(bottomLeft, transparent)
	assert.EqualValues(center, transparent)

	// SrcAtop
	op.Set(SrcAtop)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, magenta)
	assert.EqualValues(bottomLeft, transparent)
	assert.EqualValues(center, cyan)

	// DstAtop
	op.Set(DstAtop)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, transparent)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, magenta)

	// Xor
	op.Set(Xor)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, magenta)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, transparent)
	// DstAtop
	op.Set(DstAtop)
	op.Draw(bmp, source, backdrop, nil)

	topRight = bmp.Img.At(9, 0)
	bottomLeft = bmp.Img.At(0, 9)
	center = bmp.Img.At(5, 5)

	assert.EqualValues(topRight, transparent)
	assert.EqualValues(bottomLeft, cyan)
	assert.EqualValues(center, magenta)
}
