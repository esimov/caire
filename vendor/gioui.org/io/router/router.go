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
	"io"
	"math"
	"strings"
	"time"

	"gioui.org/f32"
	f32internal "gioui.org/internal/f32"
	"gioui.org/internal/ops"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/profile"
	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/io/transfer"
	"gioui.org/op"
)

// Router is a Queue implementation that routes events
// to handlers declared in operation lists.
type Router struct {
	savedTrans []f32.Affine2D
	transStack []f32.Affine2D
	pointer    struct {
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

// SemanticNode represents a node in the tree describing the components
// contained in a frame.
type SemanticNode struct {
	ID       SemanticID
	ParentID SemanticID
	Children []SemanticNode
	Desc     SemanticDesc

	areaIdx int
}

// SemanticDesc provides a semantic description of a UI component.
type SemanticDesc struct {
	Class       semantic.ClassOp
	Description string
	Label       string
	Selected    bool
	Disabled    bool
	Gestures    SemanticGestures
	Bounds      image.Rectangle
}

// SemanticGestures is a bit-set of supported gestures.
type SemanticGestures int

const (
	ClickGesture SemanticGestures = 1 << iota
	ScrollGesture
)

// SemanticID uniquely identifies a SemanticDescription.
//
// By convention, the zero value denotes the non-existent ID.
type SemanticID uint64

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

// Queue key events to the topmost handler.
func (q *Router) QueueTopmost(events ...key.Event) bool {
	var topmost event.Tag
	pq := &q.pointer.queue
	for _, h := range pq.hitTree {
		if h.ktag != nil {
			topmost = h.ktag
			break
		}
	}
	if topmost == nil {
		return false
	}
	for _, e := range events {
		q.handlers.Add(topmost, e)
	}
	return q.handlers.HadEvents()
}

// Queue events and report whether at least one handler had an event queued.
func (q *Router) Queue(events ...event.Event) bool {
	for _, e := range events {
		switch e := e.(type) {
		case profile.Event:
			q.profile = e
		case pointer.Event:
			q.pointer.queue.Push(e, &q.handlers)
		case key.Event:
			q.queueKeyEvent(e)
		case key.SnippetEvent:
			// Expand existing, overlapping snippet.
			if r := q.key.queue.content.Snippet.Range; rangeOverlaps(r, key.Range(e)) {
				if e.Start > r.Start {
					e.Start = r.Start
				}
				if e.End < r.End {
					e.End = r.End
				}
			}
			if f := q.key.queue.focus; f != nil {
				q.handlers.Add(f, e)
			}
		case key.EditEvent, key.FocusEvent, key.SelectionEvent:
			if f := q.key.queue.focus; f != nil {
				q.handlers.Add(f, e)
			}
		case clipboard.Event:
			q.cqueue.Push(e, &q.handlers)
		}
	}
	return q.handlers.HadEvents()
}

func rangeOverlaps(r1, r2 key.Range) bool {
	r1 = rangeNorm(r1)
	r2 = rangeNorm(r2)
	return r1.Start <= r2.Start && r2.Start < r1.End ||
		r1.Start <= r2.End && r2.End < r1.End
}

func rangeNorm(r key.Range) key.Range {
	if r.End < r.Start {
		r.End, r.Start = r.Start, r.End
	}
	return r
}

func (q *Router) queueKeyEvent(e key.Event) {
	kq := &q.key.queue
	f := q.key.queue.focus
	if f != nil && kq.Accepts(f, e) {
		q.handlers.Add(f, e)
		return
	}
	pq := &q.pointer.queue
	idx := len(pq.hitTree) - 1
	focused := f != nil
	if focused {
		// If there is a focused tag, traverse its ancestry through the
		// hit tree to search for handlers.
		for ; pq.hitTree[idx].ktag != f; idx-- {
		}
	}
	for idx != -1 {
		n := &pq.hitTree[idx]
		if focused {
			idx = n.next
		} else {
			idx--
		}
		if n.ktag == nil {
			continue
		}
		if kq.Accepts(n.ktag, e) {
			q.handlers.Add(n.ktag, e)
			break
		}
	}
}

func (q *Router) MoveFocus(dir FocusDirection) bool {
	return q.key.queue.MoveFocus(dir, &q.handlers)
}

// RevealFocus scrolls the current focus (if any) into viewport
// if there are scrollable parent handlers.
func (q *Router) RevealFocus(viewport image.Rectangle) {
	focus := q.key.queue.focus
	if focus == nil {
		return
	}
	bounds := q.key.queue.BoundsFor(focus)
	area := q.key.queue.AreaFor(focus)
	viewport = q.pointer.queue.ClipFor(area, viewport)

	topleft := bounds.Min.Sub(viewport.Min)
	topleft = max(topleft, bounds.Max.Sub(viewport.Max))
	topleft = min(image.Pt(0, 0), topleft)
	bottomright := bounds.Max.Sub(viewport.Max)
	bottomright = min(bottomright, bounds.Min.Sub(viewport.Min))
	bottomright = max(image.Pt(0, 0), bottomright)
	s := topleft
	if s.X == 0 {
		s.X = bottomright.X
	}
	if s.Y == 0 {
		s.Y = bottomright.Y
	}
	q.ScrollFocus(s)
}

// ScrollFocus scrolls the focused widget, if any, by dist.
func (q *Router) ScrollFocus(dist image.Point) {
	focus := q.key.queue.focus
	if focus == nil {
		return
	}
	area := q.key.queue.AreaFor(focus)
	q.pointer.queue.Deliver(area, pointer.Event{
		Type:   pointer.Scroll,
		Source: pointer.Touch,
		Scroll: f32internal.FPt(dist),
	}, &q.handlers)
}

func max(p1, p2 image.Point) image.Point {
	m := p1
	if p2.X > m.X {
		m.X = p2.X
	}
	if p2.Y > m.Y {
		m.Y = p2.Y
	}
	return m
}

func min(p1, p2 image.Point) image.Point {
	m := p1
	if p2.X < m.X {
		m.X = p2.X
	}
	if p2.Y < m.Y {
		m.Y = p2.Y
	}
	return m
}

func (q *Router) ActionAt(p f32.Point) (system.Action, bool) {
	return q.pointer.queue.ActionAt(p)
}

func (q *Router) ClickFocus() {
	focus := q.key.queue.focus
	if focus == nil {
		return
	}
	bounds := q.key.queue.BoundsFor(focus)
	center := bounds.Max.Add(bounds.Min).Div(2)
	e := pointer.Event{
		Position: f32.Pt(float32(center.X), float32(center.Y)),
		Source:   pointer.Touch,
	}
	area := q.key.queue.AreaFor(focus)
	e.Type = pointer.Press
	q.pointer.queue.Deliver(area, e, &q.handlers)
	e.Type = pointer.Release
	q.pointer.queue.Deliver(area, e, &q.handlers)
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
func (q *Router) Cursor() pointer.Cursor {
	return q.pointer.queue.cursor
}

// SemanticAt returns the first semantic description under pos, if any.
func (q *Router) SemanticAt(pos f32.Point) (SemanticID, bool) {
	return q.pointer.queue.SemanticAt(pos)
}

// AppendSemantics appends the semantic tree to nodes, and returns the result.
// The root node is the first added.
func (q *Router) AppendSemantics(nodes []SemanticNode) []SemanticNode {
	q.pointer.collector.q = &q.pointer.queue
	q.pointer.collector.ensureRoot()
	return q.pointer.queue.AppendSemantics(nodes)
}

// EditorState returns the editor state for the focused handler, or the
// zero value if there is none.
func (q *Router) EditorState() EditorState {
	return q.key.queue.content
}

func (q *Router) collect() {
	q.transStack = q.transStack[:0]
	pc := &q.pointer.collector
	pc.q = &q.pointer.queue
	pc.reset()
	kc := &q.key.collector
	*kc = keyCollector{q: &q.key.queue}
	q.key.queue.Reset()
	var t f32.Affine2D
	bo := binary.LittleEndian
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
			if extra := id - len(q.savedTrans) + 1; extra > 0 {
				q.savedTrans = append(q.savedTrans, make([]f32.Affine2D, extra)...)
			}
			q.savedTrans[id] = t
		case ops.TypeLoad:
			id := ops.DecodeLoad(encOp.Data)
			t = q.savedTrans[id]
			pc.resetState()
			pc.setTrans(t)

		case ops.TypeClip:
			var op ops.ClipOp
			op.Decode(encOp.Data)
			pc.clip(op)
		case ops.TypePopClip:
			pc.popArea()
		case ops.TypeTransform:
			t2, push := ops.DecodeTransform(encOp.Data)
			if push {
				q.transStack = append(q.transStack, t)
			}
			t = t.Mul(t2)
			pc.setTrans(t)
		case ops.TypePopTransform:
			n := len(q.transStack)
			t = q.transStack[n-1]
			q.transStack = q.transStack[:n-1]
			pc.setTrans(t)

		// Pointer ops.
		case ops.TypePass:
			pc.pass()
		case ops.TypePopPass:
			pc.popPass()
		case ops.TypePointerInput:
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
			name := pointer.Cursor(encOp.Data[1])
			pc.cursor(name)
		case ops.TypeSource:
			op := transfer.SourceOp{
				Tag:  encOp.Refs[0].(event.Tag),
				Type: encOp.Refs[1].(string),
			}
			pc.sourceOp(op, &q.handlers)
		case ops.TypeTarget:
			op := transfer.TargetOp{
				Tag:  encOp.Refs[0].(event.Tag),
				Type: encOp.Refs[1].(string),
			}
			pc.targetOp(op, &q.handlers)
		case ops.TypeOffer:
			op := transfer.OfferOp{
				Tag:  encOp.Refs[0].(event.Tag),
				Type: encOp.Refs[1].(string),
				Data: encOp.Refs[2].(io.ReadCloser),
			}
			pc.offerOp(op, &q.handlers)
		case ops.TypeActionInput:
			act := system.Action(encOp.Data[1])
			pc.actionInputOp(act)

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
			filter := key.Set(*encOp.Refs[1].(*string))
			op := key.InputOp{
				Tag:  encOp.Refs[0].(event.Tag),
				Hint: key.InputHint(encOp.Data[1]),
				Keys: filter,
			}
			a := pc.currentArea()
			b := pc.currentAreaBounds()
			pc.keyInputOp(op)
			kc.inputOp(op, a, b)
		case ops.TypeSnippet:
			op := key.SnippetOp{
				Tag: encOp.Refs[0].(event.Tag),
				Snippet: key.Snippet{
					Range: key.Range{
						Start: int(int32(bo.Uint32(encOp.Data[1:]))),
						End:   int(int32(bo.Uint32(encOp.Data[5:]))),
					},
					Text: *(encOp.Refs[1].(*string)),
				},
			}
			kc.snippetOp(op)
		case ops.TypeSelection:
			op := key.SelectionOp{
				Tag: encOp.Refs[0].(event.Tag),
				Range: key.Range{
					Start: int(int32(bo.Uint32(encOp.Data[1:]))),
					End:   int(int32(bo.Uint32(encOp.Data[5:]))),
				},
				Caret: key.Caret{
					Pos: f32.Point{
						X: math.Float32frombits(bo.Uint32(encOp.Data[9:])),
						Y: math.Float32frombits(bo.Uint32(encOp.Data[13:])),
					},
					Ascent:  math.Float32frombits(bo.Uint32(encOp.Data[17:])),
					Descent: math.Float32frombits(bo.Uint32(encOp.Data[21:])),
				},
			}
			kc.selectionOp(t, op)

		// Semantic ops.
		case ops.TypeSemanticLabel:
			lbl := *encOp.Refs[0].(*string)
			pc.semanticLabel(lbl)
		case ops.TypeSemanticDesc:
			desc := *encOp.Refs[0].(*string)
			pc.semanticDesc(desc)
		case ops.TypeSemanticClass:
			class := semantic.ClassOp(encOp.Data[1])
			pc.semanticClass(class)
		case ops.TypeSemanticSelected:
			if encOp.Data[1] != 0 {
				pc.semanticSelected(true)
			} else {
				pc.semanticSelected(false)
			}
		case ops.TypeSemanticDisabled:
			if encOp.Data[1] != 0 {
				pc.semanticDisabled(true)
			} else {
				pc.semanticDisabled(false)
			}
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

func (s SemanticGestures) String() string {
	var gestures []string
	if s&ClickGesture != 0 {
		gestures = append(gestures, "Click")
	}
	return strings.Join(gestures, ",")
}
