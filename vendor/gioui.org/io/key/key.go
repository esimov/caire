// SPDX-License-Identifier: Unlicense OR MIT

/*
Package key implements key and text events and operations.

The InputOp operations is used for declaring key input handlers. Use
an implementation of the Queue interface from package ui to receive
events.
*/
package key

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"gioui.org/f32"
	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/op"
)

// InputOp declares a handler ready for key events.
// Key events are in general only delivered to the
// focused key handler.
type InputOp struct {
	Tag event.Tag
	// Hint describes the type of text expected by Tag.
	Hint InputHint
	// Keys is the set of keys Tag can handle. That is, Tag will only
	// receive an Event if its key and modifiers are accepted by Keys.Contains.
	// As a special case, the topmost (first added) InputOp handler receives all
	// unhandled events.
	Keys Set
}

// Set is an expression that describes a set of key combinations, in the form
// "<modifiers>-<keyset>|...".  Modifiers are separated by dashes, optional
// modifiers are enclosed by parentheses.  A key set is either a literal key
// name or a list of key names separated by commas and enclosed in brackets.
//
// The "Short" modifier matches the shortcut modifier (ModShortcut) and
// "ShortAlt" matches the alternative modifier (ModShortcutAlt).
//
// Examples:
//
//   - A|B matches the A and B keys
//   - [A,B] also matches the A and B keys
//   - Shift-A matches A key if shift is pressed, and no other modifier.
//   - Shift-(Ctrl)-A matches A if shift is pressed, and optionally ctrl.
type Set string

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

// SelectionOp updates the selection for an input handler.
type SelectionOp struct {
	Tag event.Tag
	Range
	Caret
}

// SnippetOp updates the content snippet for an input handler.
type SnippetOp struct {
	Tag event.Tag
	Snippet
}

// Range represents a range of text, such as an editor's selection.
// Start and End are in runes.
type Range struct {
	Start int
	End   int
}

// Snippet represents a snippet of text content used for communicating between
// an editor and an input method.
type Snippet struct {
	Range
	Text string
}

// Caret represents the position of a caret.
type Caret struct {
	// Pos is the intersection point of the caret and its baseline.
	Pos f32.Point
	// Ascent is the length of the caret above its baseline.
	Ascent float32
	// Descent is the length of the caret below its baseline.
	Descent float32
}

// SelectionEvent is generated when an input method changes the selection.
type SelectionEvent Range

// SnippetEvent is generated when the snippet range is updated by an
// input method.
type SnippetEvent Range

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

// An EditEvent requests an edit by an input method.
type EditEvent struct {
	// Range specifies the range to replace with Text.
	Range Range
	Text  string
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
	// HintPassword hints that password input is expected. It may disable autocorrection and enable password autofill.
	HintPassword
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
	NameTab            = "Tab"
	NameSpace          = "Space"
	NameCtrl           = "Ctrl"
	NameShift          = "Shift"
	NameAlt            = "Alt"
	NameSuper          = "Super"
	NameCommand        = "⌘"
	NameF1             = "F1"
	NameF2             = "F2"
	NameF3             = "F3"
	NameF4             = "F4"
	NameF5             = "F5"
	NameF6             = "F6"
	NameF7             = "F7"
	NameF8             = "F8"
	NameF9             = "F9"
	NameF10            = "F10"
	NameF11            = "F11"
	NameF12            = "F12"
	NameBack           = "Back"
)

// Contain reports whether m contains all modifiers
// in m2.
func (m Modifiers) Contain(m2 Modifiers) bool {
	return m&m2 == m2
}

func (k Set) Contains(name string, mods Modifiers) bool {
	ks := string(k)
	for len(ks) > 0 {
		// Cut next key expression.
		chord, rest, _ := cut(ks, "|")
		ks = rest
		// Separate key set and modifier set.
		var modSet, keySet string
		sep := strings.LastIndex(chord, "-")
		if sep != -1 {
			modSet, keySet = chord[:sep], chord[sep+1:]
		} else {
			modSet, keySet = "", chord
		}
		if !keySetContains(keySet, name) {
			continue
		}
		if modSetContains(modSet, mods) {
			return true
		}
	}
	return false
}

