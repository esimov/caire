// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
)

type Bool struct {
	Value bool

	clk Clickable

	changed bool
}

// Changed reports whether Value has changed since the last
// call to Changed.
func (b *Bool) Changed() bool {
	changed := b.changed
	b.changed = false
	return changed
}

// Hovered reports whether pointer is over the element.
func (b *Bool) Hovered() bool {
	return b.clk.Hovered()
}

// Pressed reports whether pointer is pressing the element.
func (b *Bool) Pressed() bool {
	return b.clk.Pressed()
}

// Focused reports whether b has focus.
func (b *Bool) Focused() bool {
	return b.clk.Focused()
}

func (b *Bool) History() []Press {
	return b.clk.History()
}

func (b *Bool) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	dims := b.clk.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		for b.clk.Clicked() {
			b.Value = !b.Value
			b.changed = true
		}
		semantic.SelectedOp(b.Value).Add(gtx.Ops)
		semantic.DisabledOp(gtx.Queue == nil).Add(gtx.Ops)
		return w(gtx)
	})
	return dims
}
