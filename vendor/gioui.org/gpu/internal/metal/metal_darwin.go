// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"errors"
	"fmt"
	"image"
	"unsafe"

	"gioui.org/gpu/internal/driver"
	"gioui.org/shader"
)

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc
#cgo LDFLAGS: -framework CoreGraphics

@import Metal;

#include <CoreFoundation/CoreFoundation.h>
#include <Metal/Metal.h>

typedef struct {
	void *addr;
	NSUInteger size;
} slice;

static CFTypeRef queueNewBuffer(CFTypeRef queueRef) {
	@autoreleasepool {
		id<MTLCommandQueue> queue = (__bridge id<MTLCommandQueue>)queueRef;
		return CFBridgingRetain([queue commandBuffer]);
	}
}

static void cmdBufferCommit(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		[cmdBuf commit];
	}
}

static void cmdBufferWaitUntilCompleted(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		[cmdBuf waitUntilCompleted];
	}
}

static CFTypeRef cmdBufferRenderEncoder(CFTypeRef cmdBufRef, CFTypeRef textureRef, MTLLoadAction act, float r, float g, float b, float a) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		MTLRenderPassDescriptor *desc = [MTLRenderPassDescriptor new];
		desc.colorAttachments[0].texture = (__bridge id<MTLTexture>)textureRef;
		desc.colorAttachments[0].loadAction = act;
		desc.colorAttachments[0].clearColor = MTLClearColorMake(r, g, b, a);
		return CFBridgingRetain([cmdBuf renderCommandEncoderWithDescriptor:desc]);
	}
}

static CFTypeRef cmdBufferComputeEncoder(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		return CFBridgingRetain([cmdBuf computeCommandEncoder]);
	}
}

static CFTypeRef cmdBufferBlitEncoder(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		return CFBridgingRetain([cmdBuf blitCommandEncoder]);
	}
}

static void renderEncEnd(CFTypeRef renderEncRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc endEncoding];
	}
}

static void renderEncViewport(CFTypeRef renderEncRef, MTLViewport viewport) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc setViewport:viewport];
	}
}

static void renderEncSetFragmentTexture(CFTypeRef renderEncRef, NSUInteger index, CFTypeRef texRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLTexture> tex = (__bridge id<MTLTexture>)texRef;
		[enc setFragmentTexture:tex atIndex:index];
	}
}

static void renderEncSetFragmentSamplerState(CFTypeRef renderEncRef, NSUInteger index, CFTypeRef samplerRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLSamplerState> sampler = (__bridge id<MTLSamplerState>)samplerRef;
		[enc setFragmentSamplerState:sampler atIndex:index];
	}
}

static void renderEncSetVertexBuffer(CFTypeRef renderEncRef, CFTypeRef bufRef, NSUInteger idx, NSUInteger offset) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc setVertexBuffer:buf offset:offset atIndex:idx];
	}
}

static void renderEncSetFragmentBuffer(CFTypeRef renderEncRef, CFTypeRef bufRef, NSUInteger idx, NSUInteger offset) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc setFragmentBuffer:buf offset:offset atIndex:idx];
	}
}

static void renderEncSetFragmentBytes(CFTypeRef renderEncRef, const void *bytes, NSUInteger length, NSUInteger idx) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc setFragmentBytes:bytes length:length atIndex:idx];
	}
}

static void renderEncSetVertexBytes(CFTypeRef renderEncRef, const void *bytes, NSUInteger length, NSUInteger idx) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc setVertexBytes:bytes length:length atIndex:idx];
	}
}

static void renderEncSetRenderPipelineState(CFTypeRef renderEncRef, CFTypeRef pipeRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLRenderPipelineState> pipe = (__bridge id<MTLRenderPipelineState>)pipeRef;
		[enc setRenderPipelineState:pipe];
	}
}

static void renderEncDrawPrimitives(CFTypeRef renderEncRef, MTLPrimitiveType type, NSUInteger start, NSUInteger count) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc drawPrimitives:type vertexStart:start vertexCount:count];
	}
}

static void renderEncDrawIndexedPrimitives(CFTypeRef renderEncRef, MTLPrimitiveType type, CFTypeRef bufRef, NSUInteger offset, NSUInteger count) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc drawIndexedPrimitives:type indexCount:count indexType:MTLIndexTypeUInt16 indexBuffer:buf indexBufferOffset:offset];
	}
}

static void computeEncSetPipeline(CFTypeRef computeEncRef, CFTypeRef pipeRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		id<MTLComputePipelineState> pipe = (__bridge id<MTLComputePipelineState>)pipeRef;
		[enc setComputePipelineState:pipe];
	}
}

static void computeEncSetTexture(CFTypeRef computeEncRef, NSUInteger index, CFTypeRef texRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		id<MTLTexture> tex = (__bridge id<MTLTexture>)texRef;
		[enc setTexture:tex atIndex:index];
	}
}

static void computeEncEnd(CFTypeRef computeEncRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		[enc endEncoding];
	}
}

