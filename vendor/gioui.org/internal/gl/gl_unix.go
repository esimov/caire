// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin || linux || freebsd || openbsd
// +build darwin linux freebsd openbsd

package gl

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

/*
#cgo CFLAGS: -Werror
#cgo linux freebsd LDFLAGS: -ldl

#include <stdint.h>
#include <stdlib.h>
#include <sys/types.h>
#define __USE_GNU
#include <dlfcn.h>

typedef unsigned int GLenum;
typedef unsigned int GLuint;
typedef char GLchar;
typedef float GLfloat;
typedef ssize_t GLsizeiptr;
typedef intptr_t GLintptr;
typedef unsigned int GLbitfield;
typedef int GLint;
typedef unsigned char GLboolean;
typedef int GLsizei;
typedef uint8_t GLubyte;

typedef void (*_glActiveTexture)(GLenum texture);
typedef void (*_glAttachShader)(GLuint program, GLuint shader);
typedef void (*_glBindAttribLocation)(GLuint program, GLuint index, const GLchar *name);
typedef void (*_glBindBuffer)(GLenum target, GLuint buffer);
typedef void (*_glBindFramebuffer)(GLenum target, GLuint framebuffer);
typedef void (*_glBindRenderbuffer)(GLenum target, GLuint renderbuffer);
typedef void (*_glBindTexture)(GLenum target, GLuint texture);
typedef void (*_glBlendEquation)(GLenum mode);
typedef void (*_glBlendFuncSeparate)(GLenum srcRGB, GLenum dstRGB, GLenum srcA, GLenum dstA);
typedef void (*_glBufferData)(GLenum target, GLsizeiptr size, const void *data, GLenum usage);
typedef void (*_glBufferSubData)(GLenum target, GLintptr offset, GLsizeiptr size, const void *data);
typedef GLenum (*_glCheckFramebufferStatus)(GLenum target);
typedef void (*_glClear)(GLbitfield mask);
typedef void (*_glClearColor)(GLfloat red, GLfloat green, GLfloat blue, GLfloat alpha);
typedef void (*_glClearDepthf)(GLfloat d);
typedef void (*_glCompileShader)(GLuint shader);
typedef void (*_glCopyTexSubImage2D)(GLenum target, GLint level, GLint xoffset, GLint yoffset, GLint x, GLint y, GLsizei width, GLsizei height);
typedef GLuint (*_glCreateProgram)(void);
typedef GLuint (*_glCreateShader)(GLenum type);
typedef void (*_glDeleteBuffers)(GLsizei n, const GLuint *buffers);
typedef void (*_glDeleteFramebuffers)(GLsizei n, const GLuint *framebuffers);
typedef void (*_glDeleteProgram)(GLuint program);
typedef void (*_glDeleteRenderbuffers)(GLsizei n, const GLuint *renderbuffers);
typedef void (*_glDeleteShader)(GLuint shader);
typedef void (*_glDeleteTextures)(GLsizei n, const GLuint *textures);
typedef void (*_glDepthFunc)(GLenum func);
typedef void (*_glDepthMask)(GLboolean flag);
typedef void (*_glDisable)(GLenum cap);
typedef void (*_glDisableVertexAttribArray)(GLuint index);
typedef void (*_glDrawArrays)(GLenum mode, GLint first, GLsizei count);
typedef void (*_glDrawElements)(GLenum mode, GLsizei count, GLenum type, const void *indices);
typedef void (*_glEnable)(GLenum cap);
typedef void (*_glEnableVertexAttribArray)(GLuint index);
typedef void (*_glFinish)(void);
typedef void (*_glFlush)(void);
typedef void (*_glFramebufferRenderbuffer)(GLenum target, GLenum attachment, GLenum renderbuffertarget, GLuint renderbuffer);
typedef void (*_glFramebufferTexture2D)(GLenum target, GLenum attachment, GLenum textarget, GLuint texture, GLint level);
typedef void (*_glGenBuffers)(GLsizei n, GLuint *buffers);
typedef void (*_glGenerateMipmap)(GLenum target);
typedef void (*_glGenFramebuffers)(GLsizei n, GLuint *framebuffers);
typedef void (*_glGenRenderbuffers)(GLsizei n, GLuint *renderbuffers);
typedef void (*_glGenTextures)(GLsizei n, GLuint *textures);
typedef GLenum (*_glGetError)(void);
typedef void (*_glGetFramebufferAttachmentParameteriv)(GLenum target, GLenum attachment, GLenum pname, GLint *params);
typedef void (*_glGetFloatv)(GLenum pname, GLfloat *data);
typedef void (*_glGetIntegerv)(GLenum pname, GLint *data);
typedef void (*_glGetIntegeri_v)(GLenum pname, GLuint idx, GLint *data);
typedef void (*_glGetProgramiv)(GLuint program, GLenum pname, GLint *params);
typedef void (*_glGetProgramInfoLog)(GLuint program, GLsizei bufSize, GLsizei *length, GLchar *infoLog);
typedef void (*_glGetRenderbufferParameteriv)(GLenum target, GLenum pname, GLint *params);
typedef void (*_glGetShaderiv)(GLuint shader, GLenum pname, GLint *params);
typedef void (*_glGetShaderInfoLog)(GLuint shader, GLsizei bufSize, GLsizei *length, GLchar *infoLog);
typedef const GLubyte *(*_glGetString)(GLenum name);
typedef GLint (*_glGetUniformLocation)(GLuint program, const GLchar *name);
typedef void (*_glGetVertexAttribiv)(GLuint index, GLenum pname, GLint *params);
typedef void (*_glGetVertexAttribPointerv)(GLuint index, GLenum pname, void **params);
typedef GLboolean (*_glIsEnabled)(GLenum cap);
typedef void (*_glLinkProgram)(GLuint program);
typedef void (*_glPixelStorei)(GLenum pname, GLint param);
typedef void (*_glReadPixels)(GLint x, GLint y, GLsizei width, GLsizei height, GLenum format, GLenum type, void *pixels);
typedef void (*_glRenderbufferStorage)(GLenum target, GLenum internalformat, GLsizei width, GLsizei height);
typedef void (*_glScissor)(GLint x, GLint y, GLsizei width, GLsizei height);
typedef void (*_glShaderSource)(GLuint shader, GLsizei count, const GLchar *const*string, const GLint *length);
typedef void (*_glTexImage2D)(GLenum target, GLint level, GLint internalformat, GLsizei width, GLsizei height, GLint border, GLenum format, GLenum type, const void *pixels);
typedef void (*_glTexParameteri)(GLenum target, GLenum pname, GLint param);
typedef void (*_glTexSubImage2D)(GLenum target, GLint level, GLint xoffset, GLint yoffset, GLsizei width, GLsizei height, GLenum format, GLenum type, const void *pixels);
typedef void (*_glUniform1f)(GLint location, GLfloat v0);
typedef void (*_glUniform1i)(GLint location, GLint v0);
typedef void (*_glUniform2f)(GLint location, GLfloat v0, GLfloat v1);
typedef void (*_glUniform3f)(GLint location, GLfloat v0, GLfloat v1, GLfloat v2);
typedef void (*_glUniform4f)(GLint location, GLfloat v0, GLfloat v1, GLfloat v2, GLfloat v3);
typedef void (*_glUseProgram)(GLuint program);
typedef void (*_glVertexAttribPointer)(GLuint index, GLint size, GLenum type, GLboolean normalized, GLsizei stride, const void *pointer);
typedef void (*_glViewport)(GLint x, GLint y, GLsizei width, GLsizei height);
typedef void (*_glBindVertexArray)(GLuint array);
typedef void (*_glBindBufferBase)(GLenum target, GLuint index, GLuint buffer);
typedef GLuint (*_glGetUniformBlockIndex)(GLuint program, const GLchar *uniformBlockName);
typedef void (*_glUniformBlockBinding)(GLuint program, GLuint uniformBlockIndex, GLuint uniformBlockBinding);
typedef void (*_glInvalidateFramebuffer)(GLenum target, GLsizei numAttachments, const GLenum *attachments);
typedef void (*_glBeginQuery)(GLenum target, GLuint id);
typedef void (*_glDeleteQueries)(GLsizei n, const GLuint *ids);
typedef void (*_glDeleteVertexArrays)(GLsizei n, const GLuint *ids);
typedef void (*_glEndQuery)(GLenum target);
typedef void (*_glGenQueries)(GLsizei n, GLuint *ids);
typedef void (*_glGenVertexArrays)(GLsizei n, GLuint *ids);
typedef void (*_glGetProgramBinary)(GLuint program, GLsizei bufsize, GLsizei *length, GLenum *binaryFormat, void *binary);
typedef void (*_glGetQueryObjectuiv)(GLuint id, GLenum pname, GLuint *params);
typedef const GLubyte* (*_glGetStringi)(GLenum name, GLuint index);
typedef void (*_glDispatchCompute)(GLuint x, GLuint y, GLuint z);
typedef void (*_glMemoryBarrier)(GLbitfield barriers);
typedef void* (*_glMapBufferRange)(GLenum target, GLintptr offset, GLsizeiptr length, GLbitfield access);
typedef GLboolean (*_glUnmapBuffer)(GLenum target);
typedef void (*_glBindImageTexture)(GLuint unit, GLuint texture, GLint level, GLboolean layered, GLint layer, GLenum access, GLenum format);
typedef void (*_glTexStorage2D)(GLenum target, GLsizei levels, GLenum internalformat, GLsizei width, GLsizei height);
typedef void (*_glBlitFramebuffer)(GLint srcX0, GLint srcY0, GLint srcX1, GLint srcY1, GLint dstX0, GLint dstY0, GLint dstX1, GLint dstY1, GLbitfield mask, GLenum filter);

static void glActiveTexture(_glActiveTexture f, GLenum texture) {
	f(texture);
}

static void glAttachShader(_glAttachShader f, GLuint program, GLuint shader) {
	f(program, shader);
}

static void glBindAttribLocation(_glBindAttribLocation f, GLuint program, GLuint index, const GLchar *name) {
	f(program, index, name);
}

static void glBindBuffer(_glBindBuffer f, GLenum target, GLuint buffer) {
	f(target, buffer);
}

static void glBindFramebuffer(_glBindFramebuffer f, GLenum target, GLuint framebuffer) {
	f(target, framebuffer);
}

static void glBindRenderbuffer(_glBindRenderbuffer f, GLenum target, GLuint renderbuffer) {
	f(target, renderbuffer);
}

static void glBindTexture(_glBindTexture f, GLenum target, GLuint texture) {
	f(target, texture);
}

static void glBindVertexArray(_glBindVertexArray f, GLuint array) {
	f(array);
}

static void glBlendEquation(_glBlendEquation f, GLenum mode) {
	f(mode);
}

static void glBlendFuncSeparate(_glBlendFuncSeparate f, GLenum srcRGB, GLenum dstRGB, GLenum srcA, GLenum dstA) {
	f(srcRGB, dstRGB, srcA, dstA);
}

static void glBufferData(_glBufferData f, GLenum target, GLsizeiptr size, const void *data, GLenum usage) {
	f(target, size, data, usage);
}

static void glBufferSubData(_glBufferSubData f, GLenum target, GLintptr offset, GLsizeiptr size, const void *data) {
	f(target, offset, size, data);
}

static GLenum glCheckFramebufferStatus(_glCheckFramebufferStatus f, GLenum target) {
	return f(target);
}

static void glClear(_glClear f, GLbitfield mask) {
	f(mask);
}

static void glClearColor(_glClearColor f, GLfloat red, GLfloat green, GLfloat blue, GLfloat alpha) {
	f(red, green, blue, alpha);
}

static void glClearDepthf(_glClearDepthf f, GLfloat d) {
	f(d);
}

static void glCompileShader(_glCompileShader f, GLuint shader) {
	f(shader);
}

static void glCopyTexSubImage2D(_glCopyTexSubImage2D f, GLenum target, GLint level, GLint xoffset, GLint yoffset, GLint x, GLint y, GLsizei width, GLsizei height) {
	f(target, level, xoffset, yoffset, x, y, width, height);
}

static GLuint glCreateProgram(_glCreateProgram f) {
	return f();
}

static GLuint glCreateShader(_glCreateShader f, GLenum type) {
	return f(type);
}

static void glDeleteBuffers(_glDeleteBuffers f, GLsizei n, const GLuint *buffers) {
	f(n, buffers);
}

static void glDeleteFramebuffers(_glDeleteFramebuffers f, GLsizei n, const GLuint *framebuffers) {
	f(n, framebuffers);
}

static void glDeleteProgram(_glDeleteProgram f, GLuint program) {
	f(program);
}

static void glDeleteRenderbuffers(_glDeleteRenderbuffers f, GLsizei n, const GLuint *renderbuffers) {
	f(n, renderbuffers);
}

static void glDeleteShader(_glDeleteShader f, GLuint shader) {
	f(shader);
}

static void glDeleteTextures(_glDeleteTextures f, GLsizei n, const GLuint *textures) {
	f(n, textures);
}

static void glDepthFunc(_glDepthFunc f, GLenum func) {
	f(func);
}

static void glDepthMask(_glDepthMask f, GLboolean flag) {
	f(flag);
}

static void glDisable(_glDisable f, GLenum cap) {
	f(cap);
}

static void glDisableVertexAttribArray(_glDisableVertexAttribArray f, GLuint index) {
	f(index);
}

static void glDrawArrays(_glDrawArrays f, GLenum mode, GLint first, GLsizei count) {
	f(mode, first, count);
}

// offset is defined as an uintptr_t to omit Cgo pointer checks.
static void glDrawElements(_glDrawElements f, GLenum mode, GLsizei count, GLenum type, const uintptr_t offset) {
	f(mode, count, type, (const void *)offset);
}

static void glEnable(_glEnable f, GLenum cap) {
	f(cap);
}

static void glEnableVertexAttribArray(_glEnableVertexAttribArray f, GLuint index) {
	f(index);
}

static void glFinish(_glFinish f) {
	f();
}

static void glFlush(_glFlush f) {
	f();
}

static void glFramebufferRenderbuffer(_glFramebufferRenderbuffer f, GLenum target, GLenum attachment, GLenum renderbuffertarget, GLuint renderbuffer) {
	f(target, attachment, renderbuffertarget, renderbuffer);
}

static void glFramebufferTexture2D(_glFramebufferTexture2D f, GLenum target, GLenum attachment, GLenum textarget, GLuint texture, GLint level) {
	f(target, attachment, textarget, texture, level);
}

static void glGenBuffers(_glGenBuffers f, GLsizei n, GLuint *buffers) {
	f(n, buffers);
}

static void glGenerateMipmap(_glGenerateMipmap f, GLenum target) {
	f(target);
}

static void glGenFramebuffers(_glGenFramebuffers f, GLsizei n, GLuint *framebuffers) {
	f(n, framebuffers);
}

static void glGenRenderbuffers(_glGenRenderbuffers f, GLsizei n, GLuint *renderbuffers) {
	f(n, renderbuffers);
}

static void glGenTextures(_glGenTextures f, GLsizei n, GLuint *textures) {
	f(n, textures);
}

static GLenum glGetError(_glGetError f) {
	return f();
}

static void glGetFramebufferAttachmentParameteriv(_glGetFramebufferAttachmentParameteriv f, GLenum target, GLenum attachment, GLenum pname, GLint *params) {
	f(target, attachment, pname, params);
}

static void glGetIntegerv(_glGetIntegerv f, GLenum pname, GLint *data) {
	f(pname, data);
}

static void glGetFloatv(_glGetFloatv f, GLenum pname, GLfloat *data) {
	f(pname, data);
}

static void glGetIntegeri_v(_glGetIntegeri_v f, GLenum pname, GLuint idx, GLint *data) {
	f(pname, idx, data);
}

static void glGetProgramiv(_glGetProgramiv f, GLuint program, GLenum pname, GLint *params) {
	f(program, pname, params);
}

static void glGetProgramInfoLog(_glGetProgramInfoLog f, GLuint program, GLsizei bufSize, GLsizei *length, GLchar *infoLog) {
	f(program, bufSize, length, infoLog);
}

static void glGetRenderbufferParameteriv(_glGetRenderbufferParameteriv f, GLenum target, GLenum pname, GLint *params) {
	f(target, pname, params);
}

static void glGetShaderiv(_glGetShaderiv f, GLuint shader, GLenum pname, GLint *params) {
	f(shader, pname, params);
}

static void glGetShaderInfoLog(_glGetShaderInfoLog f, GLuint shader, GLsizei bufSize, GLsizei *length, GLchar *infoLog) {
	f(shader, bufSize, length, infoLog);
}

static const GLubyte *glGetString(_glGetString f, GLenum name) {
	return f(name);
}

static GLint glGetUniformLocation(_glGetUniformLocation f, GLuint program, const GLchar *name) {
	return f(program, name);
}

static void glGetVertexAttribiv(_glGetVertexAttribiv f, GLuint index, GLenum pname, GLint *data) {
	f(index, pname, data);
}

// Return uintptr_t to avoid Cgo pointer check.
static uintptr_t glGetVertexAttribPointerv(_glGetVertexAttribPointerv f, GLuint index, GLenum pname) {
	void *ptrs;
	f(index, pname, &ptrs);
	return (uintptr_t)ptrs;
}

static GLboolean glIsEnabled(_glIsEnabled f, GLenum cap) {
	return f(cap);
}

static void glLinkProgram(_glLinkProgram f, GLuint program) {
	f(program);
}

static void glPixelStorei(_glPixelStorei f, GLenum pname, GLint param) {
	f(pname, param);
}

static void glReadPixels(_glReadPixels f, GLint x, GLint y, GLsizei width, GLsizei height, GLenum format, GLenum type, void *pixels) {
	f(x, y, width, height, format, type, pixels);
}

static void glRenderbufferStorage(_glRenderbufferStorage f, GLenum target, GLenum internalformat, GLsizei width, GLsizei height) {
	f(target, internalformat, width, height);
}

static void glScissor(_glScissor f, GLint x, GLint y, GLsizei width, GLsizei height) {
	f(x, y, width, height);
}

static void glShaderSource(_glShaderSource f, GLuint shader, GLsizei count, const GLchar *const*string, const GLint *length) {
	f(shader, count, string, length);
}

static void glTexImage2D(_glTexImage2D f, GLenum target, GLint level, GLint internalformat, GLsizei width, GLsizei height, GLint border, GLenum format, GLenum type, const void *pixels) {
	f(target, level, internalformat, width, height, border, format, type, pixels);
}

static void glTexParameteri(_glTexParameteri f, GLenum target, GLenum pname, GLint param) {
	f(target, pname, param);
}

static void glTexSubImage2D(_glTexSubImage2D f, GLenum target, GLint level, GLint xoffset, GLint yoffset, GLsizei width, GLsizei height, GLenum format, GLenum type, const void *pixels) {
	f(target, level, xoffset, yoffset, width, height, format, type, pixels);
}

static void glUniform1f(_glUniform1f f, GLint location, GLfloat v0) {
	f(location, v0);
}

static void glUniform1i(_glUniform1i f, GLint location, GLint v0) {
	f(location, v0);
}

static void glUniform2f(_glUniform2f f, GLint location, GLfloat v0, GLfloat v1) {
	f(location, v0, v1);
}

static void glUniform3f(_glUniform3f f, GLint location, GLfloat v0, GLfloat v1, GLfloat v2) {
	f(location, v0, v1, v2);
}

static void glUniform4f(_glUniform4f f, GLint location, GLfloat v0, GLfloat v1, GLfloat v2, GLfloat v3) {
	f(location, v0, v1, v2, v3);
}

static void glUseProgram(_glUseProgram f, GLuint program) {
	f(program);
}

// offset is defined as an uintptr_t to omit Cgo pointer checks.
static void glVertexAttribPointer(_glVertexAttribPointer f, GLuint index, GLint size, GLenum type, GLboolean normalized, GLsizei stride, uintptr_t offset) {
	f(index, size, type, normalized, stride, (const void *)offset);
}

static void glViewport(_glViewport f, GLint x, GLint y, GLsizei width, GLsizei height) {
	f(x, y, width, height);
}

static void glBindBufferBase(_glBindBufferBase f, GLenum target, GLuint index, GLuint buffer) {
	f(target, index, buffer);
}

static void glUniformBlockBinding(_glUniformBlockBinding f, GLuint program, GLuint uniformBlockIndex, GLuint uniformBlockBinding) {
	f(program, uniformBlockIndex, uniformBlockBinding);
}

static GLuint glGetUniformBlockIndex(_glGetUniformBlockIndex f, GLuint program, const GLchar *uniformBlockName) {
	return f(program, uniformBlockName);
}

static void glInvalidateFramebuffer(_glInvalidateFramebuffer f, GLenum target, GLenum attachment) {
	// Framebuffer invalidation is just a hint and can safely be ignored.
	if (f != NULL) {
		f(target, 1, &attachment);
	}
}

static void glBeginQuery(_glBeginQuery f, GLenum target, GLenum attachment) {
	f(target, attachment);
}

static void glDeleteQueries(_glDeleteQueries f, GLsizei n, const GLuint *ids) {
	f(n, ids);
}

static void glDeleteVertexArrays(_glDeleteVertexArrays f, GLsizei n, const GLuint *ids) {
	f(n, ids);
}

static void glEndQuery(_glEndQuery f, GLenum target) {
	f(target);
}

static const GLubyte* glGetStringi(_glGetStringi f, GLenum name, GLuint index) {
	return f(name, index);
}

static void glGenQueries(_glGenQueries f, GLsizei n, GLuint *ids) {
	f(n, ids);
}

static void glGenVertexArrays(_glGenVertexArrays f, GLsizei n, GLuint *ids) {
	f(n, ids);
}

static void glGetProgramBinary(_glGetProgramBinary f, GLuint program, GLsizei bufsize, GLsizei *length, GLenum *binaryFormat, void *binary) {
	f(program, bufsize, length, binaryFormat, binary);
}

static void glGetQueryObjectuiv(_glGetQueryObjectuiv f, GLuint id, GLenum pname, GLuint *params) {
	f(id, pname, params);
}

static void glMemoryBarrier(_glMemoryBarrier f, GLbitfield barriers) {
	f(barriers);
}

static void glDispatchCompute(_glDispatchCompute f, GLuint x, GLuint y, GLuint z) {
	f(x, y, z);
}

static void *glMapBufferRange(_glMapBufferRange f, GLenum target, GLintptr offset, GLsizeiptr length, GLbitfield access) {
	return f(target, offset, length, access);
}

static GLboolean glUnmapBuffer(_glUnmapBuffer f, GLenum target) {
	return f(target);
}

static void glBindImageTexture(_glBindImageTexture f, GLuint unit, GLuint texture, GLint level, GLboolean layered, GLint layer, GLenum access, GLenum format) {
	f(unit, texture, level, layered, layer, access, format);
}

static void glTexStorage2D(_glTexStorage2D f, GLenum target, GLsizei levels, GLenum internalFormat, GLsizei width, GLsizei height) {
	f(target, levels, internalFormat, width, height);
}

static void glBlitFramebuffer(_glBlitFramebuffer f, GLint srcX0, GLint srcY0, GLint srcX1, GLint srcY1, GLint dstX0, GLint dstY0, GLint dstX1, GLint dstY1, GLbitfield mask, GLenum filter) {
	f(srcX0, srcY0, srcX1, srcY1, dstX0, dstY0, dstX1, dstY1, mask, filter);
}
*/
import "C"

