// SPDX-License-Identifier: Unlicense OR MIT

package egl

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	syscall "golang.org/x/sys/windows"

	"gioui.org/internal/gl"
)

type (
	_EGLint           int32
	_EGLDisplay       uintptr
	_EGLConfig        uintptr
	_EGLContext       uintptr
	_EGLSurface       uintptr
	NativeDisplayType uintptr
	NativeWindowType  uintptr
)

var (
	libEGL                  = syscall.NewLazyDLL("libEGL.dll")
	_eglChooseConfig        = libEGL.NewProc("eglChooseConfig")
	_eglCreateContext       = libEGL.NewProc("eglCreateContext")
	_eglCreateWindowSurface = libEGL.NewProc("eglCreateWindowSurface")
	_eglDestroyContext      = libEGL.NewProc("eglDestroyContext")
	_eglDestroySurface      = libEGL.NewProc("eglDestroySurface")
	_eglGetConfigAttrib     = libEGL.NewProc("eglGetConfigAttrib")
	_eglGetDisplay          = libEGL.NewProc("eglGetDisplay")
	_eglGetError            = libEGL.NewProc("eglGetError")
	_eglInitialize          = libEGL.NewProc("eglInitialize")
	_eglMakeCurrent         = libEGL.NewProc("eglMakeCurrent")
	_eglReleaseThread       = libEGL.NewProc("eglReleaseThread")
	_eglSwapInterval        = libEGL.NewProc("eglSwapInterval")
	_eglSwapBuffers         = libEGL.NewProc("eglSwapBuffers")
	_eglTerminate           = libEGL.NewProc("eglTerminate")
	_eglQueryString         = libEGL.NewProc("eglQueryString")
	_eglWaitClient          = libEGL.NewProc("eglWaitClient")
)

var loadOnce sync.Once

func loadEGL() error {
	var err error
	loadOnce.Do(func() {
		err = loadDLLs()
	})
	return err
}

func loadDLLs() error {
	if err := loadDLL(libEGL, "libEGL.dll"); err != nil {
		return err
	}
	if err := loadDLL(gl.LibGLESv2, "libGLESv2.dll"); err != nil {
		return err
	}
	// d3dcompiler_47.dll is needed internally for shader compilation to function.
	return loadDLL(syscall.NewLazyDLL("d3dcompiler_47.dll"), "d3dcompiler_47.dll")
}

func loadDLL(dll *syscall.LazyDLL, name string) error {
	err := dll.Load()
	if err != nil {
		return fmt.Errorf("egl: failed to load %s: %v", name, err)
	}
	return nil
}

func eglChooseConfig(disp _EGLDisplay, attribs []_EGLint) (_EGLConfig, bool) {
	var cfg _EGLConfig
	var ncfg _EGLint
	a := &attribs[0]
	r, _, _ := _eglChooseConfig.Call(uintptr(disp), uintptr(unsafe.Pointer(a)), uintptr(unsafe.Pointer(&cfg)), 1, uintptr(unsafe.Pointer(&ncfg)))
	issue34474KeepAlive(a)
	return cfg, r != 0
}

func eglCreateContext(disp _EGLDisplay, cfg _EGLConfig, shareCtx _EGLContext, attribs []_EGLint) _EGLContext {
	a := &attribs[0]
	c, _, _ := _eglCreateContext.Call(uintptr(disp), uintptr(cfg), uintptr(shareCtx), uintptr(unsafe.Pointer(a)))
	issue34474KeepAlive(a)
	return _EGLContext(c)
}

func eglCreateWindowSurface(disp _EGLDisplay, cfg _EGLConfig, win NativeWindowType, attribs []_EGLint) _EGLSurface {
	a := &attribs[0]
	s, _, _ := _eglCreateWindowSurface.Call(uintptr(disp), uintptr(cfg), uintptr(win), uintptr(unsafe.Pointer(a)))
	issue34474KeepAlive(a)
	return _EGLSurface(s)
}

func eglDestroySurface(disp _EGLDisplay, surf _EGLSurface) bool {
	r, _, _ := _eglDestroySurface.Call(uintptr(disp), uintptr(surf))
	return r != 0
}

func eglDestroyContext(disp _EGLDisplay, ctx _EGLContext) bool {
	r, _, _ := _eglDestroyContext.Call(uintptr(disp), uintptr(ctx))
	return r != 0
}

func eglGetConfigAttrib(disp _EGLDisplay, cfg _EGLConfig, attr _EGLint) (_EGLint, bool) {
	var val uintptr
	r, _, _ := _eglGetConfigAttrib.Call(uintptr(disp), uintptr(cfg), uintptr(attr), uintptr(unsafe.Pointer(&val)))
	return _EGLint(val), r != 0
}

func eglGetDisplay(disp NativeDisplayType) _EGLDisplay {
	d, _, _ := _eglGetDisplay.Call(uintptr(disp))
	return _EGLDisplay(d)
}

func eglGetError() _EGLint {
	e, _, _ := _eglGetError.Call()
	return _EGLint(e)
}

func eglInitialize(disp _EGLDisplay) (_EGLint, _EGLint, bool) {
	var maj, min uintptr
	r, _, _ := _eglInitialize.Call(uintptr(disp), uintptr(unsafe.Pointer(&maj)), uintptr(unsafe.Pointer(&min)))
	return _EGLint(maj), _EGLint(min), r != 0
}

func eglMakeCurrent(disp _EGLDisplay, draw, read _EGLSurface, ctx _EGLContext) bool {
	r, _, _ := _eglMakeCurrent.Call(uintptr(disp), uintptr(draw), uintptr(read), uintptr(ctx))
	return r != 0
}

func eglReleaseThread() bool {
	r, _, _ := _eglReleaseThread.Call()
	return r != 0
}

func eglSwapInterval(disp _EGLDisplay, interval _EGLint) bool {
	r, _, _ := _eglSwapInterval.Call(uintptr(disp), uintptr(interval))
	return r != 0
}

func eglSwapBuffers(disp _EGLDisplay, surf _EGLSurface) bool {
	r, _, _ := _eglSwapBuffers.Call(uintptr(disp), uintptr(surf))
	return r != 0
}

func eglTerminate(disp _EGLDisplay) bool {
	r, _, _ := _eglTerminate.Call(uintptr(disp))
	return r != 0
}

func eglQueryString(disp _EGLDisplay, name _EGLint) string {
	r, _, _ := _eglQueryString.Call(uintptr(disp), uintptr(name))
	return syscall.BytePtrToString((*byte)(unsafe.Pointer(r)))
}

func eglWaitClient() bool {
	r, _, _ := _eglWaitClient.Call()
	return r != 0
}

// issue34474KeepAlive calls runtime.KeepAlive as a
// workaround for golang.org/issue/34474.
func issue34474KeepAlive(v interface{}) {
	runtime.KeepAlive(v)
}
