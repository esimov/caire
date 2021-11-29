// SPDX-License-Identifier: Unlicense OR MIT

//go:build !android
// +build !android

package app

import "os"

func dataDir() (string, error) {
	return os.UserConfigDir()
}
