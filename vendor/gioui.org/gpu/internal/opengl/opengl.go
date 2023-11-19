// SPDX-License-Identifier: Unlicense OR MIT

package opengl

import (
	"errors"
	"fmt"
	"image"
	"math/bits"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/gl"
	"gioui.org/shader"
)

// Backend implements driver.Device.
type Backend struct {
	funcs *gl.Functions

	clear      bool
	glstate    glState
	state      state
	savedState glState
	sharedCtx  bool

	glver [2]int
	gles  bool
	feats driver.Caps
	// floatTriple holds the settings for floating point
	// textures.
	floatTriple textureTriple
	// Single channel alpha textures.
	alphaTriple textureTriple
	srgbaTriple textureTriple
	storage     [storageBindings]*buffer

	outputFBO gl.Framebuffer
	sRGBFBO   *SRGBFBO

	// vertArray is bound during a frame. We don't need it, but
	// core desktop OpenGL profile 3.3 requires some array bound.
	vertArray gl.VertexArray
}

// State tracking.
type glState struct {
	drawFBO     gl.Framebuffer
	readFBO     gl.Framebuffer
	vertAttribs [5]struct {
		obj        gl.Buffer
		enabled    bool
		size       int
		typ        gl.Enum
		normalized bool
		stride     int
		offset     uintptr
	}
	prog     gl.Program
	texUnits struct {
		active gl.Enum
		binds  [2]gl.Texture
	}
	arrayBuf  gl.Buffer
	elemBuf   gl.Buffer
	uniBuf    gl.Buffer
	uniBufs   [2]gl.Buffer
	storeBuf  gl.Buffer
	storeBufs [4]gl.Buffer
	vertArray gl.VertexArray
	srgb      bool
	blend     struct {
		enable         bool
		srcRGB, dstRGB gl.Enum
		srcA, dstA     gl.Enum
	}
	clearColor        [4]float32
	viewport          [4]int
	unpack_row_length int
	pack_row_length   int
}

type state struct {
	pipeline *pipeline
	buffer   bufferBinding
}

type bufferBinding struct {
	obj    gl.Buffer
	offset int
}

type timer struct {
	funcs *gl.Functions
	obj   gl.Query
}

type texture struct {
	backend  *Backend
	obj      gl.Texture
	fbo      gl.Framebuffer
	hasFBO   bool
	triple   textureTriple
	width    int
	height   int
	mipmap   bool
	bindings driver.BufferBinding
	foreign  bool
}

type pipeline struct {
	prog     *program
	inputs   []shader.InputLocation
	layout   driver.VertexLayout
	blend    driver.BlendDesc
	topology driver.Topology
}

type buffer struct {
	backend   *Backend
	hasBuffer bool
	obj       gl.Buffer
	typ       driver.BufferBinding
	size      int
	immutable bool
	// For emulation of uniform buffers.
	data []byte
}

type glshader struct {
	backend *Backend
	obj     gl.Shader
	src     shader.Sources
}

type program struct {
	backend      *Backend
	obj          gl.Program
	vertUniforms uniforms
	fragUniforms uniforms
}

type uniforms struct {
	locs []uniformLocation
	size int
}

type uniformLocation struct {
	uniform gl.Uniform
	offset  int
	typ     shader.DataType
	size    int
}

// textureTriple holds the type settings for
// a TexImage2D call.
type textureTriple struct {
	internalFormat gl.Enum
	format         gl.Enum
	typ            gl.Enum
}

const (
	storageBindings = 32
)

func init() {
	driver.NewOpenGLDevice = newOpenGLDevice
}

// Supporting compute programs is theoretically possible with OpenGL ES 3.1. In
// practice, there are too many driver issues, especially on Android (e.g.
// Google Pixel, Samsung J2 are both broken i different ways). Disable support
// and rely on Vulkan for devices that support it, and the CPU fallback for
// devices that don't.
const brokenGLES31 = true

func newOpenGLDevice(api driver.OpenGL) (driver.Device, error) {
	f, err := gl.NewFunctions(api.Context, api.ES)
	if err != nil {
		return nil, err
	}
	exts := strings.Split(f.GetString(gl.EXTENSIONS), " ")
	glVer := f.GetString(gl.VERSION)
	ver, gles, err := gl.ParseGLVersion(glVer)
	if err != nil {
		return nil, err
	}
	floatTriple, ffboErr := floatTripleFor(f, ver, exts)
	srgbaTriple, srgbErr := srgbaTripleFor(ver, exts)
	gles31 := gles && (ver[0] > 3 || (ver[0] == 3 && ver[1] >= 1))
	b := &Backend{
		glver:       ver,
		gles:        gles,
		funcs:       f,
		floatTriple: floatTriple,
		alphaTriple: alphaTripleFor(ver),
		srgbaTriple: srgbaTriple,
		sharedCtx:   api.Shared,
	}
	b.feats.BottomLeftOrigin = true
	if srgbErr == nil {
		b.feats.Features |= driver.FeatureSRGB
	}
	if ffboErr == nil {
		b.feats.Features |= driver.FeatureFloatRenderTargets
	}
	if gles31 && !brokenGLES31 {
		b.feats.Features |= driver.FeatureCompute
	}
	if hasExtension(exts, "GL_EXT_disjoint_timer_query_webgl2") || hasExtension(exts, "GL_EXT_disjoint_timer_query") {
		b.feats.Features |= driver.FeatureTimers
	}
	b.feats.MaxTextureSize = f.GetInteger(gl.MAX_TEXTURE_SIZE)
	if !b.sharedCtx {
		// We have exclusive access to the context, so query the GL state once
		// instead of at each frame.
		b.glstate = b.queryState()
	}
	return b, nil
}

