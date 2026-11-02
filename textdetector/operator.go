package textdetector

// Operator is the generic interface for edge detection.
// Whatever edge detection algorithm can be plugged in which satisfies this method definition.
// We are going to focus on the Canny Edge detector, because it's much accurate and the
// original paper is also relying on this method.
type Operator interface {
	EdgeDetect(grayscale []uint8) ([]bool, []gradient, error)

	ContainsEdges(edges []bool) bool
}
