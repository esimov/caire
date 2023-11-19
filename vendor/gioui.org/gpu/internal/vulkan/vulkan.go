// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux || freebsd) && !novulkan
// +build linux freebsd
// +build !novulkan

package vulkan

import (
	"errors"
	"fmt"
	"image"
	"math/bits"

	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/vk"
	"gioui.org/shader"
)

type Backend struct {
	physDev vk.PhysicalDevice
	dev     vk.Device
	queue   vk.Queue
	cmdPool struct {
		current vk.CommandBuffer
		pool    vk.CommandPool
		used    int
		buffers []vk.CommandBuffer
	}
	outFormat vk.Format
	staging   struct {
		buf  *Buffer
		mem  []byte
		size int
		cap  int
	}
	defers     []func(d vk.Device)
	frameSig   vk.Semaphore
	frameFence vk.Fence
	waitSems   []vk.Semaphore
	waitStages []vk.PipelineStageFlags
	sigSems    []vk.Semaphore
	fence      vk.Fence

	allPipes []*Pipeline

	pipe *Pipeline

	passes map[passKey]vk.RenderPass

	// bindings and offset are temporary storage for BindVertexBuffer.
	bindings []vk.Buffer
	offsets  []vk.DeviceSize

	desc struct {
		dirty    bool
		texBinds [texUnits]*Texture
		bufBinds [storageUnits]*Buffer
	}

	caps driver.Features
}

type passKey struct {
	fmt         vk.Format
	loadAct     vk.AttachmentLoadOp
	initLayout  vk.ImageLayout
	finalLayout vk.ImageLayout
}

type Texture struct {
	backend    *Backend
	img        vk.Image
	mem        vk.DeviceMemory
	view       vk.ImageView
	sampler    vk.Sampler
	fbo        vk.Framebuffer
	format     vk.Format
	mipmaps    int
	layout     vk.ImageLayout
	passLayout vk.ImageLayout
	width      int
	height     int
	acquire    vk.Semaphore
	foreign    bool

	scope struct {
		stage  vk.PipelineStageFlags
		access vk.AccessFlags
	}
}

type Shader struct {
	dev       vk.Device
	module    vk.ShaderModule
	pushRange vk.PushConstantRange
	src       shader.Sources
}

type Pipeline struct {
	backend    *Backend
	pipe       vk.Pipeline
	pushRanges []vk.PushConstantRange
	ninputs    int
	desc       *descPool
}

type descPool struct {
	layout     vk.PipelineLayout
	descLayout vk.DescriptorSetLayout
	pool       vk.DescriptorPool
	sets       []vk.DescriptorSet
	size       int
	texBinds   []int
	imgBinds   []int
	bufBinds   []int
}

type Buffer struct {
	backend *Backend
	buf     vk.Buffer
	store   []byte
	mem     vk.DeviceMemory
	usage   vk.BufferUsageFlags

	scope struct {
		stage  vk.PipelineStageFlags
		access vk.AccessFlags
	}
}

const (
	texUnits     = 4
	storageUnits = 4
)

func init() {
	driver.NewVulkanDevice = newVulkanDevice
}

func newVulkanDevice(api driver.Vulkan) (driver.Device, error) {
	b := &Backend{
		physDev:   vk.PhysicalDevice(api.PhysDevice),
		dev:       vk.Device(api.Device),
		outFormat: vk.Format(api.Format),
		caps:      driver.FeatureCompute,
		passes:    make(map[passKey]vk.RenderPass),
	}
	b.queue = vk.GetDeviceQueue(b.dev, api.QueueFamily, api.QueueIndex)
	cmdPool, err := vk.CreateCommandPool(b.dev, api.QueueFamily)
	if err != nil {
		return nil, err
	}
	b.cmdPool.pool = cmdPool
	props := vk.GetPhysicalDeviceFormatProperties(b.physDev, vk.FORMAT_R16_SFLOAT)
	reqs := vk.FORMAT_FEATURE_COLOR_ATTACHMENT_BIT | vk.FORMAT_FEATURE_SAMPLED_IMAGE_BIT
	if props&reqs == reqs {
		b.caps |= driver.FeatureFloatRenderTargets
	}
	reqs = vk.FORMAT_FEATURE_COLOR_ATTACHMENT_BLEND_BIT | vk.FORMAT_FEATURE_SAMPLED_IMAGE_BIT | vk.FORMAT_FEATURE_SAMPLED_IMAGE_FILTER_LINEAR_BIT
	props = vk.GetPhysicalDeviceFormatProperties(b.physDev, vk.FORMAT_R8G8B8A8_SRGB)
	if props&reqs == reqs {
		b.caps |= driver.FeatureSRGB
	}
	fence, err := vk.CreateFence(b.dev, 0)
	if err != nil {
		return nil, mapErr(err)
	}
	b.fence = fence
	return b, nil
}

func (b *Backend) BeginFrame(target driver.RenderTarget, clear bool, viewport image.Point) driver.Texture {
	b.staging.size = 0
	b.cmdPool.used = 0
	b.runDefers()
	b.resetPipes()

	if target == nil {
		return nil
	}
	switch t := target.(type) {
	case driver.VulkanRenderTarget:
		layout := vk.IMAGE_LAYOUT_UNDEFINED
		if !clear {
			layout = vk.IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL
		}
		b.frameSig = vk.Semaphore(t.SignalSem)
		b.frameFence = vk.Fence(t.Fence)
		tex := &Texture{
			img:        vk.Image(t.Image),
			fbo:        vk.Framebuffer(t.Framebuffer),
			width:      viewport.X,
			height:     viewport.Y,
			layout:     layout,
			passLayout: vk.IMAGE_LAYOUT_PRESENT_SRC_KHR,
			format:     b.outFormat,
			acquire:    vk.Semaphore(t.WaitSem),
			foreign:    true,
		}
		return tex
	case *Texture:
		return t
	default:
		panic(fmt.Sprintf("vulkan: unsupported render target type: %T", t))
	}
}

