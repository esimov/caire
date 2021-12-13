package caire

import (
	"image"
	"math"
	"os"

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
	MaxScreenX = 1366
	MaxScreenY = 768
)

// showPreview spawn a new Gio GUI window and updates its content
// with the resized image recived from a channel.
func (p *Processor) showPreview(
	workerChan <-chan worker,
	errChan chan<- error,
	guiParams struct {
		width  int
		height int
	},
) {
	width, height := guiParams.width, guiParams.height
	newWidth, newHeight := float64(width), float64(height)

	// Resize the image but retain the aspect ratio in case the
	// image width and height is greater than the predefined window.
	if width > MaxScreenX && height > MaxScreenY {
		widthRatio := float64(MaxScreenX) / float64(width)
		heightRatio := float64(MaxScreenY) / float64(height)
		ratio := math.Min(widthRatio, heightRatio)

		newWidth = float64(width) * ratio
		newHeight = float64(height) * ratio
	}

	// Create a new window.
	w := app.NewWindow(
		app.Title("Image resize in progress..."),
		app.Size(unit.Px(float32(newWidth)), unit.Px(float32(newHeight))),
	)

	// Run the Gio GUI app in a seperate goroutine
	go func() {
		if err := p.run(w, workerChan); err != nil {
			errChan <- err
		}
		// It's important to call os.Exit(0) in order to terminate
		// the execution of the GUI app when pressing ESC key.
		os.Exit(0)
	}()
}

// run the Gio main thread until a DestroyEvent or an ESC key event is captured.
func (p *Processor) run(w *app.Window, workerChan <-chan worker) error {
	var (
		ops op.Ops
		img image.Image
	)

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				w.Invalidate()

				if img != nil {
					src := paint.NewImageOp(img)
					src.Add(gtx.Ops)

					imgWidget := widget.Image{
						Src:   src,
						Scale: 1 / float32(gtx.Px(unit.Dp(1))),
						Fit:   widget.Contain,
					}
					imgWidget.Layout(gtx)
				}
				e.Frame(gtx.Ops)
			case key.Event:
				switch e.Name {
				case key.NameEscape:
					p.Spinner.RestoreCursor()
					w.Close()
				}
			case system.DestroyEvent:
				return e.Err
			}
		case worker := <-workerChan:
			img = worker.img
			if p.vRes {
				img = worker.carver.RotateImage270(img.(*image.NRGBA))
			}
			w.Invalidate()
		}
	}
}
