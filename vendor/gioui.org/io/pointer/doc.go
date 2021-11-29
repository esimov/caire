// SPDX-License-Identifier: Unlicense OR MIT

/*
Package pointer implements pointer events and operations.
A pointer is either a mouse controlled cursor or a touch
object such as a finger.

The InputOp operation is used to declare a handler ready for pointer
events. Use an event.Queue to receive events.

Types

Only events that match a specified list of types are delivered to a handler.

For example, to receive Press, Drag, and Release events (but not Move, Enter,
Leave, or Scroll):

	var ops op.Ops
	var h *Handler = ...

	pointer.InputOp{
		Tag:   h,
		Types: pointer.Press | pointer.Drag | pointer.Release,
	}.Add(ops)

Cancel events are always delivered.

Hit areas

Clip operations from package op/clip are used for specifying
hit areas where subsequent InputOps are active.

For example, to set up a handler with a rectangular hit area:

	r := image.Rectangle{...}
	area := clip.Rect(r).Push(ops)
	pointer.InputOp{Tag: h}.Add(ops)
	area.Pop()

Note that hit areas behave similar to painting: the effective area of a stack
of multiple area operations is the intersection of the areas.

BUG: Clip operations other than clip.Rect and clip.Ellipse are approximated
with their bounding boxes.

Matching events

Areas form an implicit tree, with input handlers as leaves. The children of
an area is every area and handler added between its Push and corresponding Pop.

For example:

	ops := new(op.Ops)
	var h1, h2 *Handler

	area := clip.Rect(...).Push(ops)
	pointer.InputOp{Tag: h1}.Add(Ops)
	area.Pop()

	area := clip.Rect(...).Push(ops)
	pointer.InputOp{Tag: h2}.Add(ops)
	area.Pop()

implies a tree of two inner nodes, each with one pointer handler attached.

The matching proceeds as follows.

First, the foremost area that contains the event is found. Only areas whose
parent areas all contain the event is considered.

Then, every handler attached to the area is matched with the event.

If all attached handlers are marked pass-through or if no handlers are
attached, the matching repeats with the next foremost (sibling) area. Otherwise
the matching repeats with the parent area.

In the example above, all events will go to h2 because it and h1 are siblings
and none are pass-through.

Pass-through

The PassOp operations controls the pass-through setting. All handlers added
inside one or more PassOp scopes are marked pass-through.

Pass-through is useful for overlay widgets. Consider a hidden side drawer: when
the user touches the side, both the (transparent) drawer handle and the
interface below should receive pointer events. This effect is achieved by
marking the drawer handle pass-through.

Disambiguation

When more than one handler matches a pointer event, the event queue
follows a set of rules for distributing the event.

As long as the pointer has not received a Press event, all
matching handlers receive all events.

When a pointer is pressed, the set of matching handlers is
recorded. The set is not updated according to the pointer position
and hit areas. Rather, handlers stay in the matching set until they
no longer appear in a InputOp or when another handler in the set
grabs the pointer.

A handler can exclude all other handler from its matching sets
by setting the Grab flag in its InputOp. The Grab flag is sticky
and stays in effect until the handler no longer appears in any
matching sets.

The losing handlers are notified by a Cancel event.

For multiple grabbing handlers, the foremost handler wins.

Priorities

Handlers know their position in a matching set of a pointer through
event priorities. The Shared priority is for matching sets with
multiple handlers; the Grabbed priority indicate exclusive access.

Priorities are useful for deferred gesture matching.

Consider a scrollable list of clickable elements. When the user touches an
element, it is unknown whether the gesture is a click on the element
or a drag (scroll) of the list. While the click handler might light up
the element in anticipation of a click, the scrolling handler does not
scroll on finger movements with lower than Grabbed priority.

Should the user release the finger, the click handler registers a click.

However, if the finger moves beyond a threshold, the scrolling handler
determines that the gesture is a drag and sets its Grab flag. The
click handler receives a Cancel (removing the highlight) and further
movements for the scroll handler has priority Grabbed, scrolling the
list.
*/
package pointer
