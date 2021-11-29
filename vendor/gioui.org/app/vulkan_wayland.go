// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland && !novulkan
// +build linux,!android freebsd
// +build !nowayland
// +build !novulkan

package app

import (
	"unsafe"

	"gioui.org/gpu"
	"gioui.org/internal/vk"
)

type wlVkContext struct {
	win  *window
	inst vk.Instance
	surf vk.Surface
	ctx  *vkContext
}

func init() {
	newWaylandVulkanContext = func(w *window) (context, error) {
		inst, err := vk.CreateInstance("VK_KHR_surface", "VK_KHR_wayland_surface")
		if err != nil {
			return nil, err
		}
		disp := w.display()
		wlSurf, _, _ := w.surface()
		surf, err := vk.CreateWaylandSurface(inst, unsafe.Pointer(disp), unsafe.Pointer(wlSurf))
		if err != nil {
			vk.DestroyInstance(inst)
			return nil, err
		}
		ctx, err := newVulkanContext(inst, surf)
		if err != nil {
			vk.DestroySurface(inst, surf)
			vk.DestroyInstance(inst)
			return nil, err
		}
		c := &wlVkContext{
			win:  w,
			inst: inst,
			surf: surf,
			ctx:  ctx,
		}
		return c, nil
	}
}

func (c *wlVkContext) RenderTarget() (gpu.RenderTarget, error) {
	return c.ctx.RenderTarget()
}

func (c *wlVkContext) API() gpu.API {
	return c.ctx.api()
}

func (c *wlVkContext) Release() {
	c.ctx.release()
	vk.DestroySurface(c.inst, c.surf)
	vk.DestroyInstance(c.inst)
	*c = wlVkContext{}
}

func (c *wlVkContext) Present() error {
	return c.ctx.present()
}

func (c *wlVkContext) Lock() error {
	return nil
}

func (c *wlVkContext) Unlock() {}

func (c *wlVkContext) Refresh() error {
	_, w, h := c.win.surface()
	return c.ctx.refresh(c.surf, w, h)
}
