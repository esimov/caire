/*
Package segment is about Unicode text segmenting.

Typical Usage

Segmenter provides an interface similar to bufio.Scanner for reading data
such as a file of Unicode text.
Similar to Scanner's Scan() function, successive calls to a segmenter's
Next() method will step through the 'segments' of a file.
Clients are able to get runes of the segment by calling Bytes() or Text().
Unlike Scanner, segmenters are calculating a 'penalty' for breaking
at this segment. Penalties are numeric values and reflect costs, where
negative values are to be interpreted as negative costs, i.e. merits.

Clients instantiate a UnicodeBreaker object and use it as the
breaking engine for a segmenter. Multiple breaking engines may be
supplied (where the first one is called the primary breaker and any
following breaker is a secondary breaker).

    breaker1 := ...
    breaker2 := ...
    segmenter := unicode.NewSegmenter(breaker1, breaker2)
    segmenter.Init(...)
    for segmenter.Next() {
       // do something with segmenter.Text(), .Runes() or .Bytes()
    }

An example for an UnicodeBreaker is "uax29.WordBreak", a breaker
implementing the UAX#29 word breaking algorithm.

_______________________________________________________________________

License

This project is provided under the terms of the UNLICENSE or
the 3-Clause BSD license denoted by the following SPDX identifier:

SPDX-License-Identifier: 'Unlicense' OR 'BSD-3-Clause'

You may use the project under the terms of either license.

Licenses are reproduced in the license file in the root folder of this module.

Copyright © 2021 Norbert Pillmayer <norbert@pillmayer.com>

*/
package segment

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/gioui/uax"
	"github.com/gioui/uax/internal/tracing"
)

// ErrBoundReached is returned from BoundedNext() if the reason for returning false
// (and therefore stopping to iterato over the input text) is, that the bound given
// by the client has been reached.
var ErrBoundReached = errors.New("bound reached")

// A Segmenter receives a sequence of code-points from an io.RuneReader and
// segments it into smaller parts, called segments.
//
// The specification of a segment is defined by a breaker function of type
// UnicodeBreaker; the default UnicodeBreaker breaks the input into words,
// using whitespace as boundaries. For more sophisticated breakers see
// sub-packages of package uax.
type Segmenter struct {
	deque                      *deque               // where we collect runes and penalties
	reader                     io.RuneReader        // where we get the next runes from
	breakers                   []uax.UnicodeBreaker // our work horses
	runesBuf                   runewrite            // rune buffer for segment output (active segement)
	maxSegmentLen              int                  // maximum length allowed for segments
	lastPenalties              [2]int               // penalties at last break opportunity
	pos                        int                  // current position in input text
	breakOnZero                [2]bool              // treat zero value as a valid breakpoint?
	longestActiveMatch         int                  // bookkeeping for matching
	positionOfBreakOpportunity int                  // bookkeeping for matching
	atEOF                      bool                 // at end of input buffer
	inUse                      bool                 // Next() has been called; buffer is in use.
	err                        error                // collects the first error occured
}

// MaxSegmentSize is the maximum size used to buffer a segment
// unless the user provides an explicit buffer with Segmenter.Buffer().
const MaxSegmentSize = 64 * 1024
const startBufSize = 4096 // Size of initial allocation for buffer.

// ErrTooLong flags a buffer overflow.
// ErrNotInitialized is returned if a segmenters Next-function is called without
// first setting an input source.
var (
	ErrTooLong        = errors.New("UAX segmenter: segment too long for buffer")
	ErrNotInitialized = errors.New("UAX segmenter not initialized; must call Init(...) first")
)

// NewSegmenter creates a new Segmenter by providing breaking logic (UnicodeBreaker).
// Clients may provide more than one UnicodeBreaker. Specifying no
// UnicodeBreaker results in getting a SimpleWordBreaker, which will
// break on whitespace (see SimpleWordBreaker in this package).
//
// Before using newly created segmenters, clients will have to call Init(...)
// on them, i.e. initialize them for a rune reader.
func NewSegmenter(breakers ...uax.UnicodeBreaker) *Segmenter {
	s := &Segmenter{}
	if len(breakers) == 0 {
		breakers = []uax.UnicodeBreaker{NewSimpleWordBreaker()}
	}
	s.breakers = breakers
	return s
}

// Init initializes a Segmenter with an io.RuneReader to read from.
// s is either a newly created segmenter to be initialized, or
// re-initializes a segmenter already in use.
func (s *Segmenter) Init(reader io.RuneReader) {
	if reader == nil {
		reader = strings.NewReader("")
	}
	s.reader = reader
	s.runesBuf = s.runesBuf.Reset(nil)
	s.init()
}

