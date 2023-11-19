// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
)

// Scrollbar holds the persistent state for an area that can
// display a scrollbar. In particular, it tracks the position of a
// viewport along a one-dimensional region of content. The viewport's
// position can be adjusted by drag operations along the display area,
// or by clicks within the display area.
//
// Scrollbar additionally detects when a scroll indicator region is
// hovered.
type Scrollbar struct {
	track, indicator gesture.Click
	drag             gesture.Drag
	delta            float32

	dragging   bool
	oldDragPos float32
}

// Layout updates the internal state of the scrollbar based on events
// since the previous call to Layout. The provided axis will be used to
// normalize input event coordinates and constraints into an axis-
// independent format. viewportStart is the position of the beginning
// of the scrollable viewport relative to the underlying content expressed
// as a value in the range [0,1]. viewportEnd is the position of the end
// of the viewport relative to the underlying content, also expressed
// as a value in the range [0,1]. For example, if viewportStart is 0.25
// and viewportEnd is .5, the viewport described by the scrollbar is
// currently showing the second quarter of the underlying content.
func (s *Scrollbar) Layout(gtx layout.Context, axis layout.Axis, viewportStart, viewportEnd float32) layout.Dimensions {
	// Calculate the length of the major axis of the scrollbar. This is
	// the length of the track within which pointer events occur, and is
	// used to scale those interactions.
	trackHeight := float32(axis.Convert(gtx.Constraints.Max).X)
	s.delta = 0

	centerOnClick := func(normalizedPos float32) {
		// When the user clicks on the scrollbar we center on that point, respecting the limits of the beginning and end
		// of the scrollbar.
		//
		// Centering gives a consistent experience whether the user clicks above or below the indicator.
		target := normalizedPos - (viewportEnd-viewportStart)/2
		s.delta += target - viewportStart
		if s.delta < -viewportStart {
			s.delta = -viewportStart
		} else if s.delta > 1-viewportEnd {
			s.delta = 1 - viewportEnd
		}
	}

	// Jump to a click in the track.
	for _, event := range s.track.Events(gtx) {
		if event.Type != gesture.TypeClick ||
			event.Modifiers != key.Modifiers(0) ||
			event.NumClicks > 1 {
			continue
		}
		pos := axis.Convert(image.Point{
			X: int(event.Position.X),
			Y: int(event.Position.Y),
		})
		normalizedPos := float32(pos.X) / trackHeight
		// Clicking on the indicator should not jump to that position on the track. The user might've just intended to
		// drag and changed their mind.
		if !(normalizedPos >= viewportStart && normalizedPos <= viewportEnd) {
			centerOnClick(normalizedPos)
		}
	}

	// Offset to account for any drags.
	for _, event := range s.drag.Events(gtx.Metric, gtx, gesture.Axis(axis)) {
		switch event.Type {
		case pointer.Drag:
		case pointer.Release, pointer.Cancel:
			s.dragging = false
			continue
		default:
			continue
		}
		dragOffset := axis.FConvert(event.Position).X
		// The user can drag outside of the constraints, or even the window. Limit dragging to within the scrollbar.
		if dragOffset < 0 {
			dragOffset = 0
		} else if dragOffset > trackHeight {
			dragOffset = trackHeight
		}
		normalizedDragOffset := dragOffset / trackHeight

		if !s.dragging {
			s.dragging = true
			s.oldDragPos = normalizedDragOffset

			if normalizedDragOffset < viewportStart || normalizedDragOffset > viewportEnd {
				// The user started dragging somewhere on the track that isn't covered by the indicator. Consider this a
				// click in addition to a drag and jump to the clicked point.
				//
				// TODO(dh): this isn't perfect. We only get the pointer.Drag event once the user has actually dragged,
				// which means that if the user presses the mouse button and neither releases it nor drags it, nothing
				// will happen.
				pos := axis.Convert(image.Point{
					X: int(event.Position.X),
					Y: int(event.Position.Y),
				})
				normalizedPos := float32(pos.X) / trackHeight
				centerOnClick(normalizedPos)
			}
		} else {
			s.delta += normalizedDragOffset - s.oldDragPos

			if viewportStart+s.delta < 0 {
				// Adjust normalizedDragOffset - and thus the future s.oldDragPos - so that futile dragging up has to be
				// countered with dragging down again. Otherwise, dragging up would have no effect, but dragging down would
				// immediately start scrolling. We want the user to undo their ineffective drag first.
				normalizedDragOffset -= viewportStart + s.delta
				// Limit s.delta to the maximum amount scrollable
				s.delta = -viewportStart
			} else if viewportEnd+s.delta > 1 {
				normalizedDragOffset += (1 - viewportEnd) - s.delta
				s.delta = 1 - viewportEnd
			}
			s.oldDragPos = normalizedDragOffset
		}
	}

	// Process events from the indicator so that hover is
	// detected properly.
	_ = s.indicator.Events(gtx)

	return layout.Dimensions{}
}

// AddTrack configures the track click listener for the scrollbar to use
// the current clip area.
func (s *Scrollbar) AddTrack(ops *op.Ops) {
	s.track.Add(ops)
}

// AddIndicator configures the indicator click listener for the scrollbar to use
// the current clip area.
func (s *Scrollbar) AddIndicator(ops *op.Ops) {
	s.indicator.Add(ops)
}

// AddDrag configures the drag listener for the scrollbar to use
// the current clip area.
func (s *Scrollbar) AddDrag(ops *op.Ops) {
	s.drag.Add(ops)
}

// IndicatorHovered reports whether the scroll indicator is currently being
// hovered by the pointer.
func (s *Scrollbar) IndicatorHovered() bool {
	return s.indicator.Hovered()
}

// TrackHovered reports whether the scroll track is being hovered by the
// pointer.
func (s *Scrollbar) TrackHovered() bool {
	return s.track.Hovered()
}

// ScrollDistance returns the normalized distance that the scrollbar
// moved during the last call to Layout as a value in the range [-1,1].
func (s *Scrollbar) ScrollDistance() float32 {
	return s.delta
}

// Dragging reports whether the user is currently performing a drag gesture
// on the indicator. Note that this can return false while ScrollDistance is nonzero
// if the user scrolls using a different control than the scrollbar (like a mouse
// wheel).
func (s *Scrollbar) Dragging() bool {
	return s.dragging
}

// List holds the persistent state for a layout.List that has a
// scrollbar attached.
type List struct {
	Scrollbar
	layout.List
}
