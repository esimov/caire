// SPDX-License-Identifier: Unlicense OR MIT

package layout

import (
	"image"

	"gioui.org/op"
)

// Flex lays out child elements along an axis,
// according to alignment and weights.
type Flex struct {
	// Axis is the main axis, either Horizontal or Vertical.
	Axis Axis
	// Spacing controls the distribution of space left after
	// layout.
	Spacing Spacing
	// Alignment is the alignment in the cross axis.
	Alignment Alignment
	// WeightSum is the sum of weights used for the weighted
	// size of Flexed children. If WeightSum is zero, the sum
	// of all Flexed weights is used.
	WeightSum float32
}

// FlexChild is the descriptor for a Flex child.
type FlexChild struct {
	flex   bool
	weight float32

	widget Widget

	// Scratch space.
	call op.CallOp
	dims Dimensions
}

// Spacing determine the spacing mode for a Flex.
type Spacing uint8

const (
	// SpaceEnd leaves space at the end.
	SpaceEnd Spacing = iota
	// SpaceStart leaves space at the start.
	SpaceStart
	// SpaceSides shares space between the start and end.
	SpaceSides
	// SpaceAround distributes space evenly between children,
	// with half as much space at the start and end.
	SpaceAround
	// SpaceBetween distributes space evenly between children,
	// leaving no space at the start and end.
	SpaceBetween
	// SpaceEvenly distributes space evenly between children and
	// at the start and end.
	SpaceEvenly
)

// Rigid returns a Flex child with a maximal constraint of the
// remaining space.
func Rigid(widget Widget) FlexChild {
	return FlexChild{
		widget: widget,
	}
}

// Flexed returns a Flex child forced to take up weight fraction of the
// space left over from Rigid children. The fraction is weight
// divided by either the weight sum of all Flexed children or the Flex
// WeightSum if non zero.
func Flexed(weight float32, widget Widget) FlexChild {
	return FlexChild{
		flex:   true,
		weight: weight,
		widget: widget,
	}
}

// Layout a list of children. The position of the children are
// determined by the specified order, but Rigid children are laid out
// before Flexed children.
func (f Flex) Layout(gtx Context, children ...FlexChild) Dimensions {
	size := 0
	cs := gtx.Constraints
	mainMin, mainMax := f.Axis.mainConstraint(cs)
	crossMin, crossMax := f.Axis.crossConstraint(cs)
	remaining := mainMax
	var totalWeight float32
	cgtx := gtx
	// Lay out Rigid children.
	for i, child := range children {
		if child.flex {
			totalWeight += child.weight
			continue
		}
		macro := op.Record(gtx.Ops)
		cgtx.Constraints = f.Axis.constraints(0, remaining, crossMin, crossMax)
		dims := child.widget(cgtx)
		c := macro.Stop()
		sz := f.Axis.Convert(dims.Size).X
		size += sz
		remaining -= sz
		if remaining < 0 {
			remaining = 0
		}
		children[i].call = c
		children[i].dims = dims
	}
	if w := f.WeightSum; w != 0 {
		totalWeight = w
	}
	// fraction is the rounding error from a Flex weighting.
	var fraction float32
	flexTotal := remaining
	// Lay out Flexed children.
	for i, child := range children {
		if !child.flex {
			continue
		}
		var flexSize int
		if remaining > 0 && totalWeight > 0 {
			// Apply weight and add any leftover fraction from a
			// previous Flexed.
			childSize := float32(flexTotal) * child.weight / totalWeight
			flexSize = int(childSize + fraction + .5)
			fraction = childSize - float32(flexSize)
			if flexSize > remaining {
				flexSize = remaining
			}
		}
		macro := op.Record(gtx.Ops)
		cgtx.Constraints = f.Axis.constraints(flexSize, flexSize, crossMin, crossMax)
		dims := child.widget(cgtx)
		c := macro.Stop()
		sz := f.Axis.Convert(dims.Size).X
		size += sz
		remaining -= sz
		if remaining < 0 {
			remaining = 0
		}
		children[i].call = c
		children[i].dims = dims
	}
	var maxCross int
	var maxBaseline int
	for _, child := range children {
		if c := f.Axis.Convert(child.dims.Size).Y; c > maxCross {
			maxCross = c
		}
		if b := child.dims.Size.Y - child.dims.Baseline; b > maxBaseline {
			maxBaseline = b
		}
	}
	var space int
	if mainMin > size {
		space = mainMin - size
	}
	var mainSize int
	switch f.Spacing {
	case SpaceSides:
		mainSize += space / 2
	case SpaceStart:
		mainSize += space
	case SpaceEvenly:
		mainSize += space / (1 + len(children))
	case SpaceAround:
		if len(children) > 0 {
			mainSize += space / (len(children) * 2)
		}
	}
	for i, child := range children {
		dims := child.dims
		b := dims.Size.Y - dims.Baseline
		var cross int
		switch f.Alignment {
		case End:
			cross = maxCross - f.Axis.Convert(dims.Size).Y
		case Middle:
			cross = (maxCross - f.Axis.Convert(dims.Size).Y) / 2
		case Baseline:
			if f.Axis == Horizontal {
				cross = maxBaseline - b
			}
		}
		pt := f.Axis.Convert(image.Pt(mainSize, cross))
		trans := op.Offset(FPt(pt)).Push(gtx.Ops)
		child.call.Add(gtx.Ops)
		trans.Pop()
		mainSize += f.Axis.Convert(dims.Size).X
		if i < len(children)-1 {
			switch f.Spacing {
			case SpaceEvenly:
				mainSize += space / (1 + len(children))
			case SpaceAround:
				if len(children) > 0 {
					mainSize += space / len(children)
				}
			case SpaceBetween:
				if len(children) > 1 {
					mainSize += space / (len(children) - 1)
				}
			}
		}
	}
	switch f.Spacing {
	case SpaceSides:
		mainSize += space / 2
	case SpaceEnd:
		mainSize += space
	case SpaceEvenly:
		mainSize += space / (1 + len(children))
	case SpaceAround:
		if len(children) > 0 {
			mainSize += space / (len(children) * 2)
		}
	}
	sz := f.Axis.Convert(image.Pt(mainSize, maxCross))
	return Dimensions{Size: sz, Baseline: sz.Y - maxBaseline}
}

func (s Spacing) String() string {
	switch s {
	case SpaceEnd:
		return "SpaceEnd"
	case SpaceStart:
		return "SpaceStart"
	case SpaceSides:
		return "SpaceSides"
	case SpaceAround:
		return "SpaceAround"
	case SpaceBetween:
		return "SpaceAround"
	case SpaceEvenly:
		return "SpaceEvenly"
	default:
		panic("unreachable")
	}
}
