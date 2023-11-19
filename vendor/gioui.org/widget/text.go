package widget

import (
	"bufio"
	"image"
	"io"
	"math"
	"sort"
	"unicode"
	"unicode/utf8"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"golang.org/x/exp/slices"
	"golang.org/x/image/math/fixed"
)

// textSource provides text data for use in widgets. If the underlying data type
// can fail due to I/O errors, it is the responsibility of that type to provide
// its own mechanism to surface and handle those errors. They will not always
// be returned by widgets using these functions.
type textSource interface {
	io.ReaderAt
	// Size returns the total length of the data in bytes.
	Size() int64
	// Changed returns whether the contents have changed since the last call
	// to Changed.
	Changed() bool
	// ReplaceRunes replaces runeCount runes starting at byteOffset within the
	// data with the provided string. Implementations of read-only text sources
	// are free to make this a no-op.
	ReplaceRunes(byteOffset int64, runeCount int64, replacement string)
}

// textView provides efficient shaping and indexing of interactive text. When provided
// with a TextSource, textView will shape and cache the runes within that source.
// It provides methods for configuring a viewport onto the shaped text which can
// be scrolled, and for configuring and drawing text selection boxes.
type textView struct {
	Alignment text.Alignment
	// LineHeight controls the distance between the baselines of lines of text.
	// If zero, a sensible default will be used.
	LineHeight unit.Sp
	// LineHeightScale applies a scaling factor to the LineHeight. If zero, a
	// sensible default will be used.
	LineHeightScale float32
	// SingleLine forces the text to stay on a single line.
	// SingleLine also sets the scrolling direction to
	// horizontal.
	SingleLine bool
	// MaxLines limits the shaped text to a specific quantity of shaped lines.
	MaxLines int
	// Truncator is the text that will be shown at the end of the final
	// line if MaxLines is exceeded. Defaults to "…" if empty.
	Truncator string
	// WrapPolicy configures how displayed text will be broken into lines.
	WrapPolicy text.WrapPolicy
	// Mask replaces the visual display of each rune in the contents with the given rune.
	// Newline characters are not masked. When non-zero, the unmasked contents
	// are accessed by Len, Text, and SetText.
	Mask rune

	params     text.Parameters
	shaper     *text.Shaper
	seekCursor int64
	rr         textSource
	maskReader maskReader
	// graphemes tracks the indices of grapheme cluster boundaries within rr.
	graphemes []int
	// paragraphReader is used to populate graphemes.
	paragraphReader graphemeReader
	lastMask        rune
	viewSize        image.Point
	valid           bool
	regions         []Region
	dims            layout.Dimensions

	// offIndex is an index of rune index to byte offsets.
	offIndex []offEntry

	index glyphIndex

	caret struct {
		// xoff is the offset to the current position when moving between lines.
		xoff fixed.Int26_6
		// start is the current caret position in runes, and also the start position of
		// selected text. end is the end position of selected text. If start
		// == end, then there's no selection. Note that it's possible (and
		// common) that the caret (start) is after the end, e.g. after
		// Shift-DownArrow.
		start int
		end   int
	}

	scrollOff image.Point
}

func (e *textView) Changed() bool {
	return e.rr.Changed()
}

// Dimensions returns the dimensions of the visible text.
func (e *textView) Dimensions() layout.Dimensions {
	basePos := e.dims.Size.Y - e.dims.Baseline
	return layout.Dimensions{Size: e.viewSize, Baseline: e.viewSize.Y - basePos}
}

// FullDimensions returns the dimensions of all shaped text, including
// text that isn't visible within the current viewport.
func (e *textView) FullDimensions() layout.Dimensions {
	return e.dims
}

// SetSource initializes the underlying data source for the Text. This
// must be done before invoking any other methods on Text.
func (e *textView) SetSource(source textSource) {
	e.rr = source
	e.invalidate()
	e.seekCursor = 0
}

// ReadRuneAt reads the rune starting at the given byte offset, if any.
func (e *textView) ReadRuneAt(off int64) (rune, int, error) {
	var buf [utf8.UTFMax]byte
	b := buf[:]
	n, err := e.rr.ReadAt(b, off)
	b = b[:n]
	r, s := utf8.DecodeRune(b)
	return r, s, err
}