func (b *Backend) BeginFrame(target driver.RenderTarget, clear bool, viewport image.Point) driver.Texture {
	b.clear = clear
	if b.sharedCtx {
		b.glstate = b.queryState()
		b.savedState = b.glstate
	}
	b.state = state{}
	var renderFBO gl.Framebuffer
	if target != nil {
		switch t := target.(type) {
		case driver.OpenGLRenderTarget:
			renderFBO = gl.Framebuffer(t)
		case *texture:
			renderFBO = t.ensureFBO()
		default:
			panic(fmt.Errorf("opengl: invalid render target type: %T", target))
		}
	}
	b.outputFBO = renderFBO
	b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, renderFBO)
	if b.gles {
		// If the output framebuffer is not in the sRGB colorspace already, emulate it.
		fbSRGB := false
		if !b.gles || b.glver[0] > 2 {
			var fbEncoding int
			if !renderFBO.Valid() {
				fbEncoding = b.funcs.GetFramebufferAttachmentParameteri(gl.FRAMEBUFFER, gl.BACK, gl.FRAMEBUFFER_ATTACHMENT_COLOR_ENCODING)
			} else {
				fbEncoding = b.funcs.GetFramebufferAttachmentParameteri(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.FRAMEBUFFER_ATTACHMENT_COLOR_ENCODING)
			}
			fbSRGB = fbEncoding != gl.LINEAR
		}
		if !fbSRGB && viewport != (image.Point{}) {
			if b.sRGBFBO == nil {
				sfbo, err := NewSRGBFBO(b.funcs, &b.glstate)
				if err != nil {
					panic(err)
				}
				b.sRGBFBO = sfbo
			}
			if err := b.sRGBFBO.Refresh(viewport); err != nil {
				panic(err)
			}
			renderFBO = b.sRGBFBO.Framebuffer()
		} else if b.sRGBFBO != nil {
			b.sRGBFBO.Release()
			b.sRGBFBO = nil
		}
	} else {
		b.glstate.set(b.funcs, gl.FRAMEBUFFER_SRGB, true)
		if !b.vertArray.Valid() {
			b.vertArray = b.funcs.CreateVertexArray()
		}
		b.glstate.bindVertexArray(b.funcs, b.vertArray)
	}
	b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, renderFBO)
	if b.sRGBFBO != nil && !clear {
		b.clearOutput(0, 0, 0, 0)
	}
	return &texture{backend: b, fbo: renderFBO, hasFBO: true, foreign: true}
}

func (b *Backend) EndFrame() {
	if b.sRGBFBO != nil {
		b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, b.outputFBO)
		if b.clear {
			b.SetBlend(false)
		} else {
			b.BlendFunc(driver.BlendFactorOne, driver.BlendFactorOneMinusSrcAlpha)
			b.SetBlend(true)
		}
		b.sRGBFBO.Blit()
	}
	if b.sharedCtx {
		b.restoreState(b.savedState)
	} else if runtime.GOOS == "android" {
		// The Android emulator needs the output framebuffer to be current when
		// eglSwapBuffers is called.
		b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, b.outputFBO)
	}
}

func (b *Backend) queryState() glState {
	s := glState{
		prog:       gl.Program(b.funcs.GetBinding(gl.CURRENT_PROGRAM)),
		arrayBuf:   gl.Buffer(b.funcs.GetBinding(gl.ARRAY_BUFFER_BINDING)),
		elemBuf:    gl.Buffer(b.funcs.GetBinding(gl.ELEMENT_ARRAY_BUFFER_BINDING)),
		drawFBO:    gl.Framebuffer(b.funcs.GetBinding(gl.FRAMEBUFFER_BINDING)),
		clearColor: b.funcs.GetFloat4(gl.COLOR_CLEAR_VALUE),
		viewport:   b.funcs.GetInteger4(gl.VIEWPORT),
	}
	if !b.gles || b.glver[0] > 2 {
		s.unpack_row_length = b.funcs.GetInteger(gl.UNPACK_ROW_LENGTH)
		s.pack_row_length = b.funcs.GetInteger(gl.PACK_ROW_LENGTH)
	}
	s.blend.enable = b.funcs.IsEnabled(gl.BLEND)
	s.blend.srcRGB = gl.Enum(b.funcs.GetInteger(gl.BLEND_SRC_RGB))
	s.blend.dstRGB = gl.Enum(b.funcs.GetInteger(gl.BLEND_DST_RGB))
	s.blend.srcA = gl.Enum(b.funcs.GetInteger(gl.BLEND_SRC_ALPHA))
	s.blend.dstA = gl.Enum(b.funcs.GetInteger(gl.BLEND_DST_ALPHA))
	s.texUnits.active = gl.Enum(b.funcs.GetInteger(gl.ACTIVE_TEXTURE))
	if !b.gles {
		s.srgb = b.funcs.IsEnabled(gl.FRAMEBUFFER_SRGB)
	}
	if !b.gles || b.glver[0] >= 3 {
		s.vertArray = gl.VertexArray(b.funcs.GetBinding(gl.VERTEX_ARRAY_BINDING))
		s.readFBO = gl.Framebuffer(b.funcs.GetBinding(gl.READ_FRAMEBUFFER_BINDING))
		s.uniBuf = gl.Buffer(b.funcs.GetBinding(gl.UNIFORM_BUFFER_BINDING))
		for i := range s.uniBufs {
			s.uniBufs[i] = gl.Buffer(b.funcs.GetBindingi(gl.UNIFORM_BUFFER_BINDING, i))
		}
	}
	if b.gles && (b.glver[0] > 3 || (b.glver[0] == 3 && b.glver[1] >= 1)) {
		s.storeBuf = gl.Buffer(b.funcs.GetBinding(gl.SHADER_STORAGE_BUFFER_BINDING))
		for i := range s.storeBufs {
			s.storeBufs[i] = gl.Buffer(b.funcs.GetBindingi(gl.SHADER_STORAGE_BUFFER_BINDING, i))
		}
	}
	active := s.texUnits.active
	for i := range s.texUnits.binds {
		s.activeTexture(b.funcs, gl.TEXTURE0+gl.Enum(i))
		s.texUnits.binds[i] = gl.Texture(b.funcs.GetBinding(gl.TEXTURE_BINDING_2D))
	}
	s.activeTexture(b.funcs, active)
	for i := range s.vertAttribs {
		a := &s.vertAttribs[i]
		a.enabled = b.funcs.GetVertexAttrib(i, gl.VERTEX_ATTRIB_ARRAY_ENABLED) != gl.FALSE
		a.obj = gl.Buffer(b.funcs.GetVertexAttribBinding(i, gl.VERTEX_ATTRIB_ARRAY_ENABLED))
		a.size = b.funcs.GetVertexAttrib(i, gl.VERTEX_ATTRIB_ARRAY_SIZE)
		a.typ = gl.Enum(b.funcs.GetVertexAttrib(i, gl.VERTEX_ATTRIB_ARRAY_TYPE))
		a.normalized = b.funcs.GetVertexAttrib(i, gl.VERTEX_ATTRIB_ARRAY_NORMALIZED) != gl.FALSE
		a.stride = b.funcs.GetVertexAttrib(i, gl.VERTEX_ATTRIB_ARRAY_STRIDE)
		a.offset = b.funcs.GetVertexAttribPointer(i, gl.VERTEX_ATTRIB_ARRAY_POINTER)
	}
	return s
}

