// SPDX-License-Identifier: Unlicense OR MIT

// Package profiles provides access to rendering
// profiles.
package profile

import (
	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/op"
)

// Op registers a handler for receiving
// Events.
type Op struct {
	Tag event.Tag
}

// Event contains profile data from a single
// rendered frame.
type Event struct {
	// Timings. Very likely to change.
	Timings string
}

func (p Op) Add(o *op.Ops) {
	data := ops.Write1(&o.Internal, ops.TypeProfileLen, p.Tag)
	data[0] = byte(ops.TypeProfile)
}

func (p Event) ImplementsEvent() {}