static void computeEncSetBuffer(CFTypeRef computeEncRef, NSUInteger index, CFTypeRef bufRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc setBuffer:buf offset:0 atIndex:index];
	}
}

static void computeEncDispatch(CFTypeRef computeEncRef, MTLSize threadgroupsPerGrid, MTLSize threadsPerThreadgroup) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		[enc dispatchThreadgroups:threadgroupsPerGrid threadsPerThreadgroup:threadsPerThreadgroup];
	}
}

static void computeEncSetBytes(CFTypeRef computeEncRef, const void *bytes, NSUInteger length, NSUInteger index) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		[enc setBytes:bytes length:length atIndex:index];
	}
}

static void blitEncEnd(CFTypeRef blitEncRef) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		[enc endEncoding];
	}
}

static void blitEncCopyFromTexture(CFTypeRef blitEncRef, CFTypeRef srcRef, MTLOrigin srcOrig, MTLSize srcSize, CFTypeRef dstRef, MTLOrigin dstOrig) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLTexture> src = (__bridge id<MTLTexture>)srcRef;
		id<MTLTexture> dst = (__bridge id<MTLTexture>)dstRef;
		[enc copyFromTexture:src
				 sourceSlice:0
				 sourceLevel:0
			    sourceOrigin:srcOrig
				  sourceSize:srcSize
				   toTexture:dst
			destinationSlice:0
			destinationLevel:0
		   destinationOrigin:dstOrig];
	}
}

static void blitEncCopyBufferToTexture(CFTypeRef blitEncRef, CFTypeRef bufRef, CFTypeRef texRef, NSUInteger offset, NSUInteger stride, NSUInteger length, MTLSize dims, MTLOrigin orig) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLBuffer> src = (__bridge id<MTLBuffer>)bufRef;
		id<MTLTexture> dst = (__bridge id<MTLTexture>)texRef;
		[enc copyFromBuffer:src
			   sourceOffset:offset
		  sourceBytesPerRow:stride
		sourceBytesPerImage:length
				 sourceSize:dims
				  toTexture:dst
		   destinationSlice:0
		   destinationLevel:0
		  destinationOrigin:orig];
	}
}

static void blitEncCopyTextureToBuffer(CFTypeRef blitEncRef, CFTypeRef texRef, CFTypeRef bufRef, NSUInteger offset, NSUInteger stride, NSUInteger length, MTLSize dims, MTLOrigin orig) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLTexture> src = (__bridge id<MTLTexture>)texRef;
		id<MTLBuffer> dst = (__bridge id<MTLBuffer>)bufRef;
		[enc		 copyFromTexture:src
						 sourceSlice:0
						 sourceLevel:0
						sourceOrigin:orig
						  sourceSize:dims
							toBuffer:dst
				   destinationOffset:offset
			  destinationBytesPerRow:stride
			destinationBytesPerImage:length];
	}
}

static void blitEncCopyBufferToBuffer(CFTypeRef blitEncRef, CFTypeRef srcRef, CFTypeRef dstRef, NSUInteger srcOff, NSUInteger dstOff, NSUInteger size) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLBuffer> src = (__bridge id<MTLBuffer>)srcRef;
		id<MTLBuffer> dst = (__bridge id<MTLBuffer>)dstRef;
		[enc   copyFromBuffer:src
				 sourceOffset:srcOff
					 toBuffer:dst
			destinationOffset:dstOff
						 size:size];
	}
}

static CFTypeRef newTexture(CFTypeRef devRef, NSUInteger width, NSUInteger height, MTLPixelFormat format, MTLTextureUsage usage) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		MTLTextureDescriptor *mtlDesc = [MTLTextureDescriptor texture2DDescriptorWithPixelFormat: format
																						   width: width
																						  height: height
																				   	   mipmapped: NO];
		mtlDesc.usage = usage;
		mtlDesc.storageMode =  MTLStorageModePrivate;
		return CFBridgingRetain([dev newTextureWithDescriptor:mtlDesc]);
	}
}

static CFTypeRef newSampler(CFTypeRef devRef, MTLSamplerMinMagFilter minFilter, MTLSamplerMinMagFilter magFilter) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		MTLSamplerDescriptor *desc = [MTLSamplerDescriptor new];
		desc.minFilter = minFilter;
		desc.magFilter = magFilter;
		return CFBridgingRetain([dev newSamplerStateWithDescriptor:desc]);
	}
}

static CFTypeRef newBuffer(CFTypeRef devRef, NSUInteger size, MTLResourceOptions opts) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		id<MTLBuffer> buf = [dev newBufferWithLength:size
											 options:opts];
		return CFBridgingRetain(buf);
	}
}

static slice bufferContents(CFTypeRef bufRef) {
	@autoreleasepool {
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		slice s = {.addr = [buf contents], .size = [buf length]};
		return s;
	}
}