// ReadRuneAt reads the run prior to the given byte offset, if any.
func (e *textView) ReadRuneBefore(off int64) (rune, int, error) {
	var buf [utf8.UTFMax]byte
	b := buf[:]
	if off < utf8.UTFMax {
		b = b[:off]
		off = 0
	} else {
		off -= utf8.UTFMax
	}
	n, err := e.rr.ReadAt(b, off)
	b = b[:n]
	r, s := utf8.DecodeLastRune(b)
	return r, s, err
}

func (e *textView) makeValid() {
	if e.valid {
		return
	}
	e.layoutText(e.shaper)
	e.valid = true
}

func (e *textView) closestToRune(runeIdx int) combinedPos {
	e.makeValid()
	pos, _ := e.index.closestToRune(runeIdx)
	return pos
}

func (e *textView) closestToLineCol(line, col int) combinedPos {
	e.makeValid()
	return e.index.closestToLineCol(screenPos{line: line, col: col})
}

func (e *textView) closestToXY(x fixed.Int26_6, y int) combinedPos {
	e.makeValid()
	return e.index.closestToXY(x, y)
}

func (e *textView) closestToXYGraphemes(x fixed.Int26_6, y int) combinedPos {
	// Find the closest existing rune position to the provided coordinates.
	pos := e.closestToXY(x, y)
	// Resolve cluster boundaries on either side of the rune position.
	firstOption := e.moveByGraphemes(pos.runes, 0)
	distance := 1
	if firstOption > pos.runes {
		distance = -1
	}
	secondOption := e.moveByGraphemes(firstOption, distance)
	// Choose the closest grapheme cluster boundary to the desired point.
	first := e.closestToRune(firstOption)
	firstDist := absFixed(first.x - x)
	second := e.closestToRune(secondOption)
	secondDist := absFixed(second.x - x)
	if firstDist > secondDist {
		return second
	} else {
		return first
	}
}

func absFixed(i fixed.Int26_6) fixed.Int26_6 {
	if i < 0 {
		return -i
	}
	return i
}

// MaxLines moves the cursor the specified number of lines vertically, ensuring
// that the resulting position is aligned to a grapheme cluster.
func (e *textView) MoveLines(distance int, selAct selectionAction) {
	caretStart := e.closestToRune(e.caret.start)
	x := caretStart.x + e.caret.xoff
	// Seek to line.
	pos := e.closestToLineCol(caretStart.lineCol.line+distance, 0)
	pos = e.closestToXYGraphemes(x, pos.y)
	e.caret.start = pos.runes
	e.caret.xoff = x - pos.x
	e.updateSelection(selAct)
}

// calculateViewSize determines the size of the current visible content,
// ensuring that even if there is no text content, some space is reserved
// for the caret.
func (e *textView) calculateViewSize(gtx layout.Context) image.Point {
	base := e.dims.Size
	if caretWidth := e.caretWidth(gtx); base.X < caretWidth {
		base.X = caretWidth
	}
	return gtx.Constraints.Constrain(base)
}

