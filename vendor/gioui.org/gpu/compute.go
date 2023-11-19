// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/maphash"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/bits"
	"os"
	"runtime"
	"sort"
	"time"
	"unsafe"

	"gioui.org/cpu"
	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/byteslice"
	"gioui.org/internal/f32"
	"gioui.org/internal/f32color"
	"gioui.org/internal/ops"
	"gioui.org/internal/scene"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/shader"
	"gioui.org/shader/gio"
	"gioui.org/shader/piet"
)

type compute struct {
	ctx driver.Device

	collector     collector
	enc           encoder
	texOps        []textureOp
	viewport      image.Point
	maxTextureDim int
	srgb          bool
	atlases       []*textureAtlas
	frameCount    uint
	moves         []atlasMove

	programs struct {
		elements   computeProgram
		tileAlloc  computeProgram
		pathCoarse computeProgram
		backdrop   computeProgram
		binning    computeProgram
		coarse     computeProgram
		kernel4    computeProgram
	}
	buffers struct {
		config sizedBuffer
		scene  sizedBuffer
		state  sizedBuffer
		memory sizedBuffer
	}
	output struct {
		blitPipeline driver.Pipeline

		buffer sizedBuffer

		uniforms *copyUniforms
		uniBuf   driver.Buffer

		layerVertices []layerVertex
		descriptors   *piet.Kernel4DescriptorSetLayout

		nullMaterials driver.Texture
	}
	// imgAllocs maps imageOpData.handles to allocs.
	imgAllocs map[interface{}]*atlasAlloc
	// materials contains the pre-processed materials (transformed images for
	// now, gradients etc. later) packed in a texture atlas. The atlas is used
	// as source in kernel4.
	materials struct {
		// allocs maps texture ops the their atlases and FillImage offsets.
		allocs map[textureKey]materialAlloc

		pipeline driver.Pipeline
		buffer   sizedBuffer
		quads    []materialVertex
		uniforms struct {
			u   *materialUniforms
			buf driver.Buffer
		}
	}
	timers struct {
		profile string
		t       *timers
		compact *timer
		render  *timer
		blit    *timer
	}

	// CPU fallback fields.
	useCPU     bool
	dispatcher *dispatcher

	// The following fields hold scratch space to avoid garbage.
	zeroSlice []byte
	memHeader *memoryHeader
	conf      *config
}

type materialAlloc struct {
	alloc  *atlasAlloc
	offset image.Point
}

type layer struct {
	rect      image.Rectangle
	alloc     *atlasAlloc
	ops       []paintOp
	materials *textureAtlas
}

type allocQuery struct {
	atlas     *textureAtlas
	size      image.Point
	empty     bool
	format    driver.TextureFormat
	bindings  driver.BufferBinding
	nocompact bool
}

type atlasAlloc struct {
	atlas      *textureAtlas
	rect       image.Rectangle
	cpu        bool
	dead       bool
	frameCount uint
}

type atlasMove struct {
	src     *textureAtlas
	dstPos  image.Point
	srcRect image.Rectangle
	cpu     bool
}

type textureAtlas struct {
	image     driver.Texture
	format    driver.TextureFormat
	bindings  driver.BufferBinding
	hasCPU    bool
	cpuImage  cpu.ImageDescriptor
	size      image.Point
	allocs    []*atlasAlloc
	packer    packer
	realized  bool
	lastFrame uint
	compact   bool
}

type copyUniforms struct {
	scale   [2]float32
	pos     [2]float32
	uvScale [2]float32
	_       [8]byte // Pad to 16 bytes.
}

type materialUniforms struct {
	scale       [2]float32
	pos         [2]float32
	emulatesRGB float32
	_           [12]byte // Pad to 16 bytes
}

type collector struct {
	hasher     maphash.Hash
	profile    bool
	reader     ops.Reader
	states     []f32.Affine2D
	clear      bool
	clearColor f32color.RGBA
	clipStates []clipState
	order      []hashIndex
	transStack []transEntry
	prevFrame  opsCollector
	frame      opsCollector
}

type transEntry struct {
	t        f32.Affine2D
	relTrans f32.Affine2D
}

type hashIndex struct {
	index int
	hash  uint64
}

type opsCollector struct {
	paths    []byte
	clipCmds []clipCmd
	ops      []paintOp
	layers   []layer
}

type paintOp struct {
	clipStack []clipCmd
	offset    image.Point
	state     paintKey
	intersect f32.Rectangle
	hash      uint64
	layer     int
	texOpIdx  int
}

// clipCmd describes a clipping command ready to be used for the compute
// pipeline.
type clipCmd struct {
	// union of the bounds of the operations that are clipped.
	union     f32.Rectangle
	state     clipKey
	path      []byte
	pathKey   ops.Key
	absBounds f32.Rectangle
}

type encoderState struct {
	relTrans f32.Affine2D
	clip     *clipState

	paintKey
}

// clipKey completely describes a clip operation (along with its path) and is appropriate
// for hashing and equality checks.
type clipKey struct {
	bounds      f32.Rectangle
	strokeWidth float32
	relTrans    f32.Affine2D
	pathHash    uint64
}

// paintKey completely defines a paint operation. It is suitable for hashing and
// equality checks.
type paintKey struct {
	t       f32.Affine2D
	matType materialType
	// Current paint.ImageOp
	image imageOpData
	// Current paint.ColorOp, if any.
	color color.NRGBA

	// Current paint.LinearGradientOp.
	stop1  f32.Point
	stop2  f32.Point
	color1 color.NRGBA
	color2 color.NRGBA
}

type clipState struct {
	absBounds f32.Rectangle
	parent    *clipState
	path      []byte
	pathKey   ops.Key
	intersect f32.Rectangle

	clipKey
}

type layerVertex struct {
	posX, posY float32
	u, v       float32
}

// materialVertex describes a vertex of a quad used to render a transformed
// material.
type materialVertex struct {
	posX, posY float32
	u, v       float32
}

// textureKey identifies textureOp.
type textureKey struct {
	handle    interface{}
	transform f32.Affine2D
	bounds    image.Rectangle
}

// textureOp represents an paintOp that requires texture space.
type textureOp struct {
	img imageOpData
	key textureKey
	// offset is the integer offset separated from key.transform to increase cache hit rate.
	off image.Point
	// matAlloc is the atlas placement for material.
	matAlloc materialAlloc
	// imgAlloc is the atlas placement for the source image
	imgAlloc *atlasAlloc
}

type encoder struct {
	scene    []scene.Command
	npath    int
	npathseg int
	ntrans   int
}

// sizedBuffer holds a GPU buffer, or its equivalent CPU memory.
type sizedBuffer struct {
	size   int
	buffer driver.Buffer
	// cpuBuf is initialized when useCPU is true.
	cpuBuf cpu.BufferDescriptor
}

// computeProgram holds a compute program, or its equivalent CPU implementation.
type computeProgram struct {
	prog driver.Program

	// CPU fields.
	progInfo    *cpu.ProgramInfo
	descriptors unsafe.Pointer
	buffers     []*cpu.BufferDescriptor
}

// config matches Config in setup.h
type config struct {
	n_elements      uint32 // paths
	n_pathseg       uint32
	width_in_tiles  uint32
	height_in_tiles uint32
	tile_alloc      memAlloc
	bin_alloc       memAlloc
	ptcl_alloc      memAlloc
	pathseg_alloc   memAlloc
	anno_alloc      memAlloc
	trans_alloc     memAlloc
}

// memAlloc matches Alloc in mem.h
type memAlloc struct {
	offset uint32
	//size   uint32
}

// memoryHeader matches the header of Memory in mem.h.
type memoryHeader struct {
	mem_offset uint32
	mem_error  uint32
}

// rect is a oriented rectangle.
type rectangle [4]f32.Point

const (
	layersBindings    = driver.BufferBindingShaderStorageWrite | driver.BufferBindingTexture
	materialsBindings = driver.BufferBindingFramebuffer | driver.BufferBindingShaderStorageRead
	// Materials and layers can share texture storage if their bindings match.
	combinedBindings = layersBindings | materialsBindings
)

// GPU structure sizes and constants.
const (
	tileWidthPx       = 32
	tileHeightPx      = 32
	ptclInitialAlloc  = 1024
	kernel4OutputUnit = 2
	kernel4AtlasUnit  = 3

	pathSize    = 12
	binSize     = 8
	pathsegSize = 52
	annoSize    = 32
	transSize   = 24
	stateSize   = 60
	stateStride = 4 + 2*stateSize
)

// mem.h constants.
const (
	memNoError      = 0 // NO_ERROR
	memMallocFailed = 1 // ERR_MALLOC_FAILED
)

