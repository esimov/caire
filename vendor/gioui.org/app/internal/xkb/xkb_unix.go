// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux && !android) || freebsd || openbsd
// +build linux,!android freebsd openbsd

// Package xkb implements a Go interface for the X Keyboard Extension library.
package xkb

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"gioui.org/io/event"
	"gioui.org/io/key"
)

/*
#cgo linux pkg-config: xkbcommon
#cgo freebsd openbsd CFLAGS: -I/usr/local/include
#cgo freebsd openbsd LDFLAGS: -L/usr/local/lib -lxkbcommon

#include <stdlib.h>
#include <xkbcommon/xkbcommon.h>
#include <xkbcommon/xkbcommon-compose.h>
*/
import "C"

type Context struct {
	Ctx       *C.struct_xkb_context
	keyMap    *C.struct_xkb_keymap
	state     *C.struct_xkb_state
	compTable *C.struct_xkb_compose_table
	compState *C.struct_xkb_compose_state
	utf8Buf   []byte
}

var (
	_XKB_MOD_NAME_CTRL  = []byte("Control\x00")
	_XKB_MOD_NAME_SHIFT = []byte("Shift\x00")
	_XKB_MOD_NAME_ALT   = []byte("Mod1\x00")
	_XKB_MOD_NAME_LOGO  = []byte("Mod4\x00")
)

func (x *Context) Destroy() {
	if x.compState != nil {
		C.xkb_compose_state_unref(x.compState)
		x.compState = nil
	}
	if x.compTable != nil {
		C.xkb_compose_table_unref(x.compTable)
		x.compTable = nil
	}
	x.DestroyKeymapState()
	if x.Ctx != nil {
		C.xkb_context_unref(x.Ctx)
		x.Ctx = nil
	}
}

func New() (*Context, error) {
	ctx := &Context{
		Ctx: C.xkb_context_new(C.XKB_CONTEXT_NO_FLAGS),
	}
	if ctx.Ctx == nil {
		return nil, errors.New("newXKB: xkb_context_new failed")
	}
	locale := os.Getenv("LC_ALL")
	if locale == "" {
		locale = os.Getenv("LC_CTYPE")
	}
	if locale == "" {
		locale = os.Getenv("LANG")
	}
	if locale == "" {
		locale = "C"
	}
	cloc := C.CString(locale)
	defer C.free(unsafe.Pointer(cloc))
	ctx.compTable = C.xkb_compose_table_new_from_locale(ctx.Ctx, cloc, C.XKB_COMPOSE_COMPILE_NO_FLAGS)
	if ctx.compTable == nil {
		ctx.Destroy()
		return nil, errors.New("newXKB: xkb_compose_table_new_from_locale failed")
	}
	ctx.compState = C.xkb_compose_state_new(ctx.compTable, C.XKB_COMPOSE_STATE_NO_FLAGS)
	if ctx.compState == nil {
		ctx.Destroy()
		return nil, errors.New("newXKB: xkb_compose_state_new failed")
	}
	return ctx, nil
}

func (x *Context) DestroyKeymapState() {
	if x.state != nil {
		C.xkb_state_unref(x.state)
		x.state = nil
	}
	if x.keyMap != nil {
		C.xkb_keymap_unref(x.keyMap)
		x.keyMap = nil
	}
}

// SetKeymap sets the keymap and state. The context takes ownership of the
// keymap and state and frees them in Destroy.
func (x *Context) SetKeymap(xkbKeyMap, xkbState unsafe.Pointer) {
	x.DestroyKeymapState()
	x.keyMap = (*C.struct_xkb_keymap)(xkbKeyMap)
	x.state = (*C.struct_xkb_state)(xkbState)
}

