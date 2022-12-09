package utils

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"
)

// MessageType is a custom type used as a placeholder for various message types.
type MessageType int

// The message types used across the CLI application.
const (
	DefaultMessage MessageType = iota
	SuccessMessage
	ErrorMessage
	StatusMessage
)

// Colors used across the CLI application.
const (
	DefaultColor = "\x1b[0m"
	StatusColor  = "\x1b[36m"
	SuccessColor = "\x1b[32m"
	ErrorColor   = "\x1b[31m"
)

// DecorateText shows the message types in different colors.
func DecorateText(s string, msgType MessageType) string {
	switch msgType {
	case DefaultMessage:
		s = DefaultColor + s
	case StatusMessage:
		s = StatusColor + s
	case SuccessMessage:
		s = SuccessColor + s
	case ErrorMessage:
		s = ErrorColor + s
	default:
		return s
	}
	return s + DefaultColor
}

// FormatTime formats time.Duration output to a human readable value.
func FormatTime(d time.Duration) string {
	if d.Seconds() < 60.0 {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	if d.Minutes() < 60.0 {
		remainingSeconds := math.Mod(d.Seconds(), 60)
		return fmt.Sprintf("%dm %.2fs", int64(d.Minutes()), remainingSeconds)
	}
	if d.Hours() < 24.0 {
		remainingMinutes := math.Mod(d.Minutes(), 60)
		remainingSeconds := math.Mod(d.Seconds(), 60)
		return fmt.Sprintf("%dh %dm %.2fs",
			int64(d.Hours()), int64(remainingMinutes), remainingSeconds)
	}
	remainingHours := math.Mod(d.Hours(), 24)
	remainingMinutes := math.Mod(d.Minutes(), 60)
	remainingSeconds := math.Mod(d.Seconds(), 60)
	return fmt.Sprintf("%dd %dh %dm %.2fs",
		int64(d.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), remainingSeconds)
}

// HexToRGBA converts a color expressed as hexadecimal string to RGBA color.
func HexToRGBA(x string) color.NRGBA {
	var r, g, b, a uint8

	x = strings.TrimPrefix(x, "#")
	a = 255
	if len(x) == 2 {
		format := "%03x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b, &a)
	}
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// RGB returns color based on RGB in range 0..1
func RGB(r, g, b float32) color.NRGBA {
	return color.NRGBA{R: sat8(r), G: sat8(g), B: sat8(b), A: 0xFF}
}

// RGBA returns color based on RGBA in range 0..1
func RGBA(r, g, b, a float32) color.NRGBA {
	return color.NRGBA{R: sat8(r), G: sat8(g), B: sat8(b), A: sat8(a)}
}

// HSLA returns color based on HSLA in range 0..1
func HSLA(h, s, l, a float32) color.NRGBA { return RGBA(hsla(h, s, l, a)) }

// HSL returns color based on HSL in range 0..1
func HSL(h, s, l float32) color.NRGBA { return HSLA(h, s, l, 1) }

func hue(v1, v2, h float32) float32 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	if 6*h < 1 {
		return v1 + (v2-v1)*6*h
	} else if 2*h < 1 {
		return v2
	} else if 3*h < 2 {
		return v1 + (v2-v1)*(2.0/3.0-h)*6
	}

	return v1
}

func hsla(h, s, l, a float32) (r, g, b, ra float32) {
	if s == 0 {
		return l, l, l, a
	}

	h = mod32(h, 1)

	var v2 float32
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - s*l
	}

	v1 := 2*l - v2
	r = hue(v1, v2, h+1.0/3.0)
	g = hue(v1, v2, h)
	b = hue(v1, v2, h-1.0/3.0)
	ra = a

	return
}

func sat8(v float32) uint8 {
	v *= 255.0
	if v >= 255 {
		return 255
	} else if v <= 0 {
		return 0
	}
	return uint8(v)
}

func mod32(x, y float32) float32 { return float32(math.Mod(float64(x), float64(y))) }
