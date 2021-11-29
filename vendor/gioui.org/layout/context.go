// SPDX-License-Identifier: Unlicense OR MIT

package layout

import (
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/unit"
)

// Context carries the state needed by almost all layouts and widgets.
// A zero value Context never returns events, map units to pixels
// with a scale of 1.0, and returns the zero time from Now.
type Context struct {
	// Constraints track the constraints for the active widget or
	// layout.
	Constraints Constraints

	Metric unit.Metric
	// By convention, a nil Queue is a signal to widgets to draw themselves
	// in a disabled state.
	Queue event.Queue
	// Now is the animation time.
	Now time.Time

	*op.Ops
}

// NewContext is a shorthand for
//
//   Context{
//     Ops: ops,
//     Now: e.Now,
//     Queue: e.Queue,
//     Config: e.Config,
//     Constraints: Exact(e.Size),
//   }
//
// NewContext calls ops.Reset and adjusts ops for e.Insets.
func NewContext(ops *op.Ops, e system.FrameEvent) Context {
	ops.Reset()

	size := e.Size

	if e.Insets != (system.Insets{}) {
		left := e.Metric.Px(e.Insets.Left)
		top := e.Metric.Px(e.Insets.Top)
		op.Offset(f32.Point{
			X: float32(left),
			Y: float32(top),
		}).Add(ops)

		size.X -= left + e.Metric.Px(e.Insets.Right)
		size.Y -= top + e.Metric.Px(e.Insets.Bottom)
	}

	return Context{
		Ops:         ops,
		Now:         e.Now,
		Queue:       e.Queue,
		Metric:      e.Metric,
		Constraints: Exact(size),
	}
}

// Px maps the value to pixels.
func (c Context) Px(v unit.Value) int {
	return c.Metric.Px(v)
}

// Events returns the events available for the key. If no
// queue is configured, Events returns nil.
func (c Context) Events(k event.Tag) []event.Event {
	if c.Queue == nil {
		return nil
	}
	return c.Queue.Events(k)
}

// Disabled returns a copy of this context with a nil Queue,
// blocking events to widgets using it.
//
// By convention, a nil Queue is a signal to widgets to draw themselves
// in a disabled state.
func (c Context) Disabled() Context {
	c.Queue = nil
	return c
}
