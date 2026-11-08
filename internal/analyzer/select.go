package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

// handleSelect resolves a SELECT against the catalog and produces an ir.Query.
// The caller (Analyze's type switch) has already narrowed the statement to
// *parsing.SelectStatement, so no second type assertion is needed here.
func handleSelect(cat *catalog.Catalog, s *parsing.SelectStatement) (*ir.Query, error) {
	tableName := s.From.TableName
	table, ok := cat.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}

	resultColumns, err := collectColumns(cat, s)
	if err != nil {
		return &ir.Query{}, err
	}
	// TODO

	return &ir.Query{
		Kind:          ir.Select,
		TargetTable:   tableName,
		Parameters:    parameters,
		ResultColumns: resultColumns,
		Sql:           expanded.String(),
	}, nil
}

func collectColumns(cat *catalog.Catalog, s *parsing.SelectStatement) ([]ir.Column, error) {
	tableName := s.From.TableName
	table, ok := cat.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}

	var resultColumns []ir.Column
	var colNames []string // * should be expanded
	if s.All {
		if err := validateColRefs(s.ColumnNames, *table); err != nil {
			return []ir.Column{}, err
		}
		for _, col := range table.Columns {
			irCol, err := toIRColumn(tableName, col)
			if err != nil {
				return nil, err
			}
			resultColumns = append(resultColumns, irCol)
			colNames = append(colNames, col.Name)
			// we should validate that the query contains no invalid col refs
		}
	} else {
		for _, name := range s.ColumnNames {
			col, ok := table.Column(name)
			if !ok {
				return nil, fmt.Errorf("column %q does not exist in table %q", name, tableName)
			}
			irCol, err := toIRColumn(tableName, *col)
			if err != nil {
				return nil, err
			}
			resultColumns = append(resultColumns, irCol)
			colNames = append(colNames, name)
		}
	}
	return resultColumns, nil
}

func collectParams(cat *catalog.Catalog, s *parsing.SelectStatement) ([]ir.Parameter, error) {
	seen := map[int]ir.SQLType
	parameters := make([]ir.Parameter)

	// TODO
}