static CFTypeRef newLibrary(CFTypeRef devRef, char *name, void *mtllib, size_t size) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		dispatch_data_t data = dispatch_data_create(mtllib, size, DISPATCH_TARGET_QUEUE_DEFAULT, DISPATCH_DATA_DESTRUCTOR_DEFAULT);
		id<MTLLibrary> lib = [dev newLibraryWithData:data error:nil];
		lib.label = [NSString stringWithUTF8String:name];
		return CFBridgingRetain(lib);
	}
}

static CFTypeRef libraryNewFunction(CFTypeRef libRef, char *funcName) {
	@autoreleasepool {
		id<MTLLibrary> lib = (__bridge id<MTLLibrary>)libRef;
		NSString *name = [NSString stringWithUTF8String:funcName];
		return CFBridgingRetain([lib newFunctionWithName:name]);
	}
}

static CFTypeRef newComputePipeline(CFTypeRef devRef, CFTypeRef funcRef) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		id<MTLFunction> func = (__bridge id<MTLFunction>)funcRef;
		return CFBridgingRetain([dev newComputePipelineStateWithFunction:func error:nil]);
	}
}

static CFTypeRef newRenderPipeline(CFTypeRef devRef, CFTypeRef vertFunc, CFTypeRef fragFunc, MTLPixelFormat pixelFormat, NSUInteger bufIdx, NSUInteger nverts, MTLVertexFormat *fmts, NSUInteger *offsets, NSUInteger stride, int blend, MTLBlendFactor srcFactor, MTLBlendFactor dstFactor, NSUInteger nvertBufs, NSUInteger nfragBufs) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		id<MTLFunction> vfunc = (__bridge id<MTLFunction>)vertFunc;
		id<MTLFunction> ffunc = (__bridge id<MTLFunction>)fragFunc;
		MTLVertexDescriptor *vdesc = [MTLVertexDescriptor vertexDescriptor];
		vdesc.layouts[bufIdx].stride = stride;
		for (NSUInteger i = 0; i < nverts; i++) {
			vdesc.attributes[i].format = fmts[i];
			vdesc.attributes[i].offset = offsets[i];
			vdesc.attributes[i].bufferIndex = bufIdx;
		}
		MTLRenderPipelineDescriptor *desc = [MTLRenderPipelineDescriptor new];
		desc.vertexFunction = vfunc;
		desc.fragmentFunction = ffunc;
		desc.vertexDescriptor = vdesc;
		for (NSUInteger i = 0; i < nvertBufs; i++) {
			if (@available(iOS 11.0, *)) {
				desc.vertexBuffers[i].mutability = MTLMutabilityImmutable;
			}
		}
		for (NSUInteger i = 0; i < nfragBufs; i++) {
			if (@available(iOS 11.0, *)) {
				desc.fragmentBuffers[i].mutability = MTLMutabilityImmutable;
			}
		}
		desc.colorAttachments[0].pixelFormat = pixelFormat;
		desc.colorAttachments[0].blendingEnabled = blend ? YES : NO;
		desc.colorAttachments[0].sourceAlphaBlendFactor = srcFactor;
		desc.colorAttachments[0].sourceRGBBlendFactor = srcFactor;
		desc.colorAttachments[0].destinationAlphaBlendFactor = dstFactor;
		desc.colorAttachments[0].destinationRGBBlendFactor = dstFactor;
		return CFBridgingRetain([dev newRenderPipelineStateWithDescriptor:desc
																	error:nil]);
	}
}
*/
import "C"

type Backend struct {
	dev      C.CFTypeRef
	queue    C.CFTypeRef
	pixelFmt C.MTLPixelFormat

	cmdBuffer     C.CFTypeRef
	lastCmdBuffer C.CFTypeRef
	renderEnc     C.CFTypeRef
	computeEnc    C.CFTypeRef
	blitEnc       C.CFTypeRef

	prog     *Program
	topology C.MTLPrimitiveType

	stagingBuf C.CFTypeRef
	stagingOff int

	indexBuf *Buffer

	// bufSizes is scratch space for filling out the spvBufferSizeConstants
	// that spirv-cross generates for emulating buffer.length expressions in
	// shaders.
	bufSizes []uint32
}

type Texture struct {
	backend *Backend
	texture C.CFTypeRef
	sampler C.CFTypeRef
	width   int
	height  int
	foreign bool
}

type Shader struct {
	function C.CFTypeRef
	inputs   []shader.InputLocation
}

type Program struct {
	pipeline  C.CFTypeRef
	groupSize [3]int
}

type Pipeline struct {
	pipeline C.CFTypeRef
	topology C.MTLPrimitiveType
}

type Buffer struct {
	backend *Backend
	size    int
	buffer  C.CFTypeRef

	// store is the buffer contents For buffers not allocated on the GPU.
	store []byte
}

const (
	uniformBufferIndex   = 0
	attributeBufferIndex = 1

	spvBufferSizeConstantsBinding = 25
)

const (
	texUnits    = 4
	bufferUnits = 4
)

func init() {
	driver.NewMetalDevice = newMetalDevice
}