// InitFromSlice is like Init, except using a slice of runes as an input buffer.
// s is either a newly created segmenter to be initialized, or
// re-initializes a segmenter already in use.
func (s *Segmenter) InitFromSlice(buf []rune) {
	if buf == nil {
		buf = []rune{}
	}
	s.reader = &runeread{runes: buf}
	s.runesBuf = s.runesBuf.Reset(buf)
	s.init()
}

func (s *Segmenter) init() {
	if s.deque == nil {
		s.deque = &deque{} // Q of atoms
		s.maxSegmentLen = MaxSegmentSize
	} else {
		s.deque.Clear()
		s.longestActiveMatch = 0
		s.atEOF = false
		s.inUse = false
		s.lastPenalties[0], s.lastPenalties[1] = 0, 0
		s.pos = 0
	}
	s.positionOfBreakOpportunity = -1
}

// Buffer sets the initial buffer to use when scanning and the maximum size of
// buffer that may be allocated during segmenting.
// The maximum segment size is the larger of max and cap(buf).
// If max <= cap(buf), Next() will use this buffer only and do no allocation.
//
// By default, Segmenter uses an internal buffer and sets the maximum token size
// to MaxSegmentSize.
//
// Buffer panics if it is called after scanning has started. Clients will have
// to call Init(...) again to permit re-setting the buffer.
//
// Deprecated: Use RuneBuffer instead.
func (s *Segmenter) Buffer(buf []byte, max int) {
	if s.inUse {
		panic("segment.Buffer: buffer already in use; cannot be re-set")
	}
}

// RuneBuffer sets the initial buffer to use when scanning and the maximum size of
// buffer that may be allocated during segmenting.
// The maximum segment size is the larger of max and cap(buf).
// If max <= cap(buf), Next() will use this buffer only and do no allocation.
//
// By default, Segmenter uses an internal buffer and sets the maximum token size
// to MaxSegmentSize.
//
// Buffer panics if it is called after scanning has started. Clients will have
// to call Init(...) again to permit re-setting the buffer.
func (s *Segmenter) RuneBuffer(buf []rune, max int) {
	if s.inUse {
		panic("segment.Buffer: buffer already in use; cannot be re-set")
	}
	if len(buf) > 0 {
		s.runesBuf.output = buf
		s.runesBuf.maxSegLen = max
	}
}

// Err returns the first non-EOF error that was encountered by the
// Segmenter.
func (s *Segmenter) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// BreakOnZero configures a segmenter to handle zero penalty values.
// Feasible breakpoints by definition lie between uax.InfiniteMerits and uax.InfinitePenalty.
// This would include a penalty of 0. However, I prefer to have the default zero
// value to have special semantics insofar as it should be possible for a breaker
// to express 'I do not care' with it. Should the default be breaking or not?
// The default will be to not break on zero, but clients may change this to
// accept 0 as a feasible breakpoint.
//
// The breaking policy for zero may be set differently for the primary breaker and
// secondary breakers.
func (s *Segmenter) BreakOnZero(forP1, forP2 bool) {
	s.breakOnZero[0] = forP1
	s.breakOnZero[1] = forP2
}

// Penalties >= InfinitePenalty are considered too bad for being a break opportunity.
func isPossibleBreak(p int, breakOnZero bool) bool {
	if p >= uax.InfinitePenalty {
		return false
	}
	if !breakOnZero && p == 0 {
		return false
	}
	return true
}

// Bytes returns the most recent token generated by a call to Next().
// The underlying array may point to data that will be overwritten by a
// subsequent call to Next().
func (s *Segmenter) Bytes() []byte {
	buf := bytes.Buffer{}
	for _, r := range s.runesBuf.Runes() {
		buf.WriteRune(r)
	}
	return buf.Bytes()
}

// Bytes returns the most recent token generated by a call to Next().
// The underlying array may point to data that will be overwritten by a
// subsequent call to Next().
//
// If the segmenter has been initialized with a `[]rune` input argument, Runes() will return
// a slice of it. No allocation is performed.
func (s *Segmenter) Runes() []rune {
	return s.runesBuf.Runes()
}

// Text returns the most recent segment generated by a call to Next()
// as a newly allocated string holding its bytes.
func (s *Segmenter) Text() string {
	return string(s.runesBuf.Runes())
}