func (b *Backend) deferFunc(f func(d vk.Device)) {
	b.defers = append(b.defers, f)
}

func (b *Backend) runDefers() {
	for _, f := range b.defers {
		f(b.dev)
	}
	b.defers = b.defers[:0]
}

func (b *Backend) resetPipes() {
	for i := len(b.allPipes) - 1; i >= 0; i-- {
		p := b.allPipes[i]
		if p.pipe == 0 {
			// Released pipeline.
			b.allPipes = append(b.allPipes[:i], b.allPipes[:i+1]...)
			continue
		}
		if p.desc.size > 0 {
			p.desc.size = 0
		}
	}
}

func (b *Backend) EndFrame() {
	if b.frameSig != 0 {
		b.sigSems = append(b.sigSems, b.frameSig)
		b.frameSig = 0
	}
	fence := b.frameFence
	if fence == 0 {
		// We're internally synchronized.
		fence = b.fence
	}
	b.submitCmdBuf(fence)
	if b.frameFence == 0 {
		vk.WaitForFences(b.dev, fence)
		vk.ResetFences(b.dev, fence)
	}
}

func (b *Backend) Caps() driver.Caps {
	return driver.Caps{
		MaxTextureSize: 4096,
		Features:       b.caps,
	}
}

func (b *Backend) NewTimer() driver.Timer {
	panic("timers not supported")
}

func (b *Backend) IsTimeContinuous() bool {
	panic("timers not supported")
}

func (b *Backend) Release() {
	vk.DeviceWaitIdle(b.dev)
	if buf := b.staging.buf; buf != nil {
		vk.UnmapMemory(b.dev, b.staging.buf.mem)
		buf.Release()
	}
	b.runDefers()
	for _, rp := range b.passes {
		vk.DestroyRenderPass(b.dev, rp)
	}
	vk.DestroyFence(b.dev, b.fence)
	vk.FreeCommandBuffers(b.dev, b.cmdPool.pool, b.cmdPool.buffers...)
	vk.DestroyCommandPool(b.dev, b.cmdPool.pool)
	*b = Backend{}
}

func (b *Backend) NewTexture(format driver.TextureFormat, width, height int, minFilter, magFilter driver.TextureFilter, bindings driver.BufferBinding) (driver.Texture, error) {
	vkfmt := formatFor(format)
	usage := vk.IMAGE_USAGE_TRANSFER_DST_BIT | vk.IMAGE_USAGE_TRANSFER_SRC_BIT
	passLayout := vk.IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL
	if bindings&driver.BufferBindingTexture != 0 {
		usage |= vk.IMAGE_USAGE_SAMPLED_BIT
		passLayout = vk.IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL
	}
	if bindings&driver.BufferBindingFramebuffer != 0 {
		usage |= vk.IMAGE_USAGE_COLOR_ATTACHMENT_BIT
	}
	if bindings&(driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite) != 0 {
		usage |= vk.IMAGE_USAGE_STORAGE_BIT
	}
	filterFor := func(f driver.TextureFilter) vk.Filter {
		switch f {
		case driver.FilterLinear, driver.FilterLinearMipmapLinear:
			return vk.FILTER_LINEAR
		case driver.FilterNearest:
			return vk.FILTER_NEAREST
		}
		panic("unknown filter")
	}
	mipmapMode := vk.SAMPLER_MIPMAP_MODE_NEAREST
	mipmap := minFilter == driver.FilterLinearMipmapLinear
	nmipmaps := 1
	if mipmap {
		mipmapMode = vk.SAMPLER_MIPMAP_MODE_LINEAR
		dim := width
		if height > dim {
			dim = height
		}
		log2 := 32 - bits.LeadingZeros32(uint32(dim)) - 1
		nmipmaps = log2 + 1
	}
	sampler, err := vk.CreateSampler(b.dev, filterFor(minFilter), filterFor(magFilter), mipmapMode)
	if err != nil {
		return nil, mapErr(err)
	}
	img, mem, err := vk.CreateImage(b.physDev, b.dev, vkfmt, width, height, nmipmaps, usage)
	if err != nil {
		vk.DestroySampler(b.dev, sampler)
		return nil, mapErr(err)
	}
	view, err := vk.CreateImageView(b.dev, img, vkfmt)
	if err != nil {
		vk.DestroySampler(b.dev, sampler)
		vk.DestroyImage(b.dev, img)
		vk.FreeMemory(b.dev, mem)
		return nil, mapErr(err)
	}
	t := &Texture{backend: b, img: img, mem: mem, view: view, sampler: sampler, layout: vk.IMAGE_LAYOUT_UNDEFINED, passLayout: passLayout, width: width, height: height, format: vkfmt, mipmaps: nmipmaps}
	if bindings&driver.BufferBindingFramebuffer != 0 {
		pass, err := vk.CreateRenderPass(b.dev, vkfmt, vk.ATTACHMENT_LOAD_OP_DONT_CARE,
			vk.IMAGE_LAYOUT_UNDEFINED, vk.IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL, nil)
		if err != nil {
			return nil, mapErr(err)
		}
		defer vk.DestroyRenderPass(b.dev, pass)
		fbo, err := vk.CreateFramebuffer(b.dev, pass, view, width, height)
		if err != nil {
			return nil, mapErr(err)
		}
		t.fbo = fbo
	}
	return t, nil
}