type Context interface{}

type Functions struct {
	// Query caches.
	uints  [100]C.GLuint
	ints   [100]C.GLint
	floats [100]C.GLfloat

	glActiveTexture                       C._glActiveTexture
	glAttachShader                        C._glAttachShader
	glBindAttribLocation                  C._glBindAttribLocation
	glBindBuffer                          C._glBindBuffer
	glBindFramebuffer                     C._glBindFramebuffer
	glBindRenderbuffer                    C._glBindRenderbuffer
	glBindTexture                         C._glBindTexture
	glBlendEquation                       C._glBlendEquation
	glBlendFuncSeparate                   C._glBlendFuncSeparate
	glBufferData                          C._glBufferData
	glBufferSubData                       C._glBufferSubData
	glCheckFramebufferStatus              C._glCheckFramebufferStatus
	glClear                               C._glClear
	glClearColor                          C._glClearColor
	glClearDepthf                         C._glClearDepthf
	glCompileShader                       C._glCompileShader
	glCopyTexSubImage2D                   C._glCopyTexSubImage2D
	glCreateProgram                       C._glCreateProgram
	glCreateShader                        C._glCreateShader
	glDeleteBuffers                       C._glDeleteBuffers
	glDeleteFramebuffers                  C._glDeleteFramebuffers
	glDeleteProgram                       C._glDeleteProgram
	glDeleteRenderbuffers                 C._glDeleteRenderbuffers
	glDeleteShader                        C._glDeleteShader
	glDeleteTextures                      C._glDeleteTextures
	glDepthFunc                           C._glDepthFunc
	glDepthMask                           C._glDepthMask
	glDisable                             C._glDisable
	glDisableVertexAttribArray            C._glDisableVertexAttribArray
	glDrawArrays                          C._glDrawArrays
	glDrawElements                        C._glDrawElements
	glEnable                              C._glEnable
	glEnableVertexAttribArray             C._glEnableVertexAttribArray
	glFinish                              C._glFinish
	glFlush                               C._glFlush
	glFramebufferRenderbuffer             C._glFramebufferRenderbuffer
	glFramebufferTexture2D                C._glFramebufferTexture2D
	glGenBuffers                          C._glGenBuffers
	glGenerateMipmap                      C._glGenerateMipmap
	glGenFramebuffers                     C._glGenFramebuffers
	glGenRenderbuffers                    C._glGenRenderbuffers
	glGenTextures                         C._glGenTextures
	glGetError                            C._glGetError
	glGetFramebufferAttachmentParameteriv C._glGetFramebufferAttachmentParameteriv
	glGetFloatv                           C._glGetFloatv
	glGetIntegerv                         C._glGetIntegerv
	glGetIntegeri_v                       C._glGetIntegeri_v
	glGetProgramiv                        C._glGetProgramiv
	glGetProgramInfoLog                   C._glGetProgramInfoLog
	glGetRenderbufferParameteriv          C._glGetRenderbufferParameteriv
	glGetShaderiv                         C._glGetShaderiv
	glGetShaderInfoLog                    C._glGetShaderInfoLog
	glGetString                           C._glGetString
	glGetUniformLocation                  C._glGetUniformLocation
	glGetVertexAttribiv                   C._glGetVertexAttribiv
	glGetVertexAttribPointerv             C._glGetVertexAttribPointerv
	glIsEnabled                           C._glIsEnabled
	glLinkProgram                         C._glLinkProgram
	glPixelStorei                         C._glPixelStorei
	glReadPixels                          C._glReadPixels
	glRenderbufferStorage                 C._glRenderbufferStorage
	glScissor                             C._glScissor
	glShaderSource                        C._glShaderSource
	glTexImage2D                          C._glTexImage2D
	glTexParameteri                       C._glTexParameteri
	glTexSubImage2D                       C._glTexSubImage2D
	glUniform1f                           C._glUniform1f
	glUniform1i                           C._glUniform1i
	glUniform2f                           C._glUniform2f
	glUniform3f                           C._glUniform3f
	glUniform4f                           C._glUniform4f
	glUseProgram                          C._glUseProgram
	glVertexAttribPointer                 C._glVertexAttribPointer
	glViewport                            C._glViewport
	glBindVertexArray                     C._glBindVertexArray
	glBindBufferBase                      C._glBindBufferBase
	glGetUniformBlockIndex                C._glGetUniformBlockIndex
	glUniformBlockBinding                 C._glUniformBlockBinding
	glInvalidateFramebuffer               C._glInvalidateFramebuffer
	glBeginQuery                          C._glBeginQuery
	glDeleteQueries                       C._glDeleteQueries
	glDeleteVertexArrays                  C._glDeleteVertexArrays
	glEndQuery                            C._glEndQuery
	glGenQueries                          C._glGenQueries
	glGenVertexArrays                     C._glGenVertexArrays
	glGetProgramBinary                    C._glGetProgramBinary
	glGetQueryObjectuiv                   C._glGetQueryObjectuiv
	glGetStringi                          C._glGetStringi
	glDispatchCompute                     C._glDispatchCompute
	glMemoryBarrier                       C._glMemoryBarrier
	glMapBufferRange                      C._glMapBufferRange
	glUnmapBuffer                         C._glUnmapBuffer
	glBindImageTexture                    C._glBindImageTexture
	glTexStorage2D                        C._glTexStorage2D
	glBlitFramebuffer                     C._glBlitFramebuffer
}

