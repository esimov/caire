// SPDX-License-Identifier: Unlicense OR MIT

//go:build !(linux && (arm64 || arm || amd64))
// +build !linux !arm64,!arm,!amd64

package cpu

import "unsafe"

type (
	BufferDescriptor  struct{}
	ImageDescriptor   struct{}
	SamplerDescriptor struct{}

	DispatchContext struct{}
	ThreadContext   struct{}
	ProgramInfo     struct{}
)

const Supported = false

func NewBuffer(size int) BufferDescriptor {
	panic("unsupported")
}

func (d *BufferDescriptor) Data() []byte {
	panic("unsupported")
}

func (d *BufferDescriptor) Free() {
}

func NewImageRGBA(width, height int) ImageDescriptor {
	panic("unsupported")
}

func (d *ImageDescriptor) Data() []byte {
	panic("unsupported")
}

func (d *ImageDescriptor) Free() {
}

func NewDispatchContext() *DispatchContext {
	panic("unsupported")
}

func (c *DispatchContext) Free() {
}

func (c *DispatchContext) Prepare(numThreads int, prog *ProgramInfo, descSet unsafe.Pointer, x, y, z int) {
	panic("unsupported")
}

func (c *DispatchContext) Dispatch(threadIdx int, ctx *ThreadContext) {
	panic("unsupported")
}

func NewThreadContext() *ThreadContext {
	panic("unsupported")
}

func (c *ThreadContext) Free() {
}
