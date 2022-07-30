// SPDX-License-Identifier: Unlicense OR MIT

//go:build !novulkan
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
	newAndroidVulkanContext = func(w *window) (context, error) {
		inst, err := vk.CreateInstance("VK_KHR_surface", "VK_KHR_android_surface")
		if err != nil {
			return nil, err
		}
		window, _, _ := w.nativeWindow()
		surf, err := vk.CreateAndroidSurface(inst, unsafe.Pointer(window))
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
	if c.surf != 0 {
		vk.DestroySurface(c.inst, c.surf)
	}
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
	win, w, h := c.win.nativeWindow()
	if c.surf != 0 {
		c.ctx.destroySwapchain()
		vk.DestroySurface(c.inst, c.surf)
		c.surf = 0
	}
	surf, err := vk.CreateAndroidSurface(c.inst, unsafe.Pointer(win))
	if err != nil {
		return err
	}
	c.surf = surf
	return c.ctx.refresh(c.surf, w, h)
}
