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

	params, err := walkWhere(table, s.Where)
	if err != nil {
		return nil, err
	}
	parameters, err := validateParameters(params)
	if err != nil {
		return nil, err
	}

	return &ir.Query{
		Kind:          ir.Select,
		TargetTable:   tableName,
		Parameters:    parameters,
		ResultColumns: resultColumns,
		Sql:           expandedSqlString(s, resultColumns),
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

func validateParameters(params []Param) ([]ir.Parameter, error) {
	indexToParam := make(map[int]ir.Parameter)
	maxIndexSoFar := 0
	for _, param := range params {
		if param.index <= 0 {
			return []ir.Parameter{}, fmt.Errorf("illegal parameter $%d: all parameter indices must be positive")
		}
		t, err := toSQLType(param.sqlType)
		if err != nil {
			return []ir.Parameter{}, err
		}
		if cur, ok := indexToParam[param.index]; ok {
			if cur.SqlType != t {
				return []ir.Parameter{}, fmt.Errorf("type mismatch for parameter %d (%s and %s)", param.index, t, cur)
			}
		}
		// by default, params in a select should not be nullable. (col = NULL is invalid)
		indexToParam[param.index] = ir.Parameter{Index: param.index, SqlType: t, Nullable: false}
		if param.index > maxIndexSoFar {
			maxIndexSoFar = param.index
		}
	}
	res := make([]ir.Parameter, 0)
	for i := 1; i <= maxIndexSoFar; i++ {
		p, ok := indexToParam[i]
		if !ok {
			return []ir.Parameter{}, fmt.Errorf("$%d missing", i)
		}
		res = append(res, p)

	}
	return res, nil
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
				return []Param{}, fmt.Errorf("invalid column reference %s for table %s", colRef.Name, table.Name)
			}
			param := Param{index: paramRef.Index, sqlType: col.SqlType}
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
		// for now, this is not doable. in future, we can accept casting, or inferring param types from other known param types
		return []Param{}, fmt.Errorf("param $%d must be compared to a column so that types can be inferred", t.Index)
	case *parsing.Literal:
		// check that ltieral's type is valid -- all literals are stores as strings, so should think about this.
		return []Param{}, nil
	default:
		return []Param{}, fmt.Errorf("unexpected parsing.Expr type %T", t)
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

func expandedSqlString(s *parsing.SelectStatement, columns []ir.Column) string {
	var sb strings.Builder
	lastIndex := len(columns) - 1
	sb.WriteString("SELECT")
	for i, colName := range columns {
		sb.WriteString(fmt.Sprintf(" %s", colName))
		if i != lastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(s.From.String())
	if s.Where != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(s.Where.String())
	}
	sb.WriteString(";")
	return sb.String()
}
