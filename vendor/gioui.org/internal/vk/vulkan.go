// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux || freebsd
// +build linux freebsd

package vk

/*
#cgo linux freebsd LDFLAGS: -ldl
#cgo freebsd CFLAGS: -I/usr/local/include
#cgo CFLAGS: -Werror -Werror=return-type

#define VK_NO_PROTOTYPES 1
#define VK_DEFINE_NON_DISPATCHABLE_HANDLE(object) typedef uint64_t object;
#include <vulkan/vulkan.h>
#define __USE_GNU
#include <dlfcn.h>
#include <stdlib.h>

static VkResult vkCreateInstance(PFN_vkCreateInstance f, VkInstanceCreateInfo pCreateInfo, const VkAllocationCallbacks *pAllocator, VkInstance *pInstance) {
	return f(&pCreateInfo, pAllocator, pInstance);
}

static void vkDestroyInstance(PFN_vkDestroyInstance f, VkInstance instance, const VkAllocationCallbacks *pAllocator) {
	f(instance, pAllocator);
}

static VkResult vkEnumeratePhysicalDevices(PFN_vkEnumeratePhysicalDevices f, VkInstance instance, uint32_t *pPhysicalDeviceCount, VkPhysicalDevice *pPhysicalDevices) {
	return f(instance, pPhysicalDeviceCount, pPhysicalDevices);
}

static void vkGetPhysicalDeviceQueueFamilyProperties(PFN_vkGetPhysicalDeviceQueueFamilyProperties f, VkPhysicalDevice physicalDevice, uint32_t *pQueueFamilyPropertyCount, VkQueueFamilyProperties *pQueueFamilyProperties) {
	f(physicalDevice, pQueueFamilyPropertyCount, pQueueFamilyProperties);
}

static void vkGetPhysicalDeviceFormatProperties(PFN_vkGetPhysicalDeviceFormatProperties f, VkPhysicalDevice physicalDevice, VkFormat format, VkFormatProperties *pFormatProperties) {
	f(physicalDevice, format, pFormatProperties);
}

static VkResult vkCreateDevice(PFN_vkCreateDevice f, VkPhysicalDevice physicalDevice, VkDeviceCreateInfo pCreateInfo, VkDeviceQueueCreateInfo qinf, const VkAllocationCallbacks *pAllocator, VkDevice *pDevice) {
	pCreateInfo.pQueueCreateInfos = &qinf;
	return f(physicalDevice, &pCreateInfo, pAllocator, pDevice);
}

static void vkDestroyDevice(PFN_vkDestroyDevice f, VkDevice device, const VkAllocationCallbacks *pAllocator) {
	f(device, pAllocator);
}

static void vkGetDeviceQueue(PFN_vkGetDeviceQueue f, VkDevice device, uint32_t queueFamilyIndex, uint32_t queueIndex, VkQueue *pQueue) {
	f(device, queueFamilyIndex, queueIndex, pQueue);
}

static VkResult vkCreateImageView(PFN_vkCreateImageView f, VkDevice device, const VkImageViewCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkImageView *pView) {
	return f(device, pCreateInfo, pAllocator, pView);
}

static void vkDestroyImageView(PFN_vkDestroyImageView f, VkDevice device, VkImageView imageView, const VkAllocationCallbacks *pAllocator) {
	f(device, imageView, pAllocator);
}

static VkResult vkCreateFramebuffer(PFN_vkCreateFramebuffer f, VkDevice device, VkFramebufferCreateInfo pCreateInfo, const VkAllocationCallbacks *pAllocator, VkFramebuffer *pFramebuffer) {
	return f(device, &pCreateInfo, pAllocator, pFramebuffer);
}

static void vkDestroyFramebuffer(PFN_vkDestroyFramebuffer f, VkDevice device, VkFramebuffer framebuffer, const VkAllocationCallbacks *pAllocator) {
	f(device, framebuffer, pAllocator);
}

static VkResult vkDeviceWaitIdle(PFN_vkDeviceWaitIdle f, VkDevice device) {
	return f(device);
}

static VkResult vkQueueWaitIdle(PFN_vkQueueWaitIdle f, VkQueue queue) {
	return f(queue);
}

static VkResult vkCreateSemaphore(PFN_vkCreateSemaphore f, VkDevice device, const VkSemaphoreCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSemaphore *pSemaphore) {
	return f(device, pCreateInfo, pAllocator, pSemaphore);
}

static void vkDestroySemaphore(PFN_vkDestroySemaphore f, VkDevice device, VkSemaphore semaphore, const VkAllocationCallbacks *pAllocator) {
	f(device, semaphore, pAllocator);
}

static VkResult vkCreateRenderPass(PFN_vkCreateRenderPass f, VkDevice device, VkRenderPassCreateInfo pCreateInfo, VkSubpassDescription subpassInf, const VkAllocationCallbacks *pAllocator, VkRenderPass *pRenderPass) {
	pCreateInfo.pSubpasses = &subpassInf;
	return f(device, &pCreateInfo, pAllocator, pRenderPass);
}

static void vkDestroyRenderPass(PFN_vkDestroyRenderPass f, VkDevice device, VkRenderPass renderPass, const VkAllocationCallbacks *pAllocator) {
	f(device, renderPass, pAllocator);
}

static VkResult vkCreateCommandPool(PFN_vkCreateCommandPool f, VkDevice device, const VkCommandPoolCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkCommandPool *pCommandPool) {
	return f(device, pCreateInfo, pAllocator, pCommandPool);
}

static void vkDestroyCommandPool(PFN_vkDestroyCommandPool f, VkDevice device, VkCommandPool commandPool, const VkAllocationCallbacks *pAllocator) {
	f(device, commandPool, pAllocator);
}

static VkResult vkAllocateCommandBuffers(PFN_vkAllocateCommandBuffers f, VkDevice device, const VkCommandBufferAllocateInfo *pAllocateInfo, VkCommandBuffer *pCommandBuffers) {
	return f(device, pAllocateInfo, pCommandBuffers);
}

static void vkFreeCommandBuffers(PFN_vkFreeCommandBuffers f, VkDevice  device, VkCommandPool commandPool, uint32_t commandBufferCount, const VkCommandBuffer *pCommandBuffers) {
	f(device, commandPool, commandBufferCount, pCommandBuffers);
}

static VkResult vkBeginCommandBuffer(PFN_vkBeginCommandBuffer f, VkCommandBuffer commandBuffer, VkCommandBufferBeginInfo pBeginInfo) {
	return f(commandBuffer, &pBeginInfo);
}

static VkResult vkEndCommandBuffer(PFN_vkEndCommandBuffer f, VkCommandBuffer commandBuffer) {
	return f(commandBuffer);
}

static VkResult vkQueueSubmit(PFN_vkQueueSubmit f, VkQueue queue, VkSubmitInfo pSubmits, VkFence fence) {
	return f(queue, 1, &pSubmits, fence);
}

static void vkCmdBeginRenderPass(PFN_vkCmdBeginRenderPass f, VkCommandBuffer commandBuffer, VkRenderPassBeginInfo pRenderPassBegin, VkSubpassContents contents) {
	f(commandBuffer, &pRenderPassBegin, contents);
}

static void vkCmdEndRenderPass(PFN_vkCmdEndRenderPass f, VkCommandBuffer commandBuffer) {
	f(commandBuffer);
}

static void vkCmdCopyBuffer(PFN_vkCmdCopyBuffer f, VkCommandBuffer commandBuffer, VkBuffer srcBuffer, VkBuffer dstBuffer, uint32_t regionCount, const VkBufferCopy *pRegions) {
	f(commandBuffer, srcBuffer, dstBuffer, regionCount, pRegions);
}

static void vkCmdCopyBufferToImage(PFN_vkCmdCopyBufferToImage f, VkCommandBuffer commandBuffer, VkBuffer srcBuffer, VkImage dstImage, VkImageLayout dstImageLayout, uint32_t regionCount, const VkBufferImageCopy *pRegions) {
	f(commandBuffer, srcBuffer, dstImage, dstImageLayout, regionCount, pRegions);
}

static void vkCmdPipelineBarrier(PFN_vkCmdPipelineBarrier f, VkCommandBuffer commandBuffer, VkPipelineStageFlags srcStageMask, VkPipelineStageFlags dstStageMask, VkDependencyFlags dependencyFlags, uint32_t memoryBarrierCount, const VkMemoryBarrier *pMemoryBarriers, uint32_t bufferMemoryBarrierCount, const VkBufferMemoryBarrier *pBufferMemoryBarriers, uint32_t imageMemoryBarrierCount, const VkImageMemoryBarrier *pImageMemoryBarriers) {
	f(commandBuffer, srcStageMask, dstStageMask, dependencyFlags, memoryBarrierCount, pMemoryBarriers, bufferMemoryBarrierCount, pBufferMemoryBarriers, imageMemoryBarrierCount, pImageMemoryBarriers);
}

static void vkCmdPushConstants(PFN_vkCmdPushConstants f, VkCommandBuffer commandBuffer, VkPipelineLayout layout, VkShaderStageFlags stageFlags, uint32_t offset, uint32_t size, const void *pValues) {
	f(commandBuffer, layout, stageFlags, offset, size, pValues);
}

static void vkCmdBindPipeline(PFN_vkCmdBindPipeline f, VkCommandBuffer commandBuffer, VkPipelineBindPoint pipelineBindPoint, VkPipeline pipeline) {
	f(commandBuffer, pipelineBindPoint, pipeline);
}

static void vkCmdBindVertexBuffers(PFN_vkCmdBindVertexBuffers f, VkCommandBuffer commandBuffer, uint32_t firstBinding, uint32_t bindingCount, const VkBuffer *pBuffers, const VkDeviceSize *pOffsets) {
	f(commandBuffer, firstBinding, bindingCount, pBuffers, pOffsets);
}

static void vkCmdSetViewport(PFN_vkCmdSetViewport f, VkCommandBuffer commandBuffer, uint32_t firstViewport, uint32_t viewportCount, const VkViewport *pViewports) {
	f(commandBuffer, firstViewport, viewportCount, pViewports);
}

static void vkCmdBindIndexBuffer(PFN_vkCmdBindIndexBuffer f, VkCommandBuffer commandBuffer, VkBuffer buffer, VkDeviceSize offset, VkIndexType indexType) {
	f(commandBuffer, buffer, offset, indexType);
}

static void vkCmdDraw(PFN_vkCmdDraw f, VkCommandBuffer commandBuffer, uint32_t vertexCount, uint32_t instanceCount, uint32_t firstVertex, uint32_t firstInstance) {
	f(commandBuffer, vertexCount, instanceCount, firstVertex, firstInstance);
}

static void vkCmdDrawIndexed(PFN_vkCmdDrawIndexed f, VkCommandBuffer commandBuffer, uint32_t indexCount, uint32_t instanceCount, uint32_t firstIndex, int32_t vertexOffset, uint32_t firstInstance) {
	f(commandBuffer, indexCount, instanceCount, firstIndex, vertexOffset, firstInstance);
}

static void vkCmdBindDescriptorSets(PFN_vkCmdBindDescriptorSets f, VkCommandBuffer commandBuffer, VkPipelineBindPoint pipelineBindPoint, VkPipelineLayout layout, uint32_t firstSet, uint32_t descriptorSetCount, const VkDescriptorSet *pDescriptorSets, uint32_t dynamicOffsetCount, const uint32_t *pDynamicOffsets) {
	f(commandBuffer, pipelineBindPoint, layout, firstSet, descriptorSetCount, pDescriptorSets, dynamicOffsetCount, pDynamicOffsets);
}

static void vkCmdCopyImageToBuffer(PFN_vkCmdCopyImageToBuffer f, VkCommandBuffer commandBuffer, VkImage srcImage, VkImageLayout srcImageLayout, VkBuffer dstBuffer, uint32_t regionCount, const VkBufferImageCopy *pRegions) {
	f(commandBuffer, srcImage, srcImageLayout, dstBuffer, regionCount, pRegions);
}

static void vkCmdDispatch(PFN_vkCmdDispatch f, VkCommandBuffer commandBuffer, uint32_t groupCountX, uint32_t groupCountY, uint32_t groupCountZ) {
	f(commandBuffer, groupCountX, groupCountY, groupCountZ);
}

static VkResult vkCreateImage(PFN_vkCreateImage f, VkDevice device, const VkImageCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkImage *pImage) {
	return f(device, pCreateInfo, pAllocator, pImage);
}

static void vkDestroyImage(PFN_vkDestroyImage f, VkDevice device, VkImage image, const VkAllocationCallbacks *pAllocator) {
	f(device, image, pAllocator);
}

static void vkGetImageMemoryRequirements(PFN_vkGetImageMemoryRequirements f, VkDevice device, VkImage image, VkMemoryRequirements *pMemoryRequirements) {
	f(device, image, pMemoryRequirements);
}

static VkResult vkAllocateMemory(PFN_vkAllocateMemory f, VkDevice device, const VkMemoryAllocateInfo *pAllocateInfo, const VkAllocationCallbacks *pAllocator, VkDeviceMemory *pMemory) {
	return f(device, pAllocateInfo, pAllocator, pMemory);
}

static VkResult vkBindImageMemory(PFN_vkBindImageMemory f, VkDevice device, VkImage image, VkDeviceMemory memory, VkDeviceSize memoryOffset) {
	return f(device, image, memory, memoryOffset);
}

static void vkFreeMemory(PFN_vkFreeMemory f, VkDevice device, VkDeviceMemory memory, const VkAllocationCallbacks *pAllocator) {
	f(device, memory, pAllocator);
}

static void vkGetPhysicalDeviceMemoryProperties(PFN_vkGetPhysicalDeviceMemoryProperties f, VkPhysicalDevice physicalDevice, VkPhysicalDeviceMemoryProperties *pMemoryProperties) {
	f(physicalDevice, pMemoryProperties);
}

static VkResult vkCreateSampler(PFN_vkCreateSampler f,VkDevice device, const VkSamplerCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSampler *pSampler) {
	return f(device, pCreateInfo, pAllocator, pSampler);
}

static void vkDestroySampler(PFN_vkDestroySampler f, VkDevice device, VkSampler sampler, const VkAllocationCallbacks *pAllocator) {
	f(device, sampler, pAllocator);
}

static VkResult vkCreateBuffer(PFN_vkCreateBuffer f, VkDevice device, const VkBufferCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkBuffer *pBuffer) {
	return f(device, pCreateInfo, pAllocator, pBuffer);
}

static void vkDestroyBuffer(PFN_vkDestroyBuffer f, VkDevice device, VkBuffer buffer, const VkAllocationCallbacks *pAllocator) {
	f(device, buffer, pAllocator);
}

static void vkGetBufferMemoryRequirements(PFN_vkGetBufferMemoryRequirements f, VkDevice device, VkBuffer buffer, VkMemoryRequirements *pMemoryRequirements) {
	f(device, buffer, pMemoryRequirements);
}

static VkResult vkBindBufferMemory(PFN_vkBindBufferMemory f, VkDevice device, VkBuffer buffer, VkDeviceMemory memory, VkDeviceSize memoryOffset) {
	return f(device, buffer, memory, memoryOffset);
}

static VkResult vkCreateShaderModule(PFN_vkCreateShaderModule f, VkDevice device, VkShaderModuleCreateInfo pCreateInfo, const VkAllocationCallbacks *pAllocator, VkShaderModule *pShaderModule) {
	return f(device, &pCreateInfo, pAllocator, pShaderModule);
}

static void vkDestroyShaderModule(PFN_vkDestroyShaderModule f, VkDevice device, VkShaderModule shaderModule, const VkAllocationCallbacks *pAllocator) {
	f(device, shaderModule, pAllocator);
}

static VkResult vkCreateGraphicsPipelines(PFN_vkCreateGraphicsPipelines f, VkDevice device, VkPipelineCache pipelineCache, VkGraphicsPipelineCreateInfo pCreateInfo, VkPipelineDynamicStateCreateInfo dynInf, VkPipelineColorBlendStateCreateInfo blendInf, VkPipelineVertexInputStateCreateInfo vertexInf, VkPipelineViewportStateCreateInfo viewportInf, const VkAllocationCallbacks *pAllocator, VkPipeline *pPipelines) {
	pCreateInfo.pDynamicState = &dynInf;
	pCreateInfo.pViewportState = &viewportInf;
	pCreateInfo.pColorBlendState = &blendInf;
	pCreateInfo.pVertexInputState = &vertexInf;
	return f(device, pipelineCache, 1, &pCreateInfo, pAllocator, pPipelines);
}

static void vkDestroyPipeline(PFN_vkDestroyPipeline f, VkDevice device, VkPipeline pipeline, const VkAllocationCallbacks *pAllocator) {
	f(device, pipeline, pAllocator);
}

static VkResult vkCreatePipelineLayout(PFN_vkCreatePipelineLayout f, VkDevice device, VkPipelineLayoutCreateInfo pCreateInfo, const VkAllocationCallbacks *pAllocator, VkPipelineLayout *pPipelineLayout) {
	return f(device, &pCreateInfo, pAllocator, pPipelineLayout);
}

static void vkDestroyPipelineLayout(PFN_vkDestroyPipelineLayout f, VkDevice device, VkPipelineLayout pipelineLayout, const VkAllocationCallbacks *pAllocator) {
	f(device, pipelineLayout, pAllocator);
}

static VkResult vkCreateDescriptorSetLayout(PFN_vkCreateDescriptorSetLayout f, VkDevice device, VkDescriptorSetLayoutCreateInfo pCreateInfo, const VkAllocationCallbacks *pAllocator, VkDescriptorSetLayout *pSetLayout) {
	return f(device, &pCreateInfo, pAllocator, pSetLayout);
}

static void vkDestroyDescriptorSetLayout(PFN_vkDestroyDescriptorSetLayout f, VkDevice device, VkDescriptorSetLayout descriptorSetLayout, const VkAllocationCallbacks *pAllocator) {
	f(device, descriptorSetLayout, pAllocator);
}

static VkResult vkMapMemory(PFN_vkMapMemory f, VkDevice device, VkDeviceMemory memory, VkDeviceSize offset, VkDeviceSize size, VkMemoryMapFlags flags, void **ppData) {
	return f(device, memory, offset, size, flags, ppData);
}

static void vkUnmapMemory(PFN_vkUnmapMemory f, VkDevice device, VkDeviceMemory memory) {
	f(device, memory);
}

static VkResult vkResetCommandBuffer(PFN_vkResetCommandBuffer f, VkCommandBuffer commandBuffer, VkCommandBufferResetFlags flags) {
	return f(commandBuffer, flags);
}

static VkResult vkCreateDescriptorPool(PFN_vkCreateDescriptorPool f, VkDevice device, VkDescriptorPoolCreateInfo pCreateInfo, const VkAllocationCallbacks *pAllocator, VkDescriptorPool *pDescriptorPool) {
	return f(device, &pCreateInfo, pAllocator, pDescriptorPool);
}

static void vkDestroyDescriptorPool(PFN_vkDestroyDescriptorPool f, VkDevice device, VkDescriptorPool descriptorPool, const VkAllocationCallbacks *pAllocator) {
	f(device, descriptorPool, pAllocator);
}

static VkResult vkAllocateDescriptorSets(PFN_vkAllocateDescriptorSets f, VkDevice device, VkDescriptorSetAllocateInfo pAllocateInfo, VkDescriptorSet *pDescriptorSets) {
	return f(device, &pAllocateInfo, pDescriptorSets);
}

static VkResult vkFreeDescriptorSets(PFN_vkFreeDescriptorSets f, VkDevice device, VkDescriptorPool descriptorPool, uint32_t descriptorSetCount, const VkDescriptorSet *pDescriptorSets) {
	return f(device, descriptorPool, descriptorSetCount, pDescriptorSets);
}

static void vkUpdateDescriptorSets(PFN_vkUpdateDescriptorSets f, VkDevice device, VkWriteDescriptorSet pDescriptorWrite, uint32_t descriptorCopyCount, const VkCopyDescriptorSet *pDescriptorCopies) {
	f(device, 1, &pDescriptorWrite, descriptorCopyCount, pDescriptorCopies);
}

static VkResult vkResetDescriptorPool(PFN_vkResetDescriptorPool f, VkDevice device, VkDescriptorPool descriptorPool, VkDescriptorPoolResetFlags flags) {
	return f(device, descriptorPool, flags);
}

static void vkCmdBlitImage(PFN_vkCmdBlitImage f, VkCommandBuffer commandBuffer, VkImage srcImage, VkImageLayout srcImageLayout, VkImage dstImage, VkImageLayout dstImageLayout, uint32_t regionCount, const VkImageBlit* pRegions, VkFilter filter) {
	f(commandBuffer, srcImage, srcImageLayout, dstImage, dstImageLayout, regionCount, pRegions, filter);
}

static void vkCmdCopyImage(PFN_vkCmdCopyImage f, VkCommandBuffer commandBuffer, VkImage srcImage, VkImageLayout srcImageLayout, VkImage dstImage, VkImageLayout dstImageLayout, uint32_t regionCount, const VkImageCopy *pRegions) {
	f(commandBuffer, srcImage, srcImageLayout, dstImage, dstImageLayout, regionCount, pRegions);
}

static VkResult vkCreateComputePipelines(PFN_vkCreateComputePipelines f, VkDevice device, VkPipelineCache pipelineCache, uint32_t createInfoCount, const VkComputePipelineCreateInfo *pCreateInfos, const VkAllocationCallbacks *pAllocator, VkPipeline *pPipelines) {
	return f(device, pipelineCache, createInfoCount, pCreateInfos, pAllocator, pPipelines);
}

static VkResult vkCreateFence(PFN_vkCreateFence f, VkDevice device, const VkFenceCreateInfo *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkFence *pFence) {
	return f(device, pCreateInfo, pAllocator, pFence);
}

static void vkDestroyFence(PFN_vkDestroyFence f, VkDevice device, VkFence fence, const VkAllocationCallbacks *pAllocator) {
	f(device, fence, pAllocator);
}

static VkResult vkWaitForFences(PFN_vkWaitForFences f, VkDevice device, uint32_t fenceCount, const VkFence *pFences, VkBool32 waitAll, uint64_t timeout) {
	return f(device, fenceCount, pFences, waitAll, timeout);
}

static VkResult vkResetFences(PFN_vkResetFences f, VkDevice device, uint32_t fenceCount, const VkFence *pFences) {
	return f(device, fenceCount, pFences);
}

static void vkGetPhysicalDeviceProperties(PFN_vkGetPhysicalDeviceProperties f, VkPhysicalDevice physicalDevice, VkPhysicalDeviceProperties *pProperties) {
	f(physicalDevice, pProperties);
}

static VkResult vkGetPhysicalDeviceSurfaceSupportKHR(PFN_vkGetPhysicalDeviceSurfaceSupportKHR f, VkPhysicalDevice physicalDevice, uint32_t queueFamilyIndex, VkSurfaceKHR surface, VkBool32 *pSupported) {
	return f(physicalDevice, queueFamilyIndex, surface, pSupported);
}

static void vkDestroySurfaceKHR(PFN_vkDestroySurfaceKHR f, VkInstance instance, VkSurfaceKHR surface, const VkAllocationCallbacks *pAllocator) {
	f(instance, surface, pAllocator);
}

static VkResult vkGetPhysicalDeviceSurfaceFormatsKHR(PFN_vkGetPhysicalDeviceSurfaceFormatsKHR f, VkPhysicalDevice physicalDevice, VkSurfaceKHR surface, uint32_t *pSurfaceFormatCount, VkSurfaceFormatKHR *pSurfaceFormats) {
	return f(physicalDevice, surface, pSurfaceFormatCount, pSurfaceFormats);
}

static VkResult vkGetPhysicalDeviceSurfacePresentModesKHR(PFN_vkGetPhysicalDeviceSurfacePresentModesKHR f, VkPhysicalDevice physicalDevice, VkSurfaceKHR surface, uint32_t *pPresentModeCount, VkPresentModeKHR *pPresentModes) {
	return f(physicalDevice, surface, pPresentModeCount, pPresentModes);
}

static VkResult vkGetPhysicalDeviceSurfaceCapabilitiesKHR(PFN_vkGetPhysicalDeviceSurfaceCapabilitiesKHR f, VkPhysicalDevice physicalDevice, VkSurfaceKHR surface, VkSurfaceCapabilitiesKHR *pSurfaceCapabilities) {
	return f(physicalDevice, surface, pSurfaceCapabilities);
}

static VkResult vkCreateSwapchainKHR(PFN_vkCreateSwapchainKHR f, VkDevice device, const VkSwapchainCreateInfoKHR *pCreateInfo, const VkAllocationCallbacks *pAllocator, VkSwapchainKHR *pSwapchain) {
	return f(device, pCreateInfo, pAllocator, pSwapchain);
}

static void vkDestroySwapchainKHR(PFN_vkDestroySwapchainKHR f, VkDevice device, VkSwapchainKHR swapchain, const VkAllocationCallbacks *pAllocator) {
	f(device, swapchain, pAllocator);
}

static VkResult vkGetSwapchainImagesKHR(PFN_vkGetSwapchainImagesKHR f, VkDevice device, VkSwapchainKHR swapchain, uint32_t *pSwapchainImageCount, VkImage *pSwapchainImages) {
	return f(device, swapchain, pSwapchainImageCount, pSwapchainImages);
}

// indexAndResult holds both an integer and a result returned by value, to
// avoid Go heap allocation of the integer with Vulkan's return style.
struct intAndResult {
	uint32_t uint;
	VkResult res;
};

static struct intAndResult vkAcquireNextImageKHR(PFN_vkAcquireNextImageKHR f, VkDevice device, VkSwapchainKHR swapchain, uint64_t timeout, VkSemaphore semaphore, VkFence fence) {
	struct intAndResult res;
	res.res = f(device, swapchain, timeout, semaphore, fence, &res.uint);
	return res;
}

static VkResult vkQueuePresentKHR(PFN_vkQueuePresentKHR f, VkQueue queue, const VkPresentInfoKHR pPresentInfo) {
	return f(queue, &pPresentInfo);
}
*/
import "C"
import (
	"errors"
	"fmt"
	"image"
	"math"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

type (
	AttachmentLoadOp      = C.VkAttachmentLoadOp
	AccessFlags           = C.VkAccessFlags
	BlendFactor           = C.VkBlendFactor
	Buffer                = C.VkBuffer
	BufferImageCopy       = C.VkBufferImageCopy
	BufferMemoryBarrier   = C.VkBufferMemoryBarrier
	BufferUsageFlags      = C.VkBufferUsageFlags
	CommandPool           = C.VkCommandPool
	CommandBuffer         = C.VkCommandBuffer
	DependencyFlags       = C.VkDependencyFlags
	DescriptorPool        = C.VkDescriptorPool
	DescriptorPoolSize    = C.VkDescriptorPoolSize
	DescriptorSet         = C.VkDescriptorSet
	DescriptorSetLayout   = C.VkDescriptorSetLayout
	DescriptorType        = C.VkDescriptorType
	Device                = C.VkDevice
	DeviceMemory          = C.VkDeviceMemory
	DeviceSize            = C.VkDeviceSize
	Fence                 = C.VkFence
	Queue                 = C.VkQueue
	IndexType             = C.VkIndexType
	Image                 = C.VkImage
	ImageBlit             = C.VkImageBlit
	ImageCopy             = C.VkImageCopy
	ImageLayout           = C.VkImageLayout
	ImageMemoryBarrier    = C.VkImageMemoryBarrier
	ImageUsageFlags       = C.VkImageUsageFlags
	ImageView             = C.VkImageView
	Instance              = C.VkInstance
	Filter                = C.VkFilter
	Format                = C.VkFormat
	FormatFeatureFlags    = C.VkFormatFeatureFlags
	Framebuffer           = C.VkFramebuffer
	MemoryBarrier         = C.VkMemoryBarrier
	MemoryPropertyFlags   = C.VkMemoryPropertyFlags
	Pipeline              = C.VkPipeline
	PipelineBindPoint     = C.VkPipelineBindPoint
	PipelineLayout        = C.VkPipelineLayout
	PipelineStageFlags    = C.VkPipelineStageFlags
	PhysicalDevice        = C.VkPhysicalDevice
	PrimitiveTopology     = C.VkPrimitiveTopology
	PushConstantRange     = C.VkPushConstantRange
	QueueFamilyProperties = C.VkQueueFamilyProperties
	QueueFlags            = C.VkQueueFlags
	RenderPass            = C.VkRenderPass
	Sampler               = C.VkSampler
	SamplerMipmapMode     = C.VkSamplerMipmapMode
	Semaphore             = C.VkSemaphore
	ShaderModule          = C.VkShaderModule
	ShaderStageFlags      = C.VkShaderStageFlags
	SubpassDependency     = C.VkSubpassDependency
	Viewport              = C.VkViewport
	WriteDescriptorSet    = C.VkWriteDescriptorSet

	Surface             = C.VkSurfaceKHR
	SurfaceCapabilities = C.VkSurfaceCapabilitiesKHR

	Swapchain = C.VkSwapchainKHR
)

type VertexInputBindingDescription struct {
	Binding int
	Stride  int
}

type VertexInputAttributeDescription struct {
	Location int
	Binding  int
	Format   Format
	Offset   int
}

type DescriptorSetLayoutBinding struct {
	Binding        int
	DescriptorType DescriptorType
	StageFlags     ShaderStageFlags
}

type Error C.VkResult

const (
	FORMAT_R8G8B8A8_UNORM      Format = C.VK_FORMAT_R8G8B8A8_UNORM
	FORMAT_B8G8R8A8_SRGB       Format = C.VK_FORMAT_B8G8R8A8_SRGB
	FORMAT_R8G8B8A8_SRGB       Format = C.VK_FORMAT_R8G8B8A8_SRGB
	FORMAT_R16_SFLOAT          Format = C.VK_FORMAT_R16_SFLOAT
	FORMAT_R32_SFLOAT          Format = C.VK_FORMAT_R32_SFLOAT
	FORMAT_R32G32_SFLOAT       Format = C.VK_FORMAT_R32G32_SFLOAT
	FORMAT_R32G32B32_SFLOAT    Format = C.VK_FORMAT_R32G32B32_SFLOAT
	FORMAT_R32G32B32A32_SFLOAT Format = C.VK_FORMAT_R32G32B32A32_SFLOAT

	FORMAT_FEATURE_COLOR_ATTACHMENT_BIT            FormatFeatureFlags = C.VK_FORMAT_FEATURE_COLOR_ATTACHMENT_BIT
	FORMAT_FEATURE_COLOR_ATTACHMENT_BLEND_BIT      FormatFeatureFlags = C.VK_FORMAT_FEATURE_COLOR_ATTACHMENT_BLEND_BIT
	FORMAT_FEATURE_SAMPLED_IMAGE_BIT               FormatFeatureFlags = C.VK_FORMAT_FEATURE_SAMPLED_IMAGE_BIT
	FORMAT_FEATURE_SAMPLED_IMAGE_FILTER_LINEAR_BIT FormatFeatureFlags = C.VK_FORMAT_FEATURE_SAMPLED_IMAGE_FILTER_LINEAR_BIT

	IMAGE_USAGE_SAMPLED_BIT          ImageUsageFlags = C.VK_IMAGE_USAGE_SAMPLED_BIT
	IMAGE_USAGE_COLOR_ATTACHMENT_BIT ImageUsageFlags = C.VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT
	IMAGE_USAGE_STORAGE_BIT          ImageUsageFlags = C.VK_IMAGE_USAGE_STORAGE_BIT
	IMAGE_USAGE_TRANSFER_DST_BIT     ImageUsageFlags = C.VK_IMAGE_USAGE_TRANSFER_DST_BIT
	IMAGE_USAGE_TRANSFER_SRC_BIT     ImageUsageFlags = C.VK_IMAGE_USAGE_TRANSFER_SRC_BIT

	FILTER_NEAREST Filter = C.VK_FILTER_NEAREST
	FILTER_LINEAR  Filter = C.VK_FILTER_LINEAR

	ATTACHMENT_LOAD_OP_CLEAR     AttachmentLoadOp = C.VK_ATTACHMENT_LOAD_OP_CLEAR
	ATTACHMENT_LOAD_OP_DONT_CARE AttachmentLoadOp = C.VK_ATTACHMENT_LOAD_OP_DONT_CARE
	ATTACHMENT_LOAD_OP_LOAD      AttachmentLoadOp = C.VK_ATTACHMENT_LOAD_OP_LOAD

	IMAGE_LAYOUT_UNDEFINED                ImageLayout = C.VK_IMAGE_LAYOUT_UNDEFINED
	IMAGE_LAYOUT_PRESENT_SRC_KHR          ImageLayout = C.VK_IMAGE_LAYOUT_PRESENT_SRC_KHR
	IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL ImageLayout = C.VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL
	IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL ImageLayout = C.VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL
	IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL     ImageLayout = C.VK_IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL
	IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL     ImageLayout = C.VK_IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL
	IMAGE_LAYOUT_GENERAL                  ImageLayout = C.VK_IMAGE_LAYOUT_GENERAL

	BUFFER_USAGE_TRANSFER_DST_BIT   BufferUsageFlags = C.VK_BUFFER_USAGE_TRANSFER_DST_BIT
	BUFFER_USAGE_TRANSFER_SRC_BIT   BufferUsageFlags = C.VK_BUFFER_USAGE_TRANSFER_SRC_BIT
	BUFFER_USAGE_UNIFORM_BUFFER_BIT BufferUsageFlags = C.VK_BUFFER_USAGE_UNIFORM_BUFFER_BIT
	BUFFER_USAGE_STORAGE_BUFFER_BIT BufferUsageFlags = C.VK_BUFFER_USAGE_STORAGE_BUFFER_BIT
	BUFFER_USAGE_INDEX_BUFFER_BIT   BufferUsageFlags = C.VK_BUFFER_USAGE_INDEX_BUFFER_BIT
	BUFFER_USAGE_VERTEX_BUFFER_BIT  BufferUsageFlags = C.VK_BUFFER_USAGE_VERTEX_BUFFER_BIT

	ERROR_OUT_OF_DATE_KHR  = Error(C.VK_ERROR_OUT_OF_DATE_KHR)
	ERROR_SURFACE_LOST_KHR = Error(C.VK_ERROR_SURFACE_LOST_KHR)
	ERROR_DEVICE_LOST      = Error(C.VK_ERROR_DEVICE_LOST)
	SUBOPTIMAL_KHR         = Error(C.VK_SUBOPTIMAL_KHR)

	FENCE_CREATE_SIGNALED_BIT = 0x00000001

	BLEND_FACTOR_ZERO                BlendFactor = C.VK_BLEND_FACTOR_ZERO
	BLEND_FACTOR_ONE                 BlendFactor = C.VK_BLEND_FACTOR_ONE
	BLEND_FACTOR_ONE_MINUS_SRC_ALPHA BlendFactor = C.VK_BLEND_FACTOR_ONE_MINUS_SRC_ALPHA
	BLEND_FACTOR_DST_COLOR           BlendFactor = C.VK_BLEND_FACTOR_DST_COLOR

	PRIMITIVE_TOPOLOGY_TRIANGLE_LIST  PrimitiveTopology = C.VK_PRIMITIVE_TOPOLOGY_TRIANGLE_LIST
	PRIMITIVE_TOPOLOGY_TRIANGLE_STRIP PrimitiveTopology = C.VK_PRIMITIVE_TOPOLOGY_TRIANGLE_STRIP

	SHADER_STAGE_VERTEX_BIT   ShaderStageFlags = C.VK_SHADER_STAGE_VERTEX_BIT
	SHADER_STAGE_FRAGMENT_BIT ShaderStageFlags = C.VK_SHADER_STAGE_FRAGMENT_BIT
	SHADER_STAGE_COMPUTE_BIT  ShaderStageFlags = C.VK_SHADER_STAGE_COMPUTE_BIT

	DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER DescriptorType = C.VK_DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER
	DESCRIPTOR_TYPE_UNIFORM_BUFFER         DescriptorType = C.VK_DESCRIPTOR_TYPE_UNIFORM_BUFFER
	DESCRIPTOR_TYPE_STORAGE_BUFFER         DescriptorType = C.VK_DESCRIPTOR_TYPE_STORAGE_BUFFER
	DESCRIPTOR_TYPE_STORAGE_IMAGE          DescriptorType = C.VK_DESCRIPTOR_TYPE_STORAGE_IMAGE

	MEMORY_PROPERTY_DEVICE_LOCAL_BIT  MemoryPropertyFlags = C.VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT
	MEMORY_PROPERTY_HOST_VISIBLE_BIT  MemoryPropertyFlags = C.VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT
	MEMORY_PROPERTY_HOST_COHERENT_BIT MemoryPropertyFlags = C.VK_MEMORY_PROPERTY_HOST_COHERENT_BIT

	DEPENDENCY_BY_REGION_BIT DependencyFlags = C.VK_DEPENDENCY_BY_REGION_BIT

	PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT PipelineStageFlags = C.VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT
	PIPELINE_STAGE_TRANSFER_BIT                PipelineStageFlags = C.VK_PIPELINE_STAGE_TRANSFER_BIT
	PIPELINE_STAGE_FRAGMENT_SHADER_BIT         PipelineStageFlags = C.VK_PIPELINE_STAGE_FRAGMENT_SHADER_BIT
	PIPELINE_STAGE_COMPUTE_SHADER_BIT          PipelineStageFlags = C.VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT
	PIPELINE_STAGE_TOP_OF_PIPE_BIT             PipelineStageFlags = C.VK_PIPELINE_STAGE_TOP_OF_PIPE_BIT
	PIPELINE_STAGE_HOST_BIT                    PipelineStageFlags = C.VK_PIPELINE_STAGE_HOST_BIT
	PIPELINE_STAGE_VERTEX_INPUT_BIT            PipelineStageFlags = C.VK_PIPELINE_STAGE_VERTEX_INPUT_BIT
	PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT          PipelineStageFlags = C.VK_PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT

	ACCESS_MEMORY_READ_BIT            AccessFlags = C.VK_ACCESS_MEMORY_READ_BIT
	ACCESS_MEMORY_WRITE_BIT           AccessFlags = C.VK_ACCESS_MEMORY_WRITE_BIT
	ACCESS_TRANSFER_READ_BIT          AccessFlags = C.VK_ACCESS_TRANSFER_READ_BIT
	ACCESS_TRANSFER_WRITE_BIT         AccessFlags = C.VK_ACCESS_TRANSFER_WRITE_BIT
	ACCESS_SHADER_READ_BIT            AccessFlags = C.VK_ACCESS_SHADER_READ_BIT
	ACCESS_SHADER_WRITE_BIT           AccessFlags = C.VK_ACCESS_SHADER_WRITE_BIT
	ACCESS_COLOR_ATTACHMENT_READ_BIT  AccessFlags = C.VK_ACCESS_COLOR_ATTACHMENT_READ_BIT
	ACCESS_COLOR_ATTACHMENT_WRITE_BIT AccessFlags = C.VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT
	ACCESS_HOST_READ_BIT              AccessFlags = C.VK_ACCESS_HOST_READ_BIT
	ACCESS_HOST_WRITE_BIT             AccessFlags = C.VK_ACCESS_HOST_WRITE_BIT
	ACCESS_VERTEX_ATTRIBUTE_READ_BIT  AccessFlags = C.VK_ACCESS_VERTEX_ATTRIBUTE_READ_BIT
	ACCESS_INDEX_READ_BIT             AccessFlags = C.VK_ACCESS_INDEX_READ_BIT

	PIPELINE_BIND_POINT_COMPUTE  PipelineBindPoint = C.VK_PIPELINE_BIND_POINT_COMPUTE
	PIPELINE_BIND_POINT_GRAPHICS PipelineBindPoint = C.VK_PIPELINE_BIND_POINT_GRAPHICS

	INDEX_TYPE_UINT16 IndexType = C.VK_INDEX_TYPE_UINT16
	INDEX_TYPE_UINT32 IndexType = C.VK_INDEX_TYPE_UINT32

	QUEUE_GRAPHICS_BIT QueueFlags = C.VK_QUEUE_GRAPHICS_BIT
	QUEUE_COMPUTE_BIT  QueueFlags = C.VK_QUEUE_COMPUTE_BIT

	SAMPLER_MIPMAP_MODE_NEAREST SamplerMipmapMode = C.VK_SAMPLER_MIPMAP_MODE_NEAREST
	SAMPLER_MIPMAP_MODE_LINEAR  SamplerMipmapMode = C.VK_SAMPLER_MIPMAP_MODE_LINEAR

	REMAINING_MIP_LEVELS = -1
)

var (
	once    sync.Once
	loadErr error

	loadFuncs []func(dlopen func(name string) *[0]byte)
)

var funcs struct {
	vkCreateInstance                         C.PFN_vkCreateInstance
	vkDestroyInstance                        C.PFN_vkDestroyInstance
	vkEnumeratePhysicalDevices               C.PFN_vkEnumeratePhysicalDevices
	vkGetPhysicalDeviceQueueFamilyProperties C.PFN_vkGetPhysicalDeviceQueueFamilyProperties
	vkGetPhysicalDeviceFormatProperties      C.PFN_vkGetPhysicalDeviceFormatProperties
	vkCreateDevice                           C.PFN_vkCreateDevice
	vkDestroyDevice                          C.PFN_vkDestroyDevice
	vkGetDeviceQueue                         C.PFN_vkGetDeviceQueue
	vkCreateImageView                        C.PFN_vkCreateImageView
	vkDestroyImageView                       C.PFN_vkDestroyImageView
	vkCreateFramebuffer                      C.PFN_vkCreateFramebuffer
	vkDestroyFramebuffer                     C.PFN_vkDestroyFramebuffer
	vkDeviceWaitIdle                         C.PFN_vkDeviceWaitIdle
	vkQueueWaitIdle                          C.PFN_vkQueueWaitIdle
	vkCreateSemaphore                        C.PFN_vkCreateSemaphore
	vkDestroySemaphore                       C.PFN_vkDestroySemaphore
	vkCreateRenderPass                       C.PFN_vkCreateRenderPass
	vkDestroyRenderPass                      C.PFN_vkDestroyRenderPass
	vkCreateCommandPool                      C.PFN_vkCreateCommandPool
	vkDestroyCommandPool                     C.PFN_vkDestroyCommandPool
	vkAllocateCommandBuffers                 C.PFN_vkAllocateCommandBuffers
	vkFreeCommandBuffers                     C.PFN_vkFreeCommandBuffers
	vkBeginCommandBuffer                     C.PFN_vkBeginCommandBuffer
	vkEndCommandBuffer                       C.PFN_vkEndCommandBuffer
	vkQueueSubmit                            C.PFN_vkQueueSubmit
	vkCmdBeginRenderPass                     C.PFN_vkCmdBeginRenderPass
	vkCmdEndRenderPass                       C.PFN_vkCmdEndRenderPass
	vkCmdCopyBuffer                          C.PFN_vkCmdCopyBuffer
	vkCmdCopyBufferToImage                   C.PFN_vkCmdCopyBufferToImage
	vkCmdPipelineBarrier                     C.PFN_vkCmdPipelineBarrier
	vkCmdPushConstants                       C.PFN_vkCmdPushConstants
	vkCmdBindPipeline                        C.PFN_vkCmdBindPipeline
	vkCmdBindVertexBuffers                   C.PFN_vkCmdBindVertexBuffers
	vkCmdSetViewport                         C.PFN_vkCmdSetViewport
	vkCmdBindIndexBuffer                     C.PFN_vkCmdBindIndexBuffer
	vkCmdDraw                                C.PFN_vkCmdDraw
	vkCmdDrawIndexed                         C.PFN_vkCmdDrawIndexed
	vkCmdBindDescriptorSets                  C.PFN_vkCmdBindDescriptorSets
	vkCmdCopyImageToBuffer                   C.PFN_vkCmdCopyImageToBuffer
	vkCmdDispatch                            C.PFN_vkCmdDispatch
	vkCreateImage                            C.PFN_vkCreateImage
	vkDestroyImage                           C.PFN_vkDestroyImage
	vkGetImageMemoryRequirements             C.PFN_vkGetImageMemoryRequirements
	vkAllocateMemory                         C.PFN_vkAllocateMemory
	vkBindImageMemory                        C.PFN_vkBindImageMemory
	vkFreeMemory                             C.PFN_vkFreeMemory
	vkGetPhysicalDeviceMemoryProperties      C.PFN_vkGetPhysicalDeviceMemoryProperties
	vkCreateSampler                          C.PFN_vkCreateSampler
	vkDestroySampler                         C.PFN_vkDestroySampler
	vkCreateBuffer                           C.PFN_vkCreateBuffer
	vkDestroyBuffer                          C.PFN_vkDestroyBuffer
	vkGetBufferMemoryRequirements            C.PFN_vkGetBufferMemoryRequirements
	vkBindBufferMemory                       C.PFN_vkBindBufferMemory
	vkCreateShaderModule                     C.PFN_vkCreateShaderModule
	vkDestroyShaderModule                    C.PFN_vkDestroyShaderModule
	vkCreateGraphicsPipelines                C.PFN_vkCreateGraphicsPipelines
	vkDestroyPipeline                        C.PFN_vkDestroyPipeline
	vkCreatePipelineLayout                   C.PFN_vkCreatePipelineLayout
	vkDestroyPipelineLayout                  C.PFN_vkDestroyPipelineLayout
	vkCreateDescriptorSetLayout              C.PFN_vkCreateDescriptorSetLayout
	vkDestroyDescriptorSetLayout             C.PFN_vkDestroyDescriptorSetLayout
	vkMapMemory                              C.PFN_vkMapMemory
	vkUnmapMemory                            C.PFN_vkUnmapMemory
	vkResetCommandBuffer                     C.PFN_vkResetCommandBuffer
	vkCreateDescriptorPool                   C.PFN_vkCreateDescriptorPool
	vkDestroyDescriptorPool                  C.PFN_vkDestroyDescriptorPool
	vkAllocateDescriptorSets                 C.PFN_vkAllocateDescriptorSets
	vkFreeDescriptorSets                     C.PFN_vkFreeDescriptorSets
	vkUpdateDescriptorSets                   C.PFN_vkUpdateDescriptorSets
	vkResetDescriptorPool                    C.PFN_vkResetDescriptorPool
	vkCmdBlitImage                           C.PFN_vkCmdBlitImage
	vkCmdCopyImage                           C.PFN_vkCmdCopyImage
	vkCreateComputePipelines                 C.PFN_vkCreateComputePipelines
	vkCreateFence                            C.PFN_vkCreateFence
	vkDestroyFence                           C.PFN_vkDestroyFence
	vkWaitForFences                          C.PFN_vkWaitForFences
	vkResetFences                            C.PFN_vkResetFences
	vkGetPhysicalDeviceProperties            C.PFN_vkGetPhysicalDeviceProperties

	vkGetPhysicalDeviceSurfaceSupportKHR      C.PFN_vkGetPhysicalDeviceSurfaceSupportKHR
	vkDestroySurfaceKHR                       C.PFN_vkDestroySurfaceKHR
	vkGetPhysicalDeviceSurfaceFormatsKHR      C.PFN_vkGetPhysicalDeviceSurfaceFormatsKHR
	vkGetPhysicalDeviceSurfacePresentModesKHR C.PFN_vkGetPhysicalDeviceSurfacePresentModesKHR
	vkGetPhysicalDeviceSurfaceCapabilitiesKHR C.PFN_vkGetPhysicalDeviceSurfaceCapabilitiesKHR

	vkCreateSwapchainKHR    C.PFN_vkCreateSwapchainKHR
	vkDestroySwapchainKHR   C.PFN_vkDestroySwapchainKHR
	vkGetSwapchainImagesKHR C.PFN_vkGetSwapchainImagesKHR
	vkAcquireNextImageKHR   C.PFN_vkAcquireNextImageKHR
	vkQueuePresentKHR       C.PFN_vkQueuePresentKHR
}

var (
	nilSurface             C.VkSurfaceKHR
	nilSwapchain           C.VkSwapchainKHR
	nilSemaphore           C.VkSemaphore
	nilImageView           C.VkImageView
	nilRenderPass          C.VkRenderPass
	nilFramebuffer         C.VkFramebuffer
	nilCommandPool         C.VkCommandPool
	nilImage               C.VkImage
	nilDeviceMemory        C.VkDeviceMemory
	nilSampler             C.VkSampler
	nilBuffer              C.VkBuffer
	nilShaderModule        C.VkShaderModule
	nilPipeline            C.VkPipeline
	nilPipelineCache       C.VkPipelineCache
	nilPipelineLayout      C.VkPipelineLayout
	nilDescriptorSetLayout C.VkDescriptorSetLayout
	nilDescriptorPool      C.VkDescriptorPool
	nilDescriptorSet       C.VkDescriptorSet
	nilFence               C.VkFence
)

func vkInit() error {
	once.Do(func() {
		var libName string
		switch {
		case runtime.GOOS == "android":
			libName = "libvulkan.so"
		default:
			libName = "libvulkan.so.1"
		}
		lib := dlopen(libName)
		if lib == nil {
			loadErr = fmt.Errorf("vulkan: %s", C.GoString(C.dlerror()))
			return
		}
		dlopen := func(name string) *[0]byte {
			return (*[0]byte)(dlsym(lib, name))
		}
		must := func(name string) *[0]byte {
			ptr := dlopen(name)
			if ptr != nil {
				return ptr
			}
			if loadErr == nil {
				loadErr = fmt.Errorf("vulkan: function %q not found: %s", name, C.GoString(C.dlerror()))
			}
			return nil
		}
		funcs.vkCreateInstance = must("vkCreateInstance")
		funcs.vkDestroyInstance = must("vkDestroyInstance")
		funcs.vkEnumeratePhysicalDevices = must("vkEnumeratePhysicalDevices")
		funcs.vkGetPhysicalDeviceQueueFamilyProperties = must("vkGetPhysicalDeviceQueueFamilyProperties")
		funcs.vkGetPhysicalDeviceFormatProperties = must("vkGetPhysicalDeviceFormatProperties")
		funcs.vkCreateDevice = must("vkCreateDevice")
		funcs.vkDestroyDevice = must("vkDestroyDevice")
		funcs.vkGetDeviceQueue = must("vkGetDeviceQueue")
		funcs.vkCreateImageView = must("vkCreateImageView")
		funcs.vkDestroyImageView = must("vkDestroyImageView")
		funcs.vkCreateFramebuffer = must("vkCreateFramebuffer")
		funcs.vkDestroyFramebuffer = must("vkDestroyFramebuffer")
		funcs.vkDeviceWaitIdle = must("vkDeviceWaitIdle")
		funcs.vkQueueWaitIdle = must("vkQueueWaitIdle")
		funcs.vkCreateSemaphore = must("vkCreateSemaphore")
		funcs.vkDestroySemaphore = must("vkDestroySemaphore")
		funcs.vkCreateRenderPass = must("vkCreateRenderPass")
		funcs.vkDestroyRenderPass = must("vkDestroyRenderPass")
		funcs.vkCreateCommandPool = must("vkCreateCommandPool")
		funcs.vkDestroyCommandPool = must("vkDestroyCommandPool")
		funcs.vkAllocateCommandBuffers = must("vkAllocateCommandBuffers")
		funcs.vkFreeCommandBuffers = must("vkFreeCommandBuffers")
		funcs.vkBeginCommandBuffer = must("vkBeginCommandBuffer")
		funcs.vkEndCommandBuffer = must("vkEndCommandBuffer")
		funcs.vkQueueSubmit = must("vkQueueSubmit")
		funcs.vkCmdBeginRenderPass = must("vkCmdBeginRenderPass")
		funcs.vkCmdEndRenderPass = must("vkCmdEndRenderPass")
		funcs.vkCmdCopyBuffer = must("vkCmdCopyBuffer")
		funcs.vkCmdCopyBufferToImage = must("vkCmdCopyBufferToImage")
		funcs.vkCmdPipelineBarrier = must("vkCmdPipelineBarrier")
		funcs.vkCmdPushConstants = must("vkCmdPushConstants")
		funcs.vkCmdBindPipeline = must("vkCmdBindPipeline")
		funcs.vkCmdBindVertexBuffers = must("vkCmdBindVertexBuffers")
		funcs.vkCmdSetViewport = must("vkCmdSetViewport")
		funcs.vkCmdBindIndexBuffer = must("vkCmdBindIndexBuffer")
		funcs.vkCmdDraw = must("vkCmdDraw")
		funcs.vkCmdDrawIndexed = must("vkCmdDrawIndexed")
		funcs.vkCmdBindDescriptorSets = must("vkCmdBindDescriptorSets")
		funcs.vkCmdCopyImageToBuffer = must("vkCmdCopyImageToBuffer")
		funcs.vkCmdDispatch = must("vkCmdDispatch")
		funcs.vkCreateImage = must("vkCreateImage")
		funcs.vkDestroyImage = must("vkDestroyImage")
		funcs.vkGetImageMemoryRequirements = must("vkGetImageMemoryRequirements")
		funcs.vkAllocateMemory = must("vkAllocateMemory")
		funcs.vkBindImageMemory = must("vkBindImageMemory")
		funcs.vkFreeMemory = must("vkFreeMemory")
		funcs.vkGetPhysicalDeviceMemoryProperties = must("vkGetPhysicalDeviceMemoryProperties")
		funcs.vkCreateSampler = must("vkCreateSampler")
		funcs.vkDestroySampler = must("vkDestroySampler")
		funcs.vkCreateBuffer = must("vkCreateBuffer")
		funcs.vkDestroyBuffer = must("vkDestroyBuffer")
		funcs.vkGetBufferMemoryRequirements = must("vkGetBufferMemoryRequirements")
		funcs.vkBindBufferMemory = must("vkBindBufferMemory")
		funcs.vkCreateShaderModule = must("vkCreateShaderModule")
		funcs.vkDestroyShaderModule = must("vkDestroyShaderModule")
		funcs.vkCreateGraphicsPipelines = must("vkCreateGraphicsPipelines")
		funcs.vkDestroyPipeline = must("vkDestroyPipeline")
		funcs.vkCreatePipelineLayout = must("vkCreatePipelineLayout")
		funcs.vkDestroyPipelineLayout = must("vkDestroyPipelineLayout")
		funcs.vkCreateDescriptorSetLayout = must("vkCreateDescriptorSetLayout")
		funcs.vkDestroyDescriptorSetLayout = must("vkDestroyDescriptorSetLayout")
		funcs.vkMapMemory = must("vkMapMemory")
		funcs.vkUnmapMemory = must("vkUnmapMemory")
		funcs.vkResetCommandBuffer = must("vkResetCommandBuffer")
		funcs.vkCreateDescriptorPool = must("vkCreateDescriptorPool")
		funcs.vkDestroyDescriptorPool = must("vkDestroyDescriptorPool")
		funcs.vkAllocateDescriptorSets = must("vkAllocateDescriptorSets")
		funcs.vkFreeDescriptorSets = must("vkFreeDescriptorSets")
		funcs.vkUpdateDescriptorSets = must("vkUpdateDescriptorSets")
		funcs.vkResetDescriptorPool = must("vkResetDescriptorPool")
		funcs.vkCmdBlitImage = must("vkCmdBlitImage")
		funcs.vkCmdCopyImage = must("vkCmdCopyImage")
		funcs.vkCreateComputePipelines = must("vkCreateComputePipelines")
		funcs.vkCreateFence = must("vkCreateFence")
		funcs.vkDestroyFence = must("vkDestroyFence")
		funcs.vkWaitForFences = must("vkWaitForFences")
		funcs.vkResetFences = must("vkResetFences")
		funcs.vkGetPhysicalDeviceProperties = must("vkGetPhysicalDeviceProperties")

		funcs.vkGetPhysicalDeviceSurfaceSupportKHR = dlopen("vkGetPhysicalDeviceSurfaceSupportKHR")
		funcs.vkDestroySurfaceKHR = dlopen("vkDestroySurfaceKHR")
		funcs.vkGetPhysicalDeviceSurfaceFormatsKHR = dlopen("vkGetPhysicalDeviceSurfaceFormatsKHR")
		funcs.vkGetPhysicalDeviceSurfacePresentModesKHR = dlopen("vkGetPhysicalDeviceSurfacePresentModesKHR")
		funcs.vkGetPhysicalDeviceSurfaceCapabilitiesKHR = dlopen("vkGetPhysicalDeviceSurfaceCapabilitiesKHR")

		funcs.vkCreateSwapchainKHR = dlopen("vkCreateSwapchainKHR")
		funcs.vkDestroySwapchainKHR = dlopen("vkDestroySwapchainKHR")
		funcs.vkGetSwapchainImagesKHR = dlopen("vkGetSwapchainImagesKHR")
		funcs.vkAcquireNextImageKHR = dlopen("vkAcquireNextImageKHR")
		funcs.vkQueuePresentKHR = dlopen("vkQueuePresentKHR")

		for _, f := range loadFuncs {
			f(dlopen)
		}
	})
	return loadErr
}

func CreateInstance(exts ...string) (Instance, error) {
	if err := vkInit(); err != nil {
		return nil, err
	}
	// VK_MAKE_API_VERSION macro.
	makeVer := func(variant, major, minor, patch int) C.uint32_t {
		return ((((C.uint32_t)(variant)) << 29) | (((C.uint32_t)(major)) << 22) | (((C.uint32_t)(minor)) << 12) | ((C.uint32_t)(patch)))
	}
	appInf := C.VkApplicationInfo{
		sType:      C.VK_STRUCTURE_TYPE_APPLICATION_INFO,
		apiVersion: makeVer(0, 1, 0, 0),
	}
	inf := C.VkInstanceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
		// pApplicationInfo may be omitted according to the spec, but the Android
		// emulator crashes without it.
		pApplicationInfo: &appInf,
	}
	if len(exts) > 0 {
		cexts := mallocCStringArr(exts)
		defer freeCStringArr(cexts)
		inf.enabledExtensionCount = C.uint32_t(len(exts))
		inf.ppEnabledExtensionNames = &cexts[0]
	}
	var inst Instance
	if err := vkErr(C.vkCreateInstance(funcs.vkCreateInstance, inf, nil, &inst)); err != nil {
		return nil, fmt.Errorf("vulkan: vkCreateInstance: %w", err)
	}
	return inst, nil
}

