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
	Semicolon
	EOF
)

type Token struct {
	kind TokenKind
	body string
}

/* AST TYPES */

type Statement interface {
	stmtNode()
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

// FROM statement

type FromStatement struct {
	table Table
}

// SELECT statement

type SelectStatement struct {
	columns []Column
	from    *FromStatement
}

func (s *SelectStatement) statementNode() {}

// note on design: error checking should happen at the first level where this error occurs
// one couldn't detect whether we are selecting invalid columns from a table without
// the SelectStatement struct. this is therefore the lowest level at which this error can occur.
func (s SelectStatement) ValidateSELECT() (bool, error) {
	for _, c := range s.columns {
		if !s.from.table.ContainsColumn(c) {
			return false, fmt.Errorf("column %s does not appear in table %s.", c.name, s.from.table.name)
		}
	}
	return true, nil
}
