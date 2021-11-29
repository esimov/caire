// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
)

// packer packs a set of many smaller rectangles into
// much fewer larger atlases.
type packer struct {
	maxDims image.Point
	spaces  []image.Rectangle

	sizes []image.Point
	pos   image.Point
}

type placement struct {
	Idx int
	Pos image.Point
}

// add adds the given rectangle to the atlases and
// return the allocated position.
func (p *packer) add(s image.Point) (placement, bool) {
	if place, ok := p.tryAdd(s); ok {
		return place, true
	}
	p.newPage()
	return p.tryAdd(s)
}

func (p *packer) clear() {
	p.sizes = p.sizes[:0]
	p.spaces = p.spaces[:0]
}

func (p *packer) newPage() {
	p.pos = image.Point{}
	p.sizes = append(p.sizes, image.Point{})
	p.spaces = p.spaces[:0]
	p.spaces = append(p.spaces, image.Rectangle{
		Max: image.Point{X: 1e6, Y: 1e6},
	})
}

func (p *packer) tryAdd(s image.Point) (placement, bool) {
	var (
		bestIdx   = -1
		bestSpace image.Rectangle
		bestSize  = p.maxDims
	)
	// Go backwards to prioritize smaller spaces.
	for i, space := range p.spaces {
		rightSpace := space.Dx() - s.X
		bottomSpace := space.Dy() - s.Y
		if rightSpace < 0 || bottomSpace < 0 {
			continue
		}
		idx := len(p.sizes) - 1
		size := p.sizes[idx]
		if x := space.Min.X + s.X; x > size.X {
			if x > p.maxDims.X {
				continue
			}
			size.X = x
		}
		if y := space.Min.Y + s.Y; y > size.Y {
			if y > p.maxDims.Y {
				continue
			}
			size.Y = y
		}
		if size.X*size.Y < bestSize.X*bestSize.Y {
			bestIdx = i
			bestSpace = space
			bestSize = size
		}
	}
	if bestIdx == -1 {
		return placement{}, false
	}
	// Remove space.
	p.spaces[bestIdx] = p.spaces[len(p.spaces)-1]
	p.spaces = p.spaces[:len(p.spaces)-1]
	// Put s in the top left corner and add the (at most)
	// two smaller spaces.
	pos := bestSpace.Min
	if rem := bestSpace.Dy() - s.Y; rem > 0 {
		p.spaces = append(p.spaces, image.Rectangle{
			Min: image.Point{X: pos.X, Y: pos.Y + s.Y},
			Max: image.Point{X: bestSpace.Max.X, Y: bestSpace.Max.Y},
		})
	}
	if rem := bestSpace.Dx() - s.X; rem > 0 {
		p.spaces = append(p.spaces, image.Rectangle{
			Min: image.Point{X: pos.X + s.X, Y: pos.Y},
			Max: image.Point{X: bestSpace.Max.X, Y: pos.Y + s.Y},
		})
	}
	idx := len(p.sizes) - 1
	p.sizes[idx] = bestSize
	return placement{Idx: idx, Pos: pos}, true
}