func (x *Context) LoadKeymap(format int, fd int, size int) error {
	x.DestroyKeymapState()
	mapData, err := syscall.Mmap(int(fd), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("newXKB: mmap of keymap failed: %v", err)
	}
	defer syscall.Munmap(mapData)
	keyMap := C.xkb_keymap_new_from_buffer(x.Ctx, (*C.char)(unsafe.Pointer(&mapData[0])), C.size_t(size-1), C.XKB_KEYMAP_FORMAT_TEXT_V1, C.XKB_KEYMAP_COMPILE_NO_FLAGS)
	if keyMap == nil {
		return errors.New("newXKB: xkb_keymap_new_from_buffer failed")
	}
	state := C.xkb_state_new(keyMap)
	if state == nil {
		C.xkb_keymap_unref(keyMap)
		return errors.New("newXKB: xkb_state_new failed")
	}
	x.keyMap = keyMap
	x.state = state
	return nil
}

func (x *Context) Modifiers() key.Modifiers {
	var mods key.Modifiers
	if C.xkb_state_mod_name_is_active(x.state, (*C.char)(unsafe.Pointer(&_XKB_MOD_NAME_CTRL[0])), C.XKB_STATE_MODS_EFFECTIVE) == 1 {
		mods |= key.ModCtrl
	}
	if C.xkb_state_mod_name_is_active(x.state, (*C.char)(unsafe.Pointer(&_XKB_MOD_NAME_SHIFT[0])), C.XKB_STATE_MODS_EFFECTIVE) == 1 {
		mods |= key.ModShift
	}
	if C.xkb_state_mod_name_is_active(x.state, (*C.char)(unsafe.Pointer(&_XKB_MOD_NAME_ALT[0])), C.XKB_STATE_MODS_EFFECTIVE) == 1 {
		mods |= key.ModAlt
	}
	if C.xkb_state_mod_name_is_active(x.state, (*C.char)(unsafe.Pointer(&_XKB_MOD_NAME_LOGO[0])), C.XKB_STATE_MODS_EFFECTIVE) == 1 {
		mods |= key.ModSuper
	}
	return mods
}

func (x *Context) DispatchKey(keyCode uint32, state key.State) (events []event.Event) {
	if x.state == nil {
		return
	}
	kc := C.xkb_keycode_t(keyCode)
	if len(x.utf8Buf) == 0 {
		x.utf8Buf = make([]byte, 1)
	}
	sym := C.xkb_state_key_get_one_sym(x.state, kc)
	if name, ok := convertKeysym(sym); ok {
		cmd := key.Event{
			Name:      name,
			Modifiers: x.Modifiers(),
			State:     state,
		}
		// Ensure that a physical backtab key is translated to
		// Shift-Tab.
		if sym == C.XKB_KEY_ISO_Left_Tab {
			cmd.Modifiers |= key.ModShift
		}
		events = append(events, cmd)
	}
	C.xkb_compose_state_feed(x.compState, sym)
	var str []byte
	switch C.xkb_compose_state_get_status(x.compState) {
	case C.XKB_COMPOSE_CANCELLED, C.XKB_COMPOSE_COMPOSING:
		return
	case C.XKB_COMPOSE_COMPOSED:
		size := C.xkb_compose_state_get_utf8(x.compState, (*C.char)(unsafe.Pointer(&x.utf8Buf[0])), C.size_t(len(x.utf8Buf)))
		if int(size) >= len(x.utf8Buf) {
			x.utf8Buf = make([]byte, size+1)
			size = C.xkb_compose_state_get_utf8(x.compState, (*C.char)(unsafe.Pointer(&x.utf8Buf[0])), C.size_t(len(x.utf8Buf)))
		}
		C.xkb_compose_state_reset(x.compState)
		str = x.utf8Buf[:size]
	case C.XKB_COMPOSE_NOTHING:
		mod := x.Modifiers()
		if mod&(key.ModCtrl|key.ModAlt|key.ModSuper) == 0 {
			str = x.charsForKeycode(kc)
		}
	}
	// Report only printable runes.
	var n int
	for n < len(str) {
		r, s := utf8.DecodeRune(str)
		if unicode.IsPrint(r) {
			n += s
		} else {
			copy(str[n:], str[n+s:])
			str = str[:len(str)-s]
		}
	}
	if state == key.Press && len(str) > 0 {
		events = append(events, key.EditEvent{Text: string(str)})
	}
	return
}