func (b *Backend) NewBuffer(bindings driver.BufferBinding, size int) (driver.Buffer, error) {
	if bindings&driver.BufferBindingUniforms != 0 {
		// Implement uniform buffers as inline push constants.
		return &Buffer{store: make([]byte, size)}, nil
	}
	usage := vk.BUFFER_USAGE_TRANSFER_DST_BIT | vk.BUFFER_USAGE_TRANSFER_SRC_BIT
	if bindings&driver.BufferBindingIndices != 0 {
		usage |= vk.BUFFER_USAGE_INDEX_BUFFER_BIT
	}
	if bindings&(driver.BufferBindingShaderStorageRead|driver.BufferBindingShaderStorageWrite) != 0 {
		usage |= vk.BUFFER_USAGE_STORAGE_BUFFER_BIT
	}
	if bindings&driver.BufferBindingVertices != 0 {
		usage |= vk.BUFFER_USAGE_VERTEX_BUFFER_BIT
	}
	buf, err := b.newBuffer(size, usage, vk.MEMORY_PROPERTY_DEVICE_LOCAL_BIT)
	return buf, mapErr(err)
}

func (b *Backend) newBuffer(size int, usage vk.BufferUsageFlags, props vk.MemoryPropertyFlags) (*Buffer, error) {
	buf, mem, err := vk.CreateBuffer(b.physDev, b.dev, size, usage, props)
	return &Buffer{backend: b, buf: buf, mem: mem, usage: usage}, err
}

func (b *Backend) NewImmutableBuffer(typ driver.BufferBinding, data []byte) (driver.Buffer, error) {
	buf, err := b.NewBuffer(typ, len(data))
	if err != nil {
		return nil, err
	}
	buf.Upload(data)
	return buf, nil
}

func (b *Backend) NewVertexShader(src shader.Sources) (driver.VertexShader, error) {
	sh, err := b.newShader(src, vk.SHADER_STAGE_VERTEX_BIT)
	return sh, mapErr(err)
}

func (b *Backend) NewFragmentShader(src shader.Sources) (driver.FragmentShader, error) {
	sh, err := b.newShader(src, vk.SHADER_STAGE_FRAGMENT_BIT)
	return sh, mapErr(err)
}

func (b *Backend) NewPipeline(desc driver.PipelineDesc) (driver.Pipeline, error) {
	vs := desc.VertexShader.(*Shader)
	fs := desc.FragmentShader.(*Shader)
	var ranges []vk.PushConstantRange
	if r := vs.pushRange; r != (vk.PushConstantRange{}) {
		ranges = append(ranges, r)
	}
	if r := fs.pushRange; r != (vk.PushConstantRange{}) {
		ranges = append(ranges, r)
	}
	descPool, err := createPipelineLayout(b.dev, fs.src, ranges)
	if err != nil {
		return nil, mapErr(err)
	}
	blend := desc.BlendDesc
	factorFor := func(f driver.BlendFactor) vk.BlendFactor {
		switch f {
		case driver.BlendFactorZero:
			return vk.BLEND_FACTOR_ZERO
		case driver.BlendFactorOne:
			return vk.BLEND_FACTOR_ONE
		case driver.BlendFactorOneMinusSrcAlpha:
			return vk.BLEND_FACTOR_ONE_MINUS_SRC_ALPHA
		case driver.BlendFactorDstColor:
			return vk.BLEND_FACTOR_DST_COLOR
		default:
			panic("unknown blend factor")
		}
	}
	var top vk.PrimitiveTopology
	switch desc.Topology {
	case driver.TopologyTriangles:
		top = vk.PRIMITIVE_TOPOLOGY_TRIANGLE_LIST
	case driver.TopologyTriangleStrip:
		top = vk.PRIMITIVE_TOPOLOGY_TRIANGLE_STRIP
	default:
		panic("unknown topology")
	}
	var binds []vk.VertexInputBindingDescription
	var attrs []vk.VertexInputAttributeDescription
	inputs := desc.VertexLayout.Inputs
	for i, inp := range inputs {
		binds = append(binds, vk.VertexInputBindingDescription{
			Binding: i,
			Stride:  desc.VertexLayout.Stride,
		})
		attrs = append(attrs, vk.VertexInputAttributeDescription{
			Binding:  i,
			Location: vs.src.Inputs[i].Location,
			Format:   vertFormatFor(vs.src.Inputs[i]),
			Offset:   inp.Offset,
		})
	}
	fmt := b.outFormat
	if f := desc.PixelFormat; f != driver.TextureFormatOutput {
		fmt = formatFor(f)
	}
	pass, err := vk.CreateRenderPass(b.dev, fmt, vk.ATTACHMENT_LOAD_OP_DONT_CARE,
		vk.IMAGE_LAYOUT_UNDEFINED, vk.IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL, nil)
	if err != nil {
		return nil, mapErr(err)
	}
	defer vk.DestroyRenderPass(b.dev, pass)
	pipe, err := vk.CreateGraphicsPipeline(b.dev, pass, vs.module, fs.module, blend.Enable, factorFor(blend.SrcFactor), factorFor(blend.DstFactor), top, binds, attrs, descPool.layout)
	if err != nil {
		descPool.release(b.dev)
		return nil, mapErr(err)
	}
	p := &Pipeline{backend: b, pipe: pipe, desc: descPool, pushRanges: ranges, ninputs: len(inputs)}
	b.allPipes = append(b.allPipes, p)
	return p, nil
}

func (b *Backend) NewComputeProgram(src shader.Sources) (driver.Program, error) {
	sh, err := b.newShader(src, vk.SHADER_STAGE_COMPUTE_BIT)
	if err != nil {
		return nil, mapErr(err)
	}
	defer sh.Release()
	descPool, err := createPipelineLayout(b.dev, src, nil)
	if err != nil {
		return nil, mapErr(err)
	}
	pipe, err := vk.CreateComputePipeline(b.dev, sh.module, descPool.layout)
	if err != nil {
		descPool.release(b.dev)
		return nil, mapErr(err)
	}
	return &Pipeline{backend: b, pipe: pipe, desc: descPool}, nil
}

