// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"
	"image/color"
	"image/draw"

	"gioui.org/internal/f32color"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"golang.org/x/exp/shiny/iconvg"
)

type Icon struct {
	src []byte
	// Cached values.
	op       paint.ImageOp
	imgSize  int
	imgColor color.NRGBA
}

var defaultIconSize = unit.Dp(24)

// NewIcon returns a new Icon from IconVG data.
func NewIcon(data []byte) (*Icon, error) {
	_, err := iconvg.DecodeMetadata(data)
	if err != nil {
		return nil, err
	}
	return &Icon{src: data}, nil
}

// Layout displays the icon with its size set to the X minimum constraint.
func (ic *Icon) Layout(gtx layout.Context, color color.NRGBA) layout.Dimensions {
	sz := gtx.Constraints.Min.X
	if sz == 0 {
		sz = gtx.Metric.Px(defaultIconSize)
	}
	size := gtx.Constraints.Constrain(image.Pt(sz, sz))
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()

	ico := ic.image(size.X, color)
	ico.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{
		Size: ico.Size(),
	}
}

func (ic *Icon) image(sz int, color color.NRGBA) paint.ImageOp {
	if sz == ic.imgSize && color == ic.imgColor {
		return ic.op
	}
	m, _ := iconvg.DecodeMetadata(ic.src)
	dx, dy := m.ViewBox.AspectRatio()
	img := image.NewRGBA(image.Rectangle{Max: image.Point{X: sz, Y: int(float32(sz) * dy / dx)}})
	var ico iconvg.Rasterizer
	ico.SetDstImage(img, img.Bounds(), draw.Src)
	m.Palette[0] = f32color.NRGBAToLinearRGBA(color)
	iconvg.Decode(&ico, ic.src, &iconvg.DecodeOptions{
		Palette: &m.Palette,
	})
	ic.op = paint.NewImageOp(img)
	ic.imgSize = sz
	ic.imgColor = color
	return ic.op
}
