package catalog

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func New() *Catalog {
	tables := make(map[string]*Table)
	return &Catalog{Tables: tables}
}

func (c *Catalog) Apply(statement parsing.Statement) error {
	if statement == nil {
		return fmt.Errorf("nil pointer was passed to catalog.Apply")
	}
	switch s := statement.(type) {
	case *parsing.CreateTableStatement:
		tableName := s.TableName
		columns := s.Columns
		if _, ok := c.Tables[tableName]; ok {
			if !s.IfNotExists {
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
	default:
		return fmt.Errorf("invalid statement type %T", s)
	}
	return nil
}

func (c *Catalog) ApplyAll(statements []parsing.Statement) error {
	var err error
	for _, s := range statements {
		err = c.Apply(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) Table(name string) (*Table, bool) {
	table, ok := c.Tables[name]
	if !ok {
		return nil, false
	}
	return table, true
}
