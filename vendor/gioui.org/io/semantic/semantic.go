// SPDX-License-Identifier: Unlicense OR MIT

// Package semantic provides operations for semantic descriptions of a user
// interface, to facilitate presentation and interaction in external software
// such as screen readers.
//
// Semantic descriptions are organized in a tree, with clip operations as
// nodes. Operations in this package are associated with the current semantic
// node, that is the most recent pushed clip operation.
package semantic

import (
	"gioui.org/internal/ops"
	"gioui.org/op"
)

// LabelOp provides the content of a textual component.
type LabelOp string

// DescriptionOp describes a component.
type DescriptionOp string

// ClassOp provides the component class.
type ClassOp int

const (
	Unknown ClassOp = iota
	Button
	CheckBox
	Editor
	RadioButton
	Switch
)

// SelectedOp describes the selected state for components that have
// boolean state.
type SelectedOp bool

// DisabledOp describes the disabled state.
type DisabledOp bool

func (l LabelOp) Add(o *op.Ops) {
	s := string(l)
	data := ops.Write1(&o.Internal, ops.TypeSemanticLabelLen, &s)
	data[0] = byte(ops.TypeSemanticLabel)
}

func (d DescriptionOp) Add(o *op.Ops) {
	s := string(d)
	data := ops.Write1(&o.Internal, ops.TypeSemanticDescLen, &s)
	data[0] = byte(ops.TypeSemanticDesc)
}

func (c ClassOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeSemanticClassLen)
	data[0] = byte(ops.TypeSemanticClass)
	data[1] = byte(c)
}

func (s SelectedOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeSemanticSelectedLen)
	data[0] = byte(ops.TypeSemanticSelected)
	if s {
		data[1] = 1
	}
}

func (d DisabledOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeSemanticDisabledLen)
	data[0] = byte(ops.TypeSemanticDisabled)
	if d {
		data[1] = 1
	}
}

func (c ClassOp) String() string {
	switch c {
	case Unknown:
		return "Unknown"
	case Button:
		return "Button"
	case CheckBox:
		return "CheckBox"
	case Editor:
		return "Editor"
	case RadioButton:
		return "RadioButton"
	case Switch:
		return "Switch"
	default:
		panic("invalid ClassOp")
	}
}
