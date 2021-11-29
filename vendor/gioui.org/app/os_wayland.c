// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland
// +build linux,!android freebsd
// +build !nowayland

#include <wayland-client.h>
#include "wayland_xdg_shell.h"
#include "wayland_text_input.h"
#include "_cgo_export.h"

const struct wl_registry_listener gio_registry_listener = {
	// Cast away const parameter.
	.global = (void (*)(void *, struct wl_registry *, uint32_t,  const char *, uint32_t))gio_onRegistryGlobal,
	.global_remove = gio_onRegistryGlobalRemove
};

const struct wl_surface_listener gio_surface_listener = {
	.enter = gio_onSurfaceEnter,
	.leave = gio_onSurfaceLeave,
};

const struct xdg_surface_listener gio_xdg_surface_listener = {
	.configure = gio_onXdgSurfaceConfigure,
};

const struct xdg_toplevel_listener gio_xdg_toplevel_listener = {
	.configure = gio_onToplevelConfigure,
	.close = gio_onToplevelClose,
};

static void xdg_wm_base_handle_ping(void *data, struct xdg_wm_base *wm, uint32_t serial) {
	xdg_wm_base_pong(wm, serial);
}

const struct xdg_wm_base_listener gio_xdg_wm_base_listener = {
	.ping = xdg_wm_base_handle_ping,
};

const struct wl_callback_listener gio_callback_listener = {
	.done = gio_onFrameDone,
};

const struct wl_output_listener gio_output_listener = {
	// Cast away const parameter.
	.geometry = (void (*)(void *, struct wl_output *, int32_t,  int32_t,  int32_t,  int32_t,  int32_t,  const char *, const char *, int32_t))gio_onOutputGeometry,
	.mode = gio_onOutputMode,
	.done = gio_onOutputDone,
	.scale = gio_onOutputScale,
};

const struct wl_seat_listener gio_seat_listener = {
	.capabilities = gio_onSeatCapabilities,
	// Cast away const parameter.
	.name = (void (*)(void *, struct wl_seat *, const char *))gio_onSeatName,
};

const struct wl_pointer_listener gio_pointer_listener = {
	.enter = gio_onPointerEnter,
	.leave = gio_onPointerLeave,
	.motion = gio_onPointerMotion,
	.button = gio_onPointerButton,
	.axis = gio_onPointerAxis,
	.frame = gio_onPointerFrame,
	.axis_source = gio_onPointerAxisSource,
	.axis_stop = gio_onPointerAxisStop,
	.axis_discrete = gio_onPointerAxisDiscrete,
};

const struct wl_touch_listener gio_touch_listener = {
	.down = gio_onTouchDown,
	.up = gio_onTouchUp,
	.motion = gio_onTouchMotion,
	.frame = gio_onTouchFrame,
	.cancel = gio_onTouchCancel,
};

const struct wl_keyboard_listener gio_keyboard_listener = {
	.keymap = gio_onKeyboardKeymap,
	.enter = gio_onKeyboardEnter,
	.leave = gio_onKeyboardLeave,
	.key = gio_onKeyboardKey,
	.modifiers = gio_onKeyboardModifiers,
	.repeat_info = gio_onKeyboardRepeatInfo
};

const struct zwp_text_input_v3_listener gio_zwp_text_input_v3_listener = {
	.enter = gio_onTextInputEnter,
	.leave = gio_onTextInputLeave,
	// Cast away const parameter.
	.preedit_string = (void (*)(void *, struct zwp_text_input_v3 *, const char *, int32_t,  int32_t))gio_onTextInputPreeditString,
	.commit_string = (void (*)(void *, struct zwp_text_input_v3 *, const char *))gio_onTextInputCommitString,
	.delete_surrounding_text = gio_onTextInputDeleteSurroundingText,
	.done = gio_onTextInputDone
};

const struct wl_data_device_listener gio_data_device_listener = {
	.data_offer = gio_onDataDeviceOffer,
	.enter = gio_onDataDeviceEnter,
	.leave = gio_onDataDeviceLeave,
	.motion = gio_onDataDeviceMotion,
	.drop = gio_onDataDeviceDrop,
	.selection = gio_onDataDeviceSelection,
};

const struct wl_data_offer_listener gio_data_offer_listener = {
	.offer = (void (*)(void *, struct wl_data_offer *, const char *))gio_onDataOfferOffer,
	.source_actions = gio_onDataOfferSourceActions,
	.action = gio_onDataOfferAction,
};

const struct wl_data_source_listener gio_data_source_listener = {
	.target = (void (*)(void *, struct wl_data_source *, const char *))gio_onDataSourceTarget,
	.send = (void (*)(void *, struct wl_data_source *, const char *, int32_t))gio_onDataSourceSend,
	.cancelled = gio_onDataSourceCancelled,
	.dnd_drop_performed = gio_onDataSourceDNDDropPerformed,
	.dnd_finished = gio_onDataSourceDNDFinished,
	.action = gio_onDataSourceAction,
};
