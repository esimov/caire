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

typedef struct {
	void (*glActiveTexture)(GLenum texture);
	void (*glAttachShader)(GLuint program, GLuint shader);
	void (*glBindAttribLocation)(GLuint program, GLuint index, const GLchar *name);
	void (*glBindBuffer)(GLenum target, GLuint buffer);
	void (*glBindFramebuffer)(GLenum target, GLuint framebuffer);
	void (*glBindRenderbuffer)(GLenum target, GLuint renderbuffer);
	void (*glBindTexture)(GLenum target, GLuint texture);
	void (*glBlendEquation)(GLenum mode);
	void (*glBlendFuncSeparate)(GLenum srcRGB, GLenum dstRGB, GLenum srcA, GLenum dstA);
	void (*glBufferData)(GLenum target, GLsizeiptr size, const void *data, GLenum usage);
	void (*glBufferSubData)(GLenum target, GLintptr offset, GLsizeiptr size, const void *data);
	GLenum (*glCheckFramebufferStatus)(GLenum target);
	void (*glClear)(GLbitfield mask);
	void (*glClearColor)(GLfloat red, GLfloat green, GLfloat blue, GLfloat alpha);
	void (*glClearDepthf)(GLfloat d);
	void (*glCompileShader)(GLuint shader);
	void (*glCopyTexSubImage2D)(GLenum target, GLint level, GLint xoffset, GLint yoffset, GLint x, GLint y, GLsizei width, GLsizei height);
	GLuint (*glCreateProgram)(void);
	GLuint (*glCreateShader)(GLenum type);
	void (*glDeleteBuffers)(GLsizei n, const GLuint *buffers);
	void (*glDeleteFramebuffers)(GLsizei n, const GLuint *framebuffers);
	void (*glDeleteProgram)(GLuint program);
	void (*glDeleteRenderbuffers)(GLsizei n, const GLuint *renderbuffers);
	void (*glDeleteShader)(GLuint shader);
	void (*glDeleteTextures)(GLsizei n, const GLuint *textures);
	void (*glDepthFunc)(GLenum func);
	void (*glDepthMask)(GLboolean flag);
	void (*glDisable)(GLenum cap);
	void (*glDisableVertexAttribArray)(GLuint index);
	void (*glDrawArrays)(GLenum mode, GLint first, GLsizei count);
	void (*glDrawElements)(GLenum mode, GLsizei count, GLenum type, const void *indices);
	void (*glEnable)(GLenum cap);
	void (*glEnableVertexAttribArray)(GLuint index);
	void (*glFinish)(void);
	void (*glFlush)(void);
	void (*glFramebufferRenderbuffer)(GLenum target, GLenum attachment, GLenum renderbuffertarget, GLuint renderbuffer);
	void (*glFramebufferTexture2D)(GLenum target, GLenum attachment, GLenum textarget, GLuint texture, GLint level);
	void (*glGenBuffers)(GLsizei n, GLuint *buffers);
	void (*glGenFramebuffers)(GLsizei n, GLuint *framebuffers);
	void (*glGenRenderbuffers)(GLsizei n, GLuint *renderbuffers);
	void (*glGenTextures)(GLsizei n, GLuint *textures);
	GLenum (*glGetError)(void);
	void (*glGetFramebufferAttachmentParameteriv)(GLenum target, GLenum attachment, GLenum pname, GLint *params);
	void (*glGetFloatv)(GLenum pname, GLfloat *data);
	void (*glGetIntegerv)(GLenum pname, GLint *data);
	void (*glGetIntegeri_v)(GLenum pname, GLuint idx, GLint *data);
	void (*glGetProgramiv)(GLuint program, GLenum pname, GLint *params);
	void (*glGetProgramInfoLog)(GLuint program, GLsizei bufSize, GLsizei *length, GLchar *infoLog);
	void (*glGetRenderbufferParameteriv)(GLenum target, GLenum pname, GLint *params);
	void (*glGetShaderiv)(GLuint shader, GLenum pname, GLint *params);
	void (*glGetShaderInfoLog)(GLuint shader, GLsizei bufSize, GLsizei *length, GLchar *infoLog);
	const GLubyte *(*glGetString)(GLenum name);
	GLint (*glGetUniformLocation)(GLuint program, const GLchar *name);
	void (*glGetVertexAttribiv)(GLuint index, GLenum pname, GLint *params);
	void (*glGetVertexAttribPointerv)(GLuint index, GLenum pname, void **params);
	GLboolean (*glIsEnabled)(GLenum cap);
	void (*glLinkProgram)(GLuint program);
	void (*glPixelStorei)(GLenum pname, GLint param);
	void (*glReadPixels)(GLint x, GLint y, GLsizei width, GLsizei height, GLenum format, GLenum type, void *pixels);
	void (*glRenderbufferStorage)(GLenum target, GLenum internalformat, GLsizei width, GLsizei height);
	void (*glScissor)(GLint x, GLint y, GLsizei width, GLsizei height);
	void (*glShaderSource)(GLuint shader, GLsizei count, const GLchar *const*string, const GLint *length);
	void (*glTexImage2D)(GLenum target, GLint level, GLint internalformat, GLsizei width, GLsizei height, GLint border, GLenum format, GLenum type, const void *pixels);
	void (*glTexParameteri)(GLenum target, GLenum pname, GLint param);
	void (*glTexSubImage2D)(GLenum target, GLint level, GLint xoffset, GLint yoffset, GLsizei width, GLsizei height, GLenum format, GLenum type, const void *pixels);
	void (*glUniform1f)(GLint location, GLfloat v0);
	void (*glUniform1i)(GLint location, GLint v0);
	void (*glUniform2f)(GLint location, GLfloat v0, GLfloat v1);
	void (*glUniform3f)(GLint location, GLfloat v0, GLfloat v1, GLfloat v2);
	void (*glUniform4f)(GLint location, GLfloat v0, GLfloat v1, GLfloat v2, GLfloat v3);
	void (*glUseProgram)(GLuint program);
	void (*glVertexAttribPointer)(GLuint index, GLint size, GLenum type, GLboolean normalized, GLsizei stride, const void *pointer);
	void (*glViewport)(GLint x, GLint y, GLsizei width, GLsizei height);

	void (*glBindVertexArray)(GLuint array);
	void (*glBindBufferBase)(GLenum target, GLuint index, GLuint buffer);
	GLuint (*glGetUniformBlockIndex)(GLuint program, const GLchar *uniformBlockName);
	void (*glUniformBlockBinding)(GLuint program, GLuint uniformBlockIndex, GLuint uniformBlockBinding);
	void (*glInvalidateFramebuffer)(GLenum target, GLsizei numAttachments, const GLenum *attachments);
	void (*glBeginQuery)(GLenum target, GLuint id);
	void (*glDeleteQueries)(GLsizei n, const GLuint *ids);
	void (*glDeleteVertexArrays)(GLsizei n, const GLuint *ids);
	void (*glEndQuery)(GLenum target);
	void (*glGenQueries)(GLsizei n, GLuint *ids);
	void (*glGenVertexArrays)(GLsizei n, GLuint *ids);
	void (*glGetProgramBinary)(GLuint program, GLsizei bufsize, GLsizei *length, GLenum *binaryFormat, void *binary);
	void (*glGetQueryObjectuiv)(GLuint id, GLenum pname, GLuint *params);
	const GLubyte* (*glGetStringi)(GLenum name, GLuint index);
	void (*glDispatchCompute)(GLuint x, GLuint y, GLuint z);
	void (*glMemoryBarrier)(GLbitfield barriers);
	void* (*glMapBufferRange)(GLenum target, GLintptr offset, GLsizeiptr length, GLbitfield access);
	GLboolean (*glUnmapBuffer)(GLenum target);
	void (*glBindImageTexture)(GLuint unit, GLuint texture, GLint level, GLboolean layered, GLint layer, GLenum access, GLenum format);
	void (*glTexStorage2D)(GLenum target, GLsizei levels, GLenum internalformat, GLsizei width, GLsizei height);
	void (*glBlitFramebuffer)(GLint srcX0, GLint srcY0, GLint srcX1, GLint srcY1, GLint dstX0, GLint dstY0, GLint dstX1, GLint dstY1, GLbitfield mask, GLenum filter);
} glFunctions;

