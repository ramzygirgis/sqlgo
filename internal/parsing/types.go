package parsing

import (
	"fmt"
)

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
