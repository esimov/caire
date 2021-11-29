// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && (arm64 || arm || amd64)
// +build linux
// +build arm64 arm amd64

package cpu

/*
#cgo CFLAGS: -std=c11 -D_POSIX_C_SOURCE=200112L

#include <stdint.h>
#include <stdlib.h>
#include "abi.h"
#include "runtime.h"
*/
import "C"
import (
	"unsafe"
)

type (
	BufferDescriptor  = C.struct_buffer_descriptor
	ImageDescriptor   = C.struct_image_descriptor
	SamplerDescriptor = C.struct_sampler_descriptor

	DispatchContext = C.struct_dispatch_context
	ThreadContext   = C.struct_thread_context
	ProgramInfo     = C.struct_program_info
)

const Supported = true

func NewBuffer(size int) BufferDescriptor {
	return C.alloc_buffer(C.size_t(size))
}

func (d *BufferDescriptor) Data() []byte {
	return (*(*[1 << 30]byte)(d.ptr))[:d.size_in_bytes:d.size_in_bytes]
}

func (d *BufferDescriptor) Free() {
	if d.ptr != nil {
		C.free(d.ptr)
	}
	*d = BufferDescriptor{}
}

func NewImageRGBA(width, height int) ImageDescriptor {
	return C.alloc_image_rgba(C.int(width), C.int(height))
}

func (d *ImageDescriptor) Data() []byte {
	return (*(*[1 << 30]byte)(d.ptr))[:d.size_in_bytes:d.size_in_bytes]
}

func (d *ImageDescriptor) Free() {
	if d.ptr != nil {
		C.free(d.ptr)
	}
	*d = ImageDescriptor{}
}

func NewDispatchContext() *DispatchContext {
	return C.alloc_dispatch_context()
}

func (c *DispatchContext) Free() {
	C.free_dispatch_context(c)
}

func (c *DispatchContext) Prepare(numThreads int, prog *ProgramInfo, descSet unsafe.Pointer, x, y, z int) {
	C.prepare_dispatch(c, C.int(numThreads), prog, (*C.uint8_t)(descSet), C.int(x), C.int(y), C.int(z))
}

func (c *DispatchContext) Dispatch(threadIdx int, ctx *ThreadContext) {
	C.dispatch_thread(c, C.int(threadIdx), ctx)
}

func NewThreadContext() *ThreadContext {
	return C.alloc_thread_context()
}

func (c *ThreadContext) Free() {
	C.free_thread_context(c)
}
