#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

precision mediump float;

layout(location=0) in highp vec2 vUV;
layout(location=1) in highp float opacity;

{{.Header}}

layout(location = 0) out vec4 fragColor;

void main() {
	fragColor = opacity*{{.FetchColorExpr}};
}
