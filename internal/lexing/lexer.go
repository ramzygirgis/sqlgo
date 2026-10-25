package lexing

import (
	"fmt"
	"strings"
	"unicode"
)

func Lex(input string) ([]Token, error) {
	var toks []Token
	i := 0
	for i < len(input) {
		if tok, ok := matchPunct(input, i); ok {
			toks = append(toks, tok)
			i += len(tok.Body)
			continue
		}
		b := input[i]
		switch {
		case unicode.IsSpace(rune(b)):
			i++
		case b == '\'':
			tok, end, err := lexStringLiteral(input, i)
			if err != nil {
				return nil, err
			}
			toks = append(toks, tok)
			i = end
		case isDigit(b):
			tok, end, err := lexNumber(input, i)
			if err != nil {
				return nil, err
			}
			toks = append(toks, tok)
			i = end
		case b == '$':
			tok, end, err := lexParam(input, i)
			if err != nil {
				return nil, err
			}
			toks = append(toks, tok)
			i = end
		case isWordStart(b):
			tok, end, err := lexWord(input, i)
			if err != nil {
				return nil, err
			}
			toks = append(toks, tok)
			i = end
		default:
			return []Token{}, fmt.Errorf("unexpected character %q at position %d", b, i)
		}
	}
	toks = append(toks, Token{Kind: EOF, Body: "EOF"})
	return toks, nil
}

func lexStringLiteral(input string, start int) (Token, int, error) {
	i := start + 1
	for i < len(input) {
		if input[i] == '\'' {
			if i+1 < len(input) && input[i+1] == '\'' {
				i += 2
				continue
			}
			return Token{Kind: String, Body: input[start : i+1]}, i + 1, nil
		}
		i++
	}
	return Token{}, 0, fmt.Errorf("unterminated string starting at position %d", start)
}

func lexNumber(input string, start int) (Token, int, error) {
	i := start
	seenDot := false
	for i < len(input) {
		b := input[i]
		if b == '.' {
			if seenDot {
				return Token{}, 0, fmt.Errorf("invalid number at position %d: multiple decimal points", start)
			}
			seenDot = true
		} else if !isDigit(b) {
			break
		}
		i++
	}
	if i < len(input) && isWordStart(input[i]) {
		return Token{}, 0, fmt.Errorf("invalid number at position %d: unexpected %q", start, input[i])
	}
	return Token{Kind: Number, Body: input[start:i]}, i, nil
}

func lexParam(input string, start int) (Token, int, error) {
	i := start + 1 // skip the $
	for i < len(input) && isDigit(input[i]) {
		i++
	}
	if i == start+1 {
		return Token{}, 0, fmt.Errorf("invalid parameter at position %d: expected digits after $", start)
	}
	if i < len(input) && isWordStart(input[i]) {
		return Token{}, 0, fmt.Errorf("invalid parameter at position %d: unexpected %q", start, input[i])
	}
	return Token{Kind: Param, Body: input[start:i]}, i, nil
}

func lexWord(input string, start int) (Token, int, error) {
	i := start
	if start >= len(input) {
		return Token{}, 0, fmt.Errorf("invalid `start` passed to lexWord")
	}
	for i < len(input) && isWordChar(input[i]) {
		i++
	}
	if i == start {
		return Token{}, 0, fmt.Errorf("first character '%s' is not a word character.", input[i])
	}
	body := input[start:i]
	switch {
	case isKeyword(body):
		return Token{Kind: Keyword, Body: strings.ToUpper(body)}, i, nil
	case isType(body):
		return Token{Kind: Type, Body: strings.ToUpper(body)}, i, nil
	default:
		return Token{Kind: Identifier, Body: body}, i, nil
	}
}
