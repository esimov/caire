// SPDX-License-Identifier: Unlicense OR MIT

package d3d11

import (
	"fmt"
	"math"
	"syscall"
	"unsafe"

	"gioui.org/internal/f32color"

	"golang.org/x/sys/windows"
)

type DXGI_SWAP_CHAIN_DESC struct {
	BufferDesc   DXGI_MODE_DESC
	SampleDesc   DXGI_SAMPLE_DESC
	BufferUsage  uint32
	BufferCount  uint32
	OutputWindow windows.Handle
	Windowed     uint32
	SwapEffect   uint32
	Flags        uint32
}

type DXGI_SAMPLE_DESC struct {
	Count   uint32
	Quality uint32
}

type DXGI_MODE_DESC struct {
	Width            uint32
	Height           uint32
	RefreshRate      DXGI_RATIONAL
	Format           uint32
	ScanlineOrdering uint32
	Scaling          uint32
}

type DXGI_RATIONAL struct {
	Numerator   uint32
	Denominator uint32
}

type TEXTURE2D_DESC struct {
	Width          uint32
	Height         uint32
	MipLevels      uint32
	ArraySize      uint32
	Format         uint32
	SampleDesc     DXGI_SAMPLE_DESC
	Usage          uint32
	BindFlags      uint32
	CPUAccessFlags uint32
	MiscFlags      uint32
}

type SAMPLER_DESC struct {
	Filter         uint32
	AddressU       uint32
	AddressV       uint32
	AddressW       uint32
	MipLODBias     float32
	MaxAnisotropy  uint32
	ComparisonFunc uint32
	BorderColor    [4]float32
	MinLOD         float32
	MaxLOD         float32
}

type SHADER_RESOURCE_VIEW_DESC_TEX2D struct {
	SHADER_RESOURCE_VIEW_DESC
	Texture2D TEX2D_SRV
}

type SHADER_RESOURCE_VIEW_DESC_BUFFEREX struct {
	SHADER_RESOURCE_VIEW_DESC
	Buffer BUFFEREX_SRV
}

type UNORDERED_ACCESS_VIEW_DESC_TEX2D struct {
	UNORDERED_ACCESS_VIEW_DESC
	Texture2D TEX2D_UAV
}

type UNORDERED_ACCESS_VIEW_DESC_BUFFER struct {
	UNORDERED_ACCESS_VIEW_DESC
	Buffer BUFFER_UAV
}

type SHADER_RESOURCE_VIEW_DESC struct {
	Format        uint32
	ViewDimension uint32
}

type UNORDERED_ACCESS_VIEW_DESC struct {
	Format        uint32
	ViewDimension uint32
}

type TEX2D_SRV struct {
	MostDetailedMip uint32
	MipLevels       uint32
}

type BUFFEREX_SRV struct {
	FirstElement uint32
	NumElements  uint32
	Flags        uint32
}

type TEX2D_UAV struct {
	MipSlice uint32
}

type BUFFER_UAV struct {
	FirstElement uint32
	NumElements  uint32
	Flags        uint32
}

type INPUT_ELEMENT_DESC struct {
	SemanticName         *byte
	SemanticIndex        uint32
	Format               uint32
	InputSlot            uint32
	AlignedByteOffset    uint32
	InputSlotClass       uint32
	InstanceDataStepRate uint32
}

type IDXGISwapChain struct {
	Vtbl *struct {
		_IUnknownVTbl
		SetPrivateData          uintptr
		SetPrivateDataInterface uintptr
		GetPrivateData          uintptr
		GetParent               uintptr
		GetDevice               uintptr
		Present                 uintptr
		GetBuffer               uintptr
		SetFullscreenState      uintptr
		GetFullscreenState      uintptr
		GetDesc                 uintptr
		ResizeBuffers           uintptr
		ResizeTarget            uintptr
		GetContainingOutput     uintptr
		GetFrameStatistics      uintptr
		GetLastPresentCount     uintptr
	}
}

type Debug struct {
	Vtbl *struct {
		_IUnknownVTbl
		SetFeatureMask             uintptr
		GetFeatureMask             uintptr
		SetPresentPerRenderOpDelay uintptr
		GetPresentPerRenderOpDelay uintptr
		SetSwapChain               uintptr
		GetSwapChain               uintptr
		ValidateContext            uintptr
		ReportLiveDeviceObjects    uintptr
		ValidateContextForDispatch uintptr
	}
}

type Device struct {
	Vtbl *struct {
		_IUnknownVTbl
		CreateBuffer                         uintptr
		CreateTexture1D                      uintptr
		CreateTexture2D                      uintptr
		CreateTexture3D                      uintptr
		CreateShaderResourceView             uintptr
		CreateUnorderedAccessView            uintptr
		CreateRenderTargetView               uintptr
		CreateDepthStencilView               uintptr
		CreateInputLayout                    uintptr
		CreateVertexShader                   uintptr
		CreateGeometryShader                 uintptr
		CreateGeometryShaderWithStreamOutput uintptr
		CreatePixelShader                    uintptr
		CreateHullShader                     uintptr
		CreateDomainShader                   uintptr
		CreateComputeShader                  uintptr
		CreateClassLinkage                   uintptr
		CreateBlendState                     uintptr
		CreateDepthStencilState              uintptr
		CreateRasterizerState                uintptr
		CreateSamplerState                   uintptr
		CreateQuery                          uintptr
		CreatePredicate                      uintptr
		CreateCounter                        uintptr
		CreateDeferredContext                uintptr
		OpenSharedResource                   uintptr
		CheckFormatSupport                   uintptr
		CheckMultisampleQualityLevels        uintptr
		CheckCounterInfo                     uintptr
		CheckCounter                         uintptr
		CheckFeatureSupport                  uintptr
		GetPrivateData                       uintptr
		SetPrivateData                       uintptr
		SetPrivateDataInterface              uintptr
		GetFeatureLevel                      uintptr
		GetCreationFlags                     uintptr
		GetDeviceRemovedReason               uintptr
		GetImmediateContext                  uintptr
		SetExceptionMode                     uintptr
		GetExceptionMode                     uintptr
	}
}