func NewFunctions(ctx Context, forceES bool) (*Functions, error) {
	if ctx != nil {
		panic("non-nil context")
	}
	f := new(Functions)
	err := f.load(forceES)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func dlsym(handle unsafe.Pointer, s string) unsafe.Pointer {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.dlsym(handle, cs)
}

func dlopen(lib string) unsafe.Pointer {
	clib := C.CString(lib)
	defer C.free(unsafe.Pointer(clib))
	return C.dlopen(clib, C.RTLD_NOW|C.RTLD_LOCAL)
}

func (f *Functions) load(forceES bool) error {
	var (
		loadErr  error
		libNames []string
		handles  []unsafe.Pointer
	)
	switch {
	case runtime.GOOS == "darwin" && !forceES:
		libNames = []string{"/System/Library/Frameworks/OpenGL.framework/OpenGL"}
	case runtime.GOOS == "darwin" && forceES:
		libNames = []string{"libGLESv2.dylib"}
	case runtime.GOOS == "ios":
		libNames = []string{"/System/Library/Frameworks/OpenGLES.framework/OpenGLES"}
	case runtime.GOOS == "android":
		libNames = []string{"libGLESv2.so", "libGLESv3.so"}
	default:
		libNames = []string{"libGLESv2.so.2"}
	}
	for _, lib := range libNames {
		if h := dlopen(lib); h != nil {
			handles = append(handles, h)
		}
	}
	if len(handles) == 0 {
		return fmt.Errorf("gl: no OpenGL implementation could be loaded (tried %q)", libNames)
	}
	load := func(s string) *[0]byte {
		for _, h := range handles {
			if f := dlsym(h, s); f != nil {
				return (*[0]byte)(f)
			}
		}
		return nil
	}
	must := func(s string) *[0]byte {
		ptr := load(s)
		if ptr == nil {
			loadErr = fmt.Errorf("gl: failed to load symbol %q", s)
		}
		return ptr
	}
	// GL ES 2.0 functions.
	f.glActiveTexture = must("glActiveTexture")
	f.glAttachShader = must("glAttachShader")
	f.glBindAttribLocation = must("glBindAttribLocation")
	f.glBindBuffer = must("glBindBuffer")
	f.glBindFramebuffer = must("glBindFramebuffer")
	f.glBindRenderbuffer = must("glBindRenderbuffer")
	f.glBindTexture = must("glBindTexture")
	f.glBlendEquation = must("glBlendEquation")
	f.glBlendFuncSeparate = must("glBlendFuncSeparate")
	f.glBufferData = must("glBufferData")
	f.glBufferSubData = must("glBufferSubData")
	f.glCheckFramebufferStatus = must("glCheckFramebufferStatus")
	f.glClear = must("glClear")
	f.glClearColor = must("glClearColor")
	f.glClearDepthf = must("glClearDepthf")
	f.glCompileShader = must("glCompileShader")
	f.glCopyTexSubImage2D = must("glCopyTexSubImage2D")
	f.glCreateProgram = must("glCreateProgram")
	f.glCreateShader = must("glCreateShader")
	f.glDeleteBuffers = must("glDeleteBuffers")
	f.glDeleteFramebuffers = must("glDeleteFramebuffers")
	f.glDeleteProgram = must("glDeleteProgram")
	f.glDeleteRenderbuffers = must("glDeleteRenderbuffers")
	f.glDeleteShader = must("glDeleteShader")
	f.glDeleteTextures = must("glDeleteTextures")
	f.glDepthFunc = must("glDepthFunc")
	f.glDepthMask = must("glDepthMask")
	f.glDisable = must("glDisable")
	f.glDisableVertexAttribArray = must("glDisableVertexAttribArray")
	f.glDrawArrays = must("glDrawArrays")
	f.glDrawElements = must("glDrawElements")
	f.glEnable = must("glEnable")
	f.glEnableVertexAttribArray = must("glEnableVertexAttribArray")
	f.glFinish = must("glFinish")
	f.glFlush = must("glFlush")
	f.glFramebufferRenderbuffer = must("glFramebufferRenderbuffer")
	f.glFramebufferTexture2D = must("glFramebufferTexture2D")
	f.glGenBuffers = must("glGenBuffers")
	f.glGenerateMipmap = must("glGenerateMipmap")
	f.glGenFramebuffers = must("glGenFramebuffers")
	f.glGenRenderbuffers = must("glGenRenderbuffers")
	f.glGenTextures = must("glGenTextures")
	f.glGetError = must("glGetError")
	f.glGetFramebufferAttachmentParameteriv = must("glGetFramebufferAttachmentParameteriv")
	f.glGetIntegerv = must("glGetIntegerv")
	f.glGetFloatv = must("glGetFloatv")
	f.glGetProgramiv = must("glGetProgramiv")
	f.glGetProgramInfoLog = must("glGetProgramInfoLog")
	f.glGetRenderbufferParameteriv = must("glGetRenderbufferParameteriv")
	f.glGetShaderiv = must("glGetShaderiv")
	f.glGetShaderInfoLog = must("glGetShaderInfoLog")
	f.glGetString = must("glGetString")
	f.glGetUniformLocation = must("glGetUniformLocation")
	f.glGetVertexAttribiv = must("glGetVertexAttribiv")
	f.glGetVertexAttribPointerv = must("glGetVertexAttribPointerv")
	f.glIsEnabled = must("glIsEnabled")
	f.glLinkProgram = must("glLinkProgram")
	f.glPixelStorei = must("glPixelStorei")
	f.glReadPixels = must("glReadPixels")
	f.glRenderbufferStorage = must("glRenderbufferStorage")
	f.glScissor = must("glScissor")
	f.glShaderSource = must("glShaderSource")
	f.glTexImage2D = must("glTexImage2D")
	f.glTexParameteri = must("glTexParameteri")
	f.glTexSubImage2D = must("glTexSubImage2D")
	f.glUniform1f = must("glUniform1f")
	f.glUniform1i = must("glUniform1i")
	f.glUniform2f = must("glUniform2f")
	f.glUniform3f = must("glUniform3f")
	f.glUniform4f = must("glUniform4f")
	f.glUseProgram = must("glUseProgram")
	f.glVertexAttribPointer = must("glVertexAttribPointer")
	f.glViewport = must("glViewport")

	// Extensions and GL ES 3 functions.
	f.glBindBufferBase = load("glBindBufferBase")
	f.glBindVertexArray = load("glBindVertexArray")
	f.glGetIntegeri_v = load("glGetIntegeri_v")
	f.glGetUniformBlockIndex = load("glGetUniformBlockIndex")
	f.glUniformBlockBinding = load("glUniformBlockBinding")
	f.glInvalidateFramebuffer = load("glInvalidateFramebuffer")
	f.glGetStringi = load("glGetStringi")
	// Fall back to EXT_invalidate_framebuffer if available.
	if f.glInvalidateFramebuffer == nil {
		f.glInvalidateFramebuffer = load("glDiscardFramebufferEXT")
	}

	f.glBeginQuery = load("glBeginQuery")
	if f.glBeginQuery == nil {
		f.glBeginQuery = load("glBeginQueryEXT")
	}
	f.glDeleteQueries = load("glDeleteQueries")
	if f.glDeleteQueries == nil {
		f.glDeleteQueries = load("glDeleteQueriesEXT")
	}
	f.glEndQuery = load("glEndQuery")
	if f.glEndQuery == nil {
		f.glEndQuery = load("glEndQueryEXT")
	}
	f.glGenQueries = load("glGenQueries")
	if f.glGenQueries == nil {
		f.glGenQueries = load("glGenQueriesEXT")
	}
	f.glGetQueryObjectuiv = load("glGetQueryObjectuiv")
	if f.glGetQueryObjectuiv == nil {
		f.glGetQueryObjectuiv = load("glGetQueryObjectuivEXT")
	}

	f.glDeleteVertexArrays = load("glDeleteVertexArrays")
	f.glGenVertexArrays = load("glGenVertexArrays")
	f.glMemoryBarrier = load("glMemoryBarrier")
	f.glDispatchCompute = load("glDispatchCompute")
	f.glMapBufferRange = load("glMapBufferRange")
	f.glUnmapBuffer = load("glUnmapBuffer")
	f.glBindImageTexture = load("glBindImageTexture")
	f.glTexStorage2D = load("glTexStorage2D")
	f.glBlitFramebuffer = load("glBlitFramebuffer")
	f.glGetProgramBinary = load("glGetProgramBinary")

	return loadErr
}

func (f *Functions) ActiveTexture(texture Enum) {
	C.glActiveTexture(f.glActiveTexture, C.GLenum(texture))
}

func (f *Functions) AttachShader(p Program, s Shader) {
	C.glAttachShader(f.glAttachShader, C.GLuint(p.V), C.GLuint(s.V))
}

func (f *Functions) BeginQuery(target Enum, query Query) {
	C.glBeginQuery(f.glBeginQuery, C.GLenum(target), C.GLenum(query.V))
}

func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.glBindAttribLocation(f.glBindAttribLocation, C.GLuint(p.V), C.GLuint(a), cname)
}

