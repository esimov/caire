// SPDX-License-Identifier: Unlicense OR MIT

//go:build !darwin
// +build !darwin

package key

// ModShortcut is the platform's shortcut modifier, usually the Ctrl
// key. On Apple platforms it is the Cmd key.
const ModShortcut = ModCtrl