func newMetalDevice(api driver.Metal) (driver.Device, error) {
	dev := C.CFTypeRef(api.Device)
	C.CFRetain(dev)
	queue := C.CFTypeRef(api.Queue)
	C.CFRetain(queue)
	b := &Backend{
		dev:      dev,
		queue:    queue,
		pixelFmt: C.MTLPixelFormat(api.PixelFormat),
		bufSizes: make([]uint32, bufferUnits),
	}
	return b, nil
}

func (b *Backend) BeginFrame(target driver.RenderTarget, clear bool, viewport image.Point) driver.Texture {
	if b.lastCmdBuffer != 0 {
		C.cmdBufferWaitUntilCompleted(b.lastCmdBuffer)
		b.stagingOff = 0
	}
	if target == nil {
		return nil
	}
	switch t := target.(type) {
	case driver.MetalRenderTarget:
		texture := C.CFTypeRef(t.Texture)
		return &Texture{texture: texture, foreign: true}
	case *Texture:
		return t
	default:
		panic(fmt.Sprintf("metal: unsupported render target type: %T", t))
	}
}

func (b *Backend) startBlit() C.CFTypeRef {
	if b.blitEnc != 0 {
		return b.blitEnc
	}
	b.endEncoder()
	b.ensureCmdBuffer()
	b.blitEnc = C.cmdBufferBlitEncoder(b.cmdBuffer)
	if b.blitEnc == 0 {
		panic("metal: [MTLCommandBuffer blitCommandEncoder:] failed")
	}
	return b.blitEnc
}

func (b *Backend) CopyTexture(dst driver.Texture, dorig image.Point, src driver.Texture, srect image.Rectangle) {
	enc := b.startBlit()
	dstTex := dst.(*Texture).texture
	srcTex := src.(*Texture).texture
	ssz := srect.Size()
	C.blitEncCopyFromTexture(
		enc,
		srcTex,
		C.MTLOrigin{
			x: C.NSUInteger(srect.Min.X),
			y: C.NSUInteger(srect.Min.Y),
		},
		C.MTLSize{
			width:  C.NSUInteger(ssz.X),
			height: C.NSUInteger(ssz.Y),
			depth:  1,
		},
		dstTex,
		C.MTLOrigin{
			x: C.NSUInteger(dorig.X),
			y: C.NSUInteger(dorig.Y),
		},
	)
}

func (b *Backend) EndFrame() {
	b.endCmdBuffer(false)
}

func (b *Backend) endCmdBuffer(wait bool) {
	b.endEncoder()
	if b.cmdBuffer == 0 {
		return
	}
	C.cmdBufferCommit(b.cmdBuffer)
	if wait {
		C.cmdBufferWaitUntilCompleted(b.cmdBuffer)
	}
	if b.lastCmdBuffer != 0 {
		C.CFRelease(b.lastCmdBuffer)
	}
	b.lastCmdBuffer = b.cmdBuffer
	b.cmdBuffer = 0
}

func (b *Backend) Caps() driver.Caps {
	return driver.Caps{
		MaxTextureSize: 8192,
		Features:       driver.FeatureSRGB | driver.FeatureCompute | driver.FeatureFloatRenderTargets,
	}
}

func (b *Backend) NewTimer() driver.Timer {
	panic("timers not supported")
}

func (b *Backend) IsTimeContinuous() bool {
	panic("timers not supported")
}

func (b *Backend) Release() {
	if b.cmdBuffer != 0 {
		C.CFRelease(b.cmdBuffer)
	}
	if b.lastCmdBuffer != 0 {
		C.CFRelease(b.lastCmdBuffer)
	}
	if b.stagingBuf != 0 {
		C.CFRelease(b.stagingBuf)
	}
	C.CFRelease(b.queue)
	C.CFRelease(b.dev)
	*b = Backend{}
}

func (b *Backend) NewTexture(format driver.TextureFormat, width, height int, minFilter, magFilter driver.TextureFilter, bindings driver.BufferBinding) (driver.Texture, error) {
	mformat := pixelFormatFor(format)
	var usage C.MTLTextureUsage
	if bindings&(driver.BufferBindingTexture|driver.BufferBindingShaderStorageRead) != 0 {
		usage |= C.MTLTextureUsageShaderRead
	}
	if bindings&driver.BufferBindingFramebuffer != 0 {
		usage |= C.MTLTextureUsageRenderTarget
	}
	if bindings&driver.BufferBindingShaderStorageWrite != 0 {
		usage |= C.MTLTextureUsageShaderWrite
	}
	tex := C.newTexture(b.dev, C.NSUInteger(width), C.NSUInteger(height), mformat, usage)
	if tex == 0 {
		return nil, errors.New("metal: [MTLDevice newTextureWithDescriptor:] failed")
	}
	min := samplerFilterFor(minFilter)
	max := samplerFilterFor(magFilter)
	s := C.newSampler(b.dev, min, max)
	if s == 0 {
		C.CFRelease(tex)
		return nil, errors.New("metal: [MTLDevice newSamplerStateWithDescriptor:] failed")
	}
	return &Texture{backend: b, texture: tex, sampler: s, width: width, height: height}, nil
}

