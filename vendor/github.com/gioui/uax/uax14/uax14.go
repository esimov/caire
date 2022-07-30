/*
Package uax14 implements Unicode Annex #14 line wrap.

Under active development; use at your own risk

Contents

UAX#14 is the Unicode Annex for Line Breaking (Line Wrap).
It defines a bunch of code-point classes and a set of rules
for how to place break points / break inhibitors.

Typical Usage

Clients instantiate a UAX#14 line breaker object and use it as the
breaking engine for a segmenter.

  breaker := uax14.NewLineWrap()
  segmenter := unicode.NewSegmenter(breaker)
  segmenter.Init(...)
  for segmenter.Next() {
    ... // do something with segmenter.Text() or segmenter.Bytes()
  }

Before using line breakers, clients usually will want to initialize the UAX#14
classes and rules.

  SetupClasses()

This initializes all the code-point range tables. Initialization is
not done beforehand, as it consumes quite some memory, and using UAX#14
is not mandatory. SetupClasses() is called automatically, however,
if clients call NewLineWrap().

Status

The current implementation passes all tests from the UAX#14 test file, except 3:

    uax14_test.go:65: 3 TEST CASES OUT of 7001 FAILED

______________________________________________________________________

License

This project is provided under the terms of the UNLICENSE or
the 3-Clause BSD license denoted by the following SPDX identifier:

SPDX-License-Identifier: 'Unlicense' OR 'BSD-3-Clause'

You may use the project under the terms of either license.

Licenses are reproduced in the license file in the root folder of this module.

Copyright © 2021 Norbert Pillmayer <norbert@pillmayer.com>


*/
package uax14

import (
	"math"
	"sync"
	"unicode"

	"github.com/gioui/uax"
	"github.com/gioui/uax/internal/tracing"
)

const (
	sot       UAX14Class = 1000 // pseudo class
	eot       UAX14Class = 1001 // pseudo class
	optSpaces UAX14Class = 1002 // pseudo class
)

// ClassForRune gets the line breaking/wrap class for a Unicode code-point
func ClassForRune(r rune) UAX14Class {
	if r == rune(0) {
		return eot
	}
	for lbc := UAX14Class(0); lbc <= ZWJClass; lbc++ {
		urange := rangeFromUAX14Class[lbc]
		if urange == nil {
			tracing.Errorf("-- no range for class %s\n", lbc)
		} else if unicode.Is(urange, r) {
			return lbc
		}
	}
	return XXClass
}

var setupOnce sync.Once

// SetupClasses is the top-level preparation function:
// Create code-point classes for UAX#14 line breaking/wrap.
// (Concurrency-safe).
func SetupClasses() {
	setupOnce.Do(setupUAX14Classes)
}

// === UAX#14 Line Breaker ==============================================

// LineWrap is a type used by a unicode.Segmenter to break lines
// up according to UAX#14. It implements the unicode.UnicodeBreaker interface.
type LineWrap struct {
	publisher    uax.RunePublisher
	longestMatch int   // longest active match of a rule
	penalties    []int // returned to the segmenter: penalties to insert
	rules        map[UAX14Class][]uax.NfaStateFn
	lastClass    UAX14Class // we have to remember the last code-point class
	blockedRI    bool       // are rules for Regional_Indicator currently blocked?
	substituted  bool       // has the code-point class been substituted?
	shadow       UAX14Class // class before substitution
}

