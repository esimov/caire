// SPDX-License-Identifier: Unlicense OR MIT

/*
Package gpu implements the rendering of Gio drawing operations. It
is used by package app and package app/headless and is otherwise not
useful except for integrating with external window implementations.
*/
package gpu

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"reflect"
	"time"
	"unsafe"

	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/byteslice"
	"gioui.org/internal/f32"
	"gioui.org/internal/f32color"
	"gioui.org/internal/ops"
	"gioui.org/internal/scene"
	"gioui.org/internal/stroke"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/shader"
	"gioui.org/shader/gio"

	// Register backends.
	_ "gioui.org/gpu/internal/d3d11"
	_ "gioui.org/gpu/internal/metal"
	_ "gioui.org/gpu/internal/opengl"
	_ "gioui.org/gpu/internal/vulkan"
)

type GPU interface {
	// Release non-Go resources. The GPU is no longer valid after Release.
	Release()
	// Clear sets the clear color for the next Frame.
	Clear(color color.NRGBA)
	// Frame draws the graphics operations from op into a viewport of target.
	Frame(frame *op.Ops, target RenderTarget, viewport image.Point) error
	// Profile returns the last available profiling information. Profiling
	// information is requested when Frame sees an io/profile.Op, and the result
	// is available through Profile at some later time.
	Profile() string
}

type gpu struct {
	cache *resourceCache

	profile                                string
	timers                                 *timers
	frameStart                             time.Time
	stencilTimer, coverTimer, cleanupTimer *timer
	drawOps                                drawOps
	ctx                                    driver.Device
	renderer                               *renderer
}

type renderer struct {
	ctx           driver.Device
	blitter       *blitter
	pather        *pather
	packer        packer
	intersections packer
}

type drawOps struct {
	profile     bool
	reader      ops.Reader
	states      []f32.Affine2D
	transStack  []f32.Affine2D
	vertCache   []byte
	viewport    image.Point
	clear       bool
	clearColor  f32color.RGBA
	imageOps    []imageOp
	pathOps     []*pathOp
	pathOpCache []pathOp
	qs          quadSplitter
	pathCache   *opCache
}

