#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

#extension GL_GOOGLE_include_directive : enable

precision highp float;

#include "common.h"

layout(push_constant) uniform Block {
	vec2 scale;
	vec2 pos;
	vec2 uvScale;
} _block;

layout(location = 0) in vec2 pos;
layout(location = 1) in vec2 uv;

layout(location = 0) out vec2 vUV;

void main() {
	vUV = vec2(uv*_block.uvScale);
	vec2 p = vec2(pos*_block.scale + _block.pos);
	gl_Position = vec4(transform3x2(windowTransform, vec3(p, 0)), 1);
}
