// SPDX-License-Identifier: Unlicense OR MIT

package clipboard

import (
	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/op"
)

// Event is generated when the clipboard content is requested.
type Event struct {
	Text string
}

// ReadOp requests the text of the clipboard, delivered to
// the current handler through an Event.
type ReadOp struct {
	Tag event.Tag
}

// WriteOp copies Text to the clipboard.
type WriteOp struct {
	Text string
}

func (h ReadOp) Add(o *op.Ops) {
	data := ops.Write1(&o.Internal, ops.TypeClipboardReadLen, h.Tag)
	data[0] = byte(ops.TypeClipboardRead)
}

func (h WriteOp) Add(o *op.Ops) {
	data := ops.Write1(&o.Internal, ops.TypeClipboardWriteLen, &h.Text)
	data[0] = byte(ops.TypeClipboardWrite)
}

func (Event) ImplementsEvent() {}
