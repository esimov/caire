// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !nox11
// +build linux,!android freebsd openbsd
// +build !nox11

package app

import (
	"unsafe"

	"gioui.org/internal/egl"
)

type x11Context struct {
	win *x11Window
	*egl.Context
}

func init() {
	newX11EGLContext = func(w *x11Window) (context, error) {
		disp := egl.NativeDisplayType(unsafe.Pointer(w.display()))
		ctx, err := egl.NewContext(disp)
		if err != nil {
			return nil, err
		}
		return &x11Context{win: w, Context: ctx}, nil
	}
}

func (c *x11Context) Release() {
	if c.Context != nil {
		c.Context.Release()
		c.Context = nil
	}
}

func (c *x11Context) Refresh() error {
	c.Context.ReleaseSurface()
	win, width, height := c.win.window()
	eglSurf := egl.NativeWindowType(uintptr(win))
	if err := c.Context.CreateSurface(eglSurf, width, height); err != nil {
		return err
	}
	if err := c.Context.MakeCurrent(); err != nil {
		return err
	}
	c.Context.EnableVSync(true)
	c.Context.ReleaseCurrent()
	return nil
}

func (c *x11Context) Lock() error {
	return c.Context.MakeCurrent()
}

func (c *x11Context) Unlock() {
	c.Context.ReleaseCurrent()
}
