package utils

import (
	"fmt"
	"math"
	"net/http"
	"os"
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

// DetectFileContentType detects the file type by reading MIME type information of the file content.
func DetectFileContentType(fname string) (interface{}, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}

	// Reset the read pointer if necessary.
	file.Seek(0, 0)

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return string(contentType), nil
}
