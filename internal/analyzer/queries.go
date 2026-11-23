package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func AnalyzeQuery(cat *catalog.Catalog, q *parsing.AnnotatedQuery) (*ir.Query, error) {
	queryName := q.Name
	cardinality, err := toIRCardinality(q.Cardinality)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", queryName, err)
	}
	var unannotatedQuery *UnannotatedQuery
	switch s := q.Statement.(type) {
	case *parsing.SelectStatement:
		unannotatedQuery, err = handleSelect(cat, s)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", queryName, err)
		}
	case *parsing.InsertIntoStatement:
		unannotatedQuery, err = handleInsert(cat, s)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", queryName, err)
		}
	case *parsing.UpdateStatement:
		unannotatedQuery, err = handleUpdate(cat, s)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", queryName, err)
		}
	case *parsing.DeleteFromStatement:
		unannotatedQuery, err = handleDelete(cat, s)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", queryName, err)
		}
	default:
		return nil, fmt.Errorf("%s: unsupported statement type %T", queryName, s)
	}
	query := &ir.Query{
		Name:          queryName,
		Cardinality:   cardinality,
		Kind:          unannotatedQuery.Kind,
		TargetTable:   unannotatedQuery.TargetTable,
		Parameters:    unannotatedQuery.Parameters,
		TargetColumns: unannotatedQuery.TargetColumns,
		ResultColumns: unannotatedQuery.ResultColumns,
		Sql:           unannotatedQuery.Sql,
	}
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %s", queryName, err)
	}
	return query, nil
}
