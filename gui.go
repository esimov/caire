package caire

import (
	"image"
	"image/color"
	"math"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const (
	maxScreenX = 1366
	maxScreenY = 768
)

var (
	defaultBkgColor  = color.Transparent
	defaultFillColor = color.Black
)

type interval struct {
	min, max float64
}

// Gui is the basic struct containing all of the information needed for the UI operation.
// It receives the resized image transferred through a channel which is called in a separate goroutine.
type Gui struct {
	cfg struct {
		x      interval
		y      interval
		window struct {
			w     float64
			h     float64
			title string
		}
		color struct {
			background color.Color
			fill       color.Color
		}
	}
	proc struct {
		img   image.Image
		seams []Seam

		wrk <-chan worker
		err chan<- error
	}
	cp  *Processor
	ctx layout.Context
}

// newGui initializes the Gio interface.
func newGui(w, h int) *Gui {
	gui := &Gui{
		ctx: layout.Context{
			Ops: new(op.Ops),
			Constraints: layout.Constraints{
				Max: image.Pt(w, h),
			},
		},
	}
	gui.initWindow(w, h)

	return gui
}

// initWindow creates and initializes the GUI window.
func (g *Gui) initWindow(w, h int) {
	g.cfg.window.w, g.cfg.window.h = float64(w), float64(h)
	g.cfg.x = interval{min: 0, max: float64(w)}
	g.cfg.y = interval{min: 0, max: float64(h)}

	g.cfg.color.background = defaultBkgColor
	g.cfg.color.fill = defaultFillColor

	g.cfg.window.w, g.cfg.window.h = g.getWindowSize()
	g.cfg.window.title = "Image resize in progress..."
}

// getWindowSize returns the resized image dimmension.
func (g *Gui) getWindowSize() (float64, float64) {
	w, h := g.cfg.window.w, g.cfg.window.h

	// retains the aspect ratio in case the image width and height
	// is greater than the predefined window.
	r := getRatio(w, h)
	if w > maxScreenX && h > maxScreenY {
		w = float64(w) * r
		h = float64(h) * r
	}
	return w, h
}

// Run is the core method of the Gio GUI application.
// This updates the window with the resized image obtained from a channel
// and terminates when the image resizing operation completes.
func (g *Gui) Run() error {
	w := app.NewWindow(app.Title(g.cfg.window.title), app.Size(
		unit.Px(float32(g.cfg.window.w)),
		unit.Px(float32(g.cfg.window.h)),
	))

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.FrameEvent:
				g.draw(w, e)
			case key.Event:
				switch e.Name {
				case key.NameEscape:
					g.cp.Spinner.RestoreCursor()
					w.Close()
				}
			case system.DestroyEvent:
				return e.Err
			}
		case res := <-g.proc.wrk:
			if resizeBothSide {
				continue
			}
			g.proc.img = res.img
			g.proc.seams = res.carver.Seams
			if g.cp.vRes {
				g.proc.img = res.carver.RotateImage270(g.proc.img.(*image.NRGBA))
			}
			w.Invalidate()
		}
	}
}

// draw display the resized image received from a channel
// and prints the seams in case the debug mode is activated.
func (g *Gui) draw(win *app.Window, e system.FrameEvent) {
	g.ctx = layout.NewContext(g.ctx.Ops, e)
	win.Invalidate()

	c := g.setColor(g.cfg.color.background)
	paint.Fill(g.ctx.Ops, c)

	type (
		C = layout.Context
		D = layout.Dimensions
	)

	if g.proc.img != nil {
		src := paint.NewImageOp(g.proc.img)
		src.Add(g.ctx.Ops)

		layout.Flex{
			Axis: layout.Horizontal,
		}.Layout(g.ctx,
			layout.Flexed(1, func(gtx C) D {
				paint.FillShape(gtx.Ops, c,
					clip.Rect{Max: g.ctx.Constraints.Max}.Op(),
				)
				return layout.UniformInset(unit.Px(0)).Layout(gtx,
					func(gtx C) D {
						widget.Image{
							Src:   src,
							Scale: 1 / float32(g.ctx.Px(unit.Dp(1))),
							Fit:   widget.Contain,
						}.Layout(gtx)

						if g.cp.Debug {
							var ratio float32
							tr := f32.Affine2D{}
							screen := layout.FPt(g.ctx.Constraints.Max)
							width, height := float32(g.proc.img.Bounds().Dx()), float32(g.proc.img.Bounds().Dy())
							sw, sh := float32(screen.X), float32(screen.Y)

							if sw > width {
								ratio = sw / width
								tr = tr.Scale(f32.Pt(sw/2, sh/2), f32.Pt(1, ratio))
							} else if sh > height {
								ratio = sh / height
								tr = tr.Scale(f32.Pt(sw/2, sh/2), f32.Pt(ratio, 1))
							}

							if g.cp.vRes {
								angle := float32(270 * math.Pi / 180)
								half := float32(math.Round(float64(sh*0.5-height*0.5) * 0.5))

								ox := math.Abs(float64(sw - (sw - (sw/2 - sh/2))))
								oy := math.Abs(float64(sh - (sh - (sw/2 - height/2 + half))))
								tr = tr.Rotate(f32.Pt(sw/2, sh/2), -angle)

								if screen.X > screen.Y {
									tr = tr.Offset(f32.Pt(float32(ox), float32(oy)))
								} else {
									tr = tr.Offset(f32.Pt(float32(-ox), float32(-oy)))
								}
							}
							op.Affine(tr).Add(gtx.Ops)

							for _, s := range g.proc.seams {
								g.DrawSeam(g.cp.ShapeType, float64(s.X), float64(s.Y), 1)
							}
						}
						return layout.Dimensions{Size: gtx.Constraints.Max}
					})
			}),
		)

	}

	if resizeBothSide {
		msg := "Preview is not available when the image is resized horizontally and vertically at the same time!"

		var th = material.NewTheme(gofont.Collection())
		th.Palette.Fg = color.NRGBA{R: 0xFF, A: 0xFF}
		layout.UniformInset(unit.Px(20)).Layout(g.ctx, func(gtx C) D {
			return layout.Center.Layout(g.ctx, func(gtx C) D {
				return material.Label(th, unit.Sp(40), msg).Layout(gtx)
			})
		})
	}
	e.Frame(g.ctx.Ops)
}
