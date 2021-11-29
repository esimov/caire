// SPDX-License-Identifier: Unlicense OR MIT

package d3d11

import (
	"errors"
	"fmt"
	"image"
	"math"
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"

	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/d3d11"
	"gioui.org/shader"
)

type Backend struct {
	dev *d3d11.Device
	ctx *d3d11.DeviceContext

	// Temporary storage to avoid garbage.
	clearColor [4]float32
	viewport   d3d11.VIEWPORT

	pipeline *Pipeline
	vert     struct {
		buffer *Buffer
		offset int
	}

	program *Program

	caps driver.Caps

	floatFormat uint32
}

type Pipeline struct {
	vert     *d3d11.VertexShader
	frag     *d3d11.PixelShader
	layout   *d3d11.InputLayout
	blend    *d3d11.BlendState
	stride   int
	topology driver.Topology
}

type Texture struct {
	backend      *Backend
	format       uint32
	bindings     driver.BufferBinding
	tex          *d3d11.Texture2D
	sampler      *d3d11.SamplerState
	resView      *d3d11.ShaderResourceView
	uaView       *d3d11.UnorderedAccessView
	renderTarget *d3d11.RenderTargetView

	width   int
	height  int
	foreign bool
}

type VertexShader struct {
	backend *Backend
	shader  *d3d11.VertexShader
	src     shader.Sources
}

type FragmentShader struct {
	backend *Backend
	shader  *d3d11.PixelShader
}

type Program struct {
	backend *Backend
	shader  *d3d11.ComputeShader
}

type Buffer struct {
	backend   *Backend
	bind      uint32
	buf       *d3d11.Buffer
	resView   *d3d11.ShaderResourceView
	uaView    *d3d11.UnorderedAccessView
	size      int
	immutable bool
}

func init() {
	driver.NewDirect3D11Device = newDirect3D11Device
}

func detectFloatFormat(dev *d3d11.Device) (uint32, bool) {
	formats := []uint32{
		d3d11.DXGI_FORMAT_R16_FLOAT,
		d3d11.DXGI_FORMAT_R32_FLOAT,
		d3d11.DXGI_FORMAT_R16G16_FLOAT,
		d3d11.DXGI_FORMAT_R32G32_FLOAT,
		// These last two are really wasteful, but c'est la vie.
		d3d11.DXGI_FORMAT_R16G16B16A16_FLOAT,
		d3d11.DXGI_FORMAT_R32G32B32A32_FLOAT,
	}
	for _, format := range formats {
		need := uint32(d3d11.FORMAT_SUPPORT_TEXTURE2D | d3d11.FORMAT_SUPPORT_RENDER_TARGET)
		if support, _ := dev.CheckFormatSupport(format); support&need == need {
			return format, true
		}
	}
	return 0, false
}

func newDirect3D11Device(api driver.Direct3D11) (driver.Device, error) {
	dev := (*d3d11.Device)(api.Device)
	b := &Backend{
		dev: dev,
		ctx: dev.GetImmediateContext(),
		caps: driver.Caps{
			MaxTextureSize: 2048, // 9.1 maximum
			Features:       driver.FeatureSRGB,
		},
	}
	featLvl := dev.GetFeatureLevel()
	switch {
	case featLvl < d3d11.FEATURE_LEVEL_9_1:
		d3d11.IUnknownRelease(unsafe.Pointer(dev), dev.Vtbl.Release)
		d3d11.IUnknownRelease(unsafe.Pointer(b.ctx), b.ctx.Vtbl.Release)
		return nil, fmt.Errorf("d3d11: feature level too low: %d", featLvl)
	case featLvl >= d3d11.FEATURE_LEVEL_11_0:
		b.caps.MaxTextureSize = 16384
		b.caps.Features |= driver.FeatureCompute
	case featLvl >= d3d11.FEATURE_LEVEL_9_3:
		b.caps.MaxTextureSize = 4096
	}
	if fmt, ok := detectFloatFormat(dev); ok {
		b.floatFormat = fmt
		b.caps.Features |= driver.FeatureFloatRenderTargets
	}
	// Disable backface culling to match OpenGL.
	state, err := dev.CreateRasterizerState(&d3d11.RASTERIZER_DESC{
		CullMode: d3d11.CULL_NONE,
		FillMode: d3d11.FILL_SOLID,
	})
	if err != nil {
		return nil, err
	}
	defer d3d11.IUnknownRelease(unsafe.Pointer(state), state.Vtbl.Release)
	b.ctx.RSSetState(state)
	return b, nil
}

