package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func handleDelete(cat *catalog.Catalog, s *parsing.DeleteFromStatement) (*UnannotatedQuery, error) {
	tableName := s.TableName
	table, ok := cat.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}

	targetColumns := make([]ir.Column, 0)

	resultColumns := make([]ir.Column, 0)

	parameters, err := collectAndInferDeleteParams(table, s)
	if err != nil {
		return nil, err
	}

	return &UnannotatedQuery{
		Kind:          ir.Delete,
		TargetTable:   tableName,
		Parameters:    parameters,
		TargetColumns: targetColumns,
		ResultColumns: resultColumns,
		Sql:           s.String(),
	}, nil
}

func collectAndInferDeleteParams(table *catalog.Table, s *parsing.DeleteFromStatement) ([]ir.Parameter, error) {
	var params []Param
	var err error
	if s.Where != nil {
		params, err = collectAndInferWhereParams(table, s.Where)
		if err != nil {
			return nil, err
		}
	}

	parameters, err := validateParameters(params)
	if err != nil {
		return nil, err
	}

	return parameters, nil
}
