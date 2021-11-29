#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

precision mediump float;

layout(binding = 0) uniform sampler2D tex;

layout(location = 0) in highp vec2 vUV;

layout(location = 0) out vec4 fragColor;

layout(push_constant) uniform Color {
	// If emulateSRGB is set (!= 0), the input texels are sRGB encoded. We save the
	// conversion step below, at the cost of texture filtering in sRGB space.
	layout(offset=16) float emulateSRGB;
} _color;

vec3 RGBtosRGB(vec3 rgb) {
	bvec3 cutoff = greaterThanEqual(rgb, vec3(0.0031308));
	vec3 below = vec3(12.92)*rgb;
	vec3 above = vec3(1.055)*pow(rgb, vec3(0.41666)) - vec3(0.055);
	return mix(below, above, cutoff);
}

void main() {
	vec4 texel = texture(tex, vUV);
	if (_color.emulateSRGB == 0.0) {
		texel.rgb = RGBtosRGB(texel.rgb);
	}
	fragColor = texel;
}