func newCompute(ctx driver.Device) (*compute, error) {
	caps := ctx.Caps()
	maxDim := caps.MaxTextureSize
	// Large atlas textures cause artifacts due to precision loss in
	// shaders.
	if cap := 8192; maxDim > cap {
		maxDim = cap
	}
	// The compute programs can only span 128x64 tiles. Limit to 64 for now, and leave the
	// complexity of a rectangular limit for later.
	if computeCap := 4096; maxDim > computeCap {
		maxDim = computeCap
	}
	g := &compute{
		ctx:           ctx,
		maxTextureDim: maxDim,
		srgb:          caps.Features.Has(driver.FeatureSRGB),
		conf:          new(config),
		memHeader:     new(memoryHeader),
	}
	shaders := []struct {
		prog *computeProgram
		src  shader.Sources
		info *cpu.ProgramInfo
	}{
		{&g.programs.elements, piet.Shader_elements_comp, piet.ElementsProgramInfo},
		{&g.programs.tileAlloc, piet.Shader_tile_alloc_comp, piet.Tile_allocProgramInfo},
		{&g.programs.pathCoarse, piet.Shader_path_coarse_comp, piet.Path_coarseProgramInfo},
		{&g.programs.backdrop, piet.Shader_backdrop_comp, piet.BackdropProgramInfo},
		{&g.programs.binning, piet.Shader_binning_comp, piet.BinningProgramInfo},
		{&g.programs.coarse, piet.Shader_coarse_comp, piet.CoarseProgramInfo},
		{&g.programs.kernel4, piet.Shader_kernel4_comp, piet.Kernel4ProgramInfo},
	}
	if !caps.Features.Has(driver.FeatureCompute) {
		if !cpu.Supported {
			return nil, errors.New("gpu: missing support for compute programs")
		}
		g.useCPU = true
	}
	if g.useCPU {
		g.dispatcher = newDispatcher(runtime.NumCPU())
	} else {
		null, err := ctx.NewTexture(driver.TextureFormatRGBA8, 1, 1, driver.FilterNearest, driver.FilterNearest, driver.BufferBindingShaderStorageRead)
		if err != nil {
			g.Release()
			return nil, err
		}
		g.output.nullMaterials = null
	}

	copyVert, copyFrag, err := newShaders(ctx, gio.Shader_copy_vert, gio.Shader_copy_frag)
	if err != nil {
		g.Release()
		return nil, err
	}
	defer copyVert.Release()
	defer copyFrag.Release()
	pipe, err := ctx.NewPipeline(driver.PipelineDesc{
		VertexShader:   copyVert,
		FragmentShader: copyFrag,
		VertexLayout: driver.VertexLayout{
			Inputs: []driver.InputDesc{
				{Type: shader.DataTypeFloat, Size: 2, Offset: 0},
				{Type: shader.DataTypeFloat, Size: 2, Offset: 4 * 2},
			},
			Stride: int(unsafe.Sizeof(g.output.layerVertices[0])),
		},
		PixelFormat: driver.TextureFormatOutput,
		BlendDesc: driver.BlendDesc{
			Enable:    true,
			SrcFactor: driver.BlendFactorOne,
			DstFactor: driver.BlendFactorOneMinusSrcAlpha,
		},
		Topology: driver.TopologyTriangles,
	})
	if err != nil {
		g.Release()
		return nil, err
	}
	g.output.blitPipeline = pipe
	g.output.uniforms = new(copyUniforms)

	buf, err := ctx.NewBuffer(driver.BufferBindingUniforms, int(unsafe.Sizeof(*g.output.uniforms)))
	if err != nil {
		g.Release()
		return nil, err
	}
	g.output.uniBuf = buf

	materialVert, materialFrag, err := newShaders(ctx, gio.Shader_material_vert, gio.Shader_material_frag)
	if err != nil {
		g.Release()
		return nil, err
	}
	defer materialVert.Release()
	defer materialFrag.Release()
	pipe, err = ctx.NewPipeline(driver.PipelineDesc{
		VertexShader:   materialVert,
		FragmentShader: materialFrag,
		VertexLayout: driver.VertexLayout{
			Inputs: []driver.InputDesc{
				{Type: shader.DataTypeFloat, Size: 2, Offset: 0},
				{Type: shader.DataTypeFloat, Size: 2, Offset: 4 * 2},
			},
			Stride: int(unsafe.Sizeof(g.materials.quads[0])),
		},
		PixelFormat: driver.TextureFormatRGBA8,
		Topology:    driver.TopologyTriangles,
	})
	if err != nil {
		g.Release()
		return nil, err
	}
	g.materials.pipeline = pipe
	g.materials.uniforms.u = new(materialUniforms)

	buf, err = ctx.NewBuffer(driver.BufferBindingUniforms, int(unsafe.Sizeof(*g.materials.uniforms.u)))
	if err != nil {
		g.Release()
		return nil, err
	}
	g.materials.uniforms.buf = buf

	for _, shader := range shaders {
		if !g.useCPU {
			p, err := ctx.NewComputeProgram(shader.src)
			if err != nil {
				g.Release()
				return nil, err
			}
			shader.prog.prog = p
		} else {
			shader.prog.progInfo = shader.info
		}
	}
	if g.useCPU {
		{
			desc := new(piet.ElementsDescriptorSetLayout)
			g.programs.elements.descriptors = unsafe.Pointer(desc)
			g.programs.elements.buffers = []*cpu.BufferDescriptor{desc.Binding0(), desc.Binding1(), desc.Binding2(), desc.Binding3()}
		}
		{
			desc := new(piet.Tile_allocDescriptorSetLayout)
			g.programs.tileAlloc.descriptors = unsafe.Pointer(desc)
			g.programs.tileAlloc.buffers = []*cpu.BufferDescriptor{desc.Binding0(), desc.Binding1()}
		}
		{
			desc := new(piet.Path_coarseDescriptorSetLayout)
			g.programs.pathCoarse.descriptors = unsafe.Pointer(desc)
			g.programs.pathCoarse.buffers = []*cpu.BufferDescriptor{desc.Binding0(), desc.Binding1()}
		}
		{
			desc := new(piet.BackdropDescriptorSetLayout)
			g.programs.backdrop.descriptors = unsafe.Pointer(desc)
			g.programs.backdrop.buffers = []*cpu.BufferDescriptor{desc.Binding0(), desc.Binding1()}
		}
		{
			desc := new(piet.BinningDescriptorSetLayout)
			g.programs.binning.descriptors = unsafe.Pointer(desc)
			g.programs.binning.buffers = []*cpu.BufferDescriptor{desc.Binding0(), desc.Binding1()}
		}
		{
			desc := new(piet.CoarseDescriptorSetLayout)
			g.programs.coarse.descriptors = unsafe.Pointer(desc)
			g.programs.coarse.buffers = []*cpu.BufferDescriptor{desc.Binding0(), desc.Binding1()}
		}
		{
			desc := new(piet.Kernel4DescriptorSetLayout)
			g.programs.kernel4.descriptors = unsafe.Pointer(desc)
			g.programs.kernel4.buffers = []*cpu.BufferDescriptor{desc.Binding0(), desc.Binding1()}
			g.output.descriptors = desc
		}
	}
	return g, nil
}

func newShaders(ctx driver.Device, vsrc, fsrc shader.Sources) (vert driver.VertexShader, frag driver.FragmentShader, err error) {
	vert, err = ctx.NewVertexShader(vsrc)
	if err != nil {
		return
	}
	frag, err = ctx.NewFragmentShader(fsrc)
	if err != nil {
		vert.Release()
	}
	return
}

func (g *compute) Frame(frameOps *op.Ops, target RenderTarget, viewport image.Point) error {
	g.frameCount++
	g.collect(viewport, frameOps)
	return g.frame(target)
}

func (g *compute) collect(viewport image.Point, ops *op.Ops) {
	g.viewport = viewport
	g.collector.reset()

	g.texOps = g.texOps[:0]
	g.collector.collect(ops, viewport, &g.texOps)
}

func (g *compute) Clear(col color.NRGBA) {
	g.collector.clear = true
	g.collector.clearColor = f32color.LinearFromSRGB(col)
}

func (g *compute) frame(target RenderTarget) error {
	viewport := g.viewport
	defFBO := g.ctx.BeginFrame(target, g.collector.clear, viewport)
	defer g.ctx.EndFrame()

	t := &g.timers
	if g.collector.profile && t.t == nil && g.ctx.Caps().Features.Has(driver.FeatureTimers) {
		t.t = newTimers(g.ctx)
		t.compact = t.t.newTimer()
		t.render = t.t.newTimer()
		t.blit = t.t.newTimer()
	}

	if err := g.uploadImages(); err != nil {
		return err
	}
	if err := g.renderMaterials(); err != nil {
		return err
	}
	g.layer(viewport, g.texOps)
	t.render.begin()
	if err := g.renderLayers(viewport); err != nil {
		return err
	}
	t.render.end()
	d := driver.LoadDesc{
		ClearColor: g.collector.clearColor,
	}
	if g.collector.clear {
		g.collector.clear = false
		d.Action = driver.LoadActionClear
	}
	t.blit.begin()
	g.blitLayers(d, defFBO, viewport)
	t.blit.end()
	t.compact.begin()
	if err := g.compactAllocs(); err != nil {
		return err
	}
	t.compact.end()
	if g.collector.profile && t.t.ready() {
		com, ren, blit := t.compact.Elapsed, t.render.Elapsed, t.blit.Elapsed
		ft := com + ren + blit
		q := 100 * time.Microsecond
		ft = ft.Round(q)
		com, ren, blit = com.Round(q), ren.Round(q), blit.Round(q)
		t.profile = fmt.Sprintf("ft:%7s com: %7s ren:%7s blit:%7s", ft, com, ren, blit)
	}
	return nil
}

