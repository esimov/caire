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
	v := reflect.ValueOf(s)
	sz := int(v.Elem().Type().Size())
	return unsafe.Slice((*byte)(unsafe.Pointer(v.Pointer())), sz)
}

// Uint32 returns a byte slice view of a uint32 slice.
func Uint32(s []uint32) []byte {
	n := len(s)
	if n == 0 {
		return nil
	}
	blen := n * int(unsafe.Sizeof(s[0]))
	return unsafe.Slice((*byte)(unsafe.Pointer(&s[0])), blen)
}

// Slice returns a byte slice view of a slice.
func Slice(s interface{}) []byte {
	v := reflect.ValueOf(s)
	first := v.Index(0)
	sz := int(first.Type().Size())
	res := unsafe.Slice((*byte)(unsafe.Pointer(v.Pointer())), sz*v.Cap())
	return res[:sz*v.Len()]
}
