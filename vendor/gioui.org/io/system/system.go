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
	// Insets represent the space occupied by system decorations and controls.
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
	// Values are in pixels.
	Top, Bottom, Left, Right unit.Dp
}

// A StageEvent is generated whenever the stage of a
// Window changes.
type StageEvent struct {
	Stage Stage
}

// Stage of a Window.
type Stage uint8

const (
	// StagePaused is the stage for windows that have no on-screen representation.
	// Paused windows don't receive FrameEvent.
	StagePaused Stage = iota
	// StageInactive is the stage for windows that are visible, but not active.
	// Inactive windows receive FrameEvent.
	StageInactive
	// StageRunning is for active and visible Windows.
	// Running windows receive FrameEvent.
	StageRunning
)

// String implements fmt.Stringer.
func (l Stage) String() string {
	switch l {
	case StagePaused:
		return "StagePaused"
	case StageInactive:
		return "StageInactive"
	case StageRunning:
		return "StageRunning"
	default:
		panic("unexpected Stage value")
	}
}

func (FrameEvent) ImplementsEvent()   {}
func (StageEvent) ImplementsEvent()   {}
func (DestroyEvent) ImplementsEvent() {}