static void glActiveTexture(glFunctions *f, GLenum texture) {
	f->glActiveTexture(texture);
}

static void glAttachShader(glFunctions *f, GLuint program, GLuint shader) {
	f->glAttachShader(program, shader);
}

static void glBindAttribLocation(glFunctions *f, GLuint program, GLuint index, const GLchar *name) {
	f->glBindAttribLocation(program, index, name);
}

static void glBindBuffer(glFunctions *f, GLenum target, GLuint buffer) {
	f->glBindBuffer(target, buffer);
}

static void glBindFramebuffer(glFunctions *f, GLenum target, GLuint framebuffer) {
	f->glBindFramebuffer(target, framebuffer);
}

static void glBindRenderbuffer(glFunctions *f, GLenum target, GLuint renderbuffer) {
	f->glBindRenderbuffer(target, renderbuffer);
}

static void glBindTexture(glFunctions *f, GLenum target, GLuint texture) {
	f->glBindTexture(target, texture);
}

static void glBindVertexArray(glFunctions *f, GLuint array) {
	f->glBindVertexArray(array);
}

static void glBlendEquation(glFunctions *f, GLenum mode) {
	f->glBlendEquation(mode);
}

static void glBlendFuncSeparate(glFunctions *f, GLenum srcRGB, GLenum dstRGB, GLenum srcA, GLenum dstA) {
	f->glBlendFuncSeparate(srcRGB, dstRGB, srcA, dstA);
}

static void glBufferData(glFunctions *f, GLenum target, GLsizeiptr size, const void *data, GLenum usage) {
	f->glBufferData(target, size, data, usage);
}

static void glBufferSubData(glFunctions *f, GLenum target, GLintptr offset, GLsizeiptr size, const void *data) {
	f->glBufferSubData(target, offset, size, data);
}

static GLenum glCheckFramebufferStatus(glFunctions *f, GLenum target) {
	return f->glCheckFramebufferStatus(target);
}

static void glClear(glFunctions *f, GLbitfield mask) {
	f->glClear(mask);
}

static void glClearColor(glFunctions *f, GLfloat red, GLfloat green, GLfloat blue, GLfloat alpha) {
	f->glClearColor(red, green, blue, alpha);
}

static void glClearDepthf(glFunctions *f, GLfloat d) {
	f->glClearDepthf(d);
}

static void glCompileShader(glFunctions *f, GLuint shader) {
	f->glCompileShader(shader);
}

static void glCopyTexSubImage2D(glFunctions *f, GLenum target, GLint level, GLint xoffset, GLint yoffset, GLint x, GLint y, GLsizei width, GLsizei height) {
	f->glCopyTexSubImage2D(target, level, xoffset, yoffset, x, y, width, height);
}

static GLuint glCreateProgram(glFunctions *f) {
	return f->glCreateProgram();
}

static GLuint glCreateShader(glFunctions *f, GLenum type) {
	return f->glCreateShader(type);
}

static void glDeleteBuffers(glFunctions *f, GLsizei n, const GLuint *buffers) {
	f->glDeleteBuffers(n, buffers);
}

static void glDeleteFramebuffers(glFunctions *f, GLsizei n, const GLuint *framebuffers) {
	f->glDeleteFramebuffers(n, framebuffers);
}

static void glDeleteProgram(glFunctions *f, GLuint program) {
	f->glDeleteProgram(program);
}

static void glDeleteRenderbuffers(glFunctions *f, GLsizei n, const GLuint *renderbuffers) {
	f->glDeleteRenderbuffers(n, renderbuffers);
}

static void glDeleteShader(glFunctions *f, GLuint shader) {
	f->glDeleteShader(shader);
}

