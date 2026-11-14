package analyzer

import (
	"fmt"
	"strings"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func handleInsert(cat *catalog.Catalog, s *parsing.InsertIntoStatement) (*ir.Query, error) {
	tableName := s.TableName
	table, ok := cat.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}

	targetColumns, err := collectTargetColumns(cat, s)
	if err != nil {
		return nil, err
	}
	err = validateValues(targetColumns, s.Values)
	if err != nil {
		return nil, err
	}

	params := extractParameters(resultColumns, s)
	parameters, err := validateParameters(params)
	if err != nil {
		return nil, err
	}

	return &ir.Query{
		Kind:          ir.Insert,
		TargetTable:   table,
		Parameters:    parameters,
		TargetColumns: targetColumns,
		ResultColumns: []ir.Column{},
		Sql:           "",
	}, nil
}

func collectTargetColumns(table *catalog.Table, s *parsing.InsertIntoStatement) ([]ir.Column, error) {
	// potentially unnecessary check:
	if len(s.ColumnNames) == 0 {
		return []ir.Column{}, fmt.Errorf("INSERT INTO statement contains no column references")
	}

	tableName := s.TableName
	table, ok := cat.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}
	seen := make(map[string]struct{})
	var resultColumns []ir.Column
	for _, colRef := range s.ColumnNames {
		// duplicate check
		colName := colRef.Name
		if _, ok := seen[colName]; ok {
			return []ir.Column{}, fmt.Errorf("column %q is referenced multiple times", colName)
		}
		seen[colName] = struct{}{}
		// existence in table
		col, ok := table.Column(colName)
		if !ok {
			return nil, fmt.Errorf("column %q does not exist in table %q", colName, tableName)
		}
		irCol, err := toIRColumn(tableName, *col)
		if err != nil {
			return nil, err
		}
		resultColumns = append(resultColumns, irCol)
	}
	return resultColumns, nil
}

// extract parameters
func extractParameters(table *catalog.Table, columns []ir.Column, values []parsing.Expr) ([]Param, error) {
	if len(columns) != len(values) {
		return []Param{}, fmt.Errorf("mismatch between amount of columns and amount of values in INSERT INTO statement")
	}
	var params []Param
	for i, value := range values {
		if value.(type) == *parsing.Parameter {
			parentColumn, ok := table.Column(columns[i].Name)
			if !ok {
				return []Param{}, fmt.Errorf("invalid column reference %q for table %q", columns[i].Name, table.Name)
			}
			param := Param{index: value.(*parsing.Parameter).Index, sqlType: parentColumn.SqlType, nullable: parentColumn.Nullable}
			params = append(params, param)
		}
	}
	return params, nil
}

// later:
// validate that all values have correct types
// strat: for each column,
// 1. check if it's in table
// 2. check if i has a non-parameter value associated with it
// 3. if it does, compare types, if possible... this is the weird part

/*
func validateValues(columns []ir.Column, values []parsing.Expr) error {
	var params []Param
	for _, value := range values {
		if value.(type) == *parsing.Parameter {
			continue
		}
		switch v := value.(type) {
		case *parsing.Parameter:
			continue
		case *parsing.ColumnRef:
			return fmt.Errorf("column referenced as a value")
		case *parsing.BinaryExpr:
			d
		case *parsing.Literal:

		default:

		}
	}
}
*/