func (b *Backend) BeginFrame(target driver.RenderTarget, clear bool, viewport image.Point) driver.Texture {
	var (
		renderTarget *d3d11.RenderTargetView
	)
	if target != nil {
		switch t := target.(type) {
		case driver.Direct3D11RenderTarget:
			renderTarget = (*d3d11.RenderTargetView)(t.RenderTarget)
		case *Texture:
			renderTarget = t.renderTarget
		default:
			panic(fmt.Errorf("d3d11: invalid render target type: %T", target))
		}
	}
	b.ctx.OMSetRenderTargets(renderTarget, nil)
	return &Texture{backend: b, renderTarget: renderTarget, foreign: true}
}

func (b *Backend) CopyTexture(dstTex driver.Texture, dstOrigin image.Point, srcTex driver.Texture, srcRect image.Rectangle) {
	dst := (*d3d11.Resource)(unsafe.Pointer(dstTex.(*Texture).tex))
	src := (*d3d11.Resource)(srcTex.(*Texture).tex)
	b.ctx.CopySubresourceRegion(
		dst,
		0,                                           // Destination subresource.
		uint32(dstOrigin.X), uint32(dstOrigin.Y), 0, // Destination coordinates (x, y, z).
		src,
		0, // Source subresource.
		&d3d11.BOX{
			Left:   uint32(srcRect.Min.X),
			Top:    uint32(srcRect.Min.Y),
			Right:  uint32(srcRect.Max.X),
			Bottom: uint32(srcRect.Max.Y),
			Front:  0,
			Back:   1,
		},
	)
}

func (b *Backend) EndFrame() {
}

func (b *Backend) Caps() driver.Caps {
	return b.caps
}

func (b *Backend) NewTimer() driver.Timer {
	panic("timers not supported")
}

func (b *Backend) IsTimeContinuous() bool {
	panic("timers not supported")
}

func (b *Backend) Release() {
	d3d11.IUnknownRelease(unsafe.Pointer(b.ctx), b.ctx.Vtbl.Release)
	*b = Backend{}
}

