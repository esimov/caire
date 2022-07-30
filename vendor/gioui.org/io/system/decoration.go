package system

import (
	"strings"

	"gioui.org/internal/ops"
	"gioui.org/op"
)

// ActionAreaOp makes the current clip area available for
// system gestures.
//
// Note: only ActionMove is supported.
type ActionInputOp Action

// Action is a set of window decoration actions.
type Action uint

const (
	// ActionMinimize minimizes a window.
	ActionMinimize Action = 1 << iota
	// ActionMaximize maximizes a window.
	ActionMaximize
	// ActionUnmaximize restores a maximized window.
	ActionUnmaximize
	// ActionFullscreen makes a window fullscreen.
	ActionFullscreen
	// ActionRaise requests that the platform bring this window to the top of all open windows.
	// Some platforms do not allow this except under certain circumstances, such as when
	// a window from the same application already has focus. If the platform does not
	// support it, this method will do nothing.
	ActionRaise
	// ActionCenter centers the window on the screen.
	// It is ignored in Fullscreen mode and on Wayland.
	ActionCenter
	// ActionClose closes a window.
	// Only applicable on macOS, Windows, X11 and Wayland.
	ActionClose
	// ActionMove moves a window directed by the user.
	ActionMove
)

func (op ActionInputOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeActionInputLen)
	data[0] = byte(ops.TypeActionInput)
	data[1] = byte(op)
}

func (a Action) String() string {
	var buf strings.Builder
	for b := Action(1); a != 0; b <<= 1 {
		if a&b != 0 {
			if buf.Len() > 0 {
				buf.WriteByte('|')
			}
			buf.WriteString(b.string())
			a &^= b
		}
	}
	return buf.String()
}

func (a Action) string() string {
	switch a {
	case ActionMinimize:
		return "ActionMinimize"
	case ActionMaximize:
		return "ActionMaximize"
	case ActionUnmaximize:
		return "ActionUnmaximize"
	case ActionClose:
		return "ActionClose"
	case ActionMove:
		return "ActionMove"
	}
	return ""
}
