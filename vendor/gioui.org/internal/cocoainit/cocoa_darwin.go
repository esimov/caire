// SPDX-License-Identifier: Unlicense OR MIT

// Package cocoainit initializes support for multithreaded
// programs in Cocoa.
package cocoainit

/*
#cgo CFLAGS: -xobjective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>

static inline void activate_cocoa_multithreading() {
    [[NSThread new] start];
}
#pragma GCC visibility push(hidden)
*/
import "C"

func init() {
	C.activate_cocoa_multithreading()
}
