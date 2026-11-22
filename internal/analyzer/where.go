package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

// returns a list of params. not deduped -- one should process the result of this
// to look for duplicates, and ensure that no duplicates contract eachother.
func collectAndInferWhereParams(table *catalog.Table, where parsing.Expr) ([]Param, error) {
	switch t := where.(type) {
	case *parsing.BinaryExpr:
		colRef, paramRef, isPair := columnParamPair(t)
		if isPair {
			col, ok := table.Column(colRef.Name)
			if !ok {
				return []Param{}, fmt.Errorf("invalid column reference %s for table %s", colRef.Name, table.Name)
			}
			IRCol, err := ToIRColumn(table.Name, *col)
			if err != nil {
				return []Param{}, err
			}
			param := Param{index: paramRef.Index, sqlType: IRCol.SqlType, nullable: false}
			return []Param{param}, nil
		}
		whereLeft, err := collectAndInferWhereParams(table, t.Left)
		if err != nil {
			return []Param{}, err
		}
		whereRight, err := collectAndInferWhereParams(table, t.Right)
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
