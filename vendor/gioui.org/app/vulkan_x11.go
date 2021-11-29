// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nox11 && !novulkan
// +build linux,!android freebsd
// +build !nox11
// +build !novulkan

package app

import (
	"unsafe"

	"gioui.org/gpu"
	"gioui.org/internal/vk"
)

type x11VkContext struct {
	win  *x11Window
	inst vk.Instance
	surf vk.Surface
	ctx  *vkContext
}

func init() {
	newX11VulkanContext = func(w *x11Window) (context, error) {
		inst, err := vk.CreateInstance("VK_KHR_surface", "VK_KHR_xlib_surface")
		if err != nil {
			return nil, err
		}
		disp := w.display()
		window, _, _ := w.window()
		surf, err := vk.CreateXlibSurface(inst, unsafe.Pointer(disp), uintptr(window))
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
		c := &x11VkContext{
			win:  w,
			inst: inst,
			surf: surf,
			ctx:  ctx,
		}
		return c, nil
	}
}

func (c *x11VkContext) RenderTarget() (gpu.RenderTarget, error) {
	return c.ctx.RenderTarget()
}

func (c *x11VkContext) API() gpu.API {
	return c.ctx.api()
}

func (c *x11VkContext) Release() {
	c.ctx.release()
	vk.DestroySurface(c.inst, c.surf)
	vk.DestroyInstance(c.inst)
	*c = x11VkContext{}
}

func (c *x11VkContext) Present() error {
	return c.ctx.present()
}

func (c *x11VkContext) Lock() error {
	return nil
}

func (c *x11VkContext) Unlock() {}

func (c *x11VkContext) Refresh() error {
	_, w, h := c.win.window()
	return c.ctx.refresh(c.surf, w, h)
}
