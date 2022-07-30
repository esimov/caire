package uax14

import "github.com/gioui/uax"

/*
BSD License

Copyright (c) 2017–21, Norbert Pillmayer (norbert@pillmayer.com)

All rights reserved.
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:

1. Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright
notice, this list of conditions and the following disclaimer in the
documentation and/or other materials provided with the distribution.

3. Neither the name of this software nor the names of its contributors
may be used to endorse or promote products derived from this software
without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

// These are the UAX#14 line break rules, slightly adapted.
// See http://www.unicode.org/Public/UCD/latest/ucd/auxiliary/LineBreakTest.html
//
// Format: <rule-no>\t<LHS>\t<op>\t<RHS>
//
/*
0.3		!	eot
4.0	BK	!
5.01	CR	×	LF
5.02	CR	!
5.03	LF	!
5.04	NL	!
6.0		×	( BK | CR | LF | NL )
7.01		×	SP
7.02		×	ZW
8.0	ZW SP*	÷
8.1	ZWJ_O	×
9.0	[^ SP BK CR LF NL ZW]	×	CM
11.01		×	WJ
11.02	WJ	×
12.0	GL	×
12.1	[^ SP BA HY CM]	×	GL
12.2	[^ BA HY CM] CM+	×	GL
12.3	^ CM+	×	GL
13.01		×	EX
13.02	[^ NU CM]	×	(CL | CP | IS | SY)
13.03	[^ NU CM] CM+	×	(CL | CP | IS | SY)
13.04	^ CM+	×	(CL | CP | IS | SY)
14.0	OP SP*	×
15.0	QU SP*	×	OP
16.0	(CL | CP) SP*	×	NS
17.0	B2 SP*	×	B2
18.0	SP	÷
19.01		×	QU
19.02	QU	×
20.01		÷	CB
20.02	CB	÷
21.01		×	BA
21.02		×	HY
21.03		×	NS
21.04	BB	×
21.1	HL (HY | BA)	×
21.2	SY	×	HL
22.01	(AL | HL)	×	IN
22.02	EX	×	IN
22.03	(ID | EB | EM)	×	IN
22.04	IN	×	IN
22.05	NU	×	IN
23.02	(AL | HL)	×	NU
23.03	NU	×	(AL | HL)
23.12	PR	×	(ID | EB | EM)
23.13	(ID | EB | EM)	×	PO
24.02	(PR | PO)	×	(AL | HL)
24.03	(AL | HL)	×	(PR | PO)
25.01	(PR | PO)	×	( OP | HY )? NU
25.02	( OP | HY )	×	NU
25.03	NU	×	(NU | SY | IS)
25.04	NU (NU | SY | IS)*	×	(NU | SY | IS | CL | CP)
25.05	NU (NU | SY | IS)* (CL | CP)?	×	(PO | PR)
26.01	JL	×	JL | JV | H2 | H3
26.02	JV | H2	×	JV | JT
26.03	JT | H3	×	JT
27.01	JL | JV | JT | H2 | H3	×	IN
27.02	JL | JV | JT | H2 | H3	×	PO
27.03	PR	×	JL | JV | JT | H2 | H3
28.0	(AL | HL)	×	(AL | HL)
29.0	IS	×	(AL | HL)
30.01	(AL | HL | NU)	×	OP
30.02	CP	×	(AL | HL | NU)
30.11	^ (RI RI)* RI	×	RI
30.12	[^RI] (RI RI)* RI	×	RI
30.13	RI	÷	RI
30.2	EB	×	EM
*/

// ----------------------------------------------------------------------

// LB2: Never break at the start of text.
// No longer used.
/*
func rule_LB2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, PenaltyToSuppressBreak)
}
*/

// LB3 Always break at the end of text.
func rule_LB3(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, 0, PenaltyForMustBreak)
}

func rule_05_NewLine(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == BKClass || c == NLClass || c == LFClass {
		rec.MatchLen++
		return uax.DoAccept(rec, PenaltyForMustBreak)
	} else if c == CRClass {
		rec.MatchLen++
		return rule_05_CRLF
	}
	return uax.DoAbort(rec)
}

func rule_05_CRLF(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == LFClass {
		return uax.DoAccept(rec, PenaltyForMustBreak, PenaltyToSuppressBreak, PenaltyToSuppressBreak)
	}
	return uax.DoAccept(rec, 0, PenaltyForMustBreak, PenaltyToSuppressBreak)
}

