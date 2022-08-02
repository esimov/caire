// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && (arm64 || arm || amd64)
// +build linux
// +build arm64 arm amd64

#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <math.h>
#include <stdlib.h>
#include <assert.h>
#include "abi.h"
#include "runtime.h"
#include "_cgo_export.h"

// coroutines is a FIFO queue of coroutines implemented as a circular
// buffer.
struct coroutines {
	coroutine *routines;
	// start and end indexes into routines.
	unsigned int start;
	unsigned int end;
	// cap is the capacity of routines.
	unsigned int cap;
};

struct dispatch_context {
	// descriptor_set is the aligned storage for the descriptor set.
	void *descriptor_set;
	int desc_set_size;

	int nthreads;
	bool has_cbarriers;
	size_t memory_size;
	// Program entrypoints.
	routine_begin begin;
	routine_await await;
	routine_destroy destroy;

	struct program_data data;
};

struct thread_context {
	struct coroutines routines;
	size_t memory_size;
	uint8_t *memory;
};

static void *malloc_align(size_t alignment, size_t size) {
	void *ptr;
	int ret = posix_memalign(&ptr, alignment, size);
	assert(ret == 0);
	return ptr;
}

static void coroutines_dump(struct coroutines *routines) {
	fprintf(stderr, "s: %d e: %d c: %d [", routines->start, routines->end, routines->cap);
	unsigned int i = routines->start;
	while (i != routines->end) {
		fprintf(stderr, "%p,", routines->routines[routines->start]);
		i = (i + 1)%routines->cap;
	}
	fprintf(stderr, "]\n");
}

static void coroutines_push(struct coroutines *routines, coroutine r) {
	unsigned int next = routines->end + 1;
	if (next >= routines->cap) {
		next = 0;
	}
	if (next == routines->start) {
		unsigned int newcap = routines->cap*2;
		if (newcap < 10) {
			newcap = 10;
		}
		routines->routines = realloc(routines->routines, newcap*sizeof(coroutine));
		// Move elements wrapped around the old cap to the newly allocated space.
		if (routines->end < routines->start) {
			unsigned int nelems = routines->end;
			unsigned int max = newcap - routines->cap;
			// We doubled the space above, so we can assume enough room.
			assert(nelems <= max);
			memmove(&routines->routines[routines->cap], &routines->routines[0], nelems*sizeof(coroutine));
			routines->end += routines->cap;
		}
		routines->cap = newcap;
		next = (routines->end + 1)%routines->cap;
	}
	routines->routines[routines->end] = r;
	routines->end = next;
}

static bool coroutines_pop(struct coroutines *routines, coroutine *r) {
	if (routines->start == routines->end) {
		return 0;
	}
	*r = routines->routines[routines->start];
	routines->start = (routines->start + 1)%routines->cap;
	return 1;
}

static void coroutines_free(struct coroutines *routines) {
	if (routines->routines != NULL) {
		free(routines->routines);
	}
	struct coroutines clr = { 0 }; *routines = clr;
}

struct dispatch_context *alloc_dispatch_context(void) {
	struct dispatch_context *c = malloc(sizeof(*c));
	assert(c != NULL);
	struct dispatch_context clr = { 0 }; *c = clr;
	return c;
}

void free_dispatch_context(struct dispatch_context *c) {
	if (c->descriptor_set != NULL) {
		free(c->descriptor_set);
		c->descriptor_set = NULL;
	}
}

struct thread_context *alloc_thread_context(void) {
	struct thread_context *c = malloc(sizeof(*c));
	assert(c != NULL);
	struct thread_context clr = { 0 }; *c = clr;
	return c;
}

void free_thread_context(struct thread_context *c) {
	if (c->memory != NULL) {
		free(c->memory);
	}
	coroutines_free(&c->routines);
	struct thread_context clr = { 0 }; *c = clr;
}

struct buffer_descriptor alloc_buffer(size_t size) {
	void *buf = malloc_align(MIN_STORAGE_BUFFER_OFFSET_ALIGNMENT, size);
	struct buffer_descriptor desc = {
		.ptr = buf,
		.size_in_bytes = size,
		.robustness_size = size,
	};
	return desc;
}

