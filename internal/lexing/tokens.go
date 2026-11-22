package lexing

import (
	"fmt"
)

type TokenKind int

const (
	IllegalTokenKind TokenKind = iota
	OneAnnotation              // Body <==> name
	ManyAnnotation             // Body <==> name
	ExecAnnotation             // Body <==> name
	Identifier
	Keyword
	Comma
	Number
	String
	Boolean
	Param
	Type
	Comparator
	LParen
	RParen
	Star
	Semicolon
	EOF
	OutOfBounds
)

func (tk TokenKind) String() string {
	strings := [...]string{"IllegalTokenKind", "OneAnnotation",
		"ManyAnnotation", "ExecAnnotation", "Identifier", "Keyword",
		"Comma", "Number", "String", "Boolean", "Param", "Type",
		"Comparator", "LParen", "RParen", "Star",
		"Semicolon", "EOF", "OutOfBounds",
	}
	if tk < 0 || int(tk) >= len(strings) {
		return fmt.Sprintf("TokenKind(%d)", tk)
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
