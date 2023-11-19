// SPDX-License-Identifier: Unlicense OR MIT

/*
Package gesture implements common pointer gestures.

Gestures accept low level pointer Events from an event
Queue and detect higher level actions such as clicks
and scrolling.
*/
package gesture

import (
	"image"
	"math"
	"runtime"
	"time"

	"gioui.org/f32"
	"gioui.org/internal/fling"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/unit"
)

// The duration is somewhat arbitrary.
const doubleClickDuration = 200 * time.Millisecond

// Hover detects the hover gesture for a pointer area.
type Hover struct {
	// entered tracks whether the pointer is inside the gesture.
	entered bool
	// pid is the pointer.ID.
	pid pointer.ID
}

// Add the gesture to detect hovering over the current pointer area.
func (h *Hover) Add(ops *op.Ops) {
	pointer.InputOp{
		Tag:   h,
		Types: pointer.Enter | pointer.Leave,
	}.Add(ops)
}

// Hovered returns whether a pointer is inside the area.
func (h *Hover) Hovered(q event.Queue) bool {
	for _, ev := range q.Events(h) {
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Type {
		case pointer.Leave, pointer.Cancel:
			if h.entered && h.pid == e.PointerID {
				h.entered = false
			}
		case pointer.Enter:
			if !h.entered {
				h.pid = e.PointerID
			}
			if h.pid == e.PointerID {
				h.entered = true
			}
		}
	}
	return h.entered
}

// Click detects click gestures in the form
// of ClickEvents.
type Click struct {
	// clickedAt is the timestamp at which
	// the last click occurred.
	clickedAt time.Duration
	// clicks is incremented if successive clicks
	// are performed within a fixed duration.
	clicks int
	// pressed tracks whether the pointer is pressed.
	pressed bool
	// hovered tracks whether the pointer is inside the gesture.
	hovered bool
	// entered tracks whether an Enter event has been received.
	entered bool
	// pid is the pointer.ID.
	pid pointer.ID
}

// ClickEvent represent a click action, either a
// TypePress for the beginning of a click or a
// TypeClick for a completed click.
type ClickEvent struct {
	Type      ClickType
	Position  image.Point
	Source    pointer.Source
	Modifiers key.Modifiers
	// NumClicks records successive clicks occurring
	// within a short duration of each other.
	NumClicks int
}

type ClickType uint8

// Drag detects drag gestures in the form of pointer.Drag events.
type Drag struct {
	dragging bool
	pressed  bool
	pid      pointer.ID
	start    f32.Point
	grab     bool
}

// Scroll detects scroll gestures and reduces them to
// scroll distances. Scroll recognizes mouse wheel
// movements as well as drag and fling touch gestures.
type Scroll struct {
	dragging  bool
	axis      Axis
	estimator fling.Extrapolation
	flinger   fling.Animation
	pid       pointer.ID
	grab      bool
	last      int
	// Leftover scroll.
	scroll float32
}

type ScrollState uint8

type Axis uint8

const (
	Horizontal Axis = iota
	Vertical
	Both
)

const (
	// TypePress is reported for the first pointer
	// press.
	TypePress ClickType = iota
	// TypeClick is reported when a click action
	// is complete.
	TypeClick
	// TypeCancel is reported when the gesture is
	// cancelled.
	TypeCancel
)

const (
	// StateIdle is the default scroll state.
	StateIdle ScrollState = iota
	// StateDragging is reported during drag gestures.
	StateDragging
	// StateFlinging is reported when a fling is
	// in progress.
	StateFlinging
)

const touchSlop = unit.Dp(3)

// Add the handler to the operation list to receive click events.
func (c *Click) Add(ops *op.Ops) {
	pointer.InputOp{
		Tag:   c,
		Types: pointer.Press | pointer.Release | pointer.Enter | pointer.Leave,
	}.Add(ops)
}

