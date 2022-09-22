package caire

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

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

const (
	// The starting colors for the linear gradient, used when the image is resized both horzontally and vertically.
	// In this case the preview mode is deactivated and a dynamic gradient overlay is shown.
	redStart   = 137
	greenStart = 47
	blueStart  = 54

	// The ending colors for the linear gradient. The starting colors and ending colors are lerped.
	redEnd   = 255
	greenEnd = 112
	blueEnd  = 105
)

var (
	maxScreenX float32 = 1024
	maxScreenY float32 = 768

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
		chrot  bool
		angle  float32
		window struct {
			w     float32
			h     float32
			title string
		}
		color struct {
			randR uint8
			randG uint8
			randB uint8

			background color.Color
			fill       color.Color
		}
		timeStamp time.Time
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
	rand.NewSource(time.Now().UnixNano())

	g.cfg.angle = 45
	g.cfg.color.randR = uint8(random(1, 2))
	g.cfg.color.randG = uint8(random(1, 2))
	g.cfg.color.randB = uint8(random(1, 2))

	g.cfg.window.w, g.cfg.window.h = float32(w), float32(h)
	g.cfg.x = interval{min: 0, max: float64(w)}
	g.cfg.y = interval{min: 0, max: float64(h)}

	g.cfg.color.background = defaultBkgColor
	g.cfg.color.fill = defaultFillColor

	if !resizeXY {
		g.cfg.window.w, g.cfg.window.h = g.getWindowSize()
	}
	g.cfg.window.title = "Preview"
}

// getWindowSize returns the resized image dimmension.
func (g *Gui) getWindowSize() (float32, float32) {
	w, h := g.cfg.window.w, g.cfg.window.h
	// Maintain the image aspect ratio in case the image width and height is greater than the predefined window.
	r := getRatio(w, h)
	if w > maxScreenX && h > maxScreenY {
		w = w * r
		h = h * r
	}
	return w, h
}

// Run is the core method of the Gio GUI application.
// This updates the window with the resized image received from a channel
// and terminates when the image resizing operation completes.
func (g *Gui) Run() error {
	var (
		rc uint8 = redStart
		gc uint8 = greenStart
		bc uint8 = blueStart

		descRed, descGreen, descBlue bool
	)
	w := app.NewWindow(app.Title(g.cfg.window.title), app.Size(
		unit.Dp(g.cfg.window.w),
		unit.Dp(g.cfg.window.h),
	))
	g.cfg.timeStamp = time.Now()

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
					utils.DecorateText("⇢ process aborted by the user...", utils.DefaultMessage),
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
				gtx := layout.NewContext(g.ctx.Ops, e)

				key.InputOp{Tag: w, Keys: key.NameEscape}.Add(gtx.Ops)
				for _, ev := range gtx.Queue.Events(w) {
					if e, ok := ev.(key.Event); ok && e.Name == key.NameEscape {
						w.Perform(system.ActionClose)
					}
				}

				{ // red
					if descRed {
						rc--
					} else {
						rc++
					}
					if rc >= redEnd {
						descRed = !descRed
					}
					if rc == redStart {
						descRed = !descRed
					}
				}
				{ // green
					if descGreen {
						gc--
					} else {
						gc++
					}
					if gc >= greenEnd {
						descGreen = !descGreen
					}
					if gc == greenStart {
						descGreen = !descGreen
					}
				}
				{ // blue
					if descBlue {
						bc--
					} else {
						bc++
					}
					if bc >= blueEnd {
						descBlue = !descBlue
					}
					if bc == blueStart {
						descBlue = !descBlue
					}
				}
				g.draw(gtx, color.NRGBA{R: rc, G: gc, B: bc})
				e.Frame(gtx.Ops)
			case system.DestroyEvent:
				abortFn()
				return e.Err
			}
		case res := <-g.proc.wrk:
			if res.done {
				g.proc.isDone = true
				break
			}
			if resizeXY {
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

type (
	C = layout.Context
	D = layout.Dimensions
)

// draw draws the resized image in the GUI window (obtained from a channel)
// and in case the debug mode is activated it prints out the seams.
func (g *Gui) draw(gtx layout.Context, bgCol color.NRGBA) {
	g.ctx = gtx
	op.InvalidateOp{}.Add(gtx.Ops)

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
				return layout.UniformInset(unit.Dp(0)).Layout(gtx,
					func(gtx C) D {
						widget.Image{
							Src:   src,
							Scale: 1 / float32(unit.Dp(1)),
							Fit:   widget.Contain,
						}.Layout(gtx)

						if g.cp.Debug {
							var ratio float32 = 1
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
								var dpi unit.Dp
								dpx := unit.Dp(s.X)
								dpy := unit.Dp(s.Y)

								if int(screen.Y) > len(g.proc.seams) {
									// Apply the pixel to dpi conversion formula in case
									// the screen height is greather than the image height.
									dpi = unit.Dp(float32(g.cfg.window.h) * 0.4 / float32(160))
								}
								g.DrawSeam(g.cp.ShapeType, float32(dpx), float32(dpy*dpi), float32(g.cp.ShapeStroke))
							}
						}
						return layout.Dimensions{Size: gtx.Constraints.Max}
					})
			}),
		)
	}

	// Disable the preview mode and warn the user in case the image is resized both horizontally and vertically.
	if resizeXY {
		var msg string

		if !g.proc.isDone {
			msg = "Preview is not available while the image is resized both horizontally and vertically!"
		} else {
			msg = "Done, you may close this window!"
			bgCol = color.NRGBA{R: 45, G: 45, B: 42, A: 0xff}
		}
		g.displayMessage(g.ctx, bgCol, msg)
	}
}