func rule_06_HardBreak(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == BKClass || c == NLClass || c == LFClass || c == CRClass {
		//rec.MatchLen++
		return uax.DoAccept(rec, 0, PenaltyToSuppressBreak)
	}
	return uax.DoAbort(rec)
}

// LB7 Do not break before spaces or zero width space.
func rule_LB7(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, 0, PenaltyToSuppressBreak)
}

// LB8 Break before any character following a zero-width space, even if
//one or more spaces intervene.
//
// ZW SP* +
func rule_LB8(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB8
}

// LB8 ... SP* +
func finish_LB8(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == SPClass {
		return finish_LB8
	}
	return uax.DoAccept(rec, 0, -p(8))
}

// LB8a Do not break after a zero width joiner.
//
// ZWJ x
func rule_LB8a(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, p(8))
}

// LB11 Do not break before or after Word joiner and related characters.
//
// x WJ x
func rule_LB11(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, p(11), p(11))
}

// LB12 Do not break after NBSP and related characters.
//
// GL x
/*
func rule_LB12(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, p(12))
}
*/

// LB12 Do not break after NBSP and related characters.
//
// LB12a: Do not break before NBSP and related characters, except after
// spaces and hyphens.
//
// [^SP BA HY] x GL
func rule_LB12(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	//c := UAX14Class(cpClass)
	uax14 := rec.UserData.(*LineWrap)
	last := uax14.lastClass
	if last == SPClass || last == BAClass || last == HYClass {
		return uax.DoAccept(rec, p(12))
	}
	return uax.DoAccept(rec, p(12), p(12))
}

// LB13: Do not break before ‘]’ or ‘!’ or ‘;’ or ‘/’, even after spaces.
//
// x (CL | CP | EX | IS | SY)
func rule_LB13(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, 0, p(13))
}

// LB14: Do not break after ‘[’, even after spaces.
//
// OP SP* x
func rule_LB14(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB14
}

// ... SP* x
func finish_LB14(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == SPClass {
		return finish_LB14
	}
	return uax.DoAccept(rec, 0, p(14))
}

// LB15: Do not break within ‘”[’, even with intervening spaces.
//
// QU SP* x OP
func rule_LB15(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB15
}

// ... SP* x OP
func finish_LB15(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == SPClass {
		return finish_LB15
	}
	if c == OPClass {
		return uax.DoAccept(rec, 0, p(15))
	}
	return uax.DoAbort(rec)
}

// LB16 Do not break between closing punctuation and a nonstarter (lb=NS),
// even with intervening spaces.
//
// (CL | CP) SP* x NS
func rule_LB16(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB16
}

// ... SP* x NS
func finish_LB16(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == SPClass {
		return finish_LB16
	}
	if c == NSClass {
		return uax.DoAccept(rec, 0, p(16))
	}
	return uax.DoAbort(rec)
}

// LB17 Do not break within ‘——’, even with intervening spaces.
//
// B2 SP* x B2
func rule_LB17(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB17
}

// ... SP* x B2
func finish_LB17(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == SPClass {
		return finish_LB17
	}
	if c == B2Class {
		return uax.DoAccept(rec, 0, p(17))
	}
	return uax.DoAbort(rec)
}

// LB18 Break after spaces.
//
// SP +
func rule_LB18(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, -p(18))
}

// LB19 Do not break before or after quotation marks.
//
// x QU x
func rule_LB19(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, p(19), p(19))
}

// LB20 Break before and after unresolved CB.
//
// + CB +
func rule_LB20(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	uax14 := rec.UserData.(*LineWrap)
	if uax14.lastClass == CBClass && uax14.substituted {
		return uax.DoAccept(rec, -p(20), p(20))
	}
	return uax.DoAccept(rec, -p(20), -p(20))
}

// LB21 Do not break before hyphen-minus, other hyphens, fixed-width spaces,
// small kana, and other non-starters, or after acute accents.
//
// x (BA | HY | NS)
func rule_LB21(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, 0, p(21))
}

// BB x
func rule_LB21x(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	return uax.DoAccept(rec, p(21))
}

// LB21a Don't break after Hebrew + Hyphen.
//
// HL (HY | BA) x
func rule_LB21a(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB21a
}

// .. (HY | BA) x
func finish_LB21a(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == HYClass || c == BAClass {
		return uax.DoAccept(rec, p(21))
	}
	return uax.DoAbort(rec)
}

