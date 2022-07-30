// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

type Enum struct {
	Value    string
	hovered  string
	hovering bool

	focus   string
	focused bool

	changed bool

	keys []*enumKey
}

type enumKey struct {
	key   string
	click gesture.Click
	tag   struct{}
}

func (e *Enum) index(k string) *enumKey {
	for _, v := range e.keys {
		if v.key == k {
			return v
		}
	}
	return nil
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

// Focused reports the focused key, or false if no key is focused.
func (e *Enum) Focused() (string, bool) {
	return e.focus, e.focused
}

// Layout adds the event handler for the key k.
func (e *Enum) Layout(gtx layout.Context, k string, content layout.Widget) layout.Dimensions {
	m := op.Record(gtx.Ops)
	dims := content(gtx)
	c := m.Stop()
	defer clip.Rect{Max: dims.Size}.Push(gtx.Ops).Pop()

	state := e.index(k)
	if state == nil {
		state = &enumKey{
			key: k,
		}
		e.keys = append(e.keys, state)
	}
	clk := &state.click
	for _, ev := range clk.Events(gtx) {
		switch ev.Type {
		case gesture.TypePress:
			if ev.Source == pointer.Mouse {
				key.FocusOp{Tag: &state.tag}.Add(gtx.Ops)
			}
		case gesture.TypeClick:
			if state.key != e.Value {
				e.Value = state.key
				e.changed = true
			}
		}
	}
	for _, ev := range gtx.Events(&state.tag) {
		switch ev := ev.(type) {
		case key.FocusEvent:
			if ev.Focus {
				e.focused = true
				e.focus = state.key
			} else if state.key == e.focus {
				e.focused = false
			}
		case key.Event:
			if !e.focused || ev.State != key.Release {
				break
			}
			if ev.Name != key.NameReturn && ev.Name != key.NameSpace {
				break
			}
			if state.key != e.Value {
				e.Value = state.key
				e.changed = true
			}
		}
	}
	if clk.Hovered() {
		e.hovered = k
		e.hovering = true
	} else if e.hovered == k {
		e.hovering = false
	}

	clk.Add(gtx.Ops)
	disabled := gtx.Queue == nil
	if !disabled {
		key.InputOp{Tag: &state.tag, Keys: "⏎|Space"}.Add(gtx.Ops)
	} else if e.focus == k {
		e.focused = false
	}
	semantic.SelectedOp(k == e.Value).Add(gtx.Ops)
	semantic.DisabledOp(disabled).Add(gtx.Ops)
	c.Add(gtx.Ops)

	return dims
}
