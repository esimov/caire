package imop

import (
	"github.com/esimov/caire/utils"
)

const (
	Darken   = "darken"
	Lighten  = "lighten"
	Multiply = "multiply"
	Screen   = "screen"
	Overlay  = "overlay"
)

type Blend struct {
	OpType string
}

func NewBlend() *Blend {
	return &Blend{}
}

func (o *Blend) Set(opType string) {
	bModes := []string{Darken, Lighten, Multiply, Screen, Overlay}

	if utils.Contains(bModes, opType) {
		o.OpType = opType
	}
}

func (o *Blend) Get() string {
	if len(o.OpType) > 0 {
		return o.OpType
	}
	return ""
}
