// SPDX-License-Identifier: Unlicense OR MIT

// +build darwin,ios,nometal

@import UIKit;
@import OpenGLES;

#include "_cgo_export.h"

Class gio_layerClass(void) {
	return [CAEAGLLayer class];
}

int gio_renderbufferStorage(CFTypeRef ctxRef, CFTypeRef layerRef, GLenum buffer) {
	EAGLContext *ctx = (__bridge EAGLContext *)ctxRef;
	CAEAGLLayer *layer = (__bridge CAEAGLLayer *)layerRef;
	return (int)[ctx renderbufferStorage:buffer fromDrawable:layer];
}

int gio_presentRenderbuffer(CFTypeRef ctxRef, GLenum buffer) {
	EAGLContext *ctx = (__bridge EAGLContext *)ctxRef;
	return (int)[ctx presentRenderbuffer:buffer];
}

int gio_makeCurrent(CFTypeRef ctxRef) {
	EAGLContext *ctx = (__bridge EAGLContext *)ctxRef;
	return (int)[EAGLContext setCurrentContext:ctx];
}

CFTypeRef gio_createContext(void) {
	EAGLContext *ctx = [[EAGLContext alloc] initWithAPI:kEAGLRenderingAPIOpenGLES3];
	if (ctx == nil) {
		return nil;
	}
	return CFBridgingRetain(ctx);
}

CFTypeRef gio_createGLLayer(void) {
	CAEAGLLayer *layer = [[CAEAGLLayer layer] init];
	if (layer == nil) {
		return nil;
	}
	layer.drawableProperties = @{kEAGLDrawablePropertyColorFormat: kEAGLColorFormatSRGBA8};
	layer.opaque = YES;
	layer.anchorPoint = CGPointMake(0, 0);
	return CFBridgingRetain(layer);
}
