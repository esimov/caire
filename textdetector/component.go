package textdetector

import (
	"image"
	"math"
	"sort"
)

type Component struct {
	*StrokeComponent
	MinComponentHeight   int
	MaxComponentHeight   int
	MaxComponentAspect   float64
	MaxStrokeStdDevRatio float64
	MinComponentsPerLine int
	MaxHeightRatio       float64
	MaxMedianStrokeRatio float64
	MaxVerticalOffset    float64
	MaxHorizontalGap     float64
}

type StrokeComponent struct {
	Bounds             image.Rectangle
	Pixels             []point
	MeanStroke         float64
	MedianStroke       float64
	StrokeStdDev       float64
	StrokeWidth        int
	StrokeHeigth       int
	CenterX            float64
	CenterY            float64
	Density            float64
	MaxStrokeRatio     float64
	MinComponentPixels int
}

// FindConnected finds the connected components based on the strokes similarities.
// Two gradient stroke can be considered as components if they are spatially connected
// and their stroke widths are sufficiently similar.
func (s *SWT) FindConnected(width, height int, strokes []float64) [][]point {
	visited := make([]bool, width*height)
	var components [][]point

	directions := [8]point{
		{x: -1, y: -1},
		{x: 0, y: -1},
		{x: 1, y: -1},
		{x: -1, y: 0},
		{x: 1, y: 0},
		{x: -1, y: 1},
		{x: 0, y: 1},
		{x: 1, y: 1},
	}

	for y := 0; y < height; y++ {
		rowOffset := y * width

		for x := 0; x < width; x++ {
			startIndex := rowOffset + x

			if visited[startIndex] || strokes[startIndex] == unsetStroke {
				continue
			}

			visited[startIndex] = true

			queue := []point{{x: x, y: y}}
			head := 0
			var pixels []point

			for head < len(queue) {
				current := queue[head]
				head++

				pixels = append(pixels, current)

				currentIndex := current.y*width + current.x
				currentStroke := strokes[currentIndex]

				for _, delta := range directions {
					nx := current.x + delta.x
					ny := current.y + delta.y

					if nx < 0 || nx >= width || ny < 0 || ny >= height {
						continue
					}

					neighborIndex := ny*width + nx

					if visited[neighborIndex] || strokes[neighborIndex] == unsetStroke {
						continue
					}

					neighborStroke := strokes[neighborIndex]

					minimumStroke := math.Min(currentStroke, neighborStroke)
					if minimumStroke <= 0 {
						continue
					}

					ratio := math.Max(currentStroke, neighborStroke) / minimumStroke
					if ratio > s.MaxStrokeRatio {
						continue
					}

					visited[neighborIndex] = true

					queue = append(queue, point{
						x: nx,
						y: ny,
					})
				}
			}
			components = append(components, pixels)
		}
	}

	return components
}

// Filter gets the connected component candidates, applies specific filtering criteria
// over the obtained results (based on certain criteria provided as input parameters).
// The final result, after all of the filtering criteria is passed, should be the stroke component.
func (s *SWT) Filter(imgWidth int, pixels []point, strokes []float64) (*StrokeComponent, bool) {
	if len(pixels) < s.MinComponentPixels {
		return nil, false
	}

	minX := math.MaxInt
	minY := math.MaxInt
	maxX := math.MinInt
	maxY := math.MinInt

	values := make([]float64, 0, len(pixels))
	var sum float64

	for _, p := range pixels {
		minX = min(minX, p.x)
		minY = min(minY, p.y)
		maxX = max(maxX, p.x)
		maxY = max(maxY, p.y)

		value := strokes[p.y*imgWidth+p.x]

		if value == unsetStroke || math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}

		values = append(values, value)
		sum += value
	}

	if len(values) == 0 {
		return nil, false
	}

	width := maxX - minX + 1
	height := maxY - minY + 1

	if height < s.MinComponentHeight || height > s.MaxComponentHeight {
		return nil, false
	}

	if width < 2 {
		return nil, false
	}

	aspect := math.Max(
		float64(width)/float64(height),
		float64(height)/float64(width),
	)

	if aspect > s.MaxComponentAspect {
		return nil, false
	}

	meanStroke := sum / float64(len(values))
	medianStroke := calculateMedian(values)

	var variance float64

	for _, value := range values {
		difference := value - meanStroke
		variance += difference * difference
	}

	variance /= float64(len(values))
	stdDev := math.Sqrt(variance)
	if meanStroke <= 0 || stdDev/meanStroke > s.MaxStrokeStdDevRatio {
		return nil, false
	}

	area := width * height
	density := float64(len(pixels)) / float64(area)

	// Reject extremely sparse or almost solid blobs.
	if density < 0.05 || density > 0.95 {
		return nil, false
	}

	return &StrokeComponent{
		Bounds:       image.Rect(minX, minY, maxX+1, maxY+1),
		Pixels:       pixels,
		MeanStroke:   meanStroke,
		MedianStroke: medianStroke,
		StrokeStdDev: stdDev,
		StrokeWidth:  width,
		StrokeHeigth: height,
		CenterX:      float64(minX+maxX) / 2,
		CenterY:      float64(minY+maxY) / 2,
		Density:      density,
	}, true
}

