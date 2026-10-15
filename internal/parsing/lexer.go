package parsing

import (
	"fmt"
)

// QUESTION:
// at what layer should we inject a "EOF" at the end of the text?
// maybe, we add this to queryString in pre-processing. need to figure this out
func lex(input string) ([]Token, error) {
	cleanedQuery, err := cleanQuery(input)
	if err != nil {
		return []Token{}, err
	}
	var toks []Token
	i := 0
	for i < len(cleanedQuery) {
		c := cleanedQuery[i]
		switch {
		case c == ' ':
			i++
		case c == ',':
			toks = append(toks, Token{kind: Comma, body: ","})
			i++
		case c == ';':
			toks = append(toks, Token{kind: Semicolon, body: ";"})
			i++
		case isValidChar(c):
			start := i
			for {
				i++
				if i == len(cleanedQuery) || !isValidChar(cleanedQuery[i]) {
					body := cleanedQuery[start:i]
					if isKeyword(body) {
						toks = append(toks, Token{kind: Keyword, body: body})
					} else {
						toks = append(toks, Token{kind: Identifier, body: body})
					}
					break
				}
			}
		default:
			return toks, fmt.Errorf("unexpected character %s at position %d", c, i)
		}
	}
	return toks, nil
}
