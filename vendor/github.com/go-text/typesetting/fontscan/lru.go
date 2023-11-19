package fontscan

import (
	"hash/maphash"

	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/api/metadata"
)

// runeLRUEntry holds a single key-value pair for an LRU cache.
type runeLRUEntry struct {
	next, prev *runeLRUEntry
	key        runeLRUKey
	families   []string
	v          *font.Face
}

type runeLRUKey struct {
	familiesHash uint64
	aspect       metadata.Aspect
	r            rune
}

// runeLRU is a least-recently-used cache for font faces supporting a given rune.
type runeLRU struct {
	// This implementation is derived from the one here under the terms of the UNLICENSE:
	//
	// https://git.sr.ht/~eliasnaur/gio/tree/e768fe347a732056031100f2c66987d6db258ea4/item/text/lru.go
	m          map[runeLRUKey]*runeLRUEntry
	head, tail *runeLRUEntry
	maxSize    int
	seed       maphash.Seed
}

func (l *runeLRU) Clear() {
	l.m = make(map[runeLRUKey]*runeLRUEntry)
	l.seed = maphash.MakeSeed()
	l.head = new(runeLRUEntry)
	l.tail = new(runeLRUEntry)
	l.head.prev = l.tail
	l.tail.next = l.head
}

func (l *runeLRU) init() {
	if l.m == nil {
		l.Clear()
	}
}

func (l *runeLRU) KeyFor(q Query, r rune) runeLRUKey {
	l.init()
	var h maphash.Hash
	h.SetSeed(l.seed)
	for _, s := range q.Families {
		h.WriteString(s)
	}
	return runeLRUKey{
		aspect:       q.Aspect,
		familiesHash: h.Sum64(),
		r:            r,
	}
}

// Get fetches the value associated with the given key, if any.
func (l *runeLRU) Get(k runeLRUKey, q Query) (*font.Face, bool) {
	if lt, ok := l.m[k]; ok {
		if len(lt.families) != len(q.Families) {
			return nil, false
		}
		for i := range lt.families {
			if lt.families[i] != q.Families[i] {
				return nil, false
			}
		}
		l.remove(lt)
		l.insert(lt)
		return lt.v, true
	}
	return nil, false
}

func copyStrSlice(s []string) []string {
	out := make([]string, len(s))
	copy(out, s)
	return out
}

// Put inserts the given value with the given key, evicting old
// cache entries if necessary.
func (l *runeLRU) Put(k runeLRUKey, q Query, v *font.Face) {
	l.init()
	val := &runeLRUEntry{key: k, v: v, families: copyStrSlice(q.Families)}
	l.m[k] = val
	l.insert(val)
	for len(l.m) > l.maxSize {
		oldest := l.tail.next
		l.remove(oldest)
		delete(l.m, oldest.key)
	}
}

// remove cuts e out of the lru linked list.
func (l *runeLRU) remove(e *runeLRUEntry) {
	e.next.prev = e.prev
	e.prev.next = e.next
}

// insert adds e to the lru linked list.
func (l *runeLRU) insert(e *runeLRUEntry) {
	e.next = l.head
	e.prev = l.head.prev
	e.prev.next = e
	e.next.prev = e
}