func (b *Backend) restoreState(dst glState) {
	src := &b.glstate
	f := b.funcs
	for i, unit := range dst.texUnits.binds {
		src.bindTexture(f, i, unit)
	}
	src.activeTexture(f, dst.texUnits.active)
	src.bindFramebuffer(f, gl.FRAMEBUFFER, dst.drawFBO)
	src.bindFramebuffer(f, gl.READ_FRAMEBUFFER, dst.readFBO)
	src.set(f, gl.BLEND, dst.blend.enable)
	bf := dst.blend
	src.setBlendFuncSeparate(f, bf.srcRGB, bf.dstRGB, bf.srcA, bf.dstA)
	src.set(f, gl.FRAMEBUFFER_SRGB, dst.srgb)
	src.bindVertexArray(f, dst.vertArray)
	src.useProgram(f, dst.prog)
	src.bindBuffer(f, gl.ELEMENT_ARRAY_BUFFER, dst.elemBuf)
	for i, b := range dst.uniBufs {
		src.bindBufferBase(f, gl.UNIFORM_BUFFER, i, b)
	}
	src.bindBuffer(f, gl.UNIFORM_BUFFER, dst.uniBuf)
	for i, b := range dst.storeBufs {
		src.bindBufferBase(f, gl.SHADER_STORAGE_BUFFER, i, b)
	}
	src.bindBuffer(f, gl.SHADER_STORAGE_BUFFER, dst.storeBuf)
	col := dst.clearColor
	src.setClearColor(f, col[0], col[1], col[2], col[3])
	for i, attr := range dst.vertAttribs {
		src.setVertexAttribArray(f, i, attr.enabled)
		src.vertexAttribPointer(f, attr.obj, i, attr.size, attr.typ, attr.normalized, attr.stride, int(attr.offset))
	}
	src.bindBuffer(f, gl.ARRAY_BUFFER, dst.arrayBuf)
	v := dst.viewport
	src.setViewport(f, v[0], v[1], v[2], v[3])
	src.pixelStorei(f, gl.UNPACK_ROW_LENGTH, dst.unpack_row_length)
	src.pixelStorei(f, gl.PACK_ROW_LENGTH, dst.pack_row_length)
}

func (s *glState) setVertexAttribArray(f *gl.Functions, idx int, enabled bool) {
	a := &s.vertAttribs[idx]
	if enabled != a.enabled {
		if enabled {
			f.EnableVertexAttribArray(gl.Attrib(idx))
		} else {
			f.DisableVertexAttribArray(gl.Attrib(idx))
		}
		a.enabled = enabled
	}
}

func (s *glState) vertexAttribPointer(f *gl.Functions, buf gl.Buffer, idx, size int, typ gl.Enum, normalized bool, stride, offset int) {
	s.bindBuffer(f, gl.ARRAY_BUFFER, buf)
	a := &s.vertAttribs[idx]
	a.obj = buf
	a.size = size
	a.typ = typ
	a.normalized = normalized
	a.stride = stride
	a.offset = uintptr(offset)
	f.VertexAttribPointer(gl.Attrib(idx), a.size, a.typ, a.normalized, a.stride, int(a.offset))
}

func (s *glState) activeTexture(f *gl.Functions, unit gl.Enum) {
	if unit != s.texUnits.active {
		f.ActiveTexture(unit)
		s.texUnits.active = unit
	}
}

func (s *glState) bindTexture(f *gl.Functions, unit int, t gl.Texture) {
	s.activeTexture(f, gl.TEXTURE0+gl.Enum(unit))
	if !t.Equal(s.texUnits.binds[unit]) {
		f.BindTexture(gl.TEXTURE_2D, t)
		s.texUnits.binds[unit] = t
	}
}

func (s *glState) bindVertexArray(f *gl.Functions, a gl.VertexArray) {
	if !a.Equal(s.vertArray) {
		f.BindVertexArray(a)
		s.vertArray = a
	}
}

func (s *glState) deleteFramebuffer(f *gl.Functions, fbo gl.Framebuffer) {
	f.DeleteFramebuffer(fbo)
	if fbo.Equal(s.drawFBO) {
		s.drawFBO = gl.Framebuffer{}
	}
	if fbo.Equal(s.readFBO) {
		s.readFBO = gl.Framebuffer{}
	}
}

func (s *glState) deleteBuffer(f *gl.Functions, b gl.Buffer) {
	f.DeleteBuffer(b)
	if b.Equal(s.arrayBuf) {
		s.arrayBuf = gl.Buffer{}
	}
	if b.Equal(s.elemBuf) {
		s.elemBuf = gl.Buffer{}
	}
	if b.Equal(s.uniBuf) {
		s.uniBuf = gl.Buffer{}
	}
	if b.Equal(s.storeBuf) {
		s.uniBuf = gl.Buffer{}
	}
	for i, b2 := range s.storeBufs {
		if b.Equal(b2) {
			s.storeBufs[i] = gl.Buffer{}
		}
	}
	for i, b2 := range s.uniBufs {
		if b.Equal(b2) {
			s.uniBufs[i] = gl.Buffer{}
		}
	}
}

func (s *glState) deleteProgram(f *gl.Functions, p gl.Program) {
	f.DeleteProgram(p)
	if p.Equal(s.prog) {
		s.prog = gl.Program{}
	}
}

func (s *glState) deleteVertexArray(f *gl.Functions, a gl.VertexArray) {
	f.DeleteVertexArray(a)
	if a.Equal(s.vertArray) {
		s.vertArray = gl.VertexArray{}
	}
}

func (s *glState) deleteTexture(f *gl.Functions, t gl.Texture) {
	f.DeleteTexture(t)
	binds := &s.texUnits.binds
	for i, obj := range binds {
		if t.Equal(obj) {
			binds[i] = gl.Texture{}
		}
	}
}

func (s *glState) useProgram(f *gl.Functions, p gl.Program) {
	if !p.Equal(s.prog) {
		f.UseProgram(p)
		s.prog = p
	}
}

func (s *glState) bindFramebuffer(f *gl.Functions, target gl.Enum, fbo gl.Framebuffer) {
	switch target {
	case gl.FRAMEBUFFER:
		if fbo.Equal(s.drawFBO) && fbo.Equal(s.readFBO) {
			return
		}
		s.drawFBO = fbo
		s.readFBO = fbo
	case gl.READ_FRAMEBUFFER:
		if fbo.Equal(s.readFBO) {
			return
		}
		s.readFBO = fbo
	case gl.DRAW_FRAMEBUFFER:
		if fbo.Equal(s.drawFBO) {
			return
		}
		s.drawFBO = fbo
	default:
		panic("unknown target")
	}
	f.BindFramebuffer(target, fbo)
}

