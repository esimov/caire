// SPDX-License-Identifier: Unlicense OR MIT

package gl

import (
	"errors"
	"strings"
	"syscall/js"
)

type Functions struct {
	Ctx                             js.Value
	EXT_disjoint_timer_query        js.Value
	EXT_disjoint_timer_query_webgl2 js.Value

	// Cached reference to the Uint8Array JS type.
	uint8Array js.Value

	// Cached JS arrays.
	arrayBuf js.Value
	int32Buf js.Value

	isWebGL2 bool
}

type Context js.Value

func NewFunctions(ctx Context, forceES bool) (*Functions, error) {
	f := &Functions{
		Ctx:        js.Value(ctx),
		uint8Array: js.Global().Get("Uint8Array"),
	}
	if err := f.Init(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *Functions) Init() error {
	webgl2Class := js.Global().Get("WebGL2RenderingContext")
	f.isWebGL2 = !webgl2Class.IsUndefined() && f.Ctx.InstanceOf(webgl2Class)
	if !f.isWebGL2 {
		f.EXT_disjoint_timer_query = f.getExtension("EXT_disjoint_timer_query")
		if f.getExtension("OES_texture_half_float").IsNull() && f.getExtension("OES_texture_float").IsNull() {
			return errors.New("gl: no support for neither OES_texture_half_float nor OES_texture_float")
		}
		if f.getExtension("EXT_sRGB").IsNull() {
			return errors.New("gl: EXT_sRGB not supported")
		}
	} else {
		// WebGL2 extensions.
		f.EXT_disjoint_timer_query_webgl2 = f.getExtension("EXT_disjoint_timer_query_webgl2")
		if f.getExtension("EXT_color_buffer_half_float").IsNull() && f.getExtension("EXT_color_buffer_float").IsNull() {
			return errors.New("gl: no support for neither EXT_color_buffer_half_float nor EXT_color_buffer_float")
		}
	}
	return nil
}

func (f *Functions) getExtension(name string) js.Value {
	return f.Ctx.Call("getExtension", name)
}

func (f *Functions) ActiveTexture(t Enum) {
	f.Ctx.Call("activeTexture", int(t))
}
func (f *Functions) AttachShader(p Program, s Shader) {
	f.Ctx.Call("attachShader", js.Value(p), js.Value(s))
}
func (f *Functions) BeginQuery(target Enum, query Query) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		f.Ctx.Call("beginQuery", int(target), js.Value(query))
	} else {
		f.EXT_disjoint_timer_query.Call("beginQueryEXT", int(target), js.Value(query))
	}
}
func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	f.Ctx.Call("bindAttribLocation", js.Value(p), int(a), name)
}
func (f *Functions) BindBuffer(target Enum, b Buffer) {
	f.Ctx.Call("bindBuffer", int(target), js.Value(b))
}
func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	f.Ctx.Call("bindBufferBase", int(target), index, js.Value(b))
}
func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	f.Ctx.Call("bindFramebuffer", int(target), js.Value(fb))
}
func (f *Functions) BindRenderbuffer(target Enum, rb Renderbuffer) {
	f.Ctx.Call("bindRenderbuffer", int(target), js.Value(rb))
}
func (f *Functions) BindTexture(target Enum, t Texture) {
	f.Ctx.Call("bindTexture", int(target), js.Value(t))
}
func (f *Functions) BindImageTexture(unit int, t Texture, level int, layered bool, layer int, access, format Enum) {
	panic("not implemented")
}
func (f *Functions) BindVertexArray(a VertexArray) {
	panic("not supported")
}
func (f *Functions) BlendEquation(mode Enum) {
	f.Ctx.Call("blendEquation", int(mode))
}
func (f *Functions) BlendFuncSeparate(srcRGB, dstRGB, srcA, dstA Enum) {
	f.Ctx.Call("blendFunc", int(srcRGB), int(dstRGB), int(srcA), int(dstA))
}
func (f *Functions) BufferData(target Enum, size int, usage Enum, data []byte) {
	if data == nil {
		f.Ctx.Call("bufferData", int(target), size, int(usage))
	} else {
		if len(data) != size {
			panic("size mismatch")
		}
		f.Ctx.Call("bufferData", int(target), f.byteArrayOf(data), int(usage))
	}
}
func (f *Functions) BufferSubData(target Enum, offset int, src []byte) {
	f.Ctx.Call("bufferSubData", int(target), offset, f.byteArrayOf(src))
}
func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(f.Ctx.Call("checkFramebufferStatus", int(target)).Int())
}
func (f *Functions) Clear(mask Enum) {
	f.Ctx.Call("clear", int(mask))
}
func (f *Functions) ClearColor(red, green, blue, alpha float32) {
	f.Ctx.Call("clearColor", red, green, blue, alpha)
}
func (f *Functions) ClearDepthf(d float32) {
	f.Ctx.Call("clearDepth", d)
}
func (f *Functions) CompileShader(s Shader) {
	f.Ctx.Call("compileShader", js.Value(s))
}
func (f *Functions) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	f.Ctx.Call("copyTexSubImage2D", int(target), level, xoffset, yoffset, x, y, width, height)
}
func (f *Functions) CreateBuffer() Buffer {
	return Buffer(f.Ctx.Call("createBuffer"))
}
func (f *Functions) CreateFramebuffer() Framebuffer {
	return Framebuffer(f.Ctx.Call("createFramebuffer"))
}
func (f *Functions) CreateProgram() Program {
	return Program(f.Ctx.Call("createProgram"))
}
func (f *Functions) CreateQuery() Query {
	return Query(f.Ctx.Call("createQuery"))
}
func (f *Functions) CreateRenderbuffer() Renderbuffer {
	return Renderbuffer(f.Ctx.Call("createRenderbuffer"))
}
func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader(f.Ctx.Call("createShader", int(ty)))
}
func (f *Functions) CreateTexture() Texture {
	return Texture(f.Ctx.Call("createTexture"))
}
func (f *Functions) CreateVertexArray() VertexArray {
	panic("not supported")
}
func (f *Functions) DeleteBuffer(v Buffer) {
	f.Ctx.Call("deleteBuffer", js.Value(v))
}
func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	f.Ctx.Call("deleteFramebuffer", js.Value(v))
}
func (f *Functions) DeleteProgram(p Program) {
	f.Ctx.Call("deleteProgram", js.Value(p))
}
func (f *Functions) DeleteQuery(query Query) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		f.Ctx.Call("deleteQuery", js.Value(query))
	} else {
		f.EXT_disjoint_timer_query.Call("deleteQueryEXT", js.Value(query))
	}
}
func (f *Functions) DeleteShader(s Shader) {
	f.Ctx.Call("deleteShader", js.Value(s))
}
func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	f.Ctx.Call("deleteRenderbuffer", js.Value(v))
}
func (f *Functions) DeleteTexture(v Texture) {
	f.Ctx.Call("deleteTexture", js.Value(v))
}
func (f *Functions) DeleteVertexArray(a VertexArray) {
	panic("not implemented")
}
func (f *Functions) DepthFunc(fn Enum) {
	f.Ctx.Call("depthFunc", int(fn))
}
func (f *Functions) DepthMask(mask bool) {
	f.Ctx.Call("depthMask", mask)
}
func (f *Functions) DisableVertexAttribArray(a Attrib) {
	f.Ctx.Call("disableVertexAttribArray", int(a))
}
func (f *Functions) Disable(cap Enum) {
	f.Ctx.Call("disable", int(cap))
}
func (f *Functions) DrawArrays(mode Enum, first, count int) {
	f.Ctx.Call("drawArrays", int(mode), first, count)
}
func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	f.Ctx.Call("drawElements", int(mode), count, int(ty), offset)
}
func (f *Functions) DispatchCompute(x, y, z int) {
	panic("not implemented")
}
func (f *Functions) Enable(cap Enum) {
	f.Ctx.Call("enable", int(cap))
}
func (f *Functions) EnableVertexAttribArray(a Attrib) {
	f.Ctx.Call("enableVertexAttribArray", int(a))
}
func (f *Functions) EndQuery(target Enum) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		f.Ctx.Call("endQuery", int(target))
	} else {
		f.EXT_disjoint_timer_query.Call("endQueryEXT", int(target))
	}
}
func (f *Functions) Finish() {
	f.Ctx.Call("finish")
}
func (f *Functions) Flush() {
	f.Ctx.Call("flush")
}
func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	f.Ctx.Call("framebufferRenderbuffer", int(target), int(attachment), int(renderbuffertarget), js.Value(renderbuffer))
}
func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	f.Ctx.Call("framebufferTexture2D", int(target), int(attachment), int(texTarget), js.Value(t), level)
}
func (f *Functions) GetError() Enum {
	// Avoid slow getError calls. See gio#179.
	return 0
}
func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	return paramVal(f.Ctx.Call("getRenderbufferParameteri", int(pname)))
}
func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	if !f.isWebGL2 && pname == FRAMEBUFFER_ATTACHMENT_COLOR_ENCODING {
		// FRAMEBUFFER_ATTACHMENT_COLOR_ENCODING is only available on WebGL 2
		return LINEAR
	}
	return paramVal(f.Ctx.Call("getFramebufferAttachmentParameter", int(target), int(attachment), int(pname)))
}
func (f *Functions) GetBinding(pname Enum) Object {
	obj := f.Ctx.Call("getParameter", int(pname))
	if !obj.Truthy() {
		return Object{}
	}
	return Object(obj)
}
func (f *Functions) GetBindingi(pname Enum, idx int) Object {
	obj := f.Ctx.Call("getIndexedParameter", int(pname), idx)
	if !obj.Truthy() {
		return Object{}
	}
	return Object(obj)
}
func (f *Functions) GetInteger(pname Enum) int {
	if !f.isWebGL2 {
		switch pname {
		case PACK_ROW_LENGTH, UNPACK_ROW_LENGTH:
			return 0 // PACK_ROW_LENGTH and UNPACK_ROW_LENGTH is only available on WebGL 2
		}
	}
	return paramVal(f.Ctx.Call("getParameter", int(pname)))
}
func (f *Functions) GetFloat(pname Enum) float32 {
	return float32(f.Ctx.Call("getParameter", int(pname)).Float())
}
func (f *Functions) GetInteger4(pname Enum) [4]int {
	arr := f.Ctx.Call("getParameter", int(pname))
	var res [4]int
	for i := range res {
		res[i] = arr.Index(i).Int()
	}
	return res
}
func (f *Functions) GetFloat4(pname Enum) [4]float32 {
	arr := f.Ctx.Call("getParameter", int(pname))
	var res [4]float32
	for i := range res {
		res[i] = float32(arr.Index(i).Float())
	}
	return res
}
func (f *Functions) GetProgrami(p Program, pname Enum) int {
	return paramVal(f.Ctx.Call("getProgramParameter", js.Value(p), int(pname)))
}
func (f *Functions) GetProgramInfoLog(p Program) string {
	return f.Ctx.Call("getProgramInfoLog", js.Value(p)).String()
}
func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		return uint(paramVal(f.Ctx.Call("getQueryParameter", js.Value(query), int(pname))))
	} else {
		return uint(paramVal(f.EXT_disjoint_timer_query.Call("getQueryObjectEXT", js.Value(query), int(pname))))
	}
}
func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	return paramVal(f.Ctx.Call("getShaderParameter", js.Value(s), int(pname)))
}
func (f *Functions) GetShaderInfoLog(s Shader) string {
	return f.Ctx.Call("getShaderInfoLog", js.Value(s)).String()
}
func (f *Functions) GetString(pname Enum) string {
	switch pname {
	case EXTENSIONS:
		extsjs := f.Ctx.Call("getSupportedExtensions")
		var exts []string
		for i := 0; i < extsjs.Length(); i++ {
			exts = append(exts, "GL_"+extsjs.Index(i).String())
		}
		return strings.Join(exts, " ")
	default:
		return f.Ctx.Call("getParameter", int(pname)).String()
	}
}
func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	return uint(paramVal(f.Ctx.Call("getUniformBlockIndex", js.Value(p), name)))
}
func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	return Uniform(f.Ctx.Call("getUniformLocation", js.Value(p), name))
}
func (f *Functions) GetVertexAttrib(index int, pname Enum) int {
	return paramVal(f.Ctx.Call("getVertexAttrib", index, int(pname)))
}
func (f *Functions) GetVertexAttribBinding(index int, pname Enum) Object {
	obj := f.Ctx.Call("getVertexAttrib", index, int(pname))
	if !obj.Truthy() {
		return Object{}
	}
	return Object(obj)
}
func (f *Functions) GetVertexAttribPointer(index int, pname Enum) uintptr {
	return uintptr(f.Ctx.Call("getVertexAttribOffset", index, int(pname)).Int())
}
func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	fn := f.Ctx.Get("invalidateFramebuffer")
	if !fn.IsUndefined() {
		if f.int32Buf.IsUndefined() {
			f.int32Buf = js.Global().Get("Int32Array").New(1)
		}
		f.int32Buf.SetIndex(0, int32(attachment))
		f.Ctx.Call("invalidateFramebuffer", int(target), f.int32Buf)
	}
}
func (f *Functions) IsEnabled(cap Enum) bool {
	return f.Ctx.Call("isEnabled", int(cap)).Truthy()
}
func (f *Functions) LinkProgram(p Program) {
	f.Ctx.Call("linkProgram", js.Value(p))
}
func (f *Functions) PixelStorei(pname Enum, param int) {
	f.Ctx.Call("pixelStorei", int(pname), param)
}
func (f *Functions) MemoryBarrier(barriers Enum) {
	panic("not implemented")
}
func (f *Functions) MapBufferRange(target Enum, offset, length int, access Enum) []byte {
	panic("not implemented")
}
func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	f.Ctx.Call("renderbufferStorage", int(target), int(internalformat), width, height)
}
func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	ba := f.byteArrayOf(data)
	f.Ctx.Call("readPixels", x, y, width, height, int(format), int(ty), ba)
	js.CopyBytesToGo(data, ba)
}
func (f *Functions) Scissor(x, y, width, height int32) {
	f.Ctx.Call("scissor", x, y, width, height)
}
func (f *Functions) ShaderSource(s Shader, src string) {
	f.Ctx.Call("shaderSource", js.Value(s), src)
}
func (f *Functions) TexImage2D(target Enum, level int, internalFormat Enum, width, height int, format, ty Enum) {
	f.Ctx.Call("texImage2D", int(target), int(level), int(internalFormat), int(width), int(height), 0, int(format), int(ty), nil)
}
func (f *Functions) TexStorage2D(target Enum, levels int, internalFormat Enum, width, height int) {
	f.Ctx.Call("texStorage2D", int(target), levels, int(internalFormat), width, height)
}
func (f *Functions) TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	f.Ctx.Call("texSubImage2D", int(target), level, x, y, width, height, int(format), int(ty), f.byteArrayOf(data))
}
func (f *Functions) TexParameteri(target, pname Enum, param int) {
	f.Ctx.Call("texParameteri", int(target), int(pname), int(param))
}
func (f *Functions) UniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint) {
	f.Ctx.Call("uniformBlockBinding", js.Value(p), int(uniformBlockIndex), int(uniformBlockBinding))
}
func (f *Functions) Uniform1f(dst Uniform, v float32) {
	f.Ctx.Call("uniform1f", js.Value(dst), v)
}
func (f *Functions) Uniform1i(dst Uniform, v int) {
	f.Ctx.Call("uniform1i", js.Value(dst), v)
}
func (f *Functions) Uniform2f(dst Uniform, v0, v1 float32) {
	f.Ctx.Call("uniform2f", js.Value(dst), v0, v1)
}
func (f *Functions) Uniform3f(dst Uniform, v0, v1, v2 float32) {
	f.Ctx.Call("uniform3f", js.Value(dst), v0, v1, v2)
}
func (f *Functions) Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	f.Ctx.Call("uniform4f", js.Value(dst), v0, v1, v2, v3)
}
func (f *Functions) UseProgram(p Program) {
	f.Ctx.Call("useProgram", js.Value(p))
}
func (f *Functions) UnmapBuffer(target Enum) bool {
	panic("not implemented")
}
func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	f.Ctx.Call("vertexAttribPointer", int(dst), size, int(ty), normalized, stride, offset)
}
func (f *Functions) Viewport(x, y, width, height int) {
	f.Ctx.Call("viewport", x, y, width, height)
}

func (f *Functions) byteArrayOf(data []byte) js.Value {
	if len(data) == 0 {
		return js.Null()
	}
	f.resizeByteBuffer(len(data))
	ba := f.uint8Array.New(f.arrayBuf, int(0), int(len(data)))
	js.CopyBytesToJS(ba, data)
	return ba
}

func (f *Functions) resizeByteBuffer(n int) {
	if n == 0 {
		return
	}
	if !f.arrayBuf.IsUndefined() && f.arrayBuf.Length() >= n {
		return
	}
	f.arrayBuf = js.Global().Get("ArrayBuffer").New(n)
}

func paramVal(v js.Value) int {
	switch v.Type() {
	case js.TypeBoolean:
		if b := v.Bool(); b {
			return 1
		} else {
			return 0
		}
	case js.TypeNumber:
		return v.Int()
	default:
		panic("unknown parameter type")
	}
}
