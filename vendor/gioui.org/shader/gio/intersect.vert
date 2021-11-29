#version 310 es
  
// SPDX-License-Identifier: Unlicense OR MIT

#extension GL_GOOGLE_include_directive : enable

precision highp float;

#include "common.h"

layout(location = 0) in vec2 pos;
layout(location = 1) in vec2 uv;

layout(push_constant) uniform Block {
	vec4 uvTransform;
	vec4 subUVTransform;
} _block;

layout(location = 0) out vec2 vUV;

void main() {
  vec3 p = transform3x2(fboTransform, vec3(pos, 1.0));
  gl_Position = vec4(p, 1);
  vUV = uv.xy*_block.subUVTransform.xy + _block.subUVTransform.zw;
  vUV = vUV*_block.uvTransform.xy + _block.uvTransform.zw;
}
