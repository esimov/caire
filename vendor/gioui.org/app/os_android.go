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

static jboolean jni_CallBooleanMethodA(JNIEnv *env, jobject obj, jmethodID methodID, const jvalue *args) {
	return (*env)->CallBooleanMethodA(env, obj, methodID, args);
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

static jclass jni_FindClass(JNIEnv *env, char *name) {
	return (*env)->FindClass(env, name);
}

static jobject jni_NewObjectA(JNIEnv *env, jclass cls, jmethodID cons, jvalue *args) {
	return (*env)->NewObjectA(env, cls, cons, args);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/cgo"
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
	"gioui.org/io/router"
	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/unit"
)

type window struct {
	callbacks *callbacks

	view   C.jobject
	handle cgo.Handle

	dpi       int
	fontScale float32
	insets    image.Rectangle

	stage     system.Stage
	started   bool
	animating bool

	win    *C.ANativeWindow
	config Config

	semantic struct {
		hoverID router.SemanticID
		rootID  router.SemanticID
		focusID router.SemanticID
		diffs   []router.SemanticID
	}
}

// gioView hold cached JNI methods for GioView.
var gioView struct {
	once               sync.Once
	getDensity         C.jmethodID
	getFontScale       C.jmethodID
	showTextInput      C.jmethodID
	hideTextInput      C.jmethodID
	setInputHint       C.jmethodID
	postFrameCallback  C.jmethodID
	invalidate         C.jmethodID // requests draw, called from UI thread
	setCursor          C.jmethodID
	setOrientation     C.jmethodID
	setNavigationColor C.jmethodID
	setStatusColor     C.jmethodID
	setFullscreen      C.jmethodID
	unregister         C.jmethodID
	sendA11yEvent      C.jmethodID
	sendA11yChange     C.jmethodID
	isA11yActive       C.jmethodID
	restartInput       C.jmethodID
	updateSelection    C.jmethodID
	updateCaret        C.jmethodID
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

	// android.view.accessibility.AccessibilityNodeInfo class.
	accessibilityNodeInfo struct {
		cls C.jclass
		// addChild(View, int)
		addChild C.jmethodID
		// setBoundsInScreen(Rect)
		setBoundsInScreen C.jmethodID
		// setText(CharSequence)
		setText C.jmethodID
		// setContentDescription(CharSequence)
		setContentDescription C.jmethodID
		// setParent(View, int)
		setParent C.jmethodID
		// addAction(int)
		addAction C.jmethodID
		// setClassName(CharSequence)
		setClassName C.jmethodID
		// setCheckable(boolean)
		setCheckable C.jmethodID
		// setSelected(boolean)
		setSelected C.jmethodID
		// setChecked(boolean)
		setChecked C.jmethodID
		// setEnabled(boolean)
		setEnabled C.jmethodID
		// setAccessibilityFocused(boolean)
		setAccessibilityFocused C.jmethodID
	}

	// android.graphics.Rect class.
	rect struct {
		cls C.jclass
		// (int, int, int, int) constructor.
		cons C.jmethodID
	}

	strings struct {
		// "android.view.View"
		androidViewView C.jstring
		// "android.widget.Button"
		androidWidgetButton C.jstring
		// "android.widget.CheckBox"
		androidWidgetCheckBox C.jstring
		// "android.widget.EditText"
		androidWidgetEditText C.jstring
		// "android.widget.RadioButton"
		androidWidgetRadioButton C.jstring
		// "android.widget.Switch"
		androidWidgetSwitch C.jstring
	}
}

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

// AccessibilityNodeProvider.HOST_VIEW_ID.
const HOST_VIEW_ID = -1

const (
	// AccessibilityEvent constants.
	TYPE_VIEW_HOVER_ENTER = 128
	TYPE_VIEW_HOVER_EXIT  = 256
)

const (
	// AccessibilityNodeInfo constants.
	ACTION_ACCESSIBILITY_FOCUS       = 64
	ACTION_CLEAR_ACCESSIBILITY_FOCUS = 128
	ACTION_CLICK                     = 16
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

	cls := findClass(env, "android/view/accessibility/AccessibilityNodeInfo")
	android.accessibilityNodeInfo.cls = C.jclass(C.jni_NewGlobalRef(env, C.jobject(cls)))
	android.accessibilityNodeInfo.addChild = getMethodID(env, cls, "addChild", "(Landroid/view/View;I)V")
	android.accessibilityNodeInfo.setBoundsInScreen = getMethodID(env, cls, "setBoundsInScreen", "(Landroid/graphics/Rect;)V")
	android.accessibilityNodeInfo.setText = getMethodID(env, cls, "setText", "(Ljava/lang/CharSequence;)V")
	android.accessibilityNodeInfo.setContentDescription = getMethodID(env, cls, "setContentDescription", "(Ljava/lang/CharSequence;)V")
	android.accessibilityNodeInfo.setParent = getMethodID(env, cls, "setParent", "(Landroid/view/View;I)V")
	android.accessibilityNodeInfo.addAction = getMethodID(env, cls, "addAction", "(I)V")
	android.accessibilityNodeInfo.setClassName = getMethodID(env, cls, "setClassName", "(Ljava/lang/CharSequence;)V")
	android.accessibilityNodeInfo.setCheckable = getMethodID(env, cls, "setCheckable", "(Z)V")
	android.accessibilityNodeInfo.setSelected = getMethodID(env, cls, "setSelected", "(Z)V")
	android.accessibilityNodeInfo.setChecked = getMethodID(env, cls, "setChecked", "(Z)V")
	android.accessibilityNodeInfo.setEnabled = getMethodID(env, cls, "setEnabled", "(Z)V")
	android.accessibilityNodeInfo.setAccessibilityFocused = getMethodID(env, cls, "setAccessibilityFocused", "(Z)V")

	cls = findClass(env, "android/graphics/Rect")
	android.rect.cls = C.jclass(C.jni_NewGlobalRef(env, C.jobject(cls)))
	android.rect.cons = getMethodID(env, cls, "<init>", "(IIII)V")
	android.mwriteClipboard = getStaticMethodID(env, gio, "writeClipboard", "(Landroid/content/Context;Ljava/lang/String;)V")
	android.mreadClipboard = getStaticMethodID(env, gio, "readClipboard", "(Landroid/content/Context;)Ljava/lang/String;")
	android.mwakeupMainThread = getStaticMethodID(env, gio, "wakeupMainThread", "()V")

	intern := func(s string) C.jstring {
		ref := C.jni_NewGlobalRef(env, C.jobject(javaString(env, s)))
		return C.jstring(ref)
	}
	android.strings.androidViewView = intern("android.view.View")
	android.strings.androidWidgetButton = intern("android.widget.Button")
	android.strings.androidWidgetCheckBox = intern("android.widget.CheckBox")
	android.strings.androidWidgetEditText = intern("android.widget.EditText")
	android.strings.androidWidgetRadioButton = intern("android.widget.RadioButton")
	android.strings.androidWidgetSwitch = intern("android.widget.Switch")
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
		m.postFrameCallback = getMethodID(env, class, "postFrameCallback", "()V")
		m.invalidate = getMethodID(env, class, "invalidate", "()V")
		m.setCursor = getMethodID(env, class, "setCursor", "(I)V")
		m.setOrientation = getMethodID(env, class, "setOrientation", "(II)V")
		m.setNavigationColor = getMethodID(env, class, "setNavigationColor", "(II)V")
		m.setStatusColor = getMethodID(env, class, "setStatusColor", "(II)V")
		m.setFullscreen = getMethodID(env, class, "setFullscreen", "(Z)V")
		m.unregister = getMethodID(env, class, "unregister", "()V")
		m.sendA11yEvent = getMethodID(env, class, "sendA11yEvent", "(II)V")
		m.sendA11yChange = getMethodID(env, class, "sendA11yChange", "(I)V")
		m.isA11yActive = getMethodID(env, class, "isA11yActive", "()Z")
		m.restartInput = getMethodID(env, class, "restartInput", "()V")
		m.updateSelection = getMethodID(env, class, "updateSelection", "()V")
		m.updateCaret = getMethodID(env, class, "updateCaret", "(FFFFFFFFFF)V")
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
	w.handle = cgo.NewHandle(w)
	w.callbacks.SetDriver(w)
	w.loadConfig(env, class)
	w.Configure(wopts.options)
	w.SetInputHint(key.HintAny)
	w.setStage(system.StagePaused)
	w.callbacks.Event(ViewEvent{View: uintptr(view)})
	return C.jlong(w.handle)
}

//export Java_org_gioui_GioView_onDestroyView
func Java_org_gioui_GioView_onDestroyView(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := cgo.Handle(handle).Value().(*window)
	w.detach(env)
}

//export Java_org_gioui_GioView_onStopView
func Java_org_gioui_GioView_onStopView(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := cgo.Handle(handle).Value().(*window)
	w.started = false
	w.setStage(system.StagePaused)
}

//export Java_org_gioui_GioView_onStartView
func Java_org_gioui_GioView_onStartView(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := cgo.Handle(handle).Value().(*window)
	w.started = true
	if w.win != nil {
		w.setVisible(env)
	}
}

//export Java_org_gioui_GioView_onSurfaceDestroyed
func Java_org_gioui_GioView_onSurfaceDestroyed(env *C.JNIEnv, class C.jclass, handle C.jlong) {
	w := cgo.Handle(handle).Value().(*window)
	w.win = nil
	w.setStage(system.StagePaused)
}

//export Java_org_gioui_GioView_onSurfaceChanged
func Java_org_gioui_GioView_onSurfaceChanged(env *C.JNIEnv, class C.jclass, handle C.jlong, surf C.jobject) {
	w := cgo.Handle(handle).Value().(*window)
	w.win = C.ANativeWindow_fromSurface(env, surf)
	if w.started {
		w.setVisible(env)
	}
}

//export Java_org_gioui_GioView_onLowMemory
func Java_org_gioui_GioView_onLowMemory(env *C.JNIEnv, class C.jclass) {
	runtime.GC()
	debug.FreeOSMemory()
}

//export Java_org_gioui_GioView_onConfigurationChanged
func Java_org_gioui_GioView_onConfigurationChanged(env *C.JNIEnv, class C.jclass, view C.jlong) {
	w := cgo.Handle(view).Value().(*window)
	w.loadConfig(env, class)
	if w.stage >= system.StageRunning {
		w.draw(env, true)
	}
}

//export Java_org_gioui_GioView_onFrameCallback
func Java_org_gioui_GioView_onFrameCallback(env *C.JNIEnv, class C.jclass, view C.jlong) {
	w, exist := cgo.Handle(view).Value().(*window)
	if !exist {
		return
	}
	if w.stage < system.StageRunning {
		return
	}
	if w.animating {
		w.draw(env, false)
		callVoidMethod(env, w.view, gioView.postFrameCallback)
	}
}

//export Java_org_gioui_GioView_onBack
func Java_org_gioui_GioView_onBack(env *C.JNIEnv, class C.jclass, view C.jlong) C.jboolean {
	w := cgo.Handle(view).Value().(*window)
	if w.callbacks.Event(key.Event{Name: key.NameBack}) {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_org_gioui_GioView_onFocusChange
func Java_org_gioui_GioView_onFocusChange(env *C.JNIEnv, class C.jclass, view C.jlong, focus C.jboolean) {
	w := cgo.Handle(view).Value().(*window)
	w.callbacks.Event(key.FocusEvent{Focus: focus == C.JNI_TRUE})
}

//export Java_org_gioui_GioView_onWindowInsets
func Java_org_gioui_GioView_onWindowInsets(env *C.JNIEnv, class C.jclass, view C.jlong, top, right, bottom, left C.jint) {
	w := cgo.Handle(view).Value().(*window)
	w.insets = image.Rect(int(left), int(top), int(right), int(bottom))
	if w.stage >= system.StageRunning {
		w.draw(env, true)
	}
}

//export Java_org_gioui_GioView_initializeAccessibilityNodeInfo
func Java_org_gioui_GioView_initializeAccessibilityNodeInfo(env *C.JNIEnv, class C.jclass, view C.jlong, virtID, screenX, screenY C.jint, info C.jobject) C.jobject {
	w := cgo.Handle(view).Value().(*window)
	semID := w.semIDFor(virtID)
	sem, found := w.callbacks.LookupSemantic(semID)
	if found {
		off := image.Pt(int(screenX), int(screenY))
		if err := w.initAccessibilityNodeInfo(env, sem, off, info); err != nil {
			panic(err)
		}
	}
	return info
}

//export Java_org_gioui_GioView_onTouchExploration
func Java_org_gioui_GioView_onTouchExploration(env *C.JNIEnv, class C.jclass, view C.jlong, x, y C.jfloat) {
	w := cgo.Handle(view).Value().(*window)
	semID, _ := w.callbacks.SemanticAt(f32.Pt(float32(x), float32(y)))
	if w.semantic.hoverID == semID {
		return
	}
	// Android expects ENTER before EXIT.
	if semID != 0 {
		callVoidMethod(env, w.view, gioView.sendA11yEvent, TYPE_VIEW_HOVER_ENTER, jvalue(w.virtualIDFor(semID)))
	}
	if prevID := w.semantic.hoverID; prevID != 0 {
		callVoidMethod(env, w.view, gioView.sendA11yEvent, TYPE_VIEW_HOVER_EXIT, jvalue(w.virtualIDFor(prevID)))
	}
	w.semantic.hoverID = semID
}

//export Java_org_gioui_GioView_onExitTouchExploration
func Java_org_gioui_GioView_onExitTouchExploration(env *C.JNIEnv, class C.jclass, view C.jlong) {
	w := cgo.Handle(view).Value().(*window)
	if w.semantic.hoverID != 0 {
		callVoidMethod(env, w.view, gioView.sendA11yEvent, TYPE_VIEW_HOVER_EXIT, jvalue(w.virtualIDFor(w.semantic.hoverID)))
		w.semantic.hoverID = 0
	}
}

//export Java_org_gioui_GioView_onA11yFocus
func Java_org_gioui_GioView_onA11yFocus(env *C.JNIEnv, class C.jclass, view C.jlong, virtID C.jint) {
	w := cgo.Handle(view).Value().(*window)
	if semID := w.semIDFor(virtID); semID != w.semantic.focusID {
		w.semantic.focusID = semID
		// Android needs invalidate to refresh the TalkBack focus indicator.
		callVoidMethod(env, w.view, gioView.invalidate)
	}
}

//export Java_org_gioui_GioView_onClearA11yFocus
func Java_org_gioui_GioView_onClearA11yFocus(env *C.JNIEnv, class C.jclass, view C.jlong, virtID C.jint) {
	w := cgo.Handle(view).Value().(*window)
	if w.semantic.focusID == w.semIDFor(virtID) {
		w.semantic.focusID = 0
	}
}

func (w *window) initAccessibilityNodeInfo(env *C.JNIEnv, sem router.SemanticNode, off image.Point, info C.jobject) error {
	for _, ch := range sem.Children {
		err := callVoidMethod(env, info, android.accessibilityNodeInfo.addChild, jvalue(w.view), jvalue(w.virtualIDFor(ch.ID)))
		if err != nil {
			return err
		}
	}
	if sem.ParentID != 0 {
		if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setParent, jvalue(w.view), jvalue(w.virtualIDFor(sem.ParentID))); err != nil {
			return err
		}
		b := sem.Desc.Bounds.Add(off)
		rect, err := newObject(env, android.rect.cls, android.rect.cons,
			jvalue(b.Min.X),
			jvalue(b.Min.Y),
			jvalue(b.Max.X),
			jvalue(b.Max.Y),
		)
		if err != nil {
			return err
		}
		if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setBoundsInScreen, jvalue(rect)); err != nil {
			return err
		}
	}
	d := sem.Desc
	if l := d.Label; l != "" {
		jlbl := javaString(env, l)
		if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setText, jvalue(jlbl)); err != nil {
			return err
		}
	}
	if d.Description != "" {
		jd := javaString(env, d.Description)
		if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setContentDescription, jvalue(jd)); err != nil {
			return err
		}
	}
	addAction := func(act C.jint) {
		if err := callVoidMethod(env, info, android.accessibilityNodeInfo.addAction, jvalue(act)); err != nil {
			panic(err)
		}
	}
	if d.Gestures&router.ClickGesture != 0 {
		addAction(ACTION_CLICK)
	}
	clsName := android.strings.androidViewView
	selectMethod := android.accessibilityNodeInfo.setChecked
	checkable := false
	switch d.Class {
	case semantic.Button:
		clsName = android.strings.androidWidgetButton
	case semantic.CheckBox:
		checkable = true
		clsName = android.strings.androidWidgetCheckBox
	case semantic.Editor:
		clsName = android.strings.androidWidgetEditText
	case semantic.RadioButton:
		selectMethod = android.accessibilityNodeInfo.setSelected
		clsName = android.strings.androidWidgetRadioButton
	case semantic.Switch:
		checkable = true
		clsName = android.strings.androidWidgetSwitch
	}
	if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setClassName, jvalue(clsName)); err != nil {
		panic(err)
	}
	if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setCheckable, jvalue(javaBool(checkable))); err != nil {
		panic(err)
	}
	if err := callVoidMethod(env, info, selectMethod, jvalue(javaBool(d.Selected))); err != nil {
		panic(err)
	}
	if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setEnabled, jvalue(javaBool(!d.Disabled))); err != nil {
		panic(err)
	}
	isFocus := w.semantic.focusID == sem.ID
	if err := callVoidMethod(env, info, android.accessibilityNodeInfo.setAccessibilityFocused, jvalue(javaBool(isFocus))); err != nil {
		panic(err)
	}
	if isFocus {
		addAction(ACTION_CLEAR_ACCESSIBILITY_FOCUS)
	} else {
		addAction(ACTION_ACCESSIBILITY_FOCUS)
	}
	return nil
}

