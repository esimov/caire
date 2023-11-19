// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import "github.com/go-text/typesetting/language"

// Ltag is the language tags table
// See https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6ltag.html
type Ltag struct {
	version    uint32        //	Table version; currently 1
	flags      uint32        //	Table flags; currently none defined
	numTags    uint32        //	Number of language tags which follow
	tagRange   []stringRange `arrayCount:"ComputedField-numTags"` //	Range for each tag's string
	stringData []byte        `subsliceStart:"AtStart" arrayCount:"ToEnd"`
}

type stringRange struct {
	offset uint16 // Offset from the start of the table to the beginning of the string
	length uint16 // String length (in bytes)
}

func (lt Ltag) Language(i uint16) language.Language {
	r := lt.tagRange[i]
	return language.NewLanguage(string(lt.stringData[r.offset : r.offset+r.length]))
}
