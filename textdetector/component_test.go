package textdetector

import (
	"image"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSWT_FindConnectedUsesDiagonalConnectivity(t *testing.T) {
	s := &SWT{Component: &Component{StrokeComponent: &StrokeComponent{MaxStrokeRatio: 3}}}
	strokes := []float64{
		2, unsetStroke, unsetStroke,
		unsetStroke, 2, unsetStroke,
		unsetStroke, unsetStroke, 2,
	}

	components := s.FindConnected(3, 3, strokes)

	assert.Len(t, components, 1)
	assert.ElementsMatch(t, []point{{x: 0, y: 0}, {x: 1, y: 1}, {x: 2, y: 2}}, components[0])
}

func TestSWT_FindConnectedSeparatesDifferentStrokeWidths(t *testing.T) {
	s := &SWT{Component: &Component{StrokeComponent: &StrokeComponent{MaxStrokeRatio: 3}}}
	strokes := []float64{2, 10, 10}

	components := s.FindConnected(3, 1, strokes)

	assert.Len(t, components, 2)
	assert.ElementsMatch(t, []point{{x: 0, y: 0}}, components[0])
	assert.ElementsMatch(t, []point{{x: 1, y: 0}, {x: 2, y: 0}}, components[1])
}

func TestSWT_FindConnectedAcceptsStrokeRatioBoundary(t *testing.T) {
	s := &SWT{Component: &Component{StrokeComponent: &StrokeComponent{MaxStrokeRatio: 3}}}

	components := s.FindConnected(2, 1, []float64{2, 6})

	assert.Len(t, components, 1)
	assert.ElementsMatch(t, []point{{x: 0, y: 0}, {x: 1, y: 0}}, components[0])
}

func TestSWT_FindConnectedSkipsUnsetStrokes(t *testing.T) {
	s := &SWT{Component: &Component{StrokeComponent: &StrokeComponent{MaxStrokeRatio: 3}}}

	components := s.FindConnected(3, 2, []float64{
		unsetStroke, unsetStroke, unsetStroke,
		unsetStroke, unsetStroke, unsetStroke,
	})

	assert.Empty(t, components)
}

func TestSWT_FilterCreatesStrokeComponent(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent:      &StrokeComponent{MinComponentPixels: 4},
		MinComponentHeight:   2,
		MaxComponentHeight:   10,
		MaxComponentAspect:   10,
		MaxStrokeStdDevRatio: 0.5,
	}}
	pixels := []point{{x: 1, y: 1}, {x: 2, y: 1}, {x: 3, y: 1}, {x: 1, y: 2}, {x: 3, y: 2}, {x: 2, y: 3}}
	strokes := make([]float64, 25)
	for i := range strokes {
		strokes[i] = unsetStroke
	}
	for _, p := range pixels {
		strokes[p.y*5+p.x] = 2
	}

	component, ok := s.Filter(5, pixels, strokes)

	assert.True(t, ok)
	assert.NotNil(t, component)
	assert.Equal(t, image.Rect(1, 1, 4, 4), component.Bounds)
	assert.Equal(t, 3, component.StrokeWidth)
	assert.Equal(t, 3, component.StrokeHeigth)
	assert.Equal(t, 2.0, component.MeanStroke)
	assert.Equal(t, 2.0, component.MedianStroke)
	assert.Equal(t, 0.0, component.StrokeStdDev)
	assert.Equal(t, 2.0, component.CenterX)
	assert.Equal(t, 2.0, component.CenterY)
	assert.InDelta(t, 6.0/9.0, component.Density, 0.0001)
	assert.Equal(t, pixels, component.Pixels)
}

func TestSWT_FilterRejectsTooFewPixels(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent: &StrokeComponent{MinComponentPixels: 3}}}

	component, ok := s.Filter(3, []point{{x: 0, y: 0}, {x: 1, y: 0}}, []float64{2, 2, unsetStroke})

	assert.False(t, ok)
	assert.Nil(t, component)
}

func TestSWT_FilterRejectsInvalidStrokeValues(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent:      &StrokeComponent{MinComponentPixels: 2},
		MinComponentHeight:   1,
		MaxComponentHeight:   10,
		MaxComponentAspect:   10,
		MaxStrokeStdDevRatio: 1,
	}}
	pixels := []point{{x: 0, y: 0}, {x: 1, y: 0}}

	component, ok := s.Filter(2, pixels, []float64{unsetStroke, math.NaN()})

	assert.False(t, ok)
	assert.Nil(t, component)
}