static void glDeleteTextures(glFunctions *f, GLsizei n, const GLuint *textures) {
	f->glDeleteTextures(n, textures);
}

static void glDepthFunc(glFunctions *f, GLenum func) {
	f->glDepthFunc(func);
}

static void glDepthMask(glFunctions *f, GLboolean flag) {
	f->glDepthMask(flag);
}

static void glDisable(glFunctions *f, GLenum cap) {
	f->glDisable(cap);
}

static void glDisableVertexAttribArray(glFunctions *f, GLuint index) {
	f->glDisableVertexAttribArray(index);
}

static void glDrawArrays(glFunctions *f, GLenum mode, GLint first, GLsizei count) {
	f->glDrawArrays(mode, first, count);
}

// offset is defined as an uintptr_t to omit Cgo pointer checks.
static void glDrawElements(glFunctions *f, GLenum mode, GLsizei count, GLenum type, const uintptr_t offset) {
	f->glDrawElements(mode, count, type, (const void *)offset);
}

static void glEnable(glFunctions *f, GLenum cap) {
	f->glEnable(cap);
}

static void glEnableVertexAttribArray(glFunctions *f, GLuint index) {
	f->glEnableVertexAttribArray(index);
}

static void glFinish(glFunctions *f) {
	f->glFinish();
}

static void glFlush(glFunctions *f) {
	f->glFlush();
}

static void glFramebufferRenderbuffer(glFunctions *f, GLenum target, GLenum attachment, GLenum renderbuffertarget, GLuint renderbuffer) {
	f->glFramebufferRenderbuffer(target, attachment, renderbuffertarget, renderbuffer);
}

static void glFramebufferTexture2D(glFunctions *f, GLenum target, GLenum attachment, GLenum textarget, GLuint texture, GLint level) {
	f->glFramebufferTexture2D(target, attachment, textarget, texture, level);
}

static void glGenBuffers(glFunctions *f, GLsizei n, GLuint *buffers) {
	f->glGenBuffers(n, buffers);
}

static void glGenFramebuffers(glFunctions *f, GLsizei n, GLuint *framebuffers) {
	f->glGenFramebuffers(n, framebuffers);
}

static void glGenRenderbuffers(glFunctions *f, GLsizei n, GLuint *renderbuffers) {
	f->glGenRenderbuffers(n, renderbuffers);
}

static void glGenTextures(glFunctions *f, GLsizei n, GLuint *textures) {
	f->glGenTextures(n, textures);
}

static GLenum glGetError(glFunctions *f) {
	return f->glGetError();
}

static void glGetFramebufferAttachmentParameteriv(glFunctions *f, GLenum target, GLenum attachment, GLenum pname, GLint *params) {
	f->glGetFramebufferAttachmentParameteriv(target, attachment, pname, params);
}

static void glGetIntegerv(glFunctions *f, GLenum pname, GLint *data) {
	f->glGetIntegerv(pname, data);
}

static void glGetFloatv(glFunctions *f, GLenum pname, GLfloat *data) {
	f->glGetFloatv(pname, data);
}

static void glGetIntegeri_v(glFunctions *f, GLenum pname, GLuint idx, GLint *data) {
	f->glGetIntegeri_v(pname, idx, data);
}

static void glGetProgramiv(glFunctions *f, GLuint program, GLenum pname, GLint *params) {
	f->glGetProgramiv(program, pname, params);
}

static void glGetProgramInfoLog(glFunctions *f, GLuint program, GLsizei bufSize, GLsizei *length, GLchar *infoLog) {
	f->glGetProgramInfoLog(program, bufSize, length, infoLog);
}

static void glGetRenderbufferParameteriv(glFunctions *f, GLenum target, GLenum pname, GLint *params) {
	f->glGetRenderbufferParameteriv(target, pname, params);
}

static void glGetShaderiv(glFunctions *f, GLuint shader, GLenum pname, GLint *params) {
	f->glGetShaderiv(shader, pname, params);
}

static void glGetShaderInfoLog(glFunctions *f, GLuint shader, GLsizei bufSize, GLsizei *length, GLchar *infoLog) {
	f->glGetShaderInfoLog(shader, bufSize, length, infoLog);
}

static const GLubyte *glGetString(glFunctions *f, GLenum name) {
	return f->glGetString(name);
}

static GLint glGetUniformLocation(glFunctions *f, GLuint program, const GLchar *name) {
	return f->glGetUniformLocation(program, name);
}

static void glGetVertexAttribiv(glFunctions *f, GLuint index, GLenum pname, GLint *data) {
	f->glGetVertexAttribiv(index, pname, data);
}

// Return uintptr_t to avoid Cgo pointer check.
static uintptr_t glGetVertexAttribPointerv(glFunctions *f, GLuint index, GLenum pname) {
	void *ptrs;
	f->glGetVertexAttribPointerv(index, pname, &ptrs);
	return (uintptr_t)ptrs;
}

static GLboolean glIsEnabled(glFunctions *f, GLenum cap) {
	return f->glIsEnabled(cap);
}

static void glLinkProgram(glFunctions *f, GLuint program) {
	f->glLinkProgram(program);
}

static void glPixelStorei(glFunctions *f, GLenum pname, GLint param) {
	f->glPixelStorei(pname, param);
}

static void glReadPixels(glFunctions *f, GLint x, GLint y, GLsizei width, GLsizei height, GLenum format, GLenum type, void *pixels) {
	f->glReadPixels(x, y, width, height, format, type, pixels);
}

static void glRenderbufferStorage(glFunctions *f, GLenum target, GLenum internalformat, GLsizei width, GLsizei height) {
	f->glRenderbufferStorage(target, internalformat, width, height);
}

static void glScissor(glFunctions *f, GLint x, GLint y, GLsizei width, GLsizei height) {
	f->glScissor(x, y, width, height);
}

static void glShaderSource(glFunctions *f, GLuint shader, GLsizei count, const GLchar *const*string, const GLint *length) {
	f->glShaderSource(shader, count, string, length);
}

