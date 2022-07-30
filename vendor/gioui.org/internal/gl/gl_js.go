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

	_getExtension                      js.Value
	_activeTexture                     js.Value
	_attachShader                      js.Value
	_beginQuery                        js.Value
	_beginQueryEXT                     js.Value
	_bindAttribLocation                js.Value
	_bindBuffer                        js.Value
	_bindBufferBase                    js.Value
	_bindFramebuffer                   js.Value
	_bindRenderbuffer                  js.Value
	_bindTexture                       js.Value
	_blendEquation                     js.Value
	_blendFunc                         js.Value
	_bufferData                        js.Value
	_bufferSubData                     js.Value
	_checkFramebufferStatus            js.Value
	_clear                             js.Value
	_clearColor                        js.Value
	_clearDepth                        js.Value
	_compileShader                     js.Value
	_copyTexSubImage2D                 js.Value
	_createBuffer                      js.Value
	_createFramebuffer                 js.Value
	_createProgram                     js.Value
	_createQuery                       js.Value
	_createRenderbuffer                js.Value
	_createShader                      js.Value
	_createTexture                     js.Value
	_deleteBuffer                      js.Value
	_deleteFramebuffer                 js.Value
	_deleteProgram                     js.Value
	_deleteQuery                       js.Value
	_deleteQueryEXT                    js.Value
	_deleteShader                      js.Value
	_deleteRenderbuffer                js.Value
	_deleteTexture                     js.Value
	_depthFunc                         js.Value
	_depthMask                         js.Value
	_disableVertexAttribArray          js.Value
	_disable                           js.Value
	_drawArrays                        js.Value
	_drawElements                      js.Value
	_enable                            js.Value
	_enableVertexAttribArray           js.Value
	_endQuery                          js.Value
	_endQueryEXT                       js.Value
	_finish                            js.Value
	_flush                             js.Value
	_framebufferRenderbuffer           js.Value
	_framebufferTexture2D              js.Value
	_getRenderbufferParameteri         js.Value
	_getFramebufferAttachmentParameter js.Value
	_getParameter                      js.Value
	_getIndexedParameter               js.Value
	_getProgramParameter               js.Value
	_getProgramInfoLog                 js.Value
	_getQueryParameter                 js.Value
	_getQueryObjectEXT                 js.Value
	_getShaderParameter                js.Value
	_getShaderInfoLog                  js.Value
	_getSupportedExtensions            js.Value
	_getUniformBlockIndex              js.Value
	_getUniformLocation                js.Value
	_getVertexAttrib                   js.Value
	_getVertexAttribOffset             js.Value
	_invalidateFramebuffer             js.Value
	_isEnabled                         js.Value
	_linkProgram                       js.Value
	_pixelStorei                       js.Value
	_renderbufferStorage               js.Value
	_readPixels                        js.Value
	_scissor                           js.Value
	_shaderSource                      js.Value
	_texImage2D                        js.Value
	_texStorage2D                      js.Value
	_texSubImage2D                     js.Value
	_texParameteri                     js.Value
	_uniformBlockBinding               js.Value
	_uniform1f                         js.Value
	_uniform1i                         js.Value
	_uniform2f                         js.Value
	_uniform3f                         js.Value
	_uniform4f                         js.Value
	_useProgram                        js.Value
	_vertexAttribPointer               js.Value
	_viewport                          js.Value
}

type Context js.Value

