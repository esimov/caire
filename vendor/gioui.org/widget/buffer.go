// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"io"
	"strings"
	"unicode/utf8"
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

const minSpace = 5

func (e *editBuffer) Changed() bool {
	c := e.changed
	e.changed = false
	return c
}

func (e *editBuffer) deleteRunes(caret, runes int) int {
	e.moveGap(caret, 0)
	for ; runes < 0 && e.gapstart > 0; runes++ {
		_, s := utf8.DecodeLastRune(e.text[:e.gapstart])
		e.gapstart -= s
		caret -= s
		e.changed = e.changed || s > 0
	}
	for ; runes > 0 && e.gapend < len(e.text); runes-- {
		_, s := utf8.DecodeRune(e.text[e.gapend:])
		e.gapend += s
		e.changed = e.changed || s > 0
	}
	return caret
}

// moveGap moves the gap to the caret position. After returning,
// the gap is guaranteed to be at least space bytes long.
func (e *editBuffer) moveGap(caret, space int) {
	if e.gapLen() < space {
		if space < minSpace {
			space = minSpace
		}
		txt := make([]byte, e.len()+space)
		// Expand to capacity.
		txt = txt[:cap(txt)]
		gaplen := len(txt) - e.len()
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

func (e *editBuffer) len() int {
	return len(e.text) - e.gapLen()
}

func (e *editBuffer) gapLen() int {
	return e.gapend - e.gapstart
}

func (e *editBuffer) Reset() {
	e.Seek(0, io.SeekStart)
}

// Seek implements io.Seeker
func (e *editBuffer) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case io.SeekStart:
		e.pos = int(offset)
	case io.SeekCurrent:
		e.pos += int(offset)
	case io.SeekEnd:
		e.pos = e.len() - int(offset)
	}
	if e.pos < 0 {
		e.pos = 0
	} else if e.pos > e.len() {
		e.pos = e.len()
	}
	return int64(e.pos), nil
}

func (e *editBuffer) Read(p []byte) (int, error) {
	if e.pos == e.len() {
		return 0, io.EOF
	}
	var total int
	if e.pos < e.gapstart {
		n := copy(p, e.text[e.pos:e.gapstart])
		p = p[n:]
		total += n
		e.pos += n
	}
	if e.pos >= e.gapstart {
		n := copy(p, e.text[e.pos+e.gapLen():])
		total += n
		e.pos += n
	}
	if e.pos > e.len() {
		panic("hey!")
	}
	return total, nil
}

func (e *editBuffer) ReadRune() (rune, int, error) {
	if e.pos == e.len() {
		return 0, 0, io.EOF
	}
	r, s := e.runeAt(e.pos)
	e.pos += s
	return r, s, nil
}

// WriteTo implements io.WriterTo.
func (e *editBuffer) WriteTo(w io.Writer) (int64, error) {
	n1, err := w.Write(e.text[:e.gapstart])
	if err != nil {
		return int64(n1), err
	}
	n2, err := w.Write(e.text[e.gapend:])
	return int64(n1 + n2), err
}

func (e *editBuffer) String() string {
	var b strings.Builder
	b.Grow(e.len())
	b.Write(e.text[:e.gapstart])
	b.Write(e.text[e.gapend:])
	return b.String()
}

func (e *editBuffer) prepend(caret int, s string) {
	e.moveGap(caret, len(s))
	copy(e.text[caret:], s)
	e.gapstart += len(s)
	e.changed = e.changed || len(s) > 0
}

func (e *editBuffer) runeBefore(idx int) (rune, int) {
	if idx > e.gapstart {
		idx += e.gapLen()
	}
	return utf8.DecodeLastRune(e.text[:idx])
}

func (e *editBuffer) runeAt(idx int) (rune, int) {
	if idx >= e.gapstart {
		idx += e.gapLen()
	}
	return utf8.DecodeRune(e.text[idx:])
}
