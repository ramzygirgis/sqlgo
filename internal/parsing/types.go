package parsing

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