func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	C.glBindBufferBase(f.glBindBufferBase, C.GLenum(target), C.GLuint(index), C.GLuint(b.V))
}

func (f *Functions) BindBuffer(target Enum, b Buffer) {
	C.glBindBuffer(f.glBindBuffer, C.GLenum(target), C.GLuint(b.V))
}

func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	C.glBindFramebuffer(f.glBindFramebuffer, C.GLenum(target), C.GLuint(fb.V))
}

func (f *Functions) BindRenderbuffer(target Enum, fb Renderbuffer) {
	C.glBindRenderbuffer(f.glBindRenderbuffer, C.GLenum(target), C.GLuint(fb.V))
}

func (f *Functions) BindImageTexture(unit int, t Texture, level int, layered bool, layer int, access, format Enum) {
	l := C.GLboolean(FALSE)
	if layered {
		l = TRUE
	}
	C.glBindImageTexture(f.glBindImageTexture, C.GLuint(unit), C.GLuint(t.V), C.GLint(level), l, C.GLint(layer), C.GLenum(access), C.GLenum(format))
}

func (f *Functions) BindTexture(target Enum, t Texture) {
	C.glBindTexture(f.glBindTexture, C.GLenum(target), C.GLuint(t.V))
}

func (f *Functions) BindVertexArray(a VertexArray) {
	C.glBindVertexArray(f.glBindVertexArray, C.GLuint(a.V))
}

