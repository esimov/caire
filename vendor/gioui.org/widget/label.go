// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"fmt"
	"image"

	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"

	"golang.org/x/image/math/fixed"
)

// Label is a widget for laying out and drawing text.
type Label struct {
	// Alignment specify the text alignment.
	Alignment text.Alignment
	// MaxLines limits the number of lines. Zero means no limit.
	MaxLines int
}

// screenPos describes a character position (in text line and column numbers,
// not pixels): Y = line number, X = rune column.
type screenPos image.Point

const inf = 1e6

func posIsAbove(lines []text.Line, pos combinedPos, y int) bool {
	line := lines[pos.lineCol.Y]
	return pos.y+line.Bounds.Max.Y.Ceil() < y
}

func posIsBelow(lines []text.Line, pos combinedPos, y int) bool {
	line := lines[pos.lineCol.Y]
	return pos.y+line.Bounds.Min.Y.Floor() > y
}

func clipLine(lines []text.Line, alignment text.Alignment, width int, clip image.Rectangle, linePos combinedPos) (start combinedPos, end combinedPos) {
	// Seek to first (potentially) visible column.
	lineIdx := linePos.lineCol.Y
	line := lines[lineIdx]
	// runeWidth is the width of the widest rune in line.
	runeWidth := (line.Bounds.Max.X - line.Width).Ceil()
	lineStart := fixed.I(clip.Min.X - runeWidth)
	lineEnd := fixed.I(clip.Max.X + runeWidth)

	flip := line.Layout.Direction.Progression() == system.TowardOrigin
	if flip {
		lineStart, lineEnd = lineEnd, lineStart
	}
	q := combinedPos{y: start.y, x: lineStart}
	start, _ = seekPosition(lines, alignment, width, linePos, q, 0)
	// Seek to first invisible column after start.
	q = combinedPos{y: start.y, x: lineEnd}
	end, _ = seekPosition(lines, alignment, width, start, q, 0)
	if flip {
		start, end = end, start
	}

	return start, end
}

func subLayout(line text.Line, start, end combinedPos) text.Layout {
	if start.lineCol.X == line.Layout.Runes.Count {
		return text.Layout{}
	}

	startCluster := clusterIndexFor(line, start.lineCol.X, start.clusterIndex)
	endCluster := clusterIndexFor(line, end.lineCol.X, end.clusterIndex)
	if startCluster > endCluster {
		startCluster, endCluster = endCluster, startCluster
	}
	return line.Layout.Slice(startCluster, endCluster)
}

func firstPos(line text.Line, alignment text.Alignment, width int) combinedPos {
	p := combinedPos{
		x: align(alignment, line.Layout.Direction, line.Width, width),
		y: line.Ascent.Ceil(),
	}

	if line.Layout.Direction.Progression() == system.TowardOrigin {
		p.x += line.Width
	}
	return p
}

func (p1 screenPos) Less(p2 screenPos) bool {
	return p1.Y < p2.Y || (p1.Y == p2.Y && p1.X < p2.X)
}

func (l Label) Layout(gtx layout.Context, s text.Shaper, font text.Font, size unit.Sp, txt string) layout.Dimensions {
	cs := gtx.Constraints
	textSize := fixed.I(gtx.Sp(size))
	lines := s.LayoutString(font, textSize, cs.Max.X, gtx.Locale, txt)
	if max := l.MaxLines; max > 0 && len(lines) > max {
		lines = lines[:max]
	}
	dims := linesDimens(lines)
	dims.Size = cs.Constrain(dims.Size)
	if len(lines) == 0 {
		return dims
	}
	cl := textPadding(lines)
	cl.Max = cl.Max.Add(dims.Size)
	defer clip.Rect(cl).Push(gtx.Ops).Pop()
	semantic.LabelOp(txt).Add(gtx.Ops)
	pos := firstPos(lines[0], l.Alignment, dims.Size.X)
	for !posIsBelow(lines, pos, cl.Max.Y) {
		start, end := clipLine(lines, l.Alignment, dims.Size.X, cl, pos)
		line := lines[start.lineCol.Y]
		lt := subLayout(line, start, end)

		off := image.Point{X: start.x.Floor(), Y: start.y}
		t := op.Offset(off).Push(gtx.Ops)
		op := clip.Outline{Path: s.Shape(font, textSize, lt)}.Op().Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		op.Pop()
		t.Pop()

		if pos.lineCol.Y == len(lines)-1 {
			break
		}
		pos, _ = seekPosition(lines, l.Alignment, dims.Size.X, pos, combinedPos{lineCol: screenPos{Y: pos.lineCol.Y + 1}}, 0)
	}
	return dims
}

func textPadding(lines []text.Line) (padding image.Rectangle) {
	if len(lines) == 0 {
		return
	}
	first := lines[0]
	if d := first.Ascent + first.Bounds.Min.Y; d < 0 {
		padding.Min.Y = d.Ceil()
	}
	last := lines[len(lines)-1]
	if d := last.Bounds.Max.Y - last.Descent; d > 0 {
		padding.Max.Y = d.Ceil()
	}
	if d := first.Bounds.Min.X; d < 0 {
		padding.Min.X = d.Ceil()
	}
	if d := first.Bounds.Max.X - first.Width; d > 0 {
		padding.Max.X = d.Ceil()
	}
	return
}

func linesDimens(lines []text.Line) layout.Dimensions {
	var width fixed.Int26_6
	var h int
	var baseline int
	if len(lines) > 0 {
		baseline = lines[0].Ascent.Ceil()
		var prevDesc fixed.Int26_6
		for _, l := range lines {
			h += (prevDesc + l.Ascent).Ceil()
			prevDesc = l.Descent
			if l.Width > width {
				width = l.Width
			}
		}
		h += lines[len(lines)-1].Descent.Ceil()
	}
	w := width.Ceil()
	return layout.Dimensions{
		Size: image.Point{
			X: w,
			Y: h,
		},
		Baseline: h - baseline,
	}
}

// align returns the x offset that should be applied to text with width so that it
// appears correctly aligned within a space of size maxWidth and with the primary
// text direction dir.
func align(align text.Alignment, dir system.TextDirection, width fixed.Int26_6, maxWidth int) fixed.Int26_6 {
	mw := fixed.I(maxWidth)
	if dir.Progression() == system.TowardOrigin {
		switch align {
		case text.Start:
			align = text.End
		case text.End:
			align = text.Start
		}
	}
	switch align {
	case text.Middle:
		return fixed.I(((mw - width) / 2).Floor())
	case text.End:
		return fixed.I((mw - width).Floor())
	case text.Start:
		return 0
	default:
		panic(fmt.Errorf("unknown alignment %v", align))
	}
}
