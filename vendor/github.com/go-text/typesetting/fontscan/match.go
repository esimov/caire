package fontscan

import (
	"sort"

	meta "github.com/go-text/typesetting/opentype/api/metadata"
)

// Query exposes the intention of an author about the
// font to use to shape and render text.
type Query struct {
	// Families is a list of required families,
	// the first having the highest priority.
	// Each of them is tried until a suitable match is found.
	Families []string

	// Aspect selects which particular face to use among
	// the font matching the family criteria.
	Aspect meta.Aspect
}

// fontSet stores the list of fonts available for text shaping.
// It is usually build from a system font index or by manually appending
// fonts.
// footprint family names are normalized
type fontSet []footprint

// stores the possible matches with their score:
// lower is better
type familyCrible map[string]int

// clear fc but keep the underlying storage
func (fc familyCrible) reset() {
	for k := range fc {
		delete(fc, k)
	}
}

// fillWithSubstitutions starts from `family`
// and applies all the substitutions coded in the package
// to add substitutes values
func (fc familyCrible) fillWithSubstitutions(family string) {
	fl := newFamilyList([]string{family})
	for _, subs := range familySubstitution {
		fl.execute(subs)
	}

	fl.compileTo(fc)
}

type scoredFootprints struct {
	footprints []int
	scores     []int

	database fontSet
}

// keep the underlying storage
func (sf *scoredFootprints) reset(fs fontSet) {
	sf.footprints = sf.footprints[:0]
	sf.scores = sf.scores[:0]

	sf.database = fs
}

// Len is the number of elements in the collection.
func (sf scoredFootprints) Len() int { return len(sf.footprints) }

func (sf scoredFootprints) Less(i int, j int) bool {
	if sf.scores[i] < sf.scores[j] {
		return true
	} else if sf.scores[i] > sf.scores[j] {
		return false
	} else {
		indexi, indexj := sf.footprints[i], sf.footprints[j]
		return sf.database[indexi].isUserProvided && !sf.database[indexj].isUserProvided
	}
}

// Swap swaps the elements with indexes i and j.
func (sf scoredFootprints) Swap(i int, j int) {
	sf.footprints[i], sf.footprints[j] = sf.footprints[j], sf.footprints[i]
	sf.scores[i], sf.scores[j] = sf.scores[j], sf.scores[i]
}

// Generic families as defined by
// https://www.w3.org/TR/css-fonts-4/#generic-font-families
const (
	Fantasy   = "fantasy"
	Math      = "math"
	Emoji     = "emoji"
	Serif     = "serif"
	SansSerif = "sans-serif"
	Cursive   = "cursive"
	Monospace = "monospace"
)

func isGenericFamily(family string) bool {
	switch family {
	case Serif, SansSerif, Monospace, Cursive, Fantasy, Math, Emoji:
		return true
	default:
		return false
	}
}

// selectByFamily returns all the fonts in the fontmap matching
// the given `family`, with the best matches coming first.
// `substitute` controls whether or not system substitutions are applied.
// The generic families are always expanded to concrete families.
//
// If two fonts have the same family, user provided are returned first.
//
// The returned slice may be empty if no font matches the given `family`.
// buffer is used to reduce allocations
func (fm fontSet) selectByFamily(family string, substitute bool,
	footprintBuffer *scoredFootprints, cribleBuffer familyCrible,
) []int {
	// build the crible, handling substitutions
	family = meta.NormalizeFamily(family)

	footprintBuffer.reset(fm)
	cribleBuffer.reset()

	// always substitute generic families
	if substitute || isGenericFamily(family) {
		cribleBuffer.fillWithSubstitutions(family)
	} else {
		cribleBuffer = familyCrible{family: 0}
	}

	// select the matching fonts:
	// loop through `footprints` and stores the matching fonts into `dst`
	for index, footprint := range fm {
		if score, has := cribleBuffer[footprint.Family]; has {
			footprintBuffer.footprints = append(footprintBuffer.footprints, index)
			footprintBuffer.scores = append(footprintBuffer.scores, score)
		}
	}

	// sort the matched font by score (lower is better)
	sort.Stable(*footprintBuffer)

	return footprintBuffer.footprints
}

// matchStretch look for the given stretch in the font set,
// or, if not found, the closest stretch
// if always return a valid value (contained in `candidates`) if `candidates` is not empty
func (fs fontSet) matchStretch(candidates []int, query meta.Stretch) meta.Stretch {
	// narrower and wider than the query
	var narrower, wider meta.Stretch

	for _, index := range candidates {
		stretch := fs[index].Aspect.Stretch
		if stretch > query { // wider candidate
			if wider == 0 || stretch-query < wider-query { // closer
				wider = stretch
			}
		} else if stretch < query { // narrower candidate
			// if narrower == 0, it is always more distant to queryStretch than stretch
			if query-stretch < query-narrower { // closer
				narrower = stretch
			}
		} else {
			// found an exact match, just return it
			return query
		}
	}

	// default to closest
	if query <= meta.StretchNormal { // narrow first
		if narrower != 0 {
			return narrower
		}
		return wider
	} else { // wide first
		if wider != 0 {
			return wider
		}
		return narrower
	}
}