func samplerFilterFor(f driver.TextureFilter) C.MTLSamplerMinMagFilter {
	switch f {
	case driver.FilterNearest:
		return C.MTLSamplerMinMagFilterNearest
	case driver.FilterLinear:
		return C.MTLSamplerMinMagFilterLinear
	default:
		panic("invalid texture filter")
	}
}

func (b *Backend) NewPipeline(desc driver.PipelineDesc) (driver.Pipeline, error) {
	vsh, fsh := desc.VertexShader.(*Shader), desc.FragmentShader.(*Shader)
	layout := desc.VertexLayout.Inputs
	if got, exp := len(layout), len(vsh.inputs); got != exp {
		return nil, fmt.Errorf("metal: number of input descriptors (%d) doesn't match number of inputs (%d)", got, exp)
	}
	formats := make([]C.MTLVertexFormat, len(layout))
	offsets := make([]C.NSUInteger, len(layout))
	for i, inp := range layout {
		index := vsh.inputs[i].Location
		formats[index] = vertFormatFor(vsh.inputs[i])
		offsets[index] = C.NSUInteger(inp.Offset)
	}
	var (
		fmtPtr *C.MTLVertexFormat
		offPtr *C.NSUInteger
	)
	if len(layout) > 0 {
		fmtPtr = &formats[0]
		offPtr = &offsets[0]
	}
	srcFactor := blendFactorFor(desc.BlendDesc.SrcFactor)
	dstFactor := blendFactorFor(desc.BlendDesc.DstFactor)
	blend := C.int(0)
	if desc.BlendDesc.Enable {
		blend = 1
	}
	pf := b.pixelFmt
	if f := desc.PixelFormat; f != driver.TextureFormatOutput {
		pf = pixelFormatFor(f)
	}
	pipe := C.newRenderPipeline(
		b.dev,
		vsh.function,
		fsh.function,
		pf,
		attributeBufferIndex,
		C.NSUInteger(len(layout)), fmtPtr, offPtr,
		C.NSUInteger(desc.VertexLayout.Stride),
		blend, srcFactor, dstFactor,
		2, // Number of vertex buffers.
		1, // Number of fragment buffers.
	)
	if pipe == 0 {
		return nil, errors.New("metal: pipeline construction failed")
	}
	return &Pipeline{pipeline: pipe, topology: primitiveFor(desc.Topology)}, nil
}

func dataTypeSize(d shader.DataType) int {
	switch d {
	case shader.DataTypeFloat:
		return 4
	default:
		panic("unsupported data type")
	}
}

func blendFactorFor(f driver.BlendFactor) C.MTLBlendFactor {
	switch f {
	case driver.BlendFactorZero:
		return C.MTLBlendFactorZero
	case driver.BlendFactorOne:
		return C.MTLBlendFactorOne
	case driver.BlendFactorOneMinusSrcAlpha:
		return C.MTLBlendFactorOneMinusSourceAlpha
	case driver.BlendFactorDstColor:
		return C.MTLBlendFactorDestinationColor
	default:
		panic("unsupported blend factor")
	}
}

func vertFormatFor(f shader.InputLocation) C.MTLVertexFormat {
	t := f.Type
	s := f.Size
	switch {
	case t == shader.DataTypeFloat && s == 1:
		return C.MTLVertexFormatFloat
	case t == shader.DataTypeFloat && s == 2:
		return C.MTLVertexFormatFloat2
	case t == shader.DataTypeFloat && s == 3:
		return C.MTLVertexFormatFloat3
	case t == shader.DataTypeFloat && s == 4:
		return C.MTLVertexFormatFloat4
	default:
		panic("unsupported data type")
	}
}

func pixelFormatFor(f driver.TextureFormat) C.MTLPixelFormat {
	switch f {
	case driver.TextureFormatFloat:
		return C.MTLPixelFormatR16Float
	case driver.TextureFormatRGBA8:
		return C.MTLPixelFormatRGBA8Unorm
	case driver.TextureFormatSRGBA:
		return C.MTLPixelFormatRGBA8Unorm_sRGB
	default:
		panic("unsupported pixel format")
	}
}

func (b *Backend) NewBuffer(typ driver.BufferBinding, size int) (driver.Buffer, error) {
	// Transfer buffer contents in command encoders on every use for
	// smaller buffers. The advantage is that buffer re-use during a frame
	// won't occur a GPU wait.
	// We can't do this for buffers written to by the GPU and read by the client,
	// and Metal doesn't require a buffer for indexed draws.
	if size <= 4096 && typ&(driver.BufferBindingShaderStorageWrite|driver.BufferBindingIndices) == 0 {
		return &Buffer{size: size, store: make([]byte, size)}, nil
	}
	buf := C.newBuffer(b.dev, C.NSUInteger(size), C.MTLResourceStorageModePrivate)
	return &Buffer{backend: b, size: size, buffer: buf}, nil
}