type DeviceContext struct {
	Vtbl *struct {
		_IUnknownVTbl
		GetDevice                                 uintptr
		GetPrivateData                            uintptr
		SetPrivateData                            uintptr
		SetPrivateDataInterface                   uintptr
		VSSetConstantBuffers                      uintptr
		PSSetShaderResources                      uintptr
		PSSetShader                               uintptr
		PSSetSamplers                             uintptr
		VSSetShader                               uintptr
		DrawIndexed                               uintptr
		Draw                                      uintptr
		Map                                       uintptr
		Unmap                                     uintptr
		PSSetConstantBuffers                      uintptr
		IASetInputLayout                          uintptr
		IASetVertexBuffers                        uintptr
		IASetIndexBuffer                          uintptr
		DrawIndexedInstanced                      uintptr
		DrawInstanced                             uintptr
		GSSetConstantBuffers                      uintptr
		GSSetShader                               uintptr
		IASetPrimitiveTopology                    uintptr
		VSSetShaderResources                      uintptr
		VSSetSamplers                             uintptr
		Begin                                     uintptr
		End                                       uintptr
		GetData                                   uintptr
		SetPredication                            uintptr
		GSSetShaderResources                      uintptr
		GSSetSamplers                             uintptr
		OMSetRenderTargets                        uintptr
		OMSetRenderTargetsAndUnorderedAccessViews uintptr
		OMSetBlendState                           uintptr
		OMSetDepthStencilState                    uintptr
		SOSetTargets                              uintptr
		DrawAuto                                  uintptr
		DrawIndexedInstancedIndirect              uintptr
		DrawInstancedIndirect                     uintptr
		Dispatch                                  uintptr
		DispatchIndirect                          uintptr
		RSSetState                                uintptr
		RSSetViewports                            uintptr
		RSSetScissorRects                         uintptr
		CopySubresourceRegion                     uintptr
		CopyResource                              uintptr
		UpdateSubresource                         uintptr
		CopyStructureCount                        uintptr
		ClearRenderTargetView                     uintptr
		ClearUnorderedAccessViewUint              uintptr
		ClearUnorderedAccessViewFloat             uintptr
		ClearDepthStencilView                     uintptr
		GenerateMips                              uintptr
		SetResourceMinLOD                         uintptr
		GetResourceMinLOD                         uintptr
		ResolveSubresource                        uintptr
		ExecuteCommandList                        uintptr
		HSSetShaderResources                      uintptr
		HSSetShader                               uintptr
		HSSetSamplers                             uintptr
		HSSetConstantBuffers                      uintptr
		DSSetShaderResources                      uintptr
		DSSetShader                               uintptr
		DSSetSamplers                             uintptr
		DSSetConstantBuffers                      uintptr
		CSSetShaderResources                      uintptr
		CSSetUnorderedAccessViews                 uintptr
		CSSetShader                               uintptr
		CSSetSamplers                             uintptr
		CSSetConstantBuffers                      uintptr
		VSGetConstantBuffers                      uintptr
		PSGetShaderResources                      uintptr
		PSGetShader                               uintptr
		PSGetSamplers                             uintptr
		VSGetShader                               uintptr
		PSGetConstantBuffers                      uintptr
		IAGetInputLayout                          uintptr
		IAGetVertexBuffers                        uintptr
		IAGetIndexBuffer                          uintptr
		GSGetConstantBuffers                      uintptr
		GSGetShader                               uintptr
		IAGetPrimitiveTopology                    uintptr
		VSGetShaderResources                      uintptr
		VSGetSamplers                             uintptr
		GetPredication                            uintptr
		GSGetShaderResources                      uintptr
		GSGetSamplers                             uintptr
		OMGetRenderTargets                        uintptr
		OMGetRenderTargetsAndUnorderedAccessViews uintptr
		OMGetBlendState                           uintptr
		OMGetDepthStencilState                    uintptr
		SOGetTargets                              uintptr
		RSGetState                                uintptr
		RSGetViewports                            uintptr
		RSGetScissorRects                         uintptr
		HSGetShaderResources                      uintptr
		HSGetShader                               uintptr
		HSGetSamplers                             uintptr
		HSGetConstantBuffers                      uintptr
		DSGetShaderResources                      uintptr
		DSGetShader                               uintptr
		DSGetSamplers                             uintptr
		DSGetConstantBuffers                      uintptr
		CSGetShaderResources                      uintptr
		CSGetUnorderedAccessViews                 uintptr
		CSGetShader                               uintptr
		CSGetSamplers                             uintptr
		CSGetConstantBuffers                      uintptr
		ClearState                                uintptr
		Flush                                     uintptr
		GetType                                   uintptr
		GetContextFlags                           uintptr
		FinishCommandList                         uintptr
	}
}

type RenderTargetView struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type Resource struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type Texture2D struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type Buffer struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type SamplerState struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type PixelShader struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type ShaderResourceView struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type UnorderedAccessView struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type DepthStencilView struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type BlendState struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type DepthStencilState struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type VertexShader struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type ComputeShader struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type RasterizerState struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type InputLayout struct {
	Vtbl *struct {
		_IUnknownVTbl
		GetBufferPointer uintptr
		GetBufferSize    uintptr
	}
}

type DEPTH_STENCIL_DESC struct {
	DepthEnable      uint32
	DepthWriteMask   uint32
	DepthFunc        uint32
	StencilEnable    uint32
	StencilReadMask  uint8
	StencilWriteMask uint8
	FrontFace        DEPTH_STENCILOP_DESC
	BackFace         DEPTH_STENCILOP_DESC
}

type DEPTH_STENCILOP_DESC struct {
	StencilFailOp      uint32
	StencilDepthFailOp uint32
	StencilPassOp      uint32
	StencilFunc        uint32
}

type DEPTH_STENCIL_VIEW_DESC_TEX2D struct {
	Format        uint32
	ViewDimension uint32
	Flags         uint32
	Texture2D     TEX2D_DSV
}

type TEX2D_DSV struct {
	MipSlice uint32
}

type BLEND_DESC struct {
	AlphaToCoverageEnable  uint32
	IndependentBlendEnable uint32
	RenderTarget           [8]RENDER_TARGET_BLEND_DESC
}

