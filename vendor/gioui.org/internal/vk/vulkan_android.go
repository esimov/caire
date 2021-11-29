// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nowayland
// +build !nowayland

package vk

/*
#define VK_USE_PLATFORM_ANDROID_KHR
#define VK_NO_PROTOTYPES 1
#define VK_DEFINE_NON_DISPATCHABLE_HANDLE(object) typedef uint64_t object;
#include <android/native_window.h>
#include <vulkan/vulkan.h>

static VkResult vkCreateAndroidSurfaceKHR(PFN_vkCreateAndroidSurfaceKHR f, VkInstance instance, const VkAndroidSurfaceCreateInfoKHR *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSurfaceKHR *pSurface) {
	return f(instance, pCreateInfo, pAllocator, pSurface);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

var wlFuncs struct {
	vkCreateAndroidSurfaceKHR C.PFN_vkCreateAndroidSurfaceKHR
}

func init() {
	loadFuncs = append(loadFuncs, func(dlopen func(name string) *[0]byte) {
		wlFuncs.vkCreateAndroidSurfaceKHR = dlopen("vkCreateAndroidSurfaceKHR")
	})
}

func CreateAndroidSurface(inst Instance, window unsafe.Pointer) (Surface, error) {
	inf := C.VkAndroidSurfaceCreateInfoKHR{
		sType:  C.VK_STRUCTURE_TYPE_ANDROID_SURFACE_CREATE_INFO_KHR,
		window: (*C.ANativeWindow)(window),
	}
	var surf Surface
	if err := vkErr(C.vkCreateAndroidSurfaceKHR(wlFuncs.vkCreateAndroidSurfaceKHR, inst, &inf, nil, &surf)); err != nil {
		return 0, fmt.Errorf("vulkan: vkCreateAndroidSurfaceKHR: %w", err)
	}
	return surf, nil
}
