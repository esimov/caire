// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nox11
// +build linux,!android freebsd
// +build !nox11

package vk

/*
#define VK_USE_PLATFORM_XLIB_KHR
#define VK_NO_PROTOTYPES 1
#define VK_DEFINE_NON_DISPATCHABLE_HANDLE(object) typedef uint64_t object;
#include <vulkan/vulkan.h>

static VkResult vkCreateXlibSurfaceKHR(PFN_vkCreateXlibSurfaceKHR f, VkInstance instance, const VkXlibSurfaceCreateInfoKHR *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSurfaceKHR *pSurface) {
	return f(instance, pCreateInfo, pAllocator, pSurface);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

var x11Funcs struct {
	vkCreateXlibSurfaceKHR C.PFN_vkCreateXlibSurfaceKHR
}

func init() {
	loadFuncs = append(loadFuncs, func(dlopen func(name string) *[0]byte) {
		x11Funcs.vkCreateXlibSurfaceKHR = dlopen("vkCreateXlibSurfaceKHR")
	})
}

func CreateXlibSurface(inst Instance, dpy unsafe.Pointer, window uintptr) (Surface, error) {
	inf := C.VkXlibSurfaceCreateInfoKHR{
		sType:  C.VK_STRUCTURE_TYPE_XLIB_SURFACE_CREATE_INFO_KHR,
		dpy:    (*C.Display)(dpy),
		window: (C.Window)(window),
	}
	var surf Surface
	if err := vkErr(C.vkCreateXlibSurfaceKHR(x11Funcs.vkCreateXlibSurfaceKHR, inst, &inf, nil, &surf)); err != nil {
		return 0, fmt.Errorf("vulkan: vkCreateXlibSurfaceKHR: %w", err)
	}
	return surf, nil
}