func (b *Backend) NewTexture(format driver.TextureFormat, width, height int, minFilter, magFilter driver.TextureFilter, bindings driver.BufferBinding) (driver.Texture, error) {
	var d3dfmt uint32
	switch format {
	case driver.TextureFormatFloat:
		d3dfmt = b.floatFormat
	case driver.TextureFormatSRGBA:
		d3dfmt = d3d11.DXGI_FORMAT_R8G8B8A8_UNORM_SRGB
	case driver.TextureFormatRGBA8:
		d3dfmt = d3d11.DXGI_FORMAT_R8G8B8A8_UNORM
	default:
		return nil, fmt.Errorf("unsupported texture format %d", format)
	}
	tex, err := b.dev.CreateTexture2D(&d3d11.TEXTURE2D_DESC{
		Width:     uint32(width),
		Height:    uint32(height),
		MipLevels: 1,
		ArraySize: 1,
		Format:    d3dfmt,
		SampleDesc: d3d11.DXGI_SAMPLE_DESC{
			Count:   1,
			Quality: 0,
		},
		BindFlags: convBufferBinding(bindings),
	})
	if err != nil {
		return nil, err
	}
	var (
		sampler *d3d11.SamplerState
		resView *d3d11.ShaderResourceView
		uaView  *d3d11.UnorderedAccessView
		fbo     *d3d11.RenderTargetView
	)
	if bindings&driver.BufferBindingTexture != 0 {
		var filter uint32
		switch {
		case minFilter == driver.FilterNearest && magFilter == driver.FilterNearest:
			filter = d3d11.FILTER_MIN_MAG_MIP_POINT
		case minFilter == driver.FilterLinear && magFilter == driver.FilterLinear:
			filter = d3d11.FILTER_MIN_MAG_LINEAR_MIP_POINT
		default:
			d3d11.IUnknownRelease(unsafe.Pointer(tex), tex.Vtbl.Release)
			return nil, fmt.Errorf("unsupported texture filter combination %d, %d", minFilter, magFilter)
		}
		var err error
		sampler, err = b.dev.CreateSamplerState(&d3d11.SAMPLER_DESC{
			Filter:        filter,
			AddressU:      d3d11.TEXTURE_ADDRESS_CLAMP,
			AddressV:      d3d11.TEXTURE_ADDRESS_CLAMP,
			AddressW:      d3d11.TEXTURE_ADDRESS_CLAMP,
			MaxAnisotropy: 1,
			MinLOD:        -math.MaxFloat32,
			MaxLOD:        math.MaxFloat32,
		})
		if err != nil {
			d3d11.IUnknownRelease(unsafe.Pointer(tex), tex.Vtbl.Release)
			return nil, err
		}
		resView, err = b.dev.CreateShaderResourceView(
			(*d3d11.Resource)(unsafe.Pointer(tex)),
			unsafe.Pointer(&d3d11.SHADER_RESOURCE_VIEW_DESC_TEX2D{
				SHADER_RESOURCE_VIEW_DESC: d3d11.SHADER_RESOURCE_VIEW_DESC{
					Format:        d3dfmt,
					ViewDimension: d3d11.SRV_DIMENSION_TEXTURE2D,
				},
				Texture2D: d3d11.TEX2D_SRV{
					MostDetailedMip: 0,
					MipLevels:       ^uint32(0),
				},
			}),
		)
		if err != nil {
			d3d11.IUnknownRelease(unsafe.Pointer(tex), tex.Vtbl.Release)
			d3d11.IUnknownRelease(unsafe.Pointer(sampler), sampler.Vtbl.Release)
			return nil, err
		}
	}
	if bindings&driver.BufferBindingShaderStorageWrite != 0 {
		uaView, err = b.dev.CreateUnorderedAccessView(
			(*d3d11.Resource)(unsafe.Pointer(tex)),
			unsafe.Pointer(&d3d11.UNORDERED_ACCESS_VIEW_DESC_TEX2D{
				UNORDERED_ACCESS_VIEW_DESC: d3d11.UNORDERED_ACCESS_VIEW_DESC{
					Format:        d3dfmt,
					ViewDimension: d3d11.UAV_DIMENSION_TEXTURE2D,
				},
				Texture2D: d3d11.TEX2D_UAV{
					MipSlice: 0,
				},
			}),
		)
		if err != nil {
			if sampler != nil {
				d3d11.IUnknownRelease(unsafe.Pointer(sampler), sampler.Vtbl.Release)
			}
			if resView != nil {
				d3d11.IUnknownRelease(unsafe.Pointer(resView), resView.Vtbl.Release)
			}
			d3d11.IUnknownRelease(unsafe.Pointer(tex), tex.Vtbl.Release)
			return nil, err
		}
	}
	if bindings&driver.BufferBindingFramebuffer != 0 {
		resource := (*d3d11.Resource)(unsafe.Pointer(tex))
		fbo, err = b.dev.CreateRenderTargetView(resource)
		if err != nil {
			if uaView != nil {
				d3d11.IUnknownRelease(unsafe.Pointer(uaView), uaView.Vtbl.Release)
			}
			if sampler != nil {
				d3d11.IUnknownRelease(unsafe.Pointer(sampler), sampler.Vtbl.Release)
			}
			if resView != nil {
				d3d11.IUnknownRelease(unsafe.Pointer(resView), resView.Vtbl.Release)
			}
			d3d11.IUnknownRelease(unsafe.Pointer(tex), tex.Vtbl.Release)
			return nil, err
		}
	}
	return &Texture{backend: b, format: d3dfmt, tex: tex, sampler: sampler, resView: resView, uaView: uaView, renderTarget: fbo, bindings: bindings, width: width, height: height}, nil
}

