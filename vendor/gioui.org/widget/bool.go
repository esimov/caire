package widget

import (
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

// Hovered returns whether pointer is over the element.
func (b *Bool) Hovered() bool {
	return b.clk.Hovered()
}

// Pressed returns whether pointer is pressing the element.
func (b *Bool) Pressed() bool {
	return b.clk.Pressed()
}

func (b *Bool) History() []Press {
	return b.clk.History()
}

func (b *Bool) Layout(gtx layout.Context) layout.Dimensions {
	dims := b.clk.Layout(gtx)
	for b.clk.Clicked() {
		b.Value = !b.Value
		b.changed = true
	}
	return dims
}
