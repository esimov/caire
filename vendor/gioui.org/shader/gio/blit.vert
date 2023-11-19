#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

#extension GL_GOOGLE_include_directive : enable

precision highp float;

#include "common.h"

layout(push_constant) uniform Block {
	vec4 transform;
	vec4 uvTransformR1;
	vec4 uvTransformR2;
	float opacity;
	// fbo is set if blitting to a FBO, otherwise the window.
	float fbo;
} _block;

layout(location = 0) in vec2 pos;

layout(location = 1) in vec2 uv;

layout(location = 0) out vec2 vUV;
layout(location = 1) out float opacity;

void main() {
	vec2 p = pos*_block.transform.xy + _block.transform.zw;
	if (_block.fbo != 0.0) {
		gl_Position = vec4(transform3x2(fboTransform, vec3(p, 0)), 1);
	} else {
		gl_Position = vec4(transform3x2(windowTransform, vec3(p, 0)), 1);
	}
	vUV = transform3x2(m3x2(_block.uvTransformR1.xyz, _block.uvTransformR2.xyz), vec3(uv,1)).xy;
	opacity = _block.opacity;
}