func (g *compute) dumpAtlases() {
	for i, a := range g.atlases {
		dump := image.NewRGBA(image.Rectangle{Max: a.size})
		err := driver.DownloadImage(g.ctx, a.image, dump)
		if err != nil {
			panic(err)
		}
		nrgba := image.NewNRGBA(dump.Bounds())
		draw.Draw(nrgba, image.Rectangle{}, dump, image.Point{}, draw.Src)
		var buf bytes.Buffer
		if err := png.Encode(&buf, nrgba); err != nil {
			panic(err)
		}
		if err := os.WriteFile(fmt.Sprintf("dump-%d.png", i), buf.Bytes(), 0600); err != nil {
			panic(err)
		}
	}
}

func (g *compute) Profile() string {
	return g.timers.profile
}

func (g *compute) compactAllocs() error {
	const (
		maxAllocAge = 3
		maxAtlasAge = 10
	)
	atlases := g.atlases
	for _, a := range atlases {
		if len(a.allocs) > 0 && g.frameCount-a.lastFrame > maxAtlasAge {
			a.compact = true
		}
	}
	for len(atlases) > 0 {
		var (
			dstAtlas *textureAtlas
			format   driver.TextureFormat
			bindings driver.BufferBinding
		)
		g.moves = g.moves[:0]
		addedLayers := false
		useCPU := false
	fill:
		for len(atlases) > 0 {
			srcAtlas := atlases[0]
			allocs := srcAtlas.allocs
			if !srcAtlas.compact {
				atlases = atlases[1:]
				continue
			}
			if addedLayers && (format != srcAtlas.format || srcAtlas.bindings&bindings != srcAtlas.bindings) {
				break
			}
			format = srcAtlas.format
			bindings = srcAtlas.bindings
			for len(srcAtlas.allocs) > 0 {
				a := srcAtlas.allocs[0]
				n := len(srcAtlas.allocs)
				if g.frameCount-a.frameCount > maxAllocAge {
					a.dead = true
					srcAtlas.allocs[0] = srcAtlas.allocs[n-1]
					srcAtlas.allocs = srcAtlas.allocs[:n-1]
					continue
				}
				size := a.rect.Size()
				alloc, fits := g.atlasAlloc(allocQuery{
					atlas:     dstAtlas,
					size:      size,
					format:    format,
					bindings:  bindings,
					nocompact: true,
				})
				if !fits {
					break fill
				}
				dstAtlas = alloc.atlas
				allocs = append(allocs, a)
				addedLayers = true
				useCPU = useCPU || a.cpu
				dstAtlas.allocs = append(dstAtlas.allocs, a)
				pos := alloc.rect.Min
				g.moves = append(g.moves, atlasMove{
					src: srcAtlas, dstPos: pos, srcRect: a.rect, cpu: a.cpu,
				})
				a.atlas = dstAtlas
				a.rect = image.Rectangle{Min: pos, Max: pos.Add(a.rect.Size())}
				srcAtlas.allocs[0] = srcAtlas.allocs[n-1]
				srcAtlas.allocs = srcAtlas.allocs[:n-1]
			}
			srcAtlas.compact = false
			srcAtlas.realized = false
			srcAtlas.packer.clear()
			srcAtlas.packer.newPage()
			srcAtlas.packer.maxDims = image.Pt(g.maxTextureDim, g.maxTextureDim)
			atlases = atlases[1:]
		}
		if !addedLayers {
			break
		}
		outputSize := dstAtlas.packer.sizes[0]
		if err := g.realizeAtlas(dstAtlas, useCPU, outputSize); err != nil {
			return err
		}
		for _, move := range g.moves {
			if !move.cpu {
				g.ctx.CopyTexture(dstAtlas.image, move.dstPos, move.src.image, move.srcRect)
			} else {
				src := move.src.cpuImage.Data()
				dst := dstAtlas.cpuImage.Data()
				sstride := move.src.size.X * 4
				dstride := dstAtlas.size.X * 4
				copyImage(dst, dstride, move.dstPos, src, sstride, move.srcRect)
			}
		}
	}
	for i := len(g.atlases) - 1; i >= 0; i-- {
		a := g.atlases[i]
		if len(a.allocs) == 0 && g.frameCount-a.lastFrame > maxAtlasAge {
			a.Release()
			n := len(g.atlases)
			g.atlases[i] = g.atlases[n-1]
			g.atlases = g.atlases[:n-1]
		}
	}
	return nil
}

func copyImage(dst []byte, dstStride int, dstPos image.Point, src []byte, srcStride int, srcRect image.Rectangle) {
	sz := srcRect.Size()
	soff := srcRect.Min.Y*srcStride + srcRect.Min.X*4
	doff := dstPos.Y*dstStride + dstPos.X*4
	rowLen := sz.X * 4
	for y := 0; y < sz.Y; y++ {
		srow := src[soff : soff+rowLen]
		drow := dst[doff : doff+rowLen]
		copy(drow, srow)
		soff += srcStride
		doff += dstStride
	}
}

func (g *compute) renderLayers(viewport image.Point) error {
	layers := g.collector.frame.layers
	for len(layers) > 0 {
		var materials, dst *textureAtlas
		addedLayers := false
		g.enc.reset()
		for len(layers) > 0 {
			l := &layers[0]
			if l.alloc != nil {
				layers = layers[1:]
				continue
			}
			if materials != nil {
				if l.materials != nil && materials != l.materials {
					// Only one materials texture per compute pass.
					break
				}
			} else {
				materials = l.materials
			}
			size := l.rect.Size()
			alloc, fits := g.atlasAlloc(allocQuery{
				atlas:    dst,
				empty:    true,
				format:   driver.TextureFormatRGBA8,
				bindings: combinedBindings,
				// Pad to avoid overlap.
				size: size.Add(image.Pt(1, 1)),
			})
			if !fits {
				// Only one output atlas per compute pass.
				break
			}
			dst = alloc.atlas
			dst.compact = true
			addedLayers = true
			l.alloc = &alloc
			dst.allocs = append(dst.allocs, l.alloc)
			encodeLayer(*l, alloc.rect.Min, viewport, &g.enc, g.texOps)
			layers = layers[1:]
		}
		if !addedLayers {
			break
		}
		outputSize := dst.packer.sizes[0]
		tileDims := image.Point{
			X: (outputSize.X + tileWidthPx - 1) / tileWidthPx,
			Y: (outputSize.Y + tileHeightPx - 1) / tileHeightPx,
		}
		w, h := tileDims.X*tileWidthPx, tileDims.Y*tileHeightPx
		if err := g.realizeAtlas(dst, g.useCPU, image.Pt(w, h)); err != nil {
			return err
		}
		if err := g.render(materials, dst.image, dst.cpuImage, tileDims, dst.size.X*4); err != nil {
			return err
		}
	}
	return nil
}

func (g *compute) blitLayers(d driver.LoadDesc, fbo driver.Texture, viewport image.Point) {
	layers := g.collector.frame.layers
	g.output.layerVertices = g.output.layerVertices[:0]
	for _, l := range layers {
		placef := layout.FPt(l.alloc.rect.Min)
		sizef := layout.FPt(l.rect.Size())
		r := f32.FRect(l.rect)
		quad := [4]layerVertex{
			{posX: float32(r.Min.X), posY: float32(r.Min.Y), u: placef.X, v: placef.Y},
			{posX: float32(r.Max.X), posY: float32(r.Min.Y), u: placef.X + sizef.X, v: placef.Y},
			{posX: float32(r.Max.X), posY: float32(r.Max.Y), u: placef.X + sizef.X, v: placef.Y + sizef.Y},
			{posX: float32(r.Min.X), posY: float32(r.Max.Y), u: placef.X, v: placef.Y + sizef.Y},
		}
		g.output.layerVertices = append(g.output.layerVertices, quad[0], quad[1], quad[3], quad[3], quad[2], quad[1])
		g.ctx.PrepareTexture(l.alloc.atlas.image)
	}
	if len(g.output.layerVertices) > 0 {
		vertexData := byteslice.Slice(g.output.layerVertices)
		g.output.buffer.ensureCapacity(false, g.ctx, driver.BufferBindingVertices, len(vertexData))
		g.output.buffer.buffer.Upload(vertexData)
	}
	g.ctx.BeginRenderPass(fbo, d)
	defer g.ctx.EndRenderPass()
	if len(layers) == 0 {
		return
	}
	g.ctx.Viewport(0, 0, viewport.X, viewport.Y)
	g.ctx.BindPipeline(g.output.blitPipeline)
	g.ctx.BindVertexBuffer(g.output.buffer.buffer, 0)
	start := 0
	for len(layers) > 0 {
		count := 0
		atlas := layers[0].alloc.atlas
		for len(layers) > 0 {
			l := layers[0]
			if l.alloc.atlas != atlas {
				break
			}
			layers = layers[1:]
			const verticesPerQuad = 6
			count += verticesPerQuad
		}

		// Transform positions to clip space: [-1, -1] - [1, 1], and texture
		// coordinates to texture space: [0, 0] - [1, 1].
		clip := f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(2/float32(viewport.X), 2/float32(viewport.Y))).Offset(f32.Pt(-1, -1))
		sx, _, ox, _, sy, oy := clip.Elems()
		g.output.uniforms.scale = [2]float32{sx, sy}
		g.output.uniforms.pos = [2]float32{ox, oy}
		g.output.uniforms.uvScale = [2]float32{1 / float32(atlas.size.X), 1 / float32(atlas.size.Y)}
		g.output.uniBuf.Upload(byteslice.Struct(g.output.uniforms))
		g.ctx.BindUniforms(g.output.uniBuf)
		g.ctx.BindTexture(0, atlas.image)
		g.ctx.DrawArrays(start, count)
		start += count
	}
}

