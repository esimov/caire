// SPDX-License-Identifier: Unlicense OR MIT

package material

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/internal/f32color"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
)

type EditorStyle struct {
	Font font.Font
	// LineHeight controls the distance between the baselines of lines of text.
	// If zero, a sensible default will be used.
	LineHeight unit.Sp
	// LineHeightScale applies a scaling factor to the LineHeight. If zero, a
	// sensible default will be used.
	LineHeightScale float32
	TextSize        unit.Sp
	// Color is the text color.
	Color color.NRGBA
	// Hint contains the text displayed when the editor is empty.
	Hint string
	// HintColor is the color of hint text.
	HintColor color.NRGBA
	// SelectionColor is the color of the background for selected text.
	SelectionColor color.NRGBA
	Editor         *widget.Editor

	shaper *text.Shaper
}

func Editor(th *Theme, editor *widget.Editor, hint string) EditorStyle {
	return EditorStyle{
		Editor: editor,
		Font: font.Font{
			Typeface: th.Face,
		},
		TextSize:       th.TextSize,
		Color:          th.Palette.Fg,
		shaper:         th.Shaper,
		Hint:           hint,
		HintColor:      f32color.MulAlpha(th.Palette.Fg, 0xbb),
		SelectionColor: f32color.MulAlpha(th.Palette.ContrastBg, 0x60),
	}
}

func (e EditorStyle) Layout(gtx layout.Context) layout.Dimensions {
	// Choose colors.
	textColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: e.Color}.Add(gtx.Ops)
	textColor := textColorMacro.Stop()
	hintColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: e.HintColor}.Add(gtx.Ops)
	hintColor := hintColorMacro.Stop()
	selectionColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: blendDisabledColor(gtx.Queue == nil, e.SelectionColor)}.Add(gtx.Ops)
	selectionColor := selectionColorMacro.Stop()

	var maxlines int
	if e.Editor.SingleLine {
		maxlines = 1
	}

	macro := op.Record(gtx.Ops)
	tl := widget.Label{
		Alignment:       e.Editor.Alignment,
		MaxLines:        maxlines,
		LineHeight:      e.LineHeight,
		LineHeightScale: e.LineHeightScale,
	}
	dims := tl.Layout(gtx, e.shaper, e.Font, e.TextSize, e.Hint, hintColor)
	call := macro.Stop()

	if w := dims.Size.X; gtx.Constraints.Min.X < w {
		gtx.Constraints.Min.X = w
	}
	if h := dims.Size.Y; gtx.Constraints.Min.Y < h {
		gtx.Constraints.Min.Y = h
	}
	e.Editor.LineHeight = e.LineHeight
	e.Editor.LineHeightScale = e.LineHeightScale
	dims = e.Editor.Layout(gtx, e.shaper, e.Font, e.TextSize, textColor, selectionColor)
	if e.Editor.Len() == 0 {
		call.Add(gtx.Ops)
	}
	return dims
}

func blendDisabledColor(disabled bool, c color.NRGBA) color.NRGBA {
	if disabled {
		return f32color.Disabled(c)
	}
	return c
}
