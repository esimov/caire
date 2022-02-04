// SPDX-License-Identifier: Unlicense OR MIT

// Package material implements the Material design.
//
// To maximize reusability and visual flexibility, user interface controls are
// split into two parts: the stateful widget and the stateless drawing of it.
//
// For example, widget.Clickable encapsulates the state and event
// handling of all clickable areas, while the Theme is responsible to
// draw a specific area, for example a button.
//
// This snippet defines a button that prints a message when clicked:
//
//     var gtx layout.Context
//     button := new(widget.Clickable)
//
//     for button.Clicked(gtx) {
//         fmt.Println("Clicked!")
//     }
//
// Use a Theme to draw the button:
//
//     theme := material.NewTheme(...)
//
//     material.Button(theme, "Click me!").Layout(gtx, button)
//
// Customization
//
// Quite often, a program needs to customize the theme-provided defaults. Several
// options are available, depending on the nature of the change.
//
// Mandatory parameters: Some parameters are not part of the widget state but
// have no obvious default. In the program above, the button text is a
// parameter to the Theme.Button method.
//
// Theme-global parameters: For changing the look of all widgets drawn with a
// particular theme, adjust the `Theme` fields:
//
//     theme.Color.Primary = color.NRGBA{...}
//
// Widget-local parameters: For changing the look of a particular widget,
// adjust the widget specific theme object:
//
//     btn := material.Button(theme, "Click me!")
//     btn.Font.Style = text.Italic
//     btn.Layout(gtx, button)
//
// Widget variants: A widget can have several distinct representations even
// though the underlying state is the same. A widget.Clickable can be drawn as a
// round icon button:
//
//     icon := material.NewIcon(...)
//
//     material.IconButton(theme, icon).Layout(gtx, button)
//
// Specialized widgets: Theme both define a generic Label method
// that takes a text size, and specialized methods for standard text
// sizes such as Theme.H1 and Theme.Body2.
package material
