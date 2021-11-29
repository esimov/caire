#version 310 es

// SPDX-License-Identifier: Unlicense OR MIT

precision mediump float;

layout(location=0) in highp vec2 vFrom;
layout(location=1) in highp vec2 vCtrl;
layout(location=2) in highp vec2 vTo;

layout(location = 0) out vec4 fragCover;

void main() {
	float dx = vTo.x - vFrom.x;
	// Sort from and to in increasing order so the root below
	// is always the positive square root, if any.
	// We need the direction of the curve below, so this can't be
	// done from the vertex shader.
	bool increasing = vTo.x >= vFrom.x;
	vec2 left = increasing ? vFrom : vTo;
	vec2 right = increasing ? vTo : vFrom;

	// The signed horizontal extent of the fragment.
	vec2 extent = clamp(vec2(vFrom.x, vTo.x), -0.5, 0.5);
	// Find the t where the curve crosses the middle of the
	// extent, x₀.
	// Given the Bézier curve with x coordinates P₀, P₁, P₂
	// where P₀ is at the origin, its x coordinate in t
	// is given by:
	//
	// x(t) = 2(1-t)tP₁ + t²P₂
	// 
	// Rearranging:
	//
	// x(t) = (P₂ - 2P₁)t² + 2P₁t
	//
	// Setting x(t) = x₀ and using Muller's quadratic formula ("Citardauq")
	// for robustnesss,
	//
	// t = 2x₀/(2P₁±√(4P₁²+4(P₂-2P₁)x₀))
	//
	// which simplifies to
	//
	// t = x₀/(P₁±√(P₁²+(P₂-2P₁)x₀))
	//
	// Setting v = P₂-P₁,
	//
	// t = x₀/(P₁±√(P₁²+(v-P₁)x₀))
	//
	// t lie in [0; 1]; P₂ ≥ P₁ and P₁ ≥ 0 since we split curves where
	// the control point lies before the start point or after the end point.
	// It can then be shown that only the positive square root is valid.
	float midx = mix(extent.x, extent.y, 0.5);
	float x0 = midx - left.x;
	vec2 p1 = vCtrl - left;
	vec2 v = right - vCtrl;
	float t = x0/(p1.x+sqrt(p1.x*p1.x+(v.x-p1.x)*x0));
	// Find y(t) on the curve.
	float y = mix(mix(left.y, vCtrl.y, t), mix(vCtrl.y, right.y, t), t);
	// And the slope.
	vec2 d_half = mix(p1, v, t);
	float dy = d_half.y/d_half.x;
	// Together, y and dy form a line approximation.

	// Compute the fragment area above the line.
	// The area is symmetric around dy = 0. Scale slope with extent width.
	float width = extent.y - extent.x;
	dy = abs(dy*width);

	vec4 sides = vec4(dy*+0.5 + y, dy*-0.5 + y, (+0.5-y)/dy, (-0.5-y)/dy);
	sides = clamp(sides+0.5, 0.0, 1.0);

	float area = 0.5*(sides.z - sides.z*sides.y + 1.0 - sides.x+sides.x*sides.w);
	area *= width;

	// Work around issue #13.
	if (width == 0.0)
		area = 0.0;

	fragCover.r = area;
}
