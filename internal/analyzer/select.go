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

type Param struct {
	Index int
	Type  string // i feel like this should be an iota
}

// returns a list of params. not deduped -- one should process the result of this
// to look for duplicates, and ensure that no duplicates contract eachother.
func walkWhere(table *catalog.Table, where parsing.Expr) ([]Param, error) {
	switch t := where.(type) {
	case *parsing.BinaryExpr:
		colRef, paramRef, isPair := columnParamPair(t)
		if isPair {
			col, ok := table.Column(colRef.Name)
			if !ok {
				return []Param{}, fmt.Errorf("invalid column reference %s for table %s", colRef, table.Name)
			}
			param := Param{Index: paramRef.Index, Type: col.SqlType}
			return []Param{param}, nil
		}
		whereLeft, err := walkWhere(table, t.Left)
		if err != nil {
			return []Param{}, err
		}
		whereRight, err := walkWhere(table, t.Right)
		if err != nil {
			return []Param{}, err
		}
		return append(whereLeft, whereRight...), nil
	case *parsing.ColumnRef:
		if _, ok := table.Column(t.Name); !ok {
			return []Param{}, fmt.Errorf("invalid column reference %s for table %s", t.Name, table.Name)
		}
		return []Param{}, nil
	case *parsing.Parameter:
		// should probably throw here
		return []Param{}, nil
	case *parsing.Literal:
		// check that ltieral's type is valid -- all literals are stores as strings, so should think about this.
		return []Param{}, nil
	default:
		return []Param{}, fmt.Errorf("unexpected parsing.Expr type %s", t)
	}
}

func columnParamPair(b *parsing.BinaryExpr) (*parsing.ColumnRef, *parsing.Parameter, bool) {
	if b.Op == "AND" || b.Op == "OR" {
		return nil, nil, false
	}
	rightParam, rightIsParam := b.Right.(*parsing.Parameter)
	if rightIsParam {
		leftCol, leftIsCol := b.Left.(*parsing.ColumnRef)
		if leftIsCol {
			return leftCol, rightParam, true
		}
	}

	leftParam, leftIsParam := b.Left.(*parsing.Parameter)
	if leftIsParam {
		rightCol, rightIsCol := b.Right.(*parsing.ColumnRef)
		if rightIsCol {
			return rightCol, leftParam, true
		}
	}
	return nil, nil, false
}
