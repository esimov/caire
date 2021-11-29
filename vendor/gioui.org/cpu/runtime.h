// SPDX-License-Identifier: Unlicense OR MIT

#define ATTR_HIDDEN __attribute__ ((visibility ("hidden")))

// program_info contains constant parameters for a program.
struct program_info {
	// MinMemorySize is the minimum size of memory passed to dispatch.
	size_t min_memory_size;
	// has_cbarriers is 1 when the program contains control barriers.
	bool has_cbarriers;
	// desc_set_size is the size of the first descriptor set for the program.
	size_t desc_set_size;
	int workgroup_size_x;
	int workgroup_size_y;
	int workgroup_size_z;
	// Program entrypoints.
	routine_begin begin;
	routine_await await;
	routine_destroy destroy;
};

// dispatch_context contains the information a program dispatch.
struct dispatch_context;

// thread_context contains the working memory of a batch. It may be
// reused, but not concurrently.
struct thread_context;

extern struct buffer_descriptor alloc_buffer(size_t size) ATTR_HIDDEN;
extern struct image_descriptor alloc_image_rgba(int width, int height) ATTR_HIDDEN;

extern struct dispatch_context *alloc_dispatch_context(void) ATTR_HIDDEN;

extern void free_dispatch_context(struct dispatch_context *c) ATTR_HIDDEN;

extern struct thread_context *alloc_thread_context(void) ATTR_HIDDEN;

extern void free_thread_context(struct thread_context *c) ATTR_HIDDEN;

// prepare_dispatch initializes ctx to run a dispatch of a program distributed
// among nthreads threads.
extern void prepare_dispatch(struct dispatch_context *ctx, int nthreads, struct program_info *info, uint8_t *desc_set, int ngroupx, int ngroupy, int ngroupz) ATTR_HIDDEN;

// dispatch_batch executes a dispatch batch.
extern void dispatch_thread(struct dispatch_context *ctx, int thread_idx, struct thread_context *thread) ATTR_HIDDEN;
