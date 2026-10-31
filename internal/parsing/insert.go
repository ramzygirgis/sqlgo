package parsing

import (
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strings"
)

type InsertIntoStatement struct {
	TableName string
	Columns   []ColumnRef
	Values    []string
}

func (c *InsertIntoStatement) statementMarker() {}

func (c *InsertIntoStatement) String() string {
	var sb strings.Builder

	sb.WriteString("INSERT INTO ")
	sb.WriteString(c.TableName)
	if len(c.Columns) != 0 {
		sb.WriteString(" (")
	}
	for i, ref := range c.Columns {
		sb.WriteString(ref.name)
		if i == len(c.Columns)-1 {
			sb.WriteString(")\n")
			break
		}
		sb.WriteString(", ")
	}
	sb.WriteString("VALUES (")
	for i, val := range c.Values {
		sb.WriteString(val)
		if i == len(c.Values)-1 {
			sb.WriteString(");")
			break
		}
		sb.WriteString(", ")
	}
	return sb.String()
}

func (p *parser) parseInsertInto() (*InsertIntoStatement, error) {
	if err := p.expectKeyword("INSERT"); err != nil {
		return &InsertIntoStatement{}, err
	}
	if err := p.expectKeyword("INTO"); err != nil {
		return &InsertIntoStatement{}, err
	}
	tableName, err := p.expectKind(lexing.Identifier)
	if err != nil {
		return &InsertIntoStatement{}, err
	}
	colRefs := make([]ColumnRef, 0)
	if p.read().Kind == lexing.LParen {
		p.next()
		colName, err := p.expectKind(lexing.Identifier)
		if err != nil {
			return &InsertIntoStatement{}, err
		}
		colRefs = append(colRefs, ColumnRef{name: colName})
		for p.acceptComma() {
			colName, err = p.expectKind(lexing.Identifier)
			if err != nil {
				return &InsertIntoStatement{}, err
			}
			colRefs = append(colRefs, ColumnRef{name: colName})
		}
		if _, err = p.expectKind(lexing.RParen); err != nil {
			return &InsertIntoStatement{}, err
		}
	}
	if err = p.expectKeyword("VALUES"); err != nil {
		return &InsertIntoStatement{}, err
	}
	if _, err = p.expectKind(lexing.LParen); err != nil {
		return &InsertIntoStatement{}, err
	}
	values := make([]string, 0)
	val, err := p.expectKind(lexing.Identifier)
	if err != nil {
		return &InsertIntoStatement{}, err
	}
	values = append(values, val)
	for p.acceptComma() {
		val, err := p.expectKind(lexing.Identifier)
		if err != nil {
			return &InsertIntoStatement{}, err
		}
		values = append(values, val)
	}
	if _, err := p.expectKind(lexing.RParen); err != nil {
		return &InsertIntoStatement{}, err
	}
	if _, err := p.expectKind(lexing.Semicolon); err != nil {
		return &InsertIntoStatement{}, err
	}
	return &InsertIntoStatement{
		TableName: tableName,
		Columns:   colRefs,
		Values:    values,
	}, nil
}
