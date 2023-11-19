package fontscan

import (
	"container/list"
	"strings"

	"github.com/go-text/typesetting/language"
	meta "github.com/go-text/typesetting/opentype/api/metadata"
)

// this file implements the family substitution feature,
// inspired by fontconfig.
// it works by defining a set of modifications to apply
// to a user provided family
// each of them may happen one (or more) alternative family to look for


// familySubstitution maps family name to possible alias
// it is generated from fontconfig substitution rules
// the order matters, since the rules apply sequentially to the current
// state of the family list
func init() {
	// replace families keys by their no case no blank version
	for i, v := range familySubstitution {
		for i, s := range v.additionalFamilies {
			v.additionalFamilies[i] = meta.NormalizeFamily(s)
		}

		familySubstitution[i].test = v.test.normalize()
	}
}

// we want to easily insert at the start,
// the end and "around" an element
type familyList struct {
	*list.List
}

func newFamilyList(families []string) familyList {
	var out list.List
	for _, s := range families {
		out.PushBack(s)
	}
	return familyList{List: &out}
}

// returns the node equal to `family` or nil, if not found
func (fl familyList) elementEquals(family string) *list.Element {
	for l := fl.List.Front(); l != nil; l = l.Next() {
		if l.Value.(string) == family {
			return l
		}
	}
	return nil
}

// returns the first node containing `family` or nil, if not found
func (fl familyList) elementContains(family string) *list.Element {
	for l := fl.List.Front(); l != nil; l = l.Next() {
		if strings.Contains(l.Value.(string), family) {
			return l
		}
	}
	return nil
}

// return the crible corresponding to the order
func (fl familyList) compileTo(dst familyCrible) {
	i := 0
	for l := fl.List.Front(); l != nil; l, i = l.Next(), i+1 {
		family := l.Value.(string)
		if _, has := dst[family]; !has { // for duplicated entries, keep the first (best) score
			dst[family] = i
		}
	}
}

func (fl familyList) insertStart(families []string) {
	L := len(families)
	for i := range families {
		fl.List.PushFront(families[L-1-i])
	}
}

func (fl familyList) insertEnd(families []string) {
	for _, s := range families {
		fl.List.PushBack(s)
	}
}

// insertAfter inserts families right after element
func (fl familyList) insertAfter(element *list.Element, families []string) {
	for _, s := range families {
		element = fl.List.InsertAfter(s, element)
	}
}

// insertBefore inserts families right before element
func (fl familyList) insertBefore(element *list.Element, families []string) {
	L := len(families)
	for i := range families {
		element = fl.List.InsertBefore(families[L-1-i], element)
	}
}

func (fl familyList) replace(element *list.Element, families []string) {
	fl.insertAfter(element, families)
	fl.List.Remove(element)
}

// ----- substitutions ------

// where to insert the families with respect to
// the current list
type substitutionOp uint8

const (
	opAppend substitutionOp = iota
	opAppendLast
	opPrepend
	opPrependFirst
	opReplace
)

type substitutionTest interface {
	// returns a non nil element if the substitution should
	// be applied
	// for opAppendLast and opPrependFirst an arbitrary non nil element
	// could be returned
	test(list familyList) *list.Element

	// return a copy where families have been normalize
	// to their no blank no case version
	normalize() substitutionTest
}

// a family in the list must equal 'mf'
type familyEquals string

func (mf familyEquals) test(list familyList) *list.Element {
	return list.elementEquals(string(mf))
}

func (mf familyEquals) normalize() substitutionTest {
	return familyEquals(meta.NormalizeFamily(string(mf)))
}

// a family in the list must contain 'mf'
type familyContains string

func (mf familyContains) test(list familyList) *list.Element {
	return list.elementContains(string(mf))
}

func (mf familyContains) normalize() substitutionTest {
	return familyContains(meta.NormalizeFamily(string(mf)))
}

// the family list has no "serif", "sans-serif" or "monospace" generic fallback
type noGenericFamily struct{}

func (noGenericFamily) test(list familyList) *list.Element {
	for l := list.List.Front(); l != nil; l = l.Next() {
		switch l.Value.(string) {
		case "serif", "sans-serif", "monospace":
			return nil
		}
	}
	return list.List.Front()
}

func (noGenericFamily) normalize() substitutionTest {
	return noGenericFamily{}
}

// one family must equals `family`, and the queried language
// must equals `lang`
type langAndFamilyEqual struct {
	lang   language.Language
	family string
}

// TODO: for now, these tests language base tests are ignored
func (langAndFamilyEqual) test(list familyList) *list.Element {
	return nil
}

func (t langAndFamilyEqual) normalize() substitutionTest {
	t.family = meta.NormalizeFamily(t.family)
	return t
}

// one family must equals `family`, and the queried language
// must contains `lang`
type langContainsAndFamilyEquals struct {
	lang   language.Language
	family string
}

// TODO: for now, these tests language base tests are ignored
func (langContainsAndFamilyEquals) test(list familyList) *list.Element {
	return nil
}

func (t langContainsAndFamilyEquals) normalize() substitutionTest {
	t.family = meta.NormalizeFamily(t.family)
	return t
}

// no family must equals `family`, and the queried language
// must equals `lang`
type langEqualsAndNoFamily struct {
	lang   language.Language
	family string
}

// TODO: for now, these tests language base tests are ignored
func (langEqualsAndNoFamily) test(list familyList) *list.Element {
	return nil
}

func (t langEqualsAndNoFamily) normalize() substitutionTest {
	t.family = meta.NormalizeFamily(t.family)
	return t
}

type substitution struct {
	test               substitutionTest // the condition to apply
	additionalFamilies []string         // the families to add
	op                 substitutionOp   // how to insert the families
}

func (fl familyList) execute(subs substitution) {
	element := subs.test.test(fl)
	if element == nil {
		return
	}

	switch subs.op {
	case opAppend:
		fl.insertAfter(element, subs.additionalFamilies)
	case opAppendLast:
		fl.insertEnd(subs.additionalFamilies)
	case opPrepend:
		fl.insertBefore(element, subs.additionalFamilies)
	case opPrependFirst:
		fl.insertStart(subs.additionalFamilies)
	case opReplace:
		fl.replace(element, subs.additionalFamilies)
	default:
		panic("exhaustive switch")
	}
}