func (b *Backend) newInputLayout(vertexShader shader.Sources, layout []driver.InputDesc) (*d3d11.InputLayout, error) {
	if len(vertexShader.Inputs) != len(layout) {
		return nil, fmt.Errorf("NewInputLayout: got %d inputs, expected %d", len(layout), len(vertexShader.Inputs))
	}
	descs := make([]d3d11.INPUT_ELEMENT_DESC, len(layout))
	for i, l := range layout {
		inp := vertexShader.Inputs[i]
		cname, err := windows.BytePtrFromString(inp.Semantic)
		if err != nil {
			return nil, err
		}
		var format uint32
		switch l.Type {
		case shader.DataTypeFloat:
			switch l.Size {
			case 1:
				format = d3d11.DXGI_FORMAT_R32_FLOAT
			case 2:
				format = d3d11.DXGI_FORMAT_R32G32_FLOAT
			case 3:
				format = d3d11.DXGI_FORMAT_R32G32B32_FLOAT
			case 4:
				format = d3d11.DXGI_FORMAT_R32G32B32A32_FLOAT
			default:
				panic("unsupported data size")
			}
		case shader.DataTypeShort:
			switch l.Size {
			case 1:
				format = d3d11.DXGI_FORMAT_R16_SINT
			case 2:
				format = d3d11.DXGI_FORMAT_R16G16_SINT
			default:
				panic("unsupported data size")
			}
		default:
			panic("unsupported data type")
		}
		descs[i] = d3d11.INPUT_ELEMENT_DESC{
			SemanticName:      cname,
			SemanticIndex:     uint32(inp.SemanticIndex),
			Format:            format,
			AlignedByteOffset: uint32(l.Offset),
		}
	}
	return b.dev.CreateInputLayout(descs, []byte(vertexShader.DXBC))
}

func (b *Backend) NewBuffer(typ driver.BufferBinding, size int) (driver.Buffer, error) {
	return b.newBuffer(typ, size, nil, false)
}

func (b *Backend) NewImmutableBuffer(typ driver.BufferBinding, data []byte) (driver.Buffer, error) {
	return b.newBuffer(typ, len(data), data, true)
}

func (b *Backend) newBuffer(typ driver.BufferBinding, size int, data []byte, immutable bool) (*Buffer, error) {
	if typ&driver.BufferBindingUniforms != 0 {
		if typ != driver.BufferBindingUniforms {
			return nil, errors.New("uniform buffers cannot have other bindings")
		}
		if size%16 != 0 {
			return nil, fmt.Errorf("constant buffer size is %d, expected a multiple of 16", size)
		}
	}
	bind := convBufferBinding(typ)
	var usage, miscFlags, cpuFlags uint32
	if immutable {
		usage = d3d11.USAGE_IMMUTABLE
	}
	if typ&driver.BufferBindingShaderStorageWrite != 0 {
		cpuFlags = d3d11.CPU_ACCESS_READ
	}
	if typ&(driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite) != 0 {
		miscFlags |= d3d11.RESOURCE_MISC_BUFFER_ALLOW_RAW_VIEWS
	}
	buf, err := b.dev.CreateBuffer(&d3d11.BUFFER_DESC{
		ByteWidth:      uint32(size),
		Usage:          usage,
		BindFlags:      bind,
		CPUAccessFlags: cpuFlags,
		MiscFlags:      miscFlags,
	}, data)
	if err != nil {
		return nil, err
	}
	var (
		resView *d3d11.ShaderResourceView
		uaView  *d3d11.UnorderedAccessView
	)
	if typ&driver.BufferBindingShaderStorageWrite != 0 {
		uaView, err = b.dev.CreateUnorderedAccessView(
			(*d3d11.Resource)(unsafe.Pointer(buf)),
			unsafe.Pointer(&d3d11.UNORDERED_ACCESS_VIEW_DESC_BUFFER{
				UNORDERED_ACCESS_VIEW_DESC: d3d11.UNORDERED_ACCESS_VIEW_DESC{
					Format:        d3d11.DXGI_FORMAT_R32_TYPELESS,
					ViewDimension: d3d11.UAV_DIMENSION_BUFFER,
				},
				Buffer: d3d11.BUFFER_UAV{
					FirstElement: 0,
					NumElements:  uint32(size / 4),
					Flags:        d3d11.BUFFER_UAV_FLAG_RAW,
				},
			}),
		)
		if err != nil {
			d3d11.IUnknownRelease(unsafe.Pointer(buf), buf.Vtbl.Release)
			return nil, err
		}
	} else if typ&driver.BufferBindingShaderStorageRead != 0 {
		resView, err = b.dev.CreateShaderResourceView(
			(*d3d11.Resource)(unsafe.Pointer(buf)),
			unsafe.Pointer(&d3d11.SHADER_RESOURCE_VIEW_DESC_BUFFEREX{
				SHADER_RESOURCE_VIEW_DESC: d3d11.SHADER_RESOURCE_VIEW_DESC{
					Format:        d3d11.DXGI_FORMAT_R32_TYPELESS,
					ViewDimension: d3d11.SRV_DIMENSION_BUFFEREX,
				},
				Buffer: d3d11.BUFFEREX_SRV{
					FirstElement: 0,
					NumElements:  uint32(size / 4),
					Flags:        d3d11.BUFFEREX_SRV_FLAG_RAW,
				},
			}),
		)
		if err != nil {
			d3d11.IUnknownRelease(unsafe.Pointer(buf), buf.Vtbl.Release)
			return nil, err
		}
	}
	return &Buffer{backend: b, buf: buf, bind: bind, size: size, resView: resView, uaView: uaView, immutable: immutable}, nil
}

