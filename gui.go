package caire

import (
	"fmt"
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
	"github.com/esimov/caire/utils"
)

type (
	C = layout.Context
	D = layout.Dimensions
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
		isDone bool
		img    image.Image
		seams  []Seam

		wrk <-chan worker
		err chan<- error
	}
	cp  *Processor
	ctx layout.Context
}

// NewGUI initializes the Gio interface.
func NewGUI(w, h int) *Gui {
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

	// Maintain the image aspect ratio in case the image width and height is greater than the predefined window.
	r := getRatio(w, h)
	if w > maxScreenX && h > maxScreenY {
		w = float64(w) * r
		h = float64(h) * r
	}
	return w, h
}

// Run is the core method of the Gio GUI application.
// This updates the window with the resized image received from a channel
// and terminates when the image resizing operation completes.
func (g *Gui) Run() error {
	w := app.NewWindow(app.Title(g.cfg.window.title), app.Size(
		unit.Px(float32(g.cfg.window.w)),
		unit.Px(float32(g.cfg.window.h)),
	))

	abortFn := func() {
		var dx, dy int

		if g.proc.img != nil {
			bounds := g.proc.img.Bounds()
			dx, dy = bounds.Max.X, bounds.Max.Y
		}
		if !g.proc.isDone {
			if (g.cp.NewWidth > 0 && g.cp.NewWidth != dx) ||
				(g.cp.NewHeight > 0 && g.cp.NewHeight != dy) {

				errorMsg := fmt.Sprintf("%s %s %s",
					utils.DecorateText("⚡ CAIRE", utils.StatusMessage),
					utils.DecorateText("⇢ image resizing process aborted by the user...", utils.DefaultMessage),
					utils.DecorateText("✘\n", utils.ErrorMessage),
				)
				g.cp.Spinner.StopMsg = errorMsg
				g.cp.Spinner.Stop()
			}
		}
		g.cp.Spinner.RestoreCursor()
	}

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.FrameEvent:
				g.draw(w, e)
			case key.Event:
				switch e.Name {
				case key.NameEscape:
					w.Close()
				}
			case system.DestroyEvent:
				abortFn()
				return e.Err
			}
		case res := <-g.proc.wrk:
			if res.done {
				g.proc.isDone = true
				break
			}
			g.proc.img = res.img
			g.proc.seams = res.carver.Seams
			if g.cp.vRes {
				g.proc.img = res.carver.RotateImage270(g.proc.img.(*image.NRGBA))
			}

			if resizeBothSide {
				continue
			}

			w.Invalidate()
		}
	}
}

// draw draws the resized image in the GUI window (obtained from a channel)
// and in case the debug mode is activated it prints out the seams.
func (g *Gui) draw(win *app.Window, e system.FrameEvent) {
	g.ctx = layout.NewContext(g.ctx.Ops, e)
	win.Invalidate()

	c := g.setColor(g.cfg.color.background)
	paint.Fill(g.ctx.Ops, c)

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

	// Disable the preview mode and warn the user in case the image is resized both horizontally and vertically.
	if resizeBothSide {
		var (
			msg   string
			fgcol color.NRGBA
			bgcol color.NRGBA
		)

		if !g.proc.isDone {
			msg = "Preview is not available while the image is resized both horizontally and vertically!"
			bgcol = color.NRGBA{R: 245, G: 228, B: 215, A: 0xff}
			fgcol = color.NRGBA{R: 3, G: 18, B: 14, A: 0xff}
		} else {
			msg = "Done, you may close this window!"
			bgcol = color.NRGBA{R: 15, G: 139, B: 141, A: 0xff}
			fgcol = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

		}
		displayMessage(e, g.ctx, bgcol, fgcol, msg)
	}
	e.Frame(g.ctx.Ops)
}

// displayMessage show a static message when the image is resized both horizontally and vertically.
func displayMessage(e system.FrameEvent, ctx layout.Context, bgcol, fgcol color.NRGBA, msg string) {
	var th = material.NewTheme(gofont.Collection())
	th.Palette.Fg = fgcol
	paint.ColorOp{Color: bgcol}.Add(ctx.Ops)

	rect := image.Rectangle{
		Max: image.Point{X: e.Size.X, Y: e.Size.Y},
	}
	defer clip.Rect(rect).Push(ctx.Ops).Pop()
	paint.PaintOp{}.Add(ctx.Ops)

	layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(ctx,
		layout.Flexed(1, func(gtx C) D {
			return layout.UniformInset(unit.Dp(4)).Layout(ctx, func(gtx C) D {
				return layout.Center.Layout(ctx, func(gtx C) D {
					return material.Label(th, unit.Sp(45), msg).Layout(gtx)
				})
			})
		},
		))
}
