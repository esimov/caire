// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux || freebsd) && !novulkan
// +build linux freebsd
// +build !novulkan

package app

import (
	"errors"
	"unsafe"

	"gioui.org/gpu"
	"gioui.org/internal/vk"
)

type vkContext struct {
	physDev    vk.PhysicalDevice
	inst       vk.Instance
	dev        vk.Device
	queueFam   int
	queue      vk.Queue
	acquireSem vk.Semaphore
	presentSem vk.Semaphore
	fence      vk.Fence

	swchain    vk.Swapchain
	imgs       []vk.Image
	views      []vk.ImageView
	fbos       []vk.Framebuffer
	format     vk.Format
	presentIdx int
}

func newVulkanContext(inst vk.Instance, surf vk.Surface) (*vkContext, error) {
	physDev, qFam, err := vk.ChoosePhysicalDevice(inst, surf)
	if err != nil {
		return nil, err
	}
	dev, err := vk.CreateDeviceAndQueue(physDev, qFam, "VK_KHR_swapchain")
	if err != nil {
		return nil, err
	}
	acquireSem, err := vk.CreateSemaphore(dev)
	if err != nil {
		vk.DestroyDevice(dev)
		return nil, err
	}
	presentSem, err := vk.CreateSemaphore(dev)
	if err != nil {
		vk.DestroySemaphore(dev, acquireSem)
		vk.DestroyDevice(dev)
		return nil, err
	}
	fence, err := vk.CreateFence(dev, vk.FENCE_CREATE_SIGNALED_BIT)
	if err != nil {
		vk.DestroySemaphore(dev, presentSem)
		vk.DestroySemaphore(dev, acquireSem)
		vk.DestroyDevice(dev)
		return nil, err
	}
	c := &vkContext{
		physDev:    physDev,
		inst:       inst,
		dev:        dev,
		queueFam:   qFam,
		queue:      vk.GetDeviceQueue(dev, qFam, 0),
		acquireSem: acquireSem,
		presentSem: presentSem,
		fence:      fence,
	}
	return c, nil
}

func (c *vkContext) RenderTarget() (gpu.RenderTarget, error) {
	vk.WaitForFences(c.dev, c.fence)
	vk.ResetFences(c.dev, c.fence)

	imgIdx, err := vk.AcquireNextImage(c.dev, c.swchain, c.acquireSem, 0)
	if err := mapSurfaceErr(err); err != nil {
		return nil, err
	}
	c.presentIdx = imgIdx
	return gpu.VulkanRenderTarget{
		WaitSem:     uint64(c.acquireSem),
		SignalSem:   uint64(c.presentSem),
		Fence:       uint64(c.fence),
		Framebuffer: uint64(c.fbos[imgIdx]),
		Image:       uint64(c.imgs[imgIdx]),
	}, nil
}

func (c *vkContext) api() gpu.API {
	return gpu.Vulkan{
		PhysDevice:  unsafe.Pointer(c.physDev),
		Device:      unsafe.Pointer(c.dev),
		Format:      int(c.format),
		QueueFamily: c.queueFam,
		QueueIndex:  0,
	}
}

func mapErr(err error) error {
	var vkErr vk.Error
	if errors.As(err, &vkErr) && vkErr == vk.ERROR_DEVICE_LOST {
		return gpu.ErrDeviceLost
	}
	return err
}

func mapSurfaceErr(err error) error {
	var vkErr vk.Error
	if !errors.As(err, &vkErr) {
		return err
	}
	switch {
	case vkErr == vk.SUBOPTIMAL_KHR:
		// Android reports VK_SUBOPTIMAL_KHR when presenting to a rotated
		// swapchain (preTransform != currentTransform). However, we don't
		// support transforming the output ourselves, so we'll live with it.
		return nil
	case vkErr == vk.ERROR_OUT_OF_DATE_KHR:
		return errOutOfDate
	case vkErr == vk.ERROR_SURFACE_LOST_KHR:
		// Treating a lost surface as a lost device isn't accurate, but
		// probably not worth optimizing.
		return gpu.ErrDeviceLost
	}
	return mapErr(err)
}

func (c *vkContext) release() {
	vk.DeviceWaitIdle(c.dev)

	c.destroySwapchain()
	vk.DestroyFence(c.dev, c.fence)
	vk.DestroySemaphore(c.dev, c.acquireSem)
	vk.DestroySemaphore(c.dev, c.presentSem)
	vk.DestroyDevice(c.dev)
	*c = vkContext{}
}

func (c *vkContext) present() error {
	return mapSurfaceErr(vk.PresentQueue(c.queue, c.swchain, c.presentSem, c.presentIdx))
}

func (c *vkContext) destroyImageViews() {
	for _, f := range c.fbos {
		vk.DestroyFramebuffer(c.dev, f)
	}
	c.fbos = nil
	for _, view := range c.views {
		vk.DestroyImageView(c.dev, view)
	}
	c.views = nil
}

func (c *vkContext) destroySwapchain() {
	vk.DeviceWaitIdle(c.dev)

	c.destroyImageViews()
	if c.swchain != 0 {
		vk.DestroySwapchain(c.dev, c.swchain)
		c.swchain = 0
	}
}

func (c *vkContext) refresh(surf vk.Surface, width, height int) error {
	vk.DeviceWaitIdle(c.dev)

	c.destroyImageViews()
	// Check whether size is valid. That's needed on X11, where ConfigureNotify
	// is not always synchronized with the window extent.
	caps, err := vk.GetPhysicalDeviceSurfaceCapabilities(c.physDev, surf)
	if err != nil {
		return err
	}
	minExt, maxExt := caps.MinExtent(), caps.MaxExtent()
	if width < minExt.X || maxExt.X < width || height < minExt.Y || maxExt.Y < height {
		return errOutOfDate
	}
	swchain, imgs, format, err := vk.CreateSwapchain(c.physDev, c.dev, surf, width, height, c.swchain)
	if c.swchain != 0 {
		vk.DestroySwapchain(c.dev, c.swchain)
		c.swchain = 0
	}
	if err := mapSurfaceErr(err); err != nil {
		return err
	}
	c.swchain = swchain
	c.imgs = imgs
	c.format = format
	pass, err := vk.CreateRenderPass(
		c.dev,
		format,
		vk.ATTACHMENT_LOAD_OP_CLEAR,
		vk.IMAGE_LAYOUT_UNDEFINED,
		vk.IMAGE_LAYOUT_PRESENT_SRC_KHR,
		nil,
	)
	if err := mapErr(err); err != nil {
		return err
	}
	defer vk.DestroyRenderPass(c.dev, pass)
	for _, img := range imgs {
		view, err := vk.CreateImageView(c.dev, img, format)
		if err := mapErr(err); err != nil {
			return err
		}
		c.views = append(c.views, view)
		fbo, err := vk.CreateFramebuffer(c.dev, pass, view, width, height)
		if err := mapErr(err); err != nil {
			return err
		}
		c.fbos = append(c.fbos, fbo)
	}
	return nil
}