func (b *Backend) NewImmutableBuffer(typ driver.BufferBinding, data []byte) (driver.Buffer, error) {
	buf, err := b.NewBuffer(typ, len(data))
	if err != nil {
		return nil, err
	}
	buf.Upload(data)
	return buf, nil
}

func (b *Backend) NewComputeProgram(src shader.Sources) (driver.Program, error) {
	sh, err := b.newShader(src)
	if err != nil {
		return nil, err
	}
	defer sh.Release()
	pipe := C.newComputePipeline(b.dev, sh.function)
	if pipe == 0 {
		return nil, fmt.Errorf("metal: compute program %q load failed", src.Name)
	}
	return &Program{pipeline: pipe, groupSize: src.WorkgroupSize}, nil
}

func (b *Backend) NewVertexShader(src shader.Sources) (driver.VertexShader, error) {
	return b.newShader(src)
}

func (b *Backend) NewFragmentShader(src shader.Sources) (driver.FragmentShader, error) {
	return b.newShader(src)
}

func (b *Backend) newShader(src shader.Sources) (*Shader, error) {
	vsrc := []byte(src.MetalLib)
	cname := C.CString(src.Name)
	defer C.free(unsafe.Pointer(cname))
	vlib := C.newLibrary(b.dev, cname, unsafe.Pointer(&vsrc[0]), C.size_t(len(vsrc)))
	if vlib == 0 {
		return nil, fmt.Errorf("metal: vertex shader %q load failed", src.Name)
	}
	defer C.CFRelease(vlib)
	funcName := C.CString("main0")
	defer C.free(unsafe.Pointer(funcName))
	f := C.libraryNewFunction(vlib, funcName)
	if f == 0 {
		return nil, fmt.Errorf("metal: main function not found in %q", src.Name)
	}
	return &Shader{function: f, inputs: src.Inputs}, nil
}

func (b *Backend) Viewport(x, y, width, height int) {
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	C.renderEncViewport(enc, C.MTLViewport{
		originX: C.double(x),
		originY: C.double(y),
		width:   C.double(width),
		height:  C.double(height),
		znear:   0.0,
		zfar:    1.0,
	})
}

func (b *Backend) DrawArrays(off, count int) {
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	C.renderEncDrawPrimitives(enc, b.topology, C.NSUInteger(off), C.NSUInteger(count))
}

func (b *Backend) DrawElements(off, count int) {
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	C.renderEncDrawIndexedPrimitives(enc, b.topology, b.indexBuf.buffer, C.NSUInteger(off), C.NSUInteger(count))
}

func primitiveFor(mode driver.Topology) C.MTLPrimitiveType {
	switch mode {
	case driver.TopologyTriangles:
		return C.MTLPrimitiveTypeTriangle
	case driver.TopologyTriangleStrip:
		return C.MTLPrimitiveTypeTriangleStrip
	default:
		panic("metal: unknown draw mode")
	}
}

func (b *Backend) BindImageTexture(unit int, tex driver.Texture) {
	b.BindTexture(unit, tex)
}

func (b *Backend) BeginCompute() {
	b.endEncoder()
	b.ensureCmdBuffer()
	for i := range b.bufSizes {
		b.bufSizes[i] = 0
	}
	b.computeEnc = C.cmdBufferComputeEncoder(b.cmdBuffer)
	if b.computeEnc == 0 {
		panic("metal: [MTLCommandBuffer computeCommandEncoder:] failed")
	}
}

func (b *Backend) EndCompute() {
	if b.computeEnc == 0 {
		panic("no active compute pass")
	}
	C.computeEncEnd(b.computeEnc)
	C.CFRelease(b.computeEnc)
	b.computeEnc = 0
}

func (b *Backend) DispatchCompute(x, y, z int) {
	enc := b.computeEnc
	if enc == 0 {
		panic("no active compute pass")
	}
	C.computeEncSetBytes(enc, unsafe.Pointer(&b.bufSizes[0]), C.NSUInteger(len(b.bufSizes)*4), spvBufferSizeConstantsBinding)
	threadgroupsPerGrid := C.MTLSize{
		width: C.NSUInteger(x), height: C.NSUInteger(y), depth: C.NSUInteger(z),
	}
	sz := b.prog.groupSize
	threadsPerThreadgroup := C.MTLSize{
		width: C.NSUInteger(sz[0]), height: C.NSUInteger(sz[1]), depth: C.NSUInteger(sz[2]),
	}
	C.computeEncDispatch(enc, threadgroupsPerGrid, threadsPerThreadgroup)
}

