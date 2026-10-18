package parsing

import (
	"fmt"
	"strings"
)

// CREATE TABLE is not the only create statement btw...

type CreateStatement struct {
	tableName   string // can also make it a table type
	columns     []Column
	ifNotExists bool
}

func (c *CreateStatement) statementMarker() {}

func (c *CreateStatement) String() string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE")
	if c.ifNotExists {
		sb.WriteString(" IF NOT EXISTS")
	}
	sb.WriteString(c.tableName + " (\n")

	for _, col := range c.columns {
		constraintStr := ""
		for constraint := range col.contraints { // chill to use string += string in loop b/c usually # contraints <= 3
			constraintStr += fmt.Sprintf(" %s", constraint)
		}

		defaultStr := ""
		if col.hasDefault {
			defaultStr += fmt.Sprintf(" DEFAULT %s", col.defaultVal)
		}
		sb.WriteString(fmt.Sprintf("   %s %s%s%s,\n", col.name, col.sqlType, constraintStr, defaultStr))
	}
	sb.WriteString(");")
	return sb.String()
}