func vertFormatFor(f shader.InputLocation) vk.Format {
	t := f.Type
	s := f.Size
	switch {
	case t == shader.DataTypeFloat && s == 1:
		return vk.FORMAT_R32_SFLOAT
	case t == shader.DataTypeFloat && s == 2:
		return vk.FORMAT_R32G32_SFLOAT
	case t == shader.DataTypeFloat && s == 3:
		return vk.FORMAT_R32G32B32_SFLOAT
	case t == shader.DataTypeFloat && s == 4:
		return vk.FORMAT_R32G32B32A32_SFLOAT
	default:
		panic("unsupported data type")
	}
}

func createPipelineLayout(d vk.Device, src shader.Sources, ranges []vk.PushConstantRange) (*descPool, error) {
	var (
		descLayouts []vk.DescriptorSetLayout
		descLayout  vk.DescriptorSetLayout
	)
	texBinds := make([]int, len(src.Textures))
	imgBinds := make([]int, len(src.Images))
	bufBinds := make([]int, len(src.StorageBuffers))
	var descBinds []vk.DescriptorSetLayoutBinding
	for i, t := range src.Textures {
		descBinds = append(descBinds, vk.DescriptorSetLayoutBinding{
			Binding:        t.Binding,
			StageFlags:     vk.SHADER_STAGE_FRAGMENT_BIT,
			DescriptorType: vk.DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER,
		})
		texBinds[i] = t.Binding
	}
	for i, img := range src.Images {
		descBinds = append(descBinds, vk.DescriptorSetLayoutBinding{
			Binding:        img.Binding,
			StageFlags:     vk.SHADER_STAGE_COMPUTE_BIT,
			DescriptorType: vk.DESCRIPTOR_TYPE_STORAGE_IMAGE,
		})
		imgBinds[i] = img.Binding
	}
	for i, buf := range src.StorageBuffers {
		descBinds = append(descBinds, vk.DescriptorSetLayoutBinding{
			Binding:        buf.Binding,
			StageFlags:     vk.SHADER_STAGE_COMPUTE_BIT,
			DescriptorType: vk.DESCRIPTOR_TYPE_STORAGE_BUFFER,
		})
		bufBinds[i] = buf.Binding
	}
	if len(descBinds) > 0 {
		var err error
		descLayout, err = vk.CreateDescriptorSetLayout(d, descBinds)
		if err != nil {
			return nil, err
		}
		descLayouts = append(descLayouts, descLayout)
	}
	layout, err := vk.CreatePipelineLayout(d, ranges, descLayouts)
	if err != nil {
		if descLayout != 0 {
			vk.DestroyDescriptorSetLayout(d, descLayout)
		}
		return nil, err
	}
	descPool := &descPool{
		texBinds:   texBinds,
		bufBinds:   bufBinds,
		imgBinds:   imgBinds,
		layout:     layout,
		descLayout: descLayout,
	}
	return descPool, nil
}

func (b *Backend) newShader(src shader.Sources, stage vk.ShaderStageFlags) (*Shader, error) {
	mod, err := vk.CreateShaderModule(b.dev, src.SPIRV)
	if err != nil {
		return nil, err
	}

	sh := &Shader{dev: b.dev, module: mod, src: src}
	if locs := src.Uniforms.Locations; len(locs) > 0 {
		pushOffset := 0x7fffffff
		for _, l := range locs {
			if l.Offset < pushOffset {
				pushOffset = l.Offset
			}
		}
		sh.pushRange = vk.BuildPushConstantRange(stage, pushOffset, src.Uniforms.Size)
	}
	return sh, nil
}

func (b *Backend) CopyTexture(dstTex driver.Texture, dorig image.Point, srcFBO driver.Texture, srect image.Rectangle) {
	dst := dstTex.(*Texture)
	src := srcFBO.(*Texture)
	cmdBuf := b.ensureCmdBuf()
	op := vk.BuildImageCopy(srect.Min.X, srect.Min.Y, dorig.X, dorig.Y, srect.Dx(), srect.Dy())
	src.imageBarrier(cmdBuf,
		vk.IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL,
		vk.PIPELINE_STAGE_TRANSFER_BIT,
		vk.ACCESS_TRANSFER_READ_BIT,
	)
	dst.imageBarrier(cmdBuf,
		vk.IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL,
		vk.PIPELINE_STAGE_TRANSFER_BIT,
		vk.ACCESS_TRANSFER_WRITE_BIT,
	)
	vk.CmdCopyImage(cmdBuf, src.img, src.layout, dst.img, dst.layout, []vk.ImageCopy{op})
}

func (b *Backend) Viewport(x, y, width, height int) {
	cmdBuf := b.currentCmdBuf()
	vp := vk.BuildViewport(float32(x), float32(y), float32(width), float32(height))
	vk.CmdSetViewport(cmdBuf, 0, vp)
}

func (b *Backend) DrawArrays(off, count int) {
	cmdBuf := b.currentCmdBuf()
	if b.desc.dirty {
		b.pipe.desc.bindDescriptorSet(b, cmdBuf, vk.PIPELINE_BIND_POINT_GRAPHICS, b.desc.texBinds, b.desc.bufBinds)
		b.desc.dirty = false
	}
	vk.CmdDraw(cmdBuf, count, 1, off, 0)
}

func (b *Backend) DrawElements(off, count int) {
	cmdBuf := b.currentCmdBuf()
	if b.desc.dirty {
		b.pipe.desc.bindDescriptorSet(b, cmdBuf, vk.PIPELINE_BIND_POINT_GRAPHICS, b.desc.texBinds, b.desc.bufBinds)
		b.desc.dirty = false
	}
	vk.CmdDrawIndexed(cmdBuf, count, 1, off, 0, 0)
}