func TestSWT_FilterRejectsInvalidHeight(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent:      &StrokeComponent{MinComponentPixels: 2},
		MinComponentHeight:   3,
		MaxComponentHeight:   10,
		MaxComponentAspect:   10,
		MaxStrokeStdDevRatio: 1,
	}}

	component, ok := s.Filter(3, []point{{x: 0, y: 0}, {x: 1, y: 0}}, []float64{2, 2, unsetStroke})

	assert.False(t, ok)
	assert.Nil(t, component)
}

func TestSWT_FilterRejectsWidthSmallerThanTwo(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent:      &StrokeComponent{MinComponentPixels: 2},
		MinComponentHeight:   1,
		MaxComponentHeight:   10,
		MaxComponentAspect:   10,
		MaxStrokeStdDevRatio: 1,
	}}

	component, ok := s.Filter(2, []point{{x: 0, y: 0}, {x: 0, y: 1}}, []float64{2, unsetStroke, 2, unsetStroke})

	assert.False(t, ok)
	assert.Nil(t, component)
}

func TestSWT_FilterRejectsExtremeAspectRatio(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent:      &StrokeComponent{MinComponentPixels: 2},
		MinComponentHeight:   1,
		MaxComponentHeight:   10,
		MaxComponentAspect:   2,
		MaxStrokeStdDevRatio: 1,
	}}

	component, ok := s.Filter(6, []point{{x: 0, y: 0}, {x: 5, y: 0}}, []float64{2, unsetStroke, unsetStroke, unsetStroke, unsetStroke, 2})

	assert.False(t, ok)
	assert.Nil(t, component)
}

func TestSWT_FilterRejectsHighStrokeVariance(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent:      &StrokeComponent{MinComponentPixels: 4},
		MinComponentHeight:   2,
		MaxComponentHeight:   10,
		MaxComponentAspect:   10,
		MaxStrokeStdDevRatio: 0.2,
	}}
	pixels := []point{{x: 0, y: 0}, {x: 1, y: 0}, {x: 0, y: 1}, {x: 1, y: 1}}

	component, ok := s.Filter(2, pixels, []float64{1, 10, 1, 10})

	assert.False(t, ok)
	assert.Nil(t, component)
}

func TestSWT_FilterRejectsSolidBlob(t *testing.T) {
	s := &SWT{Component: &Component{
		StrokeComponent:      &StrokeComponent{MinComponentPixels: 4},
		MinComponentHeight:   2,
		MaxComponentHeight:   10,
		MaxComponentAspect:   10,
		MaxStrokeStdDevRatio: 1,
	}}
	pixels := []point{{x: 0, y: 0}, {x: 1, y: 0}, {x: 0, y: 1}, {x: 1, y: 1}}

	component, ok := s.Filter(2, pixels, []float64{2, 2, 2, 2})

	assert.False(t, ok)
	assert.Nil(t, component)
}

func TestSWT_GroupHorizontalComponents(t *testing.T) {
	s := &SWT{Component: &Component{MinComponentsPerLine: 3, MaxHeightRatio: 2, MaxMedianStrokeRatio: 2, MaxOffset: 0.5, MaxGap: 2}}
	components := []StrokeComponent{
		{Bounds: image.Rect(0, 2, 3, 6), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4, CenterX: 1, CenterY: 3.5},
		{Bounds: image.Rect(4, 2, 7, 6), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4, CenterX: 5, CenterY: 3.5},
		{Bounds: image.Rect(8, 2, 11, 6), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4, CenterX: 9, CenterY: 3.5},
	}

	regions := s.Group(components, Horizontal)

	assert.Equal(t, []image.Rectangle{image.Rect(0, 2, 11, 6)}, regions)
}

func TestSWT_GroupVerticalComponents(t *testing.T) {
	s := &SWT{Component: &Component{MinComponentsPerLine: 3, MaxHeightRatio: 2, MaxMedianStrokeRatio: 2, MaxOffset: 0.5, MaxGap: 2}}
	components := []StrokeComponent{
		{Bounds: image.Rect(2, 0, 6, 3), MedianStroke: 2, StrokeWidth: 4, StrokeHeigth: 3, CenterX: 3.5, CenterY: 1},
		{Bounds: image.Rect(2, 4, 6, 7), MedianStroke: 2, StrokeWidth: 4, StrokeHeigth: 3, CenterX: 3.5, CenterY: 5},
		{Bounds: image.Rect(2, 8, 6, 11), MedianStroke: 2, StrokeWidth: 4, StrokeHeigth: 3, CenterX: 3.5, CenterY: 9},
	}

	regions := s.Group(components, Vertical)

	assert.Equal(t, []image.Rectangle{image.Rect(2, 0, 6, 11)}, regions)
}