// Penalties returns the last penalties a segmenter calculated.
// Two penalties are returned. The first one is the penalty returned from the
// primary breaker, the second one is the aggregate of all penalties of all the
// secondary breakers (if any).
func (s *Segmenter) Penalties() (int, int) {
	return s.lastPenalties[0], s.lastPenalties[1]
}

// Next gets the next segment, together with the accumulated penalty for this break.
//
// Next() advances the Segmenter to the next segment, which will then be available
// through the Runes(), Bytes() or Text() method. It returns false when the segmenting
// stops, either by reaching the end of the input or an error.
// After Next() returns false, the Err() method will return any error
// that occurred during scanning, except for io.EOF.
// For the latter case Err() will return nil.
//
func (s *Segmenter) Next() bool {
	return s.next(math.MaxInt)
}

// BoundedNext gets the next segment, together with the accumulated penalty for this break.
//
// BoundedNext() advances the Segmenter to the next segment, which will then be available
// through the Bytes() or Text() method. It returns false when the segmenting
// stops, either by reaching the end of the input, reaching the bound, or if an
// error occurs.
// After BoundedNext() returns false, the Err() method will return any error
// that occurred during scanning, except for io.EOF.
// For the latter case Err() will return nil.
//
// BoundedNext will return the last fragment of text before the bound, even in case
// of infinite penalties. If the client does not want this break, it is up to the
// client to merge this fragment with the next one, if any.
//
// See also method `Next`.
//
func (s *Segmenter) BoundedNext(bound int) bool {
	return s.next(bound - 1) // semantics in `next` is including, API is excluding bound
}

// setErr() records the first error encountered.
func (s *Segmenter) setErr(err error) {
	if s.err == nil || s.err == io.EOF {
		s.err = err
	}
}

// How it works
// ============
//
// The segmenter uses a double-ended queue to collect runes and the
// breaking opportunities between them. The front of the queue keeps
// adding new runes, while at the end of the queue we withdraw segments
// as soon as they are available.
// (see https://github.com/npillmayer/gotype/wiki).
//
// For every rune r read, the segmenter will fire up all the rules which
// start with r. It is not uncommon that the lifetime of a lot of rules
// overlap and all those rules are adding breaking information.
//
// The main loop to iterate over rune input will consume input runes, which are
// wrapped into queue (Q) atoms and appended to the Q. After every rune consumed,
// all the active breakers are triggered to advance their state automata and report
// possible breakpoints. Breakpoints are then attached by the segmenter to the
// atoms of the Q. At any given moment the Q may look something like this:
//
//    |Q|=3:  ← ['y' p=0|0] ← ['o' p=0|0] ← ['u' p=0|900] ← ['!' p=-200|0]
//
// In this case, the input string "you!" has been consumed and the secondary breaker
// has reported a penalty of 900 for a break after 'u', while the primary breaker
// has reported a merit of -200 for a break after '?'.
//
// The segmenter algorithm scans the Q from the front and searches for the next
// feasible breakpoint (in this case a rather undesirable one after 'u').
//
// Things get complicated by the need for book-keeping of active matches: different
// NFAs may be in the process of matching different runs of runes. After a match
// they are allowed to insert penalties anywhere between runes of the active match.
// Therefore the segmenter has to keep information about the currently longest active
// match and may not output segments including parts of this active match.
//
// The second complication arises from a feature to bound the run of runes to
// find breakpoints in: An argument may be given to not break after a certain
// position in the text. BoundedNext() should return false as soon as the
// start of the longest active match has passed the bound (and we're therefore
// sure not to detect possible breaks anywhere before the bound).

// next advances the input pointer until a possible break point has been
// found or alternatively the bound has been reached while trying.
func (s *Segmenter) next(bound int) bool {
	if s.reader == nil {
		s.setErr(ErrNotInitialized)
	}
	//
	s.inUse = true
	if !s.atEOF {
		err := s.readEnoughInputAndFindBreak(bound)
		if err != nil && err != io.EOF {
			s.setErr(err)
			s.runesBuf = s.runesBuf.Reset(nil)
			return false
		}
		if s.pos-s.deque.Len() > bound {
			tracing.Debugf("exiting because of bound reached: %d-%d>%d", s.pos, s.deque.Len(), bound)
			return false
		}
	}
	//
	if s.positionOfBreakOpportunity < 0 { // didn't find a break opportunity
		s.runesBuf = s.runesBuf.Reset(nil)
		return false
	}
	//l, _ := s.getFrontSegment(bound)
	s.getFrontSegment(bound)
	// tracing.Debugf("s.positionOfBreakOpp=%d", s.positionOfBreakOpportunity)
	//tracing.P("length", strconv.Itoa(l)).Debugf("Next() = \"%s\"", s.runesBuf)
	return true
}

