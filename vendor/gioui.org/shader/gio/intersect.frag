#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

precision mediump float;

layout(location = 0) in highp vec2 vUV;

layout(binding = 0) uniform sampler2D cover;

layout(location = 0) out vec4 fragColor;

void main() {
  fragColor.r = abs(texture(cover, vUV).r);
}