// Update the text, reshaping it as necessary. If not nil, eventHandling will be invoked after reshaping the text to
// allow parent widgets to adapt to any changes in text content or positioning. If eventHandling modifies the contents
// of the textView, it is guaranteed to be reshaped (and ready for painting) before Update returns.
func (e *textView) Update(gtx layout.Context, lt *text.Shaper, font font.Font, size unit.Sp, eventHandling func(gtx layout.Context)) {
	if e.params.Locale != gtx.Locale {
		e.params.Locale = gtx.Locale
		e.invalidate()
	}
	textSize := fixed.I(gtx.Sp(size))
	if e.params.Font != font || e.params.PxPerEm != textSize {
		e.invalidate()
		e.params.Font = font
		e.params.PxPerEm = textSize
	}
	maxWidth := gtx.Constraints.Max.X
	if e.SingleLine {
		maxWidth = math.MaxInt
	}
	minWidth := gtx.Constraints.Min.X
	if maxWidth != e.params.MaxWidth {
		e.params.MaxWidth = maxWidth
		e.invalidate()
	}
	if minWidth != e.params.MinWidth {
		e.params.MinWidth = minWidth
		e.invalidate()
	}
	if lt != e.shaper {
		e.shaper = lt
		e.invalidate()
	}
	if e.Mask != e.lastMask {
		e.lastMask = e.Mask
		e.invalidate()
	}
	if e.Alignment != e.params.Alignment {
		e.params.Alignment = e.Alignment
		e.invalidate()
	}
	if e.Truncator != e.params.Truncator {
		e.params.Truncator = e.Truncator
		e.invalidate()
	}
	if e.MaxLines != e.params.MaxLines {
		e.params.MaxLines = e.MaxLines
		e.invalidate()
	}
	if e.WrapPolicy != e.params.WrapPolicy {
		e.params.WrapPolicy = e.WrapPolicy
		e.invalidate()
	}
	if lh := fixed.I(gtx.Sp(e.LineHeight)); lh != e.params.LineHeight {
		e.params.LineHeight = lh
		e.invalidate()
	}
	if e.LineHeightScale != e.params.LineHeightScale {
		e.params.LineHeightScale = e.LineHeightScale
		e.invalidate()
	}

	e.makeValid()
	if eventHandling != nil {
		eventHandling(gtx)
		e.makeValid()
	}

	if viewSize := e.calculateViewSize(gtx); viewSize != e.viewSize {
		e.viewSize = viewSize
		e.invalidate()
	}
	e.makeValid()
}

// PaintSelection clips and paints the visible text selection rectangles using
// the provided material to fill the rectangles.
func (e *textView) PaintSelection(gtx layout.Context, material op.CallOp) {
	localViewport := image.Rectangle{Max: e.viewSize}
	docViewport := image.Rectangle{Max: e.viewSize}.Add(e.scrollOff)
	defer clip.Rect(localViewport).Push(gtx.Ops).Pop()
	e.regions = e.index.locate(docViewport, e.caret.start, e.caret.end, e.regions)
	for _, region := range e.regions {
		area := clip.Rect(region.Bounds).Push(gtx.Ops)
		material.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		area.Pop()
	}
}

// PaintText clips and paints the visible text glyph outlines using the provided
// material to fill the glyphs.
func (e *textView) PaintText(gtx layout.Context, material op.CallOp) {
	m := op.Record(gtx.Ops)
	viewport := image.Rectangle{
		Min: e.scrollOff,
		Max: e.viewSize.Add(e.scrollOff),
	}
	it := textIterator{
		viewport: viewport,
		material: material,
	}

	startGlyph := 0
	for _, line := range e.index.lines {
		if line.descent.Ceil()+line.yOff >= viewport.Min.Y {
			break
		}
		startGlyph += line.glyphs
	}
	var glyphs [32]text.Glyph
	line := glyphs[:0]
	for _, g := range e.index.glyphs[startGlyph:] {
		var ok bool
		if line, ok = it.paintGlyph(gtx, e.shaper, g, line); !ok {
			break
		}
	}

	call := m.Stop()
	viewport.Min = viewport.Min.Add(it.padding.Min)
	viewport.Max = viewport.Max.Add(it.padding.Max)
	defer clip.Rect(viewport.Sub(e.scrollOff)).Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
}

// caretWidth returns the width occupied by the caret for the current
// gtx.
func (e *textView) caretWidth(gtx layout.Context) int {
	carWidth2 := gtx.Dp(1) / 2
	if carWidth2 < 1 {
		carWidth2 = 1
	}
	return carWidth2
}

