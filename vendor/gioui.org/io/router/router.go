// SPDX-License-Identifier: Unlicense OR MIT

/*
Package router implements Router, a event.Queue implementation
that that disambiguates and routes events to handlers declared
in operation lists.

Router is used by app.Window and is otherwise only useful for
using Gio with external window implementations.
*/
package router

import (
	"encoding/binary"
	"image"
	"time"

	"gioui.org/internal/ops"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/profile"
	"gioui.org/op"
)

// Router is a Queue implementation that routes events
// to handlers declared in operation lists.
type Router struct {
	pointer struct {
		queue     pointerQueue
		collector pointerCollector
	}
	key struct {
		queue     keyQueue
		collector keyCollector
	}
	cqueue clipboardQueue

	handlers handlerEvents

	reader ops.Reader

	// InvalidateOp summary.
	wakeup     bool
	wakeupTime time.Time

	// ProfileOp summary.
	profHandlers map[event.Tag]struct{}
	profile      profile.Event
}

type handlerEvents struct {
	handlers  map[event.Tag][]event.Event
	hadEvents bool
}

// Events returns the available events for the handler key.
func (q *Router) Events(k event.Tag) []event.Event {
	events := q.handlers.Events(k)
	if _, isprof := q.profHandlers[k]; isprof {
		delete(q.profHandlers, k)
		events = append(events, q.profile)
	}
	return events
}

// Frame replaces the declared handlers from the supplied
// operation list. The text input state, wakeup time and whether
// there are active profile handlers is also saved.
func (q *Router) Frame(frame *op.Ops) {
	q.handlers.Clear()
	q.wakeup = false
	for k := range q.profHandlers {
		delete(q.profHandlers, k)
	}
	var ops *ops.Ops
	if frame != nil {
		ops = &frame.Internal
	}
	q.reader.Reset(ops)
	q.collect()

	q.pointer.queue.Frame(&q.handlers)
	q.key.queue.Frame(&q.handlers, q.key.collector)
	if q.handlers.HadEvents() {
		q.wakeup = true
		q.wakeupTime = time.Time{}
	}
}

// Queue an event and report whether at least one handler had an event queued.
func (q *Router) Queue(events ...event.Event) bool {
	for _, e := range events {
		switch e := e.(type) {
		case profile.Event:
			q.profile = e
		case pointer.Event:
			q.pointer.queue.Push(e, &q.handlers)
		case key.EditEvent, key.Event, key.FocusEvent:
			q.key.queue.Push(e, &q.handlers)
		case clipboard.Event:
			q.cqueue.Push(e, &q.handlers)
		}
	}
	return q.handlers.HadEvents()
}

// TextInputState returns the input state from the most recent
// call to Frame.
func (q *Router) TextInputState() TextInputState {
	return q.key.queue.InputState()
}

// TextInputHint returns the input mode from the most recent key.InputOp.
func (q *Router) TextInputHint() (key.InputHint, bool) {
	return q.key.queue.InputHint()
}

// WriteClipboard returns the most recent text to be copied
// to the clipboard, if any.
func (q *Router) WriteClipboard() (string, bool) {
	return q.cqueue.WriteClipboard()
}

// ReadClipboard reports if any new handler is waiting
// to read the clipboard.
func (q *Router) ReadClipboard() bool {
	return q.cqueue.ReadClipboard()
}

// Cursor returns the last cursor set.
func (q *Router) Cursor() pointer.CursorName {
	return q.pointer.queue.cursor
}

