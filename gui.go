package caire

import (
	"image"
	"image/color"
	"math"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
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
		trX    func(v float64) float64 // translate from user- to system coords
		trY    func(v float64) float64 // translate from user- to system coords
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

	// Resize the image but retain the aspect ratio in case the
	// image width and height is greater than the predefined window.
	if w > maxScreenX || h > maxScreenY {
		wr := float64(maxScreenX) / float64(w) // width ratio
		hr := float64(maxScreenY) / float64(h) // height ratio
		ratio := math.Min(wr, hr)

		w = float64(w) * ratio
		h = float64(h) * ratio
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

	color := g.setColor(g.cfg.color.background)
	paint.Fill(g.ctx.Ops, color)

	if g.proc.img != nil {
		src := paint.NewImageOp(g.proc.img)
		src.Add(g.ctx.Ops)

		imgWidget := widget.Image{
			Src:   src,
			Scale: 1 / float32(g.ctx.Px(unit.Dp(1))),
			Fit:   widget.Contain,
		}

		imgWidget.Layout(g.ctx)

		if g.cp.Debug {
			for _, s := range g.proc.seams {
				g.DrawSeam(line, float64(s.X), float64(s.Y), 2)
			}
		}
	}
	e.Frame(g.ctx.Ops)
}
