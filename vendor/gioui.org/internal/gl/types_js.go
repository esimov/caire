// SPDX-License-Identifier: Unlicense OR MIT

package gl

import "syscall/js"

type (
	Object       js.Value
	Buffer       Object
	Framebuffer  Object
	Program      Object
	Renderbuffer Object
	Shader       Object
	Texture      Object
	Query        Object
	Uniform      Object
	VertexArray  Object
)

func (o Object) valid() bool {
	return js.Value(o).Truthy()
}

func (o Object) equal(o2 Object) bool {
	return js.Value(o).Equal(js.Value(o2))
}

func (b Buffer) Valid() bool {
	return Object(b).valid()
}

func (f Framebuffer) Valid() bool {
	return Object(f).valid()
}

func (p Program) Valid() bool {
	return Object(p).valid()
}

func (r Renderbuffer) Valid() bool {
	return Object(r).valid()
}

func (s Shader) Valid() bool {
	return Object(s).valid()
}

func (t Texture) Valid() bool {
	return Object(t).valid()
}

func (u Uniform) Valid() bool {
	return Object(u).valid()
}

func (a VertexArray) Valid() bool {
	return Object(a).valid()
}

func (f Framebuffer) Equal(f2 Framebuffer) bool {
	return Object(f).equal(Object(f2))
}

func (p Program) Equal(p2 Program) bool {
	return Object(p).equal(Object(p2))
}

func (s Shader) Equal(s2 Shader) bool {
	return Object(s).equal(Object(s2))
}

func (u Uniform) Equal(u2 Uniform) bool {
	return Object(u).equal(Object(u2))
}

func (a VertexArray) Equal(a2 VertexArray) bool {
	return Object(a).equal(Object(a2))
}

func (r Renderbuffer) Equal(r2 Renderbuffer) bool {
	return Object(r).equal(Object(r2))
}

func (t Texture) Equal(t2 Texture) bool {
	return Object(t).equal(Object(t2))
}

func (b Buffer) Equal(b2 Buffer) bool {
	return Object(b).equal(Object(b2))
}
