package lexing

import (
	"strings"
)

var punctuation = []struct {
	text string
	kind TokenKind
}{
	{"<=", Comparator}, {">=", Comparator}, {"<>", Comparator}, {"!=", Comparator},
	{"=", Comparator}, {"<", Comparator}, {">", Comparator}, {",", Comma},
	{";", Semicolon}, {"(", LParen}, {")", RParen}, {"*", Star},
}

func matchPunct(input string, i int) (Token, bool) {
	for _, p := range punctuation {
		if strings.HasPrefix(input[i:], p.text) {
			return Token{Kind: p.kind, Body: p.text}, true
		}
	}
	return Token{}, false
}
