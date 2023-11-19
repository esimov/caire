// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"bytes"
	"image"
	"io"
	"log"
	"os"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/fontscan"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/opentype/api"
	"github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/exp/slices"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/unicode/bidi"

	"gioui.org/f32"
	giofont "gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/internal/debug"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// document holds a collection of shaped lines and alignment information for
// those lines.
type document struct {
	lines     []line
	alignment Alignment
	// alignWidth is the width used when aligning text.
	alignWidth      int
	unreadRuneCount int
}

// append adds the lines of other to the end of l and ensures they
// are aligned to the same width.
func (l *document) append(other document) {
	l.lines = append(l.lines, other.lines...)
	l.alignWidth = max(l.alignWidth, other.alignWidth)
	calculateYOffsets(l.lines)
}

// reset empties the document in preparation to reuse its memory.
func (l *document) reset() {
	l.lines = l.lines[:0]
	l.alignment = Start
	l.alignWidth = 0
	l.unreadRuneCount = 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// A line contains the measurements of a line of text.
type line struct {
	// runs contains sequences of shaped glyphs with common attributes. The order
	// of runs is logical, meaning that the first run will contain the glyphs
	// corresponding to the first runes of data in the original text.
	runs []runLayout
	// visualOrder is a slice of indices into Runs that describes the visual positions
	// of each run of text. Iterating this slice and accessing Runs at each
	// of the values stored in this slice traverses the runs in proper visual
	// order from left to right.
	visualOrder []int
	// width is the width of the line.
	width fixed.Int26_6
	// ascent is the height above the baseline.
	ascent fixed.Int26_6
	// descent is the height below the baseline, including
	// the line gap.
	descent fixed.Int26_6
	// lineHeight captures the gap that should exist between the baseline of this
	// line and the previous (if any).
	lineHeight fixed.Int26_6
	// direction is the dominant direction of the line. This direction will be
	// used to align the text content of the line, but may not match the actual
	// direction of the runs of text within the line (such as an RTL sentence
	// within an LTR paragraph).
	direction system.TextDirection
	// runeCount is the number of text runes represented by this line's runs.
	runeCount int

	yOffset int
}

// insertTrailingSyntheticNewline adds a synthetic newline to the final logical run of the line
// with the given shaping cluster index.
func (l *line) insertTrailingSyntheticNewline(newLineClusterIdx int) {
	// If there was a newline at the end of this paragraph, insert a synthetic glyph representing it.
	finalContentRun := len(l.runs) - 1
	// If there was a trailing newline update the rune counts to include
	// it on the last line of the paragraph.
	l.runeCount += 1
	l.runs[finalContentRun].Runes.Count += 1

	syntheticGlyph := glyph{
		id:           0,
		clusterIndex: newLineClusterIdx,
		glyphCount:   0,
		runeCount:    1,
		xAdvance:     0,
		yAdvance:     0,
		xOffset:      0,
		yOffset:      0,
	}
	// Inset the synthetic newline glyph on the proper end of the run.
	if l.runs[finalContentRun].Direction.Progression() == system.FromOrigin {
		l.runs[finalContentRun].Glyphs = append(l.runs[finalContentRun].Glyphs, syntheticGlyph)
	} else {
		// Ensure capacity.
		l.runs[finalContentRun].Glyphs = append(l.runs[finalContentRun].Glyphs, glyph{})
		copy(l.runs[finalContentRun].Glyphs[1:], l.runs[finalContentRun].Glyphs)
		l.runs[finalContentRun].Glyphs[0] = syntheticGlyph
	}
}

func (l *line) setTruncatedCount(truncatedCount int) {
	// If we've truncated the text with a truncator, adjust the rune counts within the
	// truncator to make it represent the truncated text.
	finalRunIdx := len(l.runs) - 1
	l.runs[finalRunIdx].truncator = true
	finalGlyphIdx := len(l.runs[finalRunIdx].Glyphs) - 1
	// The run represents all of the truncated text.
	l.runs[finalRunIdx].Runes.Count = truncatedCount
	// Only the final glyph represents any runes, and it represents all truncated text.
	for i := range l.runs[finalRunIdx].Glyphs {
		if i == finalGlyphIdx {
			l.runs[finalRunIdx].Glyphs[finalGlyphIdx].runeCount = truncatedCount
		} else {
			l.runs[finalRunIdx].Glyphs[finalGlyphIdx].runeCount = 0
		}
	}
}

// Range describes the position and quantity of a range of text elements
// within a larger slice. The unit is usually runes of unicode data or
// glyphs of shaped font data.
type Range struct {
	// Count describes the number of items represented by the Range.
	Count int
	// Offset describes the start position of the represented
	// items within a larger list.
	Offset int
}

// glyph contains the metadata needed to render a glyph.
type glyph struct {
	// id is this glyph's identifier within the font it was shaped with.
	id GlyphID
	// clusterIndex is the identifier for the text shaping cluster that
	// this glyph is part of.
	clusterIndex int
	// glyphCount is the number of glyphs in the same cluster as this glyph.
	glyphCount int
	// runeCount is the quantity of runes in the source text that this glyph
	// corresponds to.
	runeCount int
	// xAdvance and yAdvance describe the distance the dot moves when
	// laying out the glyph on the X or Y axis.
	xAdvance, yAdvance fixed.Int26_6
	// xOffset and yOffset describe offsets from the dot that should be
	// applied when rendering the glyph.
	xOffset, yOffset fixed.Int26_6
	// bounds describes the visual bounding box of the glyph relative to
	// its dot.
	bounds fixed.Rectangle26_6
}

type runLayout struct {
	// VisualPosition describes the relative position of this run of text within
	// its line. It should be a valid index into the containing line's VisualOrder
	// slice.
	VisualPosition int
	// X is the visual offset of the dot for the first glyph in this run
	// relative to the beginning of the line.
	X fixed.Int26_6
	// Glyphs are the actual font characters for the text. They are ordered
	// from left to right regardless of the text direction of the underlying
	// text.
	Glyphs []glyph
	// Runes describes the position of the text data this layout represents
	// within the containing text.Line.
	Runes Range
	// Advance is the sum of the advances of all clusters in the Layout.
	Advance fixed.Int26_6
	// PPEM is the pixels-per-em scale used to shape this run.
	PPEM fixed.Int26_6
	// Direction is the layout direction of the glyphs.
	Direction system.TextDirection
	// face is the font face that the ID of each Glyph in the Layout refers to.
	face font.Face
	// truncator indicates that this run is a text truncator standing in for remaining
	// text.
	truncator bool
}

// shaperImpl implements the shaping and line-wrapping of opentype fonts.
type shaperImpl struct {
	// Fields for tracking fonts/faces.
	fontMap      *fontscan.FontMap
	faces        []font.Face
	faceToIndex  map[font.Font]int
	faceMeta     []giofont.Font
	defaultFaces []string
	logger       interface {
		Printf(format string, args ...any)
	}
	parser parser

	// Shaping and wrapping state.
	shaper        shaping.HarfbuzzShaper
	wrapper       shaping.LineWrapper
	bidiParagraph bidi.Paragraph

	// Scratch buffers used to avoid re-allocating slices during routine internal
	// shaping operations.
	splitScratch1, splitScratch2 []shaping.Input
	outScratchBuf                []shaping.Output
	scratchRunes                 []rune

	// bitmapGlyphCache caches extracted bitmap glyph images.
	bitmapGlyphCache bitmapCache
}

// debugLogger only logs messages if debug.Text is true.
type debugLogger struct {
	*log.Logger
}

func newDebugLogger() debugLogger {
	return debugLogger{Logger: log.New(log.Writer(), "[text] ", log.Default().Flags())}
}

func (d debugLogger) Printf(format string, args ...any) {
	if debug.Text.Load() {
		d.Logger.Printf(format, args...)
	}
}

func newShaperImpl(systemFonts bool, collection []FontFace) *shaperImpl {
	var shaper shaperImpl
	shaper.logger = newDebugLogger()
	shaper.fontMap = fontscan.NewFontMap(shaper.logger)
	shaper.faceToIndex = make(map[font.Font]int)
	if systemFonts {
		str, err := os.UserCacheDir()
		if err != nil {
			shaper.logger.Printf("failed resolving font cache dir: %v", err)
			shaper.logger.Printf("skipping system font load")
		}
		if err := shaper.fontMap.UseSystemFonts(str); err != nil {
			shaper.logger.Printf("failed loading system fonts: %v", err)
		}
	}
	for _, f := range collection {
		shaper.Load(f)
		shaper.defaultFaces = append(shaper.defaultFaces, string(f.Font.Typeface))
	}
	shaper.shaper.SetFontCacheSize(32)
	return &shaper
}

// Load registers the provided FontFace with the shaper, if it is compatible.
// It returns whether the face is now available for use. FontFaces are prioritized
// in the order in which they are loaded, with the first face being the default.
func (s *shaperImpl) Load(f FontFace) {
	s.fontMap.AddFace(f.Face.Face(), opentype.FontToDescription(f.Font))
	s.addFace(f.Face.Face(), f.Font)
}

func (s *shaperImpl) addFace(f font.Face, md giofont.Font) {
	if _, ok := s.faceToIndex[f.Font]; ok {
		return
	}
	idx := len(s.faces)
	s.faceToIndex[f.Font] = idx
	s.faces = append(s.faces, f)
	s.faceMeta = append(s.faceMeta, md)
}

// splitByScript divides the inputs into new, smaller inputs on script boundaries
// and correctly sets the text direction per-script. It will
// use buf as the backing memory for the returned slice if buf is non-nil.
func splitByScript(inputs []shaping.Input, documentDir di.Direction, buf []shaping.Input) []shaping.Input {
	var splitInputs []shaping.Input
	if buf == nil {
		splitInputs = make([]shaping.Input, 0, len(inputs))
	} else {
		splitInputs = buf
	}
	for _, input := range inputs {
		currentInput := input
		if input.RunStart == input.RunEnd {
			return []shaping.Input{input}
		}
		firstNonCommonRune := input.RunStart
		for i := firstNonCommonRune; i < input.RunEnd; i++ {
			if language.LookupScript(input.Text[i]) != language.Common {
				firstNonCommonRune = i
				break
			}
		}
		currentInput.Script = language.LookupScript(input.Text[firstNonCommonRune])
		for i := firstNonCommonRune + 1; i < input.RunEnd; i++ {
			r := input.Text[i]
			runeScript := language.LookupScript(r)

			if runeScript == language.Common || runeScript == currentInput.Script {
				continue
			}

			if i != input.RunStart {
				currentInput.RunEnd = i
				splitInputs = append(splitInputs, currentInput)
			}

			currentInput = input
			currentInput.RunStart = i
			currentInput.Script = runeScript
			// In the future, it may make sense to try to guess the language of the text here as well,
			// but this is a complex process.
		}
		// close and add the last input
		currentInput.RunEnd = input.RunEnd
		splitInputs = append(splitInputs, currentInput)
	}

	return splitInputs
}

func (s *shaperImpl) splitBidi(input shaping.Input) []shaping.Input {
	var splitInputs []shaping.Input
	if input.Direction.Axis() != di.Horizontal || input.RunStart == input.RunEnd {
		return []shaping.Input{input}
	}
	def := bidi.LeftToRight
	if input.Direction.Progression() == di.TowardTopLeft {
		def = bidi.RightToLeft
	}
	s.bidiParagraph.SetString(string(input.Text), bidi.DefaultDirection(def))
	out, err := s.bidiParagraph.Order()
	if err != nil {
		return []shaping.Input{input}
	}
	for i := 0; i < out.NumRuns(); i++ {
		currentInput := input
		run := out.Run(i)
		dir := run.Direction()
		_, endRune := run.Pos()
		currentInput.RunEnd = endRune + 1
		if dir == bidi.RightToLeft {
			currentInput.Direction = di.DirectionRTL
		} else {
			currentInput.Direction = di.DirectionLTR
		}
		splitInputs = append(splitInputs, currentInput)
		input.RunStart = currentInput.RunEnd
	}
	return splitInputs
}

// ResolveFace allows shaperImpl to implement shaping.FontMap, wrapping its fontMap
// field and ensuring that any faces loaded as part of the search are registered with
// ids so that they can be referred to by a GlyphID.
func (s *shaperImpl) ResolveFace(r rune) font.Face {
	face := s.fontMap.ResolveFace(r)
	if face != nil {
		family, aspect := s.fontMap.FontMetadata(face.Font)
		md := opentype.DescriptionToFont(metadata.Description{
			Family: family,
			Aspect: aspect,
		})
		s.addFace(face, md)
		return face
	}
	return nil
}

// splitByFaces divides the inputs by font coverage in the provided faces. It will use the slice provided in buf
// as the backing storage of the returned slice if buf is non-nil.
func (s *shaperImpl) splitByFaces(inputs []shaping.Input, buf []shaping.Input) []shaping.Input {
	var split []shaping.Input
	if buf == nil {
		split = make([]shaping.Input, 0, len(inputs))
	} else {
		split = buf
	}
	for _, input := range inputs {
		split = append(split, shaping.SplitByFace(input, s)...)
	}
	return split
}

// shapeText invokes the text shaper and returns the raw text data in the shaper's native
// format. It does not wrap lines.
func (s *shaperImpl) shapeText(ppem fixed.Int26_6, lc system.Locale, txt []rune) []shaping.Output {
	lcfg := langConfig{
		Language:  language.NewLanguage(lc.Language),
		Direction: mapDirection(lc.Direction),
	}
	// Create an initial input.
	input := toInput(nil, ppem, lcfg, txt)
	if input.RunStart == input.RunEnd && len(s.faces) > 0 {
		// Give the empty string a face. This is a necessary special case because
		// the face splitting process works by resolving faces for each rune, and
		// the empty string contains no runes.
		input.Face = s.faces[0]
	}
	// Break input on font glyph coverage.
	inputs := s.splitBidi(input)
	inputs = s.splitByFaces(inputs, s.splitScratch1[:0])
	inputs = splitByScript(inputs, lcfg.Direction, s.splitScratch2[:0])
	// Shape all inputs.
	if needed := len(inputs) - len(s.outScratchBuf); needed > 0 {
		s.outScratchBuf = slices.Grow(s.outScratchBuf, needed)
	}
	s.outScratchBuf = s.outScratchBuf[:0]
	for _, input := range inputs {
		if input.Face != nil {
			s.outScratchBuf = append(s.outScratchBuf, s.shaper.Shape(input))
		} else {
			s.outScratchBuf = append(s.outScratchBuf, shaping.Output{
				// Use the text size as the advance of the entire fake run so that
				// it doesn't occupy zero space.
				Advance: input.Size,
				Size:    input.Size,
				Glyphs: []shaping.Glyph{
					{
						Width:        input.Size,
						Height:       input.Size,
						XBearing:     0,
						YBearing:     0,
						XAdvance:     input.Size,
						YAdvance:     input.Size,
						XOffset:      0,
						YOffset:      0,
						ClusterIndex: input.RunStart,
						RuneCount:    input.RunEnd - input.RunStart,
						GlyphCount:   1,
						GlyphID:      0,
						Mask:         0,
					},
				},
				LineBounds: shaping.Bounds{
					Ascent:  input.Size,
					Descent: 0,
					Gap:     0,
				},
				GlyphBounds: shaping.Bounds{
					Ascent:  input.Size,
					Descent: 0,
					Gap:     0,
				},
				Direction: input.Direction,
				Runes: shaping.Range{
					Offset: input.RunStart,
					Count:  input.RunEnd - input.RunStart,
				},
			})
		}
	}
	return s.outScratchBuf
}

func wrapPolicyToGoText(p WrapPolicy) shaping.LineBreakPolicy {
	switch p {
	case WrapGraphemes:
		return shaping.Always
	case WrapWords:
		return shaping.Never
	default:
		return shaping.WhenNecessary
	}
}

// shapeAndWrapText invokes the text shaper and returns wrapped lines in the shaper's native format.
func (s *shaperImpl) shapeAndWrapText(params Parameters, txt []rune) (_ []shaping.Line, truncated int) {
	wc := shaping.WrapConfig{
		TruncateAfterLines: params.MaxLines,
		TextContinues:      params.forceTruncate,
		BreakPolicy:        wrapPolicyToGoText(params.WrapPolicy),
	}
	families := s.defaultFaces
	if params.Font.Typeface != "" {
		parsed, err := s.parser.parse(string(params.Font.Typeface))
		if err != nil {
			s.logger.Printf("Unable to parse typeface %q: %v", params.Font.Typeface, err)
		} else {
			families = parsed
		}
	}
	s.fontMap.SetQuery(fontscan.Query{
		Families: families,
		Aspect:   opentype.FontToDescription(params.Font).Aspect,
	})
	if wc.TruncateAfterLines > 0 {
		if len(params.Truncator) == 0 {
			params.Truncator = "…"
		}
		// We only permit a single run as the truncator, regardless of whether more were generated.
		// Just use the first one.
		wc.Truncator = s.shapeText(params.PxPerEm, params.Locale, []rune(params.Truncator))[0]
	}
	// Wrap outputs into lines.
	return s.wrapper.WrapParagraph(wc, params.MaxWidth, txt, shaping.NewSliceIterator(s.shapeText(params.PxPerEm, params.Locale, txt)))
}

// replaceControlCharacters replaces problematic unicode
// code points with spaces to ensure proper rune accounting.
func replaceControlCharacters(in []rune) []rune {
	for i, r := range in {
		switch r {
		// ASCII File separator.
		case '\u001C':
		// ASCII Group separator.
		case '\u001D':
		// ASCII Record separator.
		case '\u001E':
		case '\r':
		case '\n':
		// Unicode "next line" character.
		case '\u0085':
		// Unicode "paragraph separator".
		case '\u2029':
		default:
			continue
		}
		in[i] = ' '
	}
	return in
}

// Layout shapes and wraps the text, and returns the result in Gio's shaped text format.
func (s *shaperImpl) LayoutString(params Parameters, txt string) document {
	return s.LayoutRunes(params, []rune(txt))
}

// Layout shapes and wraps the text, and returns the result in Gio's shaped text format.
func (s *shaperImpl) Layout(params Parameters, txt io.RuneReader) document {
	s.scratchRunes = s.scratchRunes[:0]
	for r, _, err := txt.ReadRune(); err != nil; r, _, err = txt.ReadRune() {
		s.scratchRunes = append(s.scratchRunes, r)
	}
	return s.LayoutRunes(params, s.scratchRunes)
}

func calculateYOffsets(lines []line) {
	if len(lines) < 1 {
		return
	}
	// Ceil the first value to ensure that we don't baseline it too close to the top of the
	// viewport and cut off the top pixel.
	currentY := lines[0].ascent.Ceil()
	for i := range lines {
		if i > 0 {
			currentY += lines[i].lineHeight.Round()
		}
		lines[i].yOffset = currentY
	}
}

// LayoutRunes shapes and wraps the text, and returns the result in Gio's shaped text format.
func (s *shaperImpl) LayoutRunes(params Parameters, txt []rune) document {
	hasNewline := len(txt) > 0 && txt[len(txt)-1] == '\n'
	var ls []shaping.Line
	var truncated int
	if hasNewline {
		txt = txt[:len(txt)-1]
	}
	if params.MaxLines != 0 && hasNewline {
		// If we might end up truncating a trailing newline, we must insert the truncator symbol
		// on the final line (if we hit the limit).
		params.forceTruncate = true
	}
	ls, truncated = s.shapeAndWrapText(params, replaceControlCharacters(txt))

	hasTruncator := truncated > 0 || (params.forceTruncate && params.MaxLines == len(ls))
	if hasTruncator && hasNewline {
		// We have a truncator at the end of the line, so the newline is logically
		// truncated as well.
		truncated++
		hasNewline = false
	}

	// Convert to Lines.
	textLines := make([]line, len(ls))
	maxHeight := fixed.Int26_6(0)
	for i := range ls {
		otLine := toLine(s.faceToIndex, ls[i], params.Locale.Direction)
		if otLine.lineHeight > maxHeight {
			maxHeight = otLine.lineHeight
		}
		if isFinalLine := i == len(ls)-1; isFinalLine {
			if hasNewline {
				otLine.insertTrailingSyntheticNewline(len(txt))
			}
			if hasTruncator {
				otLine.setTruncatedCount(truncated)
			}
		}
		textLines[i] = otLine
	}
	if params.LineHeight != 0 {
		maxHeight = params.LineHeight
	}
	if params.LineHeightScale == 0 {
		params.LineHeightScale = 1.2
	}

	maxHeight = floatToFixed(fixedToFloat(maxHeight) * params.LineHeightScale)
	for i := range textLines {
		textLines[i].lineHeight = maxHeight
	}
	calculateYOffsets(textLines)
	return document{
		lines:      textLines,
		alignment:  params.Alignment,
		alignWidth: alignWidth(params.MinWidth, textLines),
	}
}

func alignWidth(minWidth int, lines []line) int {
	for _, l := range lines {
		minWidth = max(minWidth, l.width.Ceil())
	}
	return minWidth
}

// Shape converts the provided glyphs into a path. The path will enclose the forms
// of all vector glyphs.
func (s *shaperImpl) Shape(pathOps *op.Ops, gs []Glyph) clip.PathSpec {
	var lastPos f32.Point
	var x fixed.Int26_6
	var builder clip.Path
	builder.Begin(pathOps)
	for i, g := range gs {
		if i == 0 {
			x = g.X
		}
		ppem, faceIdx, gid := splitGlyphID(g.ID)
		if faceIdx >= len(s.faces) {
			continue
		}
		face := s.faces[faceIdx]
		if face == nil {
			continue
		}
		scaleFactor := fixedToFloat(ppem) / float32(face.Upem())
		glyphData := face.GlyphData(gid)
		switch glyphData := glyphData.(type) {
		case api.GlyphOutline:
			outline := glyphData
			// Move to glyph position.
			pos := f32.Point{
				X: fixedToFloat((g.X - x) - g.Offset.X),
				Y: -fixedToFloat(g.Offset.Y),
			}
			builder.Move(pos.Sub(lastPos))
			lastPos = pos
			var lastArg f32.Point

			// Convert fonts.Segments to relative segments.
			for _, fseg := range outline.Segments {
				nargs := 1
				switch fseg.Op {
				case api.SegmentOpQuadTo:
					nargs = 2
				case api.SegmentOpCubeTo:
					nargs = 3
				}
				var args [3]f32.Point
				for i := 0; i < nargs; i++ {
					a := f32.Point{
						X: fseg.Args[i].X * scaleFactor,
						Y: -fseg.Args[i].Y * scaleFactor,
					}
					args[i] = a.Sub(lastArg)
					if i == nargs-1 {
						lastArg = a
					}
				}
				switch fseg.Op {
				case api.SegmentOpMoveTo:
					builder.Move(args[0])
				case api.SegmentOpLineTo:
					builder.Line(args[0])
				case api.SegmentOpQuadTo:
					builder.Quad(args[0], args[1])
				case api.SegmentOpCubeTo:
					builder.Cube(args[0], args[1], args[2])
				default:
					panic("unsupported segment op")
				}
			}
			lastPos = lastPos.Add(lastArg)
		}
	}
	return builder.End()
}

func fixedToFloat(i fixed.Int26_6) float32 {
	return float32(i) / 64.0
}

func floatToFixed(f float32) fixed.Int26_6 {
	return fixed.Int26_6(f * 64)
}

// Bitmaps returns an op.CallOp that will display all bitmap glyphs within gs.
// The positioning of the bitmaps uses the same logic as Shape(), so the returned
// CallOp can be added at the same offset as the path data returned by Shape()
// and will align correctly.
func (s *shaperImpl) Bitmaps(ops *op.Ops, gs []Glyph) op.CallOp {
	var x fixed.Int26_6
	bitmapMacro := op.Record(ops)
	for i, g := range gs {
		if i == 0 {
			x = g.X
		}
		_, faceIdx, gid := splitGlyphID(g.ID)
		if faceIdx >= len(s.faces) {
			continue
		}
		face := s.faces[faceIdx]
		if face == nil {
			continue
		}
		glyphData := face.GlyphData(gid)
		switch glyphData := glyphData.(type) {
		case api.GlyphBitmap:
			var imgOp paint.ImageOp
			var imgSize image.Point
			bitmapData, ok := s.bitmapGlyphCache.Get(g.ID)
			if !ok {
				var img image.Image
				switch glyphData.Format {
				case api.PNG, api.JPG, api.TIFF:
					img, _, _ = image.Decode(bytes.NewReader(glyphData.Data))
				case api.BlackAndWhite:
					// This is a complex family of uncompressed bitmaps that don't seem to be
					// very common in practice. We can try adding support later if needed.
					fallthrough
				default:
					// Unknown format.
					continue
				}
				imgOp = paint.NewImageOp(img)
				imgSize = img.Bounds().Size()
				s.bitmapGlyphCache.Put(g.ID, bitmap{img: imgOp, size: imgSize})
			} else {
				imgOp = bitmapData.img
				imgSize = bitmapData.size
			}
			off := op.Affine(f32.Affine2D{}.Offset(f32.Point{
				X: fixedToFloat((g.X - x) - g.Offset.X),
				Y: fixedToFloat(g.Offset.Y + g.Bounds.Min.Y),
			})).Push(ops)
			cl := clip.Rect{Max: imgSize}.Push(ops)

			glyphSize := image.Rectangle{
				Min: image.Point{
					X: g.Bounds.Min.X.Round(),
					Y: g.Bounds.Min.Y.Round(),
				},
				Max: image.Point{
					X: g.Bounds.Max.X.Round(),
					Y: g.Bounds.Max.Y.Round(),
				},
			}.Size()
			aff := op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Point{
				X: float32(glyphSize.X) / float32(imgSize.X),
				Y: float32(glyphSize.Y) / float32(imgSize.Y),
			})).Push(ops)
			imgOp.Add(ops)
			paint.PaintOp{}.Add(ops)
			aff.Pop()
			cl.Pop()
			off.Pop()
		}
	}
	return bitmapMacro.Stop()
}

