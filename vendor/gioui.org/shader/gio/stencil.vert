#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

#extension GL_GOOGLE_include_directive : enable

precision highp float;

#include "common.h"

layout(push_constant) uniform Block {
	vec4 transform;
	vec2 pathOffset;
} _block;

layout(location=0) in float corner;
layout(location=1) in float maxy;
layout(location=2) in vec2 from;
layout(location=3) in vec2 ctrl;
layout(location=4) in vec2 to;

layout(location=0) out vec2 vFrom;
layout(location=1) out vec2 vCtrl;
layout(location=2) out vec2 vTo;

void main() {
	// Add a one pixel overlap so curve quads cover their
	// entire curves. Could use conservative rasterization
	// if available.
	vec2 from = from + _block.pathOffset;
	vec2 ctrl = ctrl + _block.pathOffset;
	vec2 to = to + _block.pathOffset;
	float maxy = maxy + _block.pathOffset.y;
	vec2 pos;
	float c = corner;
	if (c >= 0.375) {
		// North.
		c -= 0.5;
		pos.y = maxy + 1.0;
	} else {
		// South.
		pos.y = min(min(from.y, ctrl.y), to.y) - 1.0;
	}
	if (c >= 0.125) {
		// East.
		pos.x = max(max(from.x, ctrl.x), to.x)+1.0;
	} else {
		// West.
		pos.x = min(min(from.x, ctrl.x), to.x)-1.0;
	}
	vFrom = from-pos;
	vCtrl = ctrl-pos;
	vTo = to-pos;
	pos = pos*_block.transform.xy + _block.transform.zw;
	gl_Position = vec4(transform3x2(fboTransform, vec3(pos, 0)), 1);
}

