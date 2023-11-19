package harfbuzz

import (
	"encoding/hex"
	"strings"

	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

// ported from harfbuzz/src/hb-ot-tag.cc Copyright © 2009  Red Hat, Inc. 2011  Google, Inc. Behdad Esfahbod, Roozbeh Pournader

var (
	// OpenType script tag, `DFLT`, for features that are not script-specific.
	tagDefaultScript = loader.NewTag('D', 'F', 'L', 'T')
	// OpenType language tag, `dflt`. Not a valid language tag, but some fonts
	// mistakenly use it.
	tagDefaultLanguage = loader.NewTag('d', 'f', 'l', 't')
)

func oldTagFromScript(script language.Script) tables.Tag {
	/* This seems to be accurate as of end of 2012. */

	switch script {
	case 0:
		return tagDefaultScript
	case language.Mathematical_notation:
		return loader.NewTag('m', 'a', 't', 'h')

	/* KATAKANA and HIRAGANA both map to 'kana' */
	case language.Hiragana:
		return loader.NewTag('k', 'a', 'n', 'a')

	/* Spaces at the end are preserved, unlike ISO 15924 */
	case language.Lao:
		return loader.NewTag('l', 'a', 'o', ' ')
	case language.Yi:
		return loader.NewTag('y', 'i', ' ', ' ')
	/* Unicode-5.0 additions */
	case language.Nko:
		return loader.NewTag('n', 'k', 'o', ' ')
	/* Unicode-5.1 additions */
	case language.Vai:
		return loader.NewTag('v', 'a', 'i', ' ')
	}

	/* Else, just change first char to lowercase and return */
	return tables.Tag(script | 0x20000000)
}

func newTagFromScript(script language.Script) tables.Tag {
	switch script {
	case language.Bengali:
		return loader.NewTag('b', 'n', 'g', '2')
	case language.Devanagari:
		return loader.NewTag('d', 'e', 'v', '2')
	case language.Gujarati:
		return loader.NewTag('g', 'j', 'r', '2')
	case language.Gurmukhi:
		return loader.NewTag('g', 'u', 'r', '2')
	case language.Kannada:
		return loader.NewTag('k', 'n', 'd', '2')
	case language.Malayalam:
		return loader.NewTag('m', 'l', 'm', '2')
	case language.Oriya:
		return loader.NewTag('o', 'r', 'y', '2')
	case language.Tamil:
		return loader.NewTag('t', 'm', 'l', '2')
	case language.Telugu:
		return loader.NewTag('t', 'e', 'l', '2')
	case language.Myanmar:
		return loader.NewTag('m', 'y', 'm', '2')
	}

	return tagDefaultScript
}

// Complete list at:
// https://docs.microsoft.com/en-us/typography/opentype/spec/scripttags
//
// Most of the script tags are the same as the ISO 15924 tag but lowercased.
// So we just do that, and handle the exceptional cases in a switch.
func allTagsFromScript(script language.Script) []tables.Tag {
	var tags []tables.Tag

	tag := newTagFromScript(script)
	if tag != tagDefaultScript {
		// HB_SCRIPT_MYANMAR maps to 'mym2', but there is no 'mym3'.
		if tag != loader.NewTag('m', 'y', 'm', '2') {
			tags = append(tags, tag|'3')
		}
		tags = append(tags, tag)
	}

	oldTag := oldTagFromScript(script)
	if oldTag != tagDefaultScript {
		tags = append(tags, oldTag)
	}
	return tags
}

func otTagsFromLanguage(langStr string) []tables.Tag {
	// check for matches of multiple subtags.
	if tags := tagsFromComplexLanguage(langStr); len(tags) != 0 {
		return tags
	}

	// find a language matching in the first component.
	s := strings.IndexByte(langStr, '-')
	if s != -1 && len(langStr) >= 6 {
		extlangEnd := strings.IndexByte(langStr[s+1:], '-')
		// if there is an extended language tag, use it.
		ref := extlangEnd
		if extlangEnd == -1 {
			ref = len(langStr[s+1:])
		}
		if ref == 3 && isAlpha(langStr[s+1]) {
			langStr = langStr[s+1:]
		}
	}

	if tagIdx := bfindLanguage(langStr); tagIdx != -1 {
		for tagIdx != 0 && otLanguages[tagIdx].language == otLanguages[tagIdx-1].language {
			tagIdx--
		}
		var out []tables.Tag
		for i := 0; tagIdx+i < len(otLanguages) &&
			otLanguages[tagIdx+i].tag != 0 &&
			otLanguages[tagIdx+i].language == otLanguages[tagIdx].language; i++ {
			out = append(out, otLanguages[tagIdx+i].tag)
		}
		return out
	}

	if s == -1 {
		s = len(langStr)
	}
	if s == 3 {
		// assume it's ISO-639-3 and upper-case and use it.
		return []tables.Tag{loader.NewTag(langStr[0], langStr[1], langStr[2], ' ') & ^tables.Tag(0x20202000)}
	}

	return nil
}

// return 0 if no tag
func parsePrivateUseSubtag(privateUseSubtag string, prefix string, normalize func(byte) byte) (tables.Tag, bool) {
	s := strings.Index(privateUseSubtag, prefix)
	if s == -1 {
		return 0, false
	}

	var tag [4]byte
	L := len(privateUseSubtag)
	s += len(prefix)
	if s < L && privateUseSubtag[s] == '-' {
		s += 1
		if L < s+8 {
			return 0, false
		}
		_, err := hex.Decode(tag[:], []byte(privateUseSubtag[s:s+8]))
		if err != nil {
			return 0, false
		}
	} else {
		var i int
		for ; i < 4 && s+i < L && isAlnum(privateUseSubtag[s+i]); i++ {
			tag[i] = normalize(privateUseSubtag[s+i])
		}
		if i == 0 {
			return 0, false
		}

		for ; i < 4; i++ {
			tag[i] = ' '
		}
	}
	out := loader.NewTag(tag[0], tag[1], tag[2], tag[3])
	if (out & 0xDFDFDFDF) == tagDefaultScript {
		out ^= ^tables.Tag(0xDFDFDFDF)
	}
	return out, true
}

// newOTTagsFromScriptAndLanguage converts a `Script` and a `Language`
// to script and language tags.
func newOTTagsFromScriptAndLanguage(script language.Script, language language.Language) (scriptTags, languageTags []tables.Tag) {
	if language != "" {
		prefix, privateUseSubtag := language.SplitExtensionTags()

		s, hasScript := parsePrivateUseSubtag(string(privateUseSubtag), "-hbsc", toLower)
		if hasScript {
			scriptTags = []tables.Tag{s}
		}

		l, hasLanguage := parsePrivateUseSubtag(string(privateUseSubtag), "-hbot", toUpper)
		if hasLanguage {
			languageTags = append(languageTags, l)
		} else {
			if prefix == "" { // if the language is 'fully private'
				prefix = language
			}
			languageTags = otTagsFromLanguage(string(prefix)) // TODO:
		}
	}

	if len(scriptTags) == 0 {
		scriptTags = allTagsFromScript(script)
	}
	return
}
