package caire

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
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
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/esimov/caire/common"
	cimage "github.com/esimov/caire/image"
	"github.com/esimov/caire/imop"
)

type hudControlType uint8

const (
	hudShowSeams hudControlType = iota
	hudShowDebugMask
)

const (
	// The starting colors for the linear gradient, used when the image is resized both horizontally and vertically.
	// In this case the preview mode is deactivated and a dynamic gradient overlay is shown.
	redStart   = 137
	greenStart = 47
	blueStart  = 54

	// The ending colors for the linear gradient. The starting colors and ending colors are lerped.
	redEnd   = 255
	greenEnd = 112
	blueEnd  = 105
)

const (
	previewMessage = "Preview is not available while the image is resized both horizontally and vertically!"
	successMessage = "Done, you can close this window!"
)

var (
	maxScreenX float32 = 1152
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
			width  float32
			height float32
			title  string
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
	process struct {
		isDone bool
		img    image.Image
		seams  []Seam

		worker <-chan worker
		err    chan<- error
	}
	proc    *Processor
	compOp  *imop.Composite
	blendOp *imop.Blend
	theme   *material.Theme
	ctx     layout.Context
	huds    map[hudControlType]*hudCtrl
	view    struct {
		huds layout.List
	}
}

type hudCtrl struct {
	enabled widget.Bool
	hudType hudControlType
	title   string
}

// NewGUI initializes the Gio interface.
func NewGUI(width, height int) *Gui {
	defaultColor := color.NRGBA{R: 0x2d, G: 0x23, B: 0x2e, A: 0xff}

	gui := &Gui{
		ctx: layout.Context{
			Ops: new(op.Ops),
			Constraints: layout.Constraints{
				Max: image.Pt(width, height),
			},
		},
		compOp:  imop.InitOp(),
		blendOp: imop.NewBlend(),
		theme:   material.NewTheme(),
		huds:    make(map[hudControlType]*hudCtrl),
	}

	gui.theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	gui.theme.TextSize = unit.Sp(16)
	gui.theme.Palette.ContrastBg = defaultColor
	gui.theme.FingerSize = 10

	gui.initWindow(width, height)

	return gui
}

// AddHudControl adds a new hud control for debugging.
func (g *Gui) AddHudControl(hudControlType hudControlType, title string, enabled bool) {
	control := &hudCtrl{
		hudType: hudControlType,
		title:   title,
		enabled: widget.Bool{},
	}
	control.enabled.Value = enabled
	g.huds[hudControlType] = control
}

// initWindow creates and initializes the GUI window.
func (g *Gui) initWindow(width, height int) {
	rand.NewSource(time.Now().UnixNano())

	g.cfg.angle = 45
	g.cfg.color.randR = uint8(randomFloat(1, 2))
	g.cfg.color.randG = uint8(randomFloat(1, 2))
	g.cfg.color.randB = uint8(randomFloat(1, 2))

	g.cfg.window.width = float32(width)
	g.cfg.window.height = float32(height)

	g.cfg.x = interval{min: 0, max: float64(width)}
	g.cfg.y = interval{min: 0, max: float64(height)}

	g.cfg.color.background = defaultBkgColor
	g.cfg.color.fill = defaultFillColor

	g.cfg.window.width, g.cfg.window.height = g.windowSize()
	g.cfg.window.title = "CAIRE - Preview"
}

// windowSize returns the resized GUI window.
func (g *Gui) windowSize() (float32, float32) {
	width, height := g.cfg.window.width, g.cfg.window.height
	ratio := getRatio(width, height)

	return width * ratio, height * ratio
}

