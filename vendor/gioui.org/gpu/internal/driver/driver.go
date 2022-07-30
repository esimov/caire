// SPDX-License-Identifier: Unlicense OR MIT

package driver

import (
	"errors"
	"image"
	"time"

	"gioui.org/internal/f32color"
	"gioui.org/shader"
)

// Device represents the abstraction of underlying GPU
// APIs such as OpenGL, Direct3D useful for rendering Gio
// operations.
type Device interface {
	BeginFrame(target RenderTarget, clear bool, viewport image.Point) Texture
	EndFrame()
	Caps() Caps
	NewTimer() Timer
	// IsContinuousTime reports whether all timer measurements
	// are valid at the point of call.
	IsTimeContinuous() bool
	NewTexture(format TextureFormat, width, height int, minFilter, magFilter TextureFilter, bindings BufferBinding) (Texture, error)
	NewImmutableBuffer(typ BufferBinding, data []byte) (Buffer, error)
	NewBuffer(typ BufferBinding, size int) (Buffer, error)
	NewComputeProgram(shader shader.Sources) (Program, error)
	NewVertexShader(src shader.Sources) (VertexShader, error)
	NewFragmentShader(src shader.Sources) (FragmentShader, error)
	NewPipeline(desc PipelineDesc) (Pipeline, error)

	Viewport(x, y, width, height int)
	DrawArrays(off, count int)
	DrawElements(off, count int)

	BeginRenderPass(t Texture, desc LoadDesc)
	EndRenderPass()
	PrepareTexture(t Texture)
	BindProgram(p Program)
	BindPipeline(p Pipeline)
	BindTexture(unit int, t Texture)
	BindVertexBuffer(b Buffer, offset int)
	BindIndexBuffer(b Buffer)
	BindImageTexture(unit int, texture Texture)
	BindUniforms(buf Buffer)
	BindStorageBuffer(binding int, buf Buffer)

	BeginCompute()
	EndCompute()
	CopyTexture(dst Texture, dstOrigin image.Point, src Texture, srcRect image.Rectangle)
	DispatchCompute(x, y, z int)

	Release()
}

var ErrDeviceLost = errors.New("GPU device lost")

type LoadDesc struct {
	Action     LoadAction
	ClearColor f32color.RGBA
}

type Pipeline interface {
	Release()
}

type PipelineDesc struct {
	VertexShader   VertexShader
	FragmentShader FragmentShader
	VertexLayout   VertexLayout
	BlendDesc      BlendDesc
	PixelFormat    TextureFormat
	Topology       Topology
}

type VertexLayout struct {
	Inputs []InputDesc
	Stride int
}

// InputDesc describes a vertex attribute as laid out in a Buffer.
type InputDesc struct {
	Type shader.DataType
	Size int

	Offset int
}

type BlendDesc struct {
	Enable               bool
	SrcFactor, DstFactor BlendFactor
}

type BlendFactor uint8

type Topology uint8

type TextureFilter uint8
type TextureFormat uint8

type BufferBinding uint8

type LoadAction uint8

type Features uint

type Caps struct {
	// BottomLeftOrigin is true if the driver has the origin in the lower left
	// corner. The OpenGL driver returns true.
	BottomLeftOrigin bool
	Features         Features
	MaxTextureSize   int
}

type VertexShader interface {
	Release()
}

type FragmentShader interface {
	Release()
}

type Program interface {
	Release()
}

type Buffer interface {
	Release()
	Upload(data []byte)
	Download(data []byte) error
}

type Timer interface {
	Begin()
	End()
	Duration() (time.Duration, bool)
	Release()
}

type Texture interface {
	RenderTarget
	Upload(offset, size image.Point, pixels []byte, stride int)
	ReadPixels(src image.Rectangle, pixels []byte, stride int) error
	Release()
}

const (
	BufferBindingIndices BufferBinding = 1 << iota
	BufferBindingVertices
	BufferBindingUniforms
	BufferBindingTexture
	BufferBindingFramebuffer
	BufferBindingShaderStorageRead
	BufferBindingShaderStorageWrite
)

const (
	TextureFormatSRGBA TextureFormat = iota
	TextureFormatFloat
	TextureFormatRGBA8
	// TextureFormatOutput denotes the format used by the output framebuffer.
	TextureFormatOutput
)

const (
	FilterNearest TextureFilter = iota
	FilterLinear
)

const (
	FeatureTimers Features = 1 << iota
	FeatureFloatRenderTargets
	FeatureCompute
	FeatureSRGB
)

const (
	TopologyTriangleStrip Topology = iota
	TopologyTriangles
)

const (
	BlendFactorOne BlendFactor = iota
	BlendFactorOneMinusSrcAlpha
	BlendFactorZero
	BlendFactorDstColor
)

const (
	LoadActionKeep LoadAction = iota
	LoadActionClear
	LoadActionInvalidate
)

var ErrContentLost = errors.New("buffer content lost")

func (f Features) Has(feats Features) bool {
	return f&feats == feats
}

func DownloadImage(d Device, t Texture, img *image.RGBA) error {
	r := img.Bounds()
	if err := t.ReadPixels(r, img.Pix, img.Stride); err != nil {
		return err
	}
	if d.Caps().BottomLeftOrigin {
		// OpenGL origin is in the lower-left corner. Flip the image to
		// match.
		flipImageY(r.Dx()*4, r.Dy(), img.Pix)
	}
	return nil
}

func flipImageY(stride, height int, pixels []byte) {
	// Flip image in y-direction. OpenGL's origin is in the lower
	// left corner.
	row := make([]uint8, stride)
	for y := 0; y < height/2; y++ {
		y1 := height - y - 1
		dest := y1 * stride
		src := y * stride
		copy(row, pixels[dest:])
		copy(pixels[dest:], pixels[src:src+len(row)])
		copy(pixels[src:], row)
	}
}

func UploadImage(t Texture, offset image.Point, img *image.RGBA) {
	var pixels []byte
	size := img.Bounds().Size()
	min := img.Rect.Min
	start := img.PixOffset(min.X, min.Y)
	end := img.PixOffset(min.X+size.X, min.Y+size.Y-1)
	pixels = img.Pix[start:end]
	t.Upload(offset, size, pixels, img.Stride)
}
