package catalog

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func (c *Catalog) applyCreateTable(statement *parsing.CreateTableStatement) error {
	tableName := statement.TableName
	columns := statement.Columns
	if _, ok := c.Tables[tableName]; ok {
		if !statement.IfNotExists {
			return fmt.Errorf("table %q already exists", tableName)
		}
		return nil
	}
	var cols []Column
	colNames := make(map[string]struct{})
	for _, col := range columns {
		colName := col.Name
		sqlType := col.SqlType
		nullable := col.Nullable

		if col.PrimaryKey {
			nullable = false
		}

		if _, ok := colNames[colName]; ok {
			return fmt.Errorf("column %q already exists in table %q", colName, tableName)
		}
		colNames[colName] = struct{}{}

		col := Column{Name: colName, SqlType: sqlType, Nullable: nullable}
		cols = append(cols, col)
	}

	newTable := Table{Name: tableName, Columns: cols}
	c.Tables[tableName] = &newTable
	return nil
}