func (b *Backend) NewComputeProgram(shader shader.Sources) (driver.Program, error) {
	cs, err := b.dev.CreateComputeShader([]byte(shader.DXBC))
	if err != nil {
		return nil, err
	}
	return &Program{backend: b, shader: cs}, nil
}

func (b *Backend) NewPipeline(desc driver.PipelineDesc) (driver.Pipeline, error) {
	vsh := desc.VertexShader.(*VertexShader)
	fsh := desc.FragmentShader.(*FragmentShader)
	blend, err := b.newBlendState(desc.BlendDesc)
	if err != nil {
		return nil, err
	}
	var layout *d3d11.InputLayout
	if l := desc.VertexLayout; l.Stride > 0 {
		var err error
		layout, err = b.newInputLayout(vsh.src, l.Inputs)
		if err != nil {
			d3d11.IUnknownRelease(unsafe.Pointer(blend), blend.Vtbl.AddRef)
			return nil, err
		}
	}

	// Retain shaders.
	vshRef := vsh.shader
	fshRef := fsh.shader
	d3d11.IUnknownAddRef(unsafe.Pointer(vshRef), vshRef.Vtbl.AddRef)
	d3d11.IUnknownAddRef(unsafe.Pointer(fshRef), fshRef.Vtbl.AddRef)

	return &Pipeline{
		vert:     vshRef,
		frag:     fshRef,
		layout:   layout,
		stride:   desc.VertexLayout.Stride,
		blend:    blend,
		topology: desc.Topology,
	}, nil
}

func (b *Backend) newBlendState(desc driver.BlendDesc) (*d3d11.BlendState, error) {
	var d3ddesc d3d11.BLEND_DESC
	t0 := &d3ddesc.RenderTarget[0]
	t0.RenderTargetWriteMask = d3d11.COLOR_WRITE_ENABLE_ALL
	t0.BlendOp = d3d11.BLEND_OP_ADD
	t0.BlendOpAlpha = d3d11.BLEND_OP_ADD
	if desc.Enable {
		t0.BlendEnable = 1
	}
	scol, salpha := toBlendFactor(desc.SrcFactor)
	dcol, dalpha := toBlendFactor(desc.DstFactor)
	t0.SrcBlend = scol
	t0.SrcBlendAlpha = salpha
	t0.DestBlend = dcol
	t0.DestBlendAlpha = dalpha
	return b.dev.CreateBlendState(&d3ddesc)
}

func (b *Backend) NewVertexShader(src shader.Sources) (driver.VertexShader, error) {
	vs, err := b.dev.CreateVertexShader([]byte(src.DXBC))
	if err != nil {
		return nil, err
	}
	return &VertexShader{b, vs, src}, nil
}