func (b *Backend) BindImageTexture(unit int, tex driver.Texture) {
	t := tex.(*Texture)
	b.desc.texBinds[unit] = t
	b.desc.dirty = true
	t.imageBarrier(b.currentCmdBuf(),
		vk.IMAGE_LAYOUT_GENERAL,
		vk.PIPELINE_STAGE_COMPUTE_SHADER_BIT,
		vk.ACCESS_SHADER_READ_BIT|vk.ACCESS_SHADER_WRITE_BIT,
	)
}

func (b *Backend) DispatchCompute(x, y, z int) {
	cmdBuf := b.currentCmdBuf()
	if b.desc.dirty {
		b.pipe.desc.bindDescriptorSet(b, cmdBuf, vk.PIPELINE_BIND_POINT_COMPUTE, b.desc.texBinds, b.desc.bufBinds)
		b.desc.dirty = false
	}
	vk.CmdDispatch(cmdBuf, x, y, z)
}

func (t *Texture) Upload(offset, size image.Point, pixels []byte, stride int) {
	if stride == 0 {
		stride = size.X * 4
	}
	cmdBuf := t.backend.ensureCmdBuf()
	dstStride := size.X * 4
	n := size.Y * dstStride
	stage, mem, off := t.backend.stagingBuffer(n)
	var srcOff, dstOff int
	for y := 0; y < size.Y; y++ {
		srcRow := pixels[srcOff : srcOff+dstStride]
		dstRow := mem[dstOff : dstOff+dstStride]
		copy(dstRow, srcRow)
		dstOff += dstStride
		srcOff += stride
	}
	op := vk.BuildBufferImageCopy(off, dstStride/4, offset.X, offset.Y, size.X, size.Y)
	t.imageBarrier(cmdBuf,
		vk.IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL,
		vk.PIPELINE_STAGE_TRANSFER_BIT,
		vk.ACCESS_TRANSFER_WRITE_BIT,
	)
	vk.CmdCopyBufferToImage(cmdBuf, stage.buf, t.img, t.layout, op)
	// Build mipmaps by repeating linear blits.
	w, h := t.width, t.height
	for i := 1; i < t.mipmaps; i++ {
		nw, nh := w/2, h/2
		if nh < 1 {
			nh = 1
		}
		if nw < 1 {
			nw = 1
		}
		// Transition previous (source) level.
		b := vk.BuildImageMemoryBarrier(
			t.img,
			vk.ACCESS_TRANSFER_WRITE_BIT, vk.ACCESS_TRANSFER_READ_BIT,
			vk.IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL, vk.IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL,
			i-1, 1,
		)
		vk.CmdPipelineBarrier(cmdBuf, vk.PIPELINE_STAGE_TRANSFER_BIT, vk.PIPELINE_STAGE_TRANSFER_BIT, vk.DEPENDENCY_BY_REGION_BIT, nil, nil, []vk.ImageMemoryBarrier{b})
		// Blit to this mipmap level.
		blit := vk.BuildImageBlit(0, 0, 0, 0, w, h, nw, nh, i-1, i)
		vk.CmdBlitImage(cmdBuf, t.img, vk.IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL, t.img, vk.IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL, []vk.ImageBlit{blit}, vk.FILTER_LINEAR)
		w, h = nw, nh
	}
	if t.mipmaps > 1 {
		// Add barrier for last blit.
		b := vk.BuildImageMemoryBarrier(
			t.img,
			vk.ACCESS_TRANSFER_WRITE_BIT, vk.ACCESS_TRANSFER_READ_BIT,
			vk.IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL, vk.IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL,
			t.mipmaps-1, 1,
		)
		vk.CmdPipelineBarrier(cmdBuf, vk.PIPELINE_STAGE_TRANSFER_BIT, vk.PIPELINE_STAGE_TRANSFER_BIT, vk.DEPENDENCY_BY_REGION_BIT, nil, nil, []vk.ImageMemoryBarrier{b})
		t.layout = vk.IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL
	}
}

func (t *Texture) Release() {
	if t.foreign {
		panic("external textures cannot be released")
	}
	freet := *t
	t.backend.deferFunc(func(d vk.Device) {
		if freet.fbo != 0 {
			vk.DestroyFramebuffer(d, freet.fbo)
		}
		vk.DestroySampler(d, freet.sampler)
		vk.DestroyImageView(d, freet.view)
		vk.DestroyImage(d, freet.img)
		vk.FreeMemory(d, freet.mem)
	})
	*t = Texture{}
}

func (p *Pipeline) Release() {
	freep := *p
	p.backend.deferFunc(func(d vk.Device) {
		freep.desc.release(d)
		vk.DestroyPipeline(d, freep.pipe)
	})
	*p = Pipeline{}
}

func (p *descPool) release(d vk.Device) {
	if p := p.pool; p != 0 {
		vk.DestroyDescriptorPool(d, p)
	}
	if l := p.descLayout; l != 0 {
		vk.DestroyDescriptorSetLayout(d, l)
	}
	vk.DestroyPipelineLayout(d, p.layout)
}