type RENDER_TARGET_BLEND_DESC struct {
	BlendEnable           uint32
	SrcBlend              uint32
	DestBlend             uint32
	BlendOp               uint32
	SrcBlendAlpha         uint32
	DestBlendAlpha        uint32
	BlendOpAlpha          uint32
	RenderTargetWriteMask uint8
}

type IDXGIObject struct {
	Vtbl *struct {
		_IUnknownVTbl
		SetPrivateData          uintptr
		SetPrivateDataInterface uintptr
		GetPrivateData          uintptr
		GetParent               uintptr
	}
}

type IDXGIAdapter struct {
	Vtbl *struct {
		_IUnknownVTbl
		SetPrivateData          uintptr
		SetPrivateDataInterface uintptr
		GetPrivateData          uintptr
		GetParent               uintptr
		EnumOutputs             uintptr
		GetDesc                 uintptr
		CheckInterfaceSupport   uintptr
		GetDesc1                uintptr
	}
}

type IDXGIFactory struct {
	Vtbl *struct {
		_IUnknownVTbl
		SetPrivateData          uintptr
		SetPrivateDataInterface uintptr
		GetPrivateData          uintptr
		GetParent               uintptr
		EnumAdapters            uintptr
		MakeWindowAssociation   uintptr
		GetWindowAssociation    uintptr
		CreateSwapChain         uintptr
		CreateSoftwareAdapter   uintptr
	}
}

type IDXGIDebug struct {
	Vtbl *struct {
		_IUnknownVTbl
		ReportLiveObjects uintptr
	}
}

type IDXGIDevice struct {
	Vtbl *struct {
		_IUnknownVTbl
		SetPrivateData          uintptr
		SetPrivateDataInterface uintptr
		GetPrivateData          uintptr
		GetParent               uintptr
		GetAdapter              uintptr
		CreateSurface           uintptr
		QueryResourceResidency  uintptr
		SetGPUThreadPriority    uintptr
		GetGPUThreadPriority    uintptr
	}
}

type IUnknown struct {
	Vtbl *struct {
		_IUnknownVTbl
	}
}

type _IUnknownVTbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type BUFFER_DESC struct {
	ByteWidth           uint32
	Usage               uint32
	BindFlags           uint32
	CPUAccessFlags      uint32
	MiscFlags           uint32
	StructureByteStride uint32
}

type GUID struct {
	Data1   uint32
	Data2   uint16
	Data3   uint16
	Data4_0 uint8
	Data4_1 uint8
	Data4_2 uint8
	Data4_3 uint8
	Data4_4 uint8
	Data4_5 uint8
	Data4_6 uint8
	Data4_7 uint8
}

type VIEWPORT struct {
	TopLeftX float32
	TopLeftY float32
	Width    float32
	Height   float32
	MinDepth float32
	MaxDepth float32
}

type SUBRESOURCE_DATA struct {
	pSysMem *byte
}

type BOX struct {
	Left   uint32
	Top    uint32
	Front  uint32
	Right  uint32
	Bottom uint32
	Back   uint32
}

type MAPPED_SUBRESOURCE struct {
	PData      uintptr
	RowPitch   uint32
	DepthPitch uint32
}

type ErrorCode struct {
	Name string
	Code uint32
}

type RASTERIZER_DESC struct {
	FillMode              uint32
	CullMode              uint32
	FrontCounterClockwise uint32
	DepthBias             int32
	DepthBiasClamp        float32
	SlopeScaledDepthBias  float32
	DepthClipEnable       uint32
	ScissorEnable         uint32
	MultisampleEnable     uint32
	AntialiasedLineEnable uint32
}

var (
	IID_Texture2D    = GUID{0x6f15aaf2, 0xd208, 0x4e89, 0x9a, 0xb4, 0x48, 0x95, 0x35, 0xd3, 0x4f, 0x9c}
	IID_IDXGIDebug   = GUID{0x119E7452, 0xDE9E, 0x40fe, 0x88, 0x06, 0x88, 0xF9, 0x0C, 0x12, 0xB4, 0x41}
	IID_IDXGIDevice  = GUID{0x54ec77fa, 0x1377, 0x44e6, 0x8c, 0x32, 0x88, 0xfd, 0x5f, 0x44, 0xc8, 0x4c}
	IID_IDXGIFactory = GUID{0x7b7166ec, 0x21c7, 0x44ae, 0xb2, 0x1a, 0xc9, 0xae, 0x32, 0x1a, 0xe3, 0x69}
	IID_ID3D11Debug  = GUID{0x79cf2233, 0x7536, 0x4948, 0x9d, 0x36, 0x1e, 0x46, 0x92, 0xdc, 0x57, 0x60}

	DXGI_DEBUG_ALL = GUID{0xe48ae283, 0xda80, 0x490b, 0x87, 0xe6, 0x43, 0xe9, 0xa9, 0xcf, 0xda, 0x8}
)

var (
	d3d11 = windows.NewLazySystemDLL("d3d11.dll")

	_D3D11CreateDevice             = d3d11.NewProc("D3D11CreateDevice")
	_D3D11CreateDeviceAndSwapChain = d3d11.NewProc("D3D11CreateDeviceAndSwapChain")

	dxgi = windows.NewLazySystemDLL("dxgi.dll")

	_DXGIGetDebugInterface1 = dxgi.NewProc("DXGIGetDebugInterface1")
)

