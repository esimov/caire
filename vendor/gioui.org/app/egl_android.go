// SPDX-License-Identifier: Unlicense OR MIT

package app

/*
#include <android/native_window_jni.h>
#include <EGL/egl.h>
*/
import "C"

import (
	"unsafe"

	"gioui.org/internal/egl"
)

type androidContext struct {
	win           *window
	eglSurf       egl.NativeWindowType
	width, height int
	*egl.Context
}

func init() {
	newAndroidGLESContext = func(w *window) (context, error) {
		ctx, err := egl.NewContext(nil)
		if err != nil {
			return nil, err
		}
		return &androidContext{win: w, Context: ctx}, nil
	}
}

func (c *androidContext) Release() {
	if c.Context != nil {
		c.Context.Release()
		c.Context = nil
	}
}

func (c *androidContext) Refresh() error {
	c.Context.ReleaseSurface()
	if err := c.win.setVisual(c.Context.VisualID()); err != nil {
		return err
	}
	win, width, height := c.win.nativeWindow()
	c.eglSurf = egl.NativeWindowType(unsafe.Pointer(win))
	c.width, c.height = width, height
	return nil
}

func (c *androidContext) Lock() error {
	// The Android emulator creates a broken surface if it is not
	// created on the same thread as the context is made current.
	if c.eglSurf != nil {
		if err := c.Context.CreateSurface(c.eglSurf, c.width, c.height); err != nil {
			return err
		}
		c.eglSurf = nil
	}
	return c.Context.MakeCurrent()
}

func (c *androidContext) Unlock() {
	c.Context.ReleaseCurrent()
}