func (x *Context) charsForKeycode(keyCode C.xkb_keycode_t) []byte {
	size := C.xkb_state_key_get_utf8(x.state, keyCode, (*C.char)(unsafe.Pointer(&x.utf8Buf[0])), C.size_t(len(x.utf8Buf)))
	if int(size) >= len(x.utf8Buf) {
		x.utf8Buf = make([]byte, size+1)
		size = C.xkb_state_key_get_utf8(x.state, keyCode, (*C.char)(unsafe.Pointer(&x.utf8Buf[0])), C.size_t(len(x.utf8Buf)))
	}
	return x.utf8Buf[:size]
}

func (x *Context) IsRepeatKey(keyCode uint32) bool {
	kc := C.xkb_keycode_t(keyCode)
	return C.xkb_keymap_key_repeats(x.keyMap, kc) == 1
}

func (x *Context) UpdateMask(depressed, latched, locked, depressedGroup, latchedGroup, lockedGroup uint32) {
	if x.state == nil {
		return
	}
	C.xkb_state_update_mask(x.state, C.xkb_mod_mask_t(depressed), C.xkb_mod_mask_t(latched), C.xkb_mod_mask_t(locked),
		C.xkb_layout_index_t(depressedGroup), C.xkb_layout_index_t(latchedGroup), C.xkb_layout_index_t(lockedGroup))
}

func convertKeysym(s C.xkb_keysym_t) (string, bool) {
	if 'a' <= s && s <= 'z' {
		return string(rune(s - 'a' + 'A')), true
	}
	if ' ' < s && s <= '~' {
		return string(rune(s)), true
	}
	var n string
	switch s {
	case C.XKB_KEY_Escape:
		n = key.NameEscape
	case C.XKB_KEY_Left:
		n = key.NameLeftArrow
	case C.XKB_KEY_Right:
		n = key.NameRightArrow
	case C.XKB_KEY_Return:
		n = key.NameReturn
	case C.XKB_KEY_KP_Enter:
		n = key.NameEnter
	case C.XKB_KEY_Up:
		n = key.NameUpArrow
	case C.XKB_KEY_Down:
		n = key.NameDownArrow
	case C.XKB_KEY_Home:
		n = key.NameHome
	case C.XKB_KEY_End:
		n = key.NameEnd
	case C.XKB_KEY_BackSpace:
		n = key.NameDeleteBackward
	case C.XKB_KEY_Delete:
		n = key.NameDeleteForward
	case C.XKB_KEY_Page_Up:
		n = key.NamePageUp
	case C.XKB_KEY_Page_Down:
		n = key.NamePageDown
	case C.XKB_KEY_F1:
		n = "F1"
	case C.XKB_KEY_F2:
		n = "F2"
	case C.XKB_KEY_F3:
		n = "F3"
	case C.XKB_KEY_F4:
		n = "F4"
	case C.XKB_KEY_F5:
		n = "F5"
	case C.XKB_KEY_F6:
		n = "F6"
	case C.XKB_KEY_F7:
		n = "F7"
	case C.XKB_KEY_F8:
		n = "F8"
	case C.XKB_KEY_F9:
		n = "F9"
	case C.XKB_KEY_F10:
		n = "F10"
	case C.XKB_KEY_F11:
		n = "F11"
	case C.XKB_KEY_F12:
		n = "F12"
	case C.XKB_KEY_Tab, C.XKB_KEY_KP_Tab, C.XKB_KEY_ISO_Left_Tab:
		n = key.NameTab
	case 0x20, C.XKB_KEY_KP_Space:
		n = key.NameSpace
	default:
		return "", false
	}
	return n, true
}
