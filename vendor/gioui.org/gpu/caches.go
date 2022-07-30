// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"fmt"

	"gioui.org/internal/f32"
)

type resourceCache struct {
	res map[interface{}]resourceCacheValue
}

type resourceCacheValue struct {
	used     bool
	resource resource
}

// opCache is like a resourceCache but using concrete types and a
// freelist instead of two maps to avoid runtime.mapaccess2 calls
// since benchmarking showed them as a bottleneck.
type opCache struct {
	// store the index + 1 in cache this key is stored in
	index map[opKey]int
	// list of indexes in cache that are free and can be used
	freelist []int
	cache    []opCacheValue
}

type opCacheValue struct {
	data pathData

	bounds f32.Rectangle
	// the fields below are handled by opCache
	key  opKey
	keep bool
}

func newResourceCache() *resourceCache {
	return &resourceCache{
		res: make(map[interface{}]resourceCacheValue),
	}
}

func (r *resourceCache) get(key interface{}) (resource, bool) {
	v, exists := r.res[key]
	if !exists {
		return nil, false
	}
	if !v.used {
		v.used = true
		r.res[key] = v
	}
	return v.resource, exists
}

func (r *resourceCache) put(key interface{}, val resource) {
	v, exists := r.res[key]
	if exists && v.used {
		panic(fmt.Errorf("key exists, %p", key))
	}
	v.used = true
	v.resource = val
	r.res[key] = v
}

func (r *resourceCache) frame() {
	for k, v := range r.res {
		if v.used {
			v.used = false
			r.res[k] = v
		} else {
			delete(r.res, k)
			v.resource.release()
		}
	}
}

func (r *resourceCache) release() {
	for _, v := range r.res {
		v.resource.release()
	}
	r.res = nil
}

func newOpCache() *opCache {
	return &opCache{
		index:    make(map[opKey]int),
		freelist: make([]int, 0),
		cache:    make([]opCacheValue, 0),
	}
}

func (r *opCache) get(key opKey) (o opCacheValue, exist bool) {
	v := r.index[key]
	if v == 0 {
		return
	}
	r.cache[v-1].keep = true
	return r.cache[v-1], true
}

func (r *opCache) put(key opKey, val opCacheValue) {
	v := r.index[key]
	val.keep = true
	val.key = key
	if v == 0 {
		// not in cache
		i := len(r.cache)
		if len(r.freelist) > 0 {
			i = r.freelist[len(r.freelist)-1]
			r.freelist = r.freelist[:len(r.freelist)-1]
			r.cache[i] = val
		} else {
			r.cache = append(r.cache, val)
		}
		r.index[key] = i + 1
	} else {
		r.cache[v-1] = val
	}
}

func (r *opCache) frame() {
	r.freelist = r.freelist[:0]
	for i, v := range r.cache {
		r.cache[i].keep = false
		if v.keep {
			continue
		}
		if v.data.data != nil {
			v.data.release()
			r.cache[i].data.data = nil
		}
		delete(r.index, v.key)
		r.freelist = append(r.freelist, i)
	}
}

func (r *opCache) release() {
	for i := range r.cache {
		r.cache[i].keep = false
	}
	r.frame()
	r.index = nil
	r.freelist = nil
	r.cache = nil
}