const (
	SDK_VERSION          = 7
	DRIVER_TYPE_HARDWARE = 1

	DXGI_FORMAT_UNKNOWN             = 0
	DXGI_FORMAT_R16_FLOAT           = 54
	DXGI_FORMAT_R32_FLOAT           = 41
	DXGI_FORMAT_R32_TYPELESS        = 39
	DXGI_FORMAT_R32G32_FLOAT        = 16
	DXGI_FORMAT_R32G32B32_FLOAT     = 6
	DXGI_FORMAT_R32G32B32A32_FLOAT  = 2
	DXGI_FORMAT_R8G8B8A8_UNORM      = 28
	DXGI_FORMAT_R8G8B8A8_UNORM_SRGB = 29
	DXGI_FORMAT_R16_SINT            = 59
	DXGI_FORMAT_R16G16_SINT         = 38
	DXGI_FORMAT_R16_UINT            = 57
	DXGI_FORMAT_D24_UNORM_S8_UINT   = 45
	DXGI_FORMAT_R16G16_FLOAT        = 34
	DXGI_FORMAT_R16G16B16A16_FLOAT  = 10

	DXGI_DEBUG_RLO_SUMMARY         = 0x1
	DXGI_DEBUG_RLO_DETAIL          = 0x2
	DXGI_DEBUG_RLO_IGNORE_INTERNAL = 0x4

	FORMAT_SUPPORT_TEXTURE2D     = 0x20
	FORMAT_SUPPORT_RENDER_TARGET = 0x4000

	DXGI_USAGE_RENDER_TARGET_OUTPUT = 1 << (1 + 4)

	CPU_ACCESS_READ = 0x20000

	MAP_READ = 1

	DXGI_SWAP_EFFECT_DISCARD = 0

	FEATURE_LEVEL_9_1  = 0x9100
	FEATURE_LEVEL_9_3  = 0x9300
	FEATURE_LEVEL_11_0 = 0xb000

	USAGE_IMMUTABLE = 1
	USAGE_STAGING   = 3

	BIND_VERTEX_BUFFER    = 0x1
	BIND_INDEX_BUFFER     = 0x2
	BIND_CONSTANT_BUFFER  = 0x4
	BIND_SHADER_RESOURCE  = 0x8
	BIND_RENDER_TARGET    = 0x20
	BIND_DEPTH_STENCIL    = 0x40
	BIND_UNORDERED_ACCESS = 0x80

	PRIMITIVE_TOPOLOGY_TRIANGLELIST  = 4
	PRIMITIVE_TOPOLOGY_TRIANGLESTRIP = 5

	FILTER_MIN_MAG_LINEAR_MIP_POINT = 0x14
	FILTER_MIN_MAG_MIP_POINT        = 0

	TEXTURE_ADDRESS_MIRROR = 2
	TEXTURE_ADDRESS_CLAMP  = 3
	TEXTURE_ADDRESS_WRAP   = 1

	SRV_DIMENSION_BUFFER    = 1
	UAV_DIMENSION_BUFFER    = 1
	SRV_DIMENSION_BUFFEREX  = 11
	SRV_DIMENSION_TEXTURE2D = 4
	UAV_DIMENSION_TEXTURE2D = 4

	BUFFER_UAV_FLAG_RAW   = 0x1
	BUFFEREX_SRV_FLAG_RAW = 0x1

	RESOURCE_MISC_BUFFER_ALLOW_RAW_VIEWS = 0x20

	CREATE_DEVICE_DEBUG = 0x2

	FILL_SOLID = 3

	CULL_NONE = 1

	CLEAR_DEPTH   = 0x1
	CLEAR_STENCIL = 0x2

	DSV_DIMENSION_TEXTURE2D = 3

	DEPTH_WRITE_MASK_ALL = 1

	COMPARISON_GREATER       = 5
	COMPARISON_GREATER_EQUAL = 7

	BLEND_OP_ADD        = 1
	BLEND_ONE           = 2
	BLEND_INV_SRC_ALPHA = 6
	BLEND_ZERO          = 1
	BLEND_DEST_COLOR    = 9
	BLEND_DEST_ALPHA    = 7

	COLOR_WRITE_ENABLE_ALL = 1 | 2 | 4 | 8

	DXGI_STATUS_OCCLUDED      = 0x087A0001
	DXGI_ERROR_DEVICE_RESET   = 0x887A0007
	DXGI_ERROR_DEVICE_REMOVED = 0x887A0005
	D3DDDIERR_DEVICEREMOVED   = 1<<31 | 0x876<<16 | 2160

	RLDO_SUMMARY         = 1
	RLDO_DETAIL          = 2
	RLDO_IGNORE_INTERNAL = 4
)

func CreateDevice(driverType uint32, flags uint32) (*Device, *DeviceContext, uint32, error) {
	var (
		dev     *Device
		ctx     *DeviceContext
		featLvl uint32
	)
	r, _, _ := _D3D11CreateDevice.Call(
		0,                                 // pAdapter
		uintptr(driverType),               // driverType
		0,                                 // Software
		uintptr(flags),                    // Flags
		0,                                 // pFeatureLevels
		0,                                 // FeatureLevels
		SDK_VERSION,                       // SDKVersion
		uintptr(unsafe.Pointer(&dev)),     // ppDevice
		uintptr(unsafe.Pointer(&featLvl)), // pFeatureLevel
		uintptr(unsafe.Pointer(&ctx)),     // ppImmediateContext
	)
	if r != 0 {
		return nil, nil, 0, ErrorCode{Name: "D3D11CreateDevice", Code: uint32(r)}
	}
	return dev, ctx, featLvl, nil
}

func CreateDeviceAndSwapChain(driverType uint32, flags uint32, swapDesc *DXGI_SWAP_CHAIN_DESC) (*Device, *DeviceContext, *IDXGISwapChain, uint32, error) {
	var (
		dev     *Device
		ctx     *DeviceContext
		swchain *IDXGISwapChain
		featLvl uint32
	)
	r, _, _ := _D3D11CreateDeviceAndSwapChain.Call(
		0,                                 // pAdapter
		uintptr(driverType),               // driverType
		0,                                 // Software
		uintptr(flags),                    // Flags
		0,                                 // pFeatureLevels
		0,                                 // FeatureLevels
		SDK_VERSION,                       // SDKVersion
		uintptr(unsafe.Pointer(swapDesc)), // pSwapChainDesc
		uintptr(unsafe.Pointer(&swchain)), // ppSwapChain
		uintptr(unsafe.Pointer(&dev)),     // ppDevice
		uintptr(unsafe.Pointer(&featLvl)), // pFeatureLevel
		uintptr(unsafe.Pointer(&ctx)),     // ppImmediateContext
	)
	if r != 0 {
		return nil, nil, nil, 0, ErrorCode{Name: "D3D11CreateDeviceAndSwapChain", Code: uint32(r)}
	}
	return dev, ctx, swchain, featLvl, nil
}