// langConfig describes the language and writing system of a body of text.
type langConfig struct {
	// Language the text is written in.
	language.Language
	// Writing system used to represent the text.
	language.Script
	// Direction of the text, usually driven by the writing system.
	di.Direction
}

// toInput converts its parameters into a shaping.Input.
func toInput(face font.Face, ppem fixed.Int26_6, lc langConfig, runes []rune) shaping.Input {
	var input shaping.Input
	input.Direction = lc.Direction
	input.Text = runes
	input.Size = ppem
	input.Face = face
	input.Language = lc.Language
	input.Script = lc.Script
	input.RunStart = 0
	input.RunEnd = len(runes)
	return input
}

func mapDirection(d system.TextDirection) di.Direction {
	switch d {
	case system.LTR:
		return di.DirectionLTR
	case system.RTL:
		return di.DirectionRTL
	}
	return di.DirectionLTR
}

func unmapDirection(d di.Direction) system.TextDirection {
	switch d {
	case di.DirectionLTR:
		return system.LTR
	case di.DirectionRTL:
		return system.RTL
	}
	return system.LTR
}

// toGioGlyphs converts text shaper glyphs into the minimal representation
// that Gio needs.
func toGioGlyphs(in []shaping.Glyph, ppem fixed.Int26_6, faceIdx int) []glyph {
	out := make([]glyph, 0, len(in))
	for _, g := range in {
		// To better understand how to calculate the bounding box, see here:
		// https://freetype.org/freetype2/docs/glyphs/glyph-metrics-3.svg
		var bounds fixed.Rectangle26_6
		bounds.Min.X = g.XBearing
		bounds.Min.Y = -g.YBearing
		bounds.Max = bounds.Min.Add(fixed.Point26_6{X: g.Width, Y: -g.Height})
		out = append(out, glyph{
			id:           newGlyphID(ppem, faceIdx, g.GlyphID),
			clusterIndex: g.ClusterIndex,
			runeCount:    g.RuneCount,
			glyphCount:   g.GlyphCount,
			xAdvance:     g.XAdvance,
			yAdvance:     g.YAdvance,
			xOffset:      g.XOffset,
			yOffset:      g.YOffset,
			bounds:       bounds,
		})
	}
	return out
}

