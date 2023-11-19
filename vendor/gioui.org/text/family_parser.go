package text

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenKind uint8

const (
	tokenStr tokenKind = iota
	tokenComma
	tokenEOF
)

type token struct {
	kind  tokenKind
	value string
}

func (t token) String() string {
	switch t.kind {
	case tokenStr:
		return t.value
	case tokenComma:
		return ","
	case tokenEOF:
		return "EOF"
	default:
		return "unknown"
	}
}

type lexState func(*lexer) lexState

func lexText(l *lexer) lexState {
	for {
		switch r := l.next(); {
		case r == -1:
			l.ignore()
			l.emit(tokenEOF)
			return nil
		case unicode.IsSpace(r):
			continue
		case r == ',':
			l.ignore()
			l.emit(tokenComma)
		case r == '"':
			l.ignore()
			return lexDquote
		case r == '\'':
			l.ignore()
			return lexSquote
		default:
			return lexBareStr
		}
	}
}

func lexBareStr(l *lexer) lexState {
	defer l.emitProcessed(tokenStr, func(s string) (string, error) {
		return strings.TrimSpace(s), nil
	})
	for {
		if strings.HasPrefix(l.input[l.pos:], `,`) {
			return lexText
		}
		switch r := l.next(); {
		case r == -1:
			return lexText
		}
	}
}

func lexDquote(l *lexer) lexState {
	return lexQuote(l, `"`)
}

func lexSquote(l *lexer) lexState {
	return lexQuote(l, `'`)
}

func unescape(s string, quote rune) (string, error) {
	var b strings.Builder
	hitNonSpace := false
	var wb strings.Builder
	for i := 0; i < len(s); {
		r, sz := utf8.DecodeRuneInString(s[i:])
		i += sz
		if unicode.IsSpace(r) {
			if !hitNonSpace {
				continue
			}
			wb.WriteRune(r)
			continue
		}
		hitNonSpace = true
		// If we get here, we're not looking at whitespace.
		// Insert any buffered up whitespace characters from
		// the gap between words.
		b.WriteString(wb.String())
		wb.Reset()
		if r == '\\' {
			r, sz := utf8.DecodeRuneInString(s[i:])
			i += sz
			switch r {
			case '\\', quote:
				b.WriteRune(r)
			default:
				return "", fmt.Errorf("illegal escape sequence \\%c", r)
			}
		} else {
			b.WriteRune(r)
		}
	}
	return b.String(), nil
}

func lexQuote(l *lexer, mark string) lexState {
	escaping := false
	for {
		if isQuote := strings.HasPrefix(l.input[l.pos:], mark); isQuote && !escaping {
			err := l.emitProcessed(tokenStr, func(s string) (string, error) {
				return unescape(s, []rune(mark)[0])
			})
			if err != nil {
				l.err = err
				return nil
			}
			l.next()
			l.ignore()
			return lexText
		}
		escaped := escaping
		switch r := l.next(); {
		case r == -1:
			l.err = fmt.Errorf("unexpected EOF while parsing %s-quoted family", mark)
			return lexText
		case r == '\\':
			if !escaped {
				escaping = true
			}
		}
		if escaped {
			escaping = false
		}
	}
}

type lexer struct {
	input  string
	pos    int
	tokens []token
	err    error
}

func (l *lexer) ignore() {
	l.input = l.input[l.pos:]
	l.pos = 0
}

// next decodes the next rune in the input and returns it.
func (l *lexer) next() int32 {
	if l.pos >= len(l.input) {
		return -1
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	return r
}

// emit adds a token of the given kind.
func (l *lexer) emit(t tokenKind) {
	l.emitProcessed(t, func(s string) (string, error) { return s, nil })
}

// emitProcessed adds a token of the given kind, but transforms its value
// with the provided closure first.
func (l *lexer) emitProcessed(t tokenKind, f func(string) (string, error)) error {
	val, err := f(l.input[:l.pos])
	l.tokens = append(l.tokens, token{
		kind:  t,
		value: val,
	})
	l.ignore()
	return err
}

// run executes the lexer on the given input.
func (l *lexer) run(input string) ([]token, error) {
	l.input = input
	l.tokens = l.tokens[:0]
	l.pos = 0
	for state := lexText; state != nil; {
		state = state(l)
	}
	return l.tokens, l.err
}

// parser implements a simple recursive descent parser for font family fallback
// expressions.
type parser struct {
	faces  []string
	lexer  lexer
	tokens []token
}

// parse the provided rule and return the extracted font families. The returned families
// are valid only until the next call to parse. If parsing fails, an error describing the
// failure is returned instead.
func (p *parser) parse(rule string) ([]string, error) {
	var err error
	p.tokens, err = p.lexer.run(rule)
	if err != nil {
		return nil, err
	}
	p.faces = p.faces[:0]
	return p.faces, p.parseList()
}

// parse implements the production:
//
//	LIST ::= <FACE> <COMMA> <LIST> | <FACE>
func (p *parser) parseList() error {
	if len(p.tokens) < 0 {
		return fmt.Errorf("expected family name, got EOF")
	}
	if head := p.tokens[0]; head.kind != tokenStr {
		return fmt.Errorf("expected family name, got %s", head)
	} else {
		p.faces = append(p.faces, head.value)
		p.tokens = p.tokens[1:]
	}

	switch head := p.tokens[0]; head.kind {
	case tokenEOF:
		return nil
	case tokenComma:
		p.tokens = p.tokens[1:]
		return p.parseList()
	default:
		return fmt.Errorf("unexpected token %s", head)
	}
}
