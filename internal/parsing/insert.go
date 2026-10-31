package parsing

import (
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
