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

	params := extractParameters(resultColumns, s)
	parameters, err := validateParameters(params)
	if err != nil {
		return nil, err
	}

	return &ir.Query{
		Kind: ir.Insert,
		TargetTable: table,
		Parameters: parameters,
		TargetColumns: targetColumns,
		ResultColumns: []ir.Column{},
		Sql: "",
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
func extractParameters(columns []ir.Column, s *parsing.InsertIntoStatement) []Param {
	var params []Param
	for i, value := range values {
		if value.(type) == *parsing.Parameter {
			parentColumn := columns[i]
			param := Param{index: v.Index, sqlType: parentColumn.SqlType, nullable: parentColumn.Nullable}
			params = append(params, param)
		}
	}
	return params
}

// validate params:
// call validateParams lol

// validate that all values have correct types
// strat: for each column,
// 1. check if it's in table
// 2. check if i has a non-parameter value associated with it
// 3. if it does, compare types, if possible... this is the weird part



func parseValues(table catalog.Table, values []parsing.Expr) ([]ir.Parameter, []ir.Values) {
	var params []Param
	for _, value := range values {
		switch v := value.(type) {
		case *parsing.Parameter:
			// prob skip????
		case *parsing.BinaryExpr:
		case *parsing.ColumnRef:
		case *parsing.Literal:
		default:

		}
	}
}


what i need to do:

get TableName
get Parameters from VALUES ( ... , ... , ... , ... )
--> 
get ResultColumns from ( ... , ... , ... , ... )