func DXGIGetDebugInterface1() (*IDXGIDebug, error) {
	var dbg *IDXGIDebug
	r, _, _ := _DXGIGetDebugInterface1.Call(
		0, // Flags
		uintptr(unsafe.Pointer(&IID_IDXGIDebug)),
		uintptr(unsafe.Pointer(&dbg)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DXGIGetDebugInterface1", Code: uint32(r)}
	}
	return dbg, nil
}

func ReportLiveObjects() error {
	dxgi, err := DXGIGetDebugInterface1()
	if err != nil {
		return err
	}
	defer IUnknownRelease(unsafe.Pointer(dxgi), dxgi.Vtbl.Release)
	dxgi.ReportLiveObjects(&DXGI_DEBUG_ALL, DXGI_DEBUG_RLO_DETAIL|DXGI_DEBUG_RLO_IGNORE_INTERNAL)
	return nil
}

func (d *IDXGIDebug) ReportLiveObjects(guid *GUID, flags uint32) {
	syscall.Syscall6(
		d.Vtbl.ReportLiveObjects,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(guid)),
		uintptr(flags),
		0,
		0,
		0,
	)
}

func (d *Device) CheckFormatSupport(format uint32) (uint32, error) {
	var support uint32
	r, _, _ := syscall.Syscall(
		d.Vtbl.CheckFormatSupport,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(format),
		uintptr(unsafe.Pointer(&support)),
	)
	if r != 0 {
		return 0, ErrorCode{Name: "DeviceCheckFormatSupport", Code: uint32(r)}
	}
	return support, nil
}

func (d *Device) CreateBuffer(desc *BUFFER_DESC, data []byte) (*Buffer, error) {
	var dataDesc *SUBRESOURCE_DATA
	if len(data) > 0 {
		dataDesc = &SUBRESOURCE_DATA{
			pSysMem: &data[0],
		}
	}
	var buf *Buffer
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateBuffer,
		4,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(dataDesc)),
		uintptr(unsafe.Pointer(&buf)),
		0, 0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateBuffer", Code: uint32(r)}
	}
	return buf, nil
}

func (d *Device) CreateDepthStencilViewTEX2D(res *Resource, desc *DEPTH_STENCIL_VIEW_DESC_TEX2D) (*DepthStencilView, error) {
	var view *DepthStencilView
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateDepthStencilView,
		4,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(res)),
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&view)),
		0, 0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateDepthStencilView", Code: uint32(r)}
	}
	return view, nil
}

func (d *Device) CreatePixelShader(bytecode []byte) (*PixelShader, error) {
	var shader *PixelShader
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreatePixelShader,
		5,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(&bytecode[0])),
		uintptr(len(bytecode)),
		0, // pClassLinkage
		uintptr(unsafe.Pointer(&shader)),
		0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreatePixelShader", Code: uint32(r)}
	}
	return shader, nil
}

func (d *Device) CreateVertexShader(bytecode []byte) (*VertexShader, error) {
	var shader *VertexShader
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateVertexShader,
		5,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(&bytecode[0])),
		uintptr(len(bytecode)),
		0, // pClassLinkage
		uintptr(unsafe.Pointer(&shader)),
		0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateVertexShader", Code: uint32(r)}
	}
	return shader, nil
}

func (d *Device) CreateComputeShader(bytecode []byte) (*ComputeShader, error) {
	var shader *ComputeShader
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateComputeShader,
		5,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(&bytecode[0])),
		uintptr(len(bytecode)),
		0, // pClassLinkage
		uintptr(unsafe.Pointer(&shader)),
		0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateComputeShader", Code: uint32(r)}
	}
	return shader, nil
}

func (d *Device) CreateShaderResourceView(res *Resource, desc unsafe.Pointer) (*ShaderResourceView, error) {
	var resView *ShaderResourceView
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateShaderResourceView,
		4,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(res)),
		uintptr(desc),
		uintptr(unsafe.Pointer(&resView)),
		0, 0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateShaderResourceView", Code: uint32(r)}
	}
	return resView, nil
}

func (d *Device) CreateUnorderedAccessView(res *Resource, desc unsafe.Pointer) (*UnorderedAccessView, error) {
	var uaView *UnorderedAccessView
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateUnorderedAccessView,
		4,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(res)),
		uintptr(desc),
		uintptr(unsafe.Pointer(&uaView)),
		0, 0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateUnorderedAccessView", Code: uint32(r)}
	}
	return uaView, nil
}

func (d *Device) CreateRasterizerState(desc *RASTERIZER_DESC) (*RasterizerState, error) {
	var state *RasterizerState
	r, _, _ := syscall.Syscall(
		d.Vtbl.CreateRasterizerState,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&state)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateRasterizerState", Code: uint32(r)}
	}
	return state, nil
}

func (d *Device) CreateInputLayout(descs []INPUT_ELEMENT_DESC, bytecode []byte) (*InputLayout, error) {
	var pdesc *INPUT_ELEMENT_DESC
	if len(descs) > 0 {
		pdesc = &descs[0]
	}
	var layout *InputLayout
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateInputLayout,
		6,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(pdesc)),
		uintptr(len(descs)),
		uintptr(unsafe.Pointer(&bytecode[0])),
		uintptr(len(bytecode)),
		uintptr(unsafe.Pointer(&layout)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateInputLayout", Code: uint32(r)}
	}
	return layout, nil
}

func (d *Device) CreateSamplerState(desc *SAMPLER_DESC) (*SamplerState, error) {
	var sampler *SamplerState
	r, _, _ := syscall.Syscall(
		d.Vtbl.CreateSamplerState,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&sampler)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateSamplerState", Code: uint32(r)}
	}
	return sampler, nil
}

func (d *Device) CreateTexture2D(desc *TEXTURE2D_DESC) (*Texture2D, error) {
	var tex *Texture2D
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateTexture2D,
		4,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(desc)),
		0, // pInitialData
		uintptr(unsafe.Pointer(&tex)),
		0, 0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "CreateTexture2D", Code: uint32(r)}
	}
	return tex, nil
}

func (d *Device) CreateRenderTargetView(res *Resource) (*RenderTargetView, error) {
	var target *RenderTargetView
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateRenderTargetView,
		4,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(res)),
		0, // pDesc
		uintptr(unsafe.Pointer(&target)),
		0, 0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateRenderTargetView", Code: uint32(r)}
	}
	return target, nil
}