static void glTexImage2D(glFunctions *f, GLenum target, GLint level, GLint internalformat, GLsizei width, GLsizei height, GLint border, GLenum format, GLenum type, const void *pixels) {
	f->glTexImage2D(target, level, internalformat, width, height, border, format, type, pixels);
}

static void glTexParameteri(glFunctions *f, GLenum target, GLenum pname, GLint param) {
	f->glTexParameteri(target, pname, param);
}

static void glTexSubImage2D(glFunctions *f, GLenum target, GLint level, GLint xoffset, GLint yoffset, GLsizei width, GLsizei height, GLenum format, GLenum type, const void *pixels) {
	f->glTexSubImage2D(target, level, xoffset, yoffset, width, height, format, type, pixels);
}

static void glUniform1f(glFunctions *f, GLint location, GLfloat v0) {
	f->glUniform1f(location, v0);
}

static void glUniform1i(glFunctions *f, GLint location, GLint v0) {
	f->glUniform1i(location, v0);
}

static void glUniform2f(glFunctions *f, GLint location, GLfloat v0, GLfloat v1) {
	f->glUniform2f(location, v0, v1);
}

static void glUniform3f(glFunctions *f, GLint location, GLfloat v0, GLfloat v1, GLfloat v2) {
	f->glUniform3f(location, v0, v1, v2);
}

static void glUniform4f(glFunctions *f, GLint location, GLfloat v0, GLfloat v1, GLfloat v2, GLfloat v3) {
	f->glUniform4f(location, v0, v1, v2, v3);
}

static void glUseProgram(glFunctions *f, GLuint program) {
	f->glUseProgram(program);
}

// offset is defined as an uintptr_t to omit Cgo pointer checks.
static void glVertexAttribPointer(glFunctions *f, GLuint index, GLint size, GLenum type, GLboolean normalized, GLsizei stride, uintptr_t offset) {
	f->glVertexAttribPointer(index, size, type, normalized, stride, (const void *)offset);
}

static void glViewport(glFunctions *f, GLint x, GLint y, GLsizei width, GLsizei height) {
	f->glViewport(x, y, width, height);
}

static void glBindBufferBase(glFunctions *f, GLenum target, GLuint index, GLuint buffer) {
	f->glBindBufferBase(target, index, buffer);
}

static void glUniformBlockBinding(glFunctions *f, GLuint program, GLuint uniformBlockIndex, GLuint uniformBlockBinding) {
	f->glUniformBlockBinding(program, uniformBlockIndex, uniformBlockBinding);
}

static GLuint glGetUniformBlockIndex(glFunctions *f, GLuint program, const GLchar *uniformBlockName) {
	return f->glGetUniformBlockIndex(program, uniformBlockName);
}

static void glInvalidateFramebuffer(glFunctions *f, GLenum target, GLenum attachment) {
	// Framebuffer invalidation is just a hint and can safely be ignored.
	if (f->glInvalidateFramebuffer != NULL) {
		f->glInvalidateFramebuffer(target, 1, &attachment);
	}
}

static void glBeginQuery(glFunctions *f, GLenum target, GLenum attachment) {
	f->glBeginQuery(target, attachment);
}

static void glDeleteQueries(glFunctions *f, GLsizei n, const GLuint *ids) {
	f->glDeleteQueries(n, ids);
}

static void glDeleteVertexArrays(glFunctions *f, GLsizei n, const GLuint *ids) {
	f->glDeleteVertexArrays(n, ids);
}

static void glEndQuery(glFunctions *f, GLenum target) {
	f->glEndQuery(target);
}

static const GLubyte* glGetStringi(glFunctions *f, GLenum name, GLuint index) {
	return f->glGetStringi(name, index);
}

static void glGenQueries(glFunctions *f, GLsizei n, GLuint *ids) {
	f->glGenQueries(n, ids);
}

static void glGenVertexArrays(glFunctions *f, GLsizei n, GLuint *ids) {
	f->glGenVertexArrays(n, ids);
}

static void glGetProgramBinary(glFunctions *f, GLuint program, GLsizei bufsize, GLsizei *length, GLenum *binaryFormat, void *binary) {
	f->glGetProgramBinary(program, bufsize, length, binaryFormat, binary);
}

static void glGetQueryObjectuiv(glFunctions *f, GLuint id, GLenum pname, GLuint *params) {
	f->glGetQueryObjectuiv(id, pname, params);
}

static void glMemoryBarrier(glFunctions *f, GLbitfield barriers) {
	f->glMemoryBarrier(barriers);
}

static void glDispatchCompute(glFunctions *f, GLuint x, GLuint y, GLuint z) {
	f->glDispatchCompute(x, y, z);
}

static void *glMapBufferRange(glFunctions *f, GLenum target, GLintptr offset, GLsizeiptr length, GLbitfield access) {
	return f->glMapBufferRange(target, offset, length, access);
}

static GLboolean glUnmapBuffer(glFunctions *f, GLenum target) {
	return f->glUnmapBuffer(target);
}

static void glBindImageTexture(glFunctions *f, GLuint unit, GLuint texture, GLint level, GLboolean layered, GLint layer, GLenum access, GLenum format) {
	f->glBindImageTexture(unit, texture, level, layered, layer, access, format);
}

static void glTexStorage2D(glFunctions *f, GLenum target, GLsizei levels, GLenum internalFormat, GLsizei width, GLsizei height) {
	f->glTexStorage2D(target, levels, internalFormat, width, height);
}

static void glBlitFramebuffer(glFunctions *f, GLint srcX0, GLint srcY0, GLint srcX1, GLint srcY1, GLint dstX0, GLint dstY0, GLint dstX1, GLint dstY1, GLbitfield mask, GLenum filter) {
	f->glBlitFramebuffer(srcX0, srcY0, srcX1, srcY1, dstX0, dstY0, dstX1, dstY1, mask, filter);
}
*/
import "C"

type Context interface{}

