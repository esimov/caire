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
)

type shapeType int

const (
	circle shapeType = iota
	line
)

// DrawSeam visualizes the seam carver in action when the preview mode is activated.
// It receives as parameters the shape type, the seam (x,y) coordinate and a size.
func (g *Gui) DrawSeam(shape shapeType, x, y, s float64) {
	switch shape {
	case circle:
		g.drawCircle(x, y, s)
	case line:
		g.drawCircle(x, y, s)
	}
}

// EncodeSeamToImg draws the seams into an image widget.
func (g *Gui) EncodeSeamToImg() {
	g.setFillColor(color.White)

	img := image.NewNRGBA(image.Rect(0, 0, int(g.cfg.window.w), int(g.cfg.window.h)))
	for _, s := range g.proc.seams {
		img.Set(s.X, s.Y, g.getFillColor())
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
func (g *Gui) drawCircle(x, y, s float64) {
	var (
		sq   float64
		p1   f32.Point
		p2   f32.Point
		orig = g.point(x-s, y)
	)

	sq = math.Sqrt(s*s - s*s)
	p1 = g.point(x+sq, y).Sub(orig)
	p2 = g.point(x-sq, y).Sub(orig)

	g.setFillColor(color.RGBA{R: 0xff, A: 0xff})

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
func (g *Gui) drawLine(x, y, s float64) {
	var (
		p1   = g.point(x, y)
		p2   = g.point(x, y+1)
		path clip.Path
	)

	path.Begin(g.ctx.Ops)
	path.Move(p1)
	path.Line(p2.Sub(path.Pos()))
	path.Close()

	g.setFillColor(color.RGBA{R: 0xff, A: 0xff})

	defer clip.Stroke{Path: path.End(), Width: float32(s)}.Op().Push(g.ctx.Ops).Pop()
	paint.ColorOp{Color: g.setColor(g.getFillColor())}.Add(g.ctx.Ops)
	paint.PaintOp{}.Add(g.ctx.Ops)
}

// point converts the seam (x,y) coordinate to Gio f32.Point.
func (g *Gui) point(x, y float64) f32.Point {
	return f32.Point{
		X: float32(x),
		Y: float32(y),
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
