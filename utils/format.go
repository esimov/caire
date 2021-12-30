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

// The message types used accross the CLI application.
const (
	DefaultMessage MessageType = iota
	SuccessMessage
	ErrorMessage
	StatusMessage
)

// Colors used accross the CLI application.
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