func (b *Backend) stagingBuffer(size int) (C.CFTypeRef, int) {
	if b.stagingBuf == 0 || b.stagingOff+size > len(bufferStore(b.stagingBuf)) {
		if b.stagingBuf != 0 {
			C.CFRelease(b.stagingBuf)
		}
		cap := 2 * (b.stagingOff + size)
		b.stagingBuf = C.newBuffer(b.dev, C.NSUInteger(cap), C.MTLResourceStorageModeShared|C.MTLResourceCPUCacheModeWriteCombined)
		if b.stagingBuf == 0 {
			panic(fmt.Errorf("metal: failed to allocate %d bytes of staging buffer", cap))
		}
		b.stagingOff = 0
	}
	off := b.stagingOff
	b.stagingOff += size
	return b.stagingBuf, off
}

func (t *Texture) Upload(offset, size image.Point, pixels []byte, stride int) {
	if len(pixels) == 0 {
		return
	}
	if stride == 0 {
		stride = size.X * 4
	}
	dstStride := size.X * 4
	n := size.Y * dstStride
	buf, off := t.backend.stagingBuffer(n)
	store := bufferSlice(buf, off, n)
	var srcOff, dstOff int
	for y := 0; y < size.Y; y++ {
		srcRow := pixels[srcOff : srcOff+dstStride]
		dstRow := store[dstOff : dstOff+dstStride]
		copy(dstRow, srcRow)
		dstOff += dstStride
		srcOff += stride
	}
	enc := t.backend.startBlit()
	orig := C.MTLOrigin{
		x: C.NSUInteger(offset.X),
		y: C.NSUInteger(offset.Y),
	}
	msize := C.MTLSize{
		width:  C.NSUInteger(size.X),
		height: C.NSUInteger(size.Y),
		depth:  1,
	}
	C.blitEncCopyBufferToTexture(enc, buf, t.texture, C.NSUInteger(off), C.NSUInteger(dstStride), C.NSUInteger(len(store)), msize, orig)
}

func (t *Texture) Release() {
	if t.foreign {
		panic("metal: release of external texture")
	}
	C.CFRelease(t.texture)
	C.CFRelease(t.sampler)
	*t = Texture{}
}

func (p *Pipeline) Release() {
	C.CFRelease(p.pipeline)
	*p = Pipeline{}
}

func (b *Backend) PrepareTexture(tex driver.Texture) {}

func (b *Backend) BindTexture(unit int, tex driver.Texture) {
	t := tex.(*Texture)
	if enc := b.renderEnc; enc != 0 {
		C.renderEncSetFragmentTexture(enc, C.NSUInteger(unit), t.texture)
		C.renderEncSetFragmentSamplerState(enc, C.NSUInteger(unit), t.sampler)
	} else if enc := b.computeEnc; enc != 0 {
		C.computeEncSetTexture(enc, C.NSUInteger(unit), t.texture)
	} else {
		panic("no active render nor compute pass")
	}
}

func (b *Backend) ensureCmdBuffer() {
	if b.cmdBuffer != 0 {
		return
	}
	b.cmdBuffer = C.queueNewBuffer(b.queue)
	if b.cmdBuffer == 0 {
		panic("metal: [MTLCommandQueue cmdBuffer] failed")
	}
}

func (b *Backend) BindPipeline(pipe driver.Pipeline) {
	p := pipe.(*Pipeline)
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	C.renderEncSetRenderPipelineState(enc, p.pipeline)
	b.topology = p.topology
}

func (b *Backend) BindProgram(prog driver.Program) {
	enc := b.computeEnc
	if enc == 0 {
		panic("no active compute pass")
	}
	p := prog.(*Program)
	C.computeEncSetPipeline(enc, p.pipeline)
	b.prog = p
}

func (s *Shader) Release() {
	C.CFRelease(s.function)
	*s = Shader{}
}

func (p *Program) Release() {
	C.CFRelease(p.pipeline)
	*p = Program{}
}

func (b *Backend) BindStorageBuffer(binding int, buffer driver.Buffer) {
	buf := buffer.(*Buffer)
	b.bufSizes[binding] = uint32(buf.size)
	enc := b.computeEnc
	if enc == 0 {
		panic("no active compute pass")
	}
	if buf.buffer != 0 {
		C.computeEncSetBuffer(enc, C.NSUInteger(binding), buf.buffer)
	} else if buf.size > 0 {
		C.computeEncSetBytes(enc, unsafe.Pointer(&buf.store[0]), C.NSUInteger(buf.size), C.NSUInteger(binding))
	}
}

func (b *Backend) BindUniforms(buf driver.Buffer) {
	bf := buf.(*Buffer)
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	if bf.buffer != 0 {
		C.renderEncSetVertexBuffer(enc, bf.buffer, uniformBufferIndex, 0)
		C.renderEncSetFragmentBuffer(enc, bf.buffer, uniformBufferIndex, 0)
	} else if bf.size > 0 {
		C.renderEncSetVertexBytes(enc, unsafe.Pointer(&bf.store[0]), C.NSUInteger(bf.size), uniformBufferIndex)
		C.renderEncSetFragmentBytes(enc, unsafe.Pointer(&bf.store[0]), C.NSUInteger(bf.size), uniformBufferIndex)
	}
}