func (q *Router) collect() {
	pc := &q.pointer.collector
	pc.reset(&q.pointer.queue)
	kc := &q.key.collector
	*kc = keyCollector{q: &q.key.queue}
	q.key.queue.Reset()
	for encOp, ok := q.reader.Decode(); ok; encOp, ok = q.reader.Decode() {
		switch ops.OpType(encOp.Data[0]) {
		case ops.TypeInvalidate:
			op := decodeInvalidateOp(encOp.Data)
			if !q.wakeup || op.At.Before(q.wakeupTime) {
				q.wakeup = true
				q.wakeupTime = op.At
			}
		case ops.TypeProfile:
			op := decodeProfileOp(encOp.Data, encOp.Refs)
			if q.profHandlers == nil {
				q.profHandlers = make(map[event.Tag]struct{})
			}
			q.profHandlers[op.Tag] = struct{}{}
		case ops.TypeClipboardRead:
			q.cqueue.ProcessReadClipboard(encOp.Refs)
		case ops.TypeClipboardWrite:
			q.cqueue.ProcessWriteClipboard(encOp.Refs)
		case ops.TypeSave:
			id := ops.DecodeSave(encOp.Data)
			pc.save(id)
		case ops.TypeLoad:
			id := ops.DecodeLoad(encOp.Data)
			pc.load(id)

		// Pointer ops.
		case ops.TypeClip:
			var op ops.ClipOp
			op.Decode(encOp.Data)
			pc.clip(op)
		case ops.TypePopClip:
			pc.popArea()
		case ops.TypePass:
			pc.pass()
		case ops.TypePopPass:
			pc.popPass()
		case ops.TypeTransform:
			t, push := ops.DecodeTransform(encOp.Data)
			pc.transform(t, push)
		case ops.TypePopTransform:
			pc.popTransform()
		case ops.TypePointerInput:
			bo := binary.LittleEndian
			op := pointer.InputOp{
				Tag:   encOp.Refs[0].(event.Tag),
				Grab:  encOp.Data[1] != 0,
				Types: pointer.Type(bo.Uint16(encOp.Data[2:])),
				ScrollBounds: image.Rectangle{
					Min: image.Point{
						X: int(int32(bo.Uint32(encOp.Data[4:]))),
						Y: int(int32(bo.Uint32(encOp.Data[8:]))),
					},
					Max: image.Point{
						X: int(int32(bo.Uint32(encOp.Data[12:]))),
						Y: int(int32(bo.Uint32(encOp.Data[16:]))),
					},
				},
			}
			pc.inputOp(op, &q.handlers)
		case ops.TypeCursor:
			name := encOp.Refs[0].(pointer.CursorName)
			pc.cursor(name)

		// Key ops.
		case ops.TypeKeyFocus:
			tag, _ := encOp.Refs[0].(event.Tag)
			op := key.FocusOp{
				Tag: tag,
			}
			kc.focusOp(op.Tag)
		case ops.TypeKeySoftKeyboard:
			op := key.SoftKeyboardOp{
				Show: encOp.Data[1] != 0,
			}
			kc.softKeyboard(op.Show)
		case ops.TypeKeyInput:
			op := key.InputOp{
				Tag:  encOp.Refs[0].(event.Tag),
				Hint: key.InputHint(encOp.Data[1]),
			}
			kc.inputOp(op)
		}
	}
}

// Profiling reports whether there was profile handlers in the
// most recent Frame call.
func (q *Router) Profiling() bool {
	return len(q.profHandlers) > 0
}

// WakeupTime returns the most recent time for doing another frame,
// as determined from the last call to Frame.
func (q *Router) WakeupTime() (time.Time, bool) {
	return q.wakeupTime, q.wakeup
}

func (h *handlerEvents) init() {
	if h.handlers == nil {
		h.handlers = make(map[event.Tag][]event.Event)
	}
}

func (h *handlerEvents) AddNoRedraw(k event.Tag, e event.Event) {
	h.init()
	h.handlers[k] = append(h.handlers[k], e)
}

func (h *handlerEvents) Add(k event.Tag, e event.Event) {
	h.AddNoRedraw(k, e)
	h.hadEvents = true
}

func (h *handlerEvents) HadEvents() bool {
	u := h.hadEvents
	h.hadEvents = false
	return u
}

func (h *handlerEvents) Events(k event.Tag) []event.Event {
	if events, ok := h.handlers[k]; ok {
		h.handlers[k] = h.handlers[k][:0]
		// Schedule another frame if we delivered events to the user
		// to flush half-updated state. This is important when an
		// event changes UI state that has already been laid out. In
		// the worst case, we waste a frame, increasing power usage.
		//
		// Gio is expected to grow the ability to construct
		// frame-to-frame differences and only render to changed
		// areas. In that case, the waste of a spurious frame should
		// be minimal.
		h.hadEvents = h.hadEvents || len(events) > 0
		return events
	}
	return nil
}

func (h *handlerEvents) Clear() {
	for k := range h.handlers {
		delete(h.handlers, k)
	}
}

func decodeProfileOp(d []byte, refs []interface{}) profile.Op {
	if ops.OpType(d[0]) != ops.TypeProfile {
		panic("invalid op")
	}
	return profile.Op{
		Tag: refs[0].(event.Tag),
	}
}

func decodeInvalidateOp(d []byte) op.InvalidateOp {
	bo := binary.LittleEndian
	if ops.OpType(d[0]) != ops.TypeInvalidate {
		panic("invalid op")
	}
	var o op.InvalidateOp
	if nanos := bo.Uint64(d[1:]); nanos > 0 {
		o.At = time.Unix(0, int64(nanos))
	}
	return o
}