//  LB21b Don’t break between Solidus and Hebrew letters.
//
// SY x HL
func rule_LB21b(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB21b
}

// ... x HL
func finish_LB21b(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == HLClass {
		return uax.DoAccept(rec, 0, p(21))
	}
	return uax.DoAbort(rec)
}

// LB22 Do not break between two ellipses, or between letters, numbers or
// exclamations and ellipsis.
//
// (AL | HL) x IN
// EX x IN
// (ID | EB | EM) x IN
// IN x IN
// NU x IN
func rule_LB22(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB22
}

// ... x IN
func finish_LB22(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == INClass {
		return uax.DoAccept(rec, 0, p(22))
	}
	return uax.DoAbort(rec)
}

// LB23 Do not break between digits and letters.
//
// (AL | HL) x NU
func rule_LB23_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB23_1
}

//  NU x (AL | HL)
func rule_LB23_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB23_2
}

// ... x NU
func finish_LB23_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == NUClass {
		return uax.DoAccept(rec, 0, p(23))
	}
	return uax.DoAbort(rec)
}

//  ... x (AL | HL)
func finish_LB23_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == ALClass || c == HLClass {
		return uax.DoAccept(rec, 0, p(23))
	}
	return uax.DoAbort(rec)
}

// LB23a Do not break between numeric prefixes and ideographs, or between
// ideographs and numeric postfixes.
//
// PR x (ID | EB | EM)
func rule_LB23a_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB23a_1
}

// (ID | EB | EM) x PO
func rule_LB23a_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB23a_2
}

// ... x (ID | EB | EM)
func finish_LB23a_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == IDClass || c == EBClass || c == EMClass {
		return uax.DoAccept(rec, 0, p(23))
	}
	return uax.DoAbort(rec)
}

// ... x PO
func finish_LB23a_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == POClass {
		return uax.DoAccept(rec, 0, p(23))
	}
	return uax.DoAbort(rec)
}

// LB24 Do not break between numeric prefix/postfix and letters, or between
// letters and prefix/postfix.
//
// (PR | PO) x (AL | HL)
func rule_LB24_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB24a_1
}

// (AL | HL) x (PR | PO)
func rule_LB24_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB24a_2
}

// ... x (AL | HL)
func finish_LB24a_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == ALClass || c == HLClass {
		return uax.DoAccept(rec, 0, p(23))
	}
	return uax.DoAbort(rec)
}

// ... x (PR | PO)
func finish_LB24a_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == PRClass || c == POClass {
		return uax.DoAccept(rec, 0, p(23))
	}
	return uax.DoAbort(rec)
}

// LB25 Do not break between the following regular expression relevant to
// numbers:
//
// ( PR | PO) ? ( OP | HY ) ? NU (NU | SY | IS) * (CL | CP) ? ( PR | PO) ?
//
// Examples: Examples: $(12.35)    2,1234    (12)¢    12.54¢.
// Example pairs: ‘$9’, ‘$[’, ‘$-’, ‘-9’, ‘/9’, ‘99’, ‘,9’, ‘9%’, ‘]%’.
func rule_LB25(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return step2_LB25
}

func step2_LB25(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == OPClass || c == HYClass {
		rec.MatchLen++
		return step3_LB25
	} else if c == NUClass {
		rec.MatchLen++
		return step4_LB25
	}
	return uax.DoAbort(rec)
}

func step3_LB25(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == NUClass {
		rec.MatchLen++
		return step4_LB25
	}
	return uax.DoAbort(rec)
}

func step4_LB25(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == NUClass || c == SYClass || c == ISClass {
		rec.MatchLen++
		return step4_LB25
	} else if c == CLClass || c == CPClass {
		rec.MatchLen++
		return step5_LB25
	} else if c == PRClass || c == POClass {
		return uax.DoAccept(rec, ps(25, p(25), rec.MatchLen)...)
	}
	return uax.DoAccept(rec, ps(25, 0, rec.MatchLen)...)
}

func step5_LB25(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == PRClass || c == POClass {
		return uax.DoAccept(rec, ps(25, p(25), rec.MatchLen)...)
	}
	return uax.DoAccept(rec, ps(25, 0, rec.MatchLen)...)
}

func finish_LB25(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	panic("rule 25 not implemented")
}

// LB26 Do not break a Korean syllable.
//
// JL x (JL | JV | H2 | H3)
func rule_LB26_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB26_1
}