// in practice, italic and oblique are synonymous
const styleOblique = meta.StyleItalic

// matchStyle look for the given style in the font set,
// or, if not found, the closest style
// if always return a valid value (contained in `fs`) if `fs` is not empty
func (fs fontSet) matchStyle(candidates []int, query meta.Style) meta.Style {
	var crible [meta.StyleItalic + 1]bool

	for _, index := range candidates {
		crible[fs[index].Aspect.Style] = true
	}

	switch query {
	case meta.StyleNormal: // StyleNormal, StyleOblique, StyleItalic
		if crible[meta.StyleNormal] {
			return meta.StyleNormal
		} else if crible[styleOblique] {
			return styleOblique
		} else {
			return meta.StyleItalic
		}
	case meta.StyleItalic: // StyleItalic, StyleOblique, StyleNormal
		if crible[meta.StyleItalic] {
			return meta.StyleItalic
		} else if crible[styleOblique] {
			return styleOblique
		} else {
			return meta.StyleNormal
		}
	}

	panic("should not happen") // query.Style is sanitized by SetDefaults
}

// matchWeight look for the given weight in the font set,
// or, if not found, the closest weight
// if always return a valid value (contained in `fs`) if `fs` is not empty
// we follow https://drafts.csswg.org/css-fonts/#font-style-matching
func (fs fontSet) matchWeight(candidates []int, query meta.Weight) meta.Weight {
	var fatter, thinner meta.Weight // approximate match
	for _, index := range candidates {
		weight := fs[index].Aspect.Weight
		if weight > query { // fatter candidate
			if fatter == 0 || weight-query < fatter-query { // weight is closer to query
				fatter = weight
			}
		} else if weight < query {
			if query-weight < query-thinner { // weight is closer to query
				thinner = weight
			}
		} else {
			// found an exact match, just return it
			return query
		}
	}

	// approximate match
	if 400 <= query && query <= 500 { // fatter until 500, then thinner then fatter
		if fatter != 0 && fatter <= 500 {
			return fatter
		} else if thinner != 0 {
			return thinner
		}
		return fatter
	} else if query < 400 { // thinner then fatter
		if thinner != 0 {
			return thinner
		}
		return fatter
	} else { // fatter then thinner
		if fatter != 0 {
			return fatter
		}
		return thinner
	}
}

// filter `candidates` in place and returns the updated slice
func (fs fontSet) filterByStretch(candidates []int, stretch meta.Stretch) []int {
	n := 0
	for _, index := range candidates {
		if fs[index].Aspect.Stretch == stretch {
			candidates[n] = index
			n++
		}
	}
	candidates = candidates[:n]
	return candidates
}

// filter `candidates` in place and returns the updated slice
func (fs fontSet) filterByStyle(candidates []int, style meta.Style) []int {
	n := 0
	for _, index := range candidates {
		if fs[index].Aspect.Style == style {
			candidates[n] = index
			n++
		}
	}
	candidates = candidates[:n]
	return candidates
}

// filter `candidates` in place and returns the updated slice
func (fs fontSet) filterByWeight(candidates []int, weight meta.Weight) []int {
	n := 0
	for _, index := range candidates {
		if fs[index].Aspect.Weight == weight {
			candidates[n] = index
			n++
		}
	}
	candidates = candidates[:n]
	return candidates
}

// retainsBestMatches narrows `candidates` to the closest footprints to `query`, according to the CSS font rules
// `candidates` is a slice of indexed into `fs`, which is mutated and returned
// if `candidates` is not empty, the returned slice is guaranteed not to be empty
func (fs fontSet) retainsBestMatches(candidates []int, query meta.Aspect) []int {
	// this follows CSS Fonts Level 3 § 5.2 [1].
	// https://drafts.csswg.org/css-fonts-3/#font-style-matching

	query.SetDefaults()

	// First step: font-stretch
	matchingStretch := fs.matchStretch(candidates, query.Stretch)
	candidates = fs.filterByStretch(candidates, matchingStretch) // only retain matching stretch

	// Second step : font-style
	matchingStyle := fs.matchStyle(candidates, query.Style)
	candidates = fs.filterByStyle(candidates, matchingStyle)

	// Third step : font-weight
	matchingWeight := fs.matchWeight(candidates, query.Weight)
	candidates = fs.filterByWeight(candidates, matchingWeight)

	return candidates
}

// filterUserProvided selects the user inserted fonts, appending to
// `candidates`, which is returned
func (fs fontSet) filterUserProvided(candidates []int) []int {
	for index, fp := range fs {
		if fp.isUserProvided {
			candidates = append(candidates, index)
		}
	}
	return candidates
}