// Hovered returns whether a pointer is inside the area.
func (c *Click) Hovered() bool {
	return c.hovered
}

// Pressed returns whether a pointer is pressing.
func (c *Click) Pressed() bool {
	return c.pressed
}

// Events returns the next click events, if any.
func (c *Click) Events(q event.Queue) []ClickEvent {
	var events []ClickEvent
	for _, evt := range q.Events(c) {
		e, ok := evt.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Type {
		case pointer.Release:
			if !c.pressed || c.pid != e.PointerID {
				break
			}
			c.pressed = false
			if !c.entered || c.hovered {
				events = append(events, ClickEvent{Type: TypeClick, Position: e.Position.Round(), Source: e.Source, Modifiers: e.Modifiers, NumClicks: c.clicks})
			} else {
				events = append(events, ClickEvent{Type: TypeCancel})
			}
		case pointer.Cancel:
			wasPressed := c.pressed
			c.pressed = false
			c.hovered = false
			c.entered = false
			if wasPressed {
				events = append(events, ClickEvent{Type: TypeCancel})
			}
		case pointer.Press:
			if c.pressed {
				break
			}
			if e.Source == pointer.Mouse && e.Buttons != pointer.ButtonPrimary {
				break
			}
			if !c.hovered {
				c.pid = e.PointerID
			}
			if c.pid != e.PointerID {
				break
			}
			c.pressed = true
			if e.Time-c.clickedAt < doubleClickDuration {
				c.clicks++
			} else {
				c.clicks = 1
			}
			c.clickedAt = e.Time
			events = append(events, ClickEvent{Type: TypePress, Position: e.Position.Round(), Source: e.Source, Modifiers: e.Modifiers, NumClicks: c.clicks})
		case pointer.Leave:
			if !c.pressed {
				c.pid = e.PointerID
			}
			if c.pid == e.PointerID {
				c.hovered = false
			}
		case pointer.Enter:
			if !c.pressed {
				c.pid = e.PointerID
			}
			if c.pid == e.PointerID {
				c.hovered = true
				c.entered = true
			}
		}
	}
	return events
}

func (ClickEvent) ImplementsEvent() {}

// Add the handler to the operation list to receive scroll events.
// The bounds variable refers to the scrolling boundaries
// as defined in io/pointer.InputOp.
func (s *Scroll) Add(ops *op.Ops, bounds image.Rectangle) {
	oph := pointer.InputOp{
		Tag:          s,
		Grab:         s.grab,
		Types:        pointer.Press | pointer.Drag | pointer.Release | pointer.Scroll,
		ScrollBounds: bounds,
	}
	oph.Add(ops)
	if s.flinger.Active() {
		op.InvalidateOp{}.Add(ops)
	}
}

// Stop any remaining fling movement.
func (s *Scroll) Stop() {
	s.flinger = fling.Animation{}
}

// Scroll detects the scrolling distance from the available events and
// ongoing fling gestures.
func (s *Scroll) Scroll(cfg unit.Metric, q event.Queue, t time.Time, axis Axis) int {
	if s.axis != axis {
		s.axis = axis
		return 0
	}
	total := 0
	for _, evt := range q.Events(s) {
		e, ok := evt.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Type {
		case pointer.Press:
			if s.dragging {
				break
			}
			// Only scroll on touch drags, or on Android where mice
			// drags also scroll by convention.
			if e.Source != pointer.Touch && runtime.GOOS != "android" {
				break
			}
			s.Stop()
			s.estimator = fling.Extrapolation{}
			v := s.val(e.Position)
			s.last = int(math.Round(float64(v)))
			s.estimator.Sample(e.Time, v)
			s.dragging = true
			s.pid = e.PointerID
		case pointer.Release:
			if s.pid != e.PointerID {
				break
			}
			fling := s.estimator.Estimate()
			if slop, d := float32(cfg.Dp(touchSlop)), fling.Distance; d < -slop || d > slop {
				s.flinger.Start(cfg, t, fling.Velocity)
			}
			fallthrough
		case pointer.Cancel:
			s.dragging = false
			s.grab = false
		case pointer.Scroll:
			switch s.axis {
			case Horizontal:
				s.scroll += e.Scroll.X
			case Vertical:
				s.scroll += e.Scroll.Y
			}
			iscroll := int(s.scroll)
			s.scroll -= float32(iscroll)
			total += iscroll
		case pointer.Drag:
			if !s.dragging || s.pid != e.PointerID {
				continue
			}
			val := s.val(e.Position)
			s.estimator.Sample(e.Time, val)
			v := int(math.Round(float64(val)))
			dist := s.last - v
			if e.Priority < pointer.Grabbed {
				slop := cfg.Dp(touchSlop)
				if dist := dist; dist >= slop || -slop >= dist {
					s.grab = true
				}
			} else {
				s.last = v
				total += dist
			}
		}
	}
	total += s.flinger.Tick(t)
	return total
}