func (p *descPool) bindDescriptorSet(b *Backend, cmdBuf vk.CommandBuffer, bindPoint vk.PipelineBindPoint, texBinds [texUnits]*Texture, bufBinds [storageUnits]*Buffer) {
	if p.size == len(p.sets) {
		l := p.descLayout
		if l == 0 {
			panic("vulkan: descriptor set is dirty, but pipeline has empty layout")
		}
		newCap := len(p.sets) * 2
		if pool := p.pool; pool != 0 {
			b.deferFunc(func(d vk.Device) {
				vk.DestroyDescriptorPool(d, pool)
			})
		}
		const initialPoolSize = 100
		if newCap < initialPoolSize {
			newCap = initialPoolSize
		}
		var poolSizes []vk.DescriptorPoolSize
		if n := len(p.texBinds); n > 0 {
			poolSizes = append(poolSizes, vk.BuildDescriptorPoolSize(vk.DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER, newCap*n))
		}
		if n := len(p.imgBinds); n > 0 {
			poolSizes = append(poolSizes, vk.BuildDescriptorPoolSize(vk.DESCRIPTOR_TYPE_STORAGE_IMAGE, newCap*n))
		}
		if n := len(p.bufBinds); n > 0 {
			poolSizes = append(poolSizes, vk.BuildDescriptorPoolSize(vk.DESCRIPTOR_TYPE_STORAGE_BUFFER, newCap*n))
		}
		pool, err := vk.CreateDescriptorPool(b.dev, newCap, poolSizes)
		if err != nil {
			panic(fmt.Errorf("vulkan: failed to allocate descriptor pool with %d descriptors: %v", newCap, err))
		}
		p.pool = pool
		sets, err := vk.AllocateDescriptorSets(b.dev, p.pool, l, newCap)
		if err != nil {
			panic(fmt.Errorf("vulkan: failed to allocate descriptor with %d sets: %v", newCap, err))
		}
		p.sets = sets
		p.size = 0
	}
	descSet := p.sets[p.size]
	p.size++
	for _, bind := range p.texBinds {
		tex := texBinds[bind]
		write := vk.BuildWriteDescriptorSetImage(descSet, bind, vk.DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER, tex.sampler, tex.view, vk.IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL)
		vk.UpdateDescriptorSet(b.dev, write)
	}
	for _, bind := range p.imgBinds {
		tex := texBinds[bind]
		write := vk.BuildWriteDescriptorSetImage(descSet, bind, vk.DESCRIPTOR_TYPE_STORAGE_IMAGE, 0, tex.view, vk.IMAGE_LAYOUT_GENERAL)
		vk.UpdateDescriptorSet(b.dev, write)
	}
	for _, bind := range p.bufBinds {
		buf := bufBinds[bind]
		write := vk.BuildWriteDescriptorSetBuffer(descSet, bind, vk.DESCRIPTOR_TYPE_STORAGE_BUFFER, buf.buf)
		vk.UpdateDescriptorSet(b.dev, write)
	}
	vk.CmdBindDescriptorSets(cmdBuf, bindPoint, p.layout, 0, []vk.DescriptorSet{descSet})
}

func (t *Texture) imageBarrier(cmdBuf vk.CommandBuffer, layout vk.ImageLayout, stage vk.PipelineStageFlags, access vk.AccessFlags) {
	srcStage := t.scope.stage
	if srcStage == 0 && t.layout == layout {
		t.scope.stage = stage
		t.scope.access = access
		return
	}
	if srcStage == 0 {
		srcStage = vk.PIPELINE_STAGE_TOP_OF_PIPE_BIT
	}
	b := vk.BuildImageMemoryBarrier(
		t.img,
		t.scope.access, access,
		t.layout, layout,
		0, vk.REMAINING_MIP_LEVELS,
	)
	vk.CmdPipelineBarrier(cmdBuf, srcStage, stage, vk.DEPENDENCY_BY_REGION_BIT, nil, nil, []vk.ImageMemoryBarrier{b})
	t.layout = layout
	t.scope.stage = stage
	t.scope.access = access
}

func (b *Backend) PrepareTexture(tex driver.Texture) {
	t := tex.(*Texture)
	cmdBuf := b.ensureCmdBuf()
	t.imageBarrier(cmdBuf,
		vk.IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL,
		vk.PIPELINE_STAGE_FRAGMENT_SHADER_BIT,
		vk.ACCESS_SHADER_READ_BIT,
	)
}

func (b *Backend) BindTexture(unit int, tex driver.Texture) {
	t := tex.(*Texture)
	b.desc.texBinds[unit] = t
	b.desc.dirty = true
}

func (b *Backend) BindPipeline(pipe driver.Pipeline) {
	b.bindPipeline(pipe.(*Pipeline), vk.PIPELINE_BIND_POINT_GRAPHICS)
}

func (b *Backend) BindProgram(prog driver.Program) {
	b.bindPipeline(prog.(*Pipeline), vk.PIPELINE_BIND_POINT_COMPUTE)
}

func (b *Backend) bindPipeline(p *Pipeline, point vk.PipelineBindPoint) {
	b.pipe = p
	b.desc.dirty = p.desc.descLayout != 0
	cmdBuf := b.currentCmdBuf()
	vk.CmdBindPipeline(cmdBuf, point, p.pipe)
}

func (s *Shader) Release() {
	vk.DestroyShaderModule(s.dev, s.module)
	*s = Shader{}
}

func (b *Backend) BindStorageBuffer(binding int, buffer driver.Buffer) {
	buf := buffer.(*Buffer)
	b.desc.bufBinds[binding] = buf
	b.desc.dirty = true
	buf.barrier(b.currentCmdBuf(),
		vk.PIPELINE_STAGE_COMPUTE_SHADER_BIT,
		vk.ACCESS_SHADER_READ_BIT|vk.ACCESS_SHADER_WRITE_BIT,
	)
}

func (b *Backend) BindUniforms(buffer driver.Buffer) {
	buf := buffer.(*Buffer)
	cmdBuf := b.currentCmdBuf()
	for _, s := range b.pipe.pushRanges {
		off := s.Offset()
		vk.CmdPushConstants(cmdBuf, b.pipe.desc.layout, s.StageFlags(), off, buf.store[off:off+s.Size()])
	}
}

func (b *Backend) BindVertexBuffer(buffer driver.Buffer, offset int) {
	buf := buffer.(*Buffer)
	cmdBuf := b.currentCmdBuf()
	b.bindings = b.bindings[:0]
	b.offsets = b.offsets[:0]
	for i := 0; i < b.pipe.ninputs; i++ {
		b.bindings = append(b.bindings, buf.buf)
		b.offsets = append(b.offsets, vk.DeviceSize(offset))
	}
	vk.CmdBindVertexBuffers(cmdBuf, 0, b.bindings, b.offsets)
}

