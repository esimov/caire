package textdetector

import (
	"math"
)

type ray struct {
	pixels []point
	width  float64
}

// Transform starts from the detected Canny edge pixels and tries to produce
// meaningful character boundaries by rejecting isolated weak responses before SWT sees them.
// This is important because if the edge is broken SWT can miss valid rays,
// otherwise if there is too much noise, then it stops at the first encountered edge.
// This is why the hysteresis operation has an especially useful role.
func (s *SWT) Transform(width, height int, edges []bool, gradients []gradient, direction float64) []float64 {
	strokes := makeFloatBuffer(width*height, math.MaxFloat64)

	var rays []ray

	for y := 1; y < height-1; y++ {
		rowOffset := y * width

		for x := 1; x < width-1; x++ {
			startIndex := rowOffset + x

			// Skip if pixel is not an edge.
			if !edges[startIndex] {
				continue
			}

			startGradient := gradients[startIndex]
			if startGradient.x == 0 && startGradient.y == 0 {
				continue
			}

			// Get stroke distance.
			dx := startGradient.x * direction
			dy := startGradient.y * direction

			// Follow the ray until another edge is not found
			// or the ray reaches the predefined boundary limit.
			pixels := s.traceRay(width, height, x, y, dx, dy)
			if len(pixels) < 2 {
				continue
			}

			for i := 1; i < len(pixels); i++ {
				end := pixels[i]
				endIndex := end.y*width + end.x

				if !edges[endIndex] {
					continue
				}

				endGradient := gradients[endIndex]

				if s.areOpposingGradients(startGradient, endGradient, direction) {
					// Calculate the gradient distance between the start and end stroke.
					strokeWidth := math.Hypot(float64(end.x-x), float64(end.y-y))
					if strokeWidth <= 0 {
						break
					}

					validPixels := append([]point{}, pixels[:i+1]...)

					rays = append(rays, ray{
						pixels: validPixels,
						width:  strokeWidth,
					})

					for _, p := range validPixels {
						index := p.y*width + p.x

						if strokeWidth < strokes[index] {
							strokes[index] = strokeWidth
						}
					}
				}

				// Stop on the first encountered edge.
				break
			}
		}
	}

	// SWT median refinement pass.
	for _, candidateRay := range rays {
		values := make([]float64, 0, len(candidateRay.pixels))

		for _, p := range candidateRay.pixels {
			value := strokes[p.y*width+p.x]

			if value != unsetStroke {
				values = append(values, value)
			}
		}

		if len(values) == 0 {
			continue
		}

		medianValue := calculateMedian(values)
		for _, p := range candidateRay.pixels {
			index := p.y*width + p.x

			if medianValue < strokes[index] {
				strokes[index] = medianValue
			}
		}
	}

	return strokes
}

// traceRay cast a ray repeatedly from startX and startY perpendicular to the
// orientation of a stroke until it reaches the ray max length.
func (s *SWT) traceRay(width, height int, startX, startY int, dx, dy float64) []point {
	result := []point{{
		x: startX,
		y: startY,
	}}
	maxRayLength := min(max(width, height)/3, s.MaxRayLength)

	lastX := startX
	lastY := startY

	for step := 1; step <= maxRayLength; step++ {
		x := int(math.Round(float64(startX) + dx*float64(step)))
		y := int(math.Round(float64(startY) + dy*float64(step)))

		if x < 0 || x >= width || y < 0 || y >= height {
			break
		}
		if x == lastX && y == lastY {
			continue
		}

		result = append(result, point{x: x, y: y})

		lastX = x
		lastY = y
	}

	return result
}

func (s *SWT) areOpposingGradients(start gradient, end gradient, direction float64) bool {
	// The ray follows direction * startGradient.
	// The opposite boundary should point back toward the start.
	ax := start.x * direction
	ay := start.y * direction

	bx := end.x * direction
	by := end.y * direction

	dot := ax*bx + ay*by

	// cos(150 degrees) ~= -0.866.
	return dot <= -0.866
}

func calculateMedian(values []float64) float64 {
	n := len(values)

	if n == 0 {
		return 0
	}

	middle := n / 2

	if n%2 != 0 {
		return quickSelect(values, middle)
	}

	upper := quickSelect(values, middle)
	lower := quickSelect(values[:middle], middle-1)

	return (lower + upper) / 2
}

// quickSelect finds the k-th smallest element (1-based index)
func quickSelect(input []float64, k int) float64 {
	left := 0
	right := len(input) - 1

	for left <= right {
		pivotIdx := partition(input, left, right)

		if pivotIdx == k {
			return input[pivotIdx]
		} else if pivotIdx > k {
			right = pivotIdx - 1
		} else {
			left = pivotIdx + 1
		}
	}

	return 0
}

// partition rearranges elements around a pivot and returns its final index
func partition(input []float64, left, right int) int {
	pivot := input[right] // choose last element as pivot
	i := left             // index for smaller elements

	for j := left; j < right; j++ {
		if input[j] < pivot {
			input[i], input[j] = input[j], input[i]
			i++
		}
	}
	// place pivot in correct position
	input[i], input[right] = input[right], input[i]

	return i
}

func makeFloatBuffer(length int, initial float64) []float64 {
	result := make([]float64, length)

	if initial != 0 {
		for i := range result {
			result[i] = initial
		}
	}

	return result
}
