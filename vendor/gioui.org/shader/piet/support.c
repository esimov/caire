// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && (arm64 || arm || amd64)
// +build linux
// +build arm64 arm amd64

#include <stdint.h>
#include <stdlib.h>
#include <assert.h>
#include "abi.h"
#include "runtime.h"

static void *malloc_align(size_t alignment, size_t size) {
	void *ptr;
	int ret = posix_memalign(&ptr, alignment, size);
	assert(ret == 0);
	return ptr;
}

ATTR_HIDDEN void *coroutine_alloc_frame(size_t size) {
	void *ptr = malloc_align(16, size);
	return ptr;
}

ATTR_HIDDEN void coroutine_free_frame(void *ptr) {
	free(ptr);
}