func (b *Backend) BindIndexBuffer(buffer driver.Buffer) {
	buf := buffer.(*Buffer)
	cmdBuf := b.currentCmdBuf()
	vk.CmdBindIndexBuffer(cmdBuf, buf.buf, 0, vk.INDEX_TYPE_UINT16)
}

func (b *Buffer) Download(data []byte) error {
	if b.buf == 0 {
		copy(data, b.store)
		return nil
	}
	stage, mem, off := b.backend.stagingBuffer(len(data))
	cmdBuf := b.backend.ensureCmdBuf()
	b.barrier(cmdBuf,
		vk.PIPELINE_STAGE_TRANSFER_BIT,
		vk.ACCESS_TRANSFER_READ_BIT,
	)
	vk.CmdCopyBuffer(cmdBuf, b.buf, stage.buf, 0, off, len(data))
	stage.scope.stage = vk.PIPELINE_STAGE_TRANSFER_BIT
	stage.scope.access = vk.ACCESS_TRANSFER_WRITE_BIT
	stage.barrier(cmdBuf,
		vk.PIPELINE_STAGE_HOST_BIT,
		vk.ACCESS_HOST_READ_BIT,
	)
	b.backend.submitCmdBuf(b.backend.fence)
	vk.WaitForFences(b.backend.dev, b.backend.fence)
	vk.ResetFences(b.backend.dev, b.backend.fence)
	copy(data, mem)
	return nil
}

func (b *Buffer) Upload(data []byte) {
	if b.buf == 0 {
		copy(b.store, data)
		return
	}
	stage, mem, off := b.backend.stagingBuffer(len(data))
	copy(mem, data)
	cmdBuf := b.backend.ensureCmdBuf()
	b.barrier(cmdBuf,
		vk.PIPELINE_STAGE_TRANSFER_BIT,
		vk.ACCESS_TRANSFER_WRITE_BIT,
	)
	vk.CmdCopyBuffer(cmdBuf, stage.buf, b.buf, off, 0, len(data))
	var access vk.AccessFlags
	if b.usage&vk.BUFFER_USAGE_INDEX_BUFFER_BIT != 0 {
		access |= vk.ACCESS_INDEX_READ_BIT
	}
	if b.usage&vk.BUFFER_USAGE_VERTEX_BUFFER_BIT != 0 {
		access |= vk.ACCESS_VERTEX_ATTRIBUTE_READ_BIT
	}
	if access != 0 {
		b.barrier(cmdBuf,
			vk.PIPELINE_STAGE_VERTEX_INPUT_BIT,
			access,
		)
	}
}

func (b *Buffer) barrier(cmdBuf vk.CommandBuffer, stage vk.PipelineStageFlags, access vk.AccessFlags) {
	srcStage := b.scope.stage
	if srcStage == 0 {
		b.scope.stage = stage
		b.scope.access = access
		return
	}
	barrier := vk.BuildBufferMemoryBarrier(
		b.buf,
		b.scope.access, access,
	)
	vk.CmdPipelineBarrier(cmdBuf, srcStage, stage, vk.DEPENDENCY_BY_REGION_BIT, nil, []vk.BufferMemoryBarrier{barrier}, nil)
	b.scope.stage = stage
	b.scope.access = access
}

func (b *Buffer) Release() {
	freeb := *b
	if freeb.buf != 0 {
		b.backend.deferFunc(func(d vk.Device) {
			vk.DestroyBuffer(d, freeb.buf)
			vk.FreeMemory(d, freeb.mem)
		})
	}
	*b = Buffer{}
}

func (t *Texture) ReadPixels(src image.Rectangle, pixels []byte, stride int) error {
	if len(pixels) == 0 {
		return nil
	}
	sz := src.Size()
	stageStride := sz.X * 4
	n := sz.Y * stageStride
	stage, mem, off := t.backend.stagingBuffer(n)
	cmdBuf := t.backend.ensureCmdBuf()
	region := vk.BuildBufferImageCopy(off, stageStride/4, src.Min.X, src.Min.Y, sz.X, sz.Y)
	t.imageBarrier(cmdBuf,
		vk.IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL,
		vk.PIPELINE_STAGE_TRANSFER_BIT,
		vk.ACCESS_TRANSFER_READ_BIT,
	)
	vk.CmdCopyImageToBuffer(cmdBuf, t.img, t.layout, stage.buf, []vk.BufferImageCopy{region})
	stage.scope.stage = vk.PIPELINE_STAGE_TRANSFER_BIT
	stage.scope.access = vk.ACCESS_TRANSFER_WRITE_BIT
	stage.barrier(cmdBuf,
		vk.PIPELINE_STAGE_HOST_BIT,
		vk.ACCESS_HOST_READ_BIT,
	)
	t.backend.submitCmdBuf(t.backend.fence)
	vk.WaitForFences(t.backend.dev, t.backend.fence)
	vk.ResetFences(t.backend.dev, t.backend.fence)
	var srcOff, dstOff int
	for y := 0; y < sz.Y; y++ {
		dstRow := pixels[srcOff : srcOff+stageStride]
		srcRow := mem[dstOff : dstOff+stageStride]
		copy(dstRow, srcRow)
		dstOff += stageStride
		srcOff += stride
	}
	return nil
}

func (b *Backend) currentCmdBuf() vk.CommandBuffer {
	cur := b.cmdPool.current
	if cur == nil {
		panic("vulkan: invalid operation outside a render or compute pass")
	}
	return cur
}