func (g *compute) renderMaterials() error {
	m := &g.materials
	for k, place := range m.allocs {
		if place.alloc.dead {
			delete(m.allocs, k)
		}
	}
	texOps := g.texOps
	for len(texOps) > 0 {
		m.quads = m.quads[:0]
		var (
			atlas    *textureAtlas
			imgAtlas *textureAtlas
		)
		// A material is clipped to avoid drawing outside its atlas bounds.
		// However, imprecision in the clipping may cause a single pixel
		// overflow.
		var padding = image.Pt(1, 1)
		var allocStart int
		for len(texOps) > 0 {
			op := &texOps[0]
			if a, exists := m.allocs[op.key]; exists {
				g.touchAlloc(a.alloc)
				op.matAlloc = a
				texOps = texOps[1:]
				continue
			}

			if imgAtlas != nil && op.imgAlloc.atlas != imgAtlas {
				// Only one image atlas per render pass.
				break
			}
			imgAtlas = op.imgAlloc.atlas
			quad := g.materialQuad(imgAtlas.size, op.key.transform, op.img, op.imgAlloc.rect.Min)
			boundsf := quadBounds(quad)
			bounds := boundsf.Round()
			bounds = bounds.Intersect(op.key.bounds)

			size := bounds.Size()
			alloc, fits := g.atlasAlloc(allocQuery{
				atlas:    atlas,
				size:     size.Add(padding),
				format:   driver.TextureFormatRGBA8,
				bindings: combinedBindings,
			})
			if !fits {
				break
			}
			if atlas == nil {
				allocStart = len(alloc.atlas.allocs)
			}
			atlas = alloc.atlas
			alloc.cpu = g.useCPU
			offsetf := layout.FPt(bounds.Min.Mul(-1))
			scale := f32.Pt(float32(size.X), float32(size.Y))
			for i := range quad {
				// Position quad to match place.
				quad[i].posX += offsetf.X
				quad[i].posY += offsetf.Y
				// Scale to match viewport [0, 1].
				quad[i].posX /= scale.X
				quad[i].posY /= scale.Y
			}
			// Draw quad as two triangles.
			m.quads = append(m.quads, quad[0], quad[1], quad[3], quad[3], quad[1], quad[2])
			if m.allocs == nil {
				m.allocs = make(map[textureKey]materialAlloc)
			}
			atlasAlloc := materialAlloc{
				alloc:  &alloc,
				offset: bounds.Min.Mul(-1),
			}
			atlas.allocs = append(atlas.allocs, atlasAlloc.alloc)
			m.allocs[op.key] = atlasAlloc
			op.matAlloc = atlasAlloc
			texOps = texOps[1:]
		}
		if len(m.quads) == 0 {
			break
		}
		realized := atlas.realized
		if err := g.realizeAtlas(atlas, g.useCPU, atlas.packer.sizes[0]); err != nil {
			return err
		}
		// Transform to clip space: [-1, -1] - [1, 1].
		*m.uniforms.u = materialUniforms{
			scale: [2]float32{2, 2},
			pos:   [2]float32{-1, -1},
		}
		if !g.srgb {
			m.uniforms.u.emulatesRGB = 1.0
		}
		m.uniforms.buf.Upload(byteslice.Struct(m.uniforms.u))
		vertexData := byteslice.Slice(m.quads)
		n := pow2Ceil(len(vertexData))
		m.buffer.ensureCapacity(false, g.ctx, driver.BufferBindingVertices, n)
		m.buffer.buffer.Upload(vertexData)
		var d driver.LoadDesc
		if !realized {
			d.Action = driver.LoadActionClear
		}
		g.ctx.PrepareTexture(imgAtlas.image)
		g.ctx.BeginRenderPass(atlas.image, d)
		g.ctx.BindTexture(0, imgAtlas.image)
		g.ctx.BindPipeline(m.pipeline)
		g.ctx.BindUniforms(m.uniforms.buf)
		g.ctx.BindVertexBuffer(m.buffer.buffer, 0)
		newAllocs := atlas.allocs[allocStart:]
		for i, a := range newAllocs {
			sz := a.rect.Size().Sub(padding)
			g.ctx.Viewport(a.rect.Min.X, a.rect.Min.Y, sz.X, sz.Y)
			g.ctx.DrawArrays(i*6, 6)
		}
		g.ctx.EndRenderPass()
		if !g.useCPU {
			continue
		}
		src := atlas.image
		data := atlas.cpuImage.Data()
		for _, a := range newAllocs {
			stride := atlas.size.X * 4
			col := a.rect.Min.X * 4
			row := stride * a.rect.Min.Y
			off := col + row
			src.ReadPixels(a.rect, data[off:], stride)
		}
	}
	return nil
}

func (g *compute) uploadImages() error {
	for k, a := range g.imgAllocs {
		if a.dead {
			delete(g.imgAllocs, k)
		}
	}
	type upload struct {
		pos image.Point
		img *image.RGBA
	}
	var uploads []upload
	format := driver.TextureFormatSRGBA
	if !g.srgb {
		format = driver.TextureFormatRGBA8
	}
	// padding is the number of pixels added to the right and below
	// images, to avoid atlas filtering artifacts.
	const padding = 1
	texOps := g.texOps
	for len(texOps) > 0 {
		uploads = uploads[:0]
		var atlas *textureAtlas
		for len(texOps) > 0 {
			op := &texOps[0]
			if a, exists := g.imgAllocs[op.img.handle]; exists {
				g.touchAlloc(a)
				op.imgAlloc = a
				texOps = texOps[1:]
				continue
			}
			size := op.img.src.Bounds().Size().Add(image.Pt(padding, padding))
			alloc, fits := g.atlasAlloc(allocQuery{
				atlas:    atlas,
				size:     size,
				format:   format,
				bindings: driver.BufferBindingTexture | driver.BufferBindingFramebuffer,
			})
			if !fits {
				break
			}
			atlas = alloc.atlas
			if g.imgAllocs == nil {
				g.imgAllocs = make(map[interface{}]*atlasAlloc)
			}
			op.imgAlloc = &alloc
			atlas.allocs = append(atlas.allocs, op.imgAlloc)
			g.imgAllocs[op.img.handle] = op.imgAlloc
			uploads = append(uploads, upload{pos: alloc.rect.Min, img: op.img.src})
			texOps = texOps[1:]
		}
		if len(uploads) == 0 {
			break
		}
		if err := g.realizeAtlas(atlas, false, atlas.packer.sizes[0]); err != nil {
			return err
		}
		for _, u := range uploads {
			size := u.img.Bounds().Size()
			driver.UploadImage(atlas.image, u.pos, u.img)
			rightPadding := image.Pt(padding, size.Y)
			atlas.image.Upload(image.Pt(u.pos.X+size.X, u.pos.Y), rightPadding, g.zeros(rightPadding.X*rightPadding.Y*4), 0)
			bottomPadding := image.Pt(size.X, padding)
			atlas.image.Upload(image.Pt(u.pos.X, u.pos.Y+size.Y), bottomPadding, g.zeros(bottomPadding.X*bottomPadding.Y*4), 0)
		}
	}
	return nil
}

func pow2Ceil(v int) int {
	exp := bits.Len(uint(v))
	if bits.OnesCount(uint(v)) == 1 {
		exp--
	}
	return 1 << exp
}