type Functions struct {
	// Query caches.
	uints  [100]C.GLuint
	ints   [100]C.GLint
	floats [100]C.GLfloat

	f C.glFunctions
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
	f.f.glActiveTexture = must("glActiveTexture")
	f.f.glAttachShader = must("glAttachShader")
	f.f.glBindAttribLocation = must("glBindAttribLocation")
	f.f.glBindBuffer = must("glBindBuffer")
	f.f.glBindFramebuffer = must("glBindFramebuffer")
	f.f.glBindRenderbuffer = must("glBindRenderbuffer")
	f.f.glBindTexture = must("glBindTexture")
	f.f.glBlendEquation = must("glBlendEquation")
	f.f.glBlendFuncSeparate = must("glBlendFuncSeparate")
	f.f.glBufferData = must("glBufferData")
	f.f.glBufferSubData = must("glBufferSubData")
	f.f.glCheckFramebufferStatus = must("glCheckFramebufferStatus")
	f.f.glClear = must("glClear")
	f.f.glClearColor = must("glClearColor")
	f.f.glClearDepthf = must("glClearDepthf")
	f.f.glCompileShader = must("glCompileShader")
	f.f.glCopyTexSubImage2D = must("glCopyTexSubImage2D")
	f.f.glCreateProgram = must("glCreateProgram")
	f.f.glCreateShader = must("glCreateShader")
	f.f.glDeleteBuffers = must("glDeleteBuffers")
	f.f.glDeleteFramebuffers = must("glDeleteFramebuffers")
	f.f.glDeleteProgram = must("glDeleteProgram")
	f.f.glDeleteRenderbuffers = must("glDeleteRenderbuffers")
	f.f.glDeleteShader = must("glDeleteShader")
	f.f.glDeleteTextures = must("glDeleteTextures")
	f.f.glDepthFunc = must("glDepthFunc")
	f.f.glDepthMask = must("glDepthMask")
	f.f.glDisable = must("glDisable")
	f.f.glDisableVertexAttribArray = must("glDisableVertexAttribArray")
	f.f.glDrawArrays = must("glDrawArrays")
	f.f.glDrawElements = must("glDrawElements")
	f.f.glEnable = must("glEnable")
	f.f.glEnableVertexAttribArray = must("glEnableVertexAttribArray")
	f.f.glFinish = must("glFinish")
	f.f.glFlush = must("glFlush")
	f.f.glFramebufferRenderbuffer = must("glFramebufferRenderbuffer")
	f.f.glFramebufferTexture2D = must("glFramebufferTexture2D")
	f.f.glGenBuffers = must("glGenBuffers")
	f.f.glGenFramebuffers = must("glGenFramebuffers")
	f.f.glGenRenderbuffers = must("glGenRenderbuffers")
	f.f.glGenTextures = must("glGenTextures")
	f.f.glGetError = must("glGetError")
	f.f.glGetFramebufferAttachmentParameteriv = must("glGetFramebufferAttachmentParameteriv")
	f.f.glGetIntegerv = must("glGetIntegerv")
	f.f.glGetFloatv = must("glGetFloatv")
	f.f.glGetProgramiv = must("glGetProgramiv")
	f.f.glGetProgramInfoLog = must("glGetProgramInfoLog")
	f.f.glGetRenderbufferParameteriv = must("glGetRenderbufferParameteriv")
	f.f.glGetShaderiv = must("glGetShaderiv")
	f.f.glGetShaderInfoLog = must("glGetShaderInfoLog")
	f.f.glGetString = must("glGetString")
	f.f.glGetUniformLocation = must("glGetUniformLocation")
	f.f.glGetVertexAttribiv = must("glGetVertexAttribiv")
	f.f.glGetVertexAttribPointerv = must("glGetVertexAttribPointerv")
	f.f.glIsEnabled = must("glIsEnabled")
	f.f.glLinkProgram = must("glLinkProgram")
	f.f.glPixelStorei = must("glPixelStorei")
	f.f.glReadPixels = must("glReadPixels")
	f.f.glRenderbufferStorage = must("glRenderbufferStorage")
	f.f.glScissor = must("glScissor")
	f.f.glShaderSource = must("glShaderSource")
	f.f.glTexImage2D = must("glTexImage2D")
	f.f.glTexParameteri = must("glTexParameteri")
	f.f.glTexSubImage2D = must("glTexSubImage2D")
	f.f.glUniform1f = must("glUniform1f")
	f.f.glUniform1i = must("glUniform1i")
	f.f.glUniform2f = must("glUniform2f")
	f.f.glUniform3f = must("glUniform3f")
	f.f.glUniform4f = must("glUniform4f")
	f.f.glUseProgram = must("glUseProgram")
	f.f.glVertexAttribPointer = must("glVertexAttribPointer")
	f.f.glViewport = must("glViewport")

	// Extensions and GL ES 3 functions.
	f.f.glBindBufferBase = load("glBindBufferBase")
	f.f.glBindVertexArray = load("glBindVertexArray")
	f.f.glGetIntegeri_v = load("glGetIntegeri_v")
	f.f.glGetUniformBlockIndex = load("glGetUniformBlockIndex")
	f.f.glUniformBlockBinding = load("glUniformBlockBinding")
	f.f.glInvalidateFramebuffer = load("glInvalidateFramebuffer")
	f.f.glGetStringi = load("glGetStringi")
	// Fall back to EXT_invalidate_framebuffer if available.
	if f.f.glInvalidateFramebuffer == nil {
		f.f.glInvalidateFramebuffer = load("glDiscardFramebufferEXT")
	}

	f.f.glBeginQuery = load("glBeginQuery")
	if f.f.glBeginQuery == nil {
		f.f.glBeginQuery = load("glBeginQueryEXT")
	}
	f.f.glDeleteQueries = load("glDeleteQueries")
	if f.f.glDeleteQueries == nil {
		f.f.glDeleteQueries = load("glDeleteQueriesEXT")
	}
	f.f.glEndQuery = load("glEndQuery")
	if f.f.glEndQuery == nil {
		f.f.glEndQuery = load("glEndQueryEXT")
	}
	f.f.glGenQueries = load("glGenQueries")
	if f.f.glGenQueries == nil {
		f.f.glGenQueries = load("glGenQueriesEXT")
	}
	f.f.glGetQueryObjectuiv = load("glGetQueryObjectuiv")
	if f.f.glGetQueryObjectuiv == nil {
		f.f.glGetQueryObjectuiv = load("glGetQueryObjectuivEXT")
	}

	f.f.glDeleteVertexArrays = load("glDeleteVertexArrays")
	f.f.glGenVertexArrays = load("glGenVertexArrays")
	f.f.glMemoryBarrier = load("glMemoryBarrier")
	f.f.glDispatchCompute = load("glDispatchCompute")
	f.f.glMapBufferRange = load("glMapBufferRange")
	f.f.glUnmapBuffer = load("glUnmapBuffer")
	f.f.glBindImageTexture = load("glBindImageTexture")
	f.f.glTexStorage2D = load("glTexStorage2D")
	f.f.glBlitFramebuffer = load("glBlitFramebuffer")
	f.f.glGetProgramBinary = load("glGetProgramBinary")

	return loadErr
}