func (s *glState) bindBufferBase(f *gl.Functions, target gl.Enum, idx int, buf gl.Buffer) {
	switch target {
	case gl.UNIFORM_BUFFER:
		if buf.Equal(s.uniBuf) && buf.Equal(s.uniBufs[idx]) {
			return
		}
		s.uniBuf = buf
		s.uniBufs[idx] = buf
	case gl.SHADER_STORAGE_BUFFER:
		if buf.Equal(s.storeBuf) && buf.Equal(s.storeBufs[idx]) {
			return
		}
		s.storeBuf = buf
		s.storeBufs[idx] = buf
	default:
		panic("unknown buffer target")
	}
	f.BindBufferBase(target, idx, buf)
}

func (s *glState) bindBuffer(f *gl.Functions, target gl.Enum, buf gl.Buffer) {
	switch target {
	case gl.ARRAY_BUFFER:
		if buf.Equal(s.arrayBuf) {
			return
		}
		s.arrayBuf = buf
	case gl.ELEMENT_ARRAY_BUFFER:
		if buf.Equal(s.elemBuf) {
			return
		}
		s.elemBuf = buf
	case gl.UNIFORM_BUFFER:
		if buf.Equal(s.uniBuf) {
			return
		}
		s.uniBuf = buf
	case gl.SHADER_STORAGE_BUFFER:
		if buf.Equal(s.storeBuf) {
			return
		}
		s.storeBuf = buf
	default:
		panic("unknown buffer target")
	}
	f.BindBuffer(target, buf)
}

func (s *glState) pixelStorei(f *gl.Functions, pname gl.Enum, val int) {
	switch pname {
	case gl.UNPACK_ROW_LENGTH:
		if val == s.unpack_row_length {
			return
		}
		s.unpack_row_length = val
	case gl.PACK_ROW_LENGTH:
		if val == s.pack_row_length {
			return
		}
		s.pack_row_length = val
	default:
		panic("unsupported PixelStorei pname")
	}
	f.PixelStorei(pname, val)
}

func (s *glState) setClearColor(f *gl.Functions, r, g, b, a float32) {
	col := [4]float32{r, g, b, a}
	if col != s.clearColor {
		f.ClearColor(r, g, b, a)
		s.clearColor = col
	}
}

func (s *glState) setViewport(f *gl.Functions, x, y, width, height int) {
	view := [4]int{x, y, width, height}
	if view != s.viewport {
		f.Viewport(x, y, width, height)
		s.viewport = view
	}
}

func (s *glState) setBlendFuncSeparate(f *gl.Functions, srcRGB, dstRGB, srcA, dstA gl.Enum) {
	if srcRGB != s.blend.srcRGB || dstRGB != s.blend.dstRGB || srcA != s.blend.srcA || dstA != s.blend.dstA {
		s.blend.srcRGB = srcRGB
		s.blend.dstRGB = dstRGB
		s.blend.srcA = srcA
		s.blend.dstA = dstA
		f.BlendFuncSeparate(srcA, dstA, srcA, dstA)
	}
}

func (s *glState) set(f *gl.Functions, target gl.Enum, enable bool) {
	switch target {
	case gl.FRAMEBUFFER_SRGB:
		if s.srgb == enable {
			return
		}
		s.srgb = enable
	case gl.BLEND:
		if enable == s.blend.enable {
			return
		}
		s.blend.enable = enable
	default:
		panic("unknown enable")
	}
	if enable {
		f.Enable(target)
	} else {
		f.Disable(target)
	}
}

func (b *Backend) Caps() driver.Caps {
	return b.feats
}

func (b *Backend) NewTimer() driver.Timer {
	return &timer{
		funcs: b.funcs,
		obj:   b.funcs.CreateQuery(),
	}
}

func (b *Backend) IsTimeContinuous() bool {
	return b.funcs.GetInteger(gl.GPU_DISJOINT_EXT) == gl.FALSE
}

func (t *texture) ensureFBO() gl.Framebuffer {
	if t.hasFBO {
		return t.fbo
	}
	b := t.backend
	oldFBO := b.glstate.drawFBO
	defer func() {
		b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, oldFBO)
	}()
	glErr(b.funcs)
	fb := b.funcs.CreateFramebuffer()
	b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, fb)
	if err := glErr(b.funcs); err != nil {
		b.funcs.DeleteFramebuffer(fb)
		panic(err)
	}
	b.funcs.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, t.obj, 0)
	if st := b.funcs.CheckFramebufferStatus(gl.FRAMEBUFFER); st != gl.FRAMEBUFFER_COMPLETE {
		b.funcs.DeleteFramebuffer(fb)
		panic(fmt.Errorf("incomplete framebuffer, status = 0x%x, err = %d", st, b.funcs.GetError()))
	}
	t.fbo = fb
	t.hasFBO = true
	return fb
}

func (b *Backend) NewTexture(format driver.TextureFormat, width, height int, minFilter, magFilter driver.TextureFilter, binding driver.BufferBinding) (driver.Texture, error) {
	glErr(b.funcs)
	tex := &texture{backend: b, obj: b.funcs.CreateTexture(), width: width, height: height, bindings: binding}
	switch format {
	case driver.TextureFormatFloat:
		tex.triple = b.floatTriple
	case driver.TextureFormatSRGBA:
		tex.triple = b.srgbaTriple
	case driver.TextureFormatRGBA8:
		tex.triple = textureTriple{gl.RGBA8, gl.RGBA, gl.UNSIGNED_BYTE}
	default:
		return nil, errors.New("unsupported texture format")
	}
	b.BindTexture(0, tex)
	min, mipmap := toTexFilter(minFilter)
	mag, _ := toTexFilter(magFilter)
	if b.gles && b.glver[0] < 3 {
		// OpenGL ES 2 only supports mipmaps for power-of-two textures.
		mipmap = false
	}
	tex.mipmap = mipmap
	b.funcs.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, mag)
	b.funcs.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, min)
	b.funcs.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.funcs.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	if mipmap {
		nmipmaps := 1
		if mipmap {
			dim := width
			if height > dim {
				dim = height
			}
			log2 := 32 - bits.LeadingZeros32(uint32(dim)) - 1
			nmipmaps = log2 + 1
		}
		// Immutable textures are required for BindImageTexture, and can't hurt otherwise.
		b.funcs.TexStorage2D(gl.TEXTURE_2D, nmipmaps, tex.triple.internalFormat, width, height)
	} else {
		b.funcs.TexImage2D(gl.TEXTURE_2D, 0, tex.triple.internalFormat, width, height, tex.triple.format, tex.triple.typ)
	}
	if err := glErr(b.funcs); err != nil {
		tex.Release()
		return nil, err
	}
	return tex, nil
}