// Run is the core method of the Gio GUI application.
// This updates the window with the resized image received from a channel
// and terminates when the image resizing operation completes.
func (g *Gui) Run() error {
	var (
		redCol   uint8 = redStart
		greenCol uint8 = greenStart
		blueCol  uint8 = blueStart

		descRed, descGreen, descBlue bool
	)

	width, height := g.windowSize()

	w := new(app.Window)
	w.Option(
		app.Title(g.cfg.window.title),
		app.Size(unit.Dp(width), unit.Dp(height)),
		app.MinSize(unit.Dp(width), unit.Dp(height)),
		app.MaxSize(unit.Dp(width), unit.Dp(height)),
	)

	// Center the window.
	w.Perform(system.ActionCenter)

	g.cfg.timeStamp = time.Now()

	if g.proc.Debug {
		g.AddHudControl(hudShowSeams, "Show seams", true)
		if g.proc.Mask != nil || g.proc.RMask != nil || g.proc.FaceDetect || g.proc.TextDetect {
			g.AddHudControl(hudShowDebugMask, "Debug mode", false)
		}
	}

	abortFn := func() {
		var dx, dy int

		if g.process.img != nil {
			bounds := g.process.img.Bounds()
			dx, dy = bounds.Max.X, bounds.Max.Y
		}

		if !g.process.isDone {
			if (g.proc.NewWidth > 0 && g.proc.NewWidth != dx) ||
				(g.proc.NewHeight > 0 && g.proc.NewHeight != dy) {

				errorMsg := fmt.Sprintf("%s %s %s",
					common.DecorateText("⚡ CAIRE", common.StatusMessage),
					common.DecorateText("⇢ process aborted by the user...", common.DefaultMessage),
					common.DecorateText("✘\n", common.ErrorMessage),
				)
				g.proc.Spinner.StopMsg = errorMsg
				g.proc.Spinner.Stop()
			}
		}
		g.proc.Spinner.RestoreCursor()
	}

	for {
		select {
		case res := <-g.process.worker:
			if res.done {
				w.Option(app.Title("CAIRE - Success!"))
				g.process.isDone = true
				break
			}
			if g.proc.resizeXY {
				continue
			}

			g.process.img = res.img
			g.process.seams = res.seams

			if mask, ok := g.huds[hudShowDebugMask]; ok {
				if mask.enabled.Value && res.mask != nil {
					bounds := res.img.Bounds()
					srcBitmap := imop.NewBitmap(bounds)
					dstBitmap := imop.NewBitmap(bounds)

					uniformCol := image.NewNRGBA(bounds)

					col := color.RGBA{R: 0x2f, G: 0xf3, B: 0xe0, A: 0xff}
					draw.Draw(uniformCol, uniformCol.Bounds(), &image.Uniform{col}, image.Point{}, draw.Src)

					_ = g.compOp.Set(imop.DstIn)
					g.compOp.Draw(srcBitmap, res.mask, uniformCol, nil)

					_ = g.blendOp.Set(imop.Screen)
					_ = g.compOp.Set(imop.SrcAtop)
					g.compOp.Draw(dstBitmap, res.img, srcBitmap.Img, g.blendOp)

					g.process.img = dstBitmap.Img
				}
			}

			if g.proc.vRes {
				g.process.img = cimage.RotateImage270(g.process.img.(*image.NRGBA))
			}
			w.Invalidate()
		default:
			switch e := w.Event().(type) {
			case app.FrameEvent:
				g.ctx = app.NewContext(g.ctx.Ops, e)

				for {
					event, ok := g.ctx.Event(key.Filter{
						Name: key.NameEscape,
					})
					if !ok {
						break
					}
					switch event := event.(type) {
					case key.Event:
						switch event.Name {
						case key.NameEscape:
							w.Perform(system.ActionClose)
							abortFn()
							return nil
						}
					}
				}

				// Disable the preview mode when the image is resized both horizontally and vertically.
				// Show a background visualization during the resize operation.
				if g.proc.resizeXY {
					{ // red
						if descRed {
							redCol--
						} else {
							redCol++
						}
						if redCol >= redEnd {
							descRed = !descRed
						}
						if redCol == redStart {
							descRed = !descRed
						}
					}
					{ // green
						if descGreen {
							greenCol--
						} else {
							greenCol++
						}
						if greenCol >= greenEnd {
							descGreen = !descGreen
						}
						if greenCol == greenStart {
							descGreen = !descGreen
						}
					}
					{ // blue
						if descBlue {
							blueCol--
						} else {
							blueCol++
						}
						if blueCol >= blueEnd {
							descBlue = !descBlue
						}
						if blueCol == blueStart {
							descBlue = !descBlue
						}
					}
					g.visualizeProcess(color.NRGBA{R: redCol, G: greenCol, B: blueCol})
				} else {
					g.draw()
				}
				e.Frame(g.ctx.Ops)
			case app.DestroyEvent:
				abortFn()
				return e.Err
			}
		}
	}
}

type (
	C = layout.Context
	D = layout.Dimensions
)

