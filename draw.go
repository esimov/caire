package caire

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/esimov/caire/utils"
)

const (
	circle = "circle"
	line   = "line"
)

// DrawSeam visualizes the seam carver in action when the preview mode is activated.
// It receives as parameters the shape type, the seam (x,y) coordinate and a size.
func (g *Gui) DrawSeam(shape string, x, y, s float32) {
	r := getRatio(g.cfg.window.w, g.cfg.window.h)

	switch shape {
	case circle:
		g.drawCircle(x*r, y*r, s)
	case line:
		g.drawLine(x*r, y*r, s)
	}
}

// EncodeSeamToImg draws the seams into an image widget.
func (g *Gui) EncodeSeamToImg() {
	c := utils.HexToRGBA(g.cp.SeamColor)
	g.setFillColor(c)

	img := image.NewNRGBA(image.Rect(0, 0, int(g.cfg.window.w), int(g.cfg.window.h)))
	r := getRatio(g.cfg.window.w, g.cfg.window.h)

	for _, s := range g.proc.seams {
		x := int(float32(s.X) * r)
		y := int(float32(s.Y) * r)
		img.Set(x, y, g.getFillColor())
	}

	src := paint.NewImageOp(img)
	src.Add(g.ctx.Ops)

	widget.Image{
		Src:   src,
		Scale: 1 / float32(g.ctx.Px(unit.Dp(1))),
		Fit:   widget.Contain,
	}.Layout(g.ctx)
}

// drawCircle draws a circle at the seam (x,y) coordinate with the provided size.
func (g *Gui) drawCircle(x, y, s float32) {
	var (
		sq   float64
		p1   f32.Point
		p2   f32.Point
		orig = g.point(x-s, y)
	)

	sq = math.Sqrt(float64(s*s) - float64(s*s))
	p1 = g.point(x+float32(sq), y).Sub(orig)
	p2 = g.point(x-float32(sq), y).Sub(orig)

	col := utils.HexToRGBA(g.cp.SeamColor)
	g.setFillColor(col)

	var path clip.Path
	path.Begin(g.ctx.Ops)
	path.Move(orig)
	path.Arc(p1, p2, 2*math.Pi)
	path.Close()

	defer clip.Outline{Path: path.End()}.Op().Push(g.ctx.Ops).Pop()
	paint.ColorOp{Color: g.setColor(g.getFillColor())}.Add(g.ctx.Ops)
	paint.PaintOp{}.Add(g.ctx.Ops)
}

// drawLine draws a line at the seam (x,y) coordinate with the provided line thickness.
func (g *Gui) drawLine(x, y, s float32) {
	var (
		p1   = g.point(x, y)
		p2   = g.point(x, y+1)
		path clip.Path
	)

	path.Begin(g.ctx.Ops)
	path.Move(p1)
	path.Line(p2.Sub(path.Pos()))
	path.Close()

	col := utils.HexToRGBA(g.cp.SeamColor)
	g.setFillColor(col)

	defer clip.Stroke{Path: path.End(), Width: float32(s)}.Op().Push(g.ctx.Ops).Pop()
	paint.ColorOp{Color: g.setColor(g.getFillColor())}.Add(g.ctx.Ops)
	paint.PaintOp{}.Add(g.ctx.Ops)
}

// point converts the seam (x,y) coordinate to Gio f32.Point.
func (g *Gui) point(x, y float32) f32.Point {
	return f32.Point{
		X: x,
		Y: y,
	}
}

// setColor sets the seam color.
func (g *Gui) setColor(c color.Color) color.NRGBA {
	rc, gc, bc, ac := c.RGBA()
	return color.NRGBA{
		R: uint8(rc >> 8),
		G: uint8(gc >> 8),
		B: uint8(bc >> 8),
		A: uint8(ac >> 8),
	}
}

// setFillColor sets the paint fill color.
func (g *Gui) setFillColor(c color.Color) {
	g.cfg.color.fill = c
}

// getFillColor retrieve the paint fill color.
func (g *Gui) getFillColor() color.Color {
	return g.cfg.color.fill
}

// getRatio returns the image aspect ratio.
func getRatio(w, h float32) float32 {
	var r float32 = 1
	if w > maxScreenX && h > maxScreenY {
		wr := float32(maxScreenX) / float32(w) // width ratio
		hr := float32(maxScreenY) / float32(h) // height ratio

		r = utils.Min(wr, hr)
	}
	return r
}
