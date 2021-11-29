package widget

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

type Enum struct {
	Value    string
	hovered  string
	hovering bool

	changed bool

	clicks []gesture.Click
	values []string
}

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// Changed reports whether Value has changed by user interaction since the last
// call to Changed.
func (e *Enum) Changed() bool {
	changed := e.changed
	e.changed = false
	return changed
}

// Hovered returns the key that is highlighted, or false if none are.
func (e *Enum) Hovered() (string, bool) {
	return e.hovered, e.hovering
}

// Layout adds the event handler for key.
func (e *Enum) Layout(gtx layout.Context, key string) layout.Dimensions {
	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Push(gtx.Ops).Pop()

	if index(e.values, key) == -1 {
		e.values = append(e.values, key)
		e.clicks = append(e.clicks, gesture.Click{})
		e.clicks[len(e.clicks)-1].Add(gtx.Ops)
	} else {
		idx := index(e.values, key)
		clk := &e.clicks[idx]
		for _, ev := range clk.Events(gtx) {
			switch ev.Type {
			case gesture.TypeClick:
				if new := e.values[idx]; new != e.Value {
					e.Value = new
					e.changed = true
				}
			}
		}
		if e.hovering && e.hovered == key {
			e.hovering = false
		}
		if clk.Hovered() {
			e.hovered = key
			e.hovering = true
		}
		clk.Add(gtx.Ops)
	}

	return layout.Dimensions{Size: gtx.Constraints.Min}
}