func (f *Functions) ActiveTexture(texture Enum) {
	C.glActiveTexture(&f.f, C.GLenum(texture))
}

func (f *Functions) AttachShader(p Program, s Shader) {
	C.glAttachShader(&f.f, C.GLuint(p.V), C.GLuint(s.V))
}

func (f *Functions) BeginQuery(target Enum, query Query) {
	C.glBeginQuery(&f.f, C.GLenum(target), C.GLenum(query.V))
}

func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.glBindAttribLocation(&f.f, C.GLuint(p.V), C.GLuint(a), cname)
}

func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	C.glBindBufferBase(&f.f, C.GLenum(target), C.GLuint(index), C.GLuint(b.V))
}

func (f *Functions) BindBuffer(target Enum, b Buffer) {
	C.glBindBuffer(&f.f, C.GLenum(target), C.GLuint(b.V))
}

func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	C.glBindFramebuffer(&f.f, C.GLenum(target), C.GLuint(fb.V))
}

func (f *Functions) BindRenderbuffer(target Enum, fb Renderbuffer) {
	C.glBindRenderbuffer(&f.f, C.GLenum(target), C.GLuint(fb.V))
}

func (f *Functions) BindImageTexture(unit int, t Texture, level int, layered bool, layer int, access, format Enum) {
	l := C.GLboolean(FALSE)
	if layered {
		l = TRUE
	}
	C.glBindImageTexture(&f.f, C.GLuint(unit), C.GLuint(t.V), C.GLint(level), l, C.GLint(layer), C.GLenum(access), C.GLenum(format))
}

func (f *Functions) BindTexture(target Enum, t Texture) {
	C.glBindTexture(&f.f, C.GLenum(target), C.GLuint(t.V))
}

func (f *Functions) BindVertexArray(a VertexArray) {
	C.glBindVertexArray(&f.f, C.GLuint(a.V))
}

func (f *Functions) BlendEquation(mode Enum) {
	C.glBlendEquation(&f.f, C.GLenum(mode))
}

func (f *Functions) BlendFuncSeparate(srcRGB, dstRGB, srcA, dstA Enum) {
	C.glBlendFuncSeparate(&f.f, C.GLenum(srcRGB), C.GLenum(dstRGB), C.GLenum(srcA), C.GLenum(dstA))
}

func (f *Functions) BlitFramebuffer(sx0, sy0, sx1, sy1, dx0, dy0, dx1, dy1 int, mask Enum, filter Enum) {
	C.glBlitFramebuffer(&f.f,
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
	C.glBufferData(&f.f, C.GLenum(target), C.GLsizeiptr(size), p, C.GLenum(usage))
}

func (f *Functions) BufferSubData(target Enum, offset int, src []byte) {
	var p unsafe.Pointer
	if len(src) > 0 {
		p = unsafe.Pointer(&src[0])
	}
	C.glBufferSubData(&f.f, C.GLenum(target), C.GLintptr(offset), C.GLsizeiptr(len(src)), p)
}

func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(C.glCheckFramebufferStatus(&f.f, C.GLenum(target)))
}

func (f *Functions) Clear(mask Enum) {
	C.glClear(&f.f, C.GLbitfield(mask))
}

func (f *Functions) ClearColor(red float32, green float32, blue float32, alpha float32) {
	C.glClearColor(&f.f, C.GLfloat(red), C.GLfloat(green), C.GLfloat(blue), C.GLfloat(alpha))
}

func (f *Functions) ClearDepthf(d float32) {
	C.glClearDepthf(&f.f, C.GLfloat(d))
}

func (f *Functions) CompileShader(s Shader) {
	C.glCompileShader(&f.f, C.GLuint(s.V))
}

func (f *Functions) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	C.glCopyTexSubImage2D(&f.f, C.GLenum(target), C.GLint(level), C.GLint(xoffset), C.GLint(yoffset), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) CreateBuffer() Buffer {
	C.glGenBuffers(&f.f, 1, &f.uints[0])
	return Buffer{uint(f.uints[0])}
}

func (f *Functions) CreateFramebuffer() Framebuffer {
	C.glGenFramebuffers(&f.f, 1, &f.uints[0])
	return Framebuffer{uint(f.uints[0])}
}

func (f *Functions) CreateProgram() Program {
	return Program{uint(C.glCreateProgram(&f.f))}
}

func (f *Functions) CreateQuery() Query {
	C.glGenQueries(&f.f, 1, &f.uints[0])
	return Query{uint(f.uints[0])}
}

func (f *Functions) CreateRenderbuffer() Renderbuffer {
	C.glGenRenderbuffers(&f.f, 1, &f.uints[0])
	return Renderbuffer{uint(f.uints[0])}
}

func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader{uint(C.glCreateShader(&f.f, C.GLenum(ty)))}
}