// displayMessage show a static message when the image is resized both horizontally and vertically.
func (g *Gui) displayMessage(ctx layout.Context, bgCol color.NRGBA, msg string) {
	th := material.NewTheme(gofont.Collection())
	th.Palette.Fg = color.NRGBA{R: 251, G: 254, B: 249, A: 0xff}
	paint.ColorOp{Color: bgCol}.Add(ctx.Ops)

	rect := image.Rectangle{
		Max: ctx.Constraints.Max,
	}

	defer clip.Rect(rect).Push(ctx.Ops).Pop()
	paint.PaintOp{}.Add(ctx.Ops)

	layout.Stack{}.Layout(ctx,
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(4)).Layout(ctx, func(gtx C) D {
				if !g.proc.isDone {
					gtx.Constraints.Min.Y = 0
					tr := f32.Affine2D{}
					dr := image.Rectangle{Max: gtx.Constraints.Min}

					tr = tr.Rotate(f32.Pt(float32(ctx.Constraints.Max.X/2), float32(ctx.Constraints.Max.Y/2)), 0.005*-g.cfg.angle)
					op.Affine(tr).Add(gtx.Ops)

					since := time.Since(g.cfg.timeStamp)

					if since.Seconds() > 5 {
						g.cfg.timeStamp = time.Now()
						g.cfg.color.randR = uint8(random(1, 2))
						g.cfg.color.randG = uint8(random(1, 2))
						g.cfg.color.randB = uint8(random(1, 2))
					}

					paint.LinearGradientOp{
						Stop1:  layout.FPt(dr.Min.Div(2)),
						Stop2:  layout.FPt(dr.Max.Mul(2)),
						Color1: color.NRGBA{R: 41, G: bgCol.G * g.cfg.color.randG, B: bgCol.B * g.cfg.color.randB, A: 0xFF},
						Color2: color.NRGBA{R: bgCol.R * g.cfg.color.randR, G: 29, B: 54, A: 0xFF},
					}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)

					if g.cfg.chrot {
						g.cfg.angle--
					} else {
						g.cfg.angle++
					}
					if g.cfg.angle == -90 || g.cfg.angle == 90 {
						g.cfg.chrot = !g.cfg.chrot
					}
				}

				return layout.Dimensions{
					Size: gtx.Constraints.Max,
				}
			})
		}),
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(4)).Layout(ctx, func(gtx C) D {
				return layout.Center.Layout(ctx, func(gtx C) D {
					return material.Label(th, unit.Sp(40), msg).Layout(gtx)
				})
			})
		}),
		layout.Stacked(func(gtx C) D {
			info := "(You will be notified once the process is finished.)"
			if g.proc.isDone {
				return layout.Dimensions{}
			}

			return layout.Inset{Top: 70}.Layout(ctx, func(gtx C) D {
				return layout.Center.Layout(ctx, func(gtx C) D {
					return material.Label(th, unit.Sp(13), info).Layout(gtx)
				})
			})
		}),
	)
}

// random generates a random number between two numbers.
func random(min, max float32) float32 {
	return rand.Float32()*(max-min) + min
}
