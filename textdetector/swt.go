// Package textdetector is used to identify and group identical characters in image.
// The implementation is based on this article:
// https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/201020CVPR20TextDetection.pdf
// This package could be integrated into the seam removal logic to avoid
// removing seams from an image where the texts should be preserved.
package textdetector

import (
	"image"
	"math"

	cimage "github.com/esimov/caire/image"
)

const unsetStroke = math.MaxFloat64

type SWT struct {
	EdgeDetector
	*Component
	Direction TextDirection
}

var _ TextDetector = (*SWT)(nil)

// Options controls the text-region detection.
type Options struct {
	minComponentHeight   int
	maxComponentHeight   int
	maxRayLength         int
	maxComponentAspect   float64
	maxStrokeStdDevRatio float64
	maxHeightRatio       float64
	maxMedianStrokeRatio float64
	maxGap               float64
	maxOffset            float64
	minComponentsPerLine int
	minComponentPixels   int
	maxStrokeRatio       float64
}

func DefaultOptions() Options {
	return Options{
		minComponentHeight:   10,
		maxComponentHeight:   200,
		maxRayLength:         200,
		maxComponentAspect:   10.0,
		maxStrokeStdDevRatio: 0.5,
		maxHeightRatio:       2.0,
		maxMedianStrokeRatio: 2.0,
		maxGap:               2.5,
		maxOffset:            0.5,
		minComponentsPerLine: 3,
		minComponentPixels:   2,
		maxStrokeRatio:       2.0,
	}
}

func NewSWT(edgeDetector EdgeDetector, options Options) *SWT {
	normalizeOptions(&options)

	component := &Component{
		MinComponentHeight:   options.minComponentHeight,
		MaxComponentHeight:   options.maxComponentHeight,
		MaxRayLength:         options.maxRayLength,
		MaxComponentAspect:   options.maxComponentAspect,
		MaxStrokeStdDevRatio: options.maxStrokeStdDevRatio,
		MinComponentsPerLine: options.minComponentsPerLine,
		MaxHeightRatio:       options.maxHeightRatio,
		MaxMedianStrokeRatio: options.maxMedianStrokeRatio,
		MaxOffset:            options.maxOffset,
		MaxGap:               options.maxGap,

		StrokeComponent: &StrokeComponent{
			MaxStrokeRatio:     options.maxStrokeRatio,
			MinComponentPixels: options.minComponentPixels,
		},
	}

	return &SWT{
		Component:    component,
		EdgeDetector: edgeDetector,
	}
}

// Detect returns candidate text regions.
//
// SWT is executed in both directions so it can detect:
//   - dark text over a light background
//   - light text over a dark background
func (swt *SWT) Detect(src *image.NRGBA) []image.Rectangle {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	gray := cimage.RgbToGrayscale(src)
	edges, gradients, _ := swt.EdgeDetector.Detect(gray)

	if !swt.EdgeDetector.ContainsEdges(edges) {
		return nil
	}

	var strokeComponents []StrokeComponent

	// Apply the transformation twice:
	// - dark text on light background (+1 follows the gradient)
	// - light text on dark background (-1 follows the inverses gradient)
	for _, direction := range []float64{1, -1} {
		strokes := swt.Transform(width, height, edges, gradients, direction)
		connectedComps := swt.FindConnected(width, height, strokes)

		for _, candidate := range connectedComps {
			if component, ok := swt.Filter(width, candidate, strokes); ok {
				strokeComponents = append(strokeComponents, *component)
			}
		}
	}

	regions := swt.Group(strokeComponents, swt.Direction)

	return regions
}

func normalizeOptions(options *Options) {
	defaults := DefaultOptions()

	if options.maxStrokeRatio <= 1 {
		options.maxStrokeRatio = defaults.maxStrokeRatio
	}

	if options.maxComponentAspect <= 1 {
		options.maxComponentAspect = defaults.maxComponentAspect
	}

	if options.maxStrokeStdDevRatio <= 0 {
		options.maxStrokeStdDevRatio = defaults.maxStrokeStdDevRatio
	}

	if options.minComponentPixels <= 0 {
		options.minComponentPixels = defaults.minComponentPixels
	}

	if options.minComponentHeight <= 0 {
		options.minComponentHeight = defaults.minComponentHeight
	}

	if options.maxHeightRatio <= 1 {
		options.maxHeightRatio = defaults.maxHeightRatio
	}

	if options.maxMedianStrokeRatio <= 1 {
		options.maxMedianStrokeRatio = defaults.maxMedianStrokeRatio
	}

	if options.maxGap <= 0 {
		options.maxGap = defaults.maxGap
	}

	if options.maxOffset <= 0 {
		options.maxOffset = defaults.maxOffset
	}

	if options.minComponentsPerLine <= 0 {
		options.minComponentsPerLine = defaults.minComponentsPerLine
	}
}
