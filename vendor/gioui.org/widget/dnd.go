package widget

import (
	"io"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// Draggable makes a widget draggable.
type Draggable struct {
	// Type contains the MIME type and matches transfer.SourceOp.
	Type string

	handle    struct{}
	drag      gesture.Drag
	click     f32.Point
	pos       f32.Point
	requested bool
	request   string
}

func (d *Draggable) Layout(gtx layout.Context, w, drag layout.Widget) layout.Dimensions {
	if gtx.Queue == nil {
		return w(gtx)
	}
	pos := d.pos
	for _, ev := range d.drag.Events(gtx.Metric, gtx.Queue, gesture.Both) {
		switch ev.Type {
		case pointer.Press:
			d.click = ev.Position
			pos = f32.Point{}
		case pointer.Drag, pointer.Release:
			pos = ev.Position.Sub(d.click)
		}
	}
	d.pos = pos

	for _, ev := range gtx.Queue.Events(&d.handle) {
		switch e := ev.(type) {
		case transfer.RequestEvent:
			d.requested = true
			d.request = e.Type
		case transfer.CancelEvent:
			d.requested = false
			d.request = ""
		}
	}

	dims := w(gtx)

	stack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	d.drag.Add(gtx.Ops)
	transfer.SourceOp{
		Tag:  &d.handle,
		Type: d.Type,
	}.Add(gtx.Ops)
	stack.Pop()

	if drag != nil && d.drag.Pressed() {
		rec := op.Record(gtx.Ops)
		op.Offset(pos.Round()).Add(gtx.Ops)
		drag(gtx)
		op.Defer(gtx.Ops, rec.Stop())
	}

	return dims
}

// Dragging returns whether d is being dragged.
func (d *Draggable) Dragging() bool {
	return d.drag.Dragging()
}

// Requested returns the MIME type, if any, for which the Draggable was requested to offer data.
func (d *Draggable) Requested() (mime string, requested bool) {
	mime = d.request
	requested = d.requested
	d.requested = false
	d.request = ""
	return
}

// Offer the data ready for a drop. Must be called after being Requested.
// The mime must be one in the requested list.
func (d *Draggable) Offer(ops *op.Ops, mime string, data io.ReadCloser) {
	transfer.OfferOp{
		Tag:  &d.handle,
		Type: mime,
		Data: data,
	}.Add(ops)
}

// Pos returns the drag position relative to its initial click position.
func (d *Draggable) Pos() f32.Point {
	return d.pos
}
