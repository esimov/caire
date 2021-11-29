// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"errors"
	"syscall/js"

	"gioui.org/gpu"
	"gioui.org/internal/gl"
)

type glContext struct {
	ctx js.Value
	cnv js.Value
}

func newContext(w *window) (*glContext, error) {
	args := map[string]interface{}{
		// Enable low latency rendering.
		// See https://developers.google.com/web/updates/2019/05/desynchronized.
		"desynchronized":        true,
		"preserveDrawingBuffer": true,
	}
	ctx := w.cnv.Call("getContext", "webgl2", args)
	if ctx.IsNull() {
		ctx = w.cnv.Call("getContext", "webgl", args)
	}
	if ctx.IsNull() {
		return nil, errors.New("app: webgl is not supported")
	}
	c := &glContext{
		ctx: ctx,
		cnv: w.cnv,
	}
	return c, nil
}

func (c *glContext) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.OpenGLRenderTarget{}, nil
}

func (c *glContext) API() gpu.API {
	return gpu.OpenGL{Context: gl.Context(c.ctx)}
}

func (c *glContext) Release() {
}

func (c *glContext) Present() error {
	if c.ctx.Call("isContextLost").Bool() {
		return errors.New("context lost")
	}
	return nil
}

func (c *glContext) Lock() error {
	return nil
}

func (c *glContext) Unlock() {}

func (c *glContext) Refresh() error {
	return nil
}

func (w *window) NewContext() (context, error) {
	return newContext(w)
}