// NewLineWrap creates a new UAX#14 line breaker.
//
// Usage:
//
//   linewrap := NewLineWrap()
//   segmenter := segment.NewSegmenter(linewrap)
//   segmenter.Init(...)
//   for segmenter.Next() ...
//
func NewLineWrap() *LineWrap {
	uax14 := &LineWrap{}
	uax14.publisher = uax.NewRunePublisher()
	uax14.rules = map[UAX14Class][]uax.NfaStateFn{
		//sot:      {rule_LB2},
		eot:      {rule_LB3},
		NLClass:  {rule_05_NewLine, rule_06_HardBreak},
		LFClass:  {rule_05_NewLine, rule_06_HardBreak},
		BKClass:  {rule_05_NewLine, rule_06_HardBreak},
		CRClass:  {rule_05_NewLine, rule_06_HardBreak},
		SPClass:  {rule_LB7, rule_LB18},
		ZWClass:  {rule_LB7, rule_LB8},
		WJClass:  {rule_LB11},
		GLClass:  {rule_LB12},
		CLClass:  {rule_LB13, rule_LB16},
		CPClass:  {rule_LB13, rule_LB16, rule_LB30_2},
		EXClass:  {rule_LB13, rule_LB22},
		ISClass:  {rule_LB13, rule_LB29},
		SYClass:  {rule_LB13, rule_LB21b},
		OPClass:  {rule_LB14, step2_LB25},
		QUClass:  {rule_LB15, rule_LB19},
		B2Class:  {rule_LB17},
		BAClass:  {rule_LB21},
		CBClass:  {rule_LB20},
		HYClass:  {rule_LB21, step2_LB25},
		NSClass:  {rule_LB21},
		BBClass:  {rule_LB21x},
		ALClass:  {rule_LB22, rule_LB23_1, rule_LB24_2, rule_LB28, rule_LB30_1},
		HLClass:  {rule_LB21a, rule_LB22, rule_LB23_1, rule_LB24_2, rule_LB28, rule_LB30_1},
		IDClass:  {rule_LB22, rule_LB23a_2},
		EBClass:  {rule_LB22, rule_LB23a_2, rule_LB30b},
		EMClass:  {rule_LB22, rule_LB23a_2},
		INClass:  {rule_LB22},
		NUClass:  {rule_LB22, rule_LB23_2, step3_LB25, rule_LB30_1},
		RIClass:  {rule_LB30a},
		PRClass:  {rule_LB23a_1, rule_LB24_1, rule_LB25, rule_LB27_2},
		POClass:  {rule_LB24_1, rule_LB25},
		JLClass:  {rule_LB26_1, rule_LB27},
		JVClass:  {rule_LB26_2, rule_LB27},
		H2Class:  {rule_LB26_2, rule_LB27},
		JTClass:  {rule_LB26_3, rule_LB27},
		H3Class:  {rule_LB26_3, rule_LB27},
		ZWJClass: {rule_LB8a},
	}
	if rangeFromUAX14Class == nil {
		tracing.Infof("UAX#14 classes not yet initialized -> initializing")
	}
	SetupClasses()
	uax14.lastClass = sot
	return uax14
}

// CodePointClassFor returns the UAX#14 code-point class for a rune (= code-point).
//
// Interface unicode.UnicodeBreaker
func (uax14 *LineWrap) CodePointClassFor(r rune) int {
	c := ClassForRune(r)
	c = resolveSomeClasses(r, c)
	cnew, shadow := substitueSomeClasses(c, uax14.lastClass)
	uax14.substituted = (c != cnew)
	uax14.shadow = shadow
	return int(cnew)
}

// StartRulesFor starts all recognizers where the starting symbol is rune r.
// r is of code-point-class cpClass.
//
// Interface unicode.UnicodeBreaker
func (uax14 *LineWrap) StartRulesFor(r rune, cpClass int) {
	c := UAX14Class(cpClass)
	if c != RIClass || !uax14.blockedRI {
		if rules := uax14.rules[c]; len(rules) > 0 {
			tracing.P("class", c).Debugf("starting %d rule(s) for class %s", len(rules), c)
			for _, rule := range rules {
				rec := uax.NewPooledRecognizer(cpClass, rule)
				rec.UserData = uax14
				uax14.publisher.SubscribeMe(rec)
			}
		} else {
			tracing.P("class", c).Debugf("starting no rule")
		}
		/*
			if uax14.shadow == ZWJClass {
				if rules := uax14.rules[uax14.shadow]; len(rules) > 0 {
					TC.P("class", c).Debugf("starting %d rule(s) for shadow class ZWJ", len(rules))
					for _, rule := range rules {
						rec := uax.NewPooledRecognizer(cpClass, rule)
						rec.UserData = uax14
						uax14.publisher.SubscribeMe(rec)
					}
				}
			}
		*/
	}
}

// LB1 Assign a line breaking class to each code point of the input.
// Resolve AI, CB, CJ, SA, SG, and XX into other line breaking classes
// depending on criteria outside the scope of this algorithm.
//
// In the absence of such criteria all characters with a specific combination of
// original class and General_Category property value are resolved as follows:
//
//   Resolved 	Original 	 General_Category
//   AL         AI, SG, XX  Any
//   CM         SA          Only Mn or Mc
//   AL         SA          Any except Mn and Mc
//   NS         CJ          Any
//
func resolveSomeClasses(r rune, c UAX14Class) UAX14Class {
	if c == AIClass || c == SGClass || c == XXClass {
		return ALClass
	} else if c == SAClass {
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) {
			return CMClass
		}
		return ALClass
	} else if c == CJClass {
		return NSClass
	}
	return c
}

// LB9: Do not break a combining character sequence;
// treat it as if it has the line breaking class of the base character in all
// of the following rules. Treat ZWJ as if it were CM.
//
//    X (CM | ZWJ)* ⟼ X.
//
// where X is any line break class except BK, CR, LF, NL, SP, or ZW.
//
// LB10: Treat any remaining combining mark or ZWJ as AL.
func substitueSomeClasses(c UAX14Class, lastClass UAX14Class) (UAX14Class, UAX14Class) {
	shadow := c
	switch lastClass {
	case sot, BKClass, CRClass, LFClass, NLClass, SPClass, ZWClass:
		if c == CMClass || c == ZWJClass {
			c = ALClass
		}
	default:
		if c == CMClass || c == ZWJClass {
			c = lastClass
		}
	}
	if shadow != c {
		tracing.Debugf("subst %+q -> %+q", shadow, c)
	}
	return c, shadow
}

