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
	String() string // for testing
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

type FromKeyword struct {
	table Table
}

func (s *FromKeyword) String() string {
	return fmt.Sprintf(" FROM %s", s.table.name)
}

// SELECT keyword/statement

type SelectStatement struct {
	columns []Column
	from    *FromKeyword
	all     bool
}

func (s *SelectStatement) statementMarker() {}

func (s *SelectStatement) String() string {
	res := "SELECT"
	if s.all {
		res += " *,"
	}
	for _, col := range s.columns {
		res += fmt.Sprintf(" %s,", col.name)
	}
	res = res[:len(res)-1]
	res += s.from.String()
	res += ";"
	return res
}