func (f *Functions) BlendEquation(mode Enum) {
	C.glBlendEquation(f.glBlendEquation, C.GLenum(mode))
}

func (f *Functions) BlendFuncSeparate(srcRGB, dstRGB, srcA, dstA Enum) {
	C.glBlendFuncSeparate(f.glBlendFuncSeparate, C.GLenum(srcRGB), C.GLenum(dstRGB), C.GLenum(srcA), C.GLenum(dstA))
}

func (f *Functions) BlitFramebuffer(sx0, sy0, sx1, sy1, dx0, dy0, dx1, dy1 int, mask Enum, filter Enum) {
	C.glBlitFramebuffer(f.glBlitFramebuffer,
		C.GLint(sx0), C.GLint(sy0), C.GLint(sx1), C.GLint(sy1),
		C.GLint(dx0), C.GLint(dy0), C.GLint(dx1), C.GLint(dy1),
		C.GLenum(mask), C.GLenum(filter),
	)
}

func (f *Functions) BufferData(target Enum, size int, usage Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glBufferData(f.glBufferData, C.GLenum(target), C.GLsizeiptr(size), p, C.GLenum(usage))
}

func (f *Functions) BufferSubData(target Enum, offset int, src []byte) {
	var p unsafe.Pointer
	if len(src) > 0 {
		p = unsafe.Pointer(&src[0])
	}
	C.glBufferSubData(f.glBufferSubData, C.GLenum(target), C.GLintptr(offset), C.GLsizeiptr(len(src)), p)
}