func mallocCStringArr(s []string) []*C.char {
	carr := make([]*C.char, len(s))
	for i, ext := range s {
		carr[i] = C.CString(ext)
	}
	return carr
}

func freeCStringArr(s []*C.char) {
	for i := range s {
		C.free(unsafe.Pointer(s[i]))
		s[i] = nil
	}
}

func DestroyInstance(inst Instance) {
	C.vkDestroyInstance(funcs.vkDestroyInstance, inst, nil)
}

func GetPhysicalDeviceQueueFamilyProperties(pd PhysicalDevice) []QueueFamilyProperties {
	var count C.uint32_t
	C.vkGetPhysicalDeviceQueueFamilyProperties(funcs.vkGetPhysicalDeviceQueueFamilyProperties, pd, &count, nil)
	if count == 0 {
		return nil
	}
	queues := make([]C.VkQueueFamilyProperties, count)
	C.vkGetPhysicalDeviceQueueFamilyProperties(funcs.vkGetPhysicalDeviceQueueFamilyProperties, pd, &count, &queues[0])
	return queues
}

func EnumeratePhysicalDevices(inst Instance) ([]PhysicalDevice, error) {
	var count C.uint32_t
	if err := vkErr(C.vkEnumeratePhysicalDevices(funcs.vkEnumeratePhysicalDevices, inst, &count, nil)); err != nil {
		return nil, fmt.Errorf("vulkan: vkEnumeratePhysicalDevices: %w", err)
	}
	if count == 0 {
		return nil, nil
	}
	devs := make([]C.VkPhysicalDevice, count)
	if err := vkErr(C.vkEnumeratePhysicalDevices(funcs.vkEnumeratePhysicalDevices, inst, &count, &devs[0])); err != nil {
		return nil, fmt.Errorf("vulkan: vkEnumeratePhysicalDevices: %w", err)
	}
	return devs, nil
}

