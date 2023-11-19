// SPDX-License-Identifier: Unlicense OR MIT

// Package opentype implements text layout and shaping for OpenType
// files.
//
// NOTE: the OpenType specification allows for fonts to include bitmap images
// in a variety of formats. In the interest of small binary sizes, the opentype
// package only automatically imports the PNG image decoder. If you have a font
// with glyphs in JPEG or TIFF formats, register those decoders with the image
// package in order to ensure those glyphs are visible in text.
package opentype

import (
	"bytes"
	"fmt"
	_ "image/png"

	giofont "gioui.org/font"
	"github.com/go-text/typesetting/font"
	fontapi "github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
)

// Face is a thread-safe representation of a loaded font. For efficiency, applications
// should construct a face for any given font file once, reusing it across different
// text shapers.
type Face struct {
	face font.Font
	font giofont.Font
}

// Parse constructs a Face from source bytes.
func Parse(src []byte) (Face, error) {
	ld, err := loader.NewLoader(bytes.NewReader(src))
	if err != nil {
		return Face{}, err
	}
	font, md, err := parseLoader(ld)
	if err != nil {
		return Face{}, fmt.Errorf("failed parsing truetype font: %w", err)
	}
	return Face{
		face: font,
		font: md,
	}, nil
}

// ParseCollection parse an Opentype font file, with support for collections.
// Single font files are supported, returning a slice with length 1.
// The returned fonts are automatically wrapped in a text.FontFace with
// inferred font metadata.
// BUG(whereswaldon): the only Variant that can be detected automatically is
// "Mono".
func ParseCollection(src []byte) ([]giofont.FontFace, error) {
	lds, err := loader.NewLoaders(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	out := make([]giofont.FontFace, len(lds))
	for i, ld := range lds {
		face, md, err := parseLoader(ld)
		if err != nil {
			return nil, fmt.Errorf("reading font %d of collection: %s", i, err)
		}
		ff := Face{
			face: face,
			font: md,
		}
		out[i] = giofont.FontFace{
			Face: ff,
			Font: ff.Font(),
		}
	}

	return out, nil
}

func DescriptionToFont(md metadata.Description) giofont.Font {
	return giofont.Font{
		Typeface: giofont.Typeface(md.Family),
		Style:    gioStyle(md.Aspect.Style),
		Weight:   gioWeight(md.Aspect.Weight),
	}
}

func FontToDescription(font giofont.Font) metadata.Description {
	return metadata.Description{
		Family: string(font.Typeface),
		Aspect: metadata.Aspect{
			Style:  mdStyle(font.Style),
			Weight: mdWeight(font.Weight),
		},
	}
}

// parseLoader parses the contents of the loader into a face and its metadata.
func parseLoader(ld *loader.Loader) (font.Font, giofont.Font, error) {
	ft, err := fontapi.NewFont(ld)
	if err != nil {
		return nil, giofont.Font{}, err
	}
	data := DescriptionToFont(metadata.Metadata(ld))
	return ft, data, nil
}

// Face returns a thread-unsafe wrapper for this Face suitable for use by a single shaper.
// Face many be invoked any number of times and is safe so long as each return value is
// only used by one goroutine.
func (f Face) Face() font.Face {
	return &fontapi.Face{Font: f.face}
}

// FontFace returns a text.Font with populated font metadata for the
// font.
// BUG(whereswaldon): the only Variant that can be detected automatically is
// "Mono".
func (f Face) Font() giofont.Font {
	return f.font
}

func gioStyle(s metadata.Style) giofont.Style {
	switch s {
	case metadata.StyleItalic:
		return giofont.Italic
	case metadata.StyleNormal:
		fallthrough
	default:
		return giofont.Regular
	}
}

func mdStyle(g giofont.Style) metadata.Style {
	switch g {
	case giofont.Italic:
		return metadata.StyleItalic
	case giofont.Regular:
		fallthrough
	default:
		return metadata.StyleNormal
	}
}

func gioWeight(w metadata.Weight) giofont.Weight {
	switch w {
	case metadata.WeightThin:
		return giofont.Thin
	case metadata.WeightExtraLight:
		return giofont.ExtraLight
	case metadata.WeightLight:
		return giofont.Light
	case metadata.WeightNormal:
		return giofont.Normal
	case metadata.WeightMedium:
		return giofont.Medium
	case metadata.WeightSemibold:
		return giofont.SemiBold
	case metadata.WeightBold:
		return giofont.Bold
	case metadata.WeightExtraBold:
		return giofont.ExtraBold
	case metadata.WeightBlack:
		return giofont.Black
	default:
		return giofont.Normal
	}
}

func mdWeight(g giofont.Weight) metadata.Weight {
	switch g {
	case giofont.Thin:
		return metadata.WeightThin
	case giofont.ExtraLight:
		return metadata.WeightExtraLight
	case giofont.Light:
		return metadata.WeightLight
	case giofont.Normal:
		return metadata.WeightNormal
	case giofont.Medium:
		return metadata.WeightMedium
	case giofont.SemiBold:
		return metadata.WeightSemibold
	case giofont.Bold:
		return metadata.WeightBold
	case giofont.ExtraBold:
		return metadata.WeightExtraBold
	case giofont.Black:
		return metadata.WeightBlack
	default:
		return metadata.WeightNormal
	}
}