func NewFunctions(ctx Context, forceES bool) (*Functions, error) {
	webgl := js.Value(ctx)
	f := &Functions{
		Ctx:                                webgl,
		uint8Array:                         js.Global().Get("Uint8Array"),
		_getExtension:                      _bind(webgl, `getExtension`),
		_activeTexture:                     _bind(webgl, `activeTexture`),
		_attachShader:                      _bind(webgl, `attachShader`),
		_beginQuery:                        _bind(webgl, `beginQuery`),
		_beginQueryEXT:                     _bind(webgl, `beginQueryEXT`),
		_bindAttribLocation:                _bind(webgl, `bindAttribLocation`),
		_bindBuffer:                        _bind(webgl, `bindBuffer`),
		_bindBufferBase:                    _bind(webgl, `bindBufferBase`),
		_bindFramebuffer:                   _bind(webgl, `bindFramebuffer`),
		_bindRenderbuffer:                  _bind(webgl, `bindRenderbuffer`),
		_bindTexture:                       _bind(webgl, `bindTexture`),
		_blendEquation:                     _bind(webgl, `blendEquation`),
		_blendFunc:                         _bind(webgl, `blendFunc`),
		_bufferData:                        _bind(webgl, `bufferData`),
		_bufferSubData:                     _bind(webgl, `bufferSubData`),
		_checkFramebufferStatus:            _bind(webgl, `checkFramebufferStatus`),
		_clear:                             _bind(webgl, `clear`),
		_clearColor:                        _bind(webgl, `clearColor`),
		_clearDepth:                        _bind(webgl, `clearDepth`),
		_compileShader:                     _bind(webgl, `compileShader`),
		_copyTexSubImage2D:                 _bind(webgl, `copyTexSubImage2D`),
		_createBuffer:                      _bind(webgl, `createBuffer`),
		_createFramebuffer:                 _bind(webgl, `createFramebuffer`),
		_createProgram:                     _bind(webgl, `createProgram`),
		_createQuery:                       _bind(webgl, `createQuery`),
		_createRenderbuffer:                _bind(webgl, `createRenderbuffer`),
		_createShader:                      _bind(webgl, `createShader`),
		_createTexture:                     _bind(webgl, `createTexture`),
		_deleteBuffer:                      _bind(webgl, `deleteBuffer`),
		_deleteFramebuffer:                 _bind(webgl, `deleteFramebuffer`),
		_deleteProgram:                     _bind(webgl, `deleteProgram`),
		_deleteQuery:                       _bind(webgl, `deleteQuery`),
		_deleteQueryEXT:                    _bind(webgl, `deleteQueryEXT`),
		_deleteShader:                      _bind(webgl, `deleteShader`),
		_deleteRenderbuffer:                _bind(webgl, `deleteRenderbuffer`),
		_deleteTexture:                     _bind(webgl, `deleteTexture`),
		_depthFunc:                         _bind(webgl, `depthFunc`),
		_depthMask:                         _bind(webgl, `depthMask`),
		_disableVertexAttribArray:          _bind(webgl, `disableVertexAttribArray`),
		_disable:                           _bind(webgl, `disable`),
		_drawArrays:                        _bind(webgl, `drawArrays`),
		_drawElements:                      _bind(webgl, `drawElements`),
		_enable:                            _bind(webgl, `enable`),
		_enableVertexAttribArray:           _bind(webgl, `enableVertexAttribArray`),
		_endQuery:                          _bind(webgl, `endQuery`),
		_endQueryEXT:                       _bind(webgl, `endQueryEXT`),
		_finish:                            _bind(webgl, `finish`),
		_flush:                             _bind(webgl, `flush`),
		_framebufferRenderbuffer:           _bind(webgl, `framebufferRenderbuffer`),
		_framebufferTexture2D:              _bind(webgl, `framebufferTexture2D`),
		_getRenderbufferParameteri:         _bind(webgl, `getRenderbufferParameteri`),
		_getFramebufferAttachmentParameter: _bind(webgl, `getFramebufferAttachmentParameter`),
		_getParameter:                      _bind(webgl, `getParameter`),
		_getIndexedParameter:               _bind(webgl, `getIndexedParameter`),
		_getProgramParameter:               _bind(webgl, `getProgramParameter`),
		_getProgramInfoLog:                 _bind(webgl, `getProgramInfoLog`),
		_getQueryParameter:                 _bind(webgl, `getQueryParameter`),
		_getQueryObjectEXT:                 _bind(webgl, `getQueryObjectEXT`),
		_getShaderParameter:                _bind(webgl, `getShaderParameter`),
		_getShaderInfoLog:                  _bind(webgl, `getShaderInfoLog`),
		_getSupportedExtensions:            _bind(webgl, `getSupportedExtensions`),
		_getUniformBlockIndex:              _bind(webgl, `getUniformBlockIndex`),
		_getUniformLocation:                _bind(webgl, `getUniformLocation`),
		_getVertexAttrib:                   _bind(webgl, `getVertexAttrib`),
		_getVertexAttribOffset:             _bind(webgl, `getVertexAttribOffset`),
		_invalidateFramebuffer:             _bind(webgl, `invalidateFramebuffer`),
		_isEnabled:                         _bind(webgl, `isEnabled`),
		_linkProgram:                       _bind(webgl, `linkProgram`),
		_pixelStorei:                       _bind(webgl, `pixelStorei`),
		_renderbufferStorage:               _bind(webgl, `renderbufferStorage`),
		_readPixels:                        _bind(webgl, `readPixels`),
		_scissor:                           _bind(webgl, `scissor`),
		_shaderSource:                      _bind(webgl, `shaderSource`),
		_texImage2D:                        _bind(webgl, `texImage2D`),
		_texStorage2D:                      _bind(webgl, `texStorage2D`),
		_texSubImage2D:                     _bind(webgl, `texSubImage2D`),
		_texParameteri:                     _bind(webgl, `texParameteri`),
		_uniformBlockBinding:               _bind(webgl, `uniformBlockBinding`),
		_uniform1f:                         _bind(webgl, `uniform1f`),
		_uniform1i:                         _bind(webgl, `uniform1i`),
		_uniform2f:                         _bind(webgl, `uniform2f`),
		_uniform3f:                         _bind(webgl, `uniform3f`),
		_uniform4f:                         _bind(webgl, `uniform4f`),
		_useProgram:                        _bind(webgl, `useProgram`),
		_vertexAttribPointer:               _bind(webgl, `vertexAttribPointer`),
		_viewport:                          _bind(webgl, `viewport`),
	}
	if err := f.Init(); err != nil {
		return nil, err
	}
	return f, nil
}