// (JV | H2) x (JV | JT)
func rule_LB26_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB26_2
}

// (JT | H3) x JT
func rule_LB26_3(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB26_3
}

// ... x (JL | JV | H2 | H3)
func finish_LB26_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == JLClass || c == JVClass || c == H2Class || c == H3Class {
		return uax.DoAccept(rec, 0, p(26))
	}
	return uax.DoAbort(rec)
}

// ... x (JV | JT)
func finish_LB26_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == JVClass || c == JTClass {
		return uax.DoAccept(rec, 0, p(26))
	}
	return uax.DoAbort(rec)
}

// ... x JT
func finish_LB26_3(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == JTClass {
		return uax.DoAccept(rec, 0, p(26))
	}
	return uax.DoAbort(rec)
}

// LB27 Treat a Korean Syllable Block the same as ID.
//
// (JL | JV | JT | H2 | H3) x IN
// (JL | JV | JT | H2 | H3) x PO
//
// When Korean uses SPACE for line breaking, the classes in rule LB26,
// as well as characters of class ID, are often tailored to AL;
// see Section 8, Customization.
func rule_LB27(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB27_1
}

// PR x (JL | JV | JT | H2 | H3)
func rule_LB27_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB27_2
}

// ... x (IN | PO)
func finish_LB27_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == INClass || c == POClass {
		return uax.DoAccept(rec, 0, p(27))
	}
	return uax.DoAbort(rec)
}

// ... x (JL | JV | JT | H2 | H3)
func finish_LB27_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == JLClass || c == JVClass || c == JTClass || c == H2Class || c == H3Class {
		return uax.DoAccept(rec, 0, p(27))
	}
	return uax.DoAbort(rec)
}

// LB28 Do not break between alphabetics.
//
// (AL | HL) × (AL | HL)
func rule_LB28(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB28
}

// ... × (AL | HL)
func finish_LB28(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == ALClass || c == HLClass {
		return uax.DoAccept(rec, 0, p(28))
	}
	return uax.DoAbort(rec)
}

// LB29 Do not break between numeric punctuation and alphabetics (“e.g.”).
//
// IS x (AL | HL)
func rule_LB29(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB29
}

// ... × (AL | HL)
func finish_LB29(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == ALClass || c == HLClass {
		return uax.DoAccept(rec, 0, p(29))
	}
	return uax.DoAbort(rec)
}

// LB30 Do not break between letters, numbers, or ordinary symbols and
// opening or closing parentheses.
//
// The purpose of this rule is to prevent breaks in common cases where a
// part of a word appears between delimiters—for example, in “person(s)”.
//
// (AL | HL | NU) x OP
func rule_LB30_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB30_1
}

// CP x (AL | HL | NU)
func rule_LB30_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB30_2
}

// ... x OP
func finish_LB30_1(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == OPClass {
		return uax.DoAccept(rec, 0, p(30))
	}
	return uax.DoAbort(rec)
}

// ... x (AL | HL | NU)
func finish_LB30_2(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	if c == ALClass || c == HLClass || c == NUClass {
		return uax.DoAccept(rec, 0, p(30))
	}
	return uax.DoAbort(rec)
}

// LB30a Break between two regional indicator symbols if and only if there
// are an even number of regional indicators preceding the position of the break.
//
// (RI RI)* RI x RI
func rule_LB30a(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	lw := rec.UserData.(*LineWrap)
	lw.block()
	rec.MatchLen++
	return finish_LB30a
}

// ... x RI
//
// TODO: This will clash with rule LB 9 and 10 !
func finish_LB30a(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)
	lw := rec.UserData.(*LineWrap)
	if c == RIClass {
		if lw.substituted {
			return finish_LB30a
		}
		lw.unblock()
		return uax.DoAccept(rec, 0, p(30))
	}
	lw.unblock()
	return uax.DoAbort(rec)
}

// LB30b Do not break between an emoji base and an emoji modifier.
//
// EB x EM
func rule_LB30b(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	rec.MatchLen++
	return finish_LB30b
}

// ... x EM
func finish_LB30b(rec *uax.Recognizer, r rune, cpClass int) uax.NfaStateFn {
	c := UAX14Class(cpClass)

	if c == EMClass {
		return uax.DoAccept(rec, 0, p(30))
	}
	return uax.DoAbort(rec)
}
