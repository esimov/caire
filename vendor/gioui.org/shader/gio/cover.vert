#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

#extension GL_GOOGLE_include_directive : enable

precision highp float;

#include "common.h"

layout(push_constant) uniform Block {
	vec4 transform;
	vec4 uvCoverTransform;
	vec4 uvTransformR1;
	vec4 uvTransformR2;
} _block;

layout(location = 0) in vec2 pos;

layout(location = 0) out vec2 vCoverUV;

layout(location = 1) in vec2 uv;
layout(location = 1) out vec2 vUV;

void main() {
	vec2 p = vec2(pos*_block.transform.xy + _block.transform.zw);
    gl_Position = vec4(transform3x2(windowTransform, vec3(p, 0)), 1);
	vUV = transform3x2(m3x2(_block.uvTransformR1.xyz, _block.uvTransformR2.xyz), vec3(uv,1)).xy;
	vec3 uv3 = vec3(uv, 1.0);
	vCoverUV = (uv3*vec3(_block.uvCoverTransform.xy, 1.0)+vec3(_block.uvCoverTransform.zw, 0.0)).xy;
}
