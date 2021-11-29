// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"gioui.org/op/clip"
	"golang.org/x/image/math/fixed"
)

type layoutCache struct {
	m          map[layoutKey]*layoutElem
	head, tail *layoutElem
}

type pathCache struct {
	m          map[pathKey]*path
	head, tail *path
}

type layoutElem struct {
	next, prev *layoutElem
	key        layoutKey
	layout     []Line
}

type path struct {
	next, prev *path
	key        pathKey
	val        clip.Op
}

type layoutKey struct {
	ppem     fixed.Int26_6
	maxWidth int
	str      string
}

type pathKey struct {
	ppem fixed.Int26_6
	str  string
}

const maxSize = 1000

func (l *layoutCache) Get(k layoutKey) ([]Line, bool) {
	if lt, ok := l.m[k]; ok {
		l.remove(lt)
		l.insert(lt)
		return lt.layout, true
	}
	return nil, false
}

func (l *layoutCache) Put(k layoutKey, lt []Line) {
	if l.m == nil {
		l.m = make(map[layoutKey]*layoutElem)
		l.head = new(layoutElem)
		l.tail = new(layoutElem)
		l.head.prev = l.tail
		l.tail.next = l.head
	}
	val := &layoutElem{key: k, layout: lt}
	l.m[k] = val
	l.insert(val)
	if len(l.m) > maxSize {
		oldest := l.tail.next
		l.remove(oldest)
		delete(l.m, oldest.key)
	}
}

func (l *layoutCache) remove(lt *layoutElem) {
	lt.next.prev = lt.prev
	lt.prev.next = lt.next
}

func (l *layoutCache) insert(lt *layoutElem) {
	lt.next = l.head
	lt.prev = l.head.prev
	lt.prev.next = lt
	lt.next.prev = lt
}

func (c *pathCache) Get(k pathKey) (clip.Op, bool) {
	if v, ok := c.m[k]; ok {
		c.remove(v)
		c.insert(v)
		return v.val, true
	}
	return clip.Op{}, false
}

func (c *pathCache) Put(k pathKey, v clip.Op) {
	if c.m == nil {
		c.m = make(map[pathKey]*path)
		c.head = new(path)
		c.tail = new(path)
		c.head.prev = c.tail
		c.tail.next = c.head
	}
	val := &path{key: k, val: v}
	c.m[k] = val
	c.insert(val)
	if len(c.m) > maxSize {
		oldest := c.tail.next
		c.remove(oldest)
		delete(c.m, oldest.key)
	}
}

func (c *pathCache) remove(v *path) {
	v.next.prev = v.prev
	v.prev.next = v.next
}

func (c *pathCache) insert(v *path) {
	v.next = c.head
	v.prev = c.head.prev
	v.prev.next = v
	v.next.prev = v
}
