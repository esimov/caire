// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"gioui.org/io/system"
	"gioui.org/op/clip"
	"github.com/benoitkugler/textlayout/fonts"
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
	val        clip.PathSpec
	gids       []fonts.GID
}

type layoutKey struct {
	ppem     fixed.Int26_6
	maxWidth int
	str      string
	locale   system.Locale
}

type pathKey struct {
	ppem    fixed.Int26_6
	gidHash uint64
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

func gidsMatch(gids []fonts.GID, l Layout) bool {
	if len(gids) != len(l.Glyphs) {
		return false
	}
	for i := range gids {
		if gids[i] != l.Glyphs[i].ID {
			return false
		}
	}
	return true
}

func (c *pathCache) Get(k pathKey, l Layout) (clip.PathSpec, bool) {
	if v, ok := c.m[k]; ok && gidsMatch(v.gids, l) {
		c.remove(v)
		c.insert(v)
		return v.val, true
	}
	return clip.PathSpec{}, false
}

func (c *pathCache) Put(k pathKey, l Layout, v clip.PathSpec) {
	if c.m == nil {
		c.m = make(map[pathKey]*path)
		c.head = new(path)
		c.tail = new(path)
		c.head.prev = c.tail
		c.tail.next = c.head
	}
	gids := make([]fonts.GID, len(l.Glyphs))
	for i := range l.Glyphs {
		gids[i] = l.Glyphs[i].ID
	}
	val := &path{key: k, val: v, gids: gids}
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
