// SPDX-License-Identifier: Unlicense OR MIT

/*
Package key implements key and text events and operations.

The InputOp operations is used for declaring key input handlers. Use
an implementation of the Queue interface from package ui to receive
events.
*/
package key

import (
	"fmt"
	"strings"

	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/op"
)

// InputOp declares a handler ready for key events.
// Key events are in general only delivered to the
// focused key handler.
type InputOp struct {
	Tag  event.Tag
	Hint InputHint
}

// SoftKeyboardOp shows or hide the on-screen keyboard, if available.
// It replaces any previous SoftKeyboardOp.
type SoftKeyboardOp struct {
	Show bool
}

// FocusOp sets or clears the keyboard focus. It replaces any previous
// FocusOp in the same frame.
type FocusOp struct {
	// Tag is the new focus. The focus is cleared if Tag is nil, or if Tag
	// has no InputOp in the same frame.
	Tag event.Tag
}

// A FocusEvent is generated when a handler gains or loses
// focus.
type FocusEvent struct {
	Focus bool
}

// An Event is generated when a key is pressed. For text input
// use EditEvent.
type Event struct {
	// Name of the key. For letters, the upper case form is used, via
	// unicode.ToUpper. The shift modifier is taken into account, all other
	// modifiers are ignored. For example, the "shift-1" and "ctrl-shift-1"
	// combinations both give the Name "!" with the US keyboard layout.
	Name string
	// Modifiers is the set of active modifiers when the key was pressed.
	Modifiers Modifiers
	// State is the state of the key when the event was fired.
	State State
}

// An EditEvent is generated when text is input.
type EditEvent struct {
	Text string
}

// InputHint changes the on-screen-keyboard type. That hints the
// type of data that might be entered by the user.
type InputHint uint8

const (
	// HintAny hints that any input is expected.
	HintAny InputHint = iota
	// HintText hints that text input is expected. It may activate auto-correction and suggestions.
	HintText
	// HintNumeric hints that numeric input is expected. It may activate shortcuts for 0-9, "." and ",".
	HintNumeric
	// HintEmail hints that email input is expected. It may activate shortcuts for common email characters, such as "@" and ".com".
	HintEmail
	// HintURL hints that URL input is expected. It may activate shortcuts for common URL fragments such as "/" and ".com".
	HintURL
	// HintTelephone hints that telephone number input is expected. It may activate shortcuts for 0-9, "#" and "*".
	HintTelephone
)

// State is the state of a key during an event.
type State uint8

const (
	// Press is the state of a pressed key.
	Press State = iota
	// Release is the state of a key that has been released.
	//
	// Note: release events are only implemented on the following platforms:
	// macOS, Linux, Windows, WebAssembly.
	Release
)

// Modifiers
type Modifiers uint32

const (
	// ModCtrl is the ctrl modifier key.
	ModCtrl Modifiers = 1 << iota
	// ModCommand is the command modifier key
	// found on Apple keyboards.
	ModCommand
	// ModShift is the shift modifier key.
	ModShift
	// ModAlt is the alt modifier key, or the option
	// key on Apple keyboards.
	ModAlt
	// ModSuper is the "logo" modifier key, often
	// represented by a Windows logo.
	ModSuper
)

const (
	// Names for special keys.
	NameLeftArrow      = "←"
	NameRightArrow     = "→"
	NameUpArrow        = "↑"
	NameDownArrow      = "↓"
	NameReturn         = "⏎"
	NameEnter          = "⌤"
	NameEscape         = "⎋"
	NameHome           = "⇱"
	NameEnd            = "⇲"
	NameDeleteBackward = "⌫"
	NameDeleteForward  = "⌦"
	NamePageUp         = "⇞"
	NamePageDown       = "⇟"
	NameTab            = "⇥"
	NameSpace          = "Space"
)

// Contain reports whether m contains all modifiers
// in m2.
func (m Modifiers) Contain(m2 Modifiers) bool {
	return m&m2 == m2
}

func (h InputOp) Add(o *op.Ops) {
	if h.Tag == nil {
		panic("Tag must be non-nil")
	}
	data := ops.Write1(&o.Internal, ops.TypeKeyInputLen, h.Tag)
	data[0] = byte(ops.TypeKeyInput)
	data[1] = byte(h.Hint)
}

func (h SoftKeyboardOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeKeySoftKeyboardLen)
	data[0] = byte(ops.TypeKeySoftKeyboard)
	if h.Show {
		data[1] = 1
	}
}

func (h FocusOp) Add(o *op.Ops) {
	data := ops.Write1(&o.Internal, ops.TypeKeyFocusLen, h.Tag)
	data[0] = byte(ops.TypeKeyFocus)
}

func (EditEvent) ImplementsEvent()  {}
func (Event) ImplementsEvent()      {}
func (FocusEvent) ImplementsEvent() {}

func (e Event) String() string {
	return fmt.Sprintf("%v %v %v}", e.Name, e.Modifiers, e.State)
}

func (m Modifiers) String() string {
	var strs []string
	if m.Contain(ModCtrl) {
		strs = append(strs, "ModCtrl")
	}
	if m.Contain(ModCommand) {
		strs = append(strs, "ModCommand")
	}
	if m.Contain(ModShift) {
		strs = append(strs, "ModShift")
	}
	if m.Contain(ModAlt) {
		strs = append(strs, "ModAlt")
	}
	if m.Contain(ModSuper) {
		strs = append(strs, "ModSuper")
	}
	return strings.Join(strs, "|")
}

func (s State) String() string {
	switch s {
	case Press:
		return "Press"
	case Release:
		return "Release"
	default:
		panic("invalid State")
	}
}
