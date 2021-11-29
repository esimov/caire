// SPDX-License-Identifier: Unlicense OR MIT

package driver

import (
	"fmt"
	"unsafe"

	"gioui.org/internal/gl"
)

// See gpu/api.go for documentation for the API types.

type API interface {
	implementsAPI()
}

type RenderTarget interface {
	ImplementsRenderTarget()
}

type OpenGLRenderTarget gl.Framebuffer

type Direct3D11RenderTarget struct {
	// RenderTarget is a *ID3D11RenderTargetView.
	RenderTarget unsafe.Pointer
}

type MetalRenderTarget struct {
	// Texture is a MTLTexture.
	Texture uintptr
}

type VulkanRenderTarget struct {
	// WaitSem is a VkSemaphore that must signaled before accessing Framebuffer.
	WaitSem uint64
	// SignalSem is a VkSemaphore that signal access to Framebuffer is complete.
	SignalSem uint64
	// Image is the VkImage to render into.
	Image uint64
	// Framebuffer is a VkFramebuffer for Image.
	Framebuffer uint64
}

type OpenGL struct {
	// ES forces the use of ANGLE OpenGL ES libraries on macOS. It is
	// ignored on all other platforms.
	ES bool
	// Context contains the WebGL context for WebAssembly platforms. It is
	// empty for all other platforms; an OpenGL context is assumed current when
	// calling NewDevice.
	Context gl.Context
	// Shared instructs users of the context to restore the GL state after
	// use.
	Shared bool
}

type Direct3D11 struct {
	// Device contains a *ID3D11Device.
	Device unsafe.Pointer
}

type Metal struct {
	// Device is an MTLDevice.
	Device uintptr
	// Queue is a MTLCommandQueue.
	Queue uintptr
	// PixelFormat is the MTLPixelFormat of the default framebuffer.
	PixelFormat int
}

type Vulkan struct {
	// PhysDevice is a VkPhysicalDevice.
	PhysDevice unsafe.Pointer
	// Device is a VkDevice.
	Device unsafe.Pointer
	// QueueFamily is the queue familily index of the queue.
	QueueFamily int
	// QueueIndex is the logical queue index of the queue.
	QueueIndex int
	// Format is a VkFormat that matches render targets.
	Format int
}

// API specific device constructors.
var (
	NewOpenGLDevice     func(api OpenGL) (Device, error)
	NewDirect3D11Device func(api Direct3D11) (Device, error)
	NewMetalDevice      func(api Metal) (Device, error)
	NewVulkanDevice     func(api Vulkan) (Device, error)
)

// NewDevice creates a new Device given the api.
//
// Note that the device does not assume ownership of the resources contained in
// api; the caller must ensure the resources are valid until the device is
// released.
func NewDevice(api API) (Device, error) {
	switch api := api.(type) {
	case OpenGL:
		if NewOpenGLDevice != nil {
			return NewOpenGLDevice(api)
		}
	case Direct3D11:
		if NewDirect3D11Device != nil {
			return NewDirect3D11Device(api)
		}
	case Metal:
		if NewMetalDevice != nil {
			return NewMetalDevice(api)
		}
	case Vulkan:
		if NewVulkanDevice != nil {
			return NewVulkanDevice(api)
		}
	}
	return nil, fmt.Errorf("driver: no driver available for the API %T", api)
}

func (OpenGL) implementsAPI()                          {}
func (Direct3D11) implementsAPI()                      {}
func (Metal) implementsAPI()                           {}
func (Vulkan) implementsAPI()                          {}
func (OpenGLRenderTarget) ImplementsRenderTarget()     {}
func (Direct3D11RenderTarget) ImplementsRenderTarget() {}
func (MetalRenderTarget) ImplementsRenderTarget()      {}
func (VulkanRenderTarget) ImplementsRenderTarget()     {}