func (s *Scroll) val(p f32.Point) float32 {
	if s.axis == Horizontal {
		return p.X
	} else {
		return p.Y
	}
}

// State reports the scroll state.
func (s *Scroll) State() ScrollState {
	switch {
	case s.flinger.Active():
		return StateFlinging
	case s.dragging:
		return StateDragging
	default:
		return StateIdle
	}
}

// Add the handler to the operation list to receive drag events.
func (d *Drag) Add(ops *op.Ops) {
	pointer.InputOp{
		Tag:   d,
		Grab:  d.grab,
		Types: pointer.Press | pointer.Drag | pointer.Release,
	}.Add(ops)
}

// Events returns the next drag events, if any.
func (d *Drag) Events(cfg unit.Metric, q event.Queue, axis Axis) []pointer.Event {
	var events []pointer.Event
	for _, e := range q.Events(d) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}

		switch e.Type {
		case pointer.Press:
			if !(e.Buttons == pointer.ButtonPrimary || e.Source == pointer.Touch) {
				continue
			}
			d.pressed = true
			if d.dragging {
				continue
			}
			d.dragging = true
			d.pid = e.PointerID
			d.start = e.Position
		case pointer.Drag:
			if !d.dragging || e.PointerID != d.pid {
				continue
			}
			switch axis {
			case Horizontal:
				e.Position.Y = d.start.Y
			case Vertical:
				e.Position.X = d.start.X
			case Both:
				// Do nothing
			}
			if e.Priority < pointer.Grabbed {
				diff := e.Position.Sub(d.start)
				slop := cfg.Dp(touchSlop)
				if diff.X*diff.X+diff.Y*diff.Y > float32(slop*slop) {
					d.grab = true
				}
			}
		case pointer.Release, pointer.Cancel:
			d.pressed = false
			if !d.dragging || e.PointerID != d.pid {
				continue
			}
			d.dragging = false
			d.grab = false
		}

		events = append(events, e)
	}

	return events
}

// Dragging reports whether it is currently in use.
func (d *Drag) Dragging() bool { return d.dragging }

// Pressed returns whether a pointer is pressing.
func (d *Drag) Pressed() bool { return d.pressed }

func (a Axis) String() string {
	switch a {
	case Horizontal:
		return "Horizontal"
	case Vertical:
		return "Vertical"
	default:
		panic("invalid Axis")
	}
}

func (ct ClickType) String() string {
	switch ct {
	case TypePress:
		return "TypePress"
	case TypeClick:
		return "TypeClick"
	case TypeCancel:
		return "TypeCancel"
	default:
		panic("invalid ClickType")
	}
}

func (s ScrollState) String() string {
	switch s {
	case StateIdle:
		return "StateIdle"
	case StateDragging:
		return "StateDragging"
	case StateFlinging:
		return "StateFlinging"
	default:
		panic("unreachable")
	}
}