// materialQuad constructs a quad that represents the transformed image. It returns the quad
// and its bounds.
func (g *compute) materialQuad(imgAtlasSize image.Point, M f32.Affine2D, img imageOpData, uvPos image.Point) [4]materialVertex {
	imgSize := layout.FPt(img.src.Bounds().Size())
	sx, hx, ox, hy, sy, oy := M.Elems()
	transOff := f32.Pt(ox, oy)
	// The 4 corners of the image rectangle transformed by M, excluding its offset, are:
	//
	// q0: M * (0, 0)   q3: M * (w, 0)
	// q1: M * (0, h)   q2: M * (w, h)
	//
	// Note that q0 = M*0 = 0, q2 = q1 + q3.
	q0 := f32.Pt(0, 0)
	q1 := f32.Pt(hx*imgSize.Y, sy*imgSize.Y)
	q3 := f32.Pt(sx*imgSize.X, hy*imgSize.X)
	q2 := q1.Add(q3)
	q0 = q0.Add(transOff)
	q1 = q1.Add(transOff)
	q2 = q2.Add(transOff)
	q3 = q3.Add(transOff)

	uvPosf := layout.FPt(uvPos)
	atlasScale := f32.Pt(1/float32(imgAtlasSize.X), 1/float32(imgAtlasSize.Y))
	uvBounds := f32.Rectangle{
		Min: uvPosf,
		Max: uvPosf.Add(imgSize),
	}
	uvBounds.Min.X *= atlasScale.X
	uvBounds.Min.Y *= atlasScale.Y
	uvBounds.Max.X *= atlasScale.X
	uvBounds.Max.Y *= atlasScale.Y
	quad := [4]materialVertex{
		{posX: q0.X, posY: q0.Y, u: uvBounds.Min.X, v: uvBounds.Min.Y},
		{posX: q1.X, posY: q1.Y, u: uvBounds.Min.X, v: uvBounds.Max.Y},
		{posX: q2.X, posY: q2.Y, u: uvBounds.Max.X, v: uvBounds.Max.Y},
		{posX: q3.X, posY: q3.Y, u: uvBounds.Max.X, v: uvBounds.Min.Y},
	}
	return quad
}

func quadBounds(q [4]materialVertex) f32.Rectangle {
	q0 := f32.Pt(q[0].posX, q[0].posY)
	q1 := f32.Pt(q[1].posX, q[1].posY)
	q2 := f32.Pt(q[2].posX, q[2].posY)
	q3 := f32.Pt(q[3].posX, q[3].posY)
	return f32.Rectangle{
		Min: min(min(q0, q1), min(q2, q3)),
		Max: max(max(q0, q1), max(q2, q3)),
	}
}

func max(p1, p2 f32.Point) f32.Point {
	p := p1
	if p2.X > p.X {
		p.X = p2.X
	}
	if p2.Y > p.Y {
		p.Y = p2.Y
	}
	return p
}

func min(p1, p2 f32.Point) f32.Point {
	p := p1
	if p2.X < p.X {
		p.X = p2.X
	}
	if p2.Y < p.Y {
		p.Y = p2.Y
	}
	return p
}

func (enc *encoder) encodePath(verts []byte, fillMode int) {
	for ; len(verts) >= scene.CommandSize+4; verts = verts[scene.CommandSize+4:] {
		cmd := ops.DecodeCommand(verts[4:])
		if cmd.Op() == scene.OpGap {
			if fillMode != scene.FillModeNonzero {
				// Skip gaps in strokes.
				continue
			}
			// Replace them by a straight line in outlines.
			cmd = scene.Line(scene.DecodeGap(cmd))
		}
		enc.scene = append(enc.scene, cmd)
		enc.npathseg++
	}
}

func (g *compute) render(images *textureAtlas, dst driver.Texture, cpuDst cpu.ImageDescriptor, tileDims image.Point, stride int) error {
	const (
		// wgSize is the largest and most common workgroup size.
		wgSize = 128
		// PARTITION_SIZE from elements.comp
		partitionSize = 32 * 4
	)
	widthInBins := (tileDims.X + 15) / 16
	heightInBins := (tileDims.Y + 7) / 8
	if widthInBins*heightInBins > wgSize {
		return fmt.Errorf("gpu: output too large (%dx%d)", tileDims.X*tileWidthPx, tileDims.Y*tileHeightPx)
	}

	enc := &g.enc
	// Pad scene with zeroes to avoid reading garbage in elements.comp.
	scenePadding := partitionSize - len(enc.scene)%partitionSize
	enc.scene = append(enc.scene, make([]scene.Command, scenePadding)...)

	scene := byteslice.Slice(enc.scene)
	if s := len(scene); s > g.buffers.scene.size {
		paddedCap := s * 11 / 10
		if err := g.buffers.scene.ensureCapacity(g.useCPU, g.ctx, driver.BufferBindingShaderStorageRead, paddedCap); err != nil {
			return err
		}
	}
	g.buffers.scene.upload(scene)

	// alloc is the number of allocated bytes for static buffers.
	var alloc uint32
	round := func(v, quantum int) int {
		return (v + quantum - 1) &^ (quantum - 1)
	}
	malloc := func(size int) memAlloc {
		size = round(size, 4)
		offset := alloc
		alloc += uint32(size)
		return memAlloc{offset /*, uint32(size)*/}
	}

	*g.conf = config{
		n_elements:      uint32(enc.npath),
		n_pathseg:       uint32(enc.npathseg),
		width_in_tiles:  uint32(tileDims.X),
		height_in_tiles: uint32(tileDims.Y),
		tile_alloc:      malloc(enc.npath * pathSize),
		bin_alloc:       malloc(round(enc.npath, wgSize) * binSize),
		ptcl_alloc:      malloc(tileDims.X * tileDims.Y * ptclInitialAlloc),
		pathseg_alloc:   malloc(enc.npathseg * pathsegSize),
		anno_alloc:      malloc(enc.npath * annoSize),
		trans_alloc:     malloc(enc.ntrans * transSize),
	}

	numPartitions := (enc.numElements() + 127) / 128
	// clearSize is the atomic partition counter plus flag and 2 states per partition.
	clearSize := 4 + numPartitions*stateStride
	if clearSize > g.buffers.state.size {
		paddedCap := clearSize * 11 / 10
		if err := g.buffers.state.ensureCapacity(g.useCPU, g.ctx, driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite, paddedCap); err != nil {
			return err
		}
	}

	confData := byteslice.Struct(g.conf)
	g.buffers.config.ensureCapacity(g.useCPU, g.ctx, driver.BufferBindingShaderStorageRead, len(confData))
	g.buffers.config.upload(confData)

	minSize := int(unsafe.Sizeof(memoryHeader{})) + int(alloc)
	if minSize > g.buffers.memory.size {
		// Add space for dynamic GPU allocations.
		const sizeBump = 4 * 1024 * 1024
		minSize += sizeBump
		if err := g.buffers.memory.ensureCapacity(g.useCPU, g.ctx, driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite, minSize); err != nil {
			return err
		}
	}

	for {
		*g.memHeader = memoryHeader{
			mem_offset: alloc,
		}
		g.buffers.memory.upload(byteslice.Struct(g.memHeader))
		g.buffers.state.upload(g.zeros(clearSize))

		if !g.useCPU {
			g.ctx.BeginCompute()
			g.ctx.BindImageTexture(kernel4OutputUnit, dst)
			img := g.output.nullMaterials
			if images != nil {
				img = images.image
			}
			g.ctx.BindImageTexture(kernel4AtlasUnit, img)
		} else {
			*g.output.descriptors.Binding2() = cpuDst
			if images != nil {
				*g.output.descriptors.Binding3() = images.cpuImage
			}
		}

		g.bindBuffers()
		g.memoryBarrier()
		g.dispatch(g.programs.elements, numPartitions, 1, 1)
		g.memoryBarrier()
		g.dispatch(g.programs.tileAlloc, (enc.npath+wgSize-1)/wgSize, 1, 1)
		g.memoryBarrier()
		g.dispatch(g.programs.pathCoarse, (enc.npathseg+31)/32, 1, 1)
		g.memoryBarrier()
		g.dispatch(g.programs.backdrop, (enc.npath+wgSize-1)/wgSize, 1, 1)
		// No barrier needed between backdrop and binning.
		g.dispatch(g.programs.binning, (enc.npath+wgSize-1)/wgSize, 1, 1)
		g.memoryBarrier()
		g.dispatch(g.programs.coarse, widthInBins, heightInBins, 1)
		g.memoryBarrier()
		g.dispatch(g.programs.kernel4, tileDims.X, tileDims.Y, 1)
		g.memoryBarrier()
		if !g.useCPU {
			g.ctx.EndCompute()
		} else {
			g.dispatcher.Sync()
		}

		if err := g.buffers.memory.download(byteslice.Struct(g.memHeader)); err != nil {
			if err == driver.ErrContentLost {
				continue
			}
			return err
		}
		switch errCode := g.memHeader.mem_error; errCode {
		case memNoError:
			if g.useCPU {
				w, h := tileDims.X*tileWidthPx, tileDims.Y*tileHeightPx
				dst.Upload(image.Pt(0, 0), image.Pt(w, h), cpuDst.Data(), stride)
			}
			return nil
		case memMallocFailed:
			// Resize memory and try again.
			sz := g.buffers.memory.size * 15 / 10
			if err := g.buffers.memory.ensureCapacity(g.useCPU, g.ctx, driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite, sz); err != nil {
				return err
			}
			continue
		default:
			return fmt.Errorf("compute: shader program failed with error %d", errCode)
		}
	}
}