// PaintCaret clips and paints the caret rectangle, adding material immediately
// before painting to set the appropriate paint material.
func (e *textView) PaintCaret(gtx layout.Context, material op.CallOp) {
	carWidth2 := e.caretWidth(gtx)
	caretPos, carAsc, carDesc := e.CaretInfo()

	carRect := image.Rectangle{
		Min: caretPos.Sub(image.Pt(carWidth2, carAsc)),
		Max: caretPos.Add(image.Pt(carWidth2, carDesc)),
	}
	cl := image.Rectangle{Max: e.viewSize}
	carRect = cl.Intersect(carRect)
	if !carRect.Empty() {
		defer clip.Rect(carRect).Push(gtx.Ops).Pop()
		material.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
}

func (e *textView) CaretInfo() (pos image.Point, ascent, descent int) {
	caretStart := e.closestToRune(e.caret.start)

	ascent = caretStart.ascent.Ceil()
	descent = caretStart.descent.Ceil()

	pos = image.Point{
		X: caretStart.x.Round(),
		Y: caretStart.y,
	}
	pos = pos.Sub(e.scrollOff)
	return
}

// ByteOffset returns the start byte of the rune at the given
// rune offset, clamped to the size of the text.
func (e *textView) ByteOffset(runeOffset int) int64 {
	return int64(e.runeOffset(e.closestToRune(runeOffset).runes))
}

// Len is the length of the editor contents, in runes.
func (e *textView) Len() int {
	e.makeValid()
	return e.closestToRune(math.MaxInt).runes
}

// Text returns the contents of the editor. If the provided buf is large enough, it will
// be filled and returned. Otherwise a new buffer will be allocated.
// Callers can guarantee that buf is large enough by giving it capacity e.Len()*utf8.UTFMax.
func (e *textView) Text(buf []byte) []byte {
	size := e.rr.Size()
	if cap(buf) < int(size) {
		buf = make([]byte, size)
	}
	buf = buf[:size]
	e.Seek(0, io.SeekStart)
	n, _ := io.ReadFull(e, buf)
	buf = buf[:n]
	return buf
}

func (e *textView) ScrollBounds() image.Rectangle {
	var b image.Rectangle
	if e.SingleLine {
		if len(e.index.lines) > 0 {
			line := e.index.lines[0]
			b.Min.X = line.xOff.Floor()
			if b.Min.X > 0 {
				b.Min.X = 0
			}
		}
		b.Max.X = e.dims.Size.X + b.Min.X - e.viewSize.X
	} else {
		b.Max.Y = e.dims.Size.Y - e.viewSize.Y
	}
	return b
}

func (e *textView) ScrollRel(dx, dy int) {
	e.scrollAbs(e.scrollOff.X+dx, e.scrollOff.Y+dy)
}

// ScrollOff returns the scroll offset of the text viewport.
func (e *textView) ScrollOff() image.Point {
	return e.scrollOff
}

func (e *textView) scrollAbs(x, y int) {
	e.scrollOff.X = x
	e.scrollOff.Y = y
	b := e.ScrollBounds()
	if e.scrollOff.X > b.Max.X {
		e.scrollOff.X = b.Max.X
	}
	if e.scrollOff.X < b.Min.X {
		e.scrollOff.X = b.Min.X
	}
	if e.scrollOff.Y > b.Max.Y {
		e.scrollOff.Y = b.Max.Y
	}
	if e.scrollOff.Y < b.Min.Y {
		e.scrollOff.Y = b.Min.Y
	}
}

// MoveCoord moves the caret to the position closest to the provided
// point that is aligned to a grapheme cluster boundary.
func (e *textView) MoveCoord(pos image.Point) {
	x := fixed.I(pos.X + e.scrollOff.X)
	y := pos.Y + e.scrollOff.Y
	e.caret.start = e.closestToXYGraphemes(x, y).runes
	e.caret.xoff = 0
}

// Truncated returns whether the text in the textView is currently
// truncated due to a restriction on the number of lines.
func (e *textView) Truncated() bool {
	return e.index.truncated
}

func (e *textView) layoutText(lt *text.Shaper) {
	e.Seek(0, io.SeekStart)
	var r io.Reader = e
	if e.Mask != 0 {
		e.maskReader.Reset(e, e.Mask)
		r = &e.maskReader
	}
	e.index.reset()
	it := textIterator{viewport: image.Rectangle{Max: image.Point{X: math.MaxInt, Y: math.MaxInt}}}
	if lt != nil {
		lt.Layout(e.params, r)
		for glyph, ok := it.processGlyph(lt.NextGlyph()); ok; glyph, ok = it.processGlyph(lt.NextGlyph()) {
			e.index.Glyph(glyph)
		}
	} else {
		// Make a fake glyph for every rune in the reader.
		b := bufio.NewReader(r)
		for _, _, err := b.ReadRune(); err != io.EOF; _, _, err = b.ReadRune() {
			g, _ := it.processGlyph(text.Glyph{Runes: 1, Flags: text.FlagClusterBreak}, true)
			e.index.Glyph(g)
		}
	}
	e.paragraphReader.SetSource(e.rr)
	e.graphemes = e.graphemes[:0]
	for g := e.paragraphReader.Graphemes(); len(g) > 0; g = e.paragraphReader.Graphemes() {
		if len(e.graphemes) > 0 && g[0] == e.graphemes[len(e.graphemes)-1] {
			g = g[1:]
		}
		e.graphemes = append(e.graphemes, g...)
	}
	dims := layout.Dimensions{Size: it.bounds.Size()}
	dims.Baseline = dims.Size.Y - it.baseline
	e.dims = dims
}

// CaretPos returns the line & column numbers of the caret.
func (e *textView) CaretPos() (line, col int) {
	pos := e.closestToRune(e.caret.start)
	return pos.lineCol.line, pos.lineCol.col
}

// CaretCoords returns the coordinates of the caret, relative to the
// editor itself.
func (e *textView) CaretCoords() f32.Point {
	pos := e.closestToRune(e.caret.start)
	return f32.Pt(float32(pos.x)/64-float32(e.scrollOff.X), float32(pos.y-e.scrollOff.Y))
}

// indexRune returns the latest rune index and byte offset no later than r.
func (e *textView) indexRune(r int) offEntry {
	// Initialize index.
	if len(e.offIndex) == 0 {
		e.offIndex = append(e.offIndex, offEntry{})
	}
	i := sort.Search(len(e.offIndex), func(i int) bool {
		entry := e.offIndex[i]
		return entry.runes >= r
	})
	// Return the entry guaranteed to be less than or equal to r.
	if i > 0 {
		i--
	}
	return e.offIndex[i]
}

// runeOffset returns the byte offset into e.rr of the r'th rune.
// r must be a valid rune index, usually returned by closestPosition.
func (e *textView) runeOffset(r int) int {
	const runesPerIndexEntry = 50
	entry := e.indexRune(r)
	lastEntry := e.offIndex[len(e.offIndex)-1].runes
	for entry.runes < r {
		if entry.runes > lastEntry && entry.runes%runesPerIndexEntry == runesPerIndexEntry-1 {
			e.offIndex = append(e.offIndex, entry)
		}
		_, s, _ := e.ReadRuneAt(int64(entry.bytes))
		entry.bytes += s
		entry.runes++
	}
	return entry.bytes
}

func (e *textView) invalidate() {
	e.offIndex = e.offIndex[:0]
	e.valid = false
}

// Replace the text between start and end with s. Indices are in runes.
// It returns the number of runes inserted.
func (e *textView) Replace(start, end int, s string) int {
	if start > end {
		start, end = end, start
	}
	startPos := e.closestToRune(start)
	endPos := e.closestToRune(end)
	startOff := e.runeOffset(startPos.runes)
	replaceSize := endPos.runes - startPos.runes
	sc := utf8.RuneCountInString(s)
	newEnd := startPos.runes + sc

	e.rr.ReplaceRunes(int64(startOff), int64(replaceSize), s)
	adjust := func(pos int) int {
		switch {
		case newEnd < pos && pos <= endPos.runes:
			pos = newEnd
		case endPos.runes < pos:
			diff := newEnd - endPos.runes
			pos = pos + diff
		}
		return pos
	}
	e.caret.start = adjust(e.caret.start)
	e.caret.end = adjust(e.caret.end)
	e.invalidate()
	return sc
}

// MovePages moves the caret position by vertical pages of text, ensuring that
// the final position is aligned to a grapheme cluster boundary.
func (e *textView) MovePages(pages int, selAct selectionAction) {
	caret := e.closestToRune(e.caret.start)
	x := caret.x + e.caret.xoff
	y := caret.y + pages*e.viewSize.Y
	pos := e.closestToXYGraphemes(x, y)
	e.caret.start = pos.runes
	e.caret.xoff = x - pos.x
	e.updateSelection(selAct)
}

// moveByGraphemes returns the rune index resulting from moving the
// specified number of grapheme clusters from startRuneidx.
func (e *textView) moveByGraphemes(startRuneidx, graphemes int) int {
	if len(e.graphemes) == 0 {
		return startRuneidx
	}
	startGraphemeIdx, _ := slices.BinarySearch(e.graphemes, startRuneidx)
	startGraphemeIdx = max(startGraphemeIdx+graphemes, 0)
	startGraphemeIdx = min(startGraphemeIdx, len(e.graphemes)-1)
	startRuneIdx := e.graphemes[startGraphemeIdx]
	return e.closestToRune(startRuneIdx).runes
}

// clampCursorToGraphemes ensures that the final start/end positions of
// the cursor are on grapheme cluster boundaries.
func (e *textView) clampCursorToGraphemes() {
	e.caret.start = e.moveByGraphemes(e.caret.start, 0)
	e.caret.end = e.moveByGraphemes(e.caret.end, 0)
}

// MoveCaret moves the caret (aka selection start) and the selection end
// relative to their current positions. Positive distances moves forward,
// negative distances moves backward. Distances are in grapheme clusters which
// better match the expectations of users than runes.
func (e *textView) MoveCaret(startDelta, endDelta int) {
	e.caret.xoff = 0
	e.caret.start = e.moveByGraphemes(e.caret.start, startDelta)
	e.caret.end = e.moveByGraphemes(e.caret.end, endDelta)
}

// MoveStart moves the caret to the start of the current line, ensuring that the resulting
// cursor position is on a grapheme cluster boundary.
func (e *textView) MoveStart(selAct selectionAction) {
	caret := e.closestToRune(e.caret.start)
	caret = e.closestToLineCol(caret.lineCol.line, 0)
	e.caret.start = caret.runes
	e.caret.xoff = -caret.x
	e.updateSelection(selAct)
	e.clampCursorToGraphemes()
}

// MoveEnd moves the caret to the end of the current line, ensuring that the resulting
// cursor position is on a grapheme cluster boundary.
func (e *textView) MoveEnd(selAct selectionAction) {
	caret := e.closestToRune(e.caret.start)
	caret = e.closestToLineCol(caret.lineCol.line, math.MaxInt)
	e.caret.start = caret.runes
	e.caret.xoff = fixed.I(e.params.MaxWidth) - caret.x
	e.updateSelection(selAct)
	e.clampCursorToGraphemes()
}

// MoveWord moves the caret to the next word in the specified direction.
// Positive is forward, negative is backward.
// Absolute values greater than one will skip that many words.
// The final caret position will be aligned to a grapheme cluster boundary.
// BUG(whereswaldon): this method's definition of a "word" is currently
// whitespace-delimited. Languages that do not use whitespace to delimit
// words will experience counter-intuitive behavior when navigating by
// word.
func (e *textView) MoveWord(distance int, selAct selectionAction) {
	// split the distance information into constituent parts to be
	// used independently.
	words, direction := distance, 1
	if distance < 0 {
		words, direction = distance*-1, -1
	}
	// atEnd if caret is at either side of the buffer.
	caret := e.closestToRune(e.caret.start)
	atEnd := func() bool {
		return caret.runes == 0 || caret.runes == e.Len()
	}
	// next returns the appropriate rune given the direction.
	next := func() (r rune) {
		off := e.runeOffset(caret.runes)
		if direction < 0 {
			r, _, _ = e.ReadRuneBefore(int64(off))
		} else {
			r, _, _ = e.ReadRuneAt(int64(off))
		}
		return r
	}
	for ii := 0; ii < words; ii++ {
		for r := next(); unicode.IsSpace(r) && !atEnd(); r = next() {
			e.MoveCaret(direction, 0)
			caret = e.closestToRune(e.caret.start)
		}
		e.MoveCaret(direction, 0)
		caret = e.closestToRune(e.caret.start)
		for r := next(); !unicode.IsSpace(r) && !atEnd(); r = next() {
			e.MoveCaret(direction, 0)
			caret = e.closestToRune(e.caret.start)
		}
	}
	e.updateSelection(selAct)
	e.clampCursorToGraphemes()
}

func (e *textView) ScrollToCaret() {
	caret := e.closestToRune(e.caret.start)
	if e.SingleLine {
		var dist int
		if d := caret.x.Floor() - e.scrollOff.X; d < 0 {
			dist = d
		} else if d := caret.x.Ceil() - (e.scrollOff.X + e.viewSize.X); d > 0 {
			dist = d
		}
		e.ScrollRel(dist, 0)
	} else {
		miny := caret.y - caret.ascent.Ceil()
		maxy := caret.y + caret.descent.Ceil()
		var dist int
		if d := miny - e.scrollOff.Y; d < 0 {
			dist = d
		} else if d := maxy - (e.scrollOff.Y + e.viewSize.Y); d > 0 {
			dist = d
		}
		e.ScrollRel(0, dist)
	}
}

// SelectionLen returns the length of the selection, in runes; it is
// equivalent to utf8.RuneCountInString(e.SelectedText()).
func (e *textView) SelectionLen() int {
	return abs(e.caret.start - e.caret.end)
}

// Selection returns the start and end of the selection, as rune offsets.
// start can be > end.
func (e *textView) Selection() (start, end int) {
	return e.caret.start, e.caret.end
}

// SetCaret moves the caret to start, and sets the selection end to end. Then
// the two ends are clamped to the nearest grapheme cluster boundary. start
// and end are in runes, and represent offsets into the editor text.
func (e *textView) SetCaret(start, end int) {
	e.caret.start = e.closestToRune(start).runes
	e.caret.end = e.closestToRune(end).runes
	e.clampCursorToGraphemes()
}

// SelectedText returns the currently selected text (if any) from the editor,
// filling the provided byte slice if it is large enough or allocating and
// returning a new byte slice if the provided one is insufficient.
// Callers can guarantee that the buf is large enough by providing a buffer
// with capacity e.SelectionLen()*utf8.UTFMax.
func (e *textView) SelectedText(buf []byte) []byte {
	startOff := e.runeOffset(e.caret.start)
	endOff := e.runeOffset(e.caret.end)
	start := min(startOff, endOff)
	end := max(startOff, endOff)
	if cap(buf) < end-start {
		buf = make([]byte, end-start)
	}
	buf = buf[:end-start]
	n, _ := e.rr.ReadAt(buf, int64(start))
	// There is no way to reasonably handle a read error here. We rely upon
	// implementations of textSource to provide other ways to signal errors
	// if the user cares about that, and here we use whatever data we were
	// able to read.
	return buf[:n]
}

func (e *textView) updateSelection(selAct selectionAction) {
	if selAct == selectionClear {
		e.ClearSelection()
	}
}

// ClearSelection clears the selection, by setting the selection end equal to
// the selection start.
func (e *textView) ClearSelection() {
	e.caret.end = e.caret.start
}

// WriteTo implements io.WriterTo.
func (e *textView) WriteTo(w io.Writer) (int64, error) {
	e.Seek(0, io.SeekStart)
	return io.Copy(w, struct{ io.Reader }{e})
}

// Seek implements io.Seeker.
func (e *textView) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		e.seekCursor = offset
	case io.SeekCurrent:
		e.seekCursor += offset
	case io.SeekEnd:
		e.seekCursor = e.rr.Size() + offset
	}
	return e.seekCursor, nil
}

// Read implements io.Reader.
func (e *textView) Read(p []byte) (int, error) {
	n, err := e.rr.ReadAt(p, e.seekCursor)
	e.seekCursor += int64(n)
	return n, err
}

// ReadAt implements io.ReaderAt.
func (e *textView) ReadAt(p []byte, offset int64) (int, error) {
	return e.rr.ReadAt(p, offset)
}

// Regions returns visible regions covering the rune range [start,end).
func (e *textView) Regions(start, end int, regions []Region) []Region {
	viewport := image.Rectangle{
		Min: e.scrollOff,
		Max: e.viewSize.Add(e.scrollOff),
	}
	return e.index.locate(viewport, start, end, regions)
}
