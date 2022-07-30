package internal

import (
	"io"

	"gioui.org/io/system"
	"gioui.org/text"
	"github.com/benoitkugler/textlayout/language"
	"github.com/gioui/uax/segment"
	"github.com/gioui/uax/uax14"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

// computeGlyphClusters populates the Clusters field of a Layout.
// The order of the clusters is visual, meaning
// that the first cluster is the leftmost cluster displayed even when
// the cluster is part of RTL text.
func computeGlyphClusters(l *text.Layout) {
	clusters := make([]text.GlyphCluster, 0, len(l.Glyphs)+1)
	if len(l.Glyphs) < 1 {
		if l.Runes.Count > 0 {
			// Empty line corresponding to a newline character.
			clusters = append(clusters, text.GlyphCluster{
				Runes: text.Range{
					Count:  1,
					Offset: l.Runes.Offset,
				},
			})
		}
		l.Clusters = clusters
		return
	}
	rtl := l.Direction == system.RTL

	// Check for trailing whitespace characters and synthesize
	// GlyphClusters to represent them.
	lastGlyph := l.Glyphs[len(l.Glyphs)-1]
	if rtl {
		lastGlyph = l.Glyphs[0]
	}
	trailingNewline := lastGlyph.ClusterIndex+lastGlyph.RuneCount < l.Runes.Count+l.Runes.Offset
	newlineCluster := text.GlyphCluster{
		Runes: text.Range{
			Count:  1,
			Offset: l.Runes.Count + l.Runes.Offset - 1,
		},
		Glyphs: text.Range{
			Offset: len(l.Glyphs),
		},
	}

	var (
		i               int = 0
		inc             int = 1
		runesProcessed  int = 0
		glyphsProcessed int = 0
	)

	if rtl {
		i = len(l.Glyphs) - 1
		inc = -inc
		glyphsProcessed = len(l.Glyphs) - 1
		newlineCluster.Glyphs.Offset = 0
	}
	// Construct clusters from the line's glyphs.
	for ; i < len(l.Glyphs) && i >= 0; i += inc {
		g := l.Glyphs[i]
		xAdv := g.XAdvance * fixed.Int26_6(inc)
		for k := 0; k < g.GlyphCount-1 && k < len(l.Glyphs); k++ {
			i += inc
			xAdv += l.Glyphs[i].XAdvance * fixed.Int26_6(inc)
		}

		startRune := runesProcessed
		runeIncrement := g.RuneCount
		startGlyph := glyphsProcessed
		glyphIncrement := g.GlyphCount * inc
		if rtl {
			startGlyph = glyphsProcessed + glyphIncrement + 1
		}
		clusters = append(clusters, text.GlyphCluster{
			Advance: xAdv,
			Runes: text.Range{
				Count:  g.RuneCount,
				Offset: startRune + l.Runes.Offset,
			},
			Glyphs: text.Range{
				Count:  g.GlyphCount,
				Offset: startGlyph,
			},
		})
		runesProcessed += runeIncrement
		glyphsProcessed += glyphIncrement
	}
	// Insert synthetic clusters at the right edge of the line.
	if trailingNewline {
		clusters = append(clusters, newlineCluster)
	}
	l.Clusters = clusters
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

// mapRunesToClusterIndices returns a slice. Each index within that slice corresponds
// to an index within the runes input slice. The value stored at that index is the
// index of the glyph at the start of the corresponding glyph cluster shaped by
// harfbuzz.
func mapRunesToClusterIndices(runes []rune, glyphs []shaping.Glyph) []int {
	mapping := make([]int, len(runes))
	glyphCursor := 0
	if len(runes) == 0 {
		return nil
	}
	// If the final cluster values are lower than the starting ones,
	// the text is RTL.
	rtl := len(glyphs) > 0 && glyphs[len(glyphs)-1].ClusterIndex < glyphs[0].ClusterIndex
	if rtl {
		glyphCursor = len(glyphs) - 1
	}
	for i := range runes {
		for glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			((rtl && glyphs[glyphCursor].ClusterIndex <= i) ||
				(!rtl && glyphs[glyphCursor].ClusterIndex < i)) {
			if rtl {
				glyphCursor--
			} else {
				glyphCursor++
			}
		}
		if rtl {
			glyphCursor++
		} else if (glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			glyphs[glyphCursor].ClusterIndex > i) ||
			(glyphCursor == len(glyphs) && len(glyphs) > 1) {
			glyphCursor--
			targetClusterIndex := glyphs[glyphCursor].ClusterIndex
			for glyphCursor-1 >= 0 && glyphs[glyphCursor-1].ClusterIndex == targetClusterIndex {
				glyphCursor--
			}
		}
		if glyphCursor < 0 {
			glyphCursor = 0
		} else if glyphCursor >= len(glyphs) {
			glyphCursor = len(glyphs) - 1
		}
		mapping[i] = glyphCursor
	}
	return mapping
}

// inclusiveGlyphRange returns the inclusive range of runes and glyphs matching
// the provided start and breakAfter rune positions.
// runeToGlyph must be a valid mapping from the rune representation to the
// glyph reprsentation produced by mapRunesToClusterIndices.
// numGlyphs is the number of glyphs in the output representing the runes
// under consideration.
func inclusiveGlyphRange(start, breakAfter int, runeToGlyph []int, numGlyphs int) (glyphStart, glyphEnd int) {
	rtl := runeToGlyph[len(runeToGlyph)-1] < runeToGlyph[0]
	runeStart := start
	runeEnd := breakAfter
	if rtl {
		glyphStart = runeToGlyph[runeEnd]
		if runeStart-1 >= 0 {
			glyphEnd = runeToGlyph[runeStart-1] - 1
		} else {
			glyphEnd = numGlyphs - 1
		}
	} else {
		glyphStart = runeToGlyph[runeStart]
		if runeEnd+1 < len(runeToGlyph) {
			glyphEnd = runeToGlyph[runeEnd+1] - 1
		} else {
			glyphEnd = numGlyphs - 1
		}
	}
	return
}

// breakOption represets a location within the rune slice at which
// it may be safe to break a line of text.
type breakOption struct {
	// breakAtRune is the index at which it is safe to break.
	breakAtRune int
	// penalty is the cost of breaking at this index. Negative
	// penalties mean that the break is beneficial, and a penalty
	// of uax14.PenaltyForMustBreak means a required break.
	penalty int
}

// getBreakOptions returns a slice of line break candidates for the
// text in the provided slice.
func getBreakOptions(text []rune) []breakOption {
	// Collect options for breaking the lines in a slice.
	var options []breakOption
	const adjust = -1
	breaker := uax14.NewLineWrap()
	segmenter := segment.NewSegmenter(breaker)
	segmenter.InitFromSlice(text)
	runeOffset := 0
	brokeAtEnd := false
	for segmenter.Next() {
		penalty, _ := segmenter.Penalties()
		// Determine the indices of the breaking runes in the runes
		// slice. Would be nice if the API provided this.
		currentSegment := segmenter.Runes()
		runeOffset += len(currentSegment)

		// Collect all break options.
		options = append(options, breakOption{
			penalty:     penalty,
			breakAtRune: runeOffset + adjust,
		})
		if options[len(options)-1].breakAtRune == len(text)-1 {
			brokeAtEnd = true
		}
	}
	if len(text) > 0 && !brokeAtEnd {
		options = append(options, breakOption{
			penalty:     uax14.PenaltyForMustBreak,
			breakAtRune: len(text) - 1,
		})
	}
	return options
}

type Shaper func(shaping.Input) (shaping.Output, error)

// paragraph shapes a single paragraph of text, breaking it into multiple lines
// to fit within the provided maxWidth.
func paragraph(shaper Shaper, face font.Face, ppem fixed.Int26_6, maxWidth int, lc langConfig, paragraph []rune) ([]output, error) {
	// TODO: handle splitting bidi text here

	// Shape the text.
	input := toInput(face, ppem, lc, paragraph)
	out, err := shaper(input)
	if err != nil {
		return nil, err
	}
	// Get a mapping from input runes to output glyphs.
	runeToGlyph := mapRunesToClusterIndices(paragraph, out.Glyphs)

	// Fetch line break candidates.
	breaks := getBreakOptions(paragraph)

	return lineWrap(out, input.Direction, paragraph, runeToGlyph, breaks, maxWidth), nil
}

// shouldKeepSegmentOnLine decides whether the segment of text from the current
// end of the line to the provided breakOption should be kept on the current
// line. It should be called successively with each available breakOption,
// and the line should be broken (without keeping the current segment)
// whenever it returns false.
//
// The parameters require some explanation:
// out - the shaping.Output that is being line-broken.
// runeToGlyph - a mapping where accessing the slice at the index of a rune
//     int out will yield the index of the first glyph corresponding to that rune.
// lineStartRune - the index of the first rune in the line.
// b - the line break candidate under consideration.
// curLineWidth - the amount of space total in the current line.
// curLineUsed - the amount of space in the current line that is already used.
// nextLineWidth - the amount of space available on the next line.
//
// This function returns both a valid shaping.Output broken at b and a boolean
// indicating whether the returned output should be used.
func shouldKeepSegmentOnLine(out shaping.Output, runeToGlyph []int, lineStartRune int, b breakOption, curLineWidth, curLineUsed, nextLineWidth int) (candidateLine shaping.Output, keep bool) {
	// Convert the break target to an inclusive index.
	glyphStart, glyphEnd := inclusiveGlyphRange(lineStartRune, b.breakAtRune, runeToGlyph, len(out.Glyphs))

	// Construct a line out of the inclusive glyph range.
	candidateLine = out
	candidateLine.Glyphs = candidateLine.Glyphs[glyphStart : glyphEnd+1]
	candidateLine.RecomputeAdvance()
	candidateAdvance := candidateLine.Advance.Ceil()
	if candidateAdvance > curLineWidth && candidateAdvance-curLineUsed <= nextLineWidth {
		// If it fits on the next line, put it there.
		return candidateLine, false
	}

	return candidateLine, true
}

// lineWrap wraps the shaped glyphs of a paragraph to a particular max width.
func lineWrap(out shaping.Output, dir di.Direction, paragraph []rune, runeToGlyph []int, breaks []breakOption, maxWidth int) []output {
	var outputs []output
	if len(breaks) == 0 {
		// Pass empty lines through as empty.
		outputs = append(outputs, output{
			Shaped: out,
			RuneRange: text.Range{
				Count: len(paragraph),
			},
		})
		return outputs
	}

	for i := 0; i < len(breaks); i++ {
		b := breaks[i]
		if b.breakAtRune+1 < len(runeToGlyph) {
			// Check if this break is valid.
			gIdx := runeToGlyph[b.breakAtRune]
			g2Idx := runeToGlyph[b.breakAtRune+1]
			cIdx := out.Glyphs[gIdx].ClusterIndex
			c2Idx := out.Glyphs[g2Idx].ClusterIndex
			if cIdx == c2Idx {
				// This break is within a harfbuzz cluster, and is
				// therefore invalid.
				copy(breaks[i:], breaks[i+1:])
				breaks = breaks[:len(breaks)-1]
				i--
			}
		}
	}

	start := 0
	runesProcessed := 0
	for i := 0; i < len(breaks); i++ {
		b := breaks[i]
		// Always keep the first segment on a line.
		good, _ := shouldKeepSegmentOnLine(out, runeToGlyph, start, b, maxWidth, 0, maxWidth)
		end := b.breakAtRune
	innerLoop:
		for k := i + 1; k < len(breaks); k++ {
			bb := breaks[k]
			candidate, ok := shouldKeepSegmentOnLine(out, runeToGlyph, start, bb, maxWidth, good.Advance.Ceil(), maxWidth)
			if ok {
				// Use this new, longer segment.
				good = candidate
				end = bb.breakAtRune
				i++
			} else {
				break innerLoop
			}
		}
		numRunes := end - start + 1
		outputs = append(outputs, output{
			Shaped: good,
			RuneRange: text.Range{
				Count:  numRunes,
				Offset: runesProcessed,
			},
		})
		runesProcessed += numRunes
		start = end + 1
	}
	return outputs
}

// output is a run of shaped text with metadata about its position
// within a text document.
type output struct {
	Shaped    shaping.Output
	RuneRange text.Range
}

func toSystemDirection(d di.Direction) system.TextDirection {
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
func toGioGlyphs(in []shaping.Glyph) []text.Glyph {
	out := make([]text.Glyph, 0, len(in))
	for _, g := range in {
		out = append(out, text.Glyph{
			ID:           g.GlyphID,
			ClusterIndex: g.ClusterIndex,
			RuneCount:    g.RuneCount,
			GlyphCount:   g.GlyphCount,
			XAdvance:     g.XAdvance,
			YAdvance:     g.YAdvance,
			XOffset:      g.XOffset,
			YOffset:      g.YOffset,
		})
	}
	return out
}

// ToLine converts the output into a text.Line
func (o output) ToLine() text.Line {
	layout := text.Layout{
		Glyphs:    toGioGlyphs(o.Shaped.Glyphs),
		Runes:     o.RuneRange,
		Direction: toSystemDirection(o.Shaped.Direction),
	}
	return text.Line{
		Layout: layout,
		Bounds: fixed.Rectangle26_6{
			Min: fixed.Point26_6{
				Y: -o.Shaped.LineBounds.Ascent,
			},
			Max: fixed.Point26_6{
				X: o.Shaped.Advance,
				Y: -o.Shaped.LineBounds.Ascent + o.Shaped.LineBounds.LineHeight(),
			},
		},
		Width:   o.Shaped.Advance,
		Ascent:  o.Shaped.LineBounds.Ascent,
		Descent: -o.Shaped.LineBounds.Descent + o.Shaped.LineBounds.Gap,
	}
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

// Document shapes text using the given font, ppem, maximum line width, language,
// and sequence of runes. It returns a slice of lines corresponding to the txt,
// broken to fit within maxWidth and on paragraph boundaries.
func Document(shaper Shaper, face font.Face, ppem fixed.Int26_6, maxWidth int, lc system.Locale, txt io.RuneReader) []text.Line {
	var (
		outputs       []text.Line
		startByte     int
		startRune     int
		paragraphText []rune
		done          bool
		langs         = make(map[language.Script]int)
	)
	for !done {
		var (
			bytes int
			runes int
		)
		newlineAdjust := 0
	paragraphLoop:
		for r, sz, re := txt.ReadRune(); !done; r, sz, re = txt.ReadRune() {
			if re != nil {
				done = true
				continue
			}
			paragraphText = append(paragraphText, r)
			script := language.LookupScript(r)
			langs[script]++
			bytes += sz
			runes++
			if r == '\n' {
				newlineAdjust = 1
				break paragraphLoop
			}
		}
		var (
			primary      language.Script
			primaryTotal int
		)
		for script, total := range langs {
			if total > primaryTotal {
				primary = script
				primaryTotal = total
			}
		}
		if lc.Language == "" {
			lc.Language = "EN"
		}
		lcfg := langConfig{
			Language:  language.NewLanguage(lc.Language),
			Script:    primary,
			Direction: mapDirection(lc.Direction),
		}
		lines, _ := paragraph(shaper, face, ppem, maxWidth, lcfg, paragraphText[:len(paragraphText)-newlineAdjust])
		for i := range lines {
			// Update the offsets of each paragraph to be correct within the
			// whole document.
			lines[i].RuneRange.Offset += startRune
			// Update the cluster values to be rune indices within the entire
			// document.
			for k := range lines[i].Shaped.Glyphs {
				lines[i].Shaped.Glyphs[k].ClusterIndex += startRune
			}
			outputs = append(outputs, lines[i].ToLine())
		}
		// If there was a trailing newline update the byte counts to include
		// it on the last line of the paragraph.
		if newlineAdjust > 0 {
			outputs[len(outputs)-1].Layout.Runes.Count += newlineAdjust
		}
		paragraphText = paragraphText[:0]
		startByte += bytes
		startRune += runes
	}
	for i := range outputs {
		computeGlyphClusters(&outputs[i].Layout)
	}
	return outputs
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
