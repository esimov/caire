// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland
// +build linux,!android freebsd
// +build !nowayland

package vk

/*
#define VK_USE_PLATFORM_WAYLAND_KHR
#define VK_NO_PROTOTYPES 1
#define VK_DEFINE_NON_DISPATCHABLE_HANDLE(object) typedef uint64_t object;
#include <vulkan/vulkan.h>

static VkResult vkCreateWaylandSurfaceKHR(PFN_vkCreateWaylandSurfaceKHR f, VkInstance instance, const VkWaylandSurfaceCreateInfoKHR *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSurfaceKHR *pSurface) {
	return f(instance, pCreateInfo, pAllocator, pSurface);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

var wlFuncs struct {
	vkCreateWaylandSurfaceKHR C.PFN_vkCreateWaylandSurfaceKHR
}

func init() {
	loadFuncs = append(loadFuncs, func(dlopen func(name string) *[0]byte) {
		wlFuncs.vkCreateWaylandSurfaceKHR = dlopen("vkCreateWaylandSurfaceKHR")
	})
}

func CreateWaylandSurface(inst Instance, disp unsafe.Pointer, wlSurf unsafe.Pointer) (Surface, error) {
	inf := C.VkWaylandSurfaceCreateInfoKHR{
		sType:   C.VK_STRUCTURE_TYPE_WAYLAND_SURFACE_CREATE_INFO_KHR,
		display: (*C.struct_wl_display)(disp),
		surface: (*C.struct_wl_surface)(wlSurf),
	}
	var surf Surface
	if err := vkErr(C.vkCreateWaylandSurfaceKHR(wlFuncs.vkCreateWaylandSurfaceKHR, inst, &inf, nil, &surf)); err != nil {
		return 0, fmt.Errorf("vulkan: vkCreateWaylandSurfaceKHR: %w", err)
	}
	return surf, nil
}