func TestSWT_GroupRejectsInsufficientComponents(t *testing.T) {
	s := &SWT{Component: &Component{MinComponentsPerLine: 3, MaxHeightRatio: 2, MaxMedianStrokeRatio: 2, MaxOffset: 1, MaxGap: 2}}
	components := []StrokeComponent{
		{Bounds: image.Rect(0, 0, 3, 4), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4, CenterX: 1, CenterY: 1.5},
		{Bounds: image.Rect(4, 0, 7, 4), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4, CenterX: 5, CenterY: 1.5},
	}

	assert.Empty(t, s.Group(components, Horizontal))
}

func TestSWT_FindCompatibleRejectsDifferentHeights(t *testing.T) {
	s := &SWT{Component: &Component{MaxHeightRatio: 2, MaxMedianStrokeRatio: 2, MaxOffset: 1, MaxGap: 3}}
	a := StrokeComponent{Bounds: image.Rect(0, 0, 3, 4), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4, CenterY: 2}
	b := StrokeComponent{Bounds: image.Rect(4, 0, 7, 12), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 12, CenterY: 6}

	assert.False(t, s.findCompatible(a, b, Horizontal))
}

func TestSWT_FindCompatibleRejectsDifferentStrokeWidths(t *testing.T) {
	s := &SWT{Component: &Component{MaxHeightRatio: 2, MaxMedianStrokeRatio: 2, MaxOffset: 1, MaxGap: 3}}
	a := StrokeComponent{Bounds: image.Rect(0, 0, 3, 4), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4, CenterY: 2}
	b := StrokeComponent{Bounds: image.Rect(4, 0, 7, 4), MedianStroke: 6, StrokeWidth: 3, StrokeHeigth: 4, CenterY: 2}

	assert.False(t, s.findCompatible(a, b, Horizontal))
}

func TestSWT_FindCompatibleRejectsInvalidDirection(t *testing.T) {
	s := &SWT{Component: &Component{MaxHeightRatio: 2, MaxMedianStrokeRatio: 2, MaxOffset: 1, MaxGap: 3}}
	a := StrokeComponent{Bounds: image.Rect(0, 0, 3, 4), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4}
	b := StrokeComponent{Bounds: image.Rect(4, 0, 7, 4), MedianStroke: 2, StrokeWidth: 3, StrokeHeigth: 4}

	assert.False(t, s.findCompatible(a, b, TextDirection(100)))
}

func TestComponent_FilterDuplicateRegions(t *testing.T) {
	c := &Component{}
	regions := []image.Rectangle{
		image.Rect(0, 0, 10, 10),
		image.Rect(2, 2, 9, 9),
		image.Rect(20, 20, 25, 25),
	}

	result := c.filterDuplicateRegions(regions)

	assert.Equal(t, []image.Rectangle{image.Rect(0, 0, 10, 10), image.Rect(20, 20, 25, 25)}, result)
}

func TestRectangleArea(t *testing.T) {
	assert.Equal(t, 50, rectangleArea(image.Rect(0, 0, 10, 5)))
	assert.Equal(t, 0, rectangleArea(image.Rectangle{}))
}

func TestRectangleHorizontalGap(t *testing.T) {
	assert.Equal(t, 3, rectangleHorizontalGap(image.Rect(0, 0, 2, 2), image.Rect(5, 0, 7, 2)))
	assert.Equal(t, 3, rectangleHorizontalGap(image.Rect(5, 0, 7, 2), image.Rect(0, 0, 2, 2)))
	assert.Equal(t, 0, rectangleHorizontalGap(image.Rect(0, 0, 5, 5), image.Rect(4, 0, 7, 5)))
}

func TestRectangleVerticalGap(t *testing.T) {
	assert.Equal(t, 3, rectangleVerticalGap(image.Rect(0, 0, 2, 2), image.Rect(0, 5, 2, 7)))
	assert.Equal(t, 3, rectangleVerticalGap(image.Rect(0, 5, 2, 7), image.Rect(0, 0, 2, 2)))
	assert.Equal(t, 0, rectangleVerticalGap(image.Rect(0, 0, 5, 5), image.Rect(0, 4, 5, 7)))
}