func _bind(ctx js.Value, p string) js.Value {
	if o := ctx.Get(p); o.Truthy() {
		return o.Call("bind", ctx)
	}
	return js.Undefined()
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
	return f._getExtension.Invoke(name)
}

func (f *Functions) ActiveTexture(t Enum) {
	f._activeTexture.Invoke(int(t))
}
func (f *Functions) AttachShader(p Program, s Shader) {
	f._attachShader.Invoke(js.Value(p), js.Value(s))
}
func (f *Functions) BeginQuery(target Enum, query Query) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		f._beginQuery.Invoke(int(target), js.Value(query))
	} else {
		f.EXT_disjoint_timer_query.Call("beginQueryEXT", int(target), js.Value(query))
	}
}
func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	f._bindAttribLocation.Invoke(js.Value(p), int(a), name)
}
func (f *Functions) BindBuffer(target Enum, b Buffer) {
	f._bindBuffer.Invoke(int(target), js.Value(b))
}
func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	f._bindBufferBase.Invoke(int(target), index, js.Value(b))
}
func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	f._bindFramebuffer.Invoke(int(target), js.Value(fb))
}
func (f *Functions) BindRenderbuffer(target Enum, rb Renderbuffer) {
	f._bindRenderbuffer.Invoke(int(target), js.Value(rb))
}
func (f *Functions) BindTexture(target Enum, t Texture) {
	f._bindTexture.Invoke(int(target), js.Value(t))
}
func (f *Functions) BindImageTexture(unit int, t Texture, level int, layered bool, layer int, access, format Enum) {
	panic("not implemented")
}
func (f *Functions) BindVertexArray(a VertexArray) {
	panic("not supported")
}
func (f *Functions) BlendEquation(mode Enum) {
	f._blendEquation.Invoke(int(mode))
}
func (f *Functions) BlendFuncSeparate(srcRGB, dstRGB, srcA, dstA Enum) {
	f._blendFunc.Invoke(int(srcRGB), int(dstRGB), int(srcA), int(dstA))
}
func (f *Functions) BufferData(target Enum, size int, usage Enum, data []byte) {
	if data == nil {
		f._bufferData.Invoke(int(target), size, int(usage))
	} else {
		if len(data) != size {
			panic("size mismatch")
		}
		f._bufferData.Invoke(int(target), f.byteArrayOf(data), int(usage))
	}
}
func (f *Functions) BufferSubData(target Enum, offset int, src []byte) {
	f._bufferSubData.Invoke(int(target), offset, f.byteArrayOf(src))
}
func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(f._checkFramebufferStatus.Invoke(int(target)).Int())
}
func (f *Functions) Clear(mask Enum) {
	f._clear.Invoke(int(mask))
}
func (f *Functions) ClearColor(red, green, blue, alpha float32) {
	f._clearColor.Invoke(red, green, blue, alpha)
}
func (f *Functions) ClearDepthf(d float32) {
	f._clearDepth.Invoke(d)
}
func (f *Functions) CompileShader(s Shader) {
	f._compileShader.Invoke(js.Value(s))
}
func (f *Functions) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	f._copyTexSubImage2D.Invoke(int(target), level, xoffset, yoffset, x, y, width, height)
}
func (f *Functions) CreateBuffer() Buffer {
	return Buffer(f._createBuffer.Invoke())
}
func (f *Functions) CreateFramebuffer() Framebuffer {
	return Framebuffer(f._createFramebuffer.Invoke())
}
func (f *Functions) CreateProgram() Program {
	return Program(f._createProgram.Invoke())
}
func (f *Functions) CreateQuery() Query {
	return Query(f._createQuery.Invoke())
}
func (f *Functions) CreateRenderbuffer() Renderbuffer {
	return Renderbuffer(f._createRenderbuffer.Invoke())
}
func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader(f._createShader.Invoke(int(ty)))
}
func (f *Functions) CreateTexture() Texture {
	return Texture(f._createTexture.Invoke())
}
func (f *Functions) CreateVertexArray() VertexArray {
	panic("not supported")
}
func (f *Functions) DeleteBuffer(v Buffer) {
	f._deleteBuffer.Invoke(js.Value(v))
}
func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	f._deleteFramebuffer.Invoke(js.Value(v))
}
func (f *Functions) DeleteProgram(p Program) {
	f._deleteProgram.Invoke(js.Value(p))
}
func (f *Functions) DeleteQuery(query Query) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		f._deleteQuery.Invoke(js.Value(query))
	} else {
		f.EXT_disjoint_timer_query.Call("deleteQueryEXT", js.Value(query))
	}
}
func (f *Functions) DeleteShader(s Shader) {
	f._deleteShader.Invoke(js.Value(s))
}
func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	f._deleteRenderbuffer.Invoke(js.Value(v))
}
func (f *Functions) DeleteTexture(v Texture) {
	f._deleteTexture.Invoke(js.Value(v))
}
func (f *Functions) DeleteVertexArray(a VertexArray) {
	panic("not implemented")
}
func (f *Functions) DepthFunc(fn Enum) {
	f._depthFunc.Invoke(int(fn))
}
func (f *Functions) DepthMask(mask bool) {
	f._depthMask.Invoke(mask)
}
func (f *Functions) DisableVertexAttribArray(a Attrib) {
	f._disableVertexAttribArray.Invoke(int(a))
}
func (f *Functions) Disable(cap Enum) {
	f._disable.Invoke(int(cap))
}
func (f *Functions) DrawArrays(mode Enum, first, count int) {
	f._drawArrays.Invoke(int(mode), first, count)
}
func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	f._drawElements.Invoke(int(mode), count, int(ty), offset)
}
func (f *Functions) DispatchCompute(x, y, z int) {
	panic("not implemented")
}
func (f *Functions) Enable(cap Enum) {
	f._enable.Invoke(int(cap))
}
func (f *Functions) EnableVertexAttribArray(a Attrib) {
	f._enableVertexAttribArray.Invoke(int(a))
}
func (f *Functions) EndQuery(target Enum) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		f._endQuery.Invoke(int(target))
	} else {
		f.EXT_disjoint_timer_query.Call("endQueryEXT", int(target))
	}
}
func (f *Functions) Finish() {
	f._finish.Invoke()
}
func (f *Functions) Flush() {
	f._flush.Invoke()
}
func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	f._framebufferRenderbuffer.Invoke(int(target), int(attachment), int(renderbuffertarget), js.Value(renderbuffer))
}
func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	f._framebufferTexture2D.Invoke(int(target), int(attachment), int(texTarget), js.Value(t), level)
}
func (f *Functions) GetError() Enum {
	// Avoid slow getError calls. See gio#179.
	return 0
}
func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	return paramVal(f._getRenderbufferParameteri.Invoke(int(pname)))
}
func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	if !f.isWebGL2 && pname == FRAMEBUFFER_ATTACHMENT_COLOR_ENCODING {
		// FRAMEBUFFER_ATTACHMENT_COLOR_ENCODING is only available on WebGL 2
		return LINEAR
	}
	return paramVal(f._getFramebufferAttachmentParameter.Invoke(int(target), int(attachment), int(pname)))
}
func (f *Functions) GetBinding(pname Enum) Object {
	obj := f._getParameter.Invoke(int(pname))
	if !obj.Truthy() {
		return Object{}
	}
	return Object(obj)
}
func (f *Functions) GetBindingi(pname Enum, idx int) Object {
	obj := f._getIndexedParameter.Invoke(int(pname), idx)
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
	return paramVal(f._getParameter.Invoke(int(pname)))
}
func (f *Functions) GetFloat(pname Enum) float32 {
	return float32(f._getParameter.Invoke(int(pname)).Float())
}
func (f *Functions) GetInteger4(pname Enum) [4]int {
	arr := f._getParameter.Invoke(int(pname))
	var res [4]int
	for i := range res {
		res[i] = arr.Index(i).Int()
	}
	return res
}
func (f *Functions) GetFloat4(pname Enum) [4]float32 {
	arr := f._getParameter.Invoke(int(pname))
	var res [4]float32
	for i := range res {
		res[i] = float32(arr.Index(i).Float())
	}
	return res
}
func (f *Functions) GetProgrami(p Program, pname Enum) int {
	return paramVal(f._getProgramParameter.Invoke(js.Value(p), int(pname)))
}
func (f *Functions) GetProgramInfoLog(p Program) string {
	return f._getProgramInfoLog.Invoke(js.Value(p)).String()
}
func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		return uint(paramVal(f._getQueryParameter.Invoke(js.Value(query), int(pname))))
	} else {
		return uint(paramVal(f.EXT_disjoint_timer_query.Call("getQueryObjectEXT", js.Value(query), int(pname))))
	}
}
func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	return paramVal(f._getShaderParameter.Invoke(js.Value(s), int(pname)))
}
func (f *Functions) GetShaderInfoLog(s Shader) string {
	return f._getShaderInfoLog.Invoke(js.Value(s)).String()
}
func (f *Functions) GetString(pname Enum) string {
	switch pname {
	case EXTENSIONS:
		extsjs := f._getSupportedExtensions.Invoke()
		var exts []string
		for i := 0; i < extsjs.Length(); i++ {
			exts = append(exts, "GL_"+extsjs.Index(i).String())
		}
		return strings.Join(exts, " ")
	default:
		return f._getParameter.Invoke(int(pname)).String()
	}
}
func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	return uint(paramVal(f._getUniformBlockIndex.Invoke(js.Value(p), name)))
}
func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	return Uniform(f._getUniformLocation.Invoke(js.Value(p), name))
}
func (f *Functions) GetVertexAttrib(index int, pname Enum) int {
	return paramVal(f._getVertexAttrib.Invoke(index, int(pname)))
}
func (f *Functions) GetVertexAttribBinding(index int, pname Enum) Object {
	obj := f._getVertexAttrib.Invoke(index, int(pname))
	if !obj.Truthy() {
		return Object{}
	}
	return Object(obj)
}
func (f *Functions) GetVertexAttribPointer(index int, pname Enum) uintptr {
	return uintptr(f._getVertexAttribOffset.Invoke(index, int(pname)).Int())
}
func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	fn := f.Ctx.Get("invalidateFramebuffer")
	if !fn.IsUndefined() {
		if f.int32Buf.IsUndefined() {
			f.int32Buf = js.Global().Get("Int32Array").New(1)
		}
		f.int32Buf.SetIndex(0, int32(attachment))
		f._invalidateFramebuffer.Invoke(int(target), f.int32Buf)
	}
}
func (f *Functions) IsEnabled(cap Enum) bool {
	return f._isEnabled.Invoke(int(cap)).Truthy()
}
func (f *Functions) LinkProgram(p Program) {
	f._linkProgram.Invoke(js.Value(p))
}
func (f *Functions) PixelStorei(pname Enum, param int) {
	f._pixelStorei.Invoke(int(pname), param)
}
func (f *Functions) MemoryBarrier(barriers Enum) {
	panic("not implemented")
}
func (f *Functions) MapBufferRange(target Enum, offset, length int, access Enum) []byte {
	panic("not implemented")
}
func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	f._renderbufferStorage.Invoke(int(target), int(internalformat), width, height)
}
func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	ba := f.byteArrayOf(data)
	f._readPixels.Invoke(x, y, width, height, int(format), int(ty), ba)
	js.CopyBytesToGo(data, ba)
}
func (f *Functions) Scissor(x, y, width, height int32) {
	f._scissor.Invoke(x, y, width, height)
}
func (f *Functions) ShaderSource(s Shader, src string) {
	f._shaderSource.Invoke(js.Value(s), src)
}
func (f *Functions) TexImage2D(target Enum, level int, internalFormat Enum, width, height int, format, ty Enum) {
	f._texImage2D.Invoke(int(target), int(level), int(internalFormat), int(width), int(height), 0, int(format), int(ty), nil)
}
func (f *Functions) TexStorage2D(target Enum, levels int, internalFormat Enum, width, height int) {
	f._texStorage2D.Invoke(int(target), levels, int(internalFormat), width, height)
}
func (f *Functions) TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	f._texSubImage2D.Invoke(int(target), level, x, y, width, height, int(format), int(ty), f.byteArrayOf(data))
}
func (f *Functions) TexParameteri(target, pname Enum, param int) {
	f._texParameteri.Invoke(int(target), int(pname), int(param))
}
func (f *Functions) UniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint) {
	f._uniformBlockBinding.Invoke(js.Value(p), int(uniformBlockIndex), int(uniformBlockBinding))
}
func (f *Functions) Uniform1f(dst Uniform, v float32) {
	f._uniform1f.Invoke(js.Value(dst), v)
}
func (f *Functions) Uniform1i(dst Uniform, v int) {
	f._uniform1i.Invoke(js.Value(dst), v)
}
func (f *Functions) Uniform2f(dst Uniform, v0, v1 float32) {
	f._uniform2f.Invoke(js.Value(dst), v0, v1)
}
func (f *Functions) Uniform3f(dst Uniform, v0, v1, v2 float32) {
	f._uniform3f.Invoke(js.Value(dst), v0, v1, v2)
}
func (f *Functions) Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	f._uniform4f.Invoke(js.Value(dst), v0, v1, v2, v3)
}
func (f *Functions) UseProgram(p Program) {
	f._useProgram.Invoke(js.Value(p))
}
func (f *Functions) UnmapBuffer(target Enum) bool {
	panic("not implemented")
}
func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	f._vertexAttribPointer.Invoke(int(dst), size, int(ty), normalized, stride, offset)
}
func (f *Functions) Viewport(x, y, width, height int) {
	f._viewport.Invoke(x, y, width, height)
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