func (b *Backend) NewFragmentShader(src shader.Sources) (driver.FragmentShader, error) {
	fs, err := b.dev.CreatePixelShader([]byte(src.DXBC))
	if err != nil {
		return nil, err
	}
	return &FragmentShader{b, fs}, nil
}

func (b *Backend) Viewport(x, y, width, height int) {
	b.viewport = d3d11.VIEWPORT{
		TopLeftX: float32(x),
		TopLeftY: float32(y),
		Width:    float32(width),
		Height:   float32(height),
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}
	b.ctx.RSSetViewports(&b.viewport)
}

func (b *Backend) DrawArrays(off, count int) {
	b.prepareDraw()
	b.ctx.Draw(uint32(count), uint32(off))
}

func (b *Backend) DrawElements(off, count int) {
	b.prepareDraw()
	b.ctx.DrawIndexed(uint32(count), uint32(off), 0)
}

func (b *Backend) prepareDraw() {
	p := b.pipeline
	if p == nil {
		return
	}
	b.ctx.VSSetShader(p.vert)
	b.ctx.PSSetShader(p.frag)
	b.ctx.IASetInputLayout(p.layout)
	b.ctx.OMSetBlendState(p.blend, nil, 0xffffffff)
	if b.vert.buffer != nil {
		b.ctx.IASetVertexBuffers(b.vert.buffer.buf, uint32(p.stride), uint32(b.vert.offset))
	}
	var topology uint32
	switch p.topology {
	case driver.TopologyTriangles:
		topology = d3d11.PRIMITIVE_TOPOLOGY_TRIANGLELIST
	case driver.TopologyTriangleStrip:
		topology = d3d11.PRIMITIVE_TOPOLOGY_TRIANGLESTRIP
	default:
		panic("unsupported draw mode")
	}
	b.ctx.IASetPrimitiveTopology(topology)
}

func (b *Backend) BindImageTexture(unit int, tex driver.Texture) {
	t := tex.(*Texture)
	if t.uaView != nil {
		b.ctx.CSSetUnorderedAccessViews(uint32(unit), t.uaView)
	} else {
		b.ctx.CSSetShaderResources(uint32(unit), t.resView)
	}
}

func (b *Backend) DispatchCompute(x, y, z int) {
	b.ctx.CSSetShader(b.program.shader)
	b.ctx.Dispatch(uint32(x), uint32(y), uint32(z))
}

func (t *Texture) Upload(offset, size image.Point, pixels []byte, stride int) {
	if stride == 0 {
		stride = size.X * 4
	}
	dst := &d3d11.BOX{
		Left:   uint32(offset.X),
		Top:    uint32(offset.Y),
		Right:  uint32(offset.X + size.X),
		Bottom: uint32(offset.Y + size.Y),
		Front:  0,
		Back:   1,
	}
	res := (*d3d11.Resource)(unsafe.Pointer(t.tex))
	t.backend.ctx.UpdateSubresource(res, dst, uint32(stride), uint32(len(pixels)), pixels)
}

func (t *Texture) Release() {
	if t.foreign {
		panic("texture not created by NewTexture")
	}
	if t.renderTarget != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(t.renderTarget), t.renderTarget.Vtbl.Release)
	}
	if t.sampler != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(t.sampler), t.sampler.Vtbl.Release)
	}
	if t.resView != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(t.resView), t.resView.Vtbl.Release)
	}
	if t.uaView != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(t.uaView), t.uaView.Vtbl.Release)
	}
	d3d11.IUnknownRelease(unsafe.Pointer(t.tex), t.tex.Vtbl.Release)
	*t = Texture{}
}

func (b *Backend) PrepareTexture(tex driver.Texture) {}

func (b *Backend) BindTexture(unit int, tex driver.Texture) {
	t := tex.(*Texture)
	b.ctx.PSSetSamplers(uint32(unit), t.sampler)
	b.ctx.PSSetShaderResources(uint32(unit), t.resView)
}

func (b *Backend) BindPipeline(pipe driver.Pipeline) {
	b.pipeline = pipe.(*Pipeline)
}