// Group groups the potential character candidates based on
// text alignment, similarities, stroke width and spacial distance.
// As a last step we filter out the overlapping rectangles.
func (s *SWT) Group(components []StrokeComponent) []image.Rectangle {
	var regions []image.Rectangle

	if len(components) == 0 {
		return nil
	}

	neighbours := make([][]int, len(components))

	for i := range components {
		for j := i + 1; j < len(components); j++ {
			if s.findCompatible(components[i], components[j]) {
				neighbours[i] = append(neighbours[i], j)
				neighbours[j] = append(neighbours[j], i)
			}
		}
	}

	visited := make([]bool, len(components))

	for i := range components {
		if visited[i] {
			continue
		}

		visited[i] = true
		queue := []int{i}
		head := 0
		var group []int

		for head < len(queue) {
			current := queue[head]
			head++

			group = append(group, current)

			for _, neighbor := range neighbours[current] {
				if visited[neighbor] {
					continue
				}

				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}

		if len(group) < s.MinComponentsPerLine {
			continue
		}

		region := components[group[0]].Bounds

		for _, index := range group[1:] {
			region = region.Union(components[index].Bounds)
		}

		if !region.Empty() {
			regions = append(regions, region)
		}
	}

	regions = s.filterDuplicateRegions(regions)

	sort.Slice(regions, func(i, j int) bool {
		if regions[i].Min.Y == regions[j].Min.Y {
			return regions[i].Min.X < regions[j].Min.X
		}
		return regions[i].Min.Y < regions[j].Min.Y
	})

	return regions
}

func (s *SWT) findCompatible(a, b StrokeComponent) bool {
	heightRatio := math.Max(
		float64(a.StrokeHeigth),
		float64(b.StrokeHeigth),
	) / math.Min(
		float64(a.StrokeHeigth),
		float64(b.StrokeHeigth),
	)

	if heightRatio > s.MaxHeightRatio {
		return false
	}

	if a.MedianStroke <= 0 || b.MedianStroke <= 0 {
		return false
	}

	strokeRatio := math.Max(a.MedianStroke, b.MedianStroke) / math.Min(a.MedianStroke, b.MedianStroke)
	if strokeRatio > s.MaxMedianStrokeRatio {
		return false
	}

	verticalOffset := math.Abs(a.CenterY - b.CenterY)
	maxVerticalOffset := float64(min(a.StrokeHeigth, b.StrokeHeigth)) * s.MaxVerticalOffset
	if verticalOffset > maxVerticalOffset {
		return false
	}

	horizontalGap := rectangleHorizontalGap(a.Bounds, b.Bounds)
	maxGap := float64(max(a.StrokeHeigth, b.StrokeHeigth)) * s.MaxHorizontalGap

	return float64(horizontalGap) <= maxGap
}

func (c *Component) filterDuplicateRegions(regions []image.Rectangle) []image.Rectangle {
	var result []image.Rectangle

	for _, candidate := range regions {
		merged := false

		for i := range result {
			intersection := candidate.Intersect(result[i])

			if intersection.Empty() {
				continue
			}

			intersectionArea := rectangleArea(intersection)

			smallerArea := min(rectangleArea(candidate), rectangleArea(result[i]))
			if smallerArea > 0 && float64(intersectionArea)/float64(smallerArea) >= 0.6 {
				result[i] = result[i].Union(candidate)
				merged = true
				break
			}
		}

		if !merged {
			result = append(result, candidate)
		}
	}

	return result
}

func rectangleArea(rect image.Rectangle) int {
	if rect.Empty() {
		return 0
	}

	return rect.Dx() * rect.Dy()
}

func rectangleHorizontalGap(a image.Rectangle, b image.Rectangle) int {
	if a.Max.X < b.Min.X {
		return b.Min.X - a.Max.X
	}

	if b.Max.X < a.Min.X {
		return a.Min.X - b.Max.X
	}

	return 0
}
