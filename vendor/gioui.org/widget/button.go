// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"
	"time"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// Clickable represents a clickable area.
type Clickable struct {
	click  gesture.Click
	clicks []Click
	// prevClicks is the index into clicks that marks the clicks
	// from the most recent Layout call. prevClicks is used to keep
	// clicks bounded.
	prevClicks int
	history    []Press

	keyTag  struct{}
	focused bool
}

// Click represents a click.
type Click struct {
	Modifiers key.Modifiers
	NumClicks int
}

// Press represents a past pointer press.
type Press struct {
	// Position of the press.
	Position image.Point
	// Start is when the press began.
	Start time.Time
	// End is when the press was ended by a release or cancel.
	// A zero End means it hasn't ended yet.
	End time.Time
	// Cancelled is true for cancelled presses.
	Cancelled bool
}

// Click executes a simple programmatic click
func (b *Clickable) Click() {
	b.clicks = append(b.clicks, Click{
		Modifiers: 0,
		NumClicks: 1,
	})
}

// Clicked reports whether there are pending clicks as would be
// reported by Clicks. If so, Clicked removes the earliest click.
func (b *Clickable) Clicked() bool {
	if len(b.clicks) == 0 {
		return false
	}
	n := copy(b.clicks, b.clicks[1:])
	b.clicks = b.clicks[:n]
	if b.prevClicks > 0 {
		b.prevClicks--
	}
	return true
}

// Hovered reports whether a pointer is over the element.
func (b *Clickable) Hovered() bool {
	return b.click.Hovered()
}

// Pressed reports whether a pointer is pressing the element.
func (b *Clickable) Pressed() bool {
	return b.click.Pressed()
}

// Focused reports whether b has focus.
func (b *Clickable) Focused() bool {
	return b.focused
}

// Clicks returns and clear the clicks since the last call to Clicks.
func (b *Clickable) Clicks() []Click {
	clicks := b.clicks
	b.clicks = nil
	b.prevClicks = 0
	return clicks
}

// History is the past pointer presses useful for drawing markers.
// History is retained for a short duration (about a second).
func (b *Clickable) History() []Press {
	return b.history
}

// Layout and update the button state
func (b *Clickable) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	b.update(gtx)
	m := op.Record(gtx.Ops)
	dims := w(gtx)
	c := m.Stop()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	disabled := gtx.Queue == nil
	semantic.DisabledOp(disabled).Add(gtx.Ops)
	b.click.Add(gtx.Ops)
	if !disabled {
		key.InputOp{Tag: &b.keyTag, Keys: "⏎|Space"}.Add(gtx.Ops)
	} else {
		b.focused = false
	}
	c.Add(gtx.Ops)
	for len(b.history) > 0 {
		c := b.history[0]
		if c.End.IsZero() || gtx.Now.Sub(c.End) < 1*time.Second {
			break
		}
		n := copy(b.history, b.history[1:])
		b.history = b.history[:n]
	}
	return dims
}

// update the button state by processing events.
func (b *Clickable) update(gtx layout.Context) {
	// Flush clicks from before the last update.
	n := copy(b.clicks, b.clicks[b.prevClicks:])
	b.clicks = b.clicks[:n]
	b.prevClicks = n

	for _, e := range b.click.Events(gtx) {
		switch e.Type {
		case gesture.TypeClick:
			b.clicks = append(b.clicks, Click{
				Modifiers: e.Modifiers,
				NumClicks: e.NumClicks,
			})
			if l := len(b.history); l > 0 {
				b.history[l-1].End = gtx.Now
			}
		case gesture.TypeCancel:
			for i := range b.history {
				b.history[i].Cancelled = true
				if b.history[i].End.IsZero() {
					b.history[i].End = gtx.Now
				}
			}
		case gesture.TypePress:
			if e.Source == pointer.Mouse {
				key.FocusOp{Tag: &b.keyTag}.Add(gtx.Ops)
			}
			b.history = append(b.history, Press{
				Position: e.Position,
				Start:    gtx.Now,
			})
		}
	}
	for _, e := range gtx.Events(&b.keyTag) {
		switch e := e.(type) {
		case key.FocusEvent:
			b.focused = e.Focus
		case key.Event:
			if !b.focused || e.State != key.Release {
				break
			}
			if e.Name != key.NameReturn && e.Name != key.NameSpace {
				break
			}
			b.clicks = append(b.clicks, Click{
				Modifiers: e.Modifiers,
				NumClicks: 1,
			})
		}
	}
}
