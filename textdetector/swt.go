package textdetector

import (
	"errors"
	"image"
	"math"

	cimage "github.com/esimov/caire/image"
)

const unsetStroke = math.MaxFloat64

// Options controls the text-region detection.
type Options struct {
	minComponentHeight   int
	maxComponentHeight   int
	maxComponentAspect   float64
	maxStrokeStdDevRatio float64
	maxHeightRatio       float64
	maxMedianStrokeRatio float64
	maxHorizontalGap     float64
	maxVerticalOffset    float64
	minComponentsPerLine int
	minComponentPixels   int
	maxStrokeRatio       float64
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
		minComponentPixels:   2,
		maxStrokeRatio:       2.0,
	}
}

// Detect returns candidate text regions.
//
// SWT is executed in both directions so it can detect:
//   - dark text over a light background
//   - light text over a dark background
func Detect(ed EdgeDetector, src image.Image) ([]image.Rectangle, error) {
	if src == nil {
		return nil, errors.New("swt: nil image")
	}

	width, heigth := src.Bounds().Dx(), src.Bounds().Dy()

	gray := cimage.RgbToGrayscale(src)
	edges, gradients, _ := ed.Detect(gray)
	if !ed.ContainsEdges(edges) {
		return nil, nil
	}

	options := defaultOptions()

	component := &Component{
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
		Width:     width,
		Height:    heigth,
		Component: component,
	}

	swt.NormalizeOptions()

	var strokeComponents []StrokeComponent

	// Apply the transformation twice:
	// - dark text on light background (+1 follows the gradient)
	// - light text on dark background (-1 follows the inverses gradient)
	for _, direction := range []float64{1, -1} {
		strokes := swt.Transform(edges, gradients, direction)
		connectedComps := swt.FindConnected(strokes)

		for _, candidate := range connectedComps {
			if ok := swt.Filter(candidate, strokes); ok {
				strokeComponents = append(strokeComponents, *swt.StrokeComponent)
			}
		}
	}

	regions := swt.Group(strokeComponents)

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