func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(C.glCheckFramebufferStatus(f.glCheckFramebufferStatus, C.GLenum(target)))
}

func (f *Functions) Clear(mask Enum) {
	C.glClear(f.glClear, C.GLbitfield(mask))
}

func (f *Functions) ClearColor(red float32, green float32, blue float32, alpha float32) {
	C.glClearColor(f.glClearColor, C.GLfloat(red), C.GLfloat(green), C.GLfloat(blue), C.GLfloat(alpha))
}

func (f *Functions) ClearDepthf(d float32) {
	C.glClearDepthf(f.glClearDepthf, C.GLfloat(d))
}

func (f *Functions) CompileShader(s Shader) {
	C.glCompileShader(f.glCompileShader, C.GLuint(s.V))
}

func (f *Functions) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	C.glCopyTexSubImage2D(f.glCopyTexSubImage2D, C.GLenum(target), C.GLint(level), C.GLint(xoffset), C.GLint(yoffset), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) CreateBuffer() Buffer {
	C.glGenBuffers(f.glGenBuffers, 1, &f.uints[0])
	return Buffer{uint(f.uints[0])}
}

func (f *Functions) CreateFramebuffer() Framebuffer {
	C.glGenFramebuffers(f.glGenFramebuffers, 1, &f.uints[0])
	return Framebuffer{uint(f.uints[0])}
}

func (f *Functions) CreateProgram() Program {
	return Program{uint(C.glCreateProgram(f.glCreateProgram))}
}