func (b *Backend) NewBuffer(typ driver.BufferBinding, size int) (driver.Buffer, error) {
	glErr(b.funcs)
	buf := &buffer{backend: b, typ: typ, size: size}
	if typ&driver.BufferBindingUniforms != 0 {
		if typ != driver.BufferBindingUniforms {
			return nil, errors.New("uniforms buffers cannot be bound as anything else")
		}
		buf.data = make([]byte, size)
	}
	if typ&^driver.BufferBindingUniforms != 0 {
		buf.hasBuffer = true
		buf.obj = b.funcs.CreateBuffer()
		if err := glErr(b.funcs); err != nil {
			buf.Release()
			return nil, err
		}
		firstBinding := firstBufferType(typ)
		b.glstate.bindBuffer(b.funcs, firstBinding, buf.obj)
		b.funcs.BufferData(firstBinding, size, gl.DYNAMIC_DRAW, nil)
	}
	return buf, nil
}

func (b *Backend) NewImmutableBuffer(typ driver.BufferBinding, data []byte) (driver.Buffer, error) {
	glErr(b.funcs)
	obj := b.funcs.CreateBuffer()
	buf := &buffer{backend: b, obj: obj, typ: typ, size: len(data), hasBuffer: true}
	firstBinding := firstBufferType(typ)
	b.glstate.bindBuffer(b.funcs, firstBinding, buf.obj)
	b.funcs.BufferData(firstBinding, len(data), gl.STATIC_DRAW, data)
	buf.immutable = true
	if err := glErr(b.funcs); err != nil {
		buf.Release()
		return nil, err
	}
	return buf, nil
}

func glErr(f *gl.Functions) error {
	if st := f.GetError(); st != gl.NO_ERROR {
		return fmt.Errorf("glGetError: %#x", st)
	}
	return nil
}

func (b *Backend) Release() {
	if b.sRGBFBO != nil {
		b.sRGBFBO.Release()
	}
	if b.vertArray.Valid() {
		b.glstate.deleteVertexArray(b.funcs, b.vertArray)
	}
	*b = Backend{}
}

func (b *Backend) DispatchCompute(x, y, z int) {
	for binding, buf := range b.storage {
		if buf != nil {
			b.glstate.bindBufferBase(b.funcs, gl.SHADER_STORAGE_BUFFER, binding, buf.obj)
		}
	}
	b.funcs.DispatchCompute(x, y, z)
	b.funcs.MemoryBarrier(gl.ALL_BARRIER_BITS)
}

func (b *Backend) BindImageTexture(unit int, tex driver.Texture) {
	t := tex.(*texture)
	var acc gl.Enum
	switch t.bindings & (driver.BufferBindingShaderStorageRead | driver.BufferBindingShaderStorageWrite) {
	case driver.BufferBindingShaderStorageRead:
		acc = gl.READ_ONLY
	case driver.BufferBindingShaderStorageWrite:
		acc = gl.WRITE_ONLY
	case driver.BufferBindingShaderStorageRead | driver.BufferBindingShaderStorageWrite:
		acc = gl.READ_WRITE
	default:
		panic("unsupported access bits")
	}
	b.funcs.BindImageTexture(unit, t.obj, 0, false, 0, acc, t.triple.internalFormat)
}

func (b *Backend) BlendFunc(sfactor, dfactor driver.BlendFactor) {
	src, dst := toGLBlendFactor(sfactor), toGLBlendFactor(dfactor)
	b.glstate.setBlendFuncSeparate(b.funcs, src, dst, src, dst)
}

func toGLBlendFactor(f driver.BlendFactor) gl.Enum {
	switch f {
	case driver.BlendFactorOne:
		return gl.ONE
	case driver.BlendFactorOneMinusSrcAlpha:
		return gl.ONE_MINUS_SRC_ALPHA
	case driver.BlendFactorZero:
		return gl.ZERO
	case driver.BlendFactorDstColor:
		return gl.DST_COLOR
	default:
		panic("unsupported blend factor")
	}
}

func (b *Backend) SetBlend(enable bool) {
	b.glstate.set(b.funcs, gl.BLEND, enable)
}

func (b *Backend) DrawElements(off, count int) {
	b.prepareDraw()
	// off is in 16-bit indices, but DrawElements take a byte offset.
	byteOff := off * 2
	b.funcs.DrawElements(toGLDrawMode(b.state.pipeline.topology), count, gl.UNSIGNED_SHORT, byteOff)
}

func (b *Backend) DrawArrays(off, count int) {
	b.prepareDraw()
	b.funcs.DrawArrays(toGLDrawMode(b.state.pipeline.topology), off, count)
}

func (b *Backend) prepareDraw() {
	p := b.state.pipeline
	if p == nil {
		return
	}
	b.setupVertexArrays()
}

func toGLDrawMode(mode driver.Topology) gl.Enum {
	switch mode {
	case driver.TopologyTriangleStrip:
		return gl.TRIANGLE_STRIP
	case driver.TopologyTriangles:
		return gl.TRIANGLES
	default:
		panic("unsupported draw mode")
	}
}

func (b *Backend) Viewport(x, y, width, height int) {
	b.glstate.setViewport(b.funcs, x, y, width, height)
}

func (b *Backend) clearOutput(colR, colG, colB, colA float32) {
	b.glstate.setClearColor(b.funcs, colR, colG, colB, colA)
	b.funcs.Clear(gl.COLOR_BUFFER_BIT)
}

func (b *Backend) NewComputeProgram(src shader.Sources) (driver.Program, error) {
	// We don't support ES 3.1 compute, see brokenGLES31 above.
	const GLES31Source = ""
	p, err := gl.CreateComputeProgram(b.funcs, GLES31Source)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", src.Name, err)
	}
	return &program{
		backend: b,
		obj:     p,
	}, nil
}

func (b *Backend) NewVertexShader(src shader.Sources) (driver.VertexShader, error) {
	glslSrc := b.glslFor(src)
	sh, err := gl.CreateShader(b.funcs, gl.VERTEX_SHADER, glslSrc)
	return &glshader{backend: b, obj: sh, src: src}, err
}