func (g *compute) memoryBarrier() {
	if g.useCPU {
		g.dispatcher.Barrier()
	}
}

func (g *compute) dispatch(p computeProgram, x, y, z int) {
	if !g.useCPU {
		g.ctx.BindProgram(p.prog)
		g.ctx.DispatchCompute(x, y, z)
	} else {
		g.dispatcher.Dispatch(p.progInfo, p.descriptors, x, y, z)
	}
}

// zeros returns a byte slice with size bytes of zeros.
func (g *compute) zeros(size int) []byte {
	if cap(g.zeroSlice) < size {
		g.zeroSlice = append(g.zeroSlice, make([]byte, size)...)
	}
	return g.zeroSlice[:size]
}

func (g *compute) touchAlloc(a *atlasAlloc) {
	if a.dead {
		panic("re-use of dead allocation")
	}
	a.frameCount = g.frameCount
	a.atlas.lastFrame = a.frameCount
}

func (g *compute) atlasAlloc(q allocQuery) (atlasAlloc, bool) {
	var (
		place placement
		fits  bool
		atlas = q.atlas
	)
	if atlas != nil {
		place, fits = atlas.packer.tryAdd(q.size)
		if !fits {
			atlas.compact = true
		}
	}
	if atlas == nil {
		// Look for matching atlas to re-use.
		for _, a := range g.atlases {
			if q.empty && len(a.allocs) > 0 {
				continue
			}
			if q.nocompact && a.compact {
				continue
			}
			if a.format != q.format || a.bindings&q.bindings != q.bindings {
				continue
			}
			place, fits = a.packer.tryAdd(q.size)
			if !fits {
				a.compact = true
				continue
			}
			atlas = a
			break
		}
	}
	if atlas == nil {
		atlas = &textureAtlas{
			format:   q.format,
			bindings: q.bindings,
		}
		atlas.packer.maxDims = image.Pt(g.maxTextureDim, g.maxTextureDim)
		atlas.packer.newPage()
		g.atlases = append(g.atlases, atlas)
		place, fits = atlas.packer.tryAdd(q.size)
		if !fits {
			panic(fmt.Errorf("compute: atlas allocation too large (%v)", q.size))
		}
	}
	if !fits {
		return atlasAlloc{}, false
	}
	atlas.lastFrame = g.frameCount
	return atlasAlloc{
		frameCount: g.frameCount,
		atlas:      atlas,
		rect:       image.Rectangle{Min: place.Pos, Max: place.Pos.Add(q.size)},
	}, true
}

func (g *compute) realizeAtlas(atlas *textureAtlas, useCPU bool, size image.Point) error {
	defer func() {
		atlas.packer.maxDims = atlas.size
		atlas.realized = true
		atlas.ensureCPUImage(useCPU)
	}()
	if atlas.size.X >= size.X && atlas.size.Y >= size.Y {
		return nil
	}
	if atlas.realized {
		panic("resizing a realized atlas")
	}
	if err := atlas.resize(g.ctx, size); err != nil {
		return err
	}
	return nil
}

func (a *textureAtlas) resize(ctx driver.Device, size image.Point) error {
	a.Release()

	img, err := ctx.NewTexture(a.format, size.X, size.Y,
		driver.FilterNearest,
		driver.FilterNearest,
		a.bindings)
	if err != nil {
		return err
	}
	a.image = img
	a.size = size
	return nil
}

func (a *textureAtlas) ensureCPUImage(useCPU bool) {
	if !useCPU || a.hasCPU {
		return
	}
	a.hasCPU = true
	a.cpuImage = cpu.NewImageRGBA(a.size.X, a.size.Y)
}

func (g *compute) Release() {
	if g.useCPU {
		g.dispatcher.Stop()
	}
	type resource interface {
		Release()
	}
	res := []resource{
		g.output.nullMaterials,
		&g.programs.elements,
		&g.programs.tileAlloc,
		&g.programs.pathCoarse,
		&g.programs.backdrop,
		&g.programs.binning,
		&g.programs.coarse,
		&g.programs.kernel4,
		g.output.blitPipeline,
		&g.output.buffer,
		g.output.uniBuf,
		&g.buffers.scene,
		&g.buffers.state,
		&g.buffers.memory,
		&g.buffers.config,
		g.materials.pipeline,
		&g.materials.buffer,
		g.materials.uniforms.buf,
		g.timers.t,
	}
	for _, r := range res {
		if r != nil {
			r.Release()
		}
	}
	for _, a := range g.atlases {
		a.Release()
	}
	g.ctx.Release()
	*g = compute{}
}

func (a *textureAtlas) Release() {
	if a.image != nil {
		a.image.Release()
		a.image = nil
	}
	a.cpuImage.Free()
	a.hasCPU = false
}

func (g *compute) bindBuffers() {
	g.bindStorageBuffers(g.programs.elements, g.buffers.memory, g.buffers.config, g.buffers.scene, g.buffers.state)
	g.bindStorageBuffers(g.programs.tileAlloc, g.buffers.memory, g.buffers.config)
	g.bindStorageBuffers(g.programs.pathCoarse, g.buffers.memory, g.buffers.config)
	g.bindStorageBuffers(g.programs.backdrop, g.buffers.memory, g.buffers.config)
	g.bindStorageBuffers(g.programs.binning, g.buffers.memory, g.buffers.config)
	g.bindStorageBuffers(g.programs.coarse, g.buffers.memory, g.buffers.config)
	g.bindStorageBuffers(g.programs.kernel4, g.buffers.memory, g.buffers.config)
}

func (p *computeProgram) Release() {
	if p.prog != nil {
		p.prog.Release()
	}
	*p = computeProgram{}
}

func (b *sizedBuffer) Release() {
	if b.buffer != nil {
		b.buffer.Release()
	}
	b.cpuBuf.Free()
	*b = sizedBuffer{}
}

func (b *sizedBuffer) ensureCapacity(useCPU bool, ctx driver.Device, binding driver.BufferBinding, size int) error {
	if b.size >= size {
		return nil
	}
	if b.buffer != nil {
		b.Release()
	}
	b.cpuBuf.Free()
	if !useCPU {
		buf, err := ctx.NewBuffer(binding, size)
		if err != nil {
			return err
		}
		b.buffer = buf
	} else {
		b.cpuBuf = cpu.NewBuffer(size)
	}
	b.size = size
	return nil
}

func (b *sizedBuffer) download(data []byte) error {
	if b.buffer != nil {
		return b.buffer.Download(data)
	} else {
		copy(data, b.cpuBuf.Data())
		return nil
	}
}

func (b *sizedBuffer) upload(data []byte) {
	if b.buffer != nil {
		b.buffer.Upload(data)
	} else {
		copy(b.cpuBuf.Data(), data)
	}
}

func (g *compute) bindStorageBuffers(prog computeProgram, buffers ...sizedBuffer) {
	for i, buf := range buffers {
		if !g.useCPU {
			g.ctx.BindStorageBuffer(i, buf.buffer)
		} else {
			*prog.buffers[i] = buf.cpuBuf
		}
	}
}

var bo = binary.LittleEndian

func (e *encoder) reset() {
	e.scene = e.scene[:0]
	e.npath = 0
	e.npathseg = 0
	e.ntrans = 0
}

func (e *encoder) numElements() int {
	return len(e.scene)
}

func (e *encoder) transform(m f32.Affine2D) {
	e.scene = append(e.scene, scene.Transform(m))
	e.ntrans++
}

func (e *encoder) lineWidth(width float32) {
	e.scene = append(e.scene, scene.SetLineWidth(width))
}

func (e *encoder) fillMode(mode scene.FillMode) {
	e.scene = append(e.scene, scene.SetFillMode(mode))
}

func (e *encoder) beginClip(bbox f32.Rectangle) {
	e.scene = append(e.scene, scene.BeginClip(bbox))
	e.npath++
}

func (e *encoder) endClip(bbox f32.Rectangle) {
	e.scene = append(e.scene, scene.EndClip(bbox))
	e.npath++
}

func (e *encoder) rect(r f32.Rectangle) {
	// Rectangle corners, clock-wise.
	c0, c1, c2, c3 := r.Min, f32.Pt(r.Min.X, r.Max.Y), r.Max, f32.Pt(r.Max.X, r.Min.Y)
	e.line(c0, c1)
	e.line(c1, c2)
	e.line(c2, c3)
	e.line(c3, c0)
}

func (e *encoder) fillColor(col color.RGBA) {
	e.scene = append(e.scene, scene.FillColor(col))
	e.npath++
}

func (e *encoder) fillImage(index int, offset image.Point) {
	e.scene = append(e.scene, scene.FillImage(index, offset))
	e.npath++
}

func (e *encoder) line(start, end f32.Point) {
	e.scene = append(e.scene, scene.Line(start, end))
	e.npathseg++
}