func (f *Functions) CreateQuery() Query {
	C.glGenQueries(f.glGenQueries, 1, &f.uints[0])
	return Query{uint(f.uints[0])}
}

func (f *Functions) CreateRenderbuffer() Renderbuffer {
	C.glGenRenderbuffers(f.glGenRenderbuffers, 1, &f.uints[0])
	return Renderbuffer{uint(f.uints[0])}
}

func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader{uint(C.glCreateShader(f.glCreateShader, C.GLenum(ty)))}
}

func (f *Functions) CreateTexture() Texture {
	C.glGenTextures(f.glGenTextures, 1, &f.uints[0])
	return Texture{uint(f.uints[0])}
}

func (f *Functions) CreateVertexArray() VertexArray {
	C.glGenVertexArrays(f.glGenVertexArrays, 1, &f.uints[0])
	return VertexArray{uint(f.uints[0])}
}

func (f *Functions) DeleteBuffer(v Buffer) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteBuffers(f.glDeleteBuffers, 1, &f.uints[0])
}

func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteFramebuffers(f.glDeleteFramebuffers, 1, &f.uints[0])
}

func (f *Functions) DeleteProgram(p Program) {
	C.glDeleteProgram(f.glDeleteProgram, C.GLuint(p.V))
}

func (f *Functions) DeleteQuery(query Query) {
	f.uints[0] = C.GLuint(query.V)
	C.glDeleteQueries(f.glDeleteQueries, 1, &f.uints[0])
}

func (f *Functions) DeleteVertexArray(array VertexArray) {
	f.uints[0] = C.GLuint(array.V)
	C.glDeleteVertexArrays(f.glDeleteVertexArrays, 1, &f.uints[0])
}

func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteRenderbuffers(f.glDeleteRenderbuffers, 1, &f.uints[0])
}

func (f *Functions) DeleteShader(s Shader) {
	C.glDeleteShader(f.glDeleteShader, C.GLuint(s.V))
}

func (f *Functions) DeleteTexture(v Texture) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteTextures(f.glDeleteTextures, 1, &f.uints[0])
}

func (f *Functions) DepthFunc(v Enum) {
	C.glDepthFunc(f.glDepthFunc, C.GLenum(v))
}

func (f *Functions) DepthMask(mask bool) {
	m := C.GLboolean(FALSE)
	if mask {
		m = C.GLboolean(TRUE)
	}
	C.glDepthMask(f.glDepthMask, m)
}

func (f *Functions) DisableVertexAttribArray(a Attrib) {
	C.glDisableVertexAttribArray(f.glDisableVertexAttribArray, C.GLuint(a))
}

func (f *Functions) Disable(cap Enum) {
	C.glDisable(f.glDisable, C.GLenum(cap))
}

func (f *Functions) DrawArrays(mode Enum, first int, count int) {
	C.glDrawArrays(f.glDrawArrays, C.GLenum(mode), C.GLint(first), C.GLsizei(count))
}

func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	C.glDrawElements(f.glDrawElements, C.GLenum(mode), C.GLsizei(count), C.GLenum(ty), C.uintptr_t(offset))
}

func (f *Functions) DispatchCompute(x, y, z int) {
	C.glDispatchCompute(f.glDispatchCompute, C.GLuint(x), C.GLuint(y), C.GLuint(z))
}

func (f *Functions) Enable(cap Enum) {
	C.glEnable(f.glEnable, C.GLenum(cap))
}

func (f *Functions) EndQuery(target Enum) {
	C.glEndQuery(f.glEndQuery, C.GLenum(target))
}

func (f *Functions) EnableVertexAttribArray(a Attrib) {
	C.glEnableVertexAttribArray(f.glEnableVertexAttribArray, C.GLuint(a))
}

func (f *Functions) Finish() {
	C.glFinish(f.glFinish)
}

func (f *Functions) Flush() {
	C.glFlush(f.glFinish)
}

func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	C.glFramebufferRenderbuffer(f.glFramebufferRenderbuffer, C.GLenum(target), C.GLenum(attachment), C.GLenum(renderbuffertarget), C.GLuint(renderbuffer.V))
}

func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	C.glFramebufferTexture2D(f.glFramebufferTexture2D, C.GLenum(target), C.GLenum(attachment), C.GLenum(texTarget), C.GLuint(t.V), C.GLint(level))
}

func (f *Functions) GenerateMipmap(target Enum) {
	C.glGenerateMipmap(f.glGenerateMipmap, C.GLenum(target))
}

func (c *Functions) GetBinding(pname Enum) Object {
	return Object{uint(c.GetInteger(pname))}
}

func (c *Functions) GetBindingi(pname Enum, idx int) Object {
	return Object{uint(c.GetIntegeri(pname, idx))}
}

func (f *Functions) GetError() Enum {
	return Enum(C.glGetError(f.glGetError))
}