func (b *Backend) BindProgram(prog driver.Program) {
	b.program = prog.(*Program)
}

func (s *VertexShader) Release() {
	d3d11.IUnknownRelease(unsafe.Pointer(s.shader), s.shader.Vtbl.Release)
	*s = VertexShader{}
}

func (s *FragmentShader) Release() {
	d3d11.IUnknownRelease(unsafe.Pointer(s.shader), s.shader.Vtbl.Release)
	*s = FragmentShader{}
}

func (s *Program) Release() {
	d3d11.IUnknownRelease(unsafe.Pointer(s.shader), s.shader.Vtbl.Release)
	*s = Program{}
}

func (p *Pipeline) Release() {
	d3d11.IUnknownRelease(unsafe.Pointer(p.vert), p.vert.Vtbl.Release)
	d3d11.IUnknownRelease(unsafe.Pointer(p.frag), p.frag.Vtbl.Release)
	d3d11.IUnknownRelease(unsafe.Pointer(p.blend), p.blend.Vtbl.Release)
	if l := p.layout; l != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(l), l.Vtbl.Release)
	}
	*p = Pipeline{}
}

func (b *Backend) BindStorageBuffer(binding int, buffer driver.Buffer) {
	buf := buffer.(*Buffer)
	if buf.resView != nil {
		b.ctx.CSSetShaderResources(uint32(binding), buf.resView)
	} else {
		b.ctx.CSSetUnorderedAccessViews(uint32(binding), buf.uaView)
	}
}

func (b *Backend) BindUniforms(buffer driver.Buffer) {
	buf := buffer.(*Buffer)
	b.ctx.VSSetConstantBuffers(buf.buf)
	b.ctx.PSSetConstantBuffers(buf.buf)
}

func (b *Backend) BindVertexBuffer(buf driver.Buffer, offset int) {
	b.vert.buffer = buf.(*Buffer)
	b.vert.offset = offset
}

func (b *Backend) BindIndexBuffer(buf driver.Buffer) {
	b.ctx.IASetIndexBuffer(buf.(*Buffer).buf, d3d11.DXGI_FORMAT_R16_UINT, 0)
}

func (b *Buffer) Download(dst []byte) error {
	res := (*d3d11.Resource)(unsafe.Pointer(b.buf))
	resMap, err := b.backend.ctx.Map(res, 0, d3d11.MAP_READ, 0)
	if err != nil {
		return fmt.Errorf("d3d11: %v", err)
	}
	defer b.backend.ctx.Unmap(res, 0)
	data := sliceOf(resMap.PData, len(dst))
	copy(dst, data)
	return nil
}

func (b *Buffer) Upload(data []byte) {
	var dst *d3d11.BOX
	if len(data) < b.size {
		dst = &d3d11.BOX{
			Left:   0,
			Right:  uint32(len(data)),
			Top:    0,
			Bottom: 1,
			Front:  0,
			Back:   1,
		}
	}
	b.backend.ctx.UpdateSubresource((*d3d11.Resource)(unsafe.Pointer(b.buf)), dst, 0, 0, data)
}

func (b *Buffer) Release() {
	if b.resView != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(b.resView), b.resView.Vtbl.Release)
	}
	if b.uaView != nil {
		d3d11.IUnknownRelease(unsafe.Pointer(b.uaView), b.uaView.Vtbl.Release)
	}
	d3d11.IUnknownRelease(unsafe.Pointer(b.buf), b.buf.Vtbl.Release)
	*b = Buffer{}
}