func ChoosePhysicalDevice(inst Instance, surf Surface) (PhysicalDevice, int, error) {
	devs, err := EnumeratePhysicalDevices(inst)
	if err != nil {
		return nil, 0, err
	}
	for _, pd := range devs {
		var props C.VkPhysicalDeviceProperties
		C.vkGetPhysicalDeviceProperties(funcs.vkGetPhysicalDeviceProperties, pd, &props)
		// The lavapipe software implementation doesn't work well rendering to a surface.
		// See https://gitlab.freedesktop.org/mesa/mesa/-/issues/5473.
		if surf != 0 && props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_CPU {
			continue
		}
		const caps = C.VK_QUEUE_GRAPHICS_BIT | C.VK_QUEUE_COMPUTE_BIT
		queueIdx, ok, err := chooseQueue(pd, surf, caps)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			continue
		}
		if surf != nilSurface {
			_, fmtFound, err := chooseFormat(pd, surf)
			if err != nil {
				return nil, 0, err
			}
			_, modFound, err := choosePresentMode(pd, surf)
			if err != nil {
				return nil, 0, err
			}
			if !fmtFound || !modFound {
				continue
			}
		}
		return pd, queueIdx, nil
	}
	return nil, 0, errors.New("vulkan: no suitable device found")
}