func (b *Backend) BindVertexBuffer(buf driver.Buffer, offset int) {
	bf := buf.(*Buffer)
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	if bf.buffer != 0 {
		C.renderEncSetVertexBuffer(enc, bf.buffer, attributeBufferIndex, C.NSUInteger(offset))
	} else if n := bf.size - offset; n > 0 {
		C.renderEncSetVertexBytes(enc, unsafe.Pointer(&bf.store[offset]), C.NSUInteger(n), attributeBufferIndex)
	}
}

func (b *Backend) BindIndexBuffer(buf driver.Buffer) {
	b.indexBuf = buf.(*Buffer)
}

func (b *Buffer) Download(data []byte) error {
	if len(data) > b.size {
		panic(fmt.Errorf("len(data) (%d) larger than len(content) (%d)", len(data), b.size))
	}
	buf, off := b.backend.stagingBuffer(len(data))
	enc := b.backend.startBlit()
	C.blitEncCopyBufferToBuffer(enc, b.buffer, buf, 0, C.NSUInteger(off), C.NSUInteger(len(data)))
	b.backend.endCmdBuffer(true)
	store := bufferSlice(buf, off, len(data))
	copy(data, store)
	return nil
}

func (b *Buffer) Upload(data []byte) {
	if len(data) > b.size {
		panic(fmt.Errorf("len(data) (%d) larger than len(content) (%d)", len(data), b.size))
	}
	if b.buffer == 0 {
		copy(b.store, data)
		return
	}
	buf, off := b.backend.stagingBuffer(len(data))
	store := bufferSlice(buf, off, len(data))
	copy(store, data)
	enc := b.backend.startBlit()
	C.blitEncCopyBufferToBuffer(enc, buf, b.buffer, C.NSUInteger(off), 0, C.NSUInteger(len(store)))
}

func bufferStore(buf C.CFTypeRef) []byte {
	contents := C.bufferContents(buf)
	return (*(*[1 << 30]byte)(contents.addr))[:contents.size:contents.size]
}

func bufferSlice(buf C.CFTypeRef, off, len int) []byte {
	store := bufferStore(buf)
	return store[off : off+len]
}

func (b *Buffer) Release() {
	if b.buffer != 0 {
		C.CFRelease(b.buffer)
	}
	*b = Buffer{}
}

func (t *Texture) ReadPixels(src image.Rectangle, pixels []byte, stride int) error {
	if len(pixels) == 0 {
		return nil
	}
	sz := src.Size()
	orig := C.MTLOrigin{
		x: C.NSUInteger(src.Min.X),
		y: C.NSUInteger(src.Min.Y),
	}
	msize := C.MTLSize{
		width:  C.NSUInteger(sz.X),
		height: C.NSUInteger(sz.Y),
		depth:  1,
	}
	stageStride := sz.X * 4
	n := sz.Y * stageStride
	buf, off := t.backend.stagingBuffer(n)
	enc := t.backend.startBlit()
	C.blitEncCopyTextureToBuffer(enc, t.texture, buf, C.NSUInteger(off), C.NSUInteger(stageStride), C.NSUInteger(n), msize, orig)
	t.backend.endCmdBuffer(true)
	store := bufferSlice(buf, off, n)
	var srcOff, dstOff int
	for y := 0; y < sz.Y; y++ {
		dstRow := pixels[srcOff : srcOff+stageStride]
		srcRow := store[dstOff : dstOff+stageStride]
		copy(dstRow, srcRow)
		dstOff += stageStride
		srcOff += stride
	}
	return nil
}

func (b *Backend) BeginRenderPass(tex driver.Texture, d driver.LoadDesc) {
	b.endEncoder()
	b.ensureCmdBuffer()
	f := tex.(*Texture)
	col := d.ClearColor
	var act C.MTLLoadAction
	switch d.Action {
	case driver.LoadActionKeep:
		act = C.MTLLoadActionLoad
	case driver.LoadActionClear:
		act = C.MTLLoadActionClear
	case driver.LoadActionInvalidate:
		act = C.MTLLoadActionDontCare
	}
	b.renderEnc = C.cmdBufferRenderEncoder(b.cmdBuffer, f.texture, act, C.float(col.R), C.float(col.G), C.float(col.B), C.float(col.A))
	if b.renderEnc == 0 {
		panic("metal: [MTLCommandBuffer renderCommandEncoderWithDescriptor:] failed")
	}
}

func (b *Backend) EndRenderPass() {
	if b.renderEnc == 0 {
		panic("no active render pass")
	}
	C.renderEncEnd(b.renderEnc)
	C.CFRelease(b.renderEnc)
	b.renderEnc = 0
}

func (b *Backend) endEncoder() {
	if b.renderEnc != 0 {
		panic("active render pass")
	}
	if b.computeEnc != 0 {
		panic("active compute pass")
	}
	if b.blitEnc != 0 {
		C.blitEncEnd(b.blitEnc)
		C.CFRelease(b.blitEnc)
		b.blitEnc = 0
	}
}

func (f *Texture) ImplementsRenderTarget() {}