func (b *Backend) ensureCmdBuf() vk.CommandBuffer {
	if b.cmdPool.current != nil {
		return b.cmdPool.current
	}
	if b.cmdPool.used < len(b.cmdPool.buffers) {
		buf := b.cmdPool.buffers[b.cmdPool.used]
		b.cmdPool.current = buf
	} else {
		buf, err := vk.AllocateCommandBuffer(b.dev, b.cmdPool.pool)
		if err != nil {
			panic(err)
		}
		b.cmdPool.buffers = append(b.cmdPool.buffers, buf)
		b.cmdPool.current = buf
	}
	b.cmdPool.used++
	buf := b.cmdPool.current
	if err := vk.BeginCommandBuffer(buf); err != nil {
		panic(err)
	}
	return buf
}

func (b *Backend) BeginRenderPass(tex driver.Texture, d driver.LoadDesc) {
	t := tex.(*Texture)
	var vkop vk.AttachmentLoadOp
	switch d.Action {
	case driver.LoadActionClear:
		vkop = vk.ATTACHMENT_LOAD_OP_CLEAR
	case driver.LoadActionInvalidate:
		vkop = vk.ATTACHMENT_LOAD_OP_DONT_CARE
	case driver.LoadActionKeep:
		vkop = vk.ATTACHMENT_LOAD_OP_LOAD
	}
	cmdBuf := b.ensureCmdBuf()
	if sem := t.acquire; sem != 0 {
		// The render pass targets a framebuffer that has an associated acquire semaphore.
		// Wait for it by forming an execution barrier.
		b.waitSems = append(b.waitSems, sem)
		b.waitStages = append(b.waitStages, vk.PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT)
		// But only for the first pass in a frame.
		t.acquire = 0
	}
	t.imageBarrier(cmdBuf,
		vk.IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL,
		vk.PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT,
		vk.ACCESS_COLOR_ATTACHMENT_READ_BIT|vk.ACCESS_COLOR_ATTACHMENT_WRITE_BIT,
	)
	pass := b.lookupPass(t.format, vkop, t.layout, t.passLayout)
	col := d.ClearColor
	vk.CmdBeginRenderPass(cmdBuf, pass, t.fbo, t.width, t.height, [4]float32{col.R, col.G, col.B, col.A})
	t.layout = t.passLayout
	// If the render pass describes an automatic image layout transition to its final layout, there
	// is an implicit image barrier with destination PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT. Make
	// sure any subsequent barrier includes the transition.
	// See also https://www.khronos.org/registry/vulkan/specs/1.0/html/vkspec.html#VkSubpassDependency.
	t.scope.stage |= vk.PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT
}

func (b *Backend) EndRenderPass() {
	vk.CmdEndRenderPass(b.cmdPool.current)
}

func (b *Backend) BeginCompute() {
	b.ensureCmdBuf()
}

func (b *Backend) EndCompute() {
}

func (b *Backend) lookupPass(fmt vk.Format, loadAct vk.AttachmentLoadOp, initLayout, finalLayout vk.ImageLayout) vk.RenderPass {
	key := passKey{fmt: fmt, loadAct: loadAct, initLayout: initLayout, finalLayout: finalLayout}
	if pass, ok := b.passes[key]; ok {
		return pass
	}
	pass, err := vk.CreateRenderPass(b.dev, fmt, loadAct, initLayout, finalLayout, nil)
	if err != nil {
		panic(err)
	}
	b.passes[key] = pass
	return pass
}

func (b *Backend) submitCmdBuf(fence vk.Fence) {
	buf := b.cmdPool.current
	if buf == nil && fence == 0 {
		return
	}
	buf = b.ensureCmdBuf()
	b.cmdPool.current = nil
	if err := vk.EndCommandBuffer(buf); err != nil {
		panic(err)
	}
	if err := vk.QueueSubmit(b.queue, buf, b.waitSems, b.waitStages, b.sigSems, fence); err != nil {
		panic(err)
	}
	b.waitSems = b.waitSems[:0]
	b.sigSems = b.sigSems[:0]
	b.waitStages = b.waitStages[:0]
}

func (b *Backend) stagingBuffer(size int) (*Buffer, []byte, int) {
	if b.staging.size+size > b.staging.cap {
		if b.staging.buf != nil {
			vk.UnmapMemory(b.dev, b.staging.buf.mem)
			b.staging.buf.Release()
			b.staging.cap = 0
		}
		cap := 2 * (b.staging.size + size)
		buf, err := b.newBuffer(cap, vk.BUFFER_USAGE_TRANSFER_SRC_BIT|vk.BUFFER_USAGE_TRANSFER_DST_BIT,
			vk.MEMORY_PROPERTY_HOST_VISIBLE_BIT|vk.MEMORY_PROPERTY_HOST_COHERENT_BIT)
		if err != nil {
			panic(err)
		}
		mem, err := vk.MapMemory(b.dev, buf.mem, 0, cap)
		if err != nil {
			buf.Release()
			panic(err)
		}
		b.staging.buf = buf
		b.staging.mem = mem
		b.staging.size = 0
		b.staging.cap = cap
	}
	off := b.staging.size
	b.staging.size += size
	mem := b.staging.mem[off : off+size]
	return b.staging.buf, mem, off
}

func formatFor(format driver.TextureFormat) vk.Format {
	switch format {
	case driver.TextureFormatRGBA8:
		return vk.FORMAT_R8G8B8A8_UNORM
	case driver.TextureFormatSRGBA:
		return vk.FORMAT_R8G8B8A8_SRGB
	case driver.TextureFormatFloat:
		return vk.FORMAT_R16_SFLOAT
	default:
		panic("unsupported texture format")
	}
}

func mapErr(err error) error {
	var vkErr vk.Error
	if errors.As(err, &vkErr) && vkErr == vk.ERROR_DEVICE_LOST {
		return driver.ErrDeviceLost
	}
	return err
}

func (f *Texture) ImplementsRenderTarget() {}
