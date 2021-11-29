// SPDX-License-Identifier: Unlicense OR MIT

#define ALIGN(bytes, type) type __attribute__((aligned(bytes)))

typedef ALIGN(8, uint8_t) byte8[8];
typedef ALIGN(8, uint16_t) word4[4];
typedef ALIGN(4, uint32_t) dword;
typedef ALIGN(16, uint32_t) dword4[4];
typedef ALIGN(8, uint64_t) qword;
typedef ALIGN(16, uint64_t) qword2[2];
typedef ALIGN(16, unsigned int) uint4[4];
typedef ALIGN(8, uint32_t) dword2[2];
typedef ALIGN(8, unsigned short) ushort4[4];
typedef ALIGN(16, float) float4[4];
typedef ALIGN(16, int) int4[4];

typedef unsigned short half;

typedef unsigned char bool;

enum {
	MAX_BOUND_DESCRIPTOR_SETS = 4,
	MAX_DESCRIPTOR_SET_UNIFORM_BUFFERS_DYNAMIC = 8,
	MAX_DESCRIPTOR_SET_STORAGE_BUFFERS_DYNAMIC = 4,
	MAX_DESCRIPTOR_SET_COMBINED_BUFFERS_DYNAMIC =
		MAX_DESCRIPTOR_SET_UNIFORM_BUFFERS_DYNAMIC +
		MAX_DESCRIPTOR_SET_STORAGE_BUFFERS_DYNAMIC,
	MAX_PUSH_CONSTANT_SIZE = 128,

	MIN_STORAGE_BUFFER_OFFSET_ALIGNMENT = 256,

	REQUIRED_MEMORY_ALIGNMENT = 16,

	SIMD_WIDTH = 4,
};

struct image_descriptor {
	ALIGN(16, void *ptr);
	int width;
	int height;
	int depth;
	int row_pitch_bytes;
	int slice_pitch_bytes;
	int sample_pitch_bytes;
	int sample_count;
	int size_in_bytes;

	void *stencil_ptr;
	int stencil_row_pitch_bytes;
	int stencil_slice_pitch_bytes;
	int stencil_sample_pitch_bytes;

	// TODO: unused?
	void *memoryOwner;
};

struct buffer_descriptor {
	ALIGN(16, void *ptr);
	int size_in_bytes;
	int robustness_size;
};

struct program_data {
	uint8_t *descriptor_sets[MAX_BOUND_DESCRIPTOR_SETS];
	uint32_t descriptor_dynamic_offsets[MAX_DESCRIPTOR_SET_COMBINED_BUFFERS_DYNAMIC];
	uint4 num_workgroups;
	uint4 workgroup_size;
	uint32_t invocations_per_subgroup;
	uint32_t subgroups_per_workgroup;
	uint32_t invocations_per_workgroup;
	unsigned char push_constants[MAX_PUSH_CONSTANT_SIZE];
	// Unused.
	void *constants;
};

typedef int32_t yield_result;

typedef void * coroutine;

typedef coroutine (*routine_begin)(struct program_data *data,
	int32_t workgroupX,
	int32_t workgroupY,
	int32_t workgroupZ,
	void *workgroupMemory,
	int32_t firstSubgroup,
	int32_t subgroupCount);

typedef bool (*routine_await)(coroutine r, yield_result *res);

typedef void (*routine_destroy)(coroutine r);

