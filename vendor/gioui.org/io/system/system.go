// SPDX-License-Identifier: Unlicense OR MIT

// Package system contains events usually handled at the top-level
// program level.
package system

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/op"
	"gioui.org/unit"
)

// A FrameEvent requests a new frame in the form of a list of
// operations that describes what to display and how to handle
// input.
type FrameEvent struct {
	// Now is the current animation. Use Now instead of time.Now to
	// synchronize animation and to avoid the time.Now call overhead.
	Now time.Time
	// Metric converts device independent dp and sp to device pixels.
	Metric unit.Metric
	// Size is the dimensions of the window.
	Size image.Point
	// Insets is the insets to apply.
	Insets Insets
	// Frame completes the FrameEvent by drawing the graphical operations
	// from ops into the window.
	Frame func(frame *op.Ops)
	// Queue supplies the events for event handlers.
	Queue event.Queue
}

// DestroyEvent is the last event sent through
// a window event channel.
type DestroyEvent struct {
	// Err is nil for normal window closures. If a
	// window is prematurely closed, Err is the cause.
	Err error
}

// Insets is the space taken up by
// system decoration such as translucent
// system bars and software keyboards.
type Insets struct {
	Top, Bottom, Left, Right unit.Value
}

// A StageEvent is generated whenever the stage of a
// Window changes.
type StageEvent struct {
	Stage Stage
}

// CommandEvent is a system event. Unlike most other events, CommandEvent is
// delivered as a pointer to allow Cancel to suppress it.
type CommandEvent struct {
	Type CommandType
	// Cancel suppress the default action of the command.
	Cancel bool
}

// Stage of a Window.
type Stage uint8

// CommandType is the type of a CommandEvent.
type CommandType uint8

const (
	// StagePaused is the Stage for inactive Windows.
	// Inactive Windows don't receive FrameEvents.
	StagePaused Stage = iota
	// StateRunning is for active Windows.
	StageRunning
)

const (
	// CommandBack is the command for a back action
	// such as the Android back button.
	CommandBack CommandType = iota
)

func (l Stage) String() string {
	switch l {
	case StagePaused:
		return "StagePaused"
	case StageRunning:
		return "StageRunning"
	default:
		panic("unexpected Stage value")
	}
}

func (FrameEvent) ImplementsEvent()    {}
func (StageEvent) ImplementsEvent()    {}
func (*CommandEvent) ImplementsEvent() {}
func (DestroyEvent) ImplementsEvent()  {}