func (c *collector) reset() {
	c.prevFrame, c.frame = c.frame, c.prevFrame
	c.profile = false
	c.clipStates = c.clipStates[:0]
	c.transStack = c.transStack[:0]
	c.frame.reset()
}

func (c *opsCollector) reset() {
	c.paths = c.paths[:0]
	c.clipCmds = c.clipCmds[:0]
	c.ops = c.ops[:0]
	c.layers = c.layers[:0]
}

func (c *collector) addClip(state *encoderState, viewport, bounds f32.Rectangle, path []byte, key ops.Key, hash uint64, strokeWidth float32, push bool) {
	// Rectangle clip regions.
	if len(path) == 0 && !push {
		// If the rectangular clip region contains a previous path it can be discarded.
		p := state.clip
		t := state.relTrans.Invert()
		for p != nil {
			// rect is the parent bounds transformed relative to the rectangle.
			rect := transformBounds(t, p.bounds)
			if rect.In(bounds) {
				return
			}
			t = p.relTrans.Invert().Mul(t)
			p = p.parent
		}
	}

	absBounds := transformBounds(state.t, bounds).Bounds()
	intersect := absBounds
	if state.clip != nil {
		intersect = state.clip.intersect.Intersect(intersect)
	}
	c.clipStates = append(c.clipStates, clipState{
		parent:    state.clip,
		absBounds: absBounds,
		path:      path,
		pathKey:   key,
		intersect: intersect,
		clipKey: clipKey{
			bounds:      bounds,
			relTrans:    state.relTrans,
			strokeWidth: strokeWidth,
			pathHash:    hash,
		},
	})
	state.clip = &c.clipStates[len(c.clipStates)-1]
	state.relTrans = f32.Affine2D{}
}

func (c *collector) collect(root *op.Ops, viewport image.Point, texOps *[]textureOp) {
	fview := f32.Rectangle{Max: layout.FPt(viewport)}
	var intOps *ops.Ops
	if root != nil {
		intOps = &root.Internal
	}
	c.reader.Reset(intOps)
	var state encoderState
	reset := func() {
		state = encoderState{
			paintKey: paintKey{
				color: color.NRGBA{A: 0xff},
			},
		}
	}
	reset()
	r := &c.reader
	var (
		pathData struct {
			data []byte
			key  ops.Key
			hash uint64
		}
		strWidth float32
	)
	c.addClip(&state, fview, fview, nil, ops.Key{}, 0, 0, false)
	for encOp, ok := r.Decode(); ok; encOp, ok = r.Decode() {
		switch ops.OpType(encOp.Data[0]) {
		case ops.TypeProfile:
			c.profile = true
		case ops.TypeTransform:
			dop, push := ops.DecodeTransform(encOp.Data)
			if push {
				c.transStack = append(c.transStack, transEntry{t: state.t, relTrans: state.relTrans})
			}
			state.t = state.t.Mul(dop)
			state.relTrans = state.relTrans.Mul(dop)
		case ops.TypePopTransform:
			n := len(c.transStack)
			st := c.transStack[n-1]
			c.transStack = c.transStack[:n-1]
			state.t = st.t
			state.relTrans = st.relTrans
		case ops.TypeStroke:
			strWidth = decodeStrokeOp(encOp.Data)
		case ops.TypePath:
			hash := bo.Uint64(encOp.Data[1:])
			encOp, ok = r.Decode()
			if !ok {
				panic("unexpected end of path operation")
			}
			pathData.data = encOp.Data[ops.TypeAuxLen:]
			pathData.key = encOp.Key
			pathData.hash = hash
		case ops.TypeClip:
			var op ops.ClipOp
			op.Decode(encOp.Data)
			bounds := f32.FRect(op.Bounds)
			c.addClip(&state, fview, bounds, pathData.data, pathData.key, pathData.hash, strWidth, true)
			pathData.data = nil
			strWidth = 0
		case ops.TypePopClip:
			state.relTrans = state.clip.relTrans.Mul(state.relTrans)
			state.clip = state.clip.parent
		case ops.TypeColor:
			state.matType = materialColor
			state.color = decodeColorOp(encOp.Data)
		case ops.TypeLinearGradient:
			state.matType = materialLinearGradient
			op := decodeLinearGradientOp(encOp.Data)
			state.stop1 = op.stop1
			state.stop2 = op.stop2
			state.color1 = op.color1
			state.color2 = op.color2
		case ops.TypeImage:
			state.matType = materialTexture
			state.image = decodeImageOp(encOp.Data, encOp.Refs)
		case ops.TypePaint:
			paintState := state
			if paintState.matType == materialTexture {
				// Clip to the bounds of the image, to hide other images in the atlas.
				sz := state.image.src.Rect.Size()
				bounds := f32.Rectangle{Max: layout.FPt(sz)}
				c.addClip(&paintState, fview, bounds, nil, ops.Key{}, 0, 0, false)
			}
			intersect := paintState.clip.intersect
			if intersect.Empty() {
				break
			}

			// If the paint is a uniform opaque color that takes up the whole
			// screen, it covers all previous paints and we can discard all
			// rendering commands recorded so far.
			if paintState.clip == nil && paintState.matType == materialColor && paintState.color.A == 255 {
				c.clearColor = f32color.LinearFromSRGB(paintState.color).Opaque()
				c.clear = true
				c.frame.reset()
				break
			}

			// Flatten clip stack.
			p := paintState.clip
			startIdx := len(c.frame.clipCmds)
			for p != nil {
				idx := len(c.frame.paths)
				c.frame.paths = append(c.frame.paths, make([]byte, len(p.path))...)
				path := c.frame.paths[idx:]
				copy(path, p.path)
				c.frame.clipCmds = append(c.frame.clipCmds, clipCmd{
					state:     p.clipKey,
					path:      path,
					pathKey:   p.pathKey,
					absBounds: p.absBounds,
				})
				p = p.parent
			}
			clipStack := c.frame.clipCmds[startIdx:]
			c.frame.ops = append(c.frame.ops, paintOp{
				clipStack: clipStack,
				state:     paintState.paintKey,
				intersect: intersect,
			})
		case ops.TypeSave:
			id := ops.DecodeSave(encOp.Data)
			c.save(id, state.t)
		case ops.TypeLoad:
			reset()
			id := ops.DecodeLoad(encOp.Data)
			state.t = c.states[id]
			state.relTrans = state.t
		}
	}
	for i := range c.frame.ops {
		op := &c.frame.ops[i]
		// For each clip, cull rectangular clip regions that contain its
		// (transformed) bounds. addClip already handled the converse case.
		// TODO: do better than O(n²) to efficiently deal with deep stacks.
		for j := 0; j < len(op.clipStack)-1; j++ {
			cl := op.clipStack[j]
			p := cl.state
			r := transformBounds(p.relTrans, p.bounds)
			for k := j + 1; k < len(op.clipStack); k++ {
				cl2 := op.clipStack[k]
				p2 := cl2.state
				if len(cl2.path) == 0 && r.In(cl2.state.bounds) {
					op.clipStack = append(op.clipStack[:k], op.clipStack[k+1:]...)
					k--
					op.clipStack[k].state.relTrans = p2.relTrans.Mul(op.clipStack[k].state.relTrans)
				}
				r = transformRect(p2.relTrans, r)
			}
		}
		// Separate the integer offset from the first transform. Two ops that differ
		// only in integer offsets may share backing storage.
		if len(op.clipStack) > 0 {
			c := &op.clipStack[len(op.clipStack)-1]
			t := c.state.relTrans
			t, off := separateTransform(t)
			c.state.relTrans = t
			op.offset = off
			op.state.t = op.state.t.Offset(layout.FPt(off.Mul(-1)))
		}
		op.hash = c.hashOp(*op)
		op.texOpIdx = -1
		switch op.state.matType {
		case materialTexture:
			op.texOpIdx = len(*texOps)
			// Separate integer offset from transformation. TextureOps that have identical transforms
			// except for their integer offsets can share a transformed image.
			t := op.state.t.Offset(layout.FPt(op.offset))
			t, off := separateTransform(t)
			bounds := op.intersect.Round().Sub(off)
			*texOps = append(*texOps, textureOp{
				img: op.state.image,
				off: off,
				key: textureKey{
					bounds:    bounds,
					transform: t,
					handle:    op.state.image.handle,
				},
			})
		}
	}
}

func (c *collector) hashOp(op paintOp) uint64 {
	c.hasher.Reset()
	for _, cl := range op.clipStack {
		k := cl.state
		keyBytes := (*[unsafe.Sizeof(k)]byte)(unsafe.Pointer(unsafe.Pointer(&k)))
		c.hasher.Write(keyBytes[:])
	}
	k := op.state
	keyBytes := (*[unsafe.Sizeof(k)]byte)(unsafe.Pointer(unsafe.Pointer(&k)))
	c.hasher.Write(keyBytes[:])
	return c.hasher.Sum64()
}

