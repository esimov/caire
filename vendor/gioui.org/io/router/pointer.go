// SPDX-License-Identifier: Unlicense OR MIT

package router

import (
	"encoding/binary"
	"image"

	"gioui.org/f32"
	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
)

type pointerQueue struct {
	hitTree  []hitNode
	areas    []areaNode
	cursors  []cursorNode
	cursor   pointer.CursorName
	handlers map[event.Tag]*pointerHandler
	pointers []pointerInfo

	scratch []event.Tag
}

type hitNode struct {
	next int
	area int

	// For handler nodes.
	tag  event.Tag
	pass bool
}

type cursorNode struct {
	name pointer.CursorName
	area int
}

type pointerInfo struct {
	id       pointer.ID
	pressed  bool
	handlers []event.Tag
	// last tracks the last pointer event received,
	// used while processing frame events.
	last pointer.Event

	// entered tracks the tags that contain the pointer.
	entered []event.Tag
}

type pointerHandler struct {
	area      int
	active    bool
	wantsGrab bool
	types     pointer.Type
	// min and max horizontal/vertical scroll
	scrollRange image.Rectangle
}

type areaOp struct {
	kind areaKind
	rect f32.Rectangle
}

type areaNode struct {
	trans f32.Affine2D
	next  int
	area  areaOp
}

type areaKind uint8

// collectState represents the state for pointerCollector.
type collectState struct {
	t f32.Affine2D
	// nodePlusOne is the current node index, plus one to
	// make the zero value collectState the initial state.
	nodePlusOne int
	pass        int
}

// pointerCollector tracks the state needed to update an pointerQueue
// from pointer ops.
type pointerCollector struct {
	q          *pointerQueue
	state      collectState
	nodeStack  []int
	transStack []f32.Affine2D
	// states holds the storage for save/restore ops.
	states []f32.Affine2D
}

const (
	areaRect areaKind = iota
	areaEllipse
)

func (c *pointerCollector) save(id int) {
	if extra := id - len(c.states) + 1; extra > 0 {
		c.states = append(c.states, make([]f32.Affine2D, extra)...)
	}
	c.states[id] = c.state.t
}

func (c *pointerCollector) load(id int) {
	c.state = collectState{t: c.states[id]}
}

