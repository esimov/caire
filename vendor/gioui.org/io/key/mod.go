// SPDX-License-Identifier: Unlicense OR MIT

//go:build !darwin
// +build !darwin

package key

// ModShortcut is the platform's shortcut modifier, usually the ctrl
// modifier. On Apple platforms it is the cmd key.
const ModShortcut = ModCtrl

// ModShortcutAlt is the platform's alternative shortcut modifier,
// usually the ctrl modifier. On Apple platforms it is the alt modifier.
const ModShortcutAlt = ModCtrl
