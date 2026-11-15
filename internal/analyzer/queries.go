package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func AnalyzeQuery(cat *catalog.Catalog, q parsing.AnnotatedQuery) (*ir.Query, error) {
	name := q.Name
	cardinality, err := toIRCardinality(q.Cardinality)
	if err != nil {
		return nil, err
	}
	var unannotatedQuery *UnannotatedQuery
	switch s := q.Statement.(type) {
	case *parsing.SelectStatement:
		unannotatedQuery, err = handleSelect(cat, s)
		if err != nil {
			return nil, err
		}
	case *parsing.InsertIntoStatement:
		unannotatedQuery, err = handleInsert(cat, s)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported statement type %T", s)
	}
	return &ir.Query{
		Name:          name,
		Cardinality:   cardinality,
		Kind:          unannotatedQuery.Kind,
		TargetTable:   unannotatedQuery.TargetTable,
		Parameters:    unannotatedQuery.Parameters,
		TargetColumns: unannotatedQuery.TargetColumns,
		ResultColumns: unannotatedQuery.ResultColumns,
		Sql:           unannotatedQuery.Sql,
	}, nil
}
