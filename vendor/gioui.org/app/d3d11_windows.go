// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"fmt"
	"unsafe"

	"gioui.org/gpu"
	"gioui.org/internal/d3d11"
)

type d3d11Context struct {
	win *window
	dev *d3d11.Device
	ctx *d3d11.DeviceContext

	swchain       *d3d11.IDXGISwapChain
	renderTarget  *d3d11.RenderTargetView
	width, height int
}

const debug = false

func init() {
	drivers = append(drivers, gpuAPI{
		priority: 1,
		initializer: func(w *window) (context, error) {
			hwnd, _, _ := w.HWND()
			var flags uint32
			if debug {
				flags |= d3d11.CREATE_DEVICE_DEBUG
			}
			dev, ctx, _, err := d3d11.CreateDevice(
				d3d11.DRIVER_TYPE_HARDWARE,
				flags,
			)
			if err != nil {
				return nil, fmt.Errorf("NewContext: %v", err)
			}
			swchain, err := d3d11.CreateSwapChain(dev, hwnd)
			if err != nil {
				d3d11.IUnknownRelease(unsafe.Pointer(ctx), ctx.Vtbl.Release)
				d3d11.IUnknownRelease(unsafe.Pointer(dev), dev.Vtbl.Release)
				return nil, err
			}
			return &d3d11Context{win: w, dev: dev, ctx: ctx, swchain: swchain}, nil
		},
	})
}

func (c *d3d11Context) API() gpu.API {
	return gpu.Direct3D11{Device: unsafe.Pointer(c.dev)}
}

func (c *d3d11Context) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.Direct3D11RenderTarget{
		RenderTarget: unsafe.Pointer(c.renderTarget),
	}, nil
}

func (c *d3d11Context) Present() error {
	err := c.swchain.Present(1, 0)
	if err == nil {
		return nil
	}
	if err, ok := err.(d3d11.ErrorCode); ok {
		switch err.Code {
		case d3d11.DXGI_STATUS_OCCLUDED:
			// Ignore
			return nil
		case d3d11.DXGI_ERROR_DEVICE_RESET, d3d11.DXGI_ERROR_DEVICE_REMOVED, d3d11.D3DDDIERR_DEVICEREMOVED:
			return gpu.ErrDeviceLost
		}
	}
	return err
}

func (c *d3d11Context) Refresh() error {
	var width, height int
	_, width, height = c.win.HWND()
	if c.renderTarget != nil && width == c.width && height == c.height {
		return nil
	}
	c.releaseFBO()
	if err := c.swchain.ResizeBuffers(0, 0, 0, d3d11.DXGI_FORMAT_UNKNOWN, 0); err != nil {
		return err
	}
	c.width = width
	c.height = height

	backBuffer, err := c.swchain.GetBuffer(0, &d3d11.IID_Texture2D)
	if err != nil {
		return err
	}
	texture := (*d3d11.Resource)(unsafe.Pointer(backBuffer))
	renderTarget, err := c.dev.CreateRenderTargetView(texture)
	d3d11.IUnknownRelease(unsafe.Pointer(backBuffer), backBuffer.Vtbl.Release)
	if err != nil {
		return err
	}
	c.renderTarget = renderTarget
	return nil
}

func (c *d3d11Context) Lock() error {
	c.ctx.OMSetRenderTargets(c.renderTarget, nil)
	return nil
}

func (c *d3d11Context) Unlock() {}

func (c *d3d11Context) Release() {
	c.releaseFBO()
	if c.swchain != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(c.swchain), c.swchain.Vtbl.Release)
	}
	if c.ctx != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(c.ctx), c.ctx.Vtbl.Release)
	}
	if c.dev != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(c.dev), c.dev.Vtbl.Release)
	}
	*c = d3d11Context{}
	if debug {
		d3d11.ReportLiveObjects()
	}
}

func (c *d3d11Context) releaseFBO() {
	if c.renderTarget != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(c.renderTarget), c.renderTarget.Vtbl.Release)
		c.renderTarget = nil
	}
}