// readRune reads in the next rune and appends it at the end of the Q.
// If EOF is encountered, s.atEOF is set and an artificial EOF atom is appended to the Q.
// If an error occurs during read, it is returned and s.atEOF is set
// (EOF is not considered an error in itself in this context).
// For all subsequent calls to readRune where s.atEOF is true, io.EOF is returned.
//
// s.pos is incremented according to the byte size of the input rune.
//
// readRune _must_ grow the Q or return an error or EOF.
//
func (s *Segmenter) readRune() error {
	if s.atEOF {
		return io.EOF
	}
	r, sz, err := s.reader.ReadRune()
	s.pos += sz
	//tracing.P("rune", r).Debugf("--------------------------------------")
	if err == nil {
		s.deque.PushBack(r, 0, 0)
		return nil
	}
	if err == io.EOF {
		s.deque.PushBack(eotAtom.r, eotAtom.penalty0, eotAtom.penalty1)
		s.atEOF = true
		err = nil
	} else { // error case, err is non-nil
		tracing.P("rune", r).Errorf("ReadRune() error: %s", err)
		s.atEOF = true
	}
	return err
}

// readEnoughInputAndFindBreak has a rather complicated logic. It iteratoes over
// the atoms in the Q, pulling runes from the input string as necessary, and
// inserts penalties aggregated from all the breakers. This results in a growing
// Q with atoms holding runes and penalties.
//
// Complication arises from the fact that searching for breaks may be bounded.
// It would be incorrect to stop scanning runes at the bound. We rather have to
// travel beyond the bound until the longest active match passes past the bound.
// We return ErrBoundReached in this case.
//
func (s *Segmenter) readEnoughInputAndFindBreak(bound int) (err error) {
	// if Q consists just of 0 rune (EOT), do nothing and return false
	//tracing.Debugf("segmenter: read enough input, EOF=%v, |Q|=%d", s.atEOF, s.deque.Len())
	for s.positionOfBreakOpportunity < 0 {
		if s.pos-s.longestActiveMatch > bound {
			tracing.Infof("segmenter: bound reached")
			return ErrBoundReached
		}
		// tracing.Debugf("pos=%d - %d = %d --> %d", s.pos, s.longestActiveMatch,
		qlen := s.deque.Len()
		from := max(0, qlen-1-s.longestActiveMatch) // current longest match limit, now old
		skipRead := false
		if from > 0 {
			// this may happen if the previous iteration hit a bound
			// or if a run of no-breaks was inserted by the breaker(s)
			//tracing.Debugf("active match is short; previous iteration did not deliver (all) breakpoints")
			if s.positionOfBreakOpportunity >= 0 {
				skipRead = true // we need not read in a rune if we had a breakpoint at bound
			}
		}
		if !skipRead { // read a rune and find new break opportunities
			if err = s.readRune(); err != nil {
				break
			}
			// if s.deque.Len() == qlen { // TODO remove this after extensive testing
			// 	panic("segmenter: code-point deque did not grow")
			// }
			// Now we've got read in a new rune and are ready to fire up the breakers for it.
			// We then harvest the new-found penalties from every breaker and insert the
			// aggregates at the atoms in the Q.
			qlen = s.deque.Len()
			s.longestActiveMatch = 0
			r := s.deque.LastRune()
			for _, breaker := range s.breakers {
				cpClass := breaker.CodePointClassFor(r)
				breaker.StartRulesFor(r, cpClass)
				breaker.ProceedWithRune(r, cpClass)
				if lam := breaker.LongestActiveMatch(); lam > s.longestActiveMatch {
					s.longestActiveMatch = lam
				}
				s.insertPenalties(s.inxForBreaker(breaker), breaker.Penalties())
			}
			//s.printQ()
			//tracing.Debugf("-- all breakers done --")
		}
		// The Q is updated. The longest active match may have been changed and
		// we have to scan from the start of the Q to the new position of active match.
		// If we find a breaking opportunity, we're done.
		boundDist := bound
		if bound < math.MaxInt {
			//boundDist = bound - s.pos + s.deque.Len()
			boundDist = bound - s.pos + qlen
		}
		s.positionOfBreakOpportunity = s.findBreakOpportunity(qlen-s.longestActiveMatch,
			boundDist)
		//s.positionOfBreakOpportunity = s.findBreakOpportunity(qlen - 1 - s.longestActiveMatch)
		// tracing.Debugf("segmenter: breakpos=%d, Q-len=%d, active match=%d",
		// 	s.positionOfBreakOpportunity, s.deque.Len(), s.longestActiveMatch)
		//s.printQ()
	}
	return err
}

