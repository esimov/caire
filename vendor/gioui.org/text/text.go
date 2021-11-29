// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"io"

	"gioui.org/op/clip"
	"golang.org/x/image/math/fixed"
)

// A Line contains the measurements of a line of text.
type Line struct {
	Layout Layout
	// Width is the width of the line.
	Width fixed.Int26_6
	// Ascent is the height above the baseline.
	Ascent fixed.Int26_6
	// Descent is the height below the baseline, including
	// the line gap.
	Descent fixed.Int26_6
	// Bounds is the visible bounds of the line.
	Bounds fixed.Rectangle26_6
}

type Layout struct {
	Text     string
	Advances []fixed.Int26_6
}

// Style is the font style.
type Style int

// Weight is a font weight, in CSS units subtracted 400 so the zero value
// is normal text weight.
type Weight int

// Font specify a particular typeface variant, style and weight.
type Font struct {
	Typeface Typeface
	Variant  Variant
	Style    Style
	// Weight is the text weight. If zero, Normal is used instead.
	Weight Weight
}

// Face implements text layout and shaping for a particular font. All
// methods must be safe for concurrent use.
type Face interface {
	Layout(ppem fixed.Int26_6, maxWidth int, txt io.Reader) ([]Line, error)
	Shape(ppem fixed.Int26_6, str Layout) clip.Op
}

// Typeface identifies a particular typeface design. The empty
// string denotes the default typeface.
type Typeface string

// Variant denotes a typeface variant such as "Mono" or "Smallcaps".
type Variant string

type Alignment uint8

const (
	Start Alignment = iota
	End
	Middle
)

const (
	Regular Style = iota
	Italic
)

const (
	Thin       Weight = 100 - 400
	Hairline   Weight = Thin
	ExtraLight Weight = 200 - 400
	UltraLight Weight = ExtraLight
	Light      Weight = 300 - 400
	Normal     Weight = 400 - 400
	Medium     Weight = 500 - 400
	SemiBold   Weight = 600 - 400
	DemiBold   Weight = SemiBold
	Bold       Weight = 700 - 400
	ExtraBold  Weight = 800 - 400
	UltraBold  Weight = ExtraBold
	Black      Weight = 900 - 400
	Heavy      Weight = Black
	ExtraBlack Weight = 950 - 400
	UltraBlack Weight = ExtraBlack
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

func (s Style) String() string {
	switch s {
	case Regular:
		return "Regular"
	case Italic:
		return "Italic"
	default:
		panic("invalid Style")
	}
}

func (w Weight) String() string {
	switch w {
	case Thin:
		return "Thin"
	case ExtraLight:
		return "ExtraLight"
	case Light:
		return "Light"
	case Normal:
		return "Normal"
	case Medium:
		return "Medium"
	case SemiBold:
		return "SemiBold"
	case Bold:
		return "Bold"
	case ExtraBold:
		return "ExtraBold"
	case Black:
		return "Black"
	case ExtraBlack:
		return "ExtraBlack"
	default:
		panic("invalid Weight")
	}
}

// weightDistance returns the distance value between two font weights.
func weightDistance(wa Weight, wb Weight) int {
	// Avoid dealing with negative Weight values.
	a := int(wa) + 400
	b := int(wb) + 400
	diff := a - b
	if diff < 0 {
		return -diff
	}
	return diff
}
