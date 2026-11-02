package textdetector

import (
	"errors"
	"fmt"
	"image"
	"math"

	cimage "github.com/esimov/caire/image"
)

const unsetStroke = math.MaxFloat64

// Options controls text-region detection.
type Options struct {
	maxStrokeRatio       float64
	maxComponentAspect   float64
	maxStrokeStdDevRatio float64
	minComponentPixels   int
	minComponentHeight   int
	maxComponentHeight   int
	maxHeightRatio       float64
	maxMedianStrokeRatio float64
	maxHorizontalGap     float64
	maxVerticalOffset    float64
	minComponentsPerLine int
}

func defaultOptions() Options {
	return Options{
		minComponentHeight:   10,
		maxComponentHeight:   300,
		maxComponentAspect:   10.0,
		maxStrokeStdDevRatio: 0.5,
		maxHeightRatio:       2.0,
		maxMedianStrokeRatio: 2.0,
		maxHorizontalGap:     3.0,
		maxVerticalOffset:    0.6,
		minComponentsPerLine: 3,
		minComponentPixels:   8,
		maxStrokeRatio:       3.0,
	}
}

// Detect returns candidate text regions.
//
// SWT is executed in both directions so it can detect:
//   - dark text over a light background
//   - light text over a dark background
func Detect(src image.Image) ([]image.Rectangle, error) {
	if src == nil {
		return nil, errors.New("swt: nil image")
	}

	width, height := src.Bounds().Dx(), src.Bounds().Dy()
	canny, err := NewCannyDetecor(width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the Canny edge detector: %w", err)
	}

	gray := cimage.RgbToGrayscale(src)
	edges, gradients, _ := canny.EdgeDetect(gray)
	if !canny.ContainsEdges(edges) {
		return nil, nil
	}

	options := defaultOptions()

	component := &Component{
		Width:                canny.width,
		Height:               canny.height,
		MinComponentHeight:   options.minComponentHeight,
		MaxComponentHeight:   options.maxComponentHeight,
		MaxComponentAspect:   options.maxComponentAspect,
		MaxStrokeStdDevRatio: options.maxStrokeStdDevRatio,
		MinComponentsPerLine: options.minComponentsPerLine,
		MaxHeightRatio:       options.maxHeightRatio,
		MaxMedianStrokeRatio: options.maxMedianStrokeRatio,
		MaxVerticalOffset:    options.maxVerticalOffset,
		MaxHorizontalGap:     options.maxHorizontalGap,

		StrokeComponent: &StrokeComponent{
			MaxStrokeRatio:     options.maxStrokeRatio,
			MinComponentPixels: options.minComponentPixels,
		},
	}

	swt := &SWT{
		Canny:     canny,
		Component: component,
	}

	swt.NormalizeOptions()

	var strokeComponents []StrokeComponent

	// +1 follows the gradient.
	// -1 follows the inverse gradient.
	for _, direction := range []float64{1, -1} {
		strokes := swt.Transform(edges, gradients, direction)
		components := component.FindConnected(strokes)

		for _, candidate := range components {
			if ok := component.Describe(candidate, strokes); ok {
				strokeComponents = append(strokeComponents, *swt.StrokeComponent)
			}
		}
	}

	regions := component.Group(strokeComponents)

	return regions, nil
}

func (swt *SWT) NormalizeOptions() {
	defaults := defaultOptions()

	if swt.MaxStrokeRatio <= 1 {
		swt.MaxStrokeRatio = defaults.maxStrokeRatio
	}

	if swt.MaxComponentAspect <= 1 {
		swt.MaxComponentAspect = defaults.maxComponentAspect
	}

	if swt.MaxStrokeStdDevRatio <= 0 {
		swt.MaxStrokeStdDevRatio = defaults.maxStrokeStdDevRatio
	}

	if swt.MinComponentPixels <= 0 {
		swt.MinComponentPixels = defaults.minComponentPixels
	}

	if swt.MinComponentHeight <= 0 {
		swt.MinComponentHeight = defaults.minComponentHeight
	}

	if swt.MaxComponentHeight <= 0 {
		swt.MaxComponentHeight = min(defaults.maxComponentHeight, swt.Height)
	}

	if swt.MaxComponentHeight < swt.MinComponentHeight {
		swt.MaxComponentHeight = swt.Height
	}

	if swt.MaxHeightRatio <= 1 {
		swt.MaxHeightRatio = defaults.maxHeightRatio
	}

	if swt.MaxMedianStrokeRatio <= 1 {
		swt.MaxMedianStrokeRatio = defaults.maxMedianStrokeRatio
	}

	if swt.MaxHorizontalGap <= 0 {
		swt.MaxHorizontalGap = defaults.maxHorizontalGap
	}

	if swt.MaxVerticalOffset <= 0 {
		swt.MaxVerticalOffset = defaults.maxVerticalOffset
	}

	if swt.MinComponentsPerLine <= 0 {
		swt.MinComponentsPerLine = defaults.minComponentsPerLine
	}
}