// findBreakOpportunity scans a segment of the Q and will find the first atom
// to carry a breakable penalty.
// Say the Q has 4 elements from reading in "are ". A breaker has just inserted
// penalties (100;0) after 'e'. Longest active match be 1 (' ').
//
//     Q (len=4):    ['a' p=0|0] <- ['r' p=0|0] <- ['e' p=100|0] <- [' ' p=0|0]
//
// We iterate over the Q to find the first pair of penalties which are considered
// breakable. In this case we find a break opportunity of pos=2 (if standard
// segmenter parameters are active).
//
// A breaking opportunity will also occur at a given bound, if any, even in the case
// of infinite penalties at the break. Parameter `boundDist` is not an absolute position,
// but rather the distance (in runes) to the bound position.
//
// Returns the Q-position to break or -1.
func (s *Segmenter) findBreakOpportunity(to int, boundDist int) int {
	if to-1 < 0 || boundDist < 0 {
		return -1
	}
	//tracing.Debugf("segmenter: searching for break opportunity from 0 to %d|%d: ", to-1, boundDist)
	breakopp := -1
	i := 0
	for ; i < to && i <= boundDist; i++ {
		//_, p0, p1 := s.deque.At(i)
		q := s.deque.AtomAt(i)
		if isPossibleBreak(q.penalty0, s.breakOnZero[0]) || (len(s.breakers) > 1 && isPossibleBreak(q.penalty1, s.breakOnZero[1])) {
			breakopp = i
			//tracing.Debugf("segmenter: penalties[%#U] = %d|%d   --- 8< ---", j, p0, p1)
			break
		}
		//tracing.Debugf("segmenter: penalties[%#U] = %d|%d", j, p0, p1)
	}
	if breakopp >= 0 {
		//tracing.Debugf("segmenter: break opportunity at %d", breakopp)
	} else if i == boundDist+1 {
		breakopp = int(boundDist)
		//tracing.Debugf("segmenter: break at bound position %d", breakopp)
	} else {
		//tracing.Debugf("segmenter: no break opportunity")
	}
	return breakopp
}

// find out if the UnicodeBreaker b is the primary breaker. If yes, return index 0,
// 1 otherwise.
func (s *Segmenter) inxForBreaker(b uax.UnicodeBreaker) int {
	if b == s.breakers[0] {
		return 0
	}
	return 1
}

// insertPenalties distributes penalties from the current breaker cycle
// between the Q atoms. Penalties may be located after the fact between
// runes of the current longest match.
func (s *Segmenter) insertPenalties(selector int, penalties []int) {
	l := s.deque.Len()
	if len(penalties) > l {
		penalties = penalties[:l] // drop excessive penalties
	}
	//var total [2]int
	//var r rune
	l -= 1 // now index of last
	for i, p := range penalties {
		at := l - i
		atom := s.deque.AtomAt(at)
		if selector == 0 {
			atom.penalty0 += p
		} else {
			atom.penalty1 += p
		}
		//r, total[0], total[1] = s.deque.At(l - i)
		//total[selector] = bounded(total[selector] + p)
		//total[selector] = total[selector] + p
		//total[selector] += p

		//s.deque.SetAt(l-i, r, total[0], total[1])
	}
}

