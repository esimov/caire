// SPDX-License-Identifier: Unlicense OR MIT

// Package byteslice provides byte slice views of other Go values  such as
// slices and structs.
package byteslice

import (
	"reflect"
	"unsafe"
)

// Struct returns a byte slice view of a struct.
func Struct(s interface{}) []byte {
	v := reflect.ValueOf(s).Elem()
	sz := int(v.Type().Size())
	var res []byte
	h := (*reflect.SliceHeader)(unsafe.Pointer(&res))
	h.Data = uintptr(unsafe.Pointer(v.UnsafeAddr()))
	h.Cap = sz
	h.Len = sz
	return res
}

// Uint32 returns a byte slice view of a uint32 slice.
func Uint32(s []uint32) []byte {
	n := len(s)
	if n == 0 {
		return nil
	}
	blen := n * int(unsafe.Sizeof(s[0]))
	return (*[1 << 30]byte)(unsafe.Pointer(&s[0]))[:blen:blen]
}

// Slice returns a byte slice view of a slice.
func Slice(s interface{}) []byte {
	v := reflect.ValueOf(s)
	first := v.Index(0)
	sz := int(first.Type().Size())
	var res []byte
	h := (*reflect.SliceHeader)(unsafe.Pointer(&res))
	h.Data = first.UnsafeAddr()
	h.Cap = v.Cap() * sz
	h.Len = v.Len() * sz
	return res
}
