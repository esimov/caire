// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"golang.org/x/sys/windows"

	"gioui.org/internal/egl"
)

type glContext struct {
	win *window
	*egl.Context
}

func init() {
	drivers = append(drivers, gpuAPI{
		priority: 2,
		initializer: func(w *window) (context, error) {
			disp := egl.NativeDisplayType(w.HDC())
			ctx, err := egl.NewContext(disp)
			if err != nil {
				return nil, err
			}
			return &glContext{win: w, Context: ctx}, nil
		},
	})
}

func (c *glContext) Release() {
	if c.Context != nil {
		c.Context.Release()
		c.Context = nil
	}
}

func (c *glContext) Refresh() error {
	c.Context.ReleaseSurface()
	var (
		win           windows.Handle
		width, height int
	)
	win, width, height = c.win.HWND()
	eglSurf := egl.NativeWindowType(win)
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

func (c *glContext) Lock() error {
	return c.Context.MakeCurrent()
}

func (c *glContext) Unlock() {
	c.Context.ReleaseCurrent()
}
