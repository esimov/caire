package widget

import (
	"fmt"
	"math/bits"

	"gioui.org/gesture"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

// Decorations handles the states of window decorations.
type Decorations struct {
	clicks []Clickable
	resize [8]struct {
		gesture.Hover
		gesture.Drag
	}
	actions   system.Action
	maximized bool
}

// LayoutMove lays out the widget that makes a window movable.
func (d *Decorations) LayoutMove(gtx layout.Context, w layout.Widget) layout.Dimensions {
	dims := w(gtx)
	defer clip.Rect{Max: dims.Size}.Push(gtx.Ops).Pop()
	system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
	return dims
}

// Clickable returns the clickable for the given single action.
func (d *Decorations) Clickable(action system.Action) *Clickable {
	if bits.OnesCount(uint(action)) != 1 {
		panic(fmt.Errorf("not a single action"))
	}
	idx := bits.TrailingZeros(uint(action))
	if n := idx - len(d.clicks); n >= 0 {
		d.clicks = append(d.clicks, make([]Clickable, n+1)...)
	}
	click := &d.clicks[idx]
	if click.Clicked() {
		if action == system.ActionMaximize {
			if d.maximized {
				d.maximized = false
				d.actions |= system.ActionUnmaximize
			} else {
				d.maximized = true
				d.actions |= system.ActionMaximize
			}
		} else {
			d.actions |= action
		}
	}
	return click
}

// Perform updates the decorations as if the specified actions were
// performed by the user.
func (d *Decorations) Perform(actions system.Action) {
	if actions&system.ActionMaximize != 0 {
		d.maximized = true
	}
	if actions&(system.ActionUnmaximize|system.ActionMinimize|system.ActionFullscreen) != 0 {
		d.maximized = false
	}
}

// Actions returns the set of actions activated by the user.
func (d *Decorations) Actions() system.Action {
	a := d.actions
	d.actions = 0
	return a
}

// Maximized returns whether the window is maximized.
func (d *Decorations) Maximized() bool {
	return d.maximized
}
