// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && nometal
// +build darwin,!ios,nometal

package app

import (
	"errors"
	"runtime"

	"unsafe"

	"gioui.org/gpu"
	"gioui.org/internal/gl"
)

/*
#include <CoreFoundation/CoreFoundation.h>
#include <CoreGraphics/CoreGraphics.h>
#include <AppKit/AppKit.h>
#include <dlfcn.h>

__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createGLContext(void);
__attribute__ ((visibility ("hidden"))) void gio_setContextView(CFTypeRef ctx, CFTypeRef view);
__attribute__ ((visibility ("hidden"))) void gio_makeCurrentContext(CFTypeRef ctx);
__attribute__ ((visibility ("hidden"))) void gio_updateContext(CFTypeRef ctx);
__attribute__ ((visibility ("hidden"))) void gio_flushContextBuffer(CFTypeRef ctx);
__attribute__ ((visibility ("hidden"))) void gio_clearCurrentContext(void);
__attribute__ ((visibility ("hidden"))) void gio_lockContext(CFTypeRef ctxRef);
__attribute__ ((visibility ("hidden"))) void gio_unlockContext(CFTypeRef ctxRef);

typedef void (*PFN_glFlush)(void);

static void glFlush(PFN_glFlush f) {
	f();
}
*/
import "C"

type glContext struct {
	c    *gl.Functions
	ctx  C.CFTypeRef
	view C.CFTypeRef

	glFlush C.PFN_glFlush
}

func newContext(w *window) (*glContext, error) {
	clib := C.CString("/System/Library/Frameworks/OpenGL.framework/OpenGL")
	defer C.free(unsafe.Pointer(clib))
	lib, err := C.dlopen(clib, C.RTLD_NOW|C.RTLD_LOCAL)
	if err != nil {
		return nil, err
	}
	csym := C.CString("glFlush")
	defer C.free(unsafe.Pointer(csym))
	glFlush := C.PFN_glFlush(C.dlsym(lib, csym))
	if glFlush == nil {
		return nil, errors.New("gl: missing symbol glFlush in the OpenGL framework")
	}
	view := w.contextView()
	ctx := C.gio_createGLContext()
	if ctx == 0 {
		return nil, errors.New("gl: failed to create NSOpenGLContext")
	}
	C.gio_setContextView(ctx, view)
	c := &glContext{
		ctx:     ctx,
		view:    view,
		glFlush: glFlush,
	}
	return c, nil
}

func (c *glContext) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.OpenGLRenderTarget{}, nil
}

func (c *glContext) API() gpu.API {
	return gpu.OpenGL{}
}

func (c *glContext) Release() {
	if c.ctx != 0 {
		C.gio_clearCurrentContext()
		C.CFRelease(c.ctx)
		c.ctx = 0
	}
}

func (c *glContext) Present() error {
	// Assume the caller already locked the context.
	C.glFlush(c.glFlush)
	return nil
}

func (c *glContext) Lock() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	C.gio_lockContext(c.ctx)
	C.gio_makeCurrentContext(c.ctx)
	return nil
}

func (c *glContext) Unlock() {
	C.gio_clearCurrentContext()
	C.gio_unlockContext(c.ctx)
}

func (c *glContext) Refresh() error {
	c.Lock()
	defer c.Unlock()
	C.gio_updateContext(c.ctx)
	return nil
}

func (w *window) NewContext() (context, error) {
	return newContext(w)
}
