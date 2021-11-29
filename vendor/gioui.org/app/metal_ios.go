// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal
// +build !nometal

package app

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc

@import UIKit;

@import QuartzCore.CAMetalLayer;

#include <CoreFoundation/CoreFoundation.h>

Class gio_layerClass(void) {
	return [CAMetalLayer class];
}

static CFTypeRef getMetalLayer(CFTypeRef viewRef) {
	@autoreleasepool {
		UIView *view = (__bridge UIView *)viewRef;
		return CFBridgingRetain(view.layer);
	}
}

static void resizeDrawable(CFTypeRef viewRef, CFTypeRef layerRef) {
	@autoreleasepool {
		UIView *view = (__bridge UIView *)viewRef;
		CAMetalLayer *layer = (__bridge CAMetalLayer *)layerRef;
		layer.contentsScale = view.contentScaleFactor;
		CGSize size = layer.bounds.size;
		size.width *= layer.contentsScale;
		size.height *= layer.contentsScale;
		layer.drawableSize = size;
	}
}
*/
import "C"

func getMetalLayer(view C.CFTypeRef) C.CFTypeRef {
	return C.getMetalLayer(view)
}

func resizeDrawable(view, layer C.CFTypeRef) {
	C.resizeDrawable(view, layer)
}