func (d *Device) CreateBlendState(desc *BLEND_DESC) (*BlendState, error) {
	var state *BlendState
	r, _, _ := syscall.Syscall(
		d.Vtbl.CreateBlendState,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&state)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateBlendState", Code: uint32(r)}
	}
	return state, nil
}

func (d *Device) CreateDepthStencilState(desc *DEPTH_STENCIL_DESC) (*DepthStencilState, error) {
	var state *DepthStencilState
	r, _, _ := syscall.Syscall(
		d.Vtbl.CreateDepthStencilState,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&state)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "DeviceCreateDepthStencilState", Code: uint32(r)}
	}
	return state, nil
}

func (d *Device) GetFeatureLevel() int {
	lvl, _, _ := syscall.Syscall(
		d.Vtbl.GetFeatureLevel,
		1,
		uintptr(unsafe.Pointer(d)),
		0, 0,
	)
	return int(lvl)
}

func (d *Device) GetImmediateContext() *DeviceContext {
	var ctx *DeviceContext
	syscall.Syscall(
		d.Vtbl.GetImmediateContext,
		2,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(&ctx)),
		0,
	)
	return ctx
}

func (d *Device) ReportLiveDeviceObjects() error {
	intf, err := IUnknownQueryInterface(unsafe.Pointer(d), d.Vtbl.QueryInterface, &IID_ID3D11Debug)
	if err != nil {
		return fmt.Errorf("ReportLiveObjects: failed to query ID3D11Debug interface: %v", err)
	}
	defer IUnknownRelease(unsafe.Pointer(intf), intf.Vtbl.Release)
	dbg := (*Debug)(unsafe.Pointer(intf))
	dbg.ReportLiveDeviceObjects(RLDO_DETAIL | RLDO_IGNORE_INTERNAL)
	return nil
}

func (d *Debug) ReportLiveDeviceObjects(flags uint32) {
	syscall.Syscall(
		d.Vtbl.ReportLiveDeviceObjects,
		2,
		uintptr(unsafe.Pointer(d)),
		uintptr(flags),
		0,
	)
}

func (s *IDXGISwapChain) GetDesc() (DXGI_SWAP_CHAIN_DESC, error) {
	var desc DXGI_SWAP_CHAIN_DESC
	r, _, _ := syscall.Syscall(
		s.Vtbl.GetDesc,
		2,
		uintptr(unsafe.Pointer(s)),
		uintptr(unsafe.Pointer(&desc)),
		0,
	)
	if r != 0 {
		return DXGI_SWAP_CHAIN_DESC{}, ErrorCode{Name: "IDXGISwapChainGetDesc", Code: uint32(r)}
	}
	return desc, nil
}

func (s *IDXGISwapChain) ResizeBuffers(buffers, width, height, newFormat, flags uint32) error {
	r, _, _ := syscall.Syscall6(
		s.Vtbl.ResizeBuffers,
		6,
		uintptr(unsafe.Pointer(s)),
		uintptr(buffers),
		uintptr(width),
		uintptr(height),
		uintptr(newFormat),
		uintptr(flags),
	)
	if r != 0 {
		return ErrorCode{Name: "IDXGISwapChainResizeBuffers", Code: uint32(r)}
	}
	return nil
}

func (s *IDXGISwapChain) Present(SyncInterval int, Flags uint32) error {
	r, _, _ := syscall.Syscall(
		s.Vtbl.Present,
		3,
		uintptr(unsafe.Pointer(s)),
		uintptr(SyncInterval),
		uintptr(Flags),
	)
	if r != 0 {
		return ErrorCode{Name: "IDXGISwapChainPresent", Code: uint32(r)}
	}
	return nil
}

func (s *IDXGISwapChain) GetBuffer(index int, riid *GUID) (*IUnknown, error) {
	var buf *IUnknown
	r, _, _ := syscall.Syscall6(
		s.Vtbl.GetBuffer,
		4,
		uintptr(unsafe.Pointer(s)),
		uintptr(index),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(&buf)),
		0,
		0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "IDXGISwapChainGetBuffer", Code: uint32(r)}
	}
	return buf, nil
}

func (c *DeviceContext) Unmap(resource *Resource, subResource uint32) {
	syscall.Syscall(
		c.Vtbl.Unmap,
		3,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(resource)),
		uintptr(subResource),
	)
}

func (c *DeviceContext) Map(resource *Resource, subResource, mapType, mapFlags uint32) (MAPPED_SUBRESOURCE, error) {
	var resMap MAPPED_SUBRESOURCE
	r, _, _ := syscall.Syscall6(
		c.Vtbl.Map,
		6,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(resource)),
		uintptr(subResource),
		uintptr(mapType),
		uintptr(mapFlags),
		uintptr(unsafe.Pointer(&resMap)),
	)
	if r != 0 {
		return resMap, ErrorCode{Name: "DeviceContextMap", Code: uint32(r)}
	}
	return resMap, nil
}

func (c *DeviceContext) CopySubresourceRegion(dst *Resource, dstSubresource, dstX, dstY, dstZ uint32, src *Resource, srcSubresource uint32, srcBox *BOX) {
	syscall.Syscall9(
		c.Vtbl.CopySubresourceRegion,
		9,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(dstSubresource),
		uintptr(dstX),
		uintptr(dstY),
		uintptr(dstZ),
		uintptr(unsafe.Pointer(src)),
		uintptr(srcSubresource),
		uintptr(unsafe.Pointer(srcBox)),
	)
}

func (c *DeviceContext) ClearDepthStencilView(target *DepthStencilView, flags uint32, depth float32, stencil uint8) {
	syscall.Syscall6(
		c.Vtbl.ClearDepthStencilView,
		5,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(target)),
		uintptr(flags),
		uintptr(math.Float32bits(depth)),
		uintptr(stencil),
		0,
	)
}

func (c *DeviceContext) ClearRenderTargetView(target *RenderTargetView, color *[4]float32) {
	syscall.Syscall(
		c.Vtbl.ClearRenderTargetView,
		3,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(target)),
		uintptr(unsafe.Pointer(color)),
	)
}