// toLine converts the output into a Line with the provided dominant text direction.
func toLine(faceToIndex map[font.Font]int, o shaping.Line, dir system.TextDirection) line {
	if len(o) < 1 {
		return line{}
	}
	line := line{
		runs:      make([]runLayout, len(o)),
		direction: dir,
	}
	maxSize := fixed.Int26_6(0)
	for i := range o {
		run := o[i]
		if run.Size > maxSize {
			maxSize = run.Size
		}
		var font font.Font
		if run.Face != nil {
			font = run.Face.Font
		}
		line.runs[i] = runLayout{
			Glyphs: toGioGlyphs(run.Glyphs, run.Size, faceToIndex[font]),
			Runes: Range{
				Count:  run.Runes.Count,
				Offset: line.runeCount,
			},
			Direction: unmapDirection(run.Direction),
			face:      run.Face,
			Advance:   run.Advance,
			PPEM:      run.Size,
		}
		line.runeCount += run.Runes.Count
		line.width += run.Advance
		if line.ascent < run.LineBounds.Ascent {
			line.ascent = run.LineBounds.Ascent
		}
		if line.descent < -run.LineBounds.Descent+run.LineBounds.Gap {
			line.descent = -run.LineBounds.Descent + run.LineBounds.Gap
		}
	}
	line.lineHeight = maxSize
	computeVisualOrder(&line)
	return line
}