// draw draws the resized image in the GUI window (obtained from a channel)
// and in case the debug mode is activated it prints out the seams.
func (g *Gui) draw() {
	g.ctx.Execute(op.InvalidateCmd{})

	c := g.setColor(g.cfg.color.background)
	paint.Fill(g.ctx.Ops, c)

	if g.process.img != nil {
		img := g.process.img
		src := paint.NewImageOp(img)
		src.Add(g.ctx.Ops)

		layout.Stack{}.Layout(g.ctx,
			layout.Stacked(func(gtx C) D {
				imageWidget := widget.Image{
					Src:      src,
					Scale:    1 / float32(unit.Dp(1)),
					Fit:      widget.Contain,
					Position: layout.Center,
				}

				imageDimension := imageWidget.Layout(gtx)

				if seam, ok := g.huds[hudShowSeams]; ok && seam.enabled.Value {
					imageWidth := float32(img.Bounds().Dx())
					imageHeight := float32(img.Bounds().Dy())

					renderedWidth := float32(imageDimension.Size.X)
					renderedHeight := float32(imageDimension.Size.Y)

					scaleX := renderedWidth / imageWidth
					scaleY := renderedHeight / imageHeight

					offsetX := float32(gtx.Constraints.Max.X-imageDimension.Size.X) / 2
					offsetY := float32(gtx.Constraints.Max.Y-imageDimension.Size.Y) / 2

					for _, seam := range g.process.seams {
						x := float32(seam.X)
						y := float32(seam.Y)

						if g.proc.vRes {
							originalHeight := float32(img.Bounds().Dx())
							x, y = originalHeight-1-y, x
						}

						px := offsetX + x*scaleX
						py := offsetY + y*scaleY

						g.DrawSeam(g.proc.ShapeType, px, py, 1)
					}
				}

				return imageDimension
			}),
		)
	}

	if g.proc.Debug {
		layout.Stack{}.Layout(g.ctx,
			layout.Stacked(func(gtx C) D {
				hudHeight := 30
				r := image.Rectangle{
					Max: image.Point{
						X: gtx.Constraints.Max.X,
						Y: hudHeight,
					},
				}
				defer op.Offset(image.Pt(0, gtx.Constraints.Max.Y-hudHeight)).Push(gtx.Ops).Pop()
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xcc}, clip.Rect(r).Op())
						return layout.Dimensions{Size: r.Max}
					}),
					layout.Stacked(func(gtx C) D {
						border := image.Rectangle{
							Max: image.Point{
								X: gtx.Constraints.Max.X,
								Y: gtx.Dp(unit.Dp(0.5)),
							},
						}
						paint.FillShape(gtx.Ops, color.NRGBA{R: 0xd0, G: 0xcd, B: 0xd7, A: 0xaa}, clip.Rect(border).Op())
						return layout.Dimensions{Size: r.Max}
					}),
					layout.Stacked(func(gtx C) D {
						return g.view.huds.Layout(gtx, len(g.huds),
							func(gtx layout.Context, index int) D {
								if hud, ok := g.huds[hudControlType(index)]; ok {
									checkbox := material.CheckBox(g.theme, &hud.enabled, fmt.Sprintf("%v", hud.title))
									checkbox.Size = 20
									return checkbox.Layout(gtx)
								}
								return D{}
							})
					}),
				)
			}),
		)
	}
}

// visualizeProcess shows a background visualization during the resize process on X and Y axis.
func (g *Gui) visualizeProcess(bgColor color.NRGBA) {
	g.ctx.Execute(op.InvalidateCmd{})

	var msg string

	if !g.process.isDone {
		msg = previewMessage
	} else {
		msg = successMessage
		bgColor = color.NRGBA{R: 45, G: 45, B: 42, A: 0xff}
	}

	g.theme.Palette.Fg = color.NRGBA{R: 251, G: 254, B: 249, A: 0xff}
	paint.ColorOp{Color: bgColor}.Add(g.ctx.Ops)

	rect := image.Rectangle{
		Max: g.ctx.Constraints.Max,
	}

	defer clip.Rect(rect).Push(g.ctx.Ops).Pop()
	paint.PaintOp{}.Add(g.ctx.Ops)

	layout.Stack{}.Layout(g.ctx,
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(4)).Layout(g.ctx, func(gtx C) D {
				if !g.process.isDone {
					gtx.Constraints.Min.Y = 0
					tr := f32.Affine2D{}
					dr := image.Rectangle{Max: gtx.Constraints.Min}

					tr = tr.Rotate(
						f32.Pt(
							float32(g.ctx.Constraints.Max.X/2),
							float32(g.ctx.Constraints.Max.Y/2)),
						0.005*-g.cfg.angle)
					op.Affine(tr).Add(gtx.Ops)

					since := time.Since(g.cfg.timeStamp)

					if since.Seconds() > 5 {
						g.cfg.timeStamp = time.Now()
						g.cfg.color.randR = uint8(randomFloat(1, 2))
						g.cfg.color.randG = uint8(randomFloat(1, 2))
						g.cfg.color.randB = uint8(randomFloat(1, 2))
					}

					paint.LinearGradientOp{
						Stop1:  layout.FPt(dr.Min.Div(2)),
						Stop2:  layout.FPt(dr.Max.Mul(2)),
						Color1: color.NRGBA{R: 41, G: bgColor.G * g.cfg.color.randG, B: bgColor.B * g.cfg.color.randB, A: 0xFF},
						Color2: color.NRGBA{R: bgColor.R * g.cfg.color.randR, G: 29, B: 54, A: 0xFF},
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
			return layout.UniformInset(unit.Dp(4)).Layout(g.ctx, func(gtx C) D {
				return layout.Center.Layout(g.ctx, func(gtx C) D {
					m := material.Label(g.theme, unit.Sp(40), msg)
					m.Alignment = text.Middle

					return m.Layout(gtx)
				})
			})
		}),
		layout.Stacked(func(gtx C) D {
			info := "(You will be notified once the process is finished.)"
			if g.process.isDone {
				return layout.Dimensions{}
			}

			return layout.Inset{Top: 70}.Layout(g.ctx, func(gtx C) D {
				return layout.Center.Layout(g.ctx, func(gtx C) D {
					return material.Label(g.theme, unit.Sp(13), info).Layout(gtx)
				})
			})
		}),
	)
}

// randomFloat generates a random float32 number between two numbers.
func randomFloat(min, max float32) float32 {
	return rand.Float32()*(max-min) + min
}