func (c *DeviceContext) CSSetShaderResources(startSlot uint32, s *ShaderResourceView) {
	syscall.Syscall6(
		c.Vtbl.CSSetShaderResources,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(startSlot),
		1, // NumViews
		uintptr(unsafe.Pointer(&s)),
		0, 0,
	)
}

func (c *DeviceContext) CSSetUnorderedAccessViews(startSlot uint32, v *UnorderedAccessView) {
	syscall.Syscall6(
		c.Vtbl.CSSetUnorderedAccessViews,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(startSlot),
		1, // NumViews
		uintptr(unsafe.Pointer(&v)),
		0, 0,
	)
}

func (c *DeviceContext) CSSetShader(s *ComputeShader) {
	syscall.Syscall6(
		c.Vtbl.CSSetShader,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(s)),
		0, // ppClassInstances
		0, // NumClassInstances
		0, 0,
	)
}

func (c *DeviceContext) RSSetViewports(viewport *VIEWPORT) {
	syscall.Syscall(
		c.Vtbl.RSSetViewports,
		3,
		uintptr(unsafe.Pointer(c)),
		1, // NumViewports
		uintptr(unsafe.Pointer(viewport)),
	)
}

func (c *DeviceContext) VSSetShader(s *VertexShader) {
	syscall.Syscall6(
		c.Vtbl.VSSetShader,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(s)),
		0, // ppClassInstances
		0, // NumClassInstances
		0, 0,
	)
}

func (c *DeviceContext) VSSetConstantBuffers(b *Buffer) {
	syscall.Syscall6(
		c.Vtbl.VSSetConstantBuffers,
		4,
		uintptr(unsafe.Pointer(c)),
		0, // StartSlot
		1, // NumBuffers
		uintptr(unsafe.Pointer(&b)),
		0, 0,
	)
}

func (c *DeviceContext) PSSetConstantBuffers(b *Buffer) {
	syscall.Syscall6(
		c.Vtbl.PSSetConstantBuffers,
		4,
		uintptr(unsafe.Pointer(c)),
		0, // StartSlot
		1, // NumBuffers
		uintptr(unsafe.Pointer(&b)),
		0, 0,
	)
}

func (c *DeviceContext) PSSetShaderResources(startSlot uint32, s *ShaderResourceView) {
	syscall.Syscall6(
		c.Vtbl.PSSetShaderResources,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(startSlot),
		1, // NumViews
		uintptr(unsafe.Pointer(&s)),
		0, 0,
	)
}

func (c *DeviceContext) PSSetSamplers(startSlot uint32, s *SamplerState) {
	syscall.Syscall6(
		c.Vtbl.PSSetSamplers,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(startSlot),
		1, // NumSamplers
		uintptr(unsafe.Pointer(&s)),
		0, 0,
	)
}

func (c *DeviceContext) PSSetShader(s *PixelShader) {
	syscall.Syscall6(
		c.Vtbl.PSSetShader,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(s)),
		0, // ppClassInstances
		0, // NumClassInstances
		0, 0,
	)
}

func (c *DeviceContext) UpdateSubresource(res *Resource, dstBox *BOX, rowPitch, depthPitch uint32, data []byte) {
	syscall.Syscall9(
		c.Vtbl.UpdateSubresource,
		7,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(res)),
		0, // DstSubresource
		uintptr(unsafe.Pointer(dstBox)),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(rowPitch),
		uintptr(depthPitch),
		0, 0,
	)
}

func (c *DeviceContext) RSSetState(state *RasterizerState) {
	syscall.Syscall(
		c.Vtbl.RSSetState,
		2,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(state)),
		0,
	)
}

func (c *DeviceContext) IASetInputLayout(layout *InputLayout) {
	syscall.Syscall(
		c.Vtbl.IASetInputLayout,
		2,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(layout)),
		0,
	)
}

func (c *DeviceContext) IASetIndexBuffer(buf *Buffer, format, offset uint32) {
	syscall.Syscall6(
		c.Vtbl.IASetIndexBuffer,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(buf)),
		uintptr(format),
		uintptr(offset),
		0, 0,
	)
}

func (c *DeviceContext) IASetVertexBuffers(buf *Buffer, stride, offset uint32) {
	syscall.Syscall6(
		c.Vtbl.IASetVertexBuffers,
		6,
		uintptr(unsafe.Pointer(c)),
		0, // StartSlot
		1, // NumBuffers,
		uintptr(unsafe.Pointer(&buf)),
		uintptr(unsafe.Pointer(&stride)),
		uintptr(unsafe.Pointer(&offset)),
	)
}

func (c *DeviceContext) IASetPrimitiveTopology(mode uint32) {
	syscall.Syscall(
		c.Vtbl.IASetPrimitiveTopology,
		2,
		uintptr(unsafe.Pointer(c)),
		uintptr(mode),
		0,
	)
}

func (c *DeviceContext) OMGetRenderTargets() (*RenderTargetView, *DepthStencilView) {
	var (
		target           *RenderTargetView
		depthStencilView *DepthStencilView
	)
	syscall.Syscall6(
		c.Vtbl.OMGetRenderTargets,
		4,
		uintptr(unsafe.Pointer(c)),
		1, // NumViews
		uintptr(unsafe.Pointer(&target)),
		uintptr(unsafe.Pointer(&depthStencilView)),
		0, 0,
	)
	return target, depthStencilView
}

func (c *DeviceContext) OMSetRenderTargets(target *RenderTargetView, depthStencil *DepthStencilView) {
	syscall.Syscall6(
		c.Vtbl.OMSetRenderTargets,
		4,
		uintptr(unsafe.Pointer(c)),
		1, // NumViews
		uintptr(unsafe.Pointer(&target)),
		uintptr(unsafe.Pointer(depthStencil)),
		0, 0,
	)
}

func (c *DeviceContext) Draw(count, start uint32) {
	syscall.Syscall(
		c.Vtbl.Draw,
		3,
		uintptr(unsafe.Pointer(c)),
		uintptr(count),
		uintptr(start),
	)
}

func (c *DeviceContext) DrawIndexed(count, start uint32, base int32) {
	syscall.Syscall6(
		c.Vtbl.DrawIndexed,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(count),
		uintptr(start),
		uintptr(base),
		0, 0,
	)
}