func (f *Functions) CreateTexture() Texture {
	C.glGenTextures(&f.f, 1, &f.uints[0])
	return Texture{uint(f.uints[0])}
}

func (f *Functions) CreateVertexArray() VertexArray {
	C.glGenVertexArrays(&f.f, 1, &f.uints[0])
	return VertexArray{uint(f.uints[0])}
}

func (f *Functions) DeleteBuffer(v Buffer) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteBuffers(&f.f, 1, &f.uints[0])
}

func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteFramebuffers(&f.f, 1, &f.uints[0])
}

func (f *Functions) DeleteProgram(p Program) {
	C.glDeleteProgram(&f.f, C.GLuint(p.V))
}

func (f *Functions) DeleteQuery(query Query) {
	f.uints[0] = C.GLuint(query.V)
	C.glDeleteQueries(&f.f, 1, &f.uints[0])
}

func (f *Functions) DeleteVertexArray(array VertexArray) {
	f.uints[0] = C.GLuint(array.V)
	C.glDeleteVertexArrays(&f.f, 1, &f.uints[0])
}

func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteRenderbuffers(&f.f, 1, &f.uints[0])
}

func (f *Functions) DeleteShader(s Shader) {
	C.glDeleteShader(&f.f, C.GLuint(s.V))
}

func (f *Functions) DeleteTexture(v Texture) {
	f.uints[0] = C.GLuint(v.V)
	C.glDeleteTextures(&f.f, 1, &f.uints[0])
}

func (f *Functions) DepthFunc(v Enum) {
	C.glDepthFunc(&f.f, C.GLenum(v))
}

func (f *Functions) DepthMask(mask bool) {
	m := C.GLboolean(FALSE)
	if mask {
		m = C.GLboolean(TRUE)
	}
	C.glDepthMask(&f.f, m)
}

func (f *Functions) DisableVertexAttribArray(a Attrib) {
	C.glDisableVertexAttribArray(&f.f, C.GLuint(a))
}

func (f *Functions) Disable(cap Enum) {
	C.glDisable(&f.f, C.GLenum(cap))
}

func (f *Functions) DrawArrays(mode Enum, first int, count int) {
	C.glDrawArrays(&f.f, C.GLenum(mode), C.GLint(first), C.GLsizei(count))
}

func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	C.glDrawElements(&f.f, C.GLenum(mode), C.GLsizei(count), C.GLenum(ty), C.uintptr_t(offset))
}

func (f *Functions) DispatchCompute(x, y, z int) {
	C.glDispatchCompute(&f.f, C.GLuint(x), C.GLuint(y), C.GLuint(z))
}

func (f *Functions) Enable(cap Enum) {
	C.glEnable(&f.f, C.GLenum(cap))
}

func (f *Functions) EndQuery(target Enum) {
	C.glEndQuery(&f.f, C.GLenum(target))
}

func (f *Functions) EnableVertexAttribArray(a Attrib) {
	C.glEnableVertexAttribArray(&f.f, C.GLuint(a))
}

func (f *Functions) Finish() {
	C.glFinish(&f.f)
}

func (f *Functions) Flush() {
	C.glFlush(&f.f)
}

func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	C.glFramebufferRenderbuffer(&f.f, C.GLenum(target), C.GLenum(attachment), C.GLenum(renderbuffertarget), C.GLuint(renderbuffer.V))
}

func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	C.glFramebufferTexture2D(&f.f, C.GLenum(target), C.GLenum(attachment), C.GLenum(texTarget), C.GLuint(t.V), C.GLint(level))
}

func (c *Functions) GetBinding(pname Enum) Object {
	return Object{uint(c.GetInteger(pname))}
}

func (c *Functions) GetBindingi(pname Enum, idx int) Object {
	return Object{uint(c.GetIntegeri(pname, idx))}
}

func (f *Functions) GetError() Enum {
	return Enum(C.glGetError(&f.f))
}