func CreateDeviceAndQueue(pd C.VkPhysicalDevice, queueIdx int, exts ...string) (Device, error) {
	priority := C.float(1.0)
	qinf := C.VkDeviceQueueCreateInfo{
		sType:            C.VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
		queueCount:       1,
		queueFamilyIndex: C.uint32_t(queueIdx),
		pQueuePriorities: &priority,
	}
	inf := C.VkDeviceCreateInfo{
		sType:                 C.VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO,
		queueCreateInfoCount:  1,
		enabledExtensionCount: C.uint32_t(len(exts)),
	}
	if len(exts) > 0 {
		cexts := mallocCStringArr(exts)
		defer freeCStringArr(cexts)
		inf.ppEnabledExtensionNames = &cexts[0]
	}
	var dev Device
	if err := vkErr(C.vkCreateDevice(funcs.vkCreateDevice, pd, inf, qinf, nil, &dev)); err != nil {
		return nil, fmt.Errorf("vulkan: vkCreateDevice: %w", err)
	}
	return dev, nil
}

func GetDeviceQueue(d Device, queueFamily, queueIndex int) Queue {
	var queue Queue
	C.vkGetDeviceQueue(funcs.vkGetDeviceQueue, d, C.uint32_t(queueFamily), C.uint32_t(queueIndex), &queue)
	return queue
}