// getFrontSegments cuts off a span of runes from the front of the Q and puts
// them into a byte-buffer.
// Normally the segment runs from the front of the Q to the position of the
// first feasible break.
//
// Things get more complicated by having a possible `bound` position, after
// which no runes are allowed to be cut off.
//
// After the cut-off getFrontSegments recursivley searches for another break
// opportunity which may have been marked by breakers with acceptable penalties.
//
func (s *Segmenter) getFrontSegment(bound int) (int, bool) {
	seglen := 0
	s.lastPenalties[0], s.lastPenalties[1] = 0, 0
	s.runesBuf = s.runesBuf.SetMark()
	l := min(s.deque.Len()-1, s.positionOfBreakOpportunity)
	if l < 0 {
		panic("l < 0")
		//return 0, true
	}
	var isCutOff bool
	start := s.pos - s.deque.Len()
	if start+l >= bound { // front segment would extend bound
		tracing.Debugf("segmenter: bound reached")
		isCutOff = true // we truncate l to Q-start---->|bound
		if l = int(bound) - int(s.pos) + s.deque.Len(); l < 0 {
			return 0, true
		}
	}
	for i := 0; i <= l; i++ {
		r, p0, p1 := s.deque.PopFront()
		written, _ := (&s.runesBuf).WriteRune(r)
		seglen += written
		s.lastPenalties[0], s.lastPenalties[1] = p0, p1
	}
	//tracing.Debugf("cutting front segment of length 0..%d: '%v'", l, s.runesBuf)
	// There may be further break opportunities between this break and the start of the
	// current longest match. Advance the pointer to the next break opportunity, if any.
	boundDist := bound
	if bound < math.MaxInt {
		boundDist = bound - s.pos + s.deque.Len()
	}
	s.positionOfBreakOpportunity = s.findBreakOpportunity(s.deque.Len()-s.longestActiveMatch,
		boundDist)
	//s.printQ()
	return seglen, isCutOff
}

// ----------------------------------------------------------------------

// Debugging helper. Print the content of the current queue to the debug log.
func (s *Segmenter) printQ() {
	var sb strings.Builder
	sb.WriteString("Q : ")
	for i := 0; i < s.deque.Len(); i++ {
		var a atom
		a.r, a.penalty0, a.penalty1 = s.deque.At(i)
		sb.WriteString(fmt.Sprintf(" <- %s", a.String()))
	}
	sb.WriteString(" .")
	tracing.P("UAX |Q|=", s.deque.Len()).Debugf(sb.String())
}

// --- Helpers ----------------------------------------------------------

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func bounded(p int) int {
	if p > uax.InfinitePenalty {
		p = uax.InfinitePenalty
	} else if p < uax.InfiniteMerits {
		p = uax.InfiniteMerits
	}
	return p
}

// runeread is a helper to wrap a `[]rune` into a cheap RuneReader.
type runeread struct {
	runes []rune
	pos   int
}

func (rr *runeread) ReadRune() (rune, int, error) {
	if rr.pos >= len(rr.runes) {
		return 0, 0, io.EOF
	}
	r := rr.runes[rr.pos]
	rr.pos++
	return r, 1, nil
}

// runewrite is a helper to wrap a `[]rune` into a cheap Buffer.
type runewrite struct {
	maxSegLen  int
	isBacked   bool
	backing    []rune // optional slice to back input
	start, end int    // limits of sub-slice
	output     []rune // output buffer
}

func (rw *runewrite) WriteRune(r rune) (int, error) {
	var err error
	if rw.isBacked {
		// we're not really writing anything, but rather just remembering the slice
		// limits of our (future) return segment
		// if r != rw.backing[rw.end] {
		// 	tracing.Debugf("rune writer and backing store out of sync: %#U", r)
		// 	err = fmt.Errorf("rune writer and backing store out of sync: %#U", r)
		// }
		rw.end++
	} else {
		// we're copying runes to an output buffer
		c := cap(rw.output)
		if c == len(rw.output) { // if buffer full
			if c >= rw.maxSegLen {
				return 0, ErrTooLong
			}
			// extend output buffer
			newSize := max(c+startBufSize, len(rw.output)+1)
			if newSize > rw.maxSegLen {
				newSize = rw.maxSegLen
			}
			old := rw.output[:]
			rw.output = make([]rune, len(old), newSize)
			copy(rw.output[:len(old)], old)
		}
		rw.output = append(rw.output, r)
	}
	return 1, err
}

func (rw runewrite) Runes() []rune {
	if rw.isBacked {
		return rw.backing[rw.start:rw.end]
	}
	return rw.output
}

func (rw runewrite) SetMark() runewrite {
	if rw.isBacked {
		rw.start = rw.end
	} else {
		rw.output = rw.output[:0]
	}
	return rw
}

func (rw runewrite) Reset(backingBuf []rune) runewrite {
	if cap(rw.output) > 0 {
		rw.output = rw.output[:0]
	}
	if backingBuf == nil {
		rw.isBacked = false
		if rw.output == nil {
			rw.output = make([]rune, 0, startBufSize)
		}
		rw.backing = nil
	} else {
		rw.isBacked = true
		rw.backing = backingBuf
	}
	rw.start, rw.end = 0, 0
	return rw
}

func (rw runewrite) String() string {
	if rw.isBacked {
		return string(rw.backing[rw.start:rw.end])
	}
	return string(rw.output)
}
