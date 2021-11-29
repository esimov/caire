// SPDX-License-Identifier: Unlicense OR MIT

package opengl

import (
	"errors"
	"fmt"
	"image"
	"runtime"
	"strings"

	"gioui.org/internal/byteslice"
	"gioui.org/internal/gl"
)

// SRGBFBO implements an intermediate sRGB FBO
// for gamma-correct rendering on platforms without
// sRGB enabled native framebuffers.
type SRGBFBO struct {
	c        *gl.Functions
	state    *glState
	viewport image.Point
	fbo      gl.Framebuffer
	tex      gl.Texture
	blitted  bool
	quad     gl.Buffer
	prog     gl.Program
	format   textureTriple
}

func NewSRGBFBO(f *gl.Functions, state *glState) (*SRGBFBO, error) {
	glVer := f.GetString(gl.VERSION)
	ver, _, err := gl.ParseGLVersion(glVer)
	if err != nil {
		return nil, err
	}
	exts := strings.Split(f.GetString(gl.EXTENSIONS), " ")
	srgbTriple, err := srgbaTripleFor(ver, exts)
	if err != nil {
		// Fall back to the linear RGB colorspace, at the cost of color precision loss.
		srgbTriple = textureTriple{gl.RGBA, gl.Enum(gl.RGBA), gl.Enum(gl.UNSIGNED_BYTE)}
	}
	s := &SRGBFBO{
		c:      f,
		state:  state,
		format: srgbTriple,
		fbo:    f.CreateFramebuffer(),
		tex:    f.CreateTexture(),
	}
	state.bindTexture(f, 0, s.tex)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	return s, nil
}

func (s *SRGBFBO) Blit() {
	if !s.blitted {
		prog, err := gl.CreateProgram(s.c, blitVSrc, blitFSrc, []string{"pos", "uv"})
		if err != nil {
			panic(err)
		}
		s.prog = prog
		s.state.useProgram(s.c, prog)
		s.c.Uniform1i(s.c.GetUniformLocation(prog, "tex"), 0)
		s.quad = s.c.CreateBuffer()
		s.state.bindBuffer(s.c, gl.ARRAY_BUFFER, s.quad)
		coords := byteslice.Slice([]float32{
			-1, +1, 0, 1,
			+1, +1, 1, 1,
			-1, -1, 0, 0,
			+1, -1, 1, 0,
		})
		s.c.BufferData(gl.ARRAY_BUFFER, len(coords), gl.STATIC_DRAW, coords)
		s.blitted = true
	}
	s.state.useProgram(s.c, s.prog)
	s.state.bindTexture(s.c, 0, s.tex)
	s.state.vertexAttribPointer(s.c, s.quad, 0 /* pos */, 2, gl.FLOAT, false, 4*4, 0)
	s.state.vertexAttribPointer(s.c, s.quad, 1 /* uv */, 2, gl.FLOAT, false, 4*4, 4*2)
	s.state.setVertexAttribArray(s.c, 0, true)
	s.state.setVertexAttribArray(s.c, 1, true)
	s.c.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	s.state.bindFramebuffer(s.c, gl.FRAMEBUFFER, s.fbo)
	s.c.InvalidateFramebuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0)
}

func (s *SRGBFBO) Framebuffer() gl.Framebuffer {
	return s.fbo
}

func (s *SRGBFBO) Refresh(viewport image.Point) error {
	if viewport.X == 0 || viewport.Y == 0 {
		return errors.New("srgb: zero-sized framebuffer")
	}
	if s.viewport == viewport {
		return nil
	}
	s.viewport = viewport
	s.state.bindTexture(s.c, 0, s.tex)
	s.c.TexImage2D(gl.TEXTURE_2D, 0, s.format.internalFormat, viewport.X, viewport.Y, s.format.format, s.format.typ)
	s.state.bindFramebuffer(s.c, gl.FRAMEBUFFER, s.fbo)
	s.c.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.tex, 0)
	if st := s.c.CheckFramebufferStatus(gl.FRAMEBUFFER); st != gl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("sRGB framebuffer incomplete (%dx%d), status: %#x error: %x", viewport.X, viewport.Y, st, s.c.GetError())
	}

	if runtime.GOOS == "js" {
		// With macOS Safari, rendering to and then reading from a SRGB8_ALPHA8
		// texture result in twice gamma corrected colors. Using a plain RGBA
		// texture seems to work.
		s.state.setClearColor(s.c, .5, .5, .5, 1.0)
		s.c.Clear(gl.COLOR_BUFFER_BIT)
		var pixel [4]byte
		s.c.ReadPixels(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, pixel[:])
		if pixel[0] == 128 { // Correct sRGB color value is ~188
			s.c.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, viewport.X, viewport.Y, gl.RGBA, gl.UNSIGNED_BYTE)
			if st := s.c.CheckFramebufferStatus(gl.FRAMEBUFFER); st != gl.FRAMEBUFFER_COMPLETE {
				return fmt.Errorf("fallback RGBA framebuffer incomplete (%dx%d), status: %#x error: %x", viewport.X, viewport.Y, st, s.c.GetError())
			}
		}
	}

	return nil
}

func (s *SRGBFBO) Release() {
	s.state.deleteFramebuffer(s.c, s.fbo)
	s.state.deleteTexture(s.c, s.tex)
	if s.blitted {
		s.state.deleteBuffer(s.c, s.quad)
		s.state.deleteProgram(s.c, s.prog)
	}
	s.c = nil
}

const (
	blitVSrc = `
#version 100

precision highp float;

attribute vec2 pos;
attribute vec2 uv;

varying vec2 vUV;

void main() {
    gl_Position = vec4(pos, 0, 1);
    vUV = uv;
}
`
	blitFSrc = `
#version 100

precision mediump float;

uniform sampler2D tex;
varying vec2 vUV;

vec3 gamma(vec3 rgb) {
	vec3 exp = vec3(1.055)*pow(rgb, vec3(0.41666)) - vec3(0.055);
	vec3 lin = rgb * vec3(12.92);
	bvec3 cut = lessThan(rgb, vec3(0.0031308));
	return vec3(cut.r ? lin.r : exp.r, cut.g ? lin.g : exp.g, cut.b ? lin.b : exp.b);
}

void main() {
    vec4 col = texture2D(tex, vUV);
	vec3 rgb = col.rgb;
	rgb = gamma(rgb);
	gl_FragColor = vec4(rgb, col.a);
}
`
)