func (t *Texture) ReadPixels(src image.Rectangle, pixels []byte, stride int) error {
	w, h := src.Dx(), src.Dy()
	tex, err := t.backend.dev.CreateTexture2D(&d3d11.TEXTURE2D_DESC{
		Width:     uint32(w),
		Height:    uint32(h),
		MipLevels: 1,
		ArraySize: 1,
		Format:    t.format,
		SampleDesc: d3d11.DXGI_SAMPLE_DESC{
			Count:   1,
			Quality: 0,
		},
		Usage:          d3d11.USAGE_STAGING,
		CPUAccessFlags: d3d11.CPU_ACCESS_READ,
	})
	if err != nil {
		return fmt.Errorf("ReadPixels: %v", err)
	}
	defer d3d11.IUnknownRelease(unsafe.Pointer(tex), tex.Vtbl.Release)
	res := (*d3d11.Resource)(unsafe.Pointer(tex))
	t.backend.ctx.CopySubresourceRegion(
		res,
		0,       // Destination subresource.
		0, 0, 0, // Destination coordinates (x, y, z).
		(*d3d11.Resource)(t.tex),
		0, // Source subresource.
		&d3d11.BOX{
			Left:   uint32(src.Min.X),
			Top:    uint32(src.Min.Y),
			Right:  uint32(src.Max.X),
			Bottom: uint32(src.Max.Y),
			Front:  0,
			Back:   1,
		},
	)
	resMap, err := t.backend.ctx.Map(res, 0, d3d11.MAP_READ, 0)
	if err != nil {
		return fmt.Errorf("ReadPixels: %v", err)
	}
	defer t.backend.ctx.Unmap(res, 0)
	srcPitch := stride
	dstPitch := int(resMap.RowPitch)
	mapSize := dstPitch * h
	data := sliceOf(resMap.PData, mapSize)
	width := w * 4
	for r := 0; r < h; r++ {
		pixels := pixels[r*srcPitch:]
		copy(pixels[:width], data[r*dstPitch:])
	}
	return nil
}

func (b *Backend) BeginCompute() {
}

func (b *Backend) EndCompute() {
}

func (b *Backend) BeginRenderPass(tex driver.Texture, d driver.LoadDesc) {
	t := tex.(*Texture)
	b.ctx.OMSetRenderTargets(t.renderTarget, nil)
	if d.Action == driver.LoadActionClear {
		c := d.ClearColor
		b.clearColor = [4]float32{c.R, c.G, c.B, c.A}
		b.ctx.ClearRenderTargetView(t.renderTarget, &b.clearColor)
	}
}

func (b *Backend) EndRenderPass() {
}

func (f *Texture) ImplementsRenderTarget() {}

func convBufferBinding(typ driver.BufferBinding) uint32 {
	var bindings uint32
	if typ&driver.BufferBindingVertices != 0 {
		bindings |= d3d11.BIND_VERTEX_BUFFER
	}
	if typ&driver.BufferBindingIndices != 0 {
		bindings |= d3d11.BIND_INDEX_BUFFER
	}
	if typ&driver.BufferBindingUniforms != 0 {
		bindings |= d3d11.BIND_CONSTANT_BUFFER
	}
	if typ&driver.BufferBindingTexture != 0 {
		bindings |= d3d11.BIND_SHADER_RESOURCE
	}
	if typ&driver.BufferBindingFramebuffer != 0 {
		bindings |= d3d11.BIND_RENDER_TARGET
	}
	if typ&driver.BufferBindingShaderStorageWrite != 0 {
		bindings |= d3d11.BIND_UNORDERED_ACCESS
	} else if typ&driver.BufferBindingShaderStorageRead != 0 {
		bindings |= d3d11.BIND_SHADER_RESOURCE
	}
	return bindings
}

func toBlendFactor(f driver.BlendFactor) (uint32, uint32) {
	switch f {
	case driver.BlendFactorOne:
		return d3d11.BLEND_ONE, d3d11.BLEND_ONE
	case driver.BlendFactorOneMinusSrcAlpha:
		return d3d11.BLEND_INV_SRC_ALPHA, d3d11.BLEND_INV_SRC_ALPHA
	case driver.BlendFactorZero:
		return d3d11.BLEND_ZERO, d3d11.BLEND_ZERO
	case driver.BlendFactorDstColor:
		return d3d11.BLEND_DEST_COLOR, d3d11.BLEND_DEST_ALPHA
	default:
		panic("unsupported blend source factor")
	}
}

// sliceOf returns a slice from a (native) pointer.
func sliceOf(ptr uintptr, cap int) []byte {
	var data []byte
	h := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	h.Data = ptr
	h.Cap = cap
	h.Len = cap
	return data
}
