#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

precision mediump float;

layout(binding = 0) uniform sampler2D tex;

layout(location = 0) in highp vec2 vUV;

layout(location = 0) out vec4 fragColor;

vec3 sRGBtoRGB(vec3 rgb) {
	bvec3 cutoff = greaterThanEqual(rgb, vec3(0.04045));
	vec3 below = rgb/vec3(12.92);
	vec3 above = pow((rgb + vec3(0.055))/vec3(1.055), vec3(2.4));
	return mix(below, above, cutoff);
}

void main() {
	vec4 texel = texture(tex, vUV);
	texel.rgb = sRGBtoRGB(texel.rgb);
	fragColor = texel;
}
