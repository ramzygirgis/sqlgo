package analyzer

import (
	"fmt"
	"strings"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

// handleSelect resolves a SELECT against the catalog and produces an ir.Query.
// The caller (Analyze's type switch) has already narrowed the statement to
// *parsing.SelectStatement, so no second type assertion is needed here.
func handleSelect(cat *catalog.Catalog, s *parsing.SelectStatement) (*UnannotatedQuery, error) {
	tableName := s.TableName
	table, ok := cat.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}

	resultColumns, err := collectResultColumns(table, s)
	if err != nil {
		return nil, err
	}

	var parameters []ir.Parameter
	if s.Where != nil {
		params, err := collectAndInferWhereParams(table, s.Where)
		if err != nil {
			return nil, err
		}
		parameters, err = validateParameters(params)
		if err != nil {
			return nil, err
		}
	}

	return &UnannotatedQuery{
		Kind:          ir.Select,
		TargetTable:   tableName,
		Parameters:    parameters,
		TargetColumns: []ir.Column{},
		ResultColumns: resultColumns,
		Sql:           expandedSqlString(s, resultColumns),
	}, nil
}

func collectResultColumns(table *catalog.Table, s *parsing.SelectStatement) ([]ir.Column, error) {
	tableName := table.Name

	var resultColumns []ir.Column
	var colNames []string // * should be expanded
	if s.All {
		if err := validateColRefs(s.ColumnNames, *table); err != nil {
			return []ir.Column{}, err
		}
		for _, col := range table.Columns {
			irCol, err := ToIRColumn(tableName, col)
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
			irCol, err := ToIRColumn(tableName, *col)
			if err != nil {
				return nil, err
			}
			resultColumns = append(resultColumns, irCol)
			colNames = append(colNames, name)
		}
	}
	return resultColumns, nil
}

func expandedSqlString(s *parsing.SelectStatement, columns []ir.Column) string {
	var sb strings.Builder
	lastIndex := len(columns) - 1
	sb.WriteString("SELECT")
	for i, col := range columns {
		colName := col.Name
		sb.WriteString(fmt.Sprintf(" %s", colName))
		if i != lastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(" FROM ")
	sb.WriteString(s.TableName)
	if s.Where != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(s.Where.String())
	}
	sb.WriteString(";")
	return sb.String()
}
