// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"encoding/binary"
	"hash/maphash"
	"image"

	giofont "gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"golang.org/x/image/math/fixed"
)

// entry holds a single key-value pair for an LRU cache.
type entry[K comparable, V any] struct {
	next, prev *entry[K, V]
	key        K
	v          V
}

// lru is a generic least-recently-used cache.
type lru[K comparable, V any] struct {
	m          map[K]*entry[K, V]
	head, tail *entry[K, V]
}

// Get fetches the value associated with the given key, if any.
func (l *lru[K, V]) Get(k K) (V, bool) {
	if lt, ok := l.m[k]; ok {
		l.remove(lt)
		l.insert(lt)
		return lt.v, true
	}
	var v V
	return v, false
}

// Put inserts the given value with the given key, evicting old
// cache entries if necessary.
func (l *lru[K, V]) Put(k K, v V) {
	if l.m == nil {
		l.m = make(map[K]*entry[K, V])
		l.head = new(entry[K, V])
		l.tail = new(entry[K, V])
		l.head.prev = l.tail
		l.tail.next = l.head
	}
	val := &entry[K, V]{key: k, v: v}
	l.m[k] = val
	l.insert(val)
	if len(l.m) > maxSize {
		oldest := l.tail.next
		l.remove(oldest)
		delete(l.m, oldest.key)
	}
}

// remove cuts e out of the lru linked list.
func (l *lru[K, V]) remove(e *entry[K, V]) {
	e.next.prev = e.prev
	e.prev.next = e.next
}

// insert adds e to the lru linked list.
func (l *lru[K, V]) insert(e *entry[K, V]) {
	e.next = l.head
	e.prev = l.head.prev
	e.prev.next = e
	e.next.prev = e
}

type bitmapCache = lru[GlyphID, bitmap]

type bitmap struct {
	img  paint.ImageOp
	size image.Point
}

type layoutCache = lru[layoutKey, document]

type glyphValue[V any] struct {
	v      V
	glyphs []glyphInfo
}

type glyphLRU[V any] struct {
	seed  maphash.Seed
	cache lru[uint64, glyphValue[V]]
}

// hashGlyphs computes a hash key based on the ID and X offset of
// every glyph in the slice.
func (c *glyphLRU[V]) hashGlyphs(gs []Glyph) uint64 {
	if c.seed == (maphash.Seed{}) {
		c.seed = maphash.MakeSeed()
	}
	var h maphash.Hash
	h.SetSeed(c.seed)
	var b [8]byte
	firstX := fixed.Int26_6(0)
	for i, g := range gs {
		if i == 0 {
			firstX = g.X
		}
		// Cache glyph X offsets relative to the first glyph.
		binary.LittleEndian.PutUint32(b[:4], uint32(g.X-firstX))
		h.Write(b[:4])
		binary.LittleEndian.PutUint64(b[:], uint64(g.ID))
		h.Write(b[:])
	}
	sum := h.Sum64()
	return sum
}

func (c *glyphLRU[V]) Get(key uint64, gs []Glyph) (V, bool) {
	if v, ok := c.cache.Get(key); ok && gidsEqual(v.glyphs, gs) {
		return v.v, true
	}
	var v V
	return v, false
}

func (c *glyphLRU[V]) Put(key uint64, glyphs []Glyph, v V) {
	gids := make([]glyphInfo, len(glyphs))
	firstX := fixed.I(0)
	for i, glyph := range glyphs {
		if i == 0 {
			firstX = glyph.X
		}
		// Cache glyph X offsets relative to the first glyph.
		gids[i] = glyphInfo{ID: glyph.ID, X: glyph.X - firstX}
	}
	val := glyphValue[V]{
		glyphs: gids,
		v:      v,
	}
	c.cache.Put(key, val)
}

type pathCache = glyphLRU[clip.PathSpec]

type bitmapShapeCache = glyphLRU[op.CallOp]

type glyphInfo struct {
	ID GlyphID
	X  fixed.Int26_6
}

type layoutKey struct {
	ppem               fixed.Int26_6
	maxWidth, minWidth int
	maxLines           int
	str                string
	truncator          string
	locale             system.Locale
	font               giofont.Font
	forceTruncate      bool
	wrapPolicy         WrapPolicy
	lineHeight         fixed.Int26_6
	lineHeightScale    float32
}

type pathKey struct {
	gidHash uint64
}

const maxSize = 1000

func gidsEqual(a []glyphInfo, glyphs []Glyph) bool {
	if len(a) != len(glyphs) {
		return false
	}
	firstX := fixed.Int26_6(0)
	for i := range a {
		if i == 0 {
			firstX = glyphs[i].X
		}
		// Cache glyph X offsets relative to the first glyph.
		if a[i].ID != glyphs[i].ID || a[i].X != (glyphs[i].X-firstX) {
			return false
		}
	}
	return true
}
