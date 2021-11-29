// SPDX-License-Identifier: Unlicense OR MIT

package f32color

import (
	"image/color"
	"math"
)

// RGBA is a 32 bit floating point linear premultiplied color space.
type RGBA struct {
	R, G, B, A float32
}

// Array returns rgba values in a [4]float32 array.
func (rgba RGBA) Array() [4]float32 {
	return [4]float32{rgba.R, rgba.G, rgba.B, rgba.A}
}

// Float32 returns r, g, b, a values.
func (col RGBA) Float32() (r, g, b, a float32) {
	return col.R, col.G, col.B, col.A
}

// SRGBA converts from linear to sRGB color space.
func (col RGBA) SRGB() color.NRGBA {
	if col.A == 0 {
		return color.NRGBA{}
	}
	return color.NRGBA{
		R: uint8(linearTosRGB(col.R/col.A)*255 + .5),
		G: uint8(linearTosRGB(col.G/col.A)*255 + .5),
		B: uint8(linearTosRGB(col.B/col.A)*255 + .5),
		A: uint8(col.A*255 + .5),
	}
}

// Luminance calculates the relative luminance of a linear RGBA color.
// Normalized to 0 for black and 1 for white.
//
// See https://www.w3.org/TR/WCAG20/#relativeluminancedef for more details
func (col RGBA) Luminance() float32 {
	return 0.2126*col.R + 0.7152*col.G + 0.0722*col.B
}

// Opaque returns the color without alpha component.
func (col RGBA) Opaque() RGBA {
	col.A = 1.0
	return col
}

// LinearFromSRGB converts from col in the sRGB colorspace to RGBA.
func LinearFromSRGB(col color.NRGBA) RGBA {
	af := float32(col.A) / 0xFF
	return RGBA{
		R: sRGBToLinear(float32(col.R)/0xff) * af,
		G: sRGBToLinear(float32(col.G)/0xff) * af,
		B: sRGBToLinear(float32(col.B)/0xff) * af,
		A: af,
	}
}

// NRGBAToRGBA converts from non-premultiplied sRGB color to premultiplied sRGB color.
//
// Each component in the result is `sRGBToLinear(c * alpha)`, where `c`
// is the linear color.
func NRGBAToRGBA(col color.NRGBA) color.RGBA {
	if col.A == 0xFF {
		return color.RGBA(col)
	}
	c := LinearFromSRGB(col)
	return color.RGBA{
		R: uint8(linearTosRGB(c.R)*255 + .5),
		G: uint8(linearTosRGB(c.G)*255 + .5),
		B: uint8(linearTosRGB(c.B)*255 + .5),
		A: col.A,
	}
}

// NRGBAToLinearRGBA converts from non-premultiplied sRGB color to premultiplied linear RGBA color.
//
// Each component in the result is `c * alpha`, where `c` is the linear color.
func NRGBAToLinearRGBA(col color.NRGBA) color.RGBA {
	if col.A == 0xFF {
		return color.RGBA(col)
	}
	c := LinearFromSRGB(col)
	return color.RGBA{
		R: uint8(c.R*255 + .5),
		G: uint8(c.G*255 + .5),
		B: uint8(c.B*255 + .5),
		A: col.A,
	}
}

// RGBAToNRGBA converts from premultiplied sRGB color to non-premultiplied sRGB color.
func RGBAToNRGBA(col color.RGBA) color.NRGBA {
	if col.A == 0xFF {
		return color.NRGBA(col)
	}

	linear := RGBA{
		R: sRGBToLinear(float32(col.R) / 0xff),
		G: sRGBToLinear(float32(col.G) / 0xff),
		B: sRGBToLinear(float32(col.B) / 0xff),
		A: float32(col.A) / 0xff,
	}

	return linear.SRGB()
}

// linearTosRGB transforms color value from linear to sRGB.
func linearTosRGB(c float32) float32 {
	// Formula from EXT_sRGB.
	switch {
	case c <= 0:
		return 0
	case 0 < c && c < 0.0031308:
		return 12.92 * c
	case 0.0031308 <= c && c < 1:
		return 1.055*float32(math.Pow(float64(c), 0.41666)) - 0.055
	}

	return 1
}

// sRGBToLinear transforms color value from sRGB to linear.
func sRGBToLinear(c float32) float32 {
	// Formula from EXT_sRGB.
	if c <= 0.04045 {
		return c / 12.92
	} else {
		return float32(math.Pow(float64((c+0.055)/1.055), 2.4))
	}
}

// MulAlpha applies the alpha to the color.
func MulAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = uint8(uint32(c.A) * uint32(alpha) / 0xFF)
	return c
}

// Disabled blends color towards the luminance and multiplies alpha.
// Blending towards luminance will desaturate the color.
// Multiplying alpha blends the color together more with the background.
func Disabled(c color.NRGBA) (d color.NRGBA) {
	const r = 80 // blend ratio
	lum := approxLuminance(c)
	return color.NRGBA{
		R: byte((int(c.R)*r + int(lum)*(256-r)) / 256),
		G: byte((int(c.G)*r + int(lum)*(256-r)) / 256),
		B: byte((int(c.B)*r + int(lum)*(256-r)) / 256),
		A: byte(int(c.A) * (128 + 32) / 256),
	}
}

// Hovered blends color towards a brighter color.
func Hovered(c color.NRGBA) (d color.NRGBA) {
	const r = 0x20 // lighten ratio
	return color.NRGBA{
		R: byte(255 - int(255-c.R)*(255-r)/256),
		G: byte(255 - int(255-c.G)*(255-r)/256),
		B: byte(255 - int(255-c.B)*(255-r)/256),
		A: c.A,
	}
}

// approxLuminance is a fast approximate version of RGBA.Luminance.
func approxLuminance(c color.NRGBA) byte {
	const (
		r = 13933 // 0.2126 * 256 * 256
		g = 46871 // 0.7152 * 256 * 256
		b = 4732  // 0.0722 * 256 * 256
		t = r + g + b
	)
	return byte((r*int(c.R) + g*int(c.G) + b*int(c.B)) / t)
}
