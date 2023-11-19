/*
Package font provides type describing font faces attributes.
*/
package font

import (
	"github.com/go-text/typesetting/font"
)

// A FontFace is a Font and a matching Face.
type FontFace struct {
	Font Font
	Face Face
}

// Style is the font style.
type Style int

// Weight is a font weight, in CSS units subtracted 400 so the zero value
// is normal text weight.
type Weight int

// Font specify a particular typeface variant, style and weight.
type Font struct {
	// Typeface specifies the name(s) of the the font faces to try. See [Typeface]
	// for details.
	Typeface Typeface
	// Style specifies the kind of text style.
	Style Style
	// Weight is the text weight.
	Weight Weight
}

// Face is an opaque handle to a typeface. The concrete implementation depends
// upon the kind of font and shaper in use.
type Face interface {
	Face() font.Face
}

// Typeface identifies a list of font families to attempt to use for displaying
// a string. The syntax is a comma-delimited list of family names. In order to
// allow for the remote possibility of needing to express a font family name
// containing a comma, name entries may be quoted using either single or double
// quotes. Within quotes, a literal quotation mark can be expressed by escaping
// it with `\`. A literal backslash may be expressed by escaping it with another
// `\`.
//
// Here's an example Typeface:
//
//	Times New Roman, Georgia, serif
//
// This is equivalent to the above:
//
//	"Times New Roman", 'Georgia', serif
//
// Here are some valid uses of escape sequences:
//
//	"Contains a literal \" doublequote", 'Literal \' Singlequote', "\\ Literal backslash", '\\ another'
//
// This syntax has the happy side effect that most CSS "font-family" rules are
// valid Typefaces (without the trailing semicolon).
//
// Generic CSS font families are supported, and are automatically expanded to lists
// of known font families with a matching style. The supported generic families are:
//
//   - fantasy
//   - math
//   - emoji
//   - serif
//   - sans-serif
//   - cursive
//   - monospace
type Typeface string

const (
	Regular Style = iota
	Italic
)

const (
	Thin       Weight = -300
	ExtraLight Weight = -200
	Light      Weight = -100
	Normal     Weight = 0
	Medium     Weight = 100
	SemiBold   Weight = 200
	Bold       Weight = 300
	ExtraBold  Weight = 400
	Black      Weight = 500
)

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
	default:
		panic("invalid Weight")
	}
}
