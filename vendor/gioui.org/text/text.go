// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"io"

	"gioui.org/io/system"
	"gioui.org/op/clip"
	"github.com/go-text/typesetting/font"
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

// Range describes the position and quantity of a range of text elements
// within a larger slice. The unit is usually runes of unicode data or
// glyphs of shaped font data.
type Range struct {
	// Count describes the number of items represented by the Range.
	Count int
	// Offset describes the start position of the represented
	// items within a larger list.
	Offset int
}

// GlyphID uniquely identifies a glyph within a specific font.
type GlyphID = font.GID

// Glyph contains the metadata needed to render a glyph.
type Glyph struct {
	// ID is this glyph's identifier within the font it was shaped with.
	ID GlyphID
	// ClusterIndex is the identifier for the text shaping cluster that
	// this glyph is part of.
	ClusterIndex int
	// GlyphCount is the number of glyphs in the same cluster as this glyph.
	GlyphCount int
	// RuneCount is the quantity of runes in the source text that this glyph
	// corresponds to.
	RuneCount int
	// XAdvance and YAdvance describe the distance the dot moves when
	// laying out the glyph on the X or Y axis.
	XAdvance, YAdvance fixed.Int26_6
	// XOffset and YOffset describe offsets from the dot that should be
	// applied when rendering the glyph.
	XOffset, YOffset fixed.Int26_6
}

// GlyphCluster provides metadata about a sequence of indivisible shaped
// glyphs.
type GlyphCluster struct {
	// Advance is the cumulative advance of all glyphs in the cluster.
	Advance fixed.Int26_6
	// Runes indicates the position and quantity of the runes represented by
	// this cluster within the text.
	Runes Range
	// Glyphs indicates the position and quantity of the glyphs within this
	// cluster in a Layout's Glyphs slice.
	Glyphs Range
}

// RuneWidth returns the effective width of one rune for this cluster.
// If the cluster contains multiple runes, the width of the glyphs of
// the cluster is divided evenly among the runes.
func (c GlyphCluster) RuneWidth() fixed.Int26_6 {
	if c.Runes.Count == 0 {
		return 0
	}
	return c.Advance / fixed.Int26_6(c.Runes.Count)
}

type Layout struct {
	// Glyphs are the actual font characters for the text. They are ordered
	// from left to right regardless of the text direction of the underlying
	// text.
	Glyphs []Glyph
	// Clusters are metadata about the shaped glyphs. They are mostly useful for
	// interactive text widgets like editors. The order of clusters is logical,
	// so the first cluster will describe the beginning of the text and may
	// refer to the final glyphs in the Glyphs field if the text is RTL.
	Clusters []GlyphCluster
	// Runes describes the position of the text data this layout represents
	// within the overall body of text being shaped.
	Runes Range
	// Direction is the layout direction of the text.
	Direction system.TextDirection
}

// Slice returns a layout starting at the glyph cluster index start
// and running through the glyph cluster index end. The Offsets field
// of the returned layout is adjusted to reflect the new rune range
// covered by the layout. The returned layout will have no Clusters.
func (l Layout) Slice(start, end int) Layout {
	if start == end || end == 0 || start == len(l.Clusters) {
		return Layout{}
	}
	newRuneStart := l.Clusters[start].Runes.Offset
	runesBefore := newRuneStart - l.Runes.Offset
	endCluster := l.Clusters[end-1]
	startCluster := l.Clusters[start]
	runesAfter := l.Runes.Offset + l.Runes.Count - (endCluster.Runes.Offset + endCluster.Runes.Count)

	if l.Direction.Progression() == system.TowardOrigin {
		startCluster, endCluster = endCluster, startCluster
	}
	glyphStart := startCluster.Glyphs.Offset
	glyphEnd := endCluster.Glyphs.Offset + endCluster.Glyphs.Count

	out := l
	out.Clusters = nil
	out.Glyphs = out.Glyphs[glyphStart:glyphEnd]
	out.Runes.Offset = newRuneStart
	out.Runes.Count -= runesBefore + runesAfter

	return out
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
	Layout(ppem fixed.Int26_6, maxWidth int, lc system.Locale, txt io.RuneReader) ([]Line, error)
	Shape(ppem fixed.Int26_6, str Layout) clip.PathSpec
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
