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
		c := cleanedQuery[i]
		switch {
		case c == ' ' || c == '\n' || c == '\t':
			i++
		case c == '*':
			toks = append(toks, Token{kind: Star, body: "*"})
			i++
		case c == ',':
			toks = append(toks, Token{kind: Comma, body: ","})
			i++
		case c == ';':
			toks = append(toks, Token{kind: Semicolon, body: ";"})
			i++
		case c == '(':
			toks = append(toks, Token{kind: Identifier, body: "("})
			i++
		case c == ')':
			toks = append(toks, Token{kind: Identifier, body: ")"})
			i++
		case isValidChar(c):
			start := i
			for {
				i++
				if i == len(cleanedQuery) || !isValidChar(cleanedQuery[i]) {
					body := cleanedQuery[start:i]
					if isKeyword(body) {
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