struct image_descriptor alloc_image_rgba(int width, int height) {
	size_t size = width*height*4;
	size = (size + 16 - 1)&~(16 - 1);
	void *storage = malloc_align(REQUIRED_MEMORY_ALIGNMENT, size);
	struct image_descriptor desc = { 0 };
	desc.ptr = storage;
	desc.width = width;
	desc.height = height;
	desc.depth = 1;
	desc.row_pitch_bytes = width*4;
	desc.slice_pitch_bytes = size;
	desc.sample_pitch_bytes = size;
	desc.sample_count = 1;
	desc.size_in_bytes = size;
	return desc;
}

void prepare_dispatch(struct dispatch_context *ctx, int nthreads, struct program_info *info, uint8_t *desc_set, int ngroupx, int ngroupy, int ngroupz) {
	if (ctx->desc_set_size < info->desc_set_size) {
		if (ctx->descriptor_set != NULL) {
			free(ctx->descriptor_set);
		}
		ctx->descriptor_set = malloc_align(16, info->desc_set_size);
		ctx->desc_set_size = info->desc_set_size;
	}
	memcpy(ctx->descriptor_set, desc_set, info->desc_set_size);

	int invocations_per_subgroup = SIMD_WIDTH;
	int invocations_per_workgroup = info->workgroup_size_x * info->workgroup_size_y * info->workgroup_size_z;
	int subgroups_per_workgroup = (invocations_per_workgroup + invocations_per_subgroup - 1) / invocations_per_subgroup;

	ctx->has_cbarriers = info->has_cbarriers;
	ctx->begin = info->begin;
	ctx->await = info->await;
	ctx->destroy = info->destroy;
	ctx->nthreads = nthreads;
	ctx->memory_size = info->min_memory_size;

	ctx->data.workgroup_size[0] = info->workgroup_size_x;
	ctx->data.workgroup_size[1] = info->workgroup_size_y;
	ctx->data.workgroup_size[2] = info->workgroup_size_z;
	ctx->data.num_workgroups[0] = ngroupx;
	ctx->data.num_workgroups[1] = ngroupy;
	ctx->data.num_workgroups[2] = ngroupz;
	ctx->data.invocations_per_subgroup = invocations_per_subgroup;
	ctx->data.invocations_per_workgroup = invocations_per_workgroup;
	ctx->data.subgroups_per_workgroup = subgroups_per_workgroup;
	ctx->data.descriptor_sets[0] = ctx->descriptor_set;
}

void dispatch_thread(struct dispatch_context *ctx, int thread_idx, struct thread_context *thread) {
	if (thread->memory_size < ctx->memory_size) {
		if (thread->memory != NULL) {
			free(thread->memory);
		}
		// SwiftShader doesn't seem to align shared memory. However, better safe
		// than subtle errors. Note that the program info generator pads
		// memory_size to ensure space for alignment.
		thread->memory = malloc_align(16, ctx->memory_size);
		thread->memory_size = ctx->memory_size;
	}
	uint8_t *memory = thread->memory;

	struct program_data *data = &ctx->data;

	int sx = data->num_workgroups[0];
	int sy = data->num_workgroups[1];
	int sz = data->num_workgroups[2];
	int ngroups = sx * sy * sz;

	for (int i = thread_idx; i < ngroups; i += ctx->nthreads) {
		int group_id = i;
		int z = group_id / (sx * sy);
		group_id -= z * sx * sy;
		int y = group_id / sx;
		group_id -= y * sx;
		int x = group_id;
		if (ctx->has_cbarriers) {
			for (int subgroup = 0; subgroup < data->subgroups_per_workgroup; subgroup++) {
				coroutine r = ctx->begin(data, x, y, z, memory, subgroup, 1);
				coroutines_push(&thread->routines, r);
			}
		} else {
			coroutine r = ctx->begin(data, x, y, z, memory, 0, data->subgroups_per_workgroup);
			coroutines_push(&thread->routines, r);
		}
		coroutine r;
		while (coroutines_pop(&thread->routines, &r)) {
			yield_result res;
			if (ctx->await(r, &res)) {
				coroutines_push(&thread->routines, r);
			} else {
				ctx->destroy(r);
			}
		}
	}
}
