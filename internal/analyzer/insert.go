package analyzer

import (
	"fmt"

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
	// the following would be a catalog bug, not sure if we validate already
	if tableName != table.Name {
		return nil, fmt.Errorf("catalog maps %q to table %q", tableName, table.Name)
	}

	targetColumns, err := collectTargetColumns(table, s)
	if err != nil {
		return nil, err
	}

	// todo: combine these into one call
	params, err := inferInsertParams(table, targetColumns, s.Values)
	if err != nil {
		return nil, err
	}
	parameters, err := validateParameters(params)
	if err != nil {
		return nil, err
	}

	return &ir.Query{
		Kind:          ir.Insert,
		TargetTable:   tableName,
		Parameters:    parameters,
		TargetColumns: targetColumns,
		ResultColumns: []ir.Column{},
		Sql:           s.String(),
	}, nil
}

func collectTargetColumns(table *catalog.Table, s *parsing.InsertIntoStatement) ([]ir.Column, error) {
	var columns []catalog.Column
	if len(s.Columns) == 0 {
		columns = table.Columns
	} else {
		seen := make(map[string]struct{})
		for _, colRef := range s.Columns {
			colName := colRef.Name
			if _, ok := seen[colName]; ok {
				return []ir.Column{}, fmt.Errorf("column %q is referenced multiple times", colName)
			}
			seen[colName] = struct{}{}
			col, ok := table.Column(colName)
			if !ok {
				return nil, fmt.Errorf("column %q does not exist in table %q", colName, table.Name)
			}
			columns = append(columns, *col)
		}
	}

	var resultColumns []ir.Column
	for _, col := range columns {
		irCol, err := toIRColumn(table.Name, col)
		if err != nil {
			return nil, err
		}
		resultColumns = append(resultColumns, irCol)
	}

	return resultColumns, nil
}

// extract parameters and validate values
func inferInsertParams(table *catalog.Table, columns []ir.Column, values []parsing.Expr) ([]Param, error) {
	// todo: handling for typechecking literal.Body, and for type inferring binary expressions
	if len(columns) != len(values) {
		return []Param{}, fmt.Errorf("mismatch between amount of columns and amount of values in INSERT INTO statement")
	}
	var params []Param
	for i, value := range values {
		switch v := value.(type) {
		case *parsing.Parameter:
			param := Param{index: v.Index, sqlType: columns[i].SqlType, nullable: columns[i].Nullable}
			params = append(params, param)
		case *parsing.ColumnRef:
			return []Param{}, fmt.Errorf("cannot pass column reference %q as a value in an INSERT INTO statement", v.Name)
		case *parsing.BinaryExpr:
			err := validateBinaryExpr(v, columns[i])
			if err != nil {
				return []Param{}, err
			}
		case *parsing.Literal:
			err := validateLiteral(v, columns[i])
			if err != nil {
				return []Param{}, err
			}
		default:
			return []Param{}, fmt.Errorf("unsupported expression type %T", v)
		}
	}
	return params, nil
}

func validateLiteral(l *parsing.Literal, column ir.Column) error {
	// todo: compare value's type with column's type, such as
	// - if a literal is castable
	if l.Body == "NULL" && !column.Nullable {
		return fmt.Errorf("NULL cannot be inserted into %q, as this column is not nullable", column.Name)
	}
	return nil
}

func validateBinaryExpr(b *parsing.BinaryExpr, column ir.Column) error {
	return fmt.Errorf("unsupported expression in VALUES; probably needs to be evaluated")
}
