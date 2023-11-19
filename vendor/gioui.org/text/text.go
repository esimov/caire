// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"fmt"

	"gioui.org/io/system"
	"golang.org/x/image/math/fixed"
)

type Alignment uint8

const (
	Start Alignment = iota
	End
	Middle
)

func (a Alignment) String() string {
	switch a {
	case Start:
		return "Start"
	case End:
		return "End"
	case Middle:
		return "Middle"
	default:
		panic("invalid Alignment")
	}
}

// Align returns the x offset that should be applied to text with width so that it
// appears correctly aligned within a space of size maxWidth and with the primary
// text direction dir.
func (a Alignment) Align(dir system.TextDirection, width fixed.Int26_6, maxWidth int) fixed.Int26_6 {
	mw := fixed.I(maxWidth)
	if dir.Progression() == system.TowardOrigin {
		switch a {
		case Start:
			a = End
		case End:
			a = Start
		}
	}
	switch a {
	case Middle:
		return (mw - width) / 2
	case End:
		return (mw - width)
	case Start:
		return 0
	default:
		panic(fmt.Errorf("unknown alignment %v", a))
	}
}
