package harfbuzz

import (
	"fmt"

	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

// ported from harfbuzz/src/hb-ot-shape-complex-myanmar.cc, .hh Copyright © 2011,2012,2013  Google, Inc.  Behdad Esfahbod

// Myanmar shaper.
type complexShaperMyanmar struct {
	complexShaperNil
}

var _ otComplexShaper = complexShaperMyanmar{}

func setMyanmarProperties(info *GlyphInfo) {
	u := info.codepoint
	type_ := indicGetCategories(u)
	cat := uint8(type_ & 0xFF)
	// pos := uint8(type_ >> 8)

	info.complexCategory = cat
	// info.complexAux = pos
}

/* Note:
 *
 * We treat Vowels and placeholders as if they were consonants.  This is safe because Vowels
 * cannot happen in a consonant syllable.  The plus side however is, we can call the
 * consonant syllable logic from the vowel syllable function and get it all right!
 *
 * Keep in sync with consonant_categories in the generator. */
const consonantFlagsMyanmar = (1 << myaSM_ex_C) | 1<<myaSM_ex_CS | myaSM_ex_Ra |
	1<<myaSM_ex_IV | 1<<myaSM_ex_GB | 1<<myaSM_ex_DOTTEDCIRCLE

func isConsonantMyanmar(info *GlyphInfo) bool {
	return isOneOf(info, consonantFlagsMyanmar)
}

/*
 * Basic features.
 * These features are applied in order, one at a time, after reordering.
 */
var myanmarBasicFeatures = [...]tables.Tag{
	loader.NewTag('r', 'p', 'h', 'f'),
	loader.NewTag('p', 'r', 'e', 'f'),
	loader.NewTag('b', 'l', 'w', 'f'),
	loader.NewTag('p', 's', 't', 'f'),
}

/*
* Other features.
* These features are applied all at once, after clearing syllables.
 */
var myanmarOtherFeatures = [...]tables.Tag{
	loader.NewTag('p', 'r', 'e', 's'),
	loader.NewTag('a', 'b', 'v', 's'),
	loader.NewTag('b', 'l', 'w', 's'),
	loader.NewTag('p', 's', 't', 's'),
}

func (complexShaperMyanmar) collectFeatures(plan *otShapePlanner) {
	map_ := &plan.map_

	/* Do this before any lookups have been applied. */
	map_.addGSUBPause(setupSyllablesMyanmar)

	map_.enableFeature(loader.NewTag('l', 'o', 'c', 'l'))
	/* The Indic specs do not require ccmp, but we apply it here since if
	* there is a use of it, it's typically at the beginning. */
	map_.enableFeature(loader.NewTag('c', 'c', 'm', 'p'))

	map_.addGSUBPause(reorderMyanmar)

	for _, feat := range myanmarBasicFeatures {
		map_.enableFeatureExt(feat, ffManualZWJ|ffPerSyllable, 1)
		map_.addGSUBPause(nil)
	}

	map_.addGSUBPause(nil)

	for _, feat := range myanmarOtherFeatures {
		map_.enableFeatureExt(feat, ffManualZWJ, 1)
	}
}

func (complexShaperMyanmar) setupMasks(_ *otShapePlan, buffer *Buffer, _ *Font) {
	// No masks, we just save information about characters.

	info := buffer.Info
	for i := range info {
		setMyanmarProperties(&info[i])
	}
}

func foundSyllableMyanmar(syllableType uint8, ts, te int, info []GlyphInfo, syllableSerial *uint8) {
	for i := ts; i < te; i++ {
		info[i].syllable = (*syllableSerial << 4) | syllableType
	}
	*syllableSerial++
	if *syllableSerial == 16 {
		*syllableSerial = 1
	}
}

func setupSyllablesMyanmar(_ *otShapePlan, _ *Font, buffer *Buffer) bool {
	findSyllablesMyanmar(buffer)
	iter, count := buffer.syllableIterator()
	for start, end := iter.next(); start < count; start, end = iter.next() {
		buffer.unsafeToBreak(start, end)
	}
	return false
}

/* Rules from:
 * https://docs.microsoft.com/en-us/typography/script-development/myanmar */
func initialReorderingConsonantSyllableMyanmar(buffer *Buffer, start, end int) {
	info := buffer.Info

	base := end
	hasReph := false

	limit := start
	if start+3 <= end &&
		info[start].complexCategory == myaSM_ex_Ra &&
		info[start+1].complexCategory == myaSM_ex_As &&
		info[start+2].complexCategory == myaSM_ex_H {
		limit += 3
		base = start
		hasReph = true
	}

	if !hasReph {
		base = limit
	}

	for i := limit; i < end; i++ {
		if isConsonantMyanmar(&info[i]) {
			base = i
			break
		}
	}

	/* Reorder! */
	i := start
	endLoop := start
	if hasReph {
		endLoop = start + 3
	}
	for ; i < endLoop; i++ {
		info[i].complexAux = posAfterMain
	}
	for ; i < base; i++ {
		info[i].complexAux = posPreC
	}
	if i < end {
		info[i].complexAux = posBaseC
		i++
	}
	var pos uint8 = posAfterMain
	/* The following loop may be ugly, but it implements all of
	 * Myanmar reordering! */
	for ; i < end; i++ {
		if info[i].complexCategory == myaSM_ex_MR /* Pre-base reordering */ {
			info[i].complexAux = posPreC
			continue
		}
		if info[i].complexCategory == myaSM_ex_VPre /* Left matra */ {
			info[i].complexAux = posPreM
			continue
		}
		if info[i].complexCategory == myaSM_ex_VS {
			info[i].complexAux = info[i-1].complexAux
			continue
		}

		if pos == posAfterMain && info[i].complexCategory == myaSM_ex_VBlw {
			pos = posBelowC
			info[i].complexAux = pos
			continue
		}

		if pos == posBelowC && info[i].complexCategory == myaSM_ex_A {
			info[i].complexAux = posBeforeSub
			continue
		}
		if pos == posBelowC && info[i].complexCategory == myaSM_ex_VBlw {
			info[i].complexAux = pos
			continue
		}
		if pos == posBelowC && info[i].complexCategory != myaSM_ex_A {
			pos = posAfterSub
			info[i].complexAux = pos
			continue
		}
		info[i].complexAux = pos
	}

	/* Sit tight, rock 'n roll! */
	buffer.sort(start, end, func(a, b *GlyphInfo) int { return int(a.complexAux) - int(b.complexAux) })

	/* Flip left-matra sequence. */
	firstLeftMatra := end
	lastLeftMatra := end
	for i := start; i < end; i++ {
		if info[i].complexAux == posPreM {
			if firstLeftMatra == end {
				firstLeftMatra = i
			}
			lastLeftMatra = i
		}
	}
	/* https://github.com/harfbuzz/harfbuzz/issues/3863 */
	if firstLeftMatra < lastLeftMatra {
		// No need to merge clusters, done already?
		buffer.reverseRange(firstLeftMatra, lastLeftMatra+1)
		// Reverse back VS, etc.
		i := firstLeftMatra
		for j := i; j <= lastLeftMatra; j++ {
			if info[j].complexCategory == myaSM_ex_VPre {
				buffer.reverseRange(i, j+1)
				i = j + 1
			}
		}
	}
}

func reorderSyllableMyanmar(buffer *Buffer, start, end int) {
	syllableType := buffer.Info[start].syllable & 0x0F
	switch syllableType {
	/* We already inserted dotted-circles, so just call the consonant_syllable. */
	case myanmarBrokenCluster, myanmarConsonantSyllable:
		initialReorderingConsonantSyllableMyanmar(buffer, start, end)
	}
}

func reorderMyanmar(_ *otShapePlan, font *Font, buffer *Buffer) bool {
	if debugMode {
		fmt.Println("MYANMAR - start reordering myanmar", buffer.Info)
	}

	ret := syllabicInsertDottedCircles(font, buffer, myanmarBrokenCluster, myaSM_ex_DOTTEDCIRCLE, -1, -1)

	iter, count := buffer.syllableIterator()
	for start, end := iter.next(); start < count; start, end = iter.next() {
		reorderSyllableMyanmar(buffer, start, end)
	}

	if debugMode {
		fmt.Println("MYANMAR - end reordering myanmar", buffer.Info)
	}

	return ret
}

func (complexShaperMyanmar) marksBehavior() (zeroWidthMarks, bool) {
	return zeroWidthMarksByGdefEarly, false
}

func (complexShaperMyanmar) normalizationPreference() normalizationMode {
	return nmComposedDiacriticsNoShortCircuit
}
