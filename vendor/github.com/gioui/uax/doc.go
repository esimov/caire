/*
Package uax is about Unicode Annexes and their algorithms.

Description

From the Unicode Consortium:

A Unicode Standard Annex (UAX) forms an integral part of the Unicode
Standard, but is published online as  a  separate  document.
The Unicode Standard may require conformance to normative content
in a Unicode Standard Annex, if so specified in  the  Conformance
chapter of that version of the Unicode Standard. The version
number of a UAX document corresponds to the version of  the  Unicode
Standard of which it forms a part.

[...]

A string of Unicode‐encoded text often needs to be broken up into
text elements programmatically. Common examples of text  elements
include  what  users  think  of as characters, words, lines (more
precisely, where line breaks are  allowed),  and  sentences.  The
precise  determination of text elements may vary according to orthographic
conventions for a given script or language.  The  goal
of matching user perceptions cannot always be met exactly because
the text alone does not always contain enough information to
unambiguously  decide  boundaries.  For example, the period (U+002E
FULL STOP) is used  ambiguously,  sometimes  for  end‐of‐sentence
purposes, sometimes for abbreviations, and sometimes for numbers.
In most cases, however, programmatic text  boundaries  can  match
user  perceptions quite closely, although sometimes the best that
can be done is not to surprise the user.

[...]

There are many different ways to divide text elements corresponding
to user‐perceived characters, words, and sentences,  and  the
Unicode  Standard does not restrict the ways in which implementations
can produce these divisions.

This specification defines default mechanisms; more sophisticated
implementations can and should tailor them for particular locales
or  environments.  For example, reliable detection of word boundaries
in languages such as Thai, Lao, Chinese,  or  Japanese
requires the use of dictionary lookup, analogous to English hyphenation.

Package Contents

Implementations of specific UAX algorithms is done in the various
sub-packages of uax. The driver type for some of the breaking-algorithms
sits in sub-package segment and will
use breaker-algorithms from other sub-packages.

Base package uax provides some of the necessary
means to implement UAX breaking algorithms. Please note that it is
in now way mandatory to use the supporting types and functions of this
package. Implementors of additional breaking algorithms are free to
ignore some or all of the helpers and instead implement their breaking
algorithms from scratch.

Every implementation of UAX breaking algorithms has to handle the trade-off
between efficiency and understandability. Algorithms as described in the
Unicodes Annex documents are no easy read when considering all the details
and edge cases. Getting it 100% right therefore sometimes may be tricky.
Implementations in the sub-packages of uax try to strike a balance between
efficiency and readability. The helper classes of uax allow implementors to
transform UAX-rules into fairly readable small functions. From a maintenance
point-of-view this is preferrable to huge and complex cascades of if-statements,
which may provide better performance, but
are hard to understand. Most of the breaking algorithms within sub-packages of uax
therefore utilize the helper types from package uax.

We perform segmenting Unicode text based on rules, which are short
regular expressions, i.e. finite state automata. This corresponds well with
the formal UAX description of rules (except for the Bidi-rules, which are
better understood as rules for a context-sensitive grammar). Every step within a
rule is performed by executing a function. This function recognizes a single
code-point class and returns another function. The returned function
represents the expectation for the next code-point(-class).
These kind of matching by function is continued until a rule is accepted
or aborted.

An example let's consider rule WB13b “Do not break from extenders” from UAX#29:

   ExtendNumLet x (ALetter | Hebrew_Letter| Numeric | Katakana)

The 'x' denotes a suppressed break. All the identifiers are UAX#29-specific
classes for code-points. Matching them will call two functions in sequence:

      rule_WB13b( … )   // match ExtendNumLet
   -> finish_WB13b( … ) // match any of ALetter … Katakana

The final return value will either signal an accept or abort.

The uax helper to perform this kind of matching is called Recognizer.
A set of Recognizers comprises an NFA and will match break opportunities
for a given UAX rule-set. Recognizers receive rune events and therefore implement
interface RuneSubscriber.

Rune Events

Walking the runes (= code-points) of a Unicode text and firing rules to match
segments will produce a high fluctuation of short-lived Recognizers.
Every Recognizer will have to react to the next rune read. Package uax
provides a publish-subscribe mechanism for signalling new runes to all active
Recognizers.

The default rune-publisher will distribute rune events to rune-subscribers
and collect return values. Subscribers are required to return active matches
and possible break-opportunities (or suppression thereof).
After all subscribers are done consuming the rune, the publisher harvests
subscribers which have ended their life-cycle (i.e., either accepted or
aborted). Dead subscribers are flagging this with Done()==true and get
unsubscribed.

Breaking algorithms are performed by `UnicodeBreaker`s (an interface type).
The UnicodeBreakers in sub-packages of this package utilize UnicodePublishers
as described above.
The segment-driver needs one or more UnicodeBreakers to perform breaking logic.

Penalties

Algorithms in this package will signal break opportunities for Unicode text.
However, breaks are not signalled with true/false, but rather with a
weighted “penalty.” Every break is connoted with an integer value,
representing the desirability of the break. Negative values denote a
negative penalty, i.e., a merit. High enough penalties signal the complete
suppression of a break opportunity, causing the segmenter to not report
this break.

The UnicodeBreakers in this package (including sub-packages)
will apply the following logic:

(1) Mandatory breaks will have a penalty/merit of ≤ -10000 (uax.InfinitePenalty).

(2) Inhibited breaks will have penalty ≥ 10000 (uax.InfiniteMerits).

(3) Neutral positions will have a penalty of 0. The segmenter can be configured
to regard the zero value as breakable or not.

The segmenter will aggregate penalties from its breakers and output aggregated
penalties to the client.

______________________________________________________________________

License

This project is provided under the terms of the UNLICENSE or
the 3-Clause BSD license denoted by the following SPDX identifier:

SPDX-License-Identifier: 'Unlicense' OR 'BSD-3-Clause'

You may use the project under the terms of either license.

Licenses are reproduced in the license file in the root folder of this module.

Copyright © 2021 Norbert Pillmayer <norbert@pillmayer.com>

*/
package uax

// We define constants for flagging break points as infinitely bad and
// infinitely good, respectively.
const (
	InfinitePenalty = 10000
	InfiniteMerits  = -10000
)
