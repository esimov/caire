// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"os"
	"strings"
)

// extraArgs contains extra arguments to append to
// os.Args. The arguments are separated with |.
// Useful for running programs on mobiles where the
// command line is not available.
// Set with the go linker flag -X.
var extraArgs string

func init() {
	if extraArgs != "" {
		args := strings.Split(extraArgs, "|")
		os.Args = append(os.Args, args...)
	}
}

// DataDir returns a path to use for application-specific
// configuration data.
// On desktop systems, DataDir use os.UserConfigDir.
// On iOS NSDocumentDirectory is queried.
// For Android Context.getFilesDir is used.
//
// BUG: DataDir blocks on Android until init functions
// have completed.
func DataDir() (string, error) {
	return dataDir()
}

// Main must be called last from the program main function.
// On most platforms Main blocks forever, for Android and
// iOS it returns immediately to give control of the main
// thread back to the system.
//
// Calling Main is necessary because some operating systems
// require control of the main thread of the program for
// running windows.
func Main() {
	osMain()
}
