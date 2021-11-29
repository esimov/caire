// SPDX-License-Identifier: Unlicense OR MIT

/*
Package layout implements layouts common to GUI programs.

Constraints and dimensions

Constraints and dimensions form the interface between layouts and
interface child elements. This package operates on Widgets, functions
that compute Dimensions from a a set of constraints for acceptable
widths and heights. Both the constraints and dimensions are maintained
in an implicit Context to keep the Widget declaration short.

For example, to add space above a widget:

	var gtx layout.Context

	// Configure a top inset.
	inset := layout.Inset{Top: unit.Dp(8), ...}
	// Use the inset to lay out a widget.
	inset.Layout(gtx, func() {
		// Lay out widget and determine its size given the constraints
		// in gtx.Constraints.
		...
		return layout.Dimensions{...}
	})

Note that the example does not generate any garbage even though the
Inset is transient. Layouts that don't accept user input are designed
to not escape to the heap during their use.

Layout operations are recursive: a child in a layout operation can
itself be another layout. That way, complex user interfaces can
be created from a few generic layouts.

This example both aligns and insets a child:

	inset := layout.Inset{...}
	inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		align := layout.Alignment(...)
		return align.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return widget.Layout(gtx, ...)
		})
	})

More complex layouts such as Stack and Flex lay out multiple children,
and stateful layouts such as List accept user input.

*/
package layout