func (c *pointerCollector) clip(op ops.ClipOp) {
	area := -1
	if i := c.state.nodePlusOne - 1; i != -1 {
		n := c.q.hitTree[i]
		area = n.area
	}
	kind := areaRect
	if op.Shape == ops.Ellipse {
		kind = areaEllipse
	}
	areaOp := areaOp{kind: kind, rect: frect(op.Bounds)}
	c.q.areas = append(c.q.areas, areaNode{trans: c.state.t, next: area, area: areaOp})
	c.nodeStack = append(c.nodeStack, c.state.nodePlusOne-1)
	c.q.hitTree = append(c.q.hitTree, hitNode{
		next: c.state.nodePlusOne - 1,
		area: len(c.q.areas) - 1,
		pass: true,
	})
	c.state.nodePlusOne = len(c.q.hitTree) - 1 + 1
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

func (c *pointerCollector) popArea() {
	n := len(c.nodeStack)
	c.state.nodePlusOne = c.nodeStack[n-1] + 1
	c.nodeStack = c.nodeStack[:n-1]
}

func (c *pointerCollector) pass() {
	c.state.pass++
}

func (c *pointerCollector) popPass() {
	c.state.pass--
}

func (c *pointerCollector) transform(t f32.Affine2D, push bool) {
	if push {
		c.transStack = append(c.transStack, c.state.t)
	}
	c.state.t = c.state.t.Mul(t)
}

func (c *pointerCollector) popTransform() {
	n := len(c.transStack)
	c.state.t = c.transStack[n-1]
	c.transStack = c.transStack[:n-1]
}

func (c *pointerCollector) inputOp(op pointer.InputOp, events *handlerEvents) {
	area := -1
	if i := c.state.nodePlusOne - 1; i != -1 {
		n := c.q.hitTree[i]
		area = n.area
	}
	c.q.hitTree = append(c.q.hitTree, hitNode{
		next: c.state.nodePlusOne - 1,
		area: area,
		tag:  op.Tag,
		pass: c.state.pass > 0,
	})
	c.state.nodePlusOne = len(c.q.hitTree) - 1 + 1
	h, ok := c.q.handlers[op.Tag]
	if !ok {
		h = new(pointerHandler)
		c.q.handlers[op.Tag] = h
		// Cancel handlers on (each) first appearance, but don't
		// trigger redraw.
		events.AddNoRedraw(op.Tag, pointer.Event{Type: pointer.Cancel})
	}
	h.active = true
	h.area = area
	h.wantsGrab = h.wantsGrab || op.Grab
	h.types = h.types | op.Types
	h.scrollRange = op.ScrollBounds
}

func (c *pointerCollector) cursor(name pointer.CursorName) {
	c.q.cursors = append(c.q.cursors, cursorNode{
		name: name,
		area: len(c.q.areas) - 1,
	})
}

func (c *pointerCollector) reset(q *pointerQueue) {
	q.reset()
	c.state = collectState{}
	c.nodeStack = c.nodeStack[:0]
	c.transStack = c.transStack[:0]
	c.q = q
}

func (q *pointerQueue) opHit(handlers *[]event.Tag, pos f32.Point) {
	// Track whether we're passing through hits.
	pass := true
	idx := len(q.hitTree) - 1
	for idx >= 0 {
		n := &q.hitTree[idx]
		hit := q.hit(n.area, pos, n.pass)
		if !hit {
			idx--
			continue
		}
		pass = pass && n.pass
		if pass {
			idx--
		} else {
			idx = n.next
		}
		if n.tag != nil {
			if _, exists := q.handlers[n.tag]; exists {
				*handlers = addHandler(*handlers, n.tag)
			}
		}
	}
}

func (q *pointerQueue) invTransform(areaIdx int, p f32.Point) f32.Point {
	if areaIdx == -1 {
		return p
	}
	return q.areas[areaIdx].trans.Invert().Transform(p)
}

func (q *pointerQueue) hit(areaIdx int, p f32.Point, pass bool) bool {
	for areaIdx != -1 {
		a := &q.areas[areaIdx]
		p := a.trans.Invert().Transform(p)
		if !a.area.Hit(p) {
			return false
		}
		areaIdx = a.next
	}
	return true
}

func (q *pointerQueue) reset() {
	if q.handlers == nil {
		q.handlers = make(map[event.Tag]*pointerHandler)
	}
	for _, h := range q.handlers {
		// Reset handler.
		h.active = false
		h.wantsGrab = false
		h.types = 0
	}
	q.hitTree = q.hitTree[:0]
	q.areas = q.areas[:0]
	q.cursors = q.cursors[:0]
}

func (q *pointerQueue) Frame(events *handlerEvents) {
	for k, h := range q.handlers {
		if !h.active {
			q.dropHandler(nil, k)
			delete(q.handlers, k)
		}
		if h.wantsGrab {
			for _, p := range q.pointers {
				if !p.pressed {
					continue
				}
				for i, k2 := range p.handlers {
					if k2 == k {
						// Drop other handlers that lost their grab.
						dropped := q.scratch[:0]
						dropped = append(dropped, p.handlers[:i]...)
						dropped = append(dropped, p.handlers[i+1:]...)
						for _, tag := range dropped {
							q.dropHandler(events, tag)
						}
						break
					}
				}
			}
		}
	}
	for i := range q.pointers {
		p := &q.pointers[i]
		q.deliverEnterLeaveEvents(p, events, p.last)
	}
}

func (q *pointerQueue) dropHandler(events *handlerEvents, tag event.Tag) {
	if events != nil {
		events.Add(tag, pointer.Event{Type: pointer.Cancel})
	}
	for i := range q.pointers {
		p := &q.pointers[i]
		for i := len(p.handlers) - 1; i >= 0; i-- {
			if p.handlers[i] == tag {
				p.handlers = append(p.handlers[:i], p.handlers[i+1:]...)
			}
		}
		for i := len(p.entered) - 1; i >= 0; i-- {
			if p.entered[i] == tag {
				p.entered = append(p.entered[:i], p.entered[i+1:]...)
			}
		}
	}
}

// pointerOf returns the pointerInfo index corresponding to the pointer in e.
func (q *pointerQueue) pointerOf(e pointer.Event) int {
	for i, p := range q.pointers {
		if p.id == e.PointerID {
			return i
		}
	}
	q.pointers = append(q.pointers, pointerInfo{id: e.PointerID})
	return len(q.pointers) - 1
}

func (q *pointerQueue) Push(e pointer.Event, events *handlerEvents) {
	if e.Type == pointer.Cancel {
		q.pointers = q.pointers[:0]
		for k := range q.handlers {
			q.dropHandler(events, k)
		}
		return
	}
	pidx := q.pointerOf(e)
	p := &q.pointers[pidx]
	p.last = e

	switch e.Type {
	case pointer.Press:
		q.deliverEnterLeaveEvents(p, events, e)
		p.pressed = true
		q.deliverEvent(p, events, e)
	case pointer.Move:
		if p.pressed {
			e.Type = pointer.Drag
		}
		q.deliverEnterLeaveEvents(p, events, e)
		q.deliverEvent(p, events, e)
	case pointer.Release:
		q.deliverEvent(p, events, e)
		p.pressed = false
		q.deliverEnterLeaveEvents(p, events, e)
	case pointer.Scroll:
		q.deliverEnterLeaveEvents(p, events, e)
		q.deliverScrollEvent(p, events, e)
	default:
		panic("unsupported pointer event type")
	}

	if !p.pressed && len(p.entered) == 0 {
		// No longer need to track pointer.
		q.pointers = append(q.pointers[:pidx], q.pointers[pidx+1:]...)
	}
}

func (q *pointerQueue) deliverEvent(p *pointerInfo, events *handlerEvents, e pointer.Event) {
	foremost := true
	if p.pressed && len(p.handlers) == 1 {
		e.Priority = pointer.Grabbed
		foremost = false
	}
	for _, k := range p.handlers {
		h := q.handlers[k]
		if e.Type&h.types == 0 {
			continue
		}
		e := e
		if foremost {
			foremost = false
			e.Priority = pointer.Foremost
		}
		e.Position = q.invTransform(h.area, e.Position)
		events.Add(k, e)
	}
}

func (q *pointerQueue) deliverScrollEvent(p *pointerInfo, events *handlerEvents, e pointer.Event) {
	foremost := true
	if p.pressed && len(p.handlers) == 1 {
		e.Priority = pointer.Grabbed
		foremost = false
	}
	var sx, sy = e.Scroll.X, e.Scroll.Y
	for _, k := range p.handlers {
		if sx == 0 && sy == 0 {
			return
		}
		h := q.handlers[k]
		// Distribute the scroll to the handler based on its ScrollRange.
		sx, e.Scroll.X = setScrollEvent(sx, h.scrollRange.Min.X, h.scrollRange.Max.X)
		sy, e.Scroll.Y = setScrollEvent(sy, h.scrollRange.Min.Y, h.scrollRange.Max.Y)
		e := e
		if foremost {
			foremost = false
			e.Priority = pointer.Foremost
		}
		e.Position = q.invTransform(h.area, e.Position)
		events.Add(k, e)
	}
}

func (q *pointerQueue) deliverEnterLeaveEvents(p *pointerInfo, events *handlerEvents, e pointer.Event) {
	q.scratch = q.scratch[:0]
	q.opHit(&q.scratch, e.Position)
	if p.pressed {
		// Filter out non-participating handlers.
		for i := len(q.scratch) - 1; i >= 0; i-- {
			if _, found := searchTag(p.handlers, q.scratch[i]); !found {
				q.scratch = append(q.scratch[:i], q.scratch[i+1:]...)
			}
		}
	} else {
		p.handlers = append(p.handlers[:0], q.scratch...)
	}
	hits := q.scratch
	if e.Source != pointer.Mouse && !p.pressed && e.Type != pointer.Press {
		// Consider non-mouse pointers leaving when they're released.
		hits = nil
	}
	// Deliver Leave events.
	for _, k := range p.entered {
		if _, found := searchTag(hits, k); found {
			continue
		}
		h := q.handlers[k]
		e.Type = pointer.Leave

		if e.Type&h.types != 0 {
			e.Position = q.invTransform(h.area, e.Position)
			events.Add(k, e)
		}
	}
	// Deliver Enter events and update cursor.
	q.cursor = pointer.CursorDefault
	for _, k := range hits {
		h := q.handlers[k]
		for i := len(q.cursors) - 1; i >= 0; i-- {
			if c := q.cursors[i]; c.area == h.area {
				q.cursor = c.name
				break
			}
		}
		if _, found := searchTag(p.entered, k); found {
			continue
		}
		e.Type = pointer.Enter

		if e.Type&h.types != 0 {
			e.Position = q.invTransform(h.area, e.Position)
			events.Add(k, e)
		}
	}
	p.entered = append(p.entered[:0], hits...)
}

func searchTag(tags []event.Tag, tag event.Tag) (int, bool) {
	for i, t := range tags {
		if t == tag {
			return i, true
		}
	}
	return 0, false
}

// addHandler adds tag to the slice if not present.
func addHandler(tags []event.Tag, tag event.Tag) []event.Tag {
	for _, t := range tags {
		if t == tag {
			return tags
		}
	}
	return append(tags, tag)
}

func opDecodeFloat32(d []byte) float32 {
	return float32(int32(binary.LittleEndian.Uint32(d)))
}

func (op *areaOp) Hit(pos f32.Point) bool {
	pos = pos.Sub(op.rect.Min)
	size := op.rect.Size()
	switch op.kind {
	case areaRect:
		return 0 <= pos.X && pos.X < size.X &&
			0 <= pos.Y && pos.Y < size.Y
	case areaEllipse:
		rx := size.X / 2
		ry := size.Y / 2
		xh := pos.X - rx
		yk := pos.Y - ry
		// The ellipse function works in all cases because
		// 0/0 is not <= 1.
		return (xh*xh)/(rx*rx)+(yk*yk)/(ry*ry) <= 1
	default:
		panic("invalid area kind")
	}
}

func setScrollEvent(scroll float32, min, max int) (left, scrolled float32) {
	if v := float32(max); scroll > v {
		return scroll - v, v
	}
	if v := float32(min); scroll < v {
		return scroll - v, v
	}
	return 0, scroll
}