func (b *Backend) NewFragmentShader(src shader.Sources) (driver.FragmentShader, error) {
	glslSrc := b.glslFor(src)
	sh, err := gl.CreateShader(b.funcs, gl.FRAGMENT_SHADER, glslSrc)
	return &glshader{backend: b, obj: sh, src: src}, err
}

func (b *Backend) glslFor(src shader.Sources) string {
	if b.gles {
		return src.GLSL100ES
	} else {
		return src.GLSL150
	}
}

func (b *Backend) NewPipeline(desc driver.PipelineDesc) (driver.Pipeline, error) {
	p, err := b.newProgram(desc)
	if err != nil {
		return nil, err
	}
	layout := desc.VertexLayout
	vsrc := desc.VertexShader.(*glshader).src
	if len(vsrc.Inputs) != len(layout.Inputs) {
		return nil, fmt.Errorf("opengl: got %d inputs, expected %d", len(layout.Inputs), len(vsrc.Inputs))
	}
	for i, inp := range vsrc.Inputs {
		if exp, got := inp.Size, layout.Inputs[i].Size; exp != got {
			return nil, fmt.Errorf("opengl: data size mismatch for %q: got %d expected %d", inp.Name, got, exp)
		}
	}
	return &pipeline{
		prog:     p,
		inputs:   vsrc.Inputs,
		layout:   layout,
		blend:    desc.BlendDesc,
		topology: desc.Topology,
	}, nil
}

func (b *Backend) newProgram(desc driver.PipelineDesc) (*program, error) {
	p := b.funcs.CreateProgram()
	if !p.Valid() {
		return nil, errors.New("opengl: glCreateProgram failed")
	}
	vsh, fsh := desc.VertexShader.(*glshader), desc.FragmentShader.(*glshader)
	b.funcs.AttachShader(p, vsh.obj)
	b.funcs.AttachShader(p, fsh.obj)
	for _, inp := range vsh.src.Inputs {
		b.funcs.BindAttribLocation(p, gl.Attrib(inp.Location), inp.Name)
	}
	b.funcs.LinkProgram(p)
	if b.funcs.GetProgrami(p, gl.LINK_STATUS) == 0 {
		log := b.funcs.GetProgramInfoLog(p)
		b.funcs.DeleteProgram(p)
		return nil, fmt.Errorf("opengl: program link failed: %s", strings.TrimSpace(log))
	}
	prog := &program{
		backend: b,
		obj:     p,
	}
	b.glstate.useProgram(b.funcs, p)
	// Bind texture uniforms.
	for _, tex := range vsh.src.Textures {
		u := b.funcs.GetUniformLocation(p, tex.Name)
		if u.Valid() {
			b.funcs.Uniform1i(u, tex.Binding)
		}
	}
	for _, tex := range fsh.src.Textures {
		u := b.funcs.GetUniformLocation(p, tex.Name)
		if u.Valid() {
			b.funcs.Uniform1i(u, tex.Binding)
		}
	}
	prog.vertUniforms.setup(b.funcs, p, vsh.src.Uniforms.Size, vsh.src.Uniforms.Locations)
	prog.fragUniforms.setup(b.funcs, p, fsh.src.Uniforms.Size, fsh.src.Uniforms.Locations)
	return prog, nil
}

func (b *Backend) BindStorageBuffer(binding int, buf driver.Buffer) {
	bf := buf.(*buffer)
	if bf.typ&(driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite) == 0 {
		panic("not a shader storage buffer")
	}
	b.storage[binding] = bf
}

func (b *Backend) BindUniforms(buf driver.Buffer) {
	bf := buf.(*buffer)
	if bf.typ&driver.BufferBindingUniforms == 0 {
		panic("not a uniform buffer")
	}
	b.state.pipeline.prog.vertUniforms.update(b.funcs, bf)
	b.state.pipeline.prog.fragUniforms.update(b.funcs, bf)
}

func (b *Backend) BindProgram(prog driver.Program) {
	p := prog.(*program)
	b.glstate.useProgram(b.funcs, p.obj)
}

func (s *glshader) Release() {
	s.backend.funcs.DeleteShader(s.obj)
}

func (p *program) Release() {
	p.backend.glstate.deleteProgram(p.backend.funcs, p.obj)
}

func (u *uniforms) setup(funcs *gl.Functions, p gl.Program, uniformSize int, uniforms []shader.UniformLocation) {
	u.locs = make([]uniformLocation, len(uniforms))
	for i, uniform := range uniforms {
		loc := funcs.GetUniformLocation(p, uniform.Name)
		u.locs[i] = uniformLocation{uniform: loc, offset: uniform.Offset, typ: uniform.Type, size: uniform.Size}
	}
	u.size = uniformSize
}

func (p *uniforms) update(funcs *gl.Functions, buf *buffer) {
	if buf.size < p.size {
		panic(fmt.Errorf("uniform buffer too small, got %d need %d", buf.size, p.size))
	}
	data := buf.data
	for _, u := range p.locs {
		if !u.uniform.Valid() {
			continue
		}
		data := data[u.offset:]
		switch {
		case u.typ == shader.DataTypeFloat && u.size == 1:
			data := data[:4]
			v := *(*[1]float32)(unsafe.Pointer(&data[0]))
			funcs.Uniform1f(u.uniform, v[0])
		case u.typ == shader.DataTypeFloat && u.size == 2:
			data := data[:8]
			v := *(*[2]float32)(unsafe.Pointer(&data[0]))
			funcs.Uniform2f(u.uniform, v[0], v[1])
		case u.typ == shader.DataTypeFloat && u.size == 3:
			data := data[:12]
			v := *(*[3]float32)(unsafe.Pointer(&data[0]))
			funcs.Uniform3f(u.uniform, v[0], v[1], v[2])
		case u.typ == shader.DataTypeFloat && u.size == 4:
			data := data[:16]
			v := *(*[4]float32)(unsafe.Pointer(&data[0]))
			funcs.Uniform4f(u.uniform, v[0], v[1], v[2], v[3])
		default:
			panic("unsupported uniform data type or size")
		}
	}
}

func (b *buffer) Upload(data []byte) {
	if b.immutable {
		panic("immutable buffer")
	}
	if len(data) > b.size {
		panic("buffer size overflow")
	}
	copy(b.data, data)
	if b.hasBuffer {
		firstBinding := firstBufferType(b.typ)
		b.backend.glstate.bindBuffer(b.backend.funcs, firstBinding, b.obj)
		if len(data) == b.size {
			// the iOS GL implementation doesn't recognize when BufferSubData
			// clears the entire buffer. Tell it and avoid GPU stalls.
			// See also https://github.com/godotengine/godot/issues/23956.
			b.backend.funcs.BufferData(firstBinding, b.size, gl.DYNAMIC_DRAW, data)
		} else {
			b.backend.funcs.BufferSubData(firstBinding, 0, data)
		}
	}
}