func GetPhysicalDeviceSurfaceCapabilities(pd PhysicalDevice, surf Surface) (SurfaceCapabilities, error) {
	var caps C.VkSurfaceCapabilitiesKHR
	err := vkErr(C.vkGetPhysicalDeviceSurfaceCapabilitiesKHR(funcs.vkGetPhysicalDeviceSurfaceCapabilitiesKHR, pd, surf, &caps))
	if err != nil {
		return SurfaceCapabilities{}, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceCapabilitiesKHR: %w", err)
	}
	return caps, nil
}

func CreateSwapchain(pd PhysicalDevice, d Device, surf Surface, width, height int, old Swapchain) (Swapchain, []Image, Format, error) {
	caps, err := GetPhysicalDeviceSurfaceCapabilities(pd, surf)
	if err != nil {
		return nilSwapchain, nil, 0, err
	}
	mode, modeOK, err := choosePresentMode(pd, surf)
	if err != nil {
		return nilSwapchain, nil, 0, err
	}
	format, fmtOK, err := chooseFormat(pd, surf)
	if err != nil {
		return nilSwapchain, nil, 0, err
	}
	if !modeOK || !fmtOK {
		// This shouldn't happen because CreateDeviceAndQueue found at least
		// one valid format and present mode.
		return nilSwapchain, nil, 0, errors.New("vulkan: no valid format and present mode found")
	}
	// Find supported alpha composite mode. It doesn't matter which one, because rendering is
	// always opaque.
	alphaComp := C.VkCompositeAlphaFlagBitsKHR(1)
	for caps.supportedCompositeAlpha&C.VkCompositeAlphaFlagsKHR(alphaComp) == 0 {
		alphaComp <<= 1
	}
	trans := C.VkSurfaceTransformFlagBitsKHR(C.VK_SURFACE_TRANSFORM_IDENTITY_BIT_KHR)
	if caps.supportedTransforms&C.VkSurfaceTransformFlagsKHR(trans) == 0 {
		return nilSwapchain, nil, 0, errors.New("vulkan: VK_SURFACE_TRANSFORM_IDENTITY_BIT_KHR not supported")
	}
	inf := C.VkSwapchainCreateInfoKHR{
		sType:            C.VK_STRUCTURE_TYPE_SWAPCHAIN_CREATE_INFO_KHR,
		surface:          surf,
		minImageCount:    caps.minImageCount,
		imageFormat:      format.format,
		imageColorSpace:  format.colorSpace,
		imageExtent:      C.VkExtent2D{width: C.uint32_t(width), height: C.uint32_t(height)},
		imageArrayLayers: 1,
		imageUsage:       C.VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT,
		imageSharingMode: C.VK_SHARING_MODE_EXCLUSIVE,
		preTransform:     trans,
		presentMode:      mode,
		compositeAlpha:   C.VkCompositeAlphaFlagBitsKHR(alphaComp),
		clipped:          C.VK_TRUE,
		oldSwapchain:     old,
	}
	var swchain Swapchain
	if err := vkErr(C.vkCreateSwapchainKHR(funcs.vkCreateSwapchainKHR, d, &inf, nil, &swchain)); err != nil {
		return nilSwapchain, nil, 0, fmt.Errorf("vulkan: vkCreateSwapchainKHR: %w", err)
	}
	var count C.uint32_t
	if err := vkErr(C.vkGetSwapchainImagesKHR(funcs.vkGetSwapchainImagesKHR, d, swchain, &count, nil)); err != nil {
		DestroySwapchain(d, swchain)
		return nilSwapchain, nil, 0, fmt.Errorf("vulkan: vkGetSwapchainImagesKHR: %w", err)
	}
	if count == 0 {
		DestroySwapchain(d, swchain)
		return nilSwapchain, nil, 0, errors.New("vulkan: vkGetSwapchainImagesKHR returned no images")
	}
	imgs := make([]Image, count)
	if err := vkErr(C.vkGetSwapchainImagesKHR(funcs.vkGetSwapchainImagesKHR, d, swchain, &count, &imgs[0])); err != nil {
		DestroySwapchain(d, swchain)
		return nilSwapchain, nil, 0, fmt.Errorf("vulkan: vkGetSwapchainImagesKHR: %w", err)
	}
	return swchain, imgs, format.format, nil
}

func DestroySwapchain(d Device, swchain Swapchain) {
	C.vkDestroySwapchainKHR(funcs.vkDestroySwapchainKHR, d, swchain, nil)
}

func AcquireNextImage(d Device, swchain Swapchain, sem Semaphore, fence Fence) (int, error) {
	res := C.vkAcquireNextImageKHR(funcs.vkAcquireNextImageKHR, d, swchain, math.MaxUint64, sem, fence)
	if err := vkErr(res.res); err != nil {
		return 0, fmt.Errorf("vulkan: vkAcquireNextImageKHR: %w", err)
	}
	return int(res.uint), nil
}

func PresentQueue(q Queue, swchain Swapchain, sem Semaphore, imgIdx int) error {
	cidx := C.uint32_t(imgIdx)
	inf := C.VkPresentInfoKHR{
		sType:          C.VK_STRUCTURE_TYPE_PRESENT_INFO_KHR,
		swapchainCount: 1,
		pSwapchains:    &swchain,
		pImageIndices:  &cidx,
	}
	if sem != nilSemaphore {
		inf.waitSemaphoreCount = 1
		inf.pWaitSemaphores = &sem
	}
	if err := vkErr(C.vkQueuePresentKHR(funcs.vkQueuePresentKHR, q, inf)); err != nil {
		return fmt.Errorf("vulkan: vkQueuePresentKHR: %w", err)
	}
	return nil
}

func CreateImageView(d Device, img Image, format Format) (ImageView, error) {
	inf := C.VkImageViewCreateInfo{
		sType:    C.VK_STRUCTURE_TYPE_IMAGE_VIEW_CREATE_INFO,
		image:    img,
		viewType: C.VK_IMAGE_VIEW_TYPE_2D,
		format:   format,
		subresourceRange: C.VkImageSubresourceRange{
			aspectMask: C.VK_IMAGE_ASPECT_COLOR_BIT,
			levelCount: C.VK_REMAINING_MIP_LEVELS,
			layerCount: C.VK_REMAINING_ARRAY_LAYERS,
		},
	}
	var view C.VkImageView
	if err := vkErr(C.vkCreateImageView(funcs.vkCreateImageView, d, &inf, nil, &view)); err != nil {
		return nilImageView, fmt.Errorf("vulkan: vkCreateImageView: %w", err)
	}
	return view, nil
}

func DestroyImageView(d Device, view ImageView) {
	C.vkDestroyImageView(funcs.vkDestroyImageView, d, view, nil)
}

func CreateRenderPass(d Device, format Format, loadOp AttachmentLoadOp, initialLayout, finalLayout ImageLayout, passDeps []SubpassDependency) (RenderPass, error) {
	att := C.VkAttachmentDescription{
		format:         format,
		samples:        C.VK_SAMPLE_COUNT_1_BIT,
		loadOp:         loadOp,
		storeOp:        C.VK_ATTACHMENT_STORE_OP_STORE,
		stencilLoadOp:  C.VK_ATTACHMENT_LOAD_OP_DONT_CARE,
		stencilStoreOp: C.VK_ATTACHMENT_STORE_OP_DONT_CARE,
		initialLayout:  initialLayout,
		finalLayout:    finalLayout,
	}

	ref := C.VkAttachmentReference{
		attachment: 0,
		layout:     C.VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL,
	}

	sub := C.VkSubpassDescription{
		pipelineBindPoint:    C.VK_PIPELINE_BIND_POINT_GRAPHICS,
		colorAttachmentCount: 1,
		pColorAttachments:    &ref,
	}

	inf := C.VkRenderPassCreateInfo{
		sType:           C.VK_STRUCTURE_TYPE_RENDER_PASS_CREATE_INFO,
		attachmentCount: 1,
		pAttachments:    &att,
		subpassCount:    1,
	}
	if n := len(passDeps); n > 0 {
		inf.dependencyCount = C.uint32_t(n)
		inf.pDependencies = &passDeps[0]
	}

	var pass RenderPass
	if err := vkErr(C.vkCreateRenderPass(funcs.vkCreateRenderPass, d, inf, sub, nil, &pass)); err != nil {
		return nilRenderPass, fmt.Errorf("vulkan: vkCreateRenderPass: %w", err)
	}
	return pass, nil
}

func DestroyRenderPass(d Device, r RenderPass) {
	C.vkDestroyRenderPass(funcs.vkDestroyRenderPass, d, r, nil)
}

func CreateFramebuffer(d Device, rp RenderPass, view ImageView, width, height int) (Framebuffer, error) {
	inf := C.VkFramebufferCreateInfo{
		sType:           C.VK_STRUCTURE_TYPE_FRAMEBUFFER_CREATE_INFO,
		renderPass:      rp,
		attachmentCount: 1,
		pAttachments:    &view,
		width:           C.uint32_t(width),
		height:          C.uint32_t(height),
		layers:          1,
	}
	var fbo Framebuffer
	if err := vkErr(C.vkCreateFramebuffer(funcs.vkCreateFramebuffer, d, inf, nil, &fbo)); err != nil {
		return nilFramebuffer, fmt.Errorf("vulkan: vkCreateFramebuffer: %w", err)
	}
	return fbo, nil

}

func DestroyFramebuffer(d Device, f Framebuffer) {
	C.vkDestroyFramebuffer(funcs.vkDestroyFramebuffer, d, f, nil)
}

