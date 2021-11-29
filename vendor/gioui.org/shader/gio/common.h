// SPDX-License-Identifier: Unlicense OR MIT

struct m3x2 {
	vec3 r0;
	vec3 r1;
};

// fboTransform is the transformation that cancels the implied transformation
// between the clip space and the framebuffer. Only two rows are returned. The
// last is implied to be [0, 0, 1].
const m3x2 fboTransform = m3x2(
#if defined(LANG_HLSL) || defined(LANG_MSL) || defined(LANG_MSLIOS)
	vec3(1.0, 0.0, 0.0),
	vec3(0.0, -1.0, 0.0)
#else
	vec3(1.0, 0.0, 0.0),
	vec3(0.0, 1.0, 0.0)
#endif
);

// windowTransform is the transformation that cancels the implied transformation
// between framebuffer space and window system coordinates.
const m3x2 windowTransform = m3x2(
#if defined(LANG_VULKAN)
	vec3(1.0, 0.0, 0.0),
	vec3(0.0, 1.0, 0.0)
#else
	vec3(1.0, 0.0, 0.0),
	vec3(0.0, -1.0, 0.0)
#endif
);

vec3 transform3x2(m3x2 t, vec3 v) {
	return vec3(dot(t.r0, v), dot(t.r1, v), dot(vec3(0.0, 0.0, 1.0), v));
}