func (b *buffer) Download(data []byte) error {
	if len(data) > b.size {
		panic("buffer size overflow")
	}
	if !b.hasBuffer {
		copy(data, b.data)
		return nil
	}
	firstBinding := firstBufferType(b.typ)
	b.backend.glstate.bindBuffer(b.backend.funcs, firstBinding, b.obj)
	bufferMap := b.backend.funcs.MapBufferRange(firstBinding, 0, len(data), gl.MAP_READ_BIT)
	if bufferMap == nil {
		return fmt.Errorf("MapBufferRange: error %#x", b.backend.funcs.GetError())
	}
	copy(data, bufferMap)
	if !b.backend.funcs.UnmapBuffer(firstBinding) {
		return driver.ErrContentLost
	}
	return nil
}

func (b *buffer) Release() {
	if b.hasBuffer {
		b.backend.glstate.deleteBuffer(b.backend.funcs, b.obj)
		b.hasBuffer = false
	}
}

func (b *Backend) BindVertexBuffer(buf driver.Buffer, offset int) {
	gbuf := buf.(*buffer)
	if gbuf.typ&driver.BufferBindingVertices == 0 {
		panic("not a vertex buffer")
	}
	b.state.buffer = bufferBinding{obj: gbuf.obj, offset: offset}
}

func (b *Backend) setupVertexArrays() {
	p := b.state.pipeline
	inputs := p.inputs
	if len(inputs) == 0 {
		return
	}
	layout := p.layout
	const max = len(b.glstate.vertAttribs)
	var enabled [max]bool
	buf := b.state.buffer
	for i, inp := range inputs {
		l := layout.Inputs[i]
		var gltyp gl.Enum
		switch l.Type {
		case shader.DataTypeFloat:
			gltyp = gl.FLOAT
		case shader.DataTypeShort:
			gltyp = gl.SHORT
		default:
			panic("unsupported data type")
		}
		enabled[inp.Location] = true
		b.glstate.vertexAttribPointer(b.funcs, buf.obj, inp.Location, l.Size, gltyp, false, p.layout.Stride, buf.offset+l.Offset)
	}
	for i := 0; i < max; i++ {
		b.glstate.setVertexAttribArray(b.funcs, i, enabled[i])
	}
}

func (b *Backend) BindIndexBuffer(buf driver.Buffer) {
	gbuf := buf.(*buffer)
	if gbuf.typ&driver.BufferBindingIndices == 0 {
		panic("not an index buffer")
	}
	b.glstate.bindBuffer(b.funcs, gl.ELEMENT_ARRAY_BUFFER, gbuf.obj)
}

func (b *Backend) CopyTexture(dst driver.Texture, dstOrigin image.Point, src driver.Texture, srcRect image.Rectangle) {
	const unit = 0
	oldTex := b.glstate.texUnits.binds[unit]
	defer func() {
		b.glstate.bindTexture(b.funcs, unit, oldTex)
	}()
	b.glstate.bindTexture(b.funcs, unit, dst.(*texture).obj)
	b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, src.(*texture).ensureFBO())
	sz := srcRect.Size()
	b.funcs.CopyTexSubImage2D(gl.TEXTURE_2D, 0, dstOrigin.X, dstOrigin.Y, srcRect.Min.X, srcRect.Min.Y, sz.X, sz.Y)
}

func (t *texture) ReadPixels(src image.Rectangle, pixels []byte, stride int) error {
	glErr(t.backend.funcs)
	t.backend.glstate.bindFramebuffer(t.backend.funcs, gl.FRAMEBUFFER, t.ensureFBO())
	w, h := src.Dx(), src.Dy()
	if len(pixels) < w*h*4 {
		return errors.New("unexpected RGBA size")
	}
	// OpenGL ES 2 doesn't support PACK_ROW_LENGTH != 0. Avoid it if possible.
	rowLen := 0
	if n := stride / 4; n != w {
		rowLen = n
	}
	if rowLen == 0 || t.backend.glver[0] > 2 {
		t.backend.glstate.pixelStorei(t.backend.funcs, gl.PACK_ROW_LENGTH, rowLen)
		t.backend.funcs.ReadPixels(src.Min.X, src.Min.Y, w, h, gl.RGBA, gl.UNSIGNED_BYTE, pixels)
	} else {
		tmp := make([]byte, w*h*4)
		t.backend.funcs.ReadPixels(src.Min.X, src.Min.Y, w, h, gl.RGBA, gl.UNSIGNED_BYTE, tmp)
		for y := 0; y < h; y++ {
			copy(pixels[y*stride:], tmp[y*w*4:])
		}
	}
	return glErr(t.backend.funcs)
}

func (b *Backend) BindPipeline(pl driver.Pipeline) {
	p := pl.(*pipeline)
	b.state.pipeline = p
	b.glstate.useProgram(b.funcs, p.prog.obj)
	b.SetBlend(p.blend.Enable)
	b.BlendFunc(p.blend.SrcFactor, p.blend.DstFactor)
}

func (b *Backend) BeginCompute() {
	b.funcs.MemoryBarrier(gl.ALL_BARRIER_BITS)
}

func (b *Backend) EndCompute() {
}

func (b *Backend) BeginRenderPass(tex driver.Texture, desc driver.LoadDesc) {
	fbo := tex.(*texture).ensureFBO()
	b.glstate.bindFramebuffer(b.funcs, gl.FRAMEBUFFER, fbo)
	switch desc.Action {
	case driver.LoadActionClear:
		c := desc.ClearColor
		b.clearOutput(c.R, c.G, c.B, c.A)
	case driver.LoadActionInvalidate:
		b.funcs.InvalidateFramebuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0)
	}
}

func (b *Backend) EndRenderPass() {
}

func (f *texture) ImplementsRenderTarget() {}

func (p *pipeline) Release() {
	p.prog.Release()
	*p = pipeline{}
}

func toTexFilter(f driver.TextureFilter) (int, bool) {
	switch f {
	case driver.FilterNearest:
		return gl.NEAREST, false
	case driver.FilterLinear:
		return gl.LINEAR, false
	case driver.FilterLinearMipmapLinear:
		return gl.LINEAR_MIPMAP_LINEAR, true
	default:
		panic("unsupported texture filter")
	}
}

func (b *Backend) PrepareTexture(tex driver.Texture) {}

