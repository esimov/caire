// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"io"
	"unicode/utf8"

	"golang.org/x/text/runes"
)

// editBuffer implements a gap buffer for text editing.
type editBuffer struct {
	// pos is the byte position for Read and ReadRune.
	pos int

	// The gap start and end in bytes.
	gapstart, gapend int
	text             []byte

	// changed tracks whether the buffer content
	// has changed since the last call to Changed.
	changed bool
}

var _ textSource = (*editBuffer)(nil)

const minSpace = 5

func (e *editBuffer) Changed() bool {
	c := e.changed
	e.changed = false
	return c
}

func (e *editBuffer) deleteRunes(caret, count int) (bytes int, runes int) {
	e.moveGap(caret, 0)
	for ; count < 0 && e.gapstart > 0; count++ {
		_, s := utf8.DecodeLastRune(e.text[:e.gapstart])
		e.gapstart -= s
		bytes += s
		runes++
		e.changed = e.changed || s > 0
	}
	for ; count > 0 && e.gapend < len(e.text); count-- {
		_, s := utf8.DecodeRune(e.text[e.gapend:])
		e.gapend += s
		e.changed = e.changed || s > 0
	}
	return
}

// moveGap moves the gap to the caret position. After returning,
// the gap is guaranteed to be at least space bytes long.
func (e *editBuffer) moveGap(caret, space int) {
	if e.gapLen() < space {
		if space < minSpace {
			space = minSpace
		}
		txt := make([]byte, int(e.Size())+space)
		// Expand to capacity.
		txt = txt[:cap(txt)]
		gaplen := len(txt) - int(e.Size())
		if caret > e.gapstart {
			copy(txt, e.text[:e.gapstart])
			copy(txt[caret+gaplen:], e.text[caret:])
			copy(txt[e.gapstart:], e.text[e.gapend:caret+e.gapLen()])
		} else {
			copy(txt, e.text[:caret])
			copy(txt[e.gapstart+gaplen:], e.text[e.gapend:])
			copy(txt[caret+gaplen:], e.text[caret:e.gapstart])
		}
		e.text = txt
		e.gapstart = caret
		e.gapend = e.gapstart + gaplen
	} else {
		if caret > e.gapstart {
			copy(e.text[e.gapstart:], e.text[e.gapend:caret+e.gapLen()])
		} else {
			copy(e.text[caret+e.gapLen():], e.text[caret:e.gapstart])
		}
		l := e.gapLen()
		e.gapstart = caret
		e.gapend = e.gapstart + l
	}
}

func (e *editBuffer) Size() int64 {
	return int64(len(e.text) - e.gapLen())
}

func (e *editBuffer) gapLen() int {
	return e.gapend - e.gapstart
}

func (e *editBuffer) ReadAt(p []byte, offset int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if offset == e.Size() {
		return 0, io.EOF
	}
	var total int
	if offset < int64(e.gapstart) {
		n := copy(p, e.text[offset:e.gapstart])
		p = p[n:]
		total += n
		offset += int64(n)
	}
	if offset >= int64(e.gapstart) {
		n := copy(p, e.text[offset+int64(e.gapLen()):])
		total += n
	}
	return total, nil
}

func (e *editBuffer) ReplaceRunes(byteOffset, runeCount int64, s string) {
	e.deleteRunes(int(byteOffset), int(runeCount))
	e.prepend(int(byteOffset), s)
}

func (e *editBuffer) prepend(caret int, s string) {
	if !utf8.ValidString(s) {
		s = runes.ReplaceIllFormed().String(s)
	}

	e.moveGap(caret, len(s))
	copy(e.text[caret:], s)
	e.gapstart += len(s)
	e.changed = e.changed || len(s) > 0
}