const (
	noBreak int = 10000
	doBreak int = -10000
)

// UAX#14 has lots of rules where a penalty is to be inserted before a
// code point without a prefix to start a rule. An example would be
//
//    LB7: Do not break before spaces or zero width space.
//         × SP
//         × ZW
//
// These are kind of 'after-the-fact' rules, as the position before the SP may
// already have been broken. This is possible if no match has been built up and
// therefore it's too late to insert a penalty for LB7.
// To provide a prefix for rules like these, they would have to be modified to
//
//         any × SP
//
// therefore introducing an any-rule with match length 1 which looks out for all
// the prefix-less rules (continuations). While this is certainly possible, we
// should alleviate the author of UAX#14 rules from such concerns and rather
// modify the general segmentation driver to allow for a delayed evaluation of the
// rightmost penalty / break opportunity. We achieve this by a simple trick:
// The uax14 breaker will always report a match length of at least 1. This works
// for two reasons: The UAX#14 rules which prohibit breaks 'after-the-fact' are
// all just one code-point long. And the segmenter will always append an artifical
// end-of-text which will then be the last pretended match and therefore allow a
// break just before it.

// ProceedWithRune is part of interface unicode.Breaker.
// A new code-point has been read and this breaker receives a message to
// consume it.
func (uax14 *LineWrap) ProceedWithRune(r rune, cpClass int) {
	c := UAX14Class(cpClass)
	uax14.longestMatch, uax14.penalties = uax14.publisher.PublishRuneEvent(r, int(c))
	x := uax14.penalties
	//fmt.Printf("   x = %v\n", x)
	if uax14.substituted && uax14.lastClass == c { // do not break between runes for rule 09
		if len(x) > 1 && x[1] == 0 {
			x[1] = noBreak
		} else if len(x) == 1 {
			x = append(x, noBreak)
		} else if len(x) == 0 {
			x = make([]int, 2)
			x[1] = noBreak
		}
	} else {
		x = setPenalty1(x, DefaultPenalty)
	}
	for i, p := range x { // positive penalties get lifted +1000
		if p > DefaultPenalty {
			p += noBreak
			x[i] = p
		}
	}
	//fmt.Printf("=> x = %v\n", x)
	uax14.penalties = x
	if c == eot { // start all over again
		c = sot
	}
	uax14.lastClass = c
}

// LongestActiveMatch is part of interface unicode.UnicodeBreaker
func (uax14 *LineWrap) LongestActiveMatch() int {
	// We return a value of at least 1, as explained above.
	return max(1, uax14.longestMatch)
}

// Penalties gets all active penalties for all active recognizers combined.
// Index 0 belongs to the most recently read rune.
//
// Interface unicode.UnicodeBreaker
func (uax14 *LineWrap) Penalties() []int {
	return uax14.penalties
}

// Helper: do not start any recognizers for class RI, until
// unblocked again.
func (uax14 *LineWrap) block() {
	uax14.blockedRI = true
}

// Helper: stop blocking new recognizers for class RI.
func (uax14 *LineWrap) unblock() {
	uax14.blockedRI = false
}

// Penalties (suppress break and mandatory break).
var (
	PenaltyToSuppressBreak = 10000  // Suppress break: ×
	PenaltyForMustBreak    = -19000 // Break: !
	DefaultPenalty         = 1      // Rule LB31: ÷    fragile, do not change!
)

// --- Helpers ---------------------------------------------------------------

// This is a small function to return a penalty value for a rule.
// w is the weight of the rule (currently I use the rule number
// directly).
func p(w int) int {
	q := 31 - w
	r := int(math.Pow(1.3, float64(q)))
	if r == DefaultPenalty {
		r = DefaultPenalty + 1
	}
	tracing.P("rule", w).Debugf("penalty %d => %d", w, r)
	return r
}

// Helper to create a slice of integer penalties, usually of length
// MatchLen for an accepting rule.
func ps(w int, first int, l int) []int {
	pp := make([]int, l+1)
	pp[1] = first
	p := p(w)
	for i := 2; i <= l; i++ {
		pp[i] = p
	}
	return pp
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func setPenalty1(P []int, p int) []int {
	if len(P) == 0 {
		P = append(P, 0)
		P = append(P, p)
	} else if len(P) == 1 {
		P = append(P, p)
	} else if P[1] == 0 {
		P[1] = p
	}
	return P
}
