// SPDX-License-Identifier: Unlicense OR MIT

package app

/*
#cgo CFLAGS: -Werror
#cgo LDFLAGS: -landroid

#include <android/native_window_jni.h>
#include <android/configuration.h>
#include <android/keycodes.h>
#include <android/input.h>
#include <stdlib.h>

static jint jni_GetEnv(JavaVM *vm, JNIEnv **env, jint version) {
	return (*vm)->GetEnv(vm, (void **)env, version);
}

static jint jni_GetJavaVM(JNIEnv *env, JavaVM **jvm) {
	return (*env)->GetJavaVM(env, jvm);
}

static jint jni_AttachCurrentThread(JavaVM *vm, JNIEnv **p_env, void *thr_args) {
	return (*vm)->AttachCurrentThread(vm, p_env, thr_args);
}

static jint jni_DetachCurrentThread(JavaVM *vm) {
	return (*vm)->DetachCurrentThread(vm);
}

static jobject jni_NewGlobalRef(JNIEnv *env, jobject obj) {
	return (*env)->NewGlobalRef(env, obj);
}

static void jni_DeleteGlobalRef(JNIEnv *env, jobject obj) {
	(*env)->DeleteGlobalRef(env, obj);
}

static jclass jni_GetObjectClass(JNIEnv *env, jobject obj) {
	return (*env)->GetObjectClass(env, obj);
}

static jmethodID jni_GetMethodID(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	return (*env)->GetMethodID(env, clazz, name, sig);
}

static jmethodID jni_GetStaticMethodID(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	return (*env)->GetStaticMethodID(env, clazz, name, sig);
}

static jfloat jni_CallFloatMethod(JNIEnv *env, jobject obj, jmethodID methodID) {
	return (*env)->CallFloatMethod(env, obj, methodID);
}

static jint jni_CallIntMethod(JNIEnv *env, jobject obj, jmethodID methodID) {
	return (*env)->CallIntMethod(env, obj, methodID);
}

static void jni_CallStaticVoidMethodA(JNIEnv *env, jclass cls, jmethodID methodID, const jvalue *args) {
	(*env)->CallStaticVoidMethodA(env, cls, methodID, args);
}

static void jni_CallVoidMethodA(JNIEnv *env, jobject obj, jmethodID methodID, const jvalue *args) {
	(*env)->CallVoidMethodA(env, obj, methodID, args);
}

static jbyte *jni_GetByteArrayElements(JNIEnv *env, jbyteArray arr) {
	return (*env)->GetByteArrayElements(env, arr, NULL);
}

static void jni_ReleaseByteArrayElements(JNIEnv *env, jbyteArray arr, jbyte *bytes) {
	(*env)->ReleaseByteArrayElements(env, arr, bytes, JNI_ABORT);
}

static jsize jni_GetArrayLength(JNIEnv *env, jbyteArray arr) {
	return (*env)->GetArrayLength(env, arr);
}

static jstring jni_NewString(JNIEnv *env, const jchar *unicodeChars, jsize len) {
	return (*env)->NewString(env, unicodeChars, len);
}

static jsize jni_GetStringLength(JNIEnv *env, jstring str) {
	return (*env)->GetStringLength(env, str);
}

static const jchar *jni_GetStringChars(JNIEnv *env, jstring str) {
	return (*env)->GetStringChars(env, str, NULL);
}

static jthrowable jni_ExceptionOccurred(JNIEnv *env) {
	return (*env)->ExceptionOccurred(env);
}

static void jni_ExceptionClear(JNIEnv *env) {
	(*env)->ExceptionClear(env);
}

static jobject jni_CallObjectMethodA(JNIEnv *env, jobject obj, jmethodID method, jvalue *args) {
	return (*env)->CallObjectMethodA(env, obj, method, args);
}

static jobject jni_CallStaticObjectMethodA(JNIEnv *env, jclass cls, jmethodID method, jvalue *args) {
	return (*env)->CallStaticObjectMethodA(env, cls, method, args);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
	"unicode/utf16"
	"unsafe"

	"gioui.org/internal/f32color"

	"gioui.org/f32"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/unit"
)

type window struct {
	callbacks *callbacks

	view C.jobject

	dpi       int
	fontScale float32
	insets    system.Insets

	stage     system.Stage
	started   bool
	animating bool

	win    *C.ANativeWindow
	config Config
}

// gioView hold cached JNI methods for GioView.
var gioView struct {
	once               sync.Once
	getDensity         C.jmethodID
	getFontScale       C.jmethodID
	showTextInput      C.jmethodID
	hideTextInput      C.jmethodID
	setInputHint       C.jmethodID
	postInvalidate     C.jmethodID // requests draw, called from non-UI thread
	invalidate         C.jmethodID // requests draw, called from UI thread
	setCursor          C.jmethodID
	setOrientation     C.jmethodID
	setNavigationColor C.jmethodID
	setStatusColor     C.jmethodID
	setFullscreen      C.jmethodID
	unregister         C.jmethodID
}

// ViewEvent is sent whenever the Window's underlying Android view
// changes.
type ViewEvent struct {
	// View is a JNI global reference to the android.view.View
	// instance backing the Window. The reference is valid until
	// the next ViewEvent is received.
	// A zero View means that there is currently no view attached.
	View uintptr
}

type jvalue uint64 // The largest JNI type fits in 64 bits.

var dataDirChan = make(chan string, 1)

var android struct {
	// mu protects all fields of this structure. However, once a
	// non-nil jvm is returned from javaVM, all the other fields may
	// be accessed unlocked.
	mu  sync.Mutex
	jvm *C.JavaVM

	// appCtx is the global Android App context.
	appCtx C.jobject
	// gioCls is the class of the Gio class.
	gioCls C.jclass

	mwriteClipboard   C.jmethodID
	mreadClipboard    C.jmethodID
	mwakeupMainThread C.jmethodID
}

// view maps from GioView JNI refenreces to windows.
var views = make(map[C.jlong]*window)

var windows = make(map[*callbacks]*window)

var mainWindow = newWindowRendezvous()

var mainFuncs = make(chan func(env *C.JNIEnv), 1)

var (
	dataDirOnce sync.Once
	dataPath    string
)

var (
	newAndroidVulkanContext func(w *window) (context, error)
	newAndroidGLESContext   func(w *window) (context, error)
)

func (w *window) NewContext() (context, error) {
	funcs := []func(w *window) (context, error){newAndroidVulkanContext, newAndroidGLESContext}
	var firstErr error
	for _, f := range funcs {
		if f == nil {
			continue
		}
		c, err := f(w)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		return c, nil
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, errors.New("x11: no available GPU backends")
}

func dataDir() (string, error) {
	dataDirOnce.Do(func() {
		dataPath = <-dataDirChan
		// Set XDG_CACHE_HOME to make os.UserCacheDir work.
		if _, exists := os.LookupEnv("XDG_CACHE_HOME"); !exists {
			cachePath := filepath.Join(dataPath, "cache")
			os.Setenv("XDG_CACHE_HOME", cachePath)
		}
		// Set XDG_CONFIG_HOME to make os.UserConfigDir work.
		if _, exists := os.LookupEnv("XDG_CONFIG_HOME"); !exists {
			cfgPath := filepath.Join(dataPath, "config")
			os.Setenv("XDG_CONFIG_HOME", cfgPath)
		}
		// Set HOME to make os.UserHomeDir work.
		if _, exists := os.LookupEnv("HOME"); !exists {
			os.Setenv("HOME", dataPath)
		}
	})
	return dataPath, nil
}

func getMethodID(env *C.JNIEnv, class C.jclass, method, sig string) C.jmethodID {
	m := C.CString(method)
	defer C.free(unsafe.Pointer(m))
	s := C.CString(sig)
	defer C.free(unsafe.Pointer(s))
	jm := C.jni_GetMethodID(env, class, m, s)
	if err := exception(env); err != nil {
		panic(err)
	}
	return jm
}

func getStaticMethodID(env *C.JNIEnv, class C.jclass, method, sig string) C.jmethodID {
	m := C.CString(method)
	defer C.free(unsafe.Pointer(m))
	s := C.CString(sig)
	defer C.free(unsafe.Pointer(s))
	jm := C.jni_GetStaticMethodID(env, class, m, s)
	if err := exception(env); err != nil {
		panic(err)
	}
	return jm
}

//export Java_org_gioui_Gio_runGoMain
func Java_org_gioui_Gio_runGoMain(env *C.JNIEnv, class C.jclass, jdataDir C.jbyteArray, context C.jobject) {
	initJVM(env, class, context)
	dirBytes := C.jni_GetByteArrayElements(env, jdataDir)
	if dirBytes == nil {
		panic("runGoMain: GetByteArrayElements failed")
	}
	n := C.jni_GetArrayLength(env, jdataDir)
	dataDir := C.GoStringN((*C.char)(unsafe.Pointer(dirBytes)), n)
	dataDirChan <- dataDir
	C.jni_ReleaseByteArrayElements(env, jdataDir, dirBytes)

	runMain()
}

func initJVM(env *C.JNIEnv, gio C.jclass, ctx C.jobject) {
	android.mu.Lock()
	defer android.mu.Unlock()
	if res := C.jni_GetJavaVM(env, &android.jvm); res != 0 {
		panic("gio: GetJavaVM failed")
	}
	android.appCtx = C.jni_NewGlobalRef(env, ctx)
	android.gioCls = C.jclass(C.jni_NewGlobalRef(env, C.jobject(gio)))
	android.mwriteClipboard = getStaticMethodID(env, gio, "writeClipboard", "(Landroid/content/Context;Ljava/lang/String;)V")
	android.mreadClipboard = getStaticMethodID(env, gio, "readClipboard", "(Landroid/content/Context;)Ljava/lang/String;")
	android.mwakeupMainThread = getStaticMethodID(env, gio, "wakeupMainThread", "()V")
}

// JavaVM returns the global JNI JavaVM.
func JavaVM() uintptr {
	jvm := javaVM()
	return uintptr(unsafe.Pointer(jvm))
}

func javaVM() *C.JavaVM {
	android.mu.Lock()
	defer android.mu.Unlock()
	return android.jvm
}

// AppContext returns the global Application context as a JNI jobject.
func AppContext() uintptr {
	android.mu.Lock()
	defer android.mu.Unlock()
	return uintptr(android.appCtx)
}

//export Java_org_gioui_GioView_onCreateView
func Java_org_gioui_GioView_onCreateView(env *C.JNIEnv, class C.jclass, view C.jobject) C.jlong {
	gioView.once.Do(func() {
		m := &gioView
		m.getDensity = getMethodID(env, class, "getDensity", "()I")
		m.getFontScale = getMethodID(env, class, "getFontScale", "()F")
		m.showTextInput = getMethodID(env, class, "showTextInput", "()V")
		m.hideTextInput = getMethodID(env, class, "hideTextInput", "()V")
		m.setInputHint = getMethodID(env, class, "setInputHint", "(I)V")
		m.postInvalidate = getMethodID(env, class, "postInvalidate", "()V")
		m.invalidate = getMethodID(env, class, "invalidate", "()V")
		m.setCursor = getMethodID(env, class, "setCursor", "(I)V")
		m.setOrientation = getMethodID(env, class, "setOrientation", "(II)V")
		m.setNavigationColor = getMethodID(env, class, "setNavigationColor", "(II)V")
		m.setStatusColor = getMethodID(env, class, "setStatusColor", "(II)V")
		m.setFullscreen = getMethodID(env, class, "setFullscreen", "(Z)V")
		m.unregister = getMethodID(env, class, "unregister", "()V")
	})
	view = C.jni_NewGlobalRef(env, view)
	wopts := <-mainWindow.out
	w, ok := windows[wopts.window]
	if !ok {
		w = &window{
			callbacks: wopts.window,
		}
		windows[wopts.window] = w
	}
	if w.view != 0 {
		w.detach(env)
	}
	w.view = view
	w.callbacks.SetDriver(w)
	handle := C.jlong(view)
	views[handle] = w
	w.loadConfig(env, class)
	w.Configure(wopts.options)
	w.setStage(system.StagePaused)
	w.callbacks.Event(ViewEvent{View: uintptr(view)})
	return handle
}

//export Java_org_gioui_GioView_onDestroyView
func Java_org_gioui_GioView_onDestroyView(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := views[handle]
	w.detach(env)
}

//export Java_org_gioui_GioView_onStopView
func Java_org_gioui_GioView_onStopView(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := views[handle]
	w.started = false
	w.setStage(system.StagePaused)
}

//export Java_org_gioui_GioView_onStartView
func Java_org_gioui_GioView_onStartView(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := views[handle]
	w.started = true
	if w.win != nil {
		w.setVisible()
	}
}

//export Java_org_gioui_GioView_onSurfaceDestroyed
func Java_org_gioui_GioView_onSurfaceDestroyed(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := views[handle]
	w.win = nil
	w.setStage(system.StagePaused)
}

//export Java_org_gioui_GioView_onSurfaceChanged
func Java_org_gioui_GioView_onSurfaceChanged(env *C.JNIEnv, class C.jclass, handle C.jlong, surf C.jobject) {
	w := views[handle]
	w.win = C.ANativeWindow_fromSurface(env, surf)
	if w.started {
		w.setVisible()
	}
}

//export Java_org_gioui_GioView_onLowMemory
func Java_org_gioui_GioView_onLowMemory(env *C.JNIEnv, class C.jclass) {
	runtime.GC()
	debug.FreeOSMemory()
}

//export Java_org_gioui_GioView_onConfigurationChanged
func Java_org_gioui_GioView_onConfigurationChanged(env *C.JNIEnv, class C.jclass, view C.jlong) {
	w := views[view]
	w.loadConfig(env, class)
	if w.stage >= system.StageRunning {
		w.draw(true)
	}
}

//export Java_org_gioui_GioView_onFrameCallback
func Java_org_gioui_GioView_onFrameCallback(env *C.JNIEnv, class C.jclass, view C.jlong) {
	w, exist := views[view]
	if !exist {
		return
	}
	if w.stage < system.StageRunning {
		return
	}
	if w.animating {
		w.draw(false)
		// Schedule the next draw immediately after this one. Since onFrameCallback runs
		// on the UI thread, View.invalidate can be used here instead of postInvalidate.
		callVoidMethod(env, w.view, gioView.invalidate)
	}
}

//export Java_org_gioui_GioView_onBack
func Java_org_gioui_GioView_onBack(env *C.JNIEnv, class C.jclass, view C.jlong) C.jboolean {
	w := views[view]
	ev := &system.CommandEvent{Type: system.CommandBack}
	w.callbacks.Event(ev)
	if ev.Cancel {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_org_gioui_GioView_onFocusChange
func Java_org_gioui_GioView_onFocusChange(env *C.JNIEnv, class C.jclass, view C.jlong, focus C.jboolean) {
	w := views[view]
	w.callbacks.Event(key.FocusEvent{Focus: focus == C.JNI_TRUE})
}

//export Java_org_gioui_GioView_onWindowInsets
func Java_org_gioui_GioView_onWindowInsets(env *C.JNIEnv, class C.jclass, view C.jlong, top, right, bottom, left C.jint) {
	w := views[view]
	w.insets = system.Insets{
		Top:    unit.Px(float32(top)),
		Right:  unit.Px(float32(right)),
		Bottom: unit.Px(float32(bottom)),
		Left:   unit.Px(float32(left)),
	}
	if w.stage >= system.StageRunning {
		w.draw(true)
	}
}

func (w *window) detach(env *C.JNIEnv) {
	callVoidMethod(env, w.view, gioView.unregister)
	w.callbacks.Event(ViewEvent{})
	w.callbacks.SetDriver(nil)
	delete(views, C.jlong(w.view))
	C.jni_DeleteGlobalRef(env, w.view)
	w.view = 0
}

func (w *window) setVisible() {
	width, height := C.ANativeWindow_getWidth(w.win), C.ANativeWindow_getHeight(w.win)
	if width == 0 || height == 0 {
		return
	}
	w.setStage(system.StageRunning)
	w.draw(true)
}

func (w *window) setStage(stage system.Stage) {
	if stage == w.stage {
		return
	}
	w.stage = stage
	w.callbacks.Event(system.StageEvent{stage})
}

func (w *window) setVisual(visID int) error {
	if C.ANativeWindow_setBuffersGeometry(w.win, 0, 0, C.int32_t(visID)) != 0 {
		return errors.New("ANativeWindow_setBuffersGeometry failed")
	}
	return nil
}

func (w *window) nativeWindow() (*C.ANativeWindow, int, int) {
	width, height := C.ANativeWindow_getWidth(w.win), C.ANativeWindow_getHeight(w.win)
	return w.win, int(width), int(height)
}

func (w *window) loadConfig(env *C.JNIEnv, class C.jclass) {
	dpi := int(C.jni_CallIntMethod(env, w.view, gioView.getDensity))
	w.fontScale = float32(C.jni_CallFloatMethod(env, w.view, gioView.getFontScale))
	switch dpi {
	case C.ACONFIGURATION_DENSITY_NONE,
		C.ACONFIGURATION_DENSITY_DEFAULT,
		C.ACONFIGURATION_DENSITY_ANY:
		// Assume standard density.
		w.dpi = C.ACONFIGURATION_DENSITY_MEDIUM
	default:
		w.dpi = int(dpi)
	}
}

func (w *window) SetAnimating(anim bool) {
	w.animating = anim
	if anim {
		runInJVM(javaVM(), func(env *C.JNIEnv) {
			callVoidMethod(env, w.view, gioView.postInvalidate)
		})
	}
}

func (w *window) draw(sync bool) {
	size := image.Pt(int(C.ANativeWindow_getWidth(w.win)), int(C.ANativeWindow_getHeight(w.win)))
	if size != w.config.Size {
		w.config.Size = size
		w.callbacks.Event(ConfigEvent{Config: w.config})
	}
	if size.X == 0 || size.Y == 0 {
		return
	}
	const inchPrDp = 1.0 / 160
	ppdp := float32(w.dpi) * inchPrDp
	w.callbacks.Event(frameEvent{
		FrameEvent: system.FrameEvent{
			Now:    time.Now(),
			Size:   w.config.Size,
			Insets: w.insets,
			Metric: unit.Metric{
				PxPerDp: ppdp,
				PxPerSp: w.fontScale * ppdp,
			},
		},
		Sync: sync,
	})
}

type keyMapper func(devId, keyCode C.int32_t) rune

func runInJVM(jvm *C.JavaVM, f func(env *C.JNIEnv)) {
	if jvm == nil {
		panic("nil JVM")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var env *C.JNIEnv
	if res := C.jni_GetEnv(jvm, &env, C.JNI_VERSION_1_6); res != C.JNI_OK {
		if res != C.JNI_EDETACHED {
			panic(fmt.Errorf("JNI GetEnv failed with error %d", res))
		}
		if C.jni_AttachCurrentThread(jvm, &env, nil) != C.JNI_OK {
			panic(errors.New("runInJVM: AttachCurrentThread failed"))
		}
		defer C.jni_DetachCurrentThread(jvm)
	}

	f(env)
}

func convertKeyCode(code C.jint) (string, bool) {
	var n string
	switch code {
	case C.AKEYCODE_DPAD_UP:
		n = key.NameUpArrow
	case C.AKEYCODE_DPAD_DOWN:
		n = key.NameDownArrow
	case C.AKEYCODE_DPAD_LEFT:
		n = key.NameLeftArrow
	case C.AKEYCODE_DPAD_RIGHT:
		n = key.NameRightArrow
	case C.AKEYCODE_FORWARD_DEL:
		n = key.NameDeleteForward
	case C.AKEYCODE_DEL:
		n = key.NameDeleteBackward
	case C.AKEYCODE_NUMPAD_ENTER:
		n = key.NameEnter
	case C.AKEYCODE_ENTER:
		n = key.NameEnter
	default:
		return "", false
	}
	return n, true
}

//export Java_org_gioui_GioView_onKeyEvent
func Java_org_gioui_GioView_onKeyEvent(env *C.JNIEnv, class C.jclass, handle C.jlong, keyCode, r C.jint, t C.jlong) {
	w := views[handle]
	if n, ok := convertKeyCode(keyCode); ok {
		w.callbacks.Event(key.Event{Name: n})
	}
	if r != 0 && r != '\n' { // Checking for "\n" to prevent duplication with key.NameEnter (gio#224).
		w.callbacks.Event(key.EditEvent{Text: string(rune(r))})
	}
}

//export Java_org_gioui_GioView_onTouchEvent
func Java_org_gioui_GioView_onTouchEvent(env *C.JNIEnv, class C.jclass, handle C.jlong, action, pointerID, tool C.jint, x, y, scrollX, scrollY C.jfloat, jbtns C.jint, t C.jlong) {
	w := views[handle]
	var typ pointer.Type
	switch action {
	case C.AMOTION_EVENT_ACTION_DOWN, C.AMOTION_EVENT_ACTION_POINTER_DOWN:
		typ = pointer.Press
	case C.AMOTION_EVENT_ACTION_UP, C.AMOTION_EVENT_ACTION_POINTER_UP:
		typ = pointer.Release
	case C.AMOTION_EVENT_ACTION_CANCEL:
		typ = pointer.Cancel
	case C.AMOTION_EVENT_ACTION_MOVE:
		typ = pointer.Move
	case C.AMOTION_EVENT_ACTION_SCROLL:
		typ = pointer.Scroll
	default:
		return
	}
	var src pointer.Source
	var btns pointer.Buttons
	if jbtns&C.AMOTION_EVENT_BUTTON_PRIMARY != 0 {
		btns |= pointer.ButtonPrimary
	}
	if jbtns&C.AMOTION_EVENT_BUTTON_SECONDARY != 0 {
		btns |= pointer.ButtonSecondary
	}
	if jbtns&C.AMOTION_EVENT_BUTTON_TERTIARY != 0 {
		btns |= pointer.ButtonTertiary
	}
	switch tool {
	case C.AMOTION_EVENT_TOOL_TYPE_FINGER:
		src = pointer.Touch
	case C.AMOTION_EVENT_TOOL_TYPE_STYLUS:
		src = pointer.Touch
	case C.AMOTION_EVENT_TOOL_TYPE_MOUSE:
		src = pointer.Mouse
	case C.AMOTION_EVENT_TOOL_TYPE_UNKNOWN:
		// For example, triggered via 'adb shell input tap'.
		// Instead of discarding it, treat it as a touch event.
		src = pointer.Touch
	default:
		return
	}
	w.callbacks.Event(pointer.Event{
		Type:      typ,
		Source:    src,
		Buttons:   btns,
		PointerID: pointer.ID(pointerID),
		Time:      time.Duration(t) * time.Millisecond,
		Position:  f32.Point{X: float32(x), Y: float32(y)},
		Scroll:    f32.Pt(float32(scrollX), float32(scrollY)),
	})
}

func (w *window) ShowTextInput(show bool) {
	runInJVM(javaVM(), func(env *C.JNIEnv) {
		if show {
			callVoidMethod(env, w.view, gioView.showTextInput)
		} else {
			callVoidMethod(env, w.view, gioView.hideTextInput)
		}
	})
}

func (w *window) SetInputHint(mode key.InputHint) {
	// Constants defined at https://developer.android.com/reference/android/text/InputType.
	const (
		TYPE_NULL                            = 0
		TYPE_CLASS_NUMBER                    = 2
		TYPE_NUMBER_FLAG_DECIMAL             = 8192
		TYPE_NUMBER_FLAG_SIGNED              = 4096
		TYPE_TEXT_FLAG_NO_SUGGESTIONS        = 524288
		TYPE_TEXT_VARIATION_VISIBLE_PASSWORD = 144
	)

	runInJVM(javaVM(), func(env *C.JNIEnv) {
		var m jvalue
		switch mode {
		case key.HintNumeric:
			m = TYPE_CLASS_NUMBER | TYPE_NUMBER_FLAG_DECIMAL | TYPE_NUMBER_FLAG_SIGNED
		default:
			// TYPE_NULL, since TYPE_CLASS_TEXT isn't currently supported.
			m = TYPE_NULL
		}

		// The TYPE_TEXT_FLAG_NO_SUGGESTIONS and TYPE_TEXT_VARIATION_VISIBLE_PASSWORD are used to fix the
		// Samsung keyboard compatibility, forcing to disable the suggests/auto-complete. gio#116.
		m = m | TYPE_TEXT_FLAG_NO_SUGGESTIONS | TYPE_TEXT_VARIATION_VISIBLE_PASSWORD

		callVoidMethod(env, w.view, gioView.setInputHint, m)
	})
}

func javaString(env *C.JNIEnv, str string) C.jstring {
	if str == "" {
		return 0
	}
	utf16Chars := utf16.Encode([]rune(str))
	return C.jni_NewString(env, (*C.jchar)(unsafe.Pointer(&utf16Chars[0])), C.int(len(utf16Chars)))
}

func varArgs(args []jvalue) *C.jvalue {
	if len(args) == 0 {
		return nil
	}
	return (*C.jvalue)(unsafe.Pointer(&args[0]))
}

func callStaticVoidMethod(env *C.JNIEnv, cls C.jclass, method C.jmethodID, args ...jvalue) error {
	C.jni_CallStaticVoidMethodA(env, cls, method, varArgs(args))
	return exception(env)
}

func callStaticObjectMethod(env *C.JNIEnv, cls C.jclass, method C.jmethodID, args ...jvalue) (C.jobject, error) {
	res := C.jni_CallStaticObjectMethodA(env, cls, method, varArgs(args))
	return res, exception(env)
}

func callVoidMethod(env *C.JNIEnv, obj C.jobject, method C.jmethodID, args ...jvalue) error {
	C.jni_CallVoidMethodA(env, obj, method, varArgs(args))
	return exception(env)
}

func callObjectMethod(env *C.JNIEnv, obj C.jobject, method C.jmethodID, args ...jvalue) (C.jobject, error) {
	res := C.jni_CallObjectMethodA(env, obj, method, varArgs(args))
	return res, exception(env)
}

// exception returns an error corresponding to the pending
// exception, or nil if no exception is pending. The pending
// exception is cleared.
func exception(env *C.JNIEnv) error {
	thr := C.jni_ExceptionOccurred(env)
	if thr == 0 {
		return nil
	}
	C.jni_ExceptionClear(env)
	cls := getObjectClass(env, C.jobject(thr))
	toString := getMethodID(env, cls, "toString", "()Ljava/lang/String;")
	msg, err := callObjectMethod(env, C.jobject(thr), toString)
	if err != nil {
		return err
	}
	return errors.New(goString(env, C.jstring(msg)))
}

func getObjectClass(env *C.JNIEnv, obj C.jobject) C.jclass {
	if obj == 0 {
		panic("null object")
	}
	cls := C.jni_GetObjectClass(env, C.jobject(obj))
	if err := exception(env); err != nil {
		// GetObjectClass should never fail.
		panic(err)
	}
	return cls
}

// goString converts the JVM jstring to a Go string.
func goString(env *C.JNIEnv, str C.jstring) string {
	if str == 0 {
		return ""
	}
	strlen := C.jni_GetStringLength(env, C.jstring(str))
	chars := C.jni_GetStringChars(env, C.jstring(str))
	var utf16Chars []uint16
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&utf16Chars))
	hdr.Data = uintptr(unsafe.Pointer(chars))
	hdr.Cap = int(strlen)
	hdr.Len = int(strlen)
	utf8 := utf16.Decode(utf16Chars)
	return string(utf8)
}

func osMain() {
}

func newWindow(window *callbacks, options []Option) error {
	mainWindow.in <- windowAndConfig{window, options}
	return <-mainWindow.errs
}

func (w *window) WriteClipboard(s string) {
	runInJVM(javaVM(), func(env *C.JNIEnv) {
		jstr := javaString(env, s)
		callStaticVoidMethod(env, android.gioCls, android.mwriteClipboard,
			jvalue(android.appCtx), jvalue(jstr))
	})
}

func (w *window) ReadClipboard() {
	runInJVM(javaVM(), func(env *C.JNIEnv) {
		c, err := callStaticObjectMethod(env, android.gioCls, android.mreadClipboard,
			jvalue(android.appCtx))
		if err != nil {
			return
		}
		content := goString(env, C.jstring(c))
		w.callbacks.Event(clipboard.Event{Text: content})
	})
}

func (w *window) Configure(options []Option) {
	runInJVM(javaVM(), func(env *C.JNIEnv) {
		prev := w.config
		cnf := w.config
		cnf.apply(unit.Metric{}, options)
		if prev.Orientation != cnf.Orientation {
			w.config.Orientation = cnf.Orientation
			setOrientation(env, w.view, cnf.Orientation)
		}
		if prev.NavigationColor != cnf.NavigationColor {
			w.config.NavigationColor = cnf.NavigationColor
			setNavigationColor(env, w.view, cnf.NavigationColor)
		}
		if prev.StatusColor != cnf.StatusColor {
			w.config.StatusColor = cnf.StatusColor
			setStatusColor(env, w.view, cnf.StatusColor)
		}
		if prev.Mode != cnf.Mode {
			switch cnf.Mode {
			case Fullscreen:
				callVoidMethod(env, w.view, gioView.setFullscreen, C.JNI_TRUE)
				w.config.Mode = Fullscreen
			case Windowed:
				callVoidMethod(env, w.view, gioView.setFullscreen, C.JNI_FALSE)
				w.config.Mode = Windowed
			}
		}
		if w.config != prev {
			w.callbacks.Event(ConfigEvent{Config: w.config})
		}
	})
}

func (w *window) Raise() {}

func (w *window) SetCursor(name pointer.CursorName) {
	runInJVM(javaVM(), func(env *C.JNIEnv) {
		setCursor(env, w.view, name)
	})
}

func (w *window) Wakeup() {
	runOnMain(func(env *C.JNIEnv) {
		w.callbacks.Event(wakeupEvent{})
	})
}

func setCursor(env *C.JNIEnv, view C.jobject, name pointer.CursorName) {
	var curID int
	switch name {
	default:
		fallthrough
	case pointer.CursorDefault:
		curID = 1000 // TYPE_ARROW
	case pointer.CursorText:
		curID = 1008 // TYPE_TEXT
	case pointer.CursorPointer:
		curID = 1002 // TYPE_HAND
	case pointer.CursorCrossHair:
		curID = 1007 // TYPE_CROSSHAIR
	case pointer.CursorColResize:
		curID = 1014 // TYPE_HORIZONTAL_DOUBLE_ARROW
	case pointer.CursorRowResize:
		curID = 1015 // TYPE_VERTICAL_DOUBLE_ARROW
	case pointer.CursorNone:
		curID = 0 // TYPE_NULL
	}
	callVoidMethod(env, view, gioView.setCursor, jvalue(curID))
}

func setOrientation(env *C.JNIEnv, view C.jobject, mode Orientation) {
	var (
		id         int
		idFallback int // Used only for SDK 17 or older.
	)
	// Constants defined at https://developer.android.com/reference/android/content/pm/ActivityInfo.
	switch mode {
	case AnyOrientation:
		id, idFallback = 2, 2 // SCREEN_ORIENTATION_USER
	case LandscapeOrientation:
		id, idFallback = 11, 0 // SCREEN_ORIENTATION_USER_LANDSCAPE (or SCREEN_ORIENTATION_LANDSCAPE)
	case PortraitOrientation:
		id, idFallback = 12, 1 // SCREEN_ORIENTATION_USER_PORTRAIT (or SCREEN_ORIENTATION_PORTRAIT)
	}
	callVoidMethod(env, view, gioView.setOrientation, jvalue(id), jvalue(idFallback))
}

func setStatusColor(env *C.JNIEnv, view C.jobject, color color.NRGBA) {
	callVoidMethod(env, view, gioView.setStatusColor,
		jvalue(uint32(color.A)<<24|uint32(color.R)<<16|uint32(color.G)<<8|uint32(color.B)),
		jvalue(int(f32color.LinearFromSRGB(color).Luminance()*255)),
	)
}

func setNavigationColor(env *C.JNIEnv, view C.jobject, color color.NRGBA) {
	callVoidMethod(env, view, gioView.setNavigationColor,
		jvalue(uint32(color.A)<<24|uint32(color.R)<<16|uint32(color.G)<<8|uint32(color.B)),
		jvalue(int(f32color.LinearFromSRGB(color).Luminance()*255)),
	)
}

// Close the window. Not implemented for Android.
func (w *window) Close() {}

// Maximize maximizes the window. Not implemented for Android.
func (w *window) Maximize() {}

// Center the window. Not implemented for Android.
func (w *window) Center() {}

// runOnMain runs a function on the Java main thread.
func runOnMain(f func(env *C.JNIEnv)) {
	go func() {
		mainFuncs <- f
		runInJVM(javaVM(), func(env *C.JNIEnv) {
			callStaticVoidMethod(env, android.gioCls, android.mwakeupMainThread)
		})
	}()
}

//export Java_org_gioui_Gio_scheduleMainFuncs
func Java_org_gioui_Gio_scheduleMainFuncs(env *C.JNIEnv, cls C.jclass) {
	for {
		select {
		case f := <-mainFuncs:
			f(env)
		default:
			return
		}
	}
}

func (_ ViewEvent) ImplementsEvent() {}
