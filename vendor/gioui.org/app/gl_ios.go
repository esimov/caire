// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && ios && nometal
// +build darwin,ios,nometal

package app

/*
@import UIKit;

#include <CoreFoundation/CoreFoundation.h>
#include <OpenGLES/ES2/gl.h>
#include <OpenGLES/ES2/glext.h>

__attribute__ ((visibility ("hidden"))) int gio_renderbufferStorage(CFTypeRef ctx, CFTypeRef layer, GLenum buffer);
__attribute__ ((visibility ("hidden"))) int gio_presentRenderbuffer(CFTypeRef ctx, GLenum buffer);
__attribute__ ((visibility ("hidden"))) int gio_makeCurrent(CFTypeRef ctx);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createContext(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createGLLayer(void);

static CFTypeRef getViewLayer(CFTypeRef viewRef) {
	@autoreleasepool {
		UIView *view = (__bridge UIView *)viewRef;
		return CFBridgingRetain(view.layer);
	}
}

*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"

	"gioui.org/gpu"
	"gioui.org/internal/gl"
)

type context struct {
	owner       *window
	c           *gl.Functions
	ctx         C.CFTypeRef
	layer       C.CFTypeRef
	init        bool
	frameBuffer gl.Framebuffer
	colorBuffer gl.Renderbuffer
}

func newContext(w *window) (*context, error) {
	ctx := C.gio_createContext()
	if ctx == 0 {
		return nil, fmt.Errorf("failed to create EAGLContext")
	}
	api := contextAPI()
	f, err := gl.NewFunctions(api.Context, api.ES)
	if err != nil {
		return nil, err
	}
	c := &context{
		ctx:   ctx,
		owner: w,
		layer: C.getViewLayer(w.contextView()),
		c:     f,
	}
	return c, nil
}

func contextAPI() gpu.OpenGL {
	return gpu.OpenGL{}
}

func (c *context) RenderTarget() gpu.RenderTarget {
	return gpu.OpenGLRenderTarget(c.frameBuffer)
}

func (c *context) API() gpu.API {
	return contextAPI()
}

func (c *context) Release() {
	if c.ctx == 0 {
		return
	}
	C.gio_renderbufferStorage(c.ctx, 0, C.GLenum(gl.RENDERBUFFER))
	c.c.DeleteFramebuffer(c.frameBuffer)
	c.c.DeleteRenderbuffer(c.colorBuffer)
	C.gio_makeCurrent(0)
	C.CFRelease(c.ctx)
	c.ctx = 0
}

func (c *context) Present() error {
	if c.layer == 0 {
		panic("context is not active")
	}
	c.c.BindRenderbuffer(gl.RENDERBUFFER, c.colorBuffer)
	if C.gio_presentRenderbuffer(c.ctx, C.GLenum(gl.RENDERBUFFER)) == 0 {
		return errors.New("presentRenderBuffer failed")
	}
	return nil
}

func (c *context) Lock() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	if C.gio_makeCurrent(c.ctx) == 0 {
		return errors.New("[EAGLContext setCurrentContext] failed")
	}
	return nil
}

func (c *context) Unlock() {
	C.gio_makeCurrent(0)
}

func (c *context) Refresh() error {
	if C.gio_makeCurrent(c.ctx) == 0 {
		return errors.New("[EAGLContext setCurrentContext] failed")
	}
	if !c.init {
		c.init = true
		c.frameBuffer = c.c.CreateFramebuffer()
		c.colorBuffer = c.c.CreateRenderbuffer()
	}
	if !c.owner.visible {
		// Make sure any in-flight GL commands are complete.
		c.c.Finish()
		return nil
	}
	currentRB := gl.Renderbuffer{uint(c.c.GetInteger(gl.RENDERBUFFER_BINDING))}
	c.c.BindRenderbuffer(gl.RENDERBUFFER, c.colorBuffer)
	if C.gio_renderbufferStorage(c.ctx, c.layer, C.GLenum(gl.RENDERBUFFER)) == 0 {
		return errors.New("renderbufferStorage failed")
	}
	c.c.BindRenderbuffer(gl.RENDERBUFFER, currentRB)
	c.c.BindFramebuffer(gl.FRAMEBUFFER, c.frameBuffer)
	c.c.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, c.colorBuffer)
	if st := c.c.CheckFramebufferStatus(gl.FRAMEBUFFER); st != gl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("framebuffer incomplete, status: %#x\n", st)
	}
	return nil
}

func (w *window) NewContext() (Context, error) {
	return newContext(w)
}
