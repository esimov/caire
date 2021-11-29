// SPDX-License-Identifier: Unlicense OR MIT

package pointer

import (
	"encoding/binary"
	"fmt"
	"image"
	"strings"
	"time"

	"gioui.org/f32"
	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// Event is a pointer event.
type Event struct {
	Type   Type
	Source Source
	// PointerID is the id for the pointer and can be used
	// to track a particular pointer from Press to
	// Release or Cancel.
	PointerID ID
	// Priority is the priority of the receiving handler
	// for this event.
	Priority Priority
	// Time is when the event was received. The
	// timestamp is relative to an undefined base.
	Time time.Duration
	// Buttons are the set of pressed mouse buttons for this event.
	Buttons Buttons
	// Position is the position of the event, relative to
	// the current transformation, as set by op.TransformOp.
	Position f32.Point
	// Scroll is the scroll amount, if any.
	Scroll f32.Point
	// Modifiers is the set of active modifiers when
	// the mouse button was pressed.
	Modifiers key.Modifiers
}

// PassOp sets the pass-through mode. InputOps added while the pass-through
// mode is set don't block events to siblings.
type PassOp struct {
}

// PassStack represents a PassOp on the pass stack.
type PassStack struct {
	ops     *ops.Ops
	id      ops.StackID
	macroID int
}

// CursorNameOp sets the cursor for the current area.
type CursorNameOp struct {
	Name CursorName
}

// InputOp declares an input handler ready for pointer
// events.
type InputOp struct {
	Tag event.Tag
	// Grab, if set, request that the handler get
	// Grabbed priority.
	Grab bool
	// Types is a bitwise-or of event types to receive.
	Types Type
	// ScrollBounds describe the maximum scrollable distances in both
	// axes. Specifically, any Event e delivered to Tag will satisfy
	//
	// ScrollBounds.Min.X <= e.Scroll.X <= ScrollBounds.Max.X (horizontal axis)
	// ScrollBounds.Min.Y <= e.Scroll.Y <= ScrollBounds.Max.Y (vertical axis)
	ScrollBounds image.Rectangle
}

type ID uint16

// Type of an Event.
type Type uint

// Priority of an Event.
type Priority uint8

// Source of an Event.
type Source uint8

// Buttons is a set of mouse buttons
type Buttons uint8

// CursorName is the name of a cursor.
type CursorName string

const (
	// CursorDefault is the default cursor.
	CursorDefault CursorName = ""
	// CursorText is the cursor for text.
	CursorText CursorName = "text"
	// CursorPointer is the cursor for a link.
	CursorPointer CursorName = "pointer"
	// CursorCrossHair is the cursor for precise location.
	CursorCrossHair CursorName = "crosshair"
	// CursorColResize is the cursor for vertical resize.
	CursorColResize CursorName = "col-resize"
	// CursorRowResize is the cursor for horizontal resize.
	CursorRowResize CursorName = "row-resize"
	// CursorGrab is the cursor for moving object in any direction.
	CursorGrab CursorName = "grab"
	// CursorNone hides the cursor. To show it again, use any other cursor.
	CursorNone CursorName = "none"
)

const (
	// A Cancel event is generated when the current gesture is
	// interrupted by other handlers or the system.
	Cancel Type = (1 << iota) >> 1
	// Press of a pointer.
	Press
	// Release of a pointer.
	Release
	// Move of a pointer.
	Move
	// Drag of a pointer.
	Drag
	// Pointer enters an area watching for pointer input
	Enter
	// Pointer leaves an area watching for pointer input
	Leave
	// Scroll of a pointer.
	Scroll
)

const (
	// Mouse generated event.
	Mouse Source = iota
	// Touch generated event.
	Touch
)

const (
	// Shared priority is for handlers that
	// are part of a matching set larger than 1.
	Shared Priority = iota
	// Foremost priority is like Shared, but the
	// handler is the foremost of the matching set.
	Foremost
	// Grabbed is used for matching sets of size 1.
	Grabbed
)

const (
	// ButtonPrimary is the primary button, usually the left button for a
	// right-handed user.
	ButtonPrimary Buttons = 1 << iota
	// ButtonSecondary is the secondary button, usually the right button for a
	// right-handed user.
	ButtonSecondary
	// ButtonTertiary is the tertiary button, usually the middle button.
	ButtonTertiary
)

// Rect constructs a rectangular hit area.
//
// Deprecated: use clip.Rect instead.
func Rect(size image.Rectangle) clip.Op {
	return clip.Rect(size).Op()
}

// Ellipse constructs an ellipsoid hit area.
//
// Deprecated: use clip.Ellipse instead.
func Ellipse(size image.Rectangle) clip.Ellipse {
	return clip.Ellipse(frect(size))
}

// frect converts a rectangle to a f32.Rectangle.
func frect(r image.Rectangle) f32.Rectangle {
	return f32.Rectangle{
		Min: fpt(r.Min), Max: fpt(r.Max),
	}
}

// fpt converts an point to a f32.Point.
func fpt(p image.Point) f32.Point {
	return f32.Point{
		X: float32(p.X), Y: float32(p.Y),
	}
}

// Push the current pass mode to the pass stack and set the pass mode.
func (p PassOp) Push(o *op.Ops) PassStack {
	id, mid := ops.PushOp(&o.Internal, ops.PassStack)
	data := ops.Write(&o.Internal, ops.TypePassLen)
	data[0] = byte(ops.TypePass)
	return PassStack{ops: &o.Internal, id: id, macroID: mid}
}

func (p PassStack) Pop() {
	ops.PopOp(p.ops, ops.PassStack, p.id, p.macroID)
	data := ops.Write(p.ops, ops.TypePopPassLen)
	data[0] = byte(ops.TypePopPass)
}

func (op CursorNameOp) Add(o *op.Ops) {
	data := ops.Write1(&o.Internal, ops.TypeCursorLen, op.Name)
	data[0] = byte(ops.TypeCursor)
}

// Add panics if the scroll range does not contain zero.
func (op InputOp) Add(o *op.Ops) {
	if op.Tag == nil {
		panic("Tag must be non-nil")
	}
	if b := op.ScrollBounds; b.Min.X > 0 || b.Max.X < 0 || b.Min.Y > 0 || b.Max.Y < 0 {
		panic(fmt.Errorf("invalid scroll range value %v", b))
	}
	if op.Types>>16 > 0 {
		panic(fmt.Errorf("value in Types overflows uint16"))
	}
	data := ops.Write1(&o.Internal, ops.TypePointerInputLen, op.Tag)
	data[0] = byte(ops.TypePointerInput)
	if op.Grab {
		data[1] = 1
	}
	bo := binary.LittleEndian
	bo.PutUint16(data[2:], uint16(op.Types))
	bo.PutUint32(data[4:], uint32(op.ScrollBounds.Min.X))
	bo.PutUint32(data[8:], uint32(op.ScrollBounds.Min.Y))
	bo.PutUint32(data[12:], uint32(op.ScrollBounds.Max.X))
	bo.PutUint32(data[16:], uint32(op.ScrollBounds.Max.Y))
}

func (t Type) String() string {
	if t == Cancel {
		return "Cancel"
	}
	var buf strings.Builder
	for tt := Type(1); tt > 0; tt <<= 1 {
		if t&tt > 0 {
			if buf.Len() > 0 {
				buf.WriteByte('|')
			}
			buf.WriteString((t & tt).string())
		}
	}
	return buf.String()
}

func (t Type) string() string {
	switch t {
	case Press:
		return "Press"
	case Release:
		return "Release"
	case Cancel:
		return "Cancel"
	case Move:
		return "Move"
	case Drag:
		return "Drag"
	case Enter:
		return "Enter"
	case Leave:
		return "Leave"
	case Scroll:
		return "Scroll"
	default:
		panic("unknown Type")
	}
}

func (p Priority) String() string {
	switch p {
	case Shared:
		return "Shared"
	case Foremost:
		return "Foremost"
	case Grabbed:
		return "Grabbed"
	default:
		panic("unknown priority")
	}
}

func (s Source) String() string {
	switch s {
	case Mouse:
		return "Mouse"
	case Touch:
		return "Touch"
	default:
		panic("unknown source")
	}
}

// Contain reports whether the set b contains
// all of the buttons.
func (b Buttons) Contain(buttons Buttons) bool {
	return b&buttons == buttons
}

func (b Buttons) String() string {
	var strs []string
	if b.Contain(ButtonPrimary) {
		strs = append(strs, "ButtonPrimary")
	}
	if b.Contain(ButtonSecondary) {
		strs = append(strs, "ButtonSecondary")
	}
	if b.Contain(ButtonTertiary) {
		strs = append(strs, "ButtonTertiary")
	}
	return strings.Join(strs, "|")
}

func (c CursorName) String() string {
	if c == CursorDefault {
		return "default"
	}
	return string(c)
}

func (Event) ImplementsEvent() {}
