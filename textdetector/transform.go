package textdetector

import (
	"math"
	"sort"
)

type SWT struct {
	Operator
	Width  int
	Height int
	*Component
}

type ray struct {
	pixels []point
	width  float64
}

func (s *SWT) Transform(edges []bool, gradients []gradient, direction float64) []float64 {
	strokes := makeFloatBuffer(s.Width*s.Height, math.MaxFloat64)

	var rays []ray

	for y := 1; y < s.Height-1; y++ {
		rowOffset := y * s.Width

		for x := 1; x < s.Width-1; x++ {
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
			pixels := s.traceRay(x, y, dx, dy)
			if len(pixels) < 2 {
				continue
			}

			for i := 1; i < len(pixels); i++ {
				end := pixels[i]
				endIndex := end.y*s.Width + end.x

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
						index := p.y*s.Width + p.x

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
			value := strokes[p.y*s.Width+p.x]

			if value != unsetStroke {
				values = append(values, value)
			}
		}

		if len(values) == 0 {
			continue
		}

		medianValue := calculateMedian(values)
		for _, p := range candidateRay.pixels {
			index := p.y*s.Width + p.x

			if medianValue < strokes[index] {
				strokes[index] = medianValue
			}
		}
	}

	return strokes
}

// traceRay cast a ray repeatedly from startX and startY perpendicular to the
// orientation of a stroke until it reaches the ray max length.
func (s *SWT) traceRay(startX int, startY int, dx float64, dy float64) []point {
	width, height := s.Width, s.Height
	result := []point{{
		x: startX,
		y: startY,
	}}
	maxRayLength := max(width, height) / 3

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

func makeFloatBuffer(length int, initial float64) []float64 {
	result := make([]float64, length)

	if initial != 0 {
		for i := range result {
			result[i] = initial
		}
	}

	return result
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64{}, values...)
	sort.Float64s(sorted)

	middle := len(sorted) / 2

	if len(sorted)%2 == 0 {
		return (sorted[middle-1] + sorted[middle]) / 2
	}

	return sorted[middle]
}