func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	C.glGetRenderbufferParameteriv(f.glGetRenderbufferParameteriv, C.GLenum(target), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	C.glGetFramebufferAttachmentParameteriv(f.glGetFramebufferAttachmentParameteriv, C.GLenum(target), C.GLenum(attachment), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetFloat4(pname Enum) [4]float32 {
	C.glGetFloatv(f.glGetFloatv, C.GLenum(pname), &f.floats[0])
	var r [4]float32
	for i := range r {
		r[i] = float32(f.floats[i])
	}
	return r
}

func (f *Functions) GetFloat(pname Enum) float32 {
	C.glGetFloatv(f.glGetFloatv, C.GLenum(pname), &f.floats[0])
	return float32(f.floats[0])
}

func (f *Functions) GetInteger4(pname Enum) [4]int {
	C.glGetIntegerv(f.glGetIntegerv, C.GLenum(pname), &f.ints[0])
	var r [4]int
	for i := range r {
		r[i] = int(f.ints[i])
	}
	return r
}

func (f *Functions) GetInteger(pname Enum) int {
	C.glGetIntegerv(f.glGetIntegerv, C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetIntegeri(pname Enum, idx int) int {
	C.glGetIntegeri_v(f.glGetIntegeri_v, C.GLenum(pname), C.GLuint(idx), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetProgrami(p Program, pname Enum) int {
	C.glGetProgramiv(f.glGetProgramiv, C.GLuint(p.V), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetProgramBinary(p Program) []byte {
	sz := f.GetProgrami(p, PROGRAM_BINARY_LENGTH)
	if sz == 0 {
		return nil
	}
	buf := make([]byte, sz)
	var format C.GLenum
	C.glGetProgramBinary(f.glGetProgramBinary, C.GLuint(p.V), C.GLsizei(sz), nil, &format, unsafe.Pointer(&buf[0]))
	return buf
}

func (f *Functions) GetProgramInfoLog(p Program) string {
	n := f.GetProgrami(p, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	C.glGetProgramInfoLog(f.glGetProgramInfoLog, C.GLuint(p.V), C.GLsizei(len(buf)), nil, (*C.GLchar)(unsafe.Pointer(&buf[0])))
	return string(buf)
}

func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	C.glGetQueryObjectuiv(f.glGetQueryObjectuiv, C.GLuint(query.V), C.GLenum(pname), &f.uints[0])
	return uint(f.uints[0])
}

func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	C.glGetShaderiv(f.glGetShaderiv, C.GLuint(s.V), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetShaderInfoLog(s Shader) string {
	n := f.GetShaderi(s, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	C.glGetShaderInfoLog(f.glGetShaderInfoLog, C.GLuint(s.V), C.GLsizei(len(buf)), nil, (*C.GLchar)(unsafe.Pointer(&buf[0])))
	return string(buf)
}

func (f *Functions) getStringi(pname Enum, index int) string {
	str := C.glGetStringi(f.glGetStringi, C.GLenum(pname), C.GLuint(index))
	if str == nil {
		return ""
	}
	return C.GoString((*C.char)(unsafe.Pointer(str)))
}

func (f *Functions) GetString(pname Enum) string {
	switch {
	case runtime.GOOS == "darwin" && pname == EXTENSIONS:
		// macOS OpenGL 3 core profile doesn't support glGetString(GL_EXTENSIONS).
		// Use glGetStringi(GL_EXTENSIONS, <index>).
		var exts []string
		nexts := f.GetInteger(NUM_EXTENSIONS)
		for i := 0; i < nexts; i++ {
			ext := f.getStringi(EXTENSIONS, i)
			exts = append(exts, ext)
		}
		return strings.Join(exts, " ")
	default:
		str := C.glGetString(f.glGetString, C.GLenum(pname))
		return C.GoString((*C.char)(unsafe.Pointer(str)))
	}
}

func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return uint(C.glGetUniformBlockIndex(f.glGetUniformBlockIndex, C.GLuint(p.V), cname))
}

func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return Uniform{int(C.glGetUniformLocation(f.glGetUniformLocation, C.GLuint(p.V), cname))}
}

func (f *Functions) GetVertexAttrib(index int, pname Enum) int {
	C.glGetVertexAttribiv(f.glGetVertexAttribiv, C.GLuint(index), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetVertexAttribBinding(index int, pname Enum) Object {
	return Object{uint(f.GetVertexAttrib(index, pname))}
}

func (f *Functions) GetVertexAttribPointer(index int, pname Enum) uintptr {
	ptr := C.glGetVertexAttribPointerv(f.glGetVertexAttribPointerv, C.GLuint(index), C.GLenum(pname))
	return uintptr(ptr)
}

func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	C.glInvalidateFramebuffer(f.glInvalidateFramebuffer, C.GLenum(target), C.GLenum(attachment))
}

func (f *Functions) IsEnabled(cap Enum) bool {
	return C.glIsEnabled(f.glIsEnabled, C.GLenum(cap)) == TRUE
}

func (f *Functions) LinkProgram(p Program) {
	C.glLinkProgram(f.glLinkProgram, C.GLuint(p.V))
}

func (f *Functions) PixelStorei(pname Enum, param int) {
	C.glPixelStorei(f.glPixelStorei, C.GLenum(pname), C.GLint(param))
}

func (f *Functions) MemoryBarrier(barriers Enum) {
	C.glMemoryBarrier(f.glMemoryBarrier, C.GLbitfield(barriers))
}

func (f *Functions) MapBufferRange(target Enum, offset, length int, access Enum) []byte {
	p := C.glMapBufferRange(f.glMapBufferRange, C.GLenum(target), C.GLintptr(offset), C.GLsizeiptr(length), C.GLbitfield(access))
	if p == nil {
		return nil
	}
	return (*[1 << 30]byte)(p)[:length:length]
}

func (f *Functions) Scissor(x, y, width, height int32) {
	C.glScissor(f.glScissor, C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glReadPixels(f.glReadPixels, C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(ty), p)
}

func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	C.glRenderbufferStorage(f.glRenderbufferStorage, C.GLenum(target), C.GLenum(internalformat), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) ShaderSource(s Shader, src string) {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))
	strlen := C.GLint(len(src))
	C.glShaderSource(f.glShaderSource, C.GLuint(s.V), 1, &csrc, &strlen)
}

func (f *Functions) TexImage2D(target Enum, level int, internalFormat Enum, width int, height int, format Enum, ty Enum) {
	C.glTexImage2D(f.glTexImage2D, C.GLenum(target), C.GLint(level), C.GLint(internalFormat), C.GLsizei(width), C.GLsizei(height), 0, C.GLenum(format), C.GLenum(ty), nil)
}

func (f *Functions) TexStorage2D(target Enum, levels int, internalFormat Enum, width, height int) {
	C.glTexStorage2D(f.glTexStorage2D, C.GLenum(target), C.GLsizei(levels), C.GLenum(internalFormat), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) TexSubImage2D(target Enum, level int, x int, y int, width int, height int, format Enum, ty Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glTexSubImage2D(f.glTexSubImage2D, C.GLenum(target), C.GLint(level), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(ty), p)
}

func (f *Functions) TexParameteri(target, pname Enum, param int) {
	C.glTexParameteri(f.glTexParameteri, C.GLenum(target), C.GLenum(pname), C.GLint(param))
}

func (f *Functions) UniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint) {
	C.glUniformBlockBinding(f.glUniformBlockBinding, C.GLuint(p.V), C.GLuint(uniformBlockIndex), C.GLuint(uniformBlockBinding))
}

func (f *Functions) Uniform1f(dst Uniform, v float32) {
	C.glUniform1f(f.glUniform1f, C.GLint(dst.V), C.GLfloat(v))
}

func (f *Functions) Uniform1i(dst Uniform, v int) {
	C.glUniform1i(f.glUniform1i, C.GLint(dst.V), C.GLint(v))
}

func (f *Functions) Uniform2f(dst Uniform, v0 float32, v1 float32) {
	C.glUniform2f(f.glUniform2f, C.GLint(dst.V), C.GLfloat(v0), C.GLfloat(v1))
}

func (f *Functions) Uniform3f(dst Uniform, v0 float32, v1 float32, v2 float32) {
	C.glUniform3f(f.glUniform3f, C.GLint(dst.V), C.GLfloat(v0), C.GLfloat(v1), C.GLfloat(v2))
}

func (f *Functions) Uniform4f(dst Uniform, v0 float32, v1 float32, v2 float32, v3 float32) {
	C.glUniform4f(f.glUniform4f, C.GLint(dst.V), C.GLfloat(v0), C.GLfloat(v1), C.GLfloat(v2), C.GLfloat(v3))
}

func (f *Functions) UseProgram(p Program) {
	C.glUseProgram(f.glUseProgram, C.GLuint(p.V))
}

func (f *Functions) UnmapBuffer(target Enum) bool {
	r := C.glUnmapBuffer(f.glUnmapBuffer, C.GLenum(target))
	return r == TRUE
}

func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride int, offset int) {
	var n C.GLboolean = FALSE
	if normalized {
		n = TRUE
	}
	C.glVertexAttribPointer(f.glVertexAttribPointer, C.GLuint(dst), C.GLint(size), C.GLenum(ty), n, C.GLsizei(stride), C.uintptr_t(offset))
}

func (f *Functions) Viewport(x int, y int, width int, height int) {
	C.glViewport(f.glViewport, C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}
