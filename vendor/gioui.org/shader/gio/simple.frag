#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

precision mediump float;

layout(location = 0) out vec4 fragColor;

void main() {
	fragColor = vec4(.25, .55, .75, 1.0);
}