type drawState struct {
	t     f32.Affine2D
	cpath *pathOp

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

type pathOp struct {
	off f32.Point
	// rect tracks whether the clip stack can be represented by a
	// pixel-aligned rectangle.
	rect bool
	// clip is the union of all
	// later clip rectangles.
	clip   image.Rectangle
	bounds f32.Rectangle
	// intersect is the intersection of bounds and all
	// previous clip bounds.
	intersect f32.Rectangle
	pathKey   opKey
	path      bool
	pathVerts []byte
	parent    *pathOp
	place     placement
}

type imageOp struct {
	path     *pathOp
	clip     image.Rectangle
	material material
	clipType clipType
	place    placement
}

func decodeStrokeOp(data []byte) float32 {
	_ = data[4]
	bo := binary.LittleEndian
	return math.Float32frombits(bo.Uint32(data[1:]))
}

type quadsOp struct {
	key opKey
	aux []byte
}

type opKey struct {
	outline        bool
	strokeWidth    float32
	sx, hx, sy, hy float32
	ops.Key
}

type material struct {
	material materialType
	opaque   bool
	// For materialTypeColor.
	color f32color.RGBA
	// For materialTypeLinearGradient.
	color1 f32color.RGBA
	color2 f32color.RGBA
	// For materialTypeTexture.
	data    imageOpData
	uvTrans f32.Affine2D
}

// imageOpData is the shadow of paint.ImageOp.
type imageOpData struct {
	src    *image.RGBA
	handle interface{}
}

type linearGradientOpData struct {
	stop1  f32.Point
	color1 color.NRGBA
	stop2  f32.Point
	color2 color.NRGBA
}

func decodeImageOp(data []byte, refs []interface{}) imageOpData {
	handle := refs[1]
	if handle == nil {
		return imageOpData{}
	}
	return imageOpData{
		src:    refs[0].(*image.RGBA),
		handle: handle,
	}
}

func decodeColorOp(data []byte) color.NRGBA {
	data = data[:ops.TypeColorLen]
	return color.NRGBA{
		R: data[1],
		G: data[2],
		B: data[3],
		A: data[4],
	}
}

func decodeLinearGradientOp(data []byte) linearGradientOpData {
	data = data[:ops.TypeLinearGradientLen]
	bo := binary.LittleEndian
	return linearGradientOpData{
		stop1: f32.Point{
			X: math.Float32frombits(bo.Uint32(data[1:])),
			Y: math.Float32frombits(bo.Uint32(data[5:])),
		},
		stop2: f32.Point{
			X: math.Float32frombits(bo.Uint32(data[9:])),
			Y: math.Float32frombits(bo.Uint32(data[13:])),
		},
		color1: color.NRGBA{
			R: data[17+0],
			G: data[17+1],
			B: data[17+2],
			A: data[17+3],
		},
		color2: color.NRGBA{
			R: data[21+0],
			G: data[21+1],
			B: data[21+2],
			A: data[21+3],
		},
	}
}

type clipType uint8

type resource interface {
	release()
}

type texture struct {
	src *image.RGBA
	tex driver.Texture
}

type blitter struct {
	ctx                    driver.Device
	viewport               image.Point
	pipelines              [3]*pipeline
	colUniforms            *blitColUniforms
	texUniforms            *blitTexUniforms
	linearGradientUniforms *blitLinearGradientUniforms
	quadVerts              driver.Buffer
}

type blitColUniforms struct {
	blitUniforms
	_ [128 - unsafe.Sizeof(blitUniforms{}) - unsafe.Sizeof(colorUniforms{})]byte // Padding to 128 bytes.
	colorUniforms
}

type blitTexUniforms struct {
	blitUniforms
}

type blitLinearGradientUniforms struct {
	blitUniforms
	_ [128 - unsafe.Sizeof(blitUniforms{}) - unsafe.Sizeof(gradientUniforms{})]byte // Padding to 128 bytes.
	gradientUniforms
}

type uniformBuffer struct {
	buf driver.Buffer
	ptr []byte
}

type pipeline struct {
	pipeline driver.Pipeline
	uniforms *uniformBuffer
}

type blitUniforms struct {
	transform     [4]float32
	uvTransformR1 [4]float32
	uvTransformR2 [4]float32
}

type colorUniforms struct {
	color f32color.RGBA
}

type gradientUniforms struct {
	color1 f32color.RGBA
	color2 f32color.RGBA
}

type materialType uint8

const (
	clipTypeNone clipType = iota
	clipTypePath
	clipTypeIntersection
)

const (
	materialColor materialType = iota
	materialLinearGradient
	materialTexture
)

// New creates a GPU for the given API.
func New(api API) (GPU, error) {
	d, err := driver.NewDevice(api)
	if err != nil {
		return nil, err
	}
	return NewWithDevice(d)
}

// NewWithDevice creates a GPU with a pre-existing device.
//
// Note: for internal use only.
func NewWithDevice(d driver.Device) (GPU, error) {
	d.BeginFrame(nil, false, image.Point{})
	defer d.EndFrame()
	forceCompute := os.Getenv("GIORENDERER") == "forcecompute"
	feats := d.Caps().Features
	switch {
	case !forceCompute && feats.Has(driver.FeatureFloatRenderTargets) && feats.Has(driver.FeatureSRGB):
		return newGPU(d)
	}
	return newCompute(d)
}

func newGPU(ctx driver.Device) (*gpu, error) {
	g := &gpu{
		cache: newResourceCache(),
	}
	g.drawOps.pathCache = newOpCache()
	if err := g.init(ctx); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *gpu) init(ctx driver.Device) error {
	g.ctx = ctx
	g.renderer = newRenderer(ctx)
	return nil
}

func (g *gpu) Clear(col color.NRGBA) {
	g.drawOps.clear = true
	g.drawOps.clearColor = f32color.LinearFromSRGB(col)
}

func (g *gpu) Release() {
	g.renderer.release()
	g.drawOps.pathCache.release()
	g.cache.release()
	if g.timers != nil {
		g.timers.Release()
	}
	g.ctx.Release()
}

func (g *gpu) Frame(frameOps *op.Ops, target RenderTarget, viewport image.Point) error {
	g.collect(viewport, frameOps)
	return g.frame(target)
}

func (g *gpu) collect(viewport image.Point, frameOps *op.Ops) {
	g.renderer.blitter.viewport = viewport
	g.renderer.pather.viewport = viewport
	g.drawOps.reset(viewport)
	g.drawOps.collect(frameOps, viewport)
	if g.drawOps.profile && g.timers == nil && g.ctx.Caps().Features.Has(driver.FeatureTimers) {
		g.frameStart = time.Now()
		g.timers = newTimers(g.ctx)
		g.stencilTimer = g.timers.newTimer()
		g.coverTimer = g.timers.newTimer()
		g.cleanupTimer = g.timers.newTimer()
	}
}

func (g *gpu) frame(target RenderTarget) error {
	viewport := g.renderer.blitter.viewport
	defFBO := g.ctx.BeginFrame(target, g.drawOps.clear, viewport)
	defer g.ctx.EndFrame()
	g.drawOps.buildPaths(g.ctx)
	for _, img := range g.drawOps.imageOps {
		expandPathOp(img.path, img.clip)
	}
	g.stencilTimer.begin()
	g.renderer.packStencils(&g.drawOps.pathOps)
	g.renderer.stencilClips(g.drawOps.pathCache, g.drawOps.pathOps)
	g.renderer.packIntersections(g.drawOps.imageOps)
	g.renderer.prepareIntersections(g.drawOps.imageOps)
	g.renderer.intersect(g.drawOps.imageOps)
	g.stencilTimer.end()
	g.coverTimer.begin()
	g.renderer.uploadImages(g.cache, g.drawOps.imageOps)
	g.renderer.prepareDrawOps(g.cache, g.drawOps.imageOps)
	d := driver.LoadDesc{
		ClearColor: g.drawOps.clearColor,
	}
	if g.drawOps.clear {
		g.drawOps.clear = false
		d.Action = driver.LoadActionClear
	}
	g.ctx.BeginRenderPass(defFBO, d)
	g.ctx.Viewport(0, 0, viewport.X, viewport.Y)
	g.renderer.drawOps(g.cache, g.drawOps.imageOps)
	g.coverTimer.end()
	g.ctx.EndRenderPass()
	g.cleanupTimer.begin()
	g.cache.frame()
	g.drawOps.pathCache.frame()
	g.cleanupTimer.end()
	if g.drawOps.profile && g.timers.ready() {
		st, covt, cleant := g.stencilTimer.Elapsed, g.coverTimer.Elapsed, g.cleanupTimer.Elapsed
		ft := st + covt + cleant
		q := 100 * time.Microsecond
		st, covt = st.Round(q), covt.Round(q)
		frameDur := time.Since(g.frameStart).Round(q)
		ft = ft.Round(q)
		g.profile = fmt.Sprintf("draw:%7s gpu:%7s st:%7s cov:%7s", frameDur, ft, st, covt)
	}
	return nil
}

func (g *gpu) Profile() string {
	return g.profile
}

func (r *renderer) texHandle(cache *resourceCache, data imageOpData) driver.Texture {
	var tex *texture
	t, exists := cache.get(data.handle)
	if !exists {
		t = &texture{
			src: data.src,
		}
		cache.put(data.handle, t)
	}
	tex = t.(*texture)
	if tex.tex != nil {
		return tex.tex
	}
	handle, err := r.ctx.NewTexture(driver.TextureFormatSRGBA, data.src.Bounds().Dx(), data.src.Bounds().Dy(), driver.FilterLinear, driver.FilterLinear, driver.BufferBindingTexture)
	if err != nil {
		panic(err)
	}
	driver.UploadImage(handle, image.Pt(0, 0), data.src)
	tex.tex = handle
	return tex.tex
}

func (t *texture) release() {
	if t.tex != nil {
		t.tex.Release()
	}
}

func newRenderer(ctx driver.Device) *renderer {
	r := &renderer{
		ctx:     ctx,
		blitter: newBlitter(ctx),
		pather:  newPather(ctx),
	}

	maxDim := ctx.Caps().MaxTextureSize
	// Large atlas textures cause artifacts due to precision loss in
	// shaders.
	if cap := 8192; maxDim > cap {
		maxDim = cap
	}

	r.packer.maxDims = image.Pt(maxDim, maxDim)
	r.intersections.maxDims = image.Pt(maxDim, maxDim)
	return r
}

func (r *renderer) release() {
	r.pather.release()
	r.blitter.release()
}

func newBlitter(ctx driver.Device) *blitter {
	quadVerts, err := ctx.NewImmutableBuffer(driver.BufferBindingVertices,
		byteslice.Slice([]float32{
			-1, -1, 0, 0,
			+1, -1, 1, 0,
			-1, +1, 0, 1,
			+1, +1, 1, 1,
		}),
	)
	if err != nil {
		panic(err)
	}
	b := &blitter{
		ctx:       ctx,
		quadVerts: quadVerts,
	}
	b.colUniforms = new(blitColUniforms)
	b.texUniforms = new(blitTexUniforms)
	b.linearGradientUniforms = new(blitLinearGradientUniforms)
	pipelines, err := createColorPrograms(ctx, gio.Shader_blit_vert, gio.Shader_blit_frag,
		[3]interface{}{b.colUniforms, b.linearGradientUniforms, b.texUniforms},
	)
	if err != nil {
		panic(err)
	}
	b.pipelines = pipelines
	return b
}

func (b *blitter) release() {
	b.quadVerts.Release()
	for _, p := range b.pipelines {
		p.Release()
	}
}

func createColorPrograms(b driver.Device, vsSrc shader.Sources, fsSrc [3]shader.Sources, uniforms [3]interface{}) ([3]*pipeline, error) {
	var pipelines [3]*pipeline
	blend := driver.BlendDesc{
		Enable:    true,
		SrcFactor: driver.BlendFactorOne,
		DstFactor: driver.BlendFactorOneMinusSrcAlpha,
	}
	layout := driver.VertexLayout{
		Inputs: []driver.InputDesc{
			{Type: shader.DataTypeFloat, Size: 2, Offset: 0},
			{Type: shader.DataTypeFloat, Size: 2, Offset: 4 * 2},
		},
		Stride: 4 * 4,
	}
	vsh, err := b.NewVertexShader(vsSrc)
	if err != nil {
		return pipelines, err
	}
	defer vsh.Release()
	{
		fsh, err := b.NewFragmentShader(fsSrc[materialTexture])
		if err != nil {
			return pipelines, err
		}
		defer fsh.Release()
		pipe, err := b.NewPipeline(driver.PipelineDesc{
			VertexShader:   vsh,
			FragmentShader: fsh,
			BlendDesc:      blend,
			VertexLayout:   layout,
			PixelFormat:    driver.TextureFormatOutput,
			Topology:       driver.TopologyTriangleStrip,
		})
		if err != nil {
			return pipelines, err
		}
		var vertBuffer *uniformBuffer
		if u := uniforms[materialTexture]; u != nil {
			vertBuffer = newUniformBuffer(b, u)
		}
		pipelines[materialTexture] = &pipeline{pipe, vertBuffer}
	}
	{
		var vertBuffer *uniformBuffer
		fsh, err := b.NewFragmentShader(fsSrc[materialColor])
		if err != nil {
			pipelines[materialTexture].Release()
			return pipelines, err
		}
		defer fsh.Release()
		pipe, err := b.NewPipeline(driver.PipelineDesc{
			VertexShader:   vsh,
			FragmentShader: fsh,
			BlendDesc:      blend,
			VertexLayout:   layout,
			PixelFormat:    driver.TextureFormatOutput,
			Topology:       driver.TopologyTriangleStrip,
		})
		if err != nil {
			pipelines[materialTexture].Release()
			return pipelines, err
		}
		if u := uniforms[materialColor]; u != nil {
			vertBuffer = newUniformBuffer(b, u)
		}
		pipelines[materialColor] = &pipeline{pipe, vertBuffer}
	}
	{
		var vertBuffer *uniformBuffer
		fsh, err := b.NewFragmentShader(fsSrc[materialLinearGradient])
		if err != nil {
			pipelines[materialTexture].Release()
			pipelines[materialColor].Release()
			return pipelines, err
		}
		defer fsh.Release()
		pipe, err := b.NewPipeline(driver.PipelineDesc{
			VertexShader:   vsh,
			FragmentShader: fsh,
			BlendDesc:      blend,
			VertexLayout:   layout,
			PixelFormat:    driver.TextureFormatOutput,
			Topology:       driver.TopologyTriangleStrip,
		})
		if err != nil {
			pipelines[materialTexture].Release()
			pipelines[materialColor].Release()
			return pipelines, err
		}
		if u := uniforms[materialLinearGradient]; u != nil {
			vertBuffer = newUniformBuffer(b, u)
		}
		pipelines[materialLinearGradient] = &pipeline{pipe, vertBuffer}
	}
	if err != nil {
		for _, p := range pipelines {
			p.Release()
		}
		return pipelines, err
	}
	return pipelines, nil
}

func (r *renderer) stencilClips(pathCache *opCache, ops []*pathOp) {
	if len(r.packer.sizes) == 0 {
		return
	}
	fbo := -1
	r.pather.begin(r.packer.sizes)
	for _, p := range ops {
		if fbo != p.place.Idx {
			if fbo != -1 {
				r.ctx.EndRenderPass()
			}
			fbo = p.place.Idx
			f := r.pather.stenciler.cover(fbo)
			r.ctx.BeginRenderPass(f.tex, driver.LoadDesc{Action: driver.LoadActionClear})
			r.ctx.BindPipeline(r.pather.stenciler.pipeline.pipeline.pipeline)
			r.ctx.BindIndexBuffer(r.pather.stenciler.indexBuf)
		}
		v, _ := pathCache.get(p.pathKey)
		r.pather.stencilPath(p.clip, p.off, p.place.Pos, v.data)
	}
	if fbo != -1 {
		r.ctx.EndRenderPass()
	}
}

func (r *renderer) prepareIntersections(ops []imageOp) {
	for _, img := range ops {
		if img.clipType != clipTypeIntersection {
			continue
		}
		fbo := r.pather.stenciler.cover(img.path.place.Idx)
		r.ctx.PrepareTexture(fbo.tex)
	}
}

func (r *renderer) intersect(ops []imageOp) {
	if len(r.intersections.sizes) == 0 {
		return
	}
	fbo := -1
	r.pather.stenciler.beginIntersect(r.intersections.sizes)
	for _, img := range ops {
		if img.clipType != clipTypeIntersection {
			continue
		}
		if fbo != img.place.Idx {
			if fbo != -1 {
				r.ctx.EndRenderPass()
			}
			fbo = img.place.Idx
			f := r.pather.stenciler.intersections.fbos[fbo]
			d := driver.LoadDesc{Action: driver.LoadActionClear}
			d.ClearColor.R = 1.0
			r.ctx.BeginRenderPass(f.tex, d)
			r.ctx.BindPipeline(r.pather.stenciler.ipipeline.pipeline.pipeline)
			r.ctx.BindVertexBuffer(r.blitter.quadVerts, 0)
		}
		r.ctx.Viewport(img.place.Pos.X, img.place.Pos.Y, img.clip.Dx(), img.clip.Dy())
		r.intersectPath(img.path, img.clip)
	}
	if fbo != -1 {
		r.ctx.EndRenderPass()
	}
}

func (r *renderer) intersectPath(p *pathOp, clip image.Rectangle) {
	if p.parent != nil {
		r.intersectPath(p.parent, clip)
	}
	if !p.path {
		return
	}
	uv := image.Rectangle{
		Min: p.place.Pos,
		Max: p.place.Pos.Add(p.clip.Size()),
	}
	o := clip.Min.Sub(p.clip.Min)
	sub := image.Rectangle{
		Min: o,
		Max: o.Add(clip.Size()),
	}
	fbo := r.pather.stenciler.cover(p.place.Idx)
	r.ctx.BindTexture(0, fbo.tex)
	coverScale, coverOff := texSpaceTransform(f32.FRect(uv), fbo.size)
	subScale, subOff := texSpaceTransform(f32.FRect(sub), p.clip.Size())
	r.pather.stenciler.ipipeline.uniforms.vert.uvTransform = [4]float32{coverScale.X, coverScale.Y, coverOff.X, coverOff.Y}
	r.pather.stenciler.ipipeline.uniforms.vert.subUVTransform = [4]float32{subScale.X, subScale.Y, subOff.X, subOff.Y}
	r.pather.stenciler.ipipeline.pipeline.UploadUniforms(r.ctx)
	r.ctx.DrawArrays(0, 4)
}

func (r *renderer) packIntersections(ops []imageOp) {
	r.intersections.clear()
	for i, img := range ops {
		var npaths int
		var onePath *pathOp
		for p := img.path; p != nil; p = p.parent {
			if p.path {
				onePath = p
				npaths++
			}
		}
		switch npaths {
		case 0:
		case 1:
			place := onePath.place
			place.Pos = place.Pos.Sub(onePath.clip.Min).Add(img.clip.Min)
			ops[i].place = place
			ops[i].clipType = clipTypePath
		default:
			sz := image.Point{X: img.clip.Dx(), Y: img.clip.Dy()}
			place, ok := r.intersections.add(sz)
			if !ok {
				panic("internal error: if the intersection fit, the intersection should fit as well")
			}
			ops[i].clipType = clipTypeIntersection
			ops[i].place = place
		}
	}
}

func (r *renderer) packStencils(pops *[]*pathOp) {
	r.packer.clear()
	ops := *pops
	// Allocate atlas space for cover textures.
	var i int
	for i < len(ops) {
		p := ops[i]
		if p.clip.Empty() {
			ops[i] = ops[len(ops)-1]
			ops = ops[:len(ops)-1]
			continue
		}
		sz := image.Point{X: p.clip.Dx(), Y: p.clip.Dy()}
		place, ok := r.packer.add(sz)
		if !ok {
			// The clip area is at most the entire screen. Hopefully no
			// screen is larger than GL_MAX_TEXTURE_SIZE.
			panic(fmt.Errorf("clip area %v is larger than maximum texture size %v", p.clip, r.packer.maxDims))
		}
		p.place = place
		i++
	}
	*pops = ops
}

func (d *drawOps) reset(viewport image.Point) {
	d.profile = false
	d.viewport = viewport
	d.imageOps = d.imageOps[:0]
	d.pathOps = d.pathOps[:0]
	d.pathOpCache = d.pathOpCache[:0]
	d.vertCache = d.vertCache[:0]
	d.transStack = d.transStack[:0]
}

func (d *drawOps) collect(root *op.Ops, viewport image.Point) {
	viewf := f32.Rectangle{
		Max: f32.Point{X: float32(viewport.X), Y: float32(viewport.Y)},
	}
	var ops *ops.Ops
	if root != nil {
		ops = &root.Internal
	}
	d.reader.Reset(ops)
	d.collectOps(&d.reader, viewf)
}

func (d *drawOps) buildPaths(ctx driver.Device) {
	for _, p := range d.pathOps {
		if v, exists := d.pathCache.get(p.pathKey); !exists || v.data.data == nil {
			data := buildPath(ctx, p.pathVerts)
			d.pathCache.put(p.pathKey, opCacheValue{
				data:   data,
				bounds: p.bounds,
			})
		}
		p.pathVerts = nil
	}
}

func (d *drawOps) newPathOp() *pathOp {
	d.pathOpCache = append(d.pathOpCache, pathOp{})
	return &d.pathOpCache[len(d.pathOpCache)-1]
}

func (d *drawOps) addClipPath(state *drawState, aux []byte, auxKey opKey, bounds f32.Rectangle, off f32.Point, push bool) {
	npath := d.newPathOp()
	*npath = pathOp{
		parent:    state.cpath,
		bounds:    bounds,
		off:       off,
		intersect: bounds.Add(off),
		rect:      true,
	}
	if npath.parent != nil {
		npath.rect = npath.parent.rect
		npath.intersect = npath.parent.intersect.Intersect(npath.intersect)
	}
	if len(aux) > 0 {
		npath.rect = false
		npath.pathKey = auxKey
		npath.path = true
		npath.pathVerts = aux
		d.pathOps = append(d.pathOps, npath)
	}
	state.cpath = npath
}

func (d *drawOps) save(id int, state f32.Affine2D) {
	if extra := id - len(d.states) + 1; extra > 0 {
		d.states = append(d.states, make([]f32.Affine2D, extra)...)
	}
	d.states[id] = state
}

func (k opKey) SetTransform(t f32.Affine2D) opKey {
	sx, hx, _, hy, sy, _ := t.Elems()
	k.sx = sx
	k.hx = hx
	k.hy = hy
	k.sy = sy
	return k
}

func (d *drawOps) collectOps(r *ops.Reader, viewport f32.Rectangle) {
	var (
		quads quadsOp
		state drawState
	)
	reset := func() {
		state = drawState{
			color: color.NRGBA{A: 0xff},
		}
	}
	reset()
loop:
	for encOp, ok := r.Decode(); ok; encOp, ok = r.Decode() {
		switch ops.OpType(encOp.Data[0]) {
		case ops.TypeProfile:
			d.profile = true
		case ops.TypeTransform:
			dop, push := ops.DecodeTransform(encOp.Data)
			if push {
				d.transStack = append(d.transStack, state.t)
			}
			state.t = state.t.Mul(dop)
		case ops.TypePopTransform:
			n := len(d.transStack)
			state.t = d.transStack[n-1]
			d.transStack = d.transStack[:n-1]

		case ops.TypeStroke:
			quads.key.strokeWidth = decodeStrokeOp(encOp.Data)

		case ops.TypePath:
			encOp, ok = r.Decode()
			if !ok {
				break loop
			}
			quads.aux = encOp.Data[ops.TypeAuxLen:]
			quads.key.Key = encOp.Key

		case ops.TypeClip:
			var op ops.ClipOp
			op.Decode(encOp.Data)
			quads.key.outline = op.Outline
			bounds := f32.FRect(op.Bounds)
			trans, off := state.t.Split()
			if len(quads.aux) > 0 {
				// There is a clipping path, build the gpu data and update the
				// cache key such that it will be equal only if the transform is the
				// same also. Use cached data if we have it.
				quads.key = quads.key.SetTransform(trans)
				if v, ok := d.pathCache.get(quads.key); ok {
					// Since the GPU data exists in the cache aux will not be used.
					// Why is this not used for the offset shapes?
					bounds = v.bounds
				} else {
					var pathData []byte
					pathData, bounds = d.buildVerts(
						quads.aux, trans, quads.key.outline, quads.key.strokeWidth,
					)
					quads.aux = pathData
					// add it to the cache, without GPU data, so the transform can be
					// reused.
					d.pathCache.put(quads.key, opCacheValue{bounds: bounds})
				}
			} else {
				quads.aux, bounds, _ = d.boundsForTransformedRect(bounds, trans)
				quads.key = opKey{Key: encOp.Key}
			}
			d.addClipPath(&state, quads.aux, quads.key, bounds, off, true)
			quads = quadsOp{}
		case ops.TypePopClip:
			state.cpath = state.cpath.parent

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
			// Transform (if needed) the painting rectangle and if so generate a clip path,
			// for those cases also compute a partialTrans that maps texture coordinates between
			// the new bounding rectangle and the transformed original paint rectangle.
			t, off := state.t.Split()
			// Fill the clip area, unless the material is a (bounded) image.
			// TODO: Find a tighter bound.
			inf := float32(1e6)
			dst := f32.Rect(-inf, -inf, inf, inf)
			if state.matType == materialTexture {
				sz := state.image.src.Rect.Size()
				dst = f32.Rectangle{Max: layout.FPt(sz)}
			}
			clipData, bnd, partialTrans := d.boundsForTransformedRect(dst, t)
			cl := viewport.Intersect(bnd.Add(off))
			if state.cpath != nil {
				cl = state.cpath.intersect.Intersect(cl)
			}
			if cl.Empty() {
				continue
			}

			if clipData != nil {
				// The paint operation is sheared or rotated, add a clip path representing
				// this transformed rectangle.
				k := opKey{Key: encOp.Key}
				k.SetTransform(t) // TODO: This call has no effect.
				d.addClipPath(&state, clipData, k, bnd, off, false)
			}

			bounds := cl.Round()
			mat := state.materialFor(bnd, off, partialTrans, bounds)

			rect := state.cpath == nil || state.cpath.rect
			if bounds.Min == (image.Point{}) && bounds.Max == d.viewport && rect && mat.opaque && (mat.material == materialColor) {
				// The image is a uniform opaque color and takes up the whole screen.
				// Scrap images up to and including this image and set clear color.
				d.imageOps = d.imageOps[:0]
				d.clearColor = mat.color.Opaque()
				d.clear = true
				continue
			}
			img := imageOp{
				path:     state.cpath,
				clip:     bounds,
				material: mat,
			}

			d.imageOps = append(d.imageOps, img)
			if clipData != nil {
				// we added a clip path that should not remain
				state.cpath = state.cpath.parent
			}
		case ops.TypeSave:
			id := ops.DecodeSave(encOp.Data)
			d.save(id, state.t)
		case ops.TypeLoad:
			reset()
			id := ops.DecodeLoad(encOp.Data)
			state.t = d.states[id]
		}
	}
}

func expandPathOp(p *pathOp, clip image.Rectangle) {
	for p != nil {
		pclip := p.clip
		if !pclip.Empty() {
			clip = clip.Union(pclip)
		}
		p.clip = clip
		p = p.parent
	}
}

func (d *drawState) materialFor(rect f32.Rectangle, off f32.Point, partTrans f32.Affine2D, clip image.Rectangle) material {
	var m material
	switch d.matType {
	case materialColor:
		m.material = materialColor
		m.color = f32color.LinearFromSRGB(d.color)
		m.opaque = m.color.A == 1.0
	case materialLinearGradient:
		m.material = materialLinearGradient

		m.color1 = f32color.LinearFromSRGB(d.color1)
		m.color2 = f32color.LinearFromSRGB(d.color2)
		m.opaque = m.color1.A == 1.0 && m.color2.A == 1.0

		m.uvTrans = partTrans.Mul(gradientSpaceTransform(clip, off, d.stop1, d.stop2))
	case materialTexture:
		m.material = materialTexture
		dr := rect.Add(off).Round()
		sz := d.image.src.Bounds().Size()
		sr := f32.Rectangle{
			Max: f32.Point{
				X: float32(sz.X),
				Y: float32(sz.Y),
			},
		}
		dx := float32(dr.Dx())
		sdx := sr.Dx()
		sr.Min.X += float32(clip.Min.X-dr.Min.X) * sdx / dx
		sr.Max.X -= float32(dr.Max.X-clip.Max.X) * sdx / dx
		dy := float32(dr.Dy())
		sdy := sr.Dy()
		sr.Min.Y += float32(clip.Min.Y-dr.Min.Y) * sdy / dy
		sr.Max.Y -= float32(dr.Max.Y-clip.Max.Y) * sdy / dy
		uvScale, uvOffset := texSpaceTransform(sr, sz)
		m.uvTrans = partTrans.Mul(f32.Affine2D{}.Scale(f32.Point{}, uvScale).Offset(uvOffset))
		m.data = d.image
	}
	return m
}

func (r *renderer) uploadImages(cache *resourceCache, ops []imageOp) {
	for _, img := range ops {
		m := img.material
		if m.material == materialTexture {
			r.texHandle(cache, m.data)
		}
	}
}

func (r *renderer) prepareDrawOps(cache *resourceCache, ops []imageOp) {
	for _, img := range ops {
		m := img.material
		switch m.material {
		case materialTexture:
			r.ctx.PrepareTexture(r.texHandle(cache, m.data))
		}

		var fbo stencilFBO
		switch img.clipType {
		case clipTypeNone:
			continue
		case clipTypePath:
			fbo = r.pather.stenciler.cover(img.place.Idx)
		case clipTypeIntersection:
			fbo = r.pather.stenciler.intersections.fbos[img.place.Idx]
		}
		r.ctx.PrepareTexture(fbo.tex)
	}
}

func (r *renderer) drawOps(cache *resourceCache, ops []imageOp) {
	var coverTex driver.Texture
	for _, img := range ops {
		m := img.material
		switch m.material {
		case materialTexture:
			r.ctx.BindTexture(0, r.texHandle(cache, m.data))
		}
		drc := img.clip

		scale, off := clipSpaceTransform(drc, r.blitter.viewport)
		var fbo stencilFBO
		switch img.clipType {
		case clipTypeNone:
			p := r.blitter.pipelines[m.material]
			r.ctx.BindPipeline(p.pipeline)
			r.ctx.BindVertexBuffer(r.blitter.quadVerts, 0)
			r.blitter.blit(m.material, m.color, m.color1, m.color2, scale, off, m.uvTrans)
			continue
		case clipTypePath:
			fbo = r.pather.stenciler.cover(img.place.Idx)
		case clipTypeIntersection:
			fbo = r.pather.stenciler.intersections.fbos[img.place.Idx]
		}
		if coverTex != fbo.tex {
			coverTex = fbo.tex
			r.ctx.BindTexture(1, coverTex)
		}
		uv := image.Rectangle{
			Min: img.place.Pos,
			Max: img.place.Pos.Add(drc.Size()),
		}
		coverScale, coverOff := texSpaceTransform(f32.FRect(uv), fbo.size)
		p := r.pather.coverer.pipelines[m.material]
		r.ctx.BindPipeline(p.pipeline)
		r.ctx.BindVertexBuffer(r.blitter.quadVerts, 0)
		r.pather.cover(m.material, m.color, m.color1, m.color2, scale, off, m.uvTrans, coverScale, coverOff)
	}
}

func (b *blitter) blit(mat materialType, col f32color.RGBA, col1, col2 f32color.RGBA, scale, off f32.Point, uvTrans f32.Affine2D) {
	p := b.pipelines[mat]
	b.ctx.BindPipeline(p.pipeline)
	var uniforms *blitUniforms
	switch mat {
	case materialColor:
		b.colUniforms.color = col
		uniforms = &b.colUniforms.blitUniforms
	case materialTexture:
		t1, t2, t3, t4, t5, t6 := uvTrans.Elems()
		b.texUniforms.blitUniforms.uvTransformR1 = [4]float32{t1, t2, t3, 0}
		b.texUniforms.blitUniforms.uvTransformR2 = [4]float32{t4, t5, t6, 0}
		uniforms = &b.texUniforms.blitUniforms
	case materialLinearGradient:
		b.linearGradientUniforms.color1 = col1
		b.linearGradientUniforms.color2 = col2

		t1, t2, t3, t4, t5, t6 := uvTrans.Elems()
		b.linearGradientUniforms.blitUniforms.uvTransformR1 = [4]float32{t1, t2, t3, 0}
		b.linearGradientUniforms.blitUniforms.uvTransformR2 = [4]float32{t4, t5, t6, 0}
		uniforms = &b.linearGradientUniforms.blitUniforms
	}
	uniforms.transform = [4]float32{scale.X, scale.Y, off.X, off.Y}
	p.UploadUniforms(b.ctx)
	b.ctx.DrawArrays(0, 4)
}

// newUniformBuffer creates a new GPU uniform buffer backed by the
// structure uniformBlock points to.
func newUniformBuffer(b driver.Device, uniformBlock interface{}) *uniformBuffer {
	ref := reflect.ValueOf(uniformBlock)
	// Determine the size of the uniforms structure, *uniforms.
	size := ref.Elem().Type().Size()
	// Map the uniforms structure as a byte slice.
	ptr := unsafe.Slice((*byte)(unsafe.Pointer(ref.Pointer())), size)
	ubuf, err := b.NewBuffer(driver.BufferBindingUniforms, len(ptr))
	if err != nil {
		panic(err)
	}
	return &uniformBuffer{buf: ubuf, ptr: ptr}
}

func (u *uniformBuffer) Upload() {
	u.buf.Upload(u.ptr)
}

func (u *uniformBuffer) Release() {
	u.buf.Release()
	u.buf = nil
}

func (p *pipeline) UploadUniforms(ctx driver.Device) {
	if p.uniforms != nil {
		p.uniforms.Upload()
		ctx.BindUniforms(p.uniforms.buf)
	}
}

func (p *pipeline) Release() {
	p.pipeline.Release()
	if p.uniforms != nil {
		p.uniforms.Release()
	}
	*p = pipeline{}
}

// texSpaceTransform return the scale and offset that transforms the given subimage
// into quad texture coordinates.
func texSpaceTransform(r f32.Rectangle, bounds image.Point) (f32.Point, f32.Point) {
	size := f32.Point{X: float32(bounds.X), Y: float32(bounds.Y)}
	scale := f32.Point{X: r.Dx() / size.X, Y: r.Dy() / size.Y}
	offset := f32.Point{X: r.Min.X / size.X, Y: r.Min.Y / size.Y}
	return scale, offset
}

// gradientSpaceTransform transforms stop1 and stop2 to [(0,0), (1,1)].
func gradientSpaceTransform(clip image.Rectangle, off f32.Point, stop1, stop2 f32.Point) f32.Affine2D {
	d := stop2.Sub(stop1)
	l := float32(math.Sqrt(float64(d.X*d.X + d.Y*d.Y)))
	a := float32(math.Atan2(float64(-d.Y), float64(d.X)))

	// TODO: optimize
	zp := f32.Point{}
	return f32.Affine2D{}.
		Scale(zp, layout.FPt(clip.Size())).            // scale to pixel space
		Offset(zp.Sub(off).Add(layout.FPt(clip.Min))). // offset to clip space
		Offset(zp.Sub(stop1)).                         // offset to first stop point
		Rotate(zp, a).                                 // rotate to align gradient
		Scale(zp, f32.Pt(1/l, 1/l))                    // scale gradient to right size
}

// clipSpaceTransform returns the scale and offset that transforms the given
// rectangle from a viewport into GPU driver device coordinates.
func clipSpaceTransform(r image.Rectangle, viewport image.Point) (f32.Point, f32.Point) {
	// First, transform UI coordinates to device coordinates:
	//
	//	[(-1, -1) (+1, -1)]
	//	[(-1, +1) (+1, +1)]
	//
	x, y := float32(r.Min.X), float32(r.Min.Y)
	w, h := float32(r.Dx()), float32(r.Dy())
	vx, vy := 2/float32(viewport.X), 2/float32(viewport.Y)
	x = x*vx - 1
	y = y*vy - 1
	w *= vx
	h *= vy

	// Then, compute the transformation from the fullscreen quad to
	// the rectangle at (x, y) and dimensions (w, h).
	scale := f32.Point{X: w * .5, Y: h * .5}
	offset := f32.Point{X: x + w*.5, Y: y + h*.5}

	return scale, offset
}

// Fill in maximal Y coordinates of the NW and NE corners.
func fillMaxY(verts []byte) {
	contour := 0
	bo := binary.LittleEndian
	for len(verts) > 0 {
		maxy := float32(math.Inf(-1))
		i := 0
		for ; i+vertStride*4 <= len(verts); i += vertStride * 4 {
			vert := verts[i : i+vertStride]
			// MaxY contains the integer contour index.
			pathContour := int(bo.Uint32(vert[int(unsafe.Offsetof(((*vertex)(nil)).MaxY)):]))
			if contour != pathContour {
				contour = pathContour
				break
			}
			fromy := math.Float32frombits(bo.Uint32(vert[int(unsafe.Offsetof(((*vertex)(nil)).FromY)):]))
			ctrly := math.Float32frombits(bo.Uint32(vert[int(unsafe.Offsetof(((*vertex)(nil)).CtrlY)):]))
			toy := math.Float32frombits(bo.Uint32(vert[int(unsafe.Offsetof(((*vertex)(nil)).ToY)):]))
			if fromy > maxy {
				maxy = fromy
			}
			if ctrly > maxy {
				maxy = ctrly
			}
			if toy > maxy {
				maxy = toy
			}
		}
		fillContourMaxY(maxy, verts[:i])
		verts = verts[i:]
	}
}

func fillContourMaxY(maxy float32, verts []byte) {
	bo := binary.LittleEndian
	for i := 0; i < len(verts); i += vertStride {
		off := int(unsafe.Offsetof(((*vertex)(nil)).MaxY))
		bo.PutUint32(verts[i+off:], math.Float32bits(maxy))
	}
}

func (d *drawOps) writeVertCache(n int) []byte {
	d.vertCache = append(d.vertCache, make([]byte, n)...)
	return d.vertCache[len(d.vertCache)-n:]
}

// transform, split paths as needed, calculate maxY, bounds and create GPU vertices.
func (d *drawOps) buildVerts(pathData []byte, tr f32.Affine2D, outline bool, strWidth float32) (verts []byte, bounds f32.Rectangle) {
	inf := float32(math.Inf(+1))
	d.qs.bounds = f32.Rectangle{
		Min: f32.Point{X: inf, Y: inf},
		Max: f32.Point{X: -inf, Y: -inf},
	}
	d.qs.d = d
	startLength := len(d.vertCache)

	switch {
	case strWidth > 0:
		// Stroke path.
		ss := stroke.StrokeStyle{
			Width: strWidth,
		}
		quads := stroke.StrokePathCommands(ss, pathData)
		for _, quad := range quads {
			d.qs.contour = quad.Contour
			quad.Quad = quad.Quad.Transform(tr)

			d.qs.splitAndEncode(quad.Quad)
		}

	case outline:
		decodeToOutlineQuads(&d.qs, tr, pathData)
	}

	fillMaxY(d.vertCache[startLength:])
	return d.vertCache[startLength:], d.qs.bounds
}

// decodeOutlineQuads decodes scene commands, splits them into quadratic béziers
// as needed and feeds them to the supplied splitter.
func decodeToOutlineQuads(qs *quadSplitter, tr f32.Affine2D, pathData []byte) {
	for len(pathData) >= scene.CommandSize+4 {
		qs.contour = bo.Uint32(pathData)
		cmd := ops.DecodeCommand(pathData[4:])
		switch cmd.Op() {
		case scene.OpLine:
			var q stroke.QuadSegment
			q.From, q.To = scene.DecodeLine(cmd)
			q.Ctrl = q.From.Add(q.To).Mul(.5)
			q = q.Transform(tr)
			qs.splitAndEncode(q)
		case scene.OpGap:
			var q stroke.QuadSegment
			q.From, q.To = scene.DecodeGap(cmd)
			q.Ctrl = q.From.Add(q.To).Mul(.5)
			q = q.Transform(tr)
			qs.splitAndEncode(q)
		case scene.OpQuad:
			var q stroke.QuadSegment
			q.From, q.Ctrl, q.To = scene.DecodeQuad(cmd)
			q = q.Transform(tr)
			qs.splitAndEncode(q)
		case scene.OpCubic:
			for _, q := range stroke.SplitCubic(scene.DecodeCubic(cmd)) {
				q = q.Transform(tr)
				qs.splitAndEncode(q)
			}
		default:
			panic("unsupported scene command")
		}
		pathData = pathData[scene.CommandSize+4:]
	}
}

// create GPU vertices for transformed r, find the bounds and establish texture transform.
func (d *drawOps) boundsForTransformedRect(r f32.Rectangle, tr f32.Affine2D) (aux []byte, bnd f32.Rectangle, ptr f32.Affine2D) {
	if isPureOffset(tr) {
		// fast-path to allow blitting of pure rectangles
		_, _, ox, _, _, oy := tr.Elems()
		off := f32.Pt(ox, oy)
		bnd.Min = r.Min.Add(off)
		bnd.Max = r.Max.Add(off)
		return
	}

	// transform all corners, find new bounds
	corners := [4]f32.Point{
		tr.Transform(r.Min), tr.Transform(f32.Pt(r.Max.X, r.Min.Y)),
		tr.Transform(r.Max), tr.Transform(f32.Pt(r.Min.X, r.Max.Y)),
	}
	bnd.Min = f32.Pt(math.MaxFloat32, math.MaxFloat32)
	bnd.Max = f32.Pt(-math.MaxFloat32, -math.MaxFloat32)
	for _, c := range corners {
		if c.X < bnd.Min.X {
			bnd.Min.X = c.X
		}
		if c.Y < bnd.Min.Y {
			bnd.Min.Y = c.Y
		}
		if c.X > bnd.Max.X {
			bnd.Max.X = c.X
		}
		if c.Y > bnd.Max.Y {
			bnd.Max.Y = c.Y
		}
	}

	// build the GPU vertices
	l := len(d.vertCache)
	d.vertCache = append(d.vertCache, make([]byte, vertStride*4*4)...)
	aux = d.vertCache[l:]
	encodeQuadTo(aux, 0, corners[0], corners[0].Add(corners[1]).Mul(0.5), corners[1])
	encodeQuadTo(aux[vertStride*4:], 0, corners[1], corners[1].Add(corners[2]).Mul(0.5), corners[2])
	encodeQuadTo(aux[vertStride*4*2:], 0, corners[2], corners[2].Add(corners[3]).Mul(0.5), corners[3])
	encodeQuadTo(aux[vertStride*4*3:], 0, corners[3], corners[3].Add(corners[0]).Mul(0.5), corners[0])
	fillMaxY(aux)

	// establish the transform mapping from bounds rectangle to transformed corners
	var P1, P2, P3 f32.Point
	P1.X = (corners[1].X - bnd.Min.X) / (bnd.Max.X - bnd.Min.X)
	P1.Y = (corners[1].Y - bnd.Min.Y) / (bnd.Max.Y - bnd.Min.Y)
	P2.X = (corners[2].X - bnd.Min.X) / (bnd.Max.X - bnd.Min.X)
	P2.Y = (corners[2].Y - bnd.Min.Y) / (bnd.Max.Y - bnd.Min.Y)
	P3.X = (corners[3].X - bnd.Min.X) / (bnd.Max.X - bnd.Min.X)
	P3.Y = (corners[3].Y - bnd.Min.Y) / (bnd.Max.Y - bnd.Min.Y)
	sx, sy := P2.X-P3.X, P2.Y-P3.Y
	ptr = f32.NewAffine2D(sx, P2.X-P1.X, P1.X-sx, sy, P2.Y-P1.Y, P1.Y-sy).Invert()

	return
}

func isPureOffset(t f32.Affine2D) bool {
	a, b, _, d, e, _ := t.Elems()
	return a == 1 && b == 0 && d == 0 && e == 1
}
