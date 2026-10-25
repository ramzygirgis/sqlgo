package lexing

import (
	"fmt"
)

type TokenKind int

const (
	Illegal TokenKind = iota
	Identifier
	Keyword
	Comma
	Number
	String
	Param
	Type
	Comparator
	LParen
	RParen
	Star
	Semicolon
	EOF
)

func (tk TokenKind) String() string {
	strings := [...]string{"Illegal", "Identifier", "Keyword",
		"Comma", "Number", "String", "Param", "Type",
		"Comparator", "LParen", "RParen", "Star",
		"Semicolon", "EOF",
	}
	if tk < 0 || int(tk) >= len(strings) {
		return "TokenKind int out of range"
	}
	return strings[tk]
}

type Token struct {
	Kind TokenKind
	Body string
}

func (t Token) String() string {
	return fmt.Sprintf("Token{\n   kind: %s\n   body: %s\n}", t.Kind, t.Body)
}