func keySetContains(keySet, name string) bool {
	// Check for single key match.
	if keySet == name {
		return true
	}
	// Check for set match.
	if len(keySet) < 2 || keySet[0] != '[' || keySet[len(keySet)-1] != ']' {
		return false
	}
	keySet = keySet[1 : len(keySet)-1]
	for len(keySet) > 0 {
		key, rest, _ := cut(keySet, ",")
		keySet = rest
		if key == name {
			return true
		}
	}
	return false
}

func modSetContains(modSet string, mods Modifiers) bool {
	var smods Modifiers
	for len(modSet) > 0 {
		mod, rest, _ := cut(modSet, "-")
		modSet = rest
		if len(mod) >= 2 && mod[0] == '(' && mod[len(mod)-1] == ')' {
			mods &^= modFor(mod[1 : len(mod)-1])
		} else {
			smods |= modFor(mod)
		}
	}
	return mods == smods
}

// cut is a copy of the standard library strings.Cut.
// TODO: remove when Go 1.18 is our minimum.
func cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

func modFor(name string) Modifiers {
	switch name {
	case NameCtrl:
		return ModCtrl
	case NameShift:
		return ModShift
	case NameAlt:
		return ModAlt
	case NameSuper:
		return ModSuper
	case NameCommand:
		return ModCommand
	case "Short":
		return ModShortcut
	case "ShortAlt":
		return ModShortcutAlt
	}
	return 0
}

func (h InputOp) Add(o *op.Ops) {
	if h.Tag == nil {
		panic("Tag must be non-nil")
	}
	data := ops.Write2String(&o.Internal, ops.TypeKeyInputLen, h.Tag, string(h.Keys))
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

func (s SnippetOp) Add(o *op.Ops) {
	data := ops.Write2String(&o.Internal, ops.TypeSnippetLen, s.Tag, s.Text)
	data[0] = byte(ops.TypeSnippet)
	bo := binary.LittleEndian
	bo.PutUint32(data[1:], uint32(s.Range.Start))
	bo.PutUint32(data[5:], uint32(s.Range.End))
}

func (s SelectionOp) Add(o *op.Ops) {
	data := ops.Write1(&o.Internal, ops.TypeSelectionLen, s.Tag)
	data[0] = byte(ops.TypeSelection)
	bo := binary.LittleEndian
	bo.PutUint32(data[1:], uint32(s.Start))
	bo.PutUint32(data[5:], uint32(s.End))
	bo.PutUint32(data[9:], math.Float32bits(s.Pos.X))
	bo.PutUint32(data[13:], math.Float32bits(s.Pos.Y))
	bo.PutUint32(data[17:], math.Float32bits(s.Ascent))
	bo.PutUint32(data[21:], math.Float32bits(s.Descent))
}

func (EditEvent) ImplementsEvent()      {}
func (Event) ImplementsEvent()          {}
func (FocusEvent) ImplementsEvent()     {}
func (SnippetEvent) ImplementsEvent()   {}
func (SelectionEvent) ImplementsEvent() {}

func (e Event) String() string {
	return fmt.Sprintf("%v %v %v}", e.Name, e.Modifiers, e.State)
}

func (m Modifiers) String() string {
	var strs []string
	if m.Contain(ModCtrl) {
		strs = append(strs, NameCtrl)
	}
	if m.Contain(ModCommand) {
		strs = append(strs, NameCommand)
	}
	if m.Contain(ModShift) {
		strs = append(strs, NameShift)
	}
	if m.Contain(ModAlt) {
		strs = append(strs, NameAlt)
	}
	if m.Contain(ModSuper) {
		strs = append(strs, NameSuper)
	}
	return strings.Join(strs, "-")
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