func (g *compute) layer(viewport image.Point, texOps []textureOp) {
	// Sort ops from previous frames by hash.
	c := &g.collector
	prevOps := c.prevFrame.ops
	c.order = c.order[:0]
	for i, op := range prevOps {
		c.order = append(c.order, hashIndex{
			index: i,
			hash:  op.hash,
		})
	}
	sort.Slice(c.order, func(i, j int) bool {
		return c.order[i].hash < c.order[j].hash
	})
	// Split layers with different materials atlas; the compute stage has only
	// one materials slot.
	splitLayer := func(ops []paintOp, prevLayerIdx int) {
		for len(ops) > 0 {
			var materials *textureAtlas
			idx := 0
			for idx < len(ops) {
				if i := ops[idx].texOpIdx; i != -1 {
					omats := texOps[i].matAlloc.alloc.atlas
					if materials != nil && omats != nil && omats != materials {
						break
					}
					materials = omats
				}
				idx++
			}
			l := layer{ops: ops[:idx], materials: materials}
			if prevLayerIdx != -1 {
				prev := c.prevFrame.layers[prevLayerIdx]
				if !prev.alloc.dead && len(prev.ops) == len(l.ops) {
					l.alloc = prev.alloc
					l.materials = prev.materials
					g.touchAlloc(l.alloc)
				}
			}
			for i, op := range l.ops {
				l.rect = l.rect.Union(op.intersect.Round())
				l.ops[i].layer = len(c.frame.layers)
			}
			c.frame.layers = append(c.frame.layers, l)
			ops = ops[idx:]
		}
	}
	ops := c.frame.ops
	idx := 0
	for idx < len(ops) {
		op := ops[idx]
		// Search for longest matching op sequence.
		// start is the earliest index of a match.
		start := searchOp(c.order, op.hash)
		layerOps, prevLayerIdx := longestLayer(prevOps, c.order[start:], ops[idx:])
		if len(layerOps) == 0 {
			idx++
			continue
		}
		if unmatched := ops[:idx]; len(unmatched) > 0 {
			// Flush layer of unmatched ops.
			splitLayer(unmatched, -1)
			ops = ops[idx:]
			idx = 0
		}
		splitLayer(layerOps, prevLayerIdx)
		ops = ops[len(layerOps):]
	}
	if len(ops) > 0 {
		splitLayer(ops, -1)
	}
}

func longestLayer(prev []paintOp, order []hashIndex, ops []paintOp) ([]paintOp, int) {
	longest := 0
	longestIdx := -1
outer:
	for len(order) > 0 {
		first := order[0]
		order = order[1:]
		match := prev[first.index:]
		// Potential match found. Now find longest matching sequence.
		end := 0
		layer := match[0].layer
		off := match[0].offset.Sub(ops[0].offset)
		for end < len(match) && end < len(ops) {
			m := match[end]
			o := ops[end]
			// End layers on previous match.
			if m.layer != layer {
				break
			}
			// End layer when the next op doesn't match.
			if m.hash != o.hash {
				if end == 0 {
					// Hashes are sorted so if the first op doesn't match, no
					// more matches are possible.
					break outer
				}
				break
			}
			if !opEqual(off, m, o) {
				break
			}
			end++
		}
		if end > longest {
			longest = end
			longestIdx = layer

		}
	}
	return ops[:longest], longestIdx
}

func searchOp(order []hashIndex, hash uint64) int {
	lo, hi := 0, len(order)
	for lo < hi {
		mid := (lo + hi) / 2
		if order[mid].hash < hash {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

func opEqual(off image.Point, o1 paintOp, o2 paintOp) bool {
	if len(o1.clipStack) != len(o2.clipStack) {
		return false
	}
	if o1.state != o2.state {
		return false
	}
	if o1.offset.Sub(o2.offset) != off {
		return false
	}
	for i, cl1 := range o1.clipStack {
		cl2 := o2.clipStack[i]
		if len(cl1.path) != len(cl2.path) {
			return false
		}
		if cl1.state != cl2.state {
			return false
		}
		if cl1.pathKey != cl2.pathKey && !bytes.Equal(cl1.path, cl2.path) {
			return false
		}
	}
	return true
}

func encodeLayer(l layer, pos image.Point, viewport image.Point, enc *encoder, texOps []textureOp) {
	off := pos.Sub(l.rect.Min)
	offf := layout.FPt(off)

	enc.transform(f32.Affine2D{}.Offset(offf))
	for _, op := range l.ops {
		encodeOp(viewport, off, enc, texOps, op)
	}
	enc.transform(f32.Affine2D{}.Offset(offf.Mul(-1)))
}

func encodeOp(viewport image.Point, absOff image.Point, enc *encoder, texOps []textureOp, op paintOp) {
	// Fill in clip bounds, which the shaders expect to be the union
	// of all affected bounds.
	var union f32.Rectangle
	for i, cl := range op.clipStack {
		union = union.Union(cl.absBounds)
		op.clipStack[i].union = union
	}

	absOfff := layout.FPt(absOff)
	fillMode := scene.FillModeNonzero
	opOff := layout.FPt(op.offset)
	inv := f32.Affine2D{}.Offset(opOff)
	enc.transform(inv)
	for i := len(op.clipStack) - 1; i >= 0; i-- {
		cl := op.clipStack[i]
		if w := cl.state.strokeWidth; w > 0 {
			enc.fillMode(scene.FillModeStroke)
			enc.lineWidth(w)
			fillMode = scene.FillModeStroke
		} else if fillMode != scene.FillModeNonzero {
			enc.fillMode(scene.FillModeNonzero)
			fillMode = scene.FillModeNonzero
		}
		enc.transform(cl.state.relTrans)
		inv = inv.Mul(cl.state.relTrans)
		if len(cl.path) == 0 {
			enc.rect(cl.state.bounds)
		} else {
			enc.encodePath(cl.path, fillMode)
		}
		if i != 0 {
			enc.beginClip(cl.union.Add(absOfff))
		}
	}
	if len(op.clipStack) == 0 {
		// No clipping; fill the entire view.
		enc.rect(f32.Rectangle{Max: layout.FPt(viewport)})
	}

	switch op.state.matType {
	case materialTexture:
		texOp := texOps[op.texOpIdx]
		off := texOp.matAlloc.alloc.rect.Min.Add(texOp.matAlloc.offset).Sub(texOp.off).Sub(absOff)
		enc.fillImage(0, off)
	case materialColor:
		enc.fillColor(f32color.NRGBAToRGBA(op.state.color))
	case materialLinearGradient:
		// TODO: implement.
		enc.fillColor(f32color.NRGBAToRGBA(op.state.color1))
	default:
		panic("not implemented")
	}
	enc.transform(inv.Invert())
	// Pop the clip stack, except the first entry used for fill.
	for i := 1; i < len(op.clipStack); i++ {
		cl := op.clipStack[i]
		enc.endClip(cl.union.Add(absOfff))
	}
	if fillMode != scene.FillModeNonzero {
		enc.fillMode(scene.FillModeNonzero)
	}
}

func (c *collector) save(id int, state f32.Affine2D) {
	if extra := id - len(c.states) + 1; extra > 0 {
		c.states = append(c.states, make([]f32.Affine2D, extra)...)
	}
	c.states[id] = state
}

func transformBounds(t f32.Affine2D, bounds f32.Rectangle) rectangle {
	return rectangle{
		t.Transform(bounds.Min), t.Transform(f32.Pt(bounds.Max.X, bounds.Min.Y)),
		t.Transform(bounds.Max), t.Transform(f32.Pt(bounds.Min.X, bounds.Max.Y)),
	}
}

func separateTransform(t f32.Affine2D) (f32.Affine2D, image.Point) {
	sx, hx, ox, hy, sy, oy := t.Elems()
	intx, fracx := math.Modf(float64(ox))
	inty, fracy := math.Modf(float64(oy))
	t = f32.NewAffine2D(sx, hx, float32(fracx), hy, sy, float32(fracy))
	return t, image.Pt(int(intx), int(inty))
}

func transformRect(t f32.Affine2D, r rectangle) rectangle {
	var tr rectangle
	for i, c := range r {
		tr[i] = t.Transform(c)
	}
	return tr
}

func (r rectangle) In(b f32.Rectangle) bool {
	for _, c := range r {
		inside := b.Min.X <= c.X && c.X <= b.Max.X &&
			b.Min.Y <= c.Y && c.Y <= b.Max.Y
		if !inside {
			return false
		}
	}
	return true
}

func (r rectangle) Contains(b f32.Rectangle) bool {
	return true
}

func (r rectangle) Bounds() f32.Rectangle {
	bounds := f32.Rectangle{
		Min: f32.Pt(math.MaxFloat32, math.MaxFloat32),
		Max: f32.Pt(-math.MaxFloat32, -math.MaxFloat32),
	}
	for _, c := range r {
		if c.X < bounds.Min.X {
			bounds.Min.X = c.X
		}
		if c.Y < bounds.Min.Y {
			bounds.Min.Y = c.Y
		}
		if c.X > bounds.Max.X {
			bounds.Max.X = c.X
		}
		if c.Y > bounds.Max.Y {
			bounds.Max.Y = c.Y
		}
	}
	return bounds
}