func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	C.glGetRenderbufferParameteriv(&f.f, C.GLenum(target), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	C.glGetFramebufferAttachmentParameteriv(&f.f, C.GLenum(target), C.GLenum(attachment), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetFloat4(pname Enum) [4]float32 {
	C.glGetFloatv(&f.f, C.GLenum(pname), &f.floats[0])
	var r [4]float32
	for i := range r {
		r[i] = float32(f.floats[i])
	}
	return r
}

func (f *Functions) GetFloat(pname Enum) float32 {
	C.glGetFloatv(&f.f, C.GLenum(pname), &f.floats[0])
	return float32(f.floats[0])
}

func (f *Functions) GetInteger4(pname Enum) [4]int {
	C.glGetIntegerv(&f.f, C.GLenum(pname), &f.ints[0])
	var r [4]int
	for i := range r {
		r[i] = int(f.ints[i])
	}
	return r
}

func (f *Functions) GetInteger(pname Enum) int {
	C.glGetIntegerv(&f.f, C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetIntegeri(pname Enum, idx int) int {
	C.glGetIntegeri_v(&f.f, C.GLenum(pname), C.GLuint(idx), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetProgrami(p Program, pname Enum) int {
	C.glGetProgramiv(&f.f, C.GLuint(p.V), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetProgramBinary(p Program) []byte {
	sz := f.GetProgrami(p, PROGRAM_BINARY_LENGTH)
	if sz == 0 {
		return nil
	}
	buf := make([]byte, sz)
	var format C.GLenum
	C.glGetProgramBinary(&f.f, C.GLuint(p.V), C.GLsizei(sz), nil, &format, unsafe.Pointer(&buf[0]))
	return buf
}

func (f *Functions) GetProgramInfoLog(p Program) string {
	n := f.GetProgrami(p, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	C.glGetProgramInfoLog(&f.f, C.GLuint(p.V), C.GLsizei(len(buf)), nil, (*C.GLchar)(unsafe.Pointer(&buf[0])))
	return string(buf)
}

func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	C.glGetQueryObjectuiv(&f.f, C.GLuint(query.V), C.GLenum(pname), &f.uints[0])
	return uint(f.uints[0])
}

func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	C.glGetShaderiv(&f.f, C.GLuint(s.V), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetShaderInfoLog(s Shader) string {
	n := f.GetShaderi(s, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	C.glGetShaderInfoLog(&f.f, C.GLuint(s.V), C.GLsizei(len(buf)), nil, (*C.GLchar)(unsafe.Pointer(&buf[0])))
	return string(buf)
}

func (f *Functions) getStringi(pname Enum, index int) string {
	str := C.glGetStringi(&f.f, C.GLenum(pname), C.GLuint(index))
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
		str := C.glGetString(&f.f, C.GLenum(pname))
		return C.GoString((*C.char)(unsafe.Pointer(str)))
	}
}

func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return uint(C.glGetUniformBlockIndex(&f.f, C.GLuint(p.V), cname))
}

func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return Uniform{int(C.glGetUniformLocation(&f.f, C.GLuint(p.V), cname))}
}

func (f *Functions) GetVertexAttrib(index int, pname Enum) int {
	C.glGetVertexAttribiv(&f.f, C.GLuint(index), C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *Functions) GetVertexAttribBinding(index int, pname Enum) Object {
	return Object{uint(f.GetVertexAttrib(index, pname))}
}

func (f *Functions) GetVertexAttribPointer(index int, pname Enum) uintptr {
	ptr := C.glGetVertexAttribPointerv(&f.f, C.GLuint(index), C.GLenum(pname))
	return uintptr(ptr)
}

func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	C.glInvalidateFramebuffer(&f.f, C.GLenum(target), C.GLenum(attachment))
}

func (f *Functions) IsEnabled(cap Enum) bool {
	return C.glIsEnabled(&f.f, C.GLenum(cap)) == TRUE
}

func (f *Functions) LinkProgram(p Program) {
	C.glLinkProgram(&f.f, C.GLuint(p.V))
}

func (f *Functions) PixelStorei(pname Enum, param int) {
	C.glPixelStorei(&f.f, C.GLenum(pname), C.GLint(param))
}

func (f *Functions) MemoryBarrier(barriers Enum) {
	C.glMemoryBarrier(&f.f, C.GLbitfield(barriers))
}

func (f *Functions) MapBufferRange(target Enum, offset, length int, access Enum) []byte {
	p := C.glMapBufferRange(&f.f, C.GLenum(target), C.GLintptr(offset), C.GLsizeiptr(length), C.GLbitfield(access))
	if p == nil {
		return nil
	}
	return (*[1 << 30]byte)(p)[:length:length]
}

func (f *Functions) Scissor(x, y, width, height int32) {
	C.glScissor(&f.f, C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glReadPixels(&f.f, C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(ty), p)
}

func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	C.glRenderbufferStorage(&f.f, C.GLenum(target), C.GLenum(internalformat), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) ShaderSource(s Shader, src string) {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))
	strlen := C.GLint(len(src))
	C.glShaderSource(&f.f, C.GLuint(s.V), 1, &csrc, &strlen)
}

func (f *Functions) TexImage2D(target Enum, level int, internalFormat Enum, width int, height int, format Enum, ty Enum) {
	C.glTexImage2D(&f.f, C.GLenum(target), C.GLint(level), C.GLint(internalFormat), C.GLsizei(width), C.GLsizei(height), 0, C.GLenum(format), C.GLenum(ty), nil)
}

func (f *Functions) TexStorage2D(target Enum, levels int, internalFormat Enum, width, height int) {
	C.glTexStorage2D(&f.f, C.GLenum(target), C.GLsizei(levels), C.GLenum(internalFormat), C.GLsizei(width), C.GLsizei(height))
}

func (f *Functions) TexSubImage2D(target Enum, level int, x int, y int, width int, height int, format Enum, ty Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glTexSubImage2D(&f.f, C.GLenum(target), C.GLint(level), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(ty), p)
}

func (f *Functions) TexParameteri(target, pname Enum, param int) {
	C.glTexParameteri(&f.f, C.GLenum(target), C.GLenum(pname), C.GLint(param))
}

func (f *Functions) UniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint) {
	C.glUniformBlockBinding(&f.f, C.GLuint(p.V), C.GLuint(uniformBlockIndex), C.GLuint(uniformBlockBinding))
}

func (f *Functions) Uniform1f(dst Uniform, v float32) {
	C.glUniform1f(&f.f, C.GLint(dst.V), C.GLfloat(v))
}

func (f *Functions) Uniform1i(dst Uniform, v int) {
	C.glUniform1i(&f.f, C.GLint(dst.V), C.GLint(v))
}

func (f *Functions) Uniform2f(dst Uniform, v0 float32, v1 float32) {
	C.glUniform2f(&f.f, C.GLint(dst.V), C.GLfloat(v0), C.GLfloat(v1))
}

func (f *Functions) Uniform3f(dst Uniform, v0 float32, v1 float32, v2 float32) {
	C.glUniform3f(&f.f, C.GLint(dst.V), C.GLfloat(v0), C.GLfloat(v1), C.GLfloat(v2))
}

func (f *Functions) Uniform4f(dst Uniform, v0 float32, v1 float32, v2 float32, v3 float32) {
	C.glUniform4f(&f.f, C.GLint(dst.V), C.GLfloat(v0), C.GLfloat(v1), C.GLfloat(v2), C.GLfloat(v3))
}

func (f *Functions) UseProgram(p Program) {
	C.glUseProgram(&f.f, C.GLuint(p.V))
}

func (f *Functions) UnmapBuffer(target Enum) bool {
	r := C.glUnmapBuffer(&f.f, C.GLenum(target))
	return r == TRUE
}

func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride int, offset int) {
	var n C.GLboolean = FALSE
	if normalized {
		n = TRUE
	}
	C.glVertexAttribPointer(&f.f, C.GLuint(dst), C.GLint(size), C.GLenum(ty), n, C.GLsizei(stride), C.uintptr_t(offset))
}

func (f *Functions) Viewport(x int, y int, width int, height int) {
	C.glViewport(&f.f, C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}
