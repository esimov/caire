// Package imop implements the Porter-Duff composition operations
// used for mixing a graphic element with its backdrop.
// Porter and Duff presented in their paper 12 different composition operation,
// but the image/draw core package implements only the source-over-destination and source.
// This package is aimed to overcome the missing composite operations.

// It is mainly used to debug the seam carving operation correctness
// with face detection and image mask enabled.
// When the GUI mode and the debugging option is activated it will show
// the image mask and the detected faces rectangles in a distinct color.
package imop

import (
	"fmt"
	"math"
	"sort"

	"github.com/esimov/caire/utils"
)

const (
	Normal     = "normal"
	Darken     = "darken"
	Lighten    = "lighten"
	Multiply   = "multiply"
	Screen     = "screen"
	Overlay    = "overlay"
	SoftLight  = "soft_light"
	HardLight  = "hard_light"
	ColorDodge = "color_dodge"
	ColorBurn  = "color_burn"
	Difference = "difference"
	Exclusion  = "exclusion"

	// Non-separable blend modes
	Hue        = "hue"
	Saturation = "saturation"
	ColorMode  = "color"
	Luminosity = "luminosity"
)

// Blend struct contains the currently active blend mode and all the supported blend modes.
type Blend struct {
	Current string
	Modes   []string
}

// Color represents the RGB channel of a specific color.
type Color struct {
	R, G, B float64
}

// NewBlend initializes a new Blend.
func NewBlend() *Blend {
	return &Blend{
		Modes: []string{
			Normal,
			Darken,
			Lighten,
			Multiply,
			Screen,
			Overlay,
			SoftLight,
			HardLight,
			ColorDodge,
			ColorBurn,
			Difference,
			Exclusion,
			Hue,
			Saturation,
			ColorMode,
			Luminosity,
		},
	}
}

// Set activate one of the supported blend modes.
func (bl *Blend) Set(blendType string) error {
	if utils.Contains(bl.Modes, blendType) {
		bl.Current = blendType
		return nil
	}
	return fmt.Errorf("unsupported blend mode")
}

// Get returns the active blend mode.
func (bl *Blend) Get() string {
	return bl.Current
}

// Lum gets the luminosity of a color.
func (bl *Blend) Lum(rgb Color) float64 {
	return 0.3*rgb.R + 0.59*rgb.G + 0.11*rgb.B
}

// SetLum set the luminosity on a color.
func (bl *Blend) SetLum(rgb Color, l float64) Color {
	delta := l - bl.Lum(rgb)
	return bl.clip(Color{
		rgb.R + delta,
		rgb.G + delta,
		rgb.B + delta,
	})
}

// clip clips the channels of a color between certain min and max values.
func (bl *Blend) clip(rgb Color) Color {
	r, g, b := rgb.R, rgb.G, rgb.B

	l := bl.Lum(rgb)
	min := utils.Min(r, g, b)
	max := utils.Max(r, g, b)

	if min < 0 {
		r = l + (((r - l) * l) / (l - min))
		g = l + (((g - l) * l) / (l - min))
		b = l + (((b - l) * l) / (l - min))
	}
	if max > 1 {
		r = l + (((r - l) * (1 - l)) / (max - l))
		g = l + (((g - l) * (1 - l)) / (max - l))
		b = l + (((b - l) * (1 - l)) / (max - l))
	}

	return Color{R: r, G: g, B: b}
}

// Sat gets the saturation of a color.
func (bl *Blend) Sat(rgb Color) float64 {
	return utils.Max(rgb.R, rgb.G, rgb.B) - utils.Min(rgb.R, rgb.G, rgb.B)
}

// channel is a key/value struct pair used for sorting the color channels
// based on the color components having the minimum, middle, and maximum
// values upon entry to the function.
// The key component holds the channel name and val is the value it has.
type channel struct {
	key string
	val float64
}

func (bl *Blend) SetSat(rgb Color, s float64) Color {
	color := map[string]float64{
		"R": rgb.R,
		"G": rgb.G,
		"B": rgb.B,
	}
	channels := make([]channel, 0, 3)
	for k, v := range color {
		channels = append(channels, channel{k, v})
	}
	// Sort the color channels based on their values.
	sort.Slice(channels, func(i, j int) bool { return channels[i].val < channels[j].val })
	minChan, midChan, maxChan := channels[0].key, channels[1].key, channels[2].key

	if color[maxChan] > color[minChan] {
		color[midChan] = (((color[midChan] - color[minChan]) * s) / (color[maxChan] - color[minChan]))
		color[maxChan] = s
	} else {
		color[midChan], color[maxChan] = 0, 0
	}
	color[minChan] = 0

	return Color{
		R: color["R"],
		G: color["G"],
		B: color["B"],
	}
}

// Applies the alpha blending formula for a blend operation.
// See: https://www.w3.org/TR/compositing-1/#blending
func (bl *Blend) AlphaCompose(
	backdropAlpha,
	sourceAlpha,
	compositeAlpha,
	backdropColor,
	sourceColor,
	compositeColor float64,
) float64 {
	return ((1 - sourceAlpha/compositeAlpha) * backdropColor) +
		(sourceAlpha / compositeAlpha *
			math.Round((1-backdropAlpha)*sourceColor+backdropAlpha*compositeColor))
}
