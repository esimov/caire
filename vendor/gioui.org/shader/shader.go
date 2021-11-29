// SPDX-License-Identifier: Unlicense OR MIT

package shader

type Sources struct {
	Name           string
	SPIRV          string
	GLSL100ES      string
	GLSL150        string
	DXBC           string
	MetalLib       string
	Uniforms       UniformsReflection
	Inputs         []InputLocation
	Textures       []TextureBinding
	StorageBuffers []BufferBinding
	Images         []ImageBinding
	WorkgroupSize  [3]int
}

type UniformsReflection struct {
	Locations []UniformLocation
	Size      int
}

type ImageBinding struct {
	Name    string
	Binding int
}

type BufferBinding struct {
	Name    string
	Binding int
}

type TextureBinding struct {
	Name    string
	Binding int
}

type UniformLocation struct {
	Name   string
	Type   DataType
	Size   int
	Offset int
}

type InputLocation struct {
	// For GLSL.
	Name     string
	Location int
	// For HLSL.
	Semantic      string
	SemanticIndex int

	Type DataType
	Size int
}

type DataType uint8

const (
	DataTypeFloat DataType = iota
	DataTypeInt
	DataTypeShort
)