func DeviceWaitIdle(d Device) error {
	if err := vkErr(C.vkDeviceWaitIdle(funcs.vkDeviceWaitIdle, d)); err != nil {
		return fmt.Errorf("vulkan: vkDeviceWaitIdle: %w", err)
	}
	return nil
}

func QueueWaitIdle(q Queue) error {
	if err := vkErr(C.vkQueueWaitIdle(funcs.vkQueueWaitIdle, q)); err != nil {
		return fmt.Errorf("vulkan: vkQueueWaitIdle: %w", err)
	}
	return nil
}

func CreateSemaphore(d Device) (Semaphore, error) {
	inf := C.VkSemaphoreCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO,
	}
	var sem Semaphore
	err := vkErr(C.vkCreateSemaphore(funcs.vkCreateSemaphore, d, &inf, nil, &sem))
	if err != nil {
		return nilSemaphore, fmt.Errorf("vulkan: vkCreateSemaphore: %w", err)
	}
	return sem, err
}

func DestroySemaphore(d Device, sem Semaphore) {
	C.vkDestroySemaphore(funcs.vkDestroySemaphore, d, sem, nil)
}

func DestroyDevice(dev Device) {
	C.vkDestroyDevice(funcs.vkDestroyDevice, dev, nil)
}

func DestroySurface(inst Instance, s Surface) {
	C.vkDestroySurfaceKHR(funcs.vkDestroySurfaceKHR, inst, s, nil)
}

func CreateCommandPool(d Device, queueIndex int) (CommandPool, error) {
	inf := C.VkCommandPoolCreateInfo{
		sType:            C.VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO,
		queueFamilyIndex: C.uint32_t(queueIndex),
		flags:            C.VK_COMMAND_POOL_CREATE_TRANSIENT_BIT | C.VK_COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT,
	}

	var pool CommandPool
	if err := vkErr(C.vkCreateCommandPool(funcs.vkCreateCommandPool, d, &inf, nil, &pool)); err != nil {
		return nilCommandPool, fmt.Errorf("vulkan: vkCreateCommandPool: %w", err)
	}
	return pool, nil
}

func DestroyCommandPool(d Device, pool CommandPool) {
	C.vkDestroyCommandPool(funcs.vkDestroyCommandPool, d, pool, nil)
}

func AllocateCommandBuffer(d Device, pool CommandPool) (CommandBuffer, error) {
	inf := C.VkCommandBufferAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
		commandPool:        pool,
		level:              C.VK_COMMAND_BUFFER_LEVEL_PRIMARY,
		commandBufferCount: 1,
	}

	var buf CommandBuffer
	if err := vkErr(C.vkAllocateCommandBuffers(funcs.vkAllocateCommandBuffers, d, &inf, &buf)); err != nil {
		return nil, fmt.Errorf("vulkan: vkAllocateCommandBuffers: %w", err)
	}
	return buf, nil
}

func FreeCommandBuffers(d Device, pool CommandPool, bufs ...CommandBuffer) {
	if len(bufs) == 0 {
		return
	}
	C.vkFreeCommandBuffers(funcs.vkFreeCommandBuffers, d, pool, C.uint32_t(len(bufs)), &bufs[0])
}

func BeginCommandBuffer(buf CommandBuffer) error {
	inf := C.VkCommandBufferBeginInfo{
		sType: C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO,
		flags: C.VK_COMMAND_BUFFER_USAGE_ONE_TIME_SUBMIT_BIT,
	}
	if err := vkErr(C.vkBeginCommandBuffer(funcs.vkBeginCommandBuffer, buf, inf)); err != nil {
		return fmt.Errorf("vulkan: vkBeginCommandBuffer: %w", err)
	}
	return nil
}

func EndCommandBuffer(buf CommandBuffer) error {
	if err := vkErr(C.vkEndCommandBuffer(funcs.vkEndCommandBuffer, buf)); err != nil {
		return fmt.Errorf("vulkan: vkEndCommandBuffer: %w", err)
	}
	return nil
}

func QueueSubmit(q Queue, buf CommandBuffer, waitSems []Semaphore, waitStages []PipelineStageFlags, sigSems []Semaphore, fence Fence) error {
	inf := C.VkSubmitInfo{
		sType:              C.VK_STRUCTURE_TYPE_SUBMIT_INFO,
		commandBufferCount: 1,
		pCommandBuffers:    &buf,
	}
	if len(waitSems) > 0 {
		if len(waitSems) != len(waitStages) {
			panic("len(waitSems) != len(waitStages)")
		}
		inf.waitSemaphoreCount = C.uint32_t(len(waitSems))
		inf.pWaitSemaphores = &waitSems[0]
		inf.pWaitDstStageMask = &waitStages[0]
	}
	if len(sigSems) > 0 {
		inf.signalSemaphoreCount = C.uint32_t(len(sigSems))
		inf.pSignalSemaphores = &sigSems[0]
	}
	if err := vkErr(C.vkQueueSubmit(funcs.vkQueueSubmit, q, inf, fence)); err != nil {
		return fmt.Errorf("vulkan: vkQueueSubmit: %w", err)
	}
	return nil
}

func CmdBeginRenderPass(buf CommandBuffer, rp RenderPass, fbo Framebuffer, width, height int, clearCol [4]float32) {
	cclearCol := [4]C.float{C.float(clearCol[0]), C.float(clearCol[1]), C.float(clearCol[2]), C.float(clearCol[3])}
	inf := C.VkRenderPassBeginInfo{
		sType:           C.VK_STRUCTURE_TYPE_RENDER_PASS_BEGIN_INFO,
		renderPass:      rp,
		framebuffer:     fbo,
		renderArea:      C.VkRect2D{extent: C.VkExtent2D{width: C.uint32_t(width), height: C.uint32_t(height)}},
		clearValueCount: 1,
		pClearValues:    (*C.VkClearValue)(unsafe.Pointer(&cclearCol)),
	}
	C.vkCmdBeginRenderPass(funcs.vkCmdBeginRenderPass, buf, inf, C.VK_SUBPASS_CONTENTS_INLINE)
}

func CmdEndRenderPass(buf CommandBuffer) {
	C.vkCmdEndRenderPass(funcs.vkCmdEndRenderPass, buf)
}

func CmdCopyBuffer(cmdBuf CommandBuffer, src, dst Buffer, srcOff, dstOff, size int) {
	C.vkCmdCopyBuffer(funcs.vkCmdCopyBuffer, cmdBuf, src, dst, 1, &C.VkBufferCopy{
		srcOffset: C.VkDeviceSize(srcOff),
		dstOffset: C.VkDeviceSize(dstOff),
		size:      C.VkDeviceSize(size),
	})
}

func CmdCopyBufferToImage(cmdBuf CommandBuffer, src Buffer, dst Image, layout ImageLayout, copy BufferImageCopy) {
	C.vkCmdCopyBufferToImage(funcs.vkCmdCopyBufferToImage, cmdBuf, src, dst, layout, 1, &copy)
}

func CmdPipelineBarrier(cmdBuf CommandBuffer, srcStage, dstStage PipelineStageFlags, flags DependencyFlags, memBarriers []MemoryBarrier, bufBarriers []BufferMemoryBarrier, imgBarriers []ImageMemoryBarrier) {
	var memPtr *MemoryBarrier
	if len(memBarriers) > 0 {
		memPtr = &memBarriers[0]
	}
	var bufPtr *BufferMemoryBarrier
	if len(bufBarriers) > 0 {
		bufPtr = &bufBarriers[0]
	}
	var imgPtr *ImageMemoryBarrier
	if len(imgBarriers) > 0 {
		imgPtr = &imgBarriers[0]
	}
	C.vkCmdPipelineBarrier(funcs.vkCmdPipelineBarrier, cmdBuf, srcStage, dstStage, flags,
		C.uint32_t(len(memBarriers)), memPtr,
		C.uint32_t(len(bufBarriers)), bufPtr,
		C.uint32_t(len(imgBarriers)), imgPtr)
}

func CmdPushConstants(cmdBuf CommandBuffer, layout PipelineLayout, stages ShaderStageFlags, offset int, data []byte) {
	if len(data) == 0 {
		return
	}
	C.vkCmdPushConstants(funcs.vkCmdPushConstants, cmdBuf, layout, stages, C.uint32_t(offset), C.uint32_t(len(data)), unsafe.Pointer(&data[0]))
}

func CmdBindPipeline(cmdBuf CommandBuffer, bindPoint PipelineBindPoint, pipe Pipeline) {
	C.vkCmdBindPipeline(funcs.vkCmdBindPipeline, cmdBuf, bindPoint, pipe)
}

func CmdBindVertexBuffers(cmdBuf CommandBuffer, first int, buffers []Buffer, sizes []DeviceSize) {
	if len(buffers) == 0 {
		return
	}
	C.vkCmdBindVertexBuffers(funcs.vkCmdBindVertexBuffers, cmdBuf, C.uint32_t(first), C.uint32_t(len(buffers)), &buffers[0], &sizes[0])
}

func CmdSetViewport(cmdBuf CommandBuffer, first int, viewports ...Viewport) {
	if len(viewports) == 0 {
		return
	}
	C.vkCmdSetViewport(funcs.vkCmdSetViewport, cmdBuf, C.uint32_t(first), C.uint32_t(len(viewports)), &viewports[0])
}

func CmdBindIndexBuffer(cmdBuf CommandBuffer, buffer Buffer, offset int, typ IndexType) {
	C.vkCmdBindIndexBuffer(funcs.vkCmdBindIndexBuffer, cmdBuf, buffer, C.VkDeviceSize(offset), typ)
}

func CmdDraw(cmdBuf CommandBuffer, vertCount, instCount, firstVert, firstInst int) {
	C.vkCmdDraw(funcs.vkCmdDraw, cmdBuf, C.uint32_t(vertCount), C.uint32_t(instCount), C.uint32_t(firstVert), C.uint32_t(firstInst))
}

func CmdDrawIndexed(cmdBuf CommandBuffer, idxCount, instCount, firstIdx, vertOff, firstInst int) {
	C.vkCmdDrawIndexed(funcs.vkCmdDrawIndexed, cmdBuf, C.uint32_t(idxCount), C.uint32_t(instCount), C.uint32_t(firstIdx), C.int32_t(vertOff), C.uint32_t(firstInst))
}

func GetPhysicalDeviceFormatProperties(physDev PhysicalDevice, format Format) FormatFeatureFlags {
	var props C.VkFormatProperties
	C.vkGetPhysicalDeviceFormatProperties(funcs.vkGetPhysicalDeviceFormatProperties, physDev, format, &props)
	return FormatFeatureFlags(props.optimalTilingFeatures)
}

func CmdBindDescriptorSets(cmdBuf CommandBuffer, point PipelineBindPoint, layout PipelineLayout, firstSet int, sets []DescriptorSet) {
	C.vkCmdBindDescriptorSets(funcs.vkCmdBindDescriptorSets, cmdBuf, point, layout, C.uint32_t(firstSet), C.uint32_t(len(sets)), &sets[0], 0, nil)
}

func CmdBlitImage(cmdBuf CommandBuffer, src Image, srcLayout ImageLayout, dst Image, dstLayout ImageLayout, regions []ImageBlit, filter Filter) {
	if len(regions) == 0 {
		return
	}
	C.vkCmdBlitImage(funcs.vkCmdBlitImage, cmdBuf, src, srcLayout, dst, dstLayout, C.uint32_t(len(regions)), &regions[0], filter)
}

func CmdCopyImage(cmdBuf CommandBuffer, src Image, srcLayout ImageLayout, dst Image, dstLayout ImageLayout, regions []ImageCopy) {
	if len(regions) == 0 {
		return
	}
	C.vkCmdCopyImage(funcs.vkCmdCopyImage, cmdBuf, src, srcLayout, dst, dstLayout, C.uint32_t(len(regions)), &regions[0])
}

func CmdCopyImageToBuffer(cmdBuf CommandBuffer, src Image, srcLayout ImageLayout, dst Buffer, regions []BufferImageCopy) {
	if len(regions) == 0 {
		return
	}
	C.vkCmdCopyImageToBuffer(funcs.vkCmdCopyImageToBuffer, cmdBuf, src, srcLayout, dst, C.uint32_t(len(regions)), &regions[0])
}

func CmdDispatch(cmdBuf CommandBuffer, x, y, z int) {
	C.vkCmdDispatch(funcs.vkCmdDispatch, cmdBuf, C.uint32_t(x), C.uint32_t(y), C.uint32_t(z))
}

func CreateImage(pd PhysicalDevice, d Device, format Format, width, height, mipmaps int, usage ImageUsageFlags) (Image, DeviceMemory, error) {
	inf := C.VkImageCreateInfo{
		sType:     C.VK_STRUCTURE_TYPE_IMAGE_CREATE_INFO,
		imageType: C.VK_IMAGE_TYPE_2D,
		format:    format,
		extent: C.VkExtent3D{
			width:  C.uint32_t(width),
			height: C.uint32_t(height),
			depth:  1,
		},
		mipLevels:     C.uint32_t(mipmaps),
		arrayLayers:   1,
		samples:       C.VK_SAMPLE_COUNT_1_BIT,
		tiling:        C.VK_IMAGE_TILING_OPTIMAL,
		usage:         usage,
		initialLayout: C.VK_IMAGE_LAYOUT_UNDEFINED,
	}
	var img C.VkImage
	if err := vkErr(C.vkCreateImage(funcs.vkCreateImage, d, &inf, nil, &img)); err != nil {
		return nilImage, nilDeviceMemory, fmt.Errorf("vulkan: vkCreateImage: %w", err)
	}
	var memReqs C.VkMemoryRequirements
	C.vkGetImageMemoryRequirements(funcs.vkGetImageMemoryRequirements, d, img, &memReqs)

	memIdx, found := findMemoryTypeIndex(pd, memReqs.memoryTypeBits, C.VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT)
	if !found {
		DestroyImage(d, img)
		return nilImage, nilDeviceMemory, errors.New("vulkan: no memory type suitable for images found")
	}

	memInf := C.VkMemoryAllocateInfo{
		sType:           C.VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		allocationSize:  memReqs.size,
		memoryTypeIndex: C.uint32_t(memIdx),
	}
	var imgMem C.VkDeviceMemory
	if err := vkErr(C.vkAllocateMemory(funcs.vkAllocateMemory, d, &memInf, nil, &imgMem)); err != nil {
		DestroyImage(d, img)
		return nilImage, nilDeviceMemory, fmt.Errorf("vulkan: vkAllocateMemory: %w", err)
	}

	if err := vkErr(C.vkBindImageMemory(funcs.vkBindImageMemory, d, img, imgMem, 0)); err != nil {
		FreeMemory(d, imgMem)
		DestroyImage(d, img)
		return nilImage, nilDeviceMemory, fmt.Errorf("vulkan: vkBindImageMemory: %w", err)
	}
	return img, imgMem, nil
}

