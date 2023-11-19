// SPDX-License-Identifier: Unlicense OR MIT

package material

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/widget"
)

type CheckBoxStyle struct {
	checkable
	CheckBox *widget.Bool
}

func CheckBox(th *Theme, checkBox *widget.Bool, label string) CheckBoxStyle {
	c := CheckBoxStyle{
		CheckBox: checkBox,
		checkable: checkable{
			Label:              label,
			Color:              th.Palette.Fg,
			IconColor:          th.Palette.ContrastBg,
			TextSize:           th.TextSize * 14.0 / 16.0,
			Size:               26,
			shaper:             th.Shaper,
			checkedStateIcon:   th.Icon.CheckBoxChecked,
			uncheckedStateIcon: th.Icon.CheckBoxUnchecked,
		},
	}
	c.checkable.Font.Typeface = th.Face
	return c
}

// Layout updates the checkBox and displays it.
func (c CheckBoxStyle) Layout(gtx layout.Context) layout.Dimensions {
	return c.CheckBox.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.CheckBox.Add(gtx.Ops)
		return c.layout(gtx, c.CheckBox.Value, c.CheckBox.Hovered() || c.CheckBox.Focused())
	})
}