func (c *DeviceContext) Dispatch(x, y, z uint32) {
	syscall.Syscall6(
		c.Vtbl.Dispatch,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(x),
		uintptr(y),
		uintptr(z),
		0, 0,
	)
}

func (c *DeviceContext) OMSetBlendState(state *BlendState, factor *f32color.RGBA, sampleMask uint32) {
	syscall.Syscall6(
		c.Vtbl.OMSetBlendState,
		4,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(state)),
		uintptr(unsafe.Pointer(factor)),
		uintptr(sampleMask),
		0, 0,
	)
}

func (c *DeviceContext) OMSetDepthStencilState(state *DepthStencilState, stencilRef uint32) {
	syscall.Syscall(
		c.Vtbl.OMSetDepthStencilState,
		3,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(state)),
		uintptr(stencilRef),
	)
}

func (d *IDXGIObject) GetParent(guid *GUID) (*IDXGIObject, error) {
	var parent *IDXGIObject
	r, _, _ := syscall.Syscall(
		d.Vtbl.GetParent,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(guid)),
		uintptr(unsafe.Pointer(&parent)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "IDXGIObjectGetParent", Code: uint32(r)}
	}
	return parent, nil
}

func (d *IDXGIFactory) CreateSwapChain(device *IUnknown, desc *DXGI_SWAP_CHAIN_DESC) (*IDXGISwapChain, error) {
	var swchain *IDXGISwapChain
	r, _, _ := syscall.Syscall6(
		d.Vtbl.CreateSwapChain,
		4,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(desc)),
		uintptr(unsafe.Pointer(&swchain)),
		0, 0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "IDXGIFactory", Code: uint32(r)}
	}
	return swchain, nil
}

func (d *IDXGIDevice) GetAdapter() (*IDXGIAdapter, error) {
	var adapter *IDXGIAdapter
	r, _, _ := syscall.Syscall(
		d.Vtbl.GetAdapter,
		2,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(&adapter)),
		0,
	)
	if r != 0 {
		return nil, ErrorCode{Name: "IDXGIDeviceGetAdapter", Code: uint32(r)}
	}
	return adapter, nil
}

func IUnknownQueryInterface(obj unsafe.Pointer, queryInterfaceMethod uintptr, guid *GUID) (*IUnknown, error) {
	var ref *IUnknown
	r, _, _ := syscall.Syscall(
		queryInterfaceMethod,
		3,
		uintptr(obj),
		uintptr(unsafe.Pointer(guid)),
		uintptr(unsafe.Pointer(&ref)),
	)
	if r != 0 {
		return nil, ErrorCode{Name: "IUnknownQueryInterface", Code: uint32(r)}
	}
	return ref, nil
}

func IUnknownAddRef(obj unsafe.Pointer, addRefMethod uintptr) {
	syscall.Syscall(
		addRefMethod,
		1,
		uintptr(obj),
		0,
		0,
	)
}

func IUnknownRelease(obj unsafe.Pointer, releaseMethod uintptr) {
	syscall.Syscall(
		releaseMethod,
		1,
		uintptr(obj),
		0,
		0,
	)
}

func (e ErrorCode) Error() string {
	return fmt.Sprintf("%s: %#x", e.Name, e.Code)
}

func CreateSwapChain(dev *Device, hwnd windows.Handle) (*IDXGISwapChain, error) {
	dxgiDev, err := IUnknownQueryInterface(unsafe.Pointer(dev), dev.Vtbl.QueryInterface, &IID_IDXGIDevice)
	if err != nil {
		return nil, fmt.Errorf("NewContext: %v", err)
	}
	adapter, err := (*IDXGIDevice)(unsafe.Pointer(dxgiDev)).GetAdapter()
	IUnknownRelease(unsafe.Pointer(dxgiDev), dxgiDev.Vtbl.Release)
	if err != nil {
		return nil, fmt.Errorf("NewContext: %v", err)
	}
	dxgiFactory, err := (*IDXGIObject)(unsafe.Pointer(adapter)).GetParent(&IID_IDXGIFactory)
	IUnknownRelease(unsafe.Pointer(adapter), adapter.Vtbl.Release)
	if err != nil {
		return nil, fmt.Errorf("NewContext: %v", err)
	}
	swchain, err := (*IDXGIFactory)(unsafe.Pointer(dxgiFactory)).CreateSwapChain(
		(*IUnknown)(unsafe.Pointer(dev)),
		&DXGI_SWAP_CHAIN_DESC{
			BufferDesc: DXGI_MODE_DESC{
				Format: DXGI_FORMAT_R8G8B8A8_UNORM_SRGB,
			},
			SampleDesc: DXGI_SAMPLE_DESC{
				Count: 1,
			},
			BufferUsage:  DXGI_USAGE_RENDER_TARGET_OUTPUT,
			BufferCount:  1,
			OutputWindow: hwnd,
			Windowed:     1,
			SwapEffect:   DXGI_SWAP_EFFECT_DISCARD,
		},
	)
	IUnknownRelease(unsafe.Pointer(dxgiFactory), dxgiFactory.Vtbl.Release)
	if err != nil {
		return nil, fmt.Errorf("NewContext: %v", err)
	}
	return swchain, nil
}

func CreateDepthView(d *Device, width, height, depthBits int) (*DepthStencilView, error) {
	depthTex, err := d.CreateTexture2D(&TEXTURE2D_DESC{
		Width:     uint32(width),
		Height:    uint32(height),
		MipLevels: 1,
		ArraySize: 1,
		Format:    DXGI_FORMAT_D24_UNORM_S8_UINT,
		SampleDesc: DXGI_SAMPLE_DESC{
			Count:   1,
			Quality: 0,
		},
		BindFlags: BIND_DEPTH_STENCIL,
	})
	if err != nil {
		return nil, err
	}
	depthView, err := d.CreateDepthStencilViewTEX2D(
		(*Resource)(unsafe.Pointer(depthTex)),
		&DEPTH_STENCIL_VIEW_DESC_TEX2D{
			Format:        DXGI_FORMAT_D24_UNORM_S8_UINT,
			ViewDimension: DSV_DIMENSION_TEXTURE2D,
		},
	)
	IUnknownRelease(unsafe.Pointer(depthTex), depthTex.Vtbl.Release)
	return depthView, err
}