func DestroyImage(d Device, img Image) {
	C.vkDestroyImage(funcs.vkDestroyImage, d, img, nil)
}

func FreeMemory(d Device, mem DeviceMemory) {
	C.vkFreeMemory(funcs.vkFreeMemory, d, mem, nil)
}

func CreateSampler(d Device, minFilter, magFilter Filter, mipmapMode SamplerMipmapMode) (Sampler, error) {
	inf := C.VkSamplerCreateInfo{
		sType:        C.VK_STRUCTURE_TYPE_SAMPLER_CREATE_INFO,
		minFilter:    minFilter,
		magFilter:    magFilter,
		mipmapMode:   mipmapMode,
		maxLod:       C.VK_LOD_CLAMP_NONE,
		addressModeU: C.VK_SAMPLER_ADDRESS_MODE_CLAMP_TO_EDGE,
		addressModeV: C.VK_SAMPLER_ADDRESS_MODE_CLAMP_TO_EDGE,
	}
	var s C.VkSampler
	if err := vkErr(C.vkCreateSampler(funcs.vkCreateSampler, d, &inf, nil, &s)); err != nil {
		return nilSampler, fmt.Errorf("vulkan: vkCreateSampler: %w", err)
	}
	return s, nil
}

func DestroySampler(d Device, sampler Sampler) {
	C.vkDestroySampler(funcs.vkDestroySampler, d, sampler, nil)
}

func CreateBuffer(pd PhysicalDevice, d Device, size int, usage BufferUsageFlags, props MemoryPropertyFlags) (Buffer, DeviceMemory, error) {
	inf := C.VkBufferCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO,
		size:  C.VkDeviceSize(size),
		usage: usage,
	}
	var buf C.VkBuffer
	if err := vkErr(C.vkCreateBuffer(funcs.vkCreateBuffer, d, &inf, nil, &buf)); err != nil {
		return nilBuffer, nilDeviceMemory, fmt.Errorf("vulkan: vkCreateBuffer: %w", err)
	}

	var memReqs C.VkMemoryRequirements
	C.vkGetBufferMemoryRequirements(funcs.vkGetBufferMemoryRequirements, d, buf, &memReqs)

	memIdx, found := findMemoryTypeIndex(pd, memReqs.memoryTypeBits, props)
	if !found {
		DestroyBuffer(d, buf)
		return nilBuffer, nilDeviceMemory, errors.New("vulkan: no memory suitable for buffers found")
	}
	memInf := C.VkMemoryAllocateInfo{
		sType:           C.VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		allocationSize:  memReqs.size,
		memoryTypeIndex: C.uint32_t(memIdx),
	}

	var mem C.VkDeviceMemory
	if err := vkErr(C.vkAllocateMemory(funcs.vkAllocateMemory, d, &memInf, nil, &mem)); err != nil {
		DestroyBuffer(d, buf)
		return nilBuffer, nilDeviceMemory, fmt.Errorf("vulkan: vkAllocateMemory: %w", err)
	}

	if err := vkErr(C.vkBindBufferMemory(funcs.vkBindBufferMemory, d, buf, mem, 0)); err != nil {
		FreeMemory(d, mem)
		DestroyBuffer(d, buf)
		return nilBuffer, nilDeviceMemory, fmt.Errorf("vulkan: vkBindBufferMemory: %w", err)
	}
	return buf, mem, nil
}

func DestroyBuffer(d Device, buf Buffer) {
	C.vkDestroyBuffer(funcs.vkDestroyBuffer, d, buf, nil)
}

func CreateShaderModule(d Device, spirv string) (ShaderModule, error) {
	ptr := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&spirv)).Data)
	inf := C.VkShaderModuleCreateInfo{
		sType:    C.VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO,
		codeSize: C.size_t(len(spirv)),
		pCode:    (*C.uint32_t)(ptr),
	}

	var mod C.VkShaderModule
	if err := vkErr(C.vkCreateShaderModule(funcs.vkCreateShaderModule, d, inf, nil, &mod)); err != nil {
		return nilShaderModule, fmt.Errorf("vulkan: vkCreateShaderModule: %w", err)
	}
	return mod, nil
}

func DestroyShaderModule(d Device, mod ShaderModule) {
	C.vkDestroyShaderModule(funcs.vkDestroyShaderModule, d, mod, nil)
}

func CreateGraphicsPipeline(d Device, pass RenderPass, vmod, fmod ShaderModule, blend bool, srcFactor, dstFactor BlendFactor, topology PrimitiveTopology, bindings []VertexInputBindingDescription, attrs []VertexInputAttributeDescription, layout PipelineLayout) (Pipeline, error) {
	main := C.CString("main")
	defer C.free(unsafe.Pointer(main))
	stages := []C.VkPipelineShaderStageCreateInfo{
		{
			sType:  C.VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
			stage:  C.VK_SHADER_STAGE_VERTEX_BIT,
			module: vmod,
			pName:  main,
		},
		{
			sType:  C.VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
			stage:  C.VK_SHADER_STAGE_FRAGMENT_BIT,
			module: fmod,
			pName:  main,
		},
	}
	dynStates := []C.VkDynamicState{C.VK_DYNAMIC_STATE_VIEWPORT}
	dynInf := C.VkPipelineDynamicStateCreateInfo{
		sType:             C.VK_STRUCTURE_TYPE_PIPELINE_DYNAMIC_STATE_CREATE_INFO,
		dynamicStateCount: C.uint32_t(len(dynStates)),
		pDynamicStates:    &dynStates[0],
	}
	const maxDim = 0x7fffffff
	scissors := []C.VkRect2D{{extent: C.VkExtent2D{width: maxDim, height: maxDim}}}
	viewportInf := C.VkPipelineViewportStateCreateInfo{
		sType:         C.VK_STRUCTURE_TYPE_PIPELINE_VIEWPORT_STATE_CREATE_INFO,
		viewportCount: 1,
		scissorCount:  C.uint32_t(len(scissors)),
		pScissors:     &scissors[0],
	}
	enable := C.VkBool32(0)
	if blend {
		enable = 1
	}
	attBlendInf := C.VkPipelineColorBlendAttachmentState{
		blendEnable:         enable,
		srcColorBlendFactor: srcFactor,
		srcAlphaBlendFactor: srcFactor,
		dstColorBlendFactor: dstFactor,
		dstAlphaBlendFactor: dstFactor,
		colorBlendOp:        C.VK_BLEND_OP_ADD,
		alphaBlendOp:        C.VK_BLEND_OP_ADD,
		colorWriteMask:      C.VK_COLOR_COMPONENT_R_BIT | C.VK_COLOR_COMPONENT_G_BIT | C.VK_COLOR_COMPONENT_B_BIT | C.VK_COLOR_COMPONENT_A_BIT,
	}
	blendInf := C.VkPipelineColorBlendStateCreateInfo{
		sType:           C.VK_STRUCTURE_TYPE_PIPELINE_COLOR_BLEND_STATE_CREATE_INFO,
		attachmentCount: 1,
		pAttachments:    &attBlendInf,
	}
	var vkBinds []C.VkVertexInputBindingDescription
	var vkAttrs []C.VkVertexInputAttributeDescription
	for _, b := range bindings {
		vkBinds = append(vkBinds, C.VkVertexInputBindingDescription{
			binding: C.uint32_t(b.Binding),
			stride:  C.uint32_t(b.Stride),
		})
	}
	for _, a := range attrs {
		vkAttrs = append(vkAttrs, C.VkVertexInputAttributeDescription{
			location: C.uint32_t(a.Location),
			binding:  C.uint32_t(a.Binding),
			format:   a.Format,
			offset:   C.uint32_t(a.Offset),
		})
	}
	vertexInf := C.VkPipelineVertexInputStateCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_PIPELINE_VERTEX_INPUT_STATE_CREATE_INFO,
	}
	if n := len(vkBinds); n > 0 {
		vertexInf.vertexBindingDescriptionCount = C.uint32_t(n)
		vertexInf.pVertexBindingDescriptions = &vkBinds[0]
	}
	if n := len(vkAttrs); n > 0 {
		vertexInf.vertexAttributeDescriptionCount = C.uint32_t(n)
		vertexInf.pVertexAttributeDescriptions = &vkAttrs[0]
	}
	inf := C.VkGraphicsPipelineCreateInfo{
		sType:      C.VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO,
		stageCount: C.uint32_t(len(stages)),
		pStages:    &stages[0],
		renderPass: pass,
		layout:     layout,
		pRasterizationState: &C.VkPipelineRasterizationStateCreateInfo{
			sType:     C.VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO,
			lineWidth: 1.0,
		},
		pMultisampleState: &C.VkPipelineMultisampleStateCreateInfo{
			sType:                C.VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO,
			rasterizationSamples: C.VK_SAMPLE_COUNT_1_BIT,
		},
		pInputAssemblyState: &C.VkPipelineInputAssemblyStateCreateInfo{
			sType:    C.VK_STRUCTURE_TYPE_PIPELINE_INPUT_ASSEMBLY_STATE_CREATE_INFO,
			topology: topology,
		},
	}

	var pipe C.VkPipeline
	if err := vkErr(C.vkCreateGraphicsPipelines(funcs.vkCreateGraphicsPipelines, d, nilPipelineCache, inf, dynInf, blendInf, vertexInf, viewportInf, nil, &pipe)); err != nil {
		return nilPipeline, fmt.Errorf("vulkan: vkCreateGraphicsPipelines: %w", err)
	}
	return pipe, nil
}

func DestroyPipeline(d Device, p Pipeline) {
	C.vkDestroyPipeline(funcs.vkDestroyPipeline, d, p, nil)
}

func CreatePipelineLayout(d Device, pushRanges []PushConstantRange, sets []DescriptorSetLayout) (PipelineLayout, error) {
	inf := C.VkPipelineLayoutCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO,
	}
	if n := len(sets); n > 0 {
		inf.setLayoutCount = C.uint32_t(n)
		inf.pSetLayouts = &sets[0]
	}
	if n := len(pushRanges); n > 0 {
		inf.pushConstantRangeCount = C.uint32_t(n)
		inf.pPushConstantRanges = &pushRanges[0]
	}
	var l C.VkPipelineLayout
	if err := vkErr(C.vkCreatePipelineLayout(funcs.vkCreatePipelineLayout, d, inf, nil, &l)); err != nil {
		return nilPipelineLayout, fmt.Errorf("vulkan: vkCreatePipelineLayout: %w", err)
	}
	return l, nil
}

func DestroyPipelineLayout(d Device, l PipelineLayout) {
	C.vkDestroyPipelineLayout(funcs.vkDestroyPipelineLayout, d, l, nil)
}

func CreateDescriptorSetLayout(d Device, bindings []DescriptorSetLayoutBinding) (DescriptorSetLayout, error) {
	var vkbinds []C.VkDescriptorSetLayoutBinding
	for _, b := range bindings {
		vkbinds = append(vkbinds, C.VkDescriptorSetLayoutBinding{
			binding:         C.uint32_t(b.Binding),
			descriptorType:  b.DescriptorType,
			descriptorCount: 1,
			stageFlags:      b.StageFlags,
		})
	}
	inf := C.VkDescriptorSetLayoutCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_DESCRIPTOR_SET_LAYOUT_CREATE_INFO,
	}
	if n := len(vkbinds); n > 0 {
		inf.bindingCount = C.uint32_t(len(vkbinds))
		inf.pBindings = &vkbinds[0]
	}
	var l C.VkDescriptorSetLayout
	if err := vkErr(C.vkCreateDescriptorSetLayout(funcs.vkCreateDescriptorSetLayout, d, inf, nil, &l)); err != nil {
		return nilDescriptorSetLayout, fmt.Errorf("vulkan: vkCreateDescriptorSetLayout: %w", err)
	}
	return l, nil
}

func DestroyDescriptorSetLayout(d Device, l DescriptorSetLayout) {
	C.vkDestroyDescriptorSetLayout(funcs.vkDestroyDescriptorSetLayout, d, l, nil)
}

func MapMemory(d Device, mem DeviceMemory, offset, size int) ([]byte, error) {
	var ptr unsafe.Pointer
	if err := vkErr(C.vkMapMemory(funcs.vkMapMemory, d, mem, C.VkDeviceSize(offset), C.VkDeviceSize(size), 0, &ptr)); err != nil {
		return nil, fmt.Errorf("vulkan: vkMapMemory: %w", err)
	}
	return ((*[1 << 30]byte)(ptr))[:size:size], nil
}

func UnmapMemory(d Device, mem DeviceMemory) {
	C.vkUnmapMemory(funcs.vkUnmapMemory, d, mem)
}

func ResetCommandBuffer(buf CommandBuffer) error {
	if err := vkErr(C.vkResetCommandBuffer(funcs.vkResetCommandBuffer, buf, 0)); err != nil {
		return fmt.Errorf("vulkan: vkResetCommandBuffer. %w", err)
	}
	return nil
}

func CreateDescriptorPool(d Device, maxSets int, sizes []DescriptorPoolSize) (DescriptorPool, error) {
	inf := C.VkDescriptorPoolCreateInfo{
		sType:         C.VK_STRUCTURE_TYPE_DESCRIPTOR_POOL_CREATE_INFO,
		maxSets:       C.uint32_t(maxSets),
		poolSizeCount: C.uint32_t(len(sizes)),
		pPoolSizes:    &sizes[0],
	}
	var pool C.VkDescriptorPool
	if err := vkErr(C.vkCreateDescriptorPool(funcs.vkCreateDescriptorPool, d, inf, nil, &pool)); err != nil {
		return nilDescriptorPool, fmt.Errorf("vulkan: vkCreateDescriptorPool: %w", err)
	}
	return pool, nil
}

func DestroyDescriptorPool(d Device, pool DescriptorPool) {
	C.vkDestroyDescriptorPool(funcs.vkDestroyDescriptorPool, d, pool, nil)
}

func ResetDescriptorPool(d Device, pool DescriptorPool) error {
	if err := vkErr(C.vkResetDescriptorPool(funcs.vkResetDescriptorPool, d, pool, 0)); err != nil {
		return fmt.Errorf("vulkan: vkResetDescriptorPool: %w", err)
	}
	return nil
}

