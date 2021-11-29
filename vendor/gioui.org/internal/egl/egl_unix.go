// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux || freebsd || openbsd
// +build linux freebsd openbsd

package egl

/*
#cgo linux,!android  pkg-config: egl
#cgo freebsd openbsd android LDFLAGS: -lEGL
#cgo freebsd CFLAGS: -I/usr/local/include
#cgo freebsd LDFLAGS: -L/usr/local/lib
#cgo openbsd CFLAGS: -I/usr/X11R6/include
#cgo openbsd LDFLAGS: -L/usr/X11R6/lib
#cgo CFLAGS: -DEGL_NO_X11

#include <EGL/egl.h>
#include <EGL/eglext.h>
*/
import "C"

type (
	_EGLint           = C.EGLint
	_EGLDisplay       = C.EGLDisplay
	_EGLConfig        = C.EGLConfig
	_EGLContext       = C.EGLContext
	_EGLSurface       = C.EGLSurface
	NativeDisplayType = C.EGLNativeDisplayType
	NativeWindowType  = C.EGLNativeWindowType
)

func loadEGL() error {
	return nil
}

func eglChooseConfig(disp _EGLDisplay, attribs []_EGLint) (_EGLConfig, bool) {
	var cfg C.EGLConfig
	var ncfg C.EGLint
	if C.eglChooseConfig(disp, &attribs[0], &cfg, 1, &ncfg) != C.EGL_TRUE {
		return nilEGLConfig, false
	}
	return _EGLConfig(cfg), true
}

func eglCreateContext(disp _EGLDisplay, cfg _EGLConfig, shareCtx _EGLContext, attribs []_EGLint) _EGLContext {
	ctx := C.eglCreateContext(disp, cfg, shareCtx, &attribs[0])
	return _EGLContext(ctx)
}

func eglDestroySurface(disp _EGLDisplay, surf _EGLSurface) bool {
	return C.eglDestroySurface(disp, surf) == C.EGL_TRUE
}

func eglDestroyContext(disp _EGLDisplay, ctx _EGLContext) bool {
	return C.eglDestroyContext(disp, ctx) == C.EGL_TRUE
}

func eglGetConfigAttrib(disp _EGLDisplay, cfg _EGLConfig, attr _EGLint) (_EGLint, bool) {
	var val _EGLint
	ret := C.eglGetConfigAttrib(disp, cfg, attr, &val)
	return val, ret == C.EGL_TRUE
}

func eglGetError() _EGLint {
	return C.eglGetError()
}

func eglInitialize(disp _EGLDisplay) (_EGLint, _EGLint, bool) {
	var maj, min _EGLint
	ret := C.eglInitialize(disp, &maj, &min)
	return maj, min, ret == C.EGL_TRUE
}

func eglMakeCurrent(disp _EGLDisplay, draw, read _EGLSurface, ctx _EGLContext) bool {
	return C.eglMakeCurrent(disp, draw, read, ctx) == C.EGL_TRUE
}

func eglReleaseThread() bool {
	return C.eglReleaseThread() == C.EGL_TRUE
}

func eglSwapBuffers(disp _EGLDisplay, surf _EGLSurface) bool {
	return C.eglSwapBuffers(disp, surf) == C.EGL_TRUE
}

func eglSwapInterval(disp _EGLDisplay, interval _EGLint) bool {
	return C.eglSwapInterval(disp, interval) == C.EGL_TRUE
}

func eglTerminate(disp _EGLDisplay) bool {
	return C.eglTerminate(disp) == C.EGL_TRUE
}

func eglQueryString(disp _EGLDisplay, name _EGLint) string {
	return C.GoString(C.eglQueryString(disp, name))
}

func eglGetDisplay(disp NativeDisplayType) _EGLDisplay {
	return C.eglGetDisplay(disp)
}

func eglCreateWindowSurface(disp _EGLDisplay, conf _EGLConfig, win NativeWindowType, attribs []_EGLint) _EGLSurface {
	eglSurf := C.eglCreateWindowSurface(disp, conf, win, &attribs[0])
	return eglSurf
}

func eglWaitClient() bool {
	return C.eglWaitClient() == C.EGL_TRUE
}
