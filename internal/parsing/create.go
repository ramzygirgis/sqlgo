package parsing

import "strings"

type CreateTableStatement struct {
	tableName   string
	columns     []Column
	ifNotExists bool
}

func (c *CreateTableStatement) statementMarker() {}

func (c *CreateTableStatement) String() string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE ")
	if c.ifNotExists {
		sb.WriteString("IF NOT EXISTS ")
	}
	sb.WriteString(c.tableName)
	sb.WriteString(" (\n")

	colStrs := make([]string, len(c.columns))
	for i, col := range c.columns {
		colStrs[i] = "   " + col.String()
	}
	sb.WriteString(strings.Join(colStrs, ",\n"))

	sb.WriteString("\n);")
	return sb.String()
}

func (col Column) String() string {
	parts := []string{col.name, col.sqlType}
	if !col.nullable {
		parts = append(parts, "NOT NULL")
	}
	if col.primaryKey {
		parts = append(parts, "PRIMARY KEY")
	}
	if col.hasDefault {
		parts = append(parts, "DEFAULT "+col.defaultVal)
	}
	return strings.Join(parts, " ")
}