func UpdateDescriptorSet(d Device, write WriteDescriptorSet) {
	C.vkUpdateDescriptorSets(funcs.vkUpdateDescriptorSets, d, write, 0, nil)
}

func AllocateDescriptorSets(d Device, pool DescriptorPool, layout DescriptorSetLayout, count int) ([]DescriptorSet, error) {
	layouts := make([]DescriptorSetLayout, count)
	for i := range layouts {
		layouts[i] = layout
	}
	inf := C.VkDescriptorSetAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_DESCRIPTOR_SET_ALLOCATE_INFO,
		descriptorPool:     pool,
		descriptorSetCount: C.uint32_t(count),
		pSetLayouts:        &layouts[0],
	}
	sets := make([]DescriptorSet, count)
	if err := vkErr(C.vkAllocateDescriptorSets(funcs.vkAllocateDescriptorSets, d, inf, &sets[0])); err != nil {
		return nil, fmt.Errorf("vulkan: vkAllocateDescriptorSets: %w", err)
	}
	return sets, nil
}

func CreateComputePipeline(d Device, mod ShaderModule, layout PipelineLayout) (Pipeline, error) {
	main := C.CString("main")
	defer C.free(unsafe.Pointer(main))
	inf := C.VkComputePipelineCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_COMPUTE_PIPELINE_CREATE_INFO,
		stage: C.VkPipelineShaderStageCreateInfo{
			sType:  C.VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
			stage:  C.VK_SHADER_STAGE_COMPUTE_BIT,
			module: mod,
			pName:  main,
		},
		layout: layout,
	}
	var pipe C.VkPipeline
	if err := vkErr(C.vkCreateComputePipelines(funcs.vkCreateComputePipelines, d, nilPipelineCache, 1, &inf, nil, &pipe)); err != nil {
		return nilPipeline, fmt.Errorf("vulkan: vkCreateComputePipelines: %w", err)
	}
	return pipe, nil
}

func CreateFence(d Device, flags int) (Fence, error) {
	inf := C.VkFenceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
		flags: C.VkFenceCreateFlags(flags),
	}
	var f C.VkFence
	if err := vkErr(C.vkCreateFence(funcs.vkCreateFence, d, &inf, nil, &f)); err != nil {
		return nilFence, fmt.Errorf("vulkan: vkCreateFence: %w", err)
	}
	return f, nil
}

func DestroyFence(d Device, f Fence) {
	C.vkDestroyFence(funcs.vkDestroyFence, d, f, nil)
}

func WaitForFences(d Device, fences ...Fence) error {
	if len(fences) == 0 {
		return nil
	}
	err := vkErr(C.vkWaitForFences(funcs.vkWaitForFences, d, C.uint32_t(len(fences)), &fences[0], C.VK_TRUE, 0xffffffffffffffff))
	if err != nil {
		return fmt.Errorf("vulkan: vkWaitForFences: %w", err)
	}
	return nil
}

func ResetFences(d Device, fences ...Fence) error {
	if len(fences) == 0 {
		return nil
	}
	err := vkErr(C.vkResetFences(funcs.vkResetFences, d, C.uint32_t(len(fences)), &fences[0]))
	if err != nil {
		return fmt.Errorf("vulkan: vkResetFences: %w", err)
	}
	return nil
}

func BuildSubpassDependency(srcStage, dstStage PipelineStageFlags, srcMask, dstMask AccessFlags, flags DependencyFlags) SubpassDependency {
	return C.VkSubpassDependency{
		srcSubpass:      C.VK_SUBPASS_EXTERNAL,
		srcStageMask:    srcStage,
		srcAccessMask:   srcMask,
		dstSubpass:      0,
		dstStageMask:    dstStage,
		dstAccessMask:   dstMask,
		dependencyFlags: flags,
	}
}

func BuildPushConstantRange(stages ShaderStageFlags, offset, size int) PushConstantRange {
	return C.VkPushConstantRange{
		stageFlags: stages,
		offset:     C.uint32_t(offset),
		size:       C.uint32_t(size),
	}
}

func BuildDescriptorPoolSize(typ DescriptorType, count int) DescriptorPoolSize {
	return C.VkDescriptorPoolSize{
		_type:           typ,
		descriptorCount: C.uint32_t(count),
	}
}

func BuildWriteDescriptorSetImage(set DescriptorSet, binding int, typ DescriptorType, sampler Sampler, view ImageView, layout ImageLayout) WriteDescriptorSet {
	return C.VkWriteDescriptorSet{
		sType:           C.VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET,
		dstSet:          set,
		dstBinding:      C.uint32_t(binding),
		descriptorCount: 1,
		descriptorType:  typ,
		pImageInfo: &C.VkDescriptorImageInfo{
			sampler:     sampler,
			imageView:   view,
			imageLayout: layout,
		},
	}
}

func BuildWriteDescriptorSetBuffer(set DescriptorSet, binding int, typ DescriptorType, buf Buffer) WriteDescriptorSet {
	return C.VkWriteDescriptorSet{
		sType:           C.VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET,
		dstSet:          set,
		dstBinding:      C.uint32_t(binding),
		descriptorCount: 1,
		descriptorType:  typ,
		pBufferInfo: &C.VkDescriptorBufferInfo{
			buffer: buf,
			_range: C.VK_WHOLE_SIZE,
		},
	}
}

func (r PushConstantRange) StageFlags() ShaderStageFlags {
	return r.stageFlags
}

func (r PushConstantRange) Offset() int {
	return int(r.offset)
}

func (r PushConstantRange) Size() int {
	return int(r.size)
}

func (p QueueFamilyProperties) Flags() QueueFlags {
	return p.queueFlags
}

func (c SurfaceCapabilities) MinExtent() image.Point {
	return image.Pt(int(c.minImageExtent.width), int(c.minImageExtent.height))
}

func (c SurfaceCapabilities) MaxExtent() image.Point {
	return image.Pt(int(c.maxImageExtent.width), int(c.maxImageExtent.height))
}

func BuildViewport(x, y, width, height float32) Viewport {
	return C.VkViewport{
		x:        C.float(x),
		y:        C.float(y),
		width:    C.float(width),
		height:   C.float(height),
		maxDepth: 1.0,
	}
}

func BuildImageMemoryBarrier(img Image, srcMask, dstMask AccessFlags, oldLayout, newLayout ImageLayout, baseMip, numMips int) ImageMemoryBarrier {
	return C.VkImageMemoryBarrier{
		sType:         C.VK_STRUCTURE_TYPE_IMAGE_MEMORY_BARRIER,
		srcAccessMask: srcMask,
		dstAccessMask: dstMask,
		oldLayout:     oldLayout,
		newLayout:     newLayout,
		image:         img,
		subresourceRange: C.VkImageSubresourceRange{
			aspectMask:   C.VK_IMAGE_ASPECT_COLOR_BIT,
			baseMipLevel: C.uint32_t(baseMip),
			levelCount:   C.uint32_t(numMips),
			layerCount:   C.VK_REMAINING_ARRAY_LAYERS,
		},
	}
}

func BuildBufferMemoryBarrier(buf Buffer, srcMask, dstMask AccessFlags) BufferMemoryBarrier {
	return C.VkBufferMemoryBarrier{
		sType:         C.VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		srcAccessMask: srcMask,
		dstAccessMask: dstMask,
		buffer:        buf,
		size:          C.VK_WHOLE_SIZE,
	}
}

func BuildMemoryBarrier(srcMask, dstMask AccessFlags) MemoryBarrier {
	return C.VkMemoryBarrier{
		sType:         C.VK_STRUCTURE_TYPE_MEMORY_BARRIER,
		srcAccessMask: srcMask,
		dstAccessMask: dstMask,
	}
}

func BuildBufferImageCopy(bufOff, bufStride, x, y, width, height int) BufferImageCopy {
	return C.VkBufferImageCopy{
		bufferOffset:    C.VkDeviceSize(bufOff),
		bufferRowLength: C.uint32_t(bufStride),
		imageSubresource: C.VkImageSubresourceLayers{
			aspectMask: C.VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
		},
		imageOffset: C.VkOffset3D{
			x: C.int32_t(x), y: C.int32_t(y), z: 0,
		},
		imageExtent: C.VkExtent3D{
			width: C.uint32_t(width), height: C.uint32_t(height), depth: 1,
		},
	}
}

func BuildImageCopy(srcX, srcY, dstX, dstY, width, height int) ImageCopy {
	return C.VkImageCopy{
		srcSubresource: C.VkImageSubresourceLayers{
			aspectMask: C.VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
		},
		srcOffset: C.VkOffset3D{
			x: C.int32_t(srcX),
			y: C.int32_t(srcY),
		},
		dstSubresource: C.VkImageSubresourceLayers{
			aspectMask: C.VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
		},
		dstOffset: C.VkOffset3D{
			x: C.int32_t(dstX),
			y: C.int32_t(dstY),
		},
		extent: C.VkExtent3D{
			width:  C.uint32_t(width),
			height: C.uint32_t(height),
			depth:  1,
		},
	}
}

func BuildImageBlit(srcX, srcY, dstX, dstY, srcWidth, srcHeight, dstWidth, dstHeight, srcMip, dstMip int) ImageBlit {
	return C.VkImageBlit{
		srcOffsets: [2]C.VkOffset3D{
			{C.int32_t(srcX), C.int32_t(srcY), 0},
			{C.int32_t(srcWidth), C.int32_t(srcHeight), 1},
		},
		srcSubresource: C.VkImageSubresourceLayers{
			aspectMask: C.VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
			mipLevel:   C.uint32_t(srcMip),
		},
		dstOffsets: [2]C.VkOffset3D{
			{C.int32_t(dstX), C.int32_t(dstY), 0},
			{C.int32_t(dstWidth), C.int32_t(dstHeight), 1},
		},
		dstSubresource: C.VkImageSubresourceLayers{
			aspectMask: C.VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
			mipLevel:   C.uint32_t(dstMip),
		},
	}
}

func findMemoryTypeIndex(pd C.VkPhysicalDevice, constraints C.uint32_t, wantProps C.VkMemoryPropertyFlags) (int, bool) {
	var memProps C.VkPhysicalDeviceMemoryProperties
	C.vkGetPhysicalDeviceMemoryProperties(funcs.vkGetPhysicalDeviceMemoryProperties, pd, &memProps)

	for i := 0; i < int(memProps.memoryTypeCount); i++ {
		if (constraints & (1 << i)) == 0 {
			continue
		}
		if (memProps.memoryTypes[i].propertyFlags & wantProps) != wantProps {
			continue
		}
		return i, true
	}

	return 0, false
}

func choosePresentMode(pd C.VkPhysicalDevice, surf Surface) (C.VkPresentModeKHR, bool, error) {
	var count C.uint32_t
	err := vkErr(C.vkGetPhysicalDeviceSurfacePresentModesKHR(funcs.vkGetPhysicalDeviceSurfacePresentModesKHR, pd, surf, &count, nil))
	if err != nil {
		return 0, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfacePresentModesKHR: %w", err)
	}
	if count == 0 {
		return 0, false, nil
	}
	modes := make([]C.VkPresentModeKHR, count)
	err = vkErr(C.vkGetPhysicalDeviceSurfacePresentModesKHR(funcs.vkGetPhysicalDeviceSurfacePresentModesKHR, pd, surf, &count, &modes[0]))
	if err != nil {
		return 0, false, fmt.Errorf("vulkan: kGetPhysicalDeviceSurfacePresentModesKHR: %w", err)
	}
	for _, m := range modes {
		if m == C.VK_PRESENT_MODE_MAILBOX_KHR || m == C.VK_PRESENT_MODE_FIFO_KHR {
			return m, true, nil
		}
	}
	return 0, false, nil
}

func chooseFormat(pd C.VkPhysicalDevice, surf Surface) (C.VkSurfaceFormatKHR, bool, error) {
	var count C.uint32_t
	err := vkErr(C.vkGetPhysicalDeviceSurfaceFormatsKHR(funcs.vkGetPhysicalDeviceSurfaceFormatsKHR, pd, surf, &count, nil))
	if err != nil {
		return C.VkSurfaceFormatKHR{}, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceFormatsKHR: %w", err)
	}
	if count == 0 {
		return C.VkSurfaceFormatKHR{}, false, nil
	}
	formats := make([]C.VkSurfaceFormatKHR, count)
	err = vkErr(C.vkGetPhysicalDeviceSurfaceFormatsKHR(funcs.vkGetPhysicalDeviceSurfaceFormatsKHR, pd, surf, &count, &formats[0]))
	if err != nil {
		return C.VkSurfaceFormatKHR{}, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceFormatsKHR: %w", err)
	}
	// Query for format with sRGB support.
	// TODO: Support devices without sRGB.
	for _, f := range formats {
		if f.colorSpace != C.VK_COLOR_SPACE_SRGB_NONLINEAR_KHR {
			continue
		}
		switch f.format {
		case C.VK_FORMAT_B8G8R8A8_SRGB, C.VK_FORMAT_R8G8B8A8_SRGB:
			return f, true, nil
		}
	}
	return C.VkSurfaceFormatKHR{}, false, nil
}

func chooseQueue(pd C.VkPhysicalDevice, surf Surface, flags C.VkQueueFlags) (int, bool, error) {
	queues := GetPhysicalDeviceQueueFamilyProperties(pd)
	for i, q := range queues {
		// Check for presentation and feature support.
		if q.queueFlags&flags != flags {
			continue
		}
		if surf != nilSurface {
			// Check for presentation support. It is possible that a device has no
			// queue with both rendering and presentation support, but not in reality.
			// See https://github.com/KhronosGroup/Vulkan-Docs/issues/1234.
			var support C.VkBool32
			if err := vkErr(C.vkGetPhysicalDeviceSurfaceSupportKHR(funcs.vkGetPhysicalDeviceSurfaceSupportKHR, pd, C.uint32_t(i), surf, &support)); err != nil {
				return 0, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceSupportKHR: %w", err)
			}
			if support != C.VK_TRUE {
				continue
			}
		}
		return i, true, nil
	}
	return 0, false, nil
}

func dlsym(handle unsafe.Pointer, s string) unsafe.Pointer {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.dlsym(handle, cs)
}

func dlopen(lib string) unsafe.Pointer {
	clib := C.CString(lib)
	defer C.free(unsafe.Pointer(clib))
	return C.dlopen(clib, C.RTLD_NOW|C.RTLD_LOCAL)
}

func vkErr(res C.VkResult) error {
	switch res {
	case C.VK_SUCCESS:
		return nil
	default:
		return Error(res)
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("error %d", e)
}
