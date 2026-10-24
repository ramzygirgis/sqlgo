package parsing

import (
	"fmt"
	"strings"
)

func Lex(input string) ([]Token, error) {
	cleanedQuery, err := cleanQuery(input)
	if err != nil {
		return []Token{}, err
	}
	var toks []Token
	i := 0
	for i < len(cleanedQuery) {
		if tok, ok := matchPunct(cleanedQuery, i); ok {
			toks = append(toks, tok)
			i += len(tok.body)
			continue
		}
		c := cleanedQuery[i]
		switch {
		case c == ' ' || c == '\n' || c == '\t':
			i++
		case isValidChar(c):
			start := i
			for {
				i++
				if i == len(cleanedQuery) || !isValidChar(cleanedQuery[i]) {
					// TODO: ONLY numbers can have a period in them (for now). need to enforce. _ at front is only for identifiers (not sure if need this)
					body := cleanedQuery[start:i]
					if strings.HasPrefix(body, "$") {
						toks = append(toks, Token{kind: Param, body: body})
					} else if strings.HasPrefix(body, "'") && strings.HasSuffix(body, "'") {
						toks = append(toks, Token{kind: String, body: body})
					} else if isNumber(body) {
						toks = append(toks, Token{kind: Number, body: body})
					} else if isKeyword(body) {
						toks = append(toks, Token{kind: Keyword, body: strings.ToUpper(body)})
					} else if isType(body) {
						toks = append(toks, Token{kind: Type, body: strings.ToUpper(body)})
					} else {
						toks = append(toks, Token{kind: Identifier, body: body})
					}
					break
				}
			}
		default:
			return []Token{}, fmt.Errorf("unexpected character %s at position %d", string(c), i)
		}
	}
	toks = append(toks, Token{kind: EOF, body: "EOF"})
	return toks, nil
}
