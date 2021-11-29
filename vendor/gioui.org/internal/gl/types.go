//go:build !js
// +build !js

package gl

type (
	Object       struct{ V uint }
	Buffer       Object
	Framebuffer  Object
	Program      Object
	Renderbuffer Object
	Shader       Object
	Texture      Object
	Query        Object
	Uniform      struct{ V int }
	VertexArray  Object
)

func (o Object) valid() bool {
	return o.V != 0
}

func (o Object) equal(o2 Object) bool {
	return o == o2
}

func (u Framebuffer) Valid() bool {
	return Object(u).valid()
}

func (u Uniform) Valid() bool {
	return u.V != -1
}

func (p Program) Valid() bool {
	return Object(p).valid()
}

func (s Shader) Valid() bool {
	return Object(s).valid()
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
	return u == u2
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
