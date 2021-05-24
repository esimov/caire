package utils

import (
	"fmt"
	"math"
	"time"
)

// MessageType is a placeholder for the various the message types.
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
	ErrorColor   = "\x1b[31m"
	SuccessColor = "\x1b[32m"
	DefaultColor = "\x1b[0m"
	StatusColor  = "\x1b[36m"
)

// DecorateText shows the message types in different colors.
func DecorateText(s string, msgType MessageType) string {
	switch msgType {
	case SuccessMessage:
		s = SuccessColor + s
	case ErrorMessage:
		s = ErrorColor + s
	case DefaultMessage:
		s = DefaultColor + s
	case StatusMessage:
		s = StatusColor + s
	default:
		return s
	}
	return s + "\x1b[0m"
}

// FormatTime formats time.Duration output to a human readable value.
func FormatTime(d time.Duration) string {
	if d.Seconds() < 60.0 {
		return fmt.Sprintf("%ds", int64(d.Seconds()))
	}
	if d.Minutes() < 60.0 {
		remainingSeconds := math.Mod(d.Seconds(), 60)
		return fmt.Sprintf("%dm:%ds", int64(d.Minutes()), int64(remainingSeconds))
	}
	if d.Hours() < 24.0 {
		remainingMinutes := math.Mod(d.Minutes(), 60)
		remainingSeconds := math.Mod(d.Seconds(), 60)
		return fmt.Sprintf("%dh:%dm:%ds",
			int64(d.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}
	remainingHours := math.Mod(d.Hours(), 24)
	remainingMinutes := math.Mod(d.Minutes(), 60)
	remainingSeconds := math.Mod(d.Seconds(), 60)
	return fmt.Sprintf("%dd:%dh:%dm:%ds",
		int64(d.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
}