func (w *window) virtualIDFor(id router.SemanticID) C.jint {
	// TODO: Android virtual IDs are 32-bit Java integers, but childID is a int64.
	if id == w.semantic.rootID {
		return HOST_VIEW_ID
	}
	return C.jint(id)
}

func (w *window) semIDFor(virtID C.jint) router.SemanticID {
	if virtID == HOST_VIEW_ID {
		return w.semantic.rootID
	}
	return router.SemanticID(virtID)
}

func (w *window) detach(env *C.JNIEnv) {
	callVoidMethod(env, w.view, gioView.unregister)
	w.callbacks.Event(ViewEvent{})
	w.callbacks.SetDriver(nil)
	w.handle.Delete()
	C.jni_DeleteGlobalRef(env, w.view)
	w.view = 0
}

func (w *window) setVisible(env *C.JNIEnv) {
	width, height := C.ANativeWindow_getWidth(w.win), C.ANativeWindow_getHeight(w.win)
	if width == 0 || height == 0 {
		return
	}
	w.setStage(system.StageRunning)
	w.draw(env, true)
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
			callVoidMethod(env, w.view, gioView.postFrameCallback)
		})
	}
}

func (w *window) draw(env *C.JNIEnv, sync bool) {
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
	dppp := unit.Dp(1.0 / ppdp)
	insets := system.Insets{
		Top:    unit.Dp(w.insets.Min.Y) * dppp,
		Bottom: unit.Dp(w.insets.Max.Y) * dppp,
		Left:   unit.Dp(w.insets.Min.X) * dppp,
		Right:  unit.Dp(w.insets.Max.X) * dppp,
	}
	w.callbacks.Event(frameEvent{
		FrameEvent: system.FrameEvent{
			Now:    time.Now(),
			Size:   w.config.Size,
			Insets: insets,
			Metric: unit.Metric{
				PxPerDp: ppdp,
				PxPerSp: w.fontScale * ppdp,
			},
		},
		Sync: sync,
	})
	a11yActive, err := callBooleanMethod(env, w.view, gioView.isA11yActive)
	if err != nil {
		panic(err)
	}
	if a11yActive {
		if newR, oldR := w.callbacks.SemanticRoot(), w.semantic.rootID; newR != oldR {
			// Remap focus and hover.
			if oldR == w.semantic.hoverID {
				w.semantic.hoverID = newR
			}
			if oldR == w.semantic.focusID {
				w.semantic.focusID = newR
			}
			w.semantic.rootID = newR
			callVoidMethod(env, w.view, gioView.sendA11yChange, jvalue(w.virtualIDFor(newR)))
		}
		w.semantic.diffs = w.callbacks.AppendSemanticDiffs(w.semantic.diffs[:0])
		for _, id := range w.semantic.diffs {
			callVoidMethod(env, w.view, gioView.sendA11yChange, jvalue(w.virtualIDFor(id)))
		}
	}
}

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
	case C.AKEYCODE_FORWARD_DEL:
		n = key.NameDeleteForward
	case C.AKEYCODE_DEL:
		n = key.NameDeleteBackward
	case C.AKEYCODE_NUMPAD_ENTER:
		n = key.NameEnter
	case C.AKEYCODE_ENTER:
		n = key.NameReturn
	case C.AKEYCODE_CTRL_LEFT, C.AKEYCODE_CTRL_RIGHT:
		n = key.NameCtrl
	case C.AKEYCODE_SHIFT_LEFT, C.AKEYCODE_SHIFT_RIGHT:
		n = key.NameShift
	case C.AKEYCODE_ALT_LEFT, C.AKEYCODE_ALT_RIGHT:
		n = key.NameAlt
	case C.AKEYCODE_META_LEFT, C.AKEYCODE_META_RIGHT:
		n = key.NameSuper
	case C.AKEYCODE_DPAD_UP:
		n = key.NameUpArrow
	case C.AKEYCODE_DPAD_DOWN:
		n = key.NameDownArrow
	case C.AKEYCODE_DPAD_LEFT:
		n = key.NameLeftArrow
	case C.AKEYCODE_DPAD_RIGHT:
		n = key.NameRightArrow
	default:
		return "", false
	}
	return n, true
}