// computeVisualOrder will populate the Line's VisualOrder field and the
// VisualPosition field of each element in Runs.
func computeVisualOrder(l *line) {
	l.visualOrder = make([]int, len(l.runs))
	const none = -1
	bidiRangeStart := none

	// visPos returns the visual position for an individual logically-indexed
	// run in this line, taking only the line's overall text direction into
	// account.
	visPos := func(logicalIndex int) int {
		if l.direction.Progression() == system.TowardOrigin {
			return len(l.runs) - 1 - logicalIndex
		}
		return logicalIndex
	}

	// resolveBidi populated the line's VisualOrder fields for the elements in the
	// half-open range [bidiRangeStart:bidiRangeEnd) indicating that those elements
	// should be displayed in reverse-visual order.
	resolveBidi := func(bidiRangeStart, bidiRangeEnd int) {
		firstVisual := bidiRangeEnd - 1
		// Just found the end of a bidi range.
		for startIdx := bidiRangeStart; startIdx < bidiRangeEnd; startIdx++ {
			pos := visPos(firstVisual)
			l.runs[startIdx].VisualPosition = pos
			l.visualOrder[pos] = startIdx
			firstVisual--
		}
		bidiRangeStart = none
	}
	for runIdx, run := range l.runs {
		if run.Direction.Progression() != l.direction.Progression() {
			if bidiRangeStart == none {
				bidiRangeStart = runIdx
			}
			continue
		} else if bidiRangeStart != none {
			// Just found the end of a bidi range.
			resolveBidi(bidiRangeStart, runIdx)
			bidiRangeStart = none
		}
		pos := visPos(runIdx)
		l.runs[runIdx].VisualPosition = pos
		l.visualOrder[pos] = runIdx
	}
	if bidiRangeStart != none {
		// We ended iteration within a bidi segment, resolve it.
		resolveBidi(bidiRangeStart, len(l.runs))
	}
	// Iterate and resolve the X of each run.
	x := fixed.Int26_6(0)
	for _, runIdx := range l.visualOrder {
		l.runs[runIdx].X = x
		x += l.runs[runIdx].Advance
	}
}