func (b *Backend) BindTexture(unit int, t driver.Texture) {
	b.glstate.bindTexture(b.funcs, unit, t.(*texture).obj)
}

func (t *texture) Release() {
	if t.foreign {
		panic("texture not created by NewTexture")
	}
	if t.hasFBO {
		t.backend.glstate.deleteFramebuffer(t.backend.funcs, t.fbo)
	}
	t.backend.glstate.deleteTexture(t.backend.funcs, t.obj)
}

func (t *texture) Upload(offset, size image.Point, pixels []byte, stride int) {
	if min := size.X * size.Y * 4; min > len(pixels) {
		panic(fmt.Errorf("size %d larger than data %d", min, len(pixels)))
	}
	t.backend.BindTexture(0, t)
	// WebGL 1 doesn't support UNPACK_ROW_LENGTH != 0. Avoid it if possible.
	rowLen := 0
	if n := stride / 4; n != size.X {
		rowLen = n
	}
	t.backend.glstate.pixelStorei(t.backend.funcs, gl.UNPACK_ROW_LENGTH, rowLen)
	t.backend.funcs.TexSubImage2D(gl.TEXTURE_2D, 0, offset.X, offset.Y, size.X, size.Y, t.triple.format, t.triple.typ, pixels)
	if t.mipmap {
		t.backend.funcs.GenerateMipmap(gl.TEXTURE_2D)
	}
}

func (t *timer) Begin() {
	t.funcs.BeginQuery(gl.TIME_ELAPSED_EXT, t.obj)
}

func (t *timer) End() {
	t.funcs.EndQuery(gl.TIME_ELAPSED_EXT)
}

func (t *timer) ready() bool {
	return t.funcs.GetQueryObjectuiv(t.obj, gl.QUERY_RESULT_AVAILABLE) == gl.TRUE
}

func (t *timer) Release() {
	t.funcs.DeleteQuery(t.obj)
}

func (t *timer) Duration() (time.Duration, bool) {
	if !t.ready() {
		return 0, false
	}
	nanos := t.funcs.GetQueryObjectuiv(t.obj, gl.QUERY_RESULT)
	return time.Duration(nanos), true
}

// floatTripleFor determines the best texture triple for floating point FBOs.
func floatTripleFor(f *gl.Functions, ver [2]int, exts []string) (textureTriple, error) {
	var triples []textureTriple
	if ver[0] >= 3 {
		triples = append(triples, textureTriple{gl.R16F, gl.Enum(gl.RED), gl.Enum(gl.HALF_FLOAT)})
	}
	// According to the OES_texture_half_float specification, EXT_color_buffer_half_float is needed to
	// render to FBOs. However, the Safari WebGL1 implementation does support half-float FBOs but does not
	// report EXT_color_buffer_half_float support. The triples are verified below, so it doesn't matter if we're
	// wrong.
	if hasExtension(exts, "GL_OES_texture_half_float") || hasExtension(exts, "GL_EXT_color_buffer_half_float") {
		// Try single channel.
		triples = append(triples, textureTriple{gl.LUMINANCE, gl.Enum(gl.LUMINANCE), gl.Enum(gl.HALF_FLOAT_OES)})
		// Fallback to 4 channels.
		triples = append(triples, textureTriple{gl.RGBA, gl.Enum(gl.RGBA), gl.Enum(gl.HALF_FLOAT_OES)})
	}
	if hasExtension(exts, "GL_OES_texture_float") || hasExtension(exts, "GL_EXT_color_buffer_float") {
		triples = append(triples, textureTriple{gl.RGBA, gl.Enum(gl.RGBA), gl.Enum(gl.FLOAT)})
	}
	tex := f.CreateTexture()
	defer f.DeleteTexture(tex)
	defTex := gl.Texture(f.GetBinding(gl.TEXTURE_BINDING_2D))
	defer f.BindTexture(gl.TEXTURE_2D, defTex)
	f.BindTexture(gl.TEXTURE_2D, tex)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	fbo := f.CreateFramebuffer()
	defer f.DeleteFramebuffer(fbo)
	defFBO := gl.Framebuffer(f.GetBinding(gl.FRAMEBUFFER_BINDING))
	f.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	defer f.BindFramebuffer(gl.FRAMEBUFFER, defFBO)
	var attempts []string
	for _, tt := range triples {
		const size = 256
		f.TexImage2D(gl.TEXTURE_2D, 0, tt.internalFormat, size, size, tt.format, tt.typ)
		f.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tex, 0)
		st := f.CheckFramebufferStatus(gl.FRAMEBUFFER)
		if st == gl.FRAMEBUFFER_COMPLETE {
			return tt, nil
		}
		attempts = append(attempts, fmt.Sprintf("(0x%x, 0x%x, 0x%x): 0x%x", tt.internalFormat, tt.format, tt.typ, st))
	}
	return textureTriple{}, fmt.Errorf("floating point fbos not supported (attempted %s)", attempts)
}

func srgbaTripleFor(ver [2]int, exts []string) (textureTriple, error) {
	switch {
	case ver[0] >= 3:
		return textureTriple{gl.SRGB8_ALPHA8, gl.Enum(gl.RGBA), gl.Enum(gl.UNSIGNED_BYTE)}, nil
	case hasExtension(exts, "GL_EXT_sRGB"):
		return textureTriple{gl.SRGB_ALPHA_EXT, gl.Enum(gl.SRGB_ALPHA_EXT), gl.Enum(gl.UNSIGNED_BYTE)}, nil
	default:
		return textureTriple{}, errors.New("no sRGB texture formats found")
	}
}

func alphaTripleFor(ver [2]int) textureTriple {
	intf, f := gl.Enum(gl.R8), gl.Enum(gl.RED)
	if ver[0] < 3 {
		// R8, RED not supported on OpenGL ES 2.0.
		intf, f = gl.LUMINANCE, gl.Enum(gl.LUMINANCE)
	}
	return textureTriple{intf, f, gl.UNSIGNED_BYTE}
}

func hasExtension(exts []string, ext string) bool {
	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}

func firstBufferType(typ driver.BufferBinding) gl.Enum {
	switch {
	case typ&driver.BufferBindingIndices != 0:
		return gl.ELEMENT_ARRAY_BUFFER
	case typ&driver.BufferBindingVertices != 0:
		return gl.ARRAY_BUFFER
	case typ&driver.BufferBindingUniforms != 0:
		return gl.UNIFORM_BUFFER
	case typ&(driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite) != 0:
		return gl.SHADER_STORAGE_BUFFER
	default:
		panic("unsupported buffer type")
	}
}
