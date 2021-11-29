// SPDX-License-Identifier: Unlicense OR MIT

// +build darwin,!ios,nometal

@import AppKit;

#include <CoreFoundation/CoreFoundation.h>
#include <OpenGL/OpenGL.h>
#include "_cgo_export.h"

CALayer *gio_layerFactory(void) {
	@autoreleasepool {
		return [CALayer layer];
	}
}

CFTypeRef gio_createGLContext(void) {
	@autoreleasepool {
		NSOpenGLPixelFormatAttribute attr[] = {
			NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core,
			NSOpenGLPFAColorSize,     24,
			NSOpenGLPFAAccelerated,
			// Opt-in to automatic GPU switching. CGL-only property.
			kCGLPFASupportsAutomaticGraphicsSwitching,
			NSOpenGLPFAAllowOfflineRenderers,
			0
		};
		NSOpenGLPixelFormat *pixFormat = [[NSOpenGLPixelFormat alloc] initWithAttributes:attr];

		NSOpenGLContext *ctx = [[NSOpenGLContext alloc] initWithFormat:pixFormat shareContext: nil];
		return CFBridgingRetain(ctx);
	}
}

void gio_setContextView(CFTypeRef ctxRef, CFTypeRef viewRef) {
	NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
	NSView *view = (__bridge NSView *)viewRef;
	[view setWantsBestResolutionOpenGLSurface:YES];
	[ctx setView:view];
}

void gio_clearCurrentContext(void) {
	@autoreleasepool {
		[NSOpenGLContext clearCurrentContext];
	}
}

void gio_updateContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		[ctx update];
	}
}

void gio_makeCurrentContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		[ctx makeCurrentContext];
	}
}

void gio_lockContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		CGLLockContext([ctx CGLContextObj]);
	}
}

void gio_unlockContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		CGLUnlockContext([ctx CGLContextObj]);
	}
}
