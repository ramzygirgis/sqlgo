package parsing

import (
	"fmt"
)

/* TOKEN TYPES */

type TokenKind int

const (
	Illegal TokenKind = iota
	Identifier
	Keyword
	Comma
	Star
	Semicolon
	EOF
)

func (tk TokenKind) String() string {
	strings := [...]string{"Illegal", "Identifier", "Keyword", "Comma", "Star", "Semicolon", "EOF"}
	if tk < 0 || int(tk) >= len(strings) {
		return "TokenKind int out of range"
	}
	return strings[tk]
}

type Token struct {
	kind TokenKind
	body string
}

func (t Token) String() string {
	return fmt.Sprintf("Token{\n   kind: %s\n   body: %s\n}", t.kind, t.body)
}

/* AST TYPES */

type Statement interface {
	statementMarker()
}

// columns (public)

type Column struct {
	name string
}

func (c Column) GetName() string {
	return c.name
}

// tables (public) (maybe refactor to just include name)

type Table struct {
	name string
}

// FROM keyword

type FromStatement struct {
	table Table
}

// SELECT keyword/statement

type SelectStatement struct {
	columns []Column
	from    *FromStatement
}

func (s *SelectStatement) statementMarker() {}