//export Java_org_gioui_GioView_onKeyEvent
func Java_org_gioui_GioView_onKeyEvent(env *C.JNIEnv, class C.jclass, handle C.jlong, keyCode, r C.jint, pressed C.jboolean, t C.jlong) {
	w := cgo.Handle(handle).Value().(*window)
	if pressed == C.JNI_TRUE && keyCode == C.AKEYCODE_DPAD_CENTER {
		w.callbacks.ClickFocus()
		return
	}
	if n, ok := convertKeyCode(keyCode); ok {
		state := key.Release
		if pressed == C.JNI_TRUE {
			state = key.Press
		}
		w.callbacks.Event(key.Event{Name: n, State: state})
	}
	if pressed == C.JNI_TRUE && r != 0 && r != '\n' { // Checking for "\n" to prevent duplication with key.NameEnter (gio#224).
		w.callbacks.EditorInsert(string(rune(r)))
	}
}

//export Java_org_gioui_GioView_onTouchEvent
func Java_org_gioui_GioView_onTouchEvent(env *C.JNIEnv, class C.jclass, handle C.jlong, action, pointerID, tool C.jint, x, y, scrollX, scrollY C.jfloat, jbtns C.jint, t C.jlong) {
	w := cgo.Handle(handle).Value().(*window)
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

//export Java_org_gioui_GioView_imeSelectionStart
func Java_org_gioui_GioView_imeSelectionStart(env *C.JNIEnv, class C.jclass, handle C.jlong) C.jint {
	w := cgo.Handle(handle).Value().(*window)
	sel := w.callbacks.EditorState().Selection
	start := sel.Start
	if sel.End < sel.Start {
		start = sel.End
	}
	return C.jint(start)
}

//export Java_org_gioui_GioView_imeSelectionEnd
func Java_org_gioui_GioView_imeSelectionEnd(env *C.JNIEnv, class C.jclass, handle C.jlong) C.jint {
	w := cgo.Handle(handle).Value().(*window)
	sel := w.callbacks.EditorState().Selection
	end := sel.End
	if sel.End < sel.Start {
		end = sel.Start
	}
	return C.jint(end)
}

//export Java_org_gioui_GioView_imeComposingStart
func Java_org_gioui_GioView_imeComposingStart(env *C.JNIEnv, class C.jclass, handle C.jlong) C.jint {
	w := cgo.Handle(handle).Value().(*window)
	comp := w.callbacks.EditorState().compose
	start := comp.Start
	if e := comp.End; e < start {
		start = e
	}
	return C.jint(start)
}

//export Java_org_gioui_GioView_imeComposingEnd
func Java_org_gioui_GioView_imeComposingEnd(env *C.JNIEnv, class C.jclass, handle C.jlong) C.jint {
	w := cgo.Handle(handle).Value().(*window)
	comp := w.callbacks.EditorState().compose
	end := comp.End
	if s := comp.Start; s > end {
		end = s
	}
	return C.jint(end)
}

//export Java_org_gioui_GioView_imeSnippet
func Java_org_gioui_GioView_imeSnippet(env *C.JNIEnv, class C.jclass, handle C.jlong) C.jstring {
	w := cgo.Handle(handle).Value().(*window)
	snip := w.callbacks.EditorState().Snippet.Text
	return javaString(env, snip)
}

//export Java_org_gioui_GioView_imeSnippetStart
func Java_org_gioui_GioView_imeSnippetStart(env *C.JNIEnv, class C.jclass, handle C.jlong) C.jint {
	w := cgo.Handle(handle).Value().(*window)
	return C.jint(w.callbacks.EditorState().Snippet.Start)
}

//export Java_org_gioui_GioView_imeSetSnippet
func Java_org_gioui_GioView_imeSetSnippet(env *C.JNIEnv, class C.jclass, handle C.jlong, start, end C.jint) {
	w := cgo.Handle(handle).Value().(*window)
	r := key.Range{Start: int(start), End: int(end)}
	w.callbacks.SetEditorSnippet(r)
}

//export Java_org_gioui_GioView_imeSetSelection
func Java_org_gioui_GioView_imeSetSelection(env *C.JNIEnv, class C.jclass, handle C.jlong, start, end C.jint) {
	w := cgo.Handle(handle).Value().(*window)
	r := key.Range{Start: int(start), End: int(end)}
	w.callbacks.SetEditorSelection(r)
}

//export Java_org_gioui_GioView_imeSetComposingRegion
func Java_org_gioui_GioView_imeSetComposingRegion(env *C.JNIEnv, class C.jclass, handle C.jlong, start, end C.jint) {
	w := cgo.Handle(handle).Value().(*window)
	w.callbacks.SetComposingRegion(key.Range{
		Start: int(start),
		End:   int(end),
	})
}

//export Java_org_gioui_GioView_imeReplace
func Java_org_gioui_GioView_imeReplace(env *C.JNIEnv, class C.jclass, handle C.jlong, start, end C.jint, jtext C.jstring) {
	w := cgo.Handle(handle).Value().(*window)
	r := key.Range{Start: int(start), End: int(end)}
	text := goString(env, jtext)
	w.callbacks.EditorReplace(r, text)
}

//export Java_org_gioui_GioView_imeToRunes
func Java_org_gioui_GioView_imeToRunes(env *C.JNIEnv, class C.jclass, handle C.jlong, chars C.jint) C.jint {
	w := cgo.Handle(handle).Value().(*window)
	state := w.callbacks.EditorState()
	return C.jint(state.RunesIndex(int(chars)))
}

//export Java_org_gioui_GioView_imeToUTF16
func Java_org_gioui_GioView_imeToUTF16(env *C.JNIEnv, class C.jclass, handle C.jlong, runes C.jint) C.jint {
	w := cgo.Handle(handle).Value().(*window)
	state := w.callbacks.EditorState()
	return C.jint(state.UTF16Index(int(runes)))
}

func (w *window) EditorStateChanged(old, new editorState) {
	runInJVM(javaVM(), func(env *C.JNIEnv) {
		if old.Snippet != new.Snippet {
			callVoidMethod(env, w.view, gioView.restartInput)
			return
		}
		if old.Selection.Range != new.Selection.Range {
			w.callbacks.SetComposingRegion(key.Range{Start: -1, End: -1})
			callVoidMethod(env, w.view, gioView.updateSelection)
		}
		if old.Selection.Transform != new.Selection.Transform || old.Selection.Caret != new.Selection.Caret {
			sel := new.Selection
			m00, m01, m02, m10, m11, m12 := sel.Transform.Elems()
			f := func(v float32) jvalue {
				return jvalue(math.Float32bits(v))
			}
			c := sel.Caret
			callVoidMethod(env, w.view, gioView.updateCaret, f(m00), f(m01), f(m02), f(m10), f(m11), f(m12), f(c.Pos.X), f(c.Pos.Y-c.Ascent), f(c.Pos.Y), f(c.Pos.Y+c.Descent))
		}
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
		TYPE_NULL = 0

		TYPE_CLASS_TEXT                   = 1
		TYPE_TEXT_VARIATION_EMAIL_ADDRESS = 32
		TYPE_TEXT_VARIATION_URI           = 16
		TYPE_TEXT_FLAG_CAP_SENTENCES      = 16384
		TYPE_TEXT_FLAG_AUTO_CORRECT       = 32768

		TYPE_CLASS_NUMBER        = 2
		TYPE_NUMBER_FLAG_DECIMAL = 8192
		TYPE_NUMBER_FLAG_SIGNED  = 4096

		TYPE_CLASS_PHONE = 3
	)

	runInJVM(javaVM(), func(env *C.JNIEnv) {
		var m jvalue
		switch mode {
		case key.HintText:
			m = TYPE_CLASS_TEXT | TYPE_TEXT_FLAG_AUTO_CORRECT | TYPE_TEXT_FLAG_CAP_SENTENCES
		case key.HintNumeric:
			m = TYPE_CLASS_NUMBER | TYPE_NUMBER_FLAG_DECIMAL | TYPE_NUMBER_FLAG_SIGNED
		case key.HintEmail:
			m = TYPE_CLASS_TEXT | TYPE_TEXT_VARIATION_EMAIL_ADDRESS
		case key.HintURL:
			m = TYPE_CLASS_TEXT | TYPE_TEXT_VARIATION_URI
		case key.HintTelephone:
			m = TYPE_CLASS_PHONE
		default:
			m = TYPE_CLASS_TEXT
		}

		callVoidMethod(env, w.view, gioView.setInputHint, m)
	})
}

func javaBool(b bool) C.jboolean {
	if b {
		return C.JNI_TRUE
	} else {
		return C.JNI_FALSE
	}
}

func javaString(env *C.JNIEnv, str string) C.jstring {
	utf16Chars := utf16.Encode([]rune(str))
	var ptr *C.jchar
	if len(utf16Chars) > 0 {
		ptr = (*C.jchar)(unsafe.Pointer(&utf16Chars[0]))
	}
	return C.jni_NewString(env, ptr, C.int(len(utf16Chars)))
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

func callBooleanMethod(env *C.JNIEnv, obj C.jobject, method C.jmethodID, args ...jvalue) (bool, error) {
	res := C.jni_CallBooleanMethodA(env, obj, method, varArgs(args))
	return res == C.JNI_TRUE, exception(env)
}

func callObjectMethod(env *C.JNIEnv, obj C.jobject, method C.jmethodID, args ...jvalue) (C.jobject, error) {
	res := C.jni_CallObjectMethodA(env, obj, method, varArgs(args))
	return res, exception(env)
}

func newObject(env *C.JNIEnv, cls C.jclass, method C.jmethodID, args ...jvalue) (C.jobject, error) {
	res := C.jni_NewObjectA(env, cls, method, varArgs(args))
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
	utf16Chars := unsafe.Slice((*uint16)(unsafe.Pointer(chars)), strlen)
	utf8 := utf16.Decode(utf16Chars)
	return string(utf8)
}

func findClass(env *C.JNIEnv, name string) C.jclass {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	return C.jni_FindClass(env, cn)
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
		// Decorations are never disabled.
		cnf.Decorated = true

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
		if cnf.Decorated != prev.Decorated {
			w.config.Decorated = cnf.Decorated
		}
		w.callbacks.Event(ConfigEvent{Config: w.config})
	})
}

func (w *window) Perform(system.Action) {}

func (w *window) SetCursor(cursor pointer.Cursor) {
	runInJVM(javaVM(), func(env *C.JNIEnv) {
		setCursor(env, w.view, cursor)
	})
}

func (w *window) Wakeup() {
	runOnMain(func(env *C.JNIEnv) {
		w.callbacks.Event(wakeupEvent{})
	})
}

var androidCursor = [...]uint16{
	pointer.CursorDefault:                  1000, // TYPE_ARROW
	pointer.CursorNone:                     0,
	pointer.CursorText:                     1008, // TYPE_TEXT
	pointer.CursorVerticalText:             1009, // TYPE_VERTICAL_TEXT
	pointer.CursorPointer:                  1002, // TYPE_HAND
	pointer.CursorCrosshair:                1007, // TYPE_CROSSHAIR
	pointer.CursorAllScroll:                1013, // TYPE_ALL_SCROLL
	pointer.CursorColResize:                1014, // TYPE_HORIZONTAL_DOUBLE_ARROW
	pointer.CursorRowResize:                1015, // TYPE_VERTICAL_DOUBLE_ARROW
	pointer.CursorGrab:                     1020, // TYPE_GRAB
	pointer.CursorGrabbing:                 1021, // TYPE_GRABBING
	pointer.CursorNotAllowed:               1012, // TYPE_NO_DROP
	pointer.CursorWait:                     1004, // TYPE_WAIT
	pointer.CursorProgress:                 1000, // TYPE_ARROW
	pointer.CursorNorthWestResize:          1017, // TYPE_TOP_LEFT_DIAGONAL_DOUBLE_ARROW
	pointer.CursorNorthEastResize:          1016, // TYPE_TOP_RIGHT_DIAGONAL_DOUBLE_ARROW
	pointer.CursorSouthWestResize:          1016, // TYPE_TOP_RIGHT_DIAGONAL_DOUBLE_ARROW
	pointer.CursorSouthEastResize:          1017, // TYPE_TOP_LEFT_DIAGONAL_DOUBLE_ARROW
	pointer.CursorNorthSouthResize:         1015, // TYPE_VERTICAL_DOUBLE_ARROW
	pointer.CursorEastWestResize:           1014, // TYPE_HORIZONTAL_DOUBLE_ARROW
	pointer.CursorWestResize:               1014, // TYPE_HORIZONTAL_DOUBLE_ARROW
	pointer.CursorEastResize:               1014, // TYPE_HORIZONTAL_DOUBLE_ARROW
	pointer.CursorNorthResize:              1015, // TYPE_VERTICAL_DOUBLE_ARROW
	pointer.CursorSouthResize:              1015, // TYPE_VERTICAL_DOUBLE_ARROW
	pointer.CursorNorthEastSouthWestResize: 1016, // TYPE_TOP_RIGHT_DIAGONAL_DOUBLE_ARROW
	pointer.CursorNorthWestSouthEastResize: 1017, // TYPE_TOP_LEFT_DIAGONAL_DOUBLE_ARROW
}

func setCursor(env *C.JNIEnv, view C.jobject, cursor pointer.Cursor) {
	curID := androidCursor[cursor]
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
