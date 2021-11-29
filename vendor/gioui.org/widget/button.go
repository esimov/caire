// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/layout"
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
}

// Click represents a click.
type Click struct {
	Modifiers key.Modifiers
	NumClicks int
}

// Press represents a past pointer press.
type Press struct {
	// Position of the press.
	Position f32.Point
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

// Hovered returns whether pointer is over the element.
func (b *Clickable) Hovered() bool {
	return b.click.Hovered()
}

// Pressed returns whether pointer is pressing the element.
func (b *Clickable) Pressed() bool {
	return b.click.Pressed()
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
func (b *Clickable) Layout(gtx layout.Context) layout.Dimensions {
	b.update(gtx)
	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Push(gtx.Ops).Pop()
	b.click.Add(gtx.Ops)
	for len(b.history) > 0 {
		c := b.history[0]
		if c.End.IsZero() || gtx.Now.Sub(c.End) < 1*time.Second {
			break
		}
		n := copy(b.history, b.history[1:])
		b.history = b.history[:n]
	}
	return layout.Dimensions{Size: gtx.Constraints.Min}
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
			b.history = append(b.history, Press{
				Position: e.Position,
				Start:    gtx.Now,
			})
		}
	}
}
