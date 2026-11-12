package textdetector

import "image"

// EdgeDetector is the generic interface for edge detection.
// Any kind of edge detection algorithm can be plugged in which satisfies this method definition.
// The Stroke Width Transform implementation uses the Canny Edge detector
// due to its higher accuracy and efficiency.
type EdgeDetector interface {
	// Detect returns the detected edges together with their gradient.
	Detect(grayscale []uint8) ([]bool, []gradient, error)

	// ContainsEdges check if there is any edge detected.
	ContainsEdges(edges []bool) bool
}

// TextDetector is the generic interface for text detection.
type TextDetector interface {
	// Detect returns the final detected text areas based on their components similarities.
	Detect(src *image.NRGBA) []image.Rectangle

	// FindConnected find the similar components
	FindConnected(width, height int, strokes []float64) [][]point

	// Filter is used to select the text component candidates with similar
	// characteristics, based on certain criteria provided as input parameters.
	Filter(width int, pixels []point, strokes []float64) (*StrokeComponent, bool)

	// Group groups together the connected components and present them as single image unit.
	Group(components []StrokeComponent, direction TextDirection) []image.Rectangle
}
