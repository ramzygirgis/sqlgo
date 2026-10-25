package parsing

import (
	"fmt"
)

type Statement interface {
	statementMarker()
	String() string // for testing
}

type Expr interface {
	expressionMarker()
	String() string // also for testing
}

type Column struct {
	name       string
	sqlType    string
	hasDefault bool
	defaultVal string
	nullable   bool
	primaryKey bool
}

type ColumnRef struct {
	name string
}

func (c *ColumnRef) expressionMarker() {}

func (c *ColumnRef) String() string {
	return c.name
}

type Table struct {
	name    string
	columns []Column
}

type Literal struct {
	body string
}

func (l *Literal) expressionMarker() {}

func (l *Literal) String() string {
	return l.body
}

type Parameter struct {
	index int
}

func (p *Parameter) expressionMarker() {}

func (p *Parameter) String() string {
	return fmt.Sprintf("$%d", p.index)
}
