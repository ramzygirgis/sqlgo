package parsing

import (
	"fmt"
	"slices"
)

/* PRASER TYPES */

type parser struct {
	toks []Token
	pos  int
}

func (p *parser) peek() Token {
	if p.pos >= len(p.toks) || p.pos < 0 {
		return Token{
			kind: EOF,
			body: "",
		}
	}
	return p.toks[p.pos]
}

func (p *parser) next() Token {
	p.pos += 1
	return p.peek()
}

func (p *parser) expect(k TokenKind, body string) (Token, error) {
	if p.peek().kind != k {
		return Token{}, fmt.Errorf("unexpected token kind: expected %v, got %v", k, p.peek().kind)
	}
	if p.peek().body != body {
		return Token{}, fmt.Errorf("unexpected token body: expected %s, got %s", body, p.peek().body)
	}
	return Token{
		kind: k,
		body: body,
	}, nil
}

/* TOKEN TYPES */

type TokenKind int

const (
	Identifier TokenKind = iota
	Keyword
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

//TODO: implement this method for all of the below

// columns (public)

type Column struct {
	name string
}

func (c Column) GetName() string {
	return c.name
}

// tables (public)

type Table struct {
	name    string
	columns []Column
}

func (t Table) GetName() string {
	return t.name
}

func (t Table) GetColumns() []Column {
	return t.columns
}

func (t Table) ContainsColumn(c Column) bool {
	return slices.Contains(t.columns, c)
}

// FROM statement

type FromStatement struct {
	table Table
}

// SELECT statement

type SelectStatement struct {
	columns []Column
	from    FromStatement
}

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

// AST

type AST struct {
	selectStatement SelectStatement
}
