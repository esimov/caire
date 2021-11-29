#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

#extension GL_GOOGLE_include_directive : enable

precision highp float;

#include "common.h"

layout(location=0) in vec4 position;

void main() {
    gl_Position = vec4(transform3x2(windowTransform, position.xyz), position.w);
}
