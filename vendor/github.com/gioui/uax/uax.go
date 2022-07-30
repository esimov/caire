package uax

import "unicode/utf8"

// PositionOfFirstLegalRune returns a legal Unicode code point
// start position and cut-off prefix, if any.
func PositionOfFirstLegalRune(s string) (int, []byte) {
	i, l, start := 0, len(s), -1
	for i < l {
		if utf8.RuneStart(s[i]) {
			r, _ := utf8.DecodeRuneInString(s[i:])
			if r != utf8.RuneError {
				start = i
			}
			break
		}
	}
	//CT().Debugf("start index = %d", start)
	return start, []byte(s[:i])
}
