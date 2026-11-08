package analyzer

import (
	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func AnalyzeQuery(catalog *catalog.Catalog, statement parsing.Statement) (*ir.Query, error) {
	switch s := statement.(type) {
	case *parsing.SelectStatement:
		kind := ir.Select
		targetTable := s.From.TableName
		parameters := ...
		sql := s.String()

		table, ok := catalog.Tables[targetTable]
		if !ok {
			return &ir.Query{}, fmt.Errorf("table %s does not appear in catalog", targetTable)
		}
		columns := table.Columns

		resultColumns := make([]ir.Column)
		for _, colName := range s.ColumnNames {
			catalogCol := table.
			// doing so also lets us get nullability
			col := ir.Column{
				Name: colName,
				SqlType: ...,
				Nullable: ...,
			}
			resultColumns = append(resultColumns, col)
		}
		return &ir.Query{
			Kind: kind,
			TargetTable: targetTable,
			Parameters: parameters, 
			ResultColumns: resultColumns,
			Sql: sql,
		}
	}
}



1. table exists
2. * expands to every col, in cat order. if * present, named are ignored.
3. each named col exists in table. ResultColumns follows order written.
4. where col refs must exist in table.
5. sql is query with * replaced



func AnalyzeQuery(catalog *catalog.Catalog, statement parsing.Statement) (*ir.Query, error) {
	switch s := statement.(type) {
	case *parsing.SelectStatement:
		kind := ir.Select
		targetTable := s.From.TableName
		parameters := ...
		sql := ...

		table, ok := catalog.Tables[targetTable]
		if !ok {
			return &ir.Query{}, fmt.Errorf("table %s does not appear in catalog", targetTable.Name)
		}

		// if s.all: do ...

		resultColumns := make([]ir.Column)

		for _, colName := range s.ColumnNames {
			if !catalog.TableHasColumn(table, colName) {
				return &ir.Query{}, fmt.Errorf("table %s does not contain column %s", targetTable, colName)
			}
			
		}
	}
}
