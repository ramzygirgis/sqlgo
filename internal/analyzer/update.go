package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func handleUpdate(cat *catalog.Catalog, s *parsing.UpdateStatement) (*UnannotatedQuery, error) {
	tableName := s.TableName
	table, ok := cat.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}

	targetColumns, err := collectUpdateTargetColumns(table, s.CvPairs)
	if err != nil {
		return nil, err
	}

	parameters, err := collectAndInferUpdateParams(table, s)
	if err != nil {
		return nil, err
	}

	resultColumns := make([]ir.Column, 0)

	return &UnannotatedQuery{
		Kind:          ir.Update,
		TargetTable:   tableName,
		Parameters:    parameters,
		TargetColumns: targetColumns,
		ResultColumns: resultColumns,
		Sql:           s.String(),
	}, nil

}

func collectUpdateTargetColumns(table *catalog.Table, cvPairs []parsing.ColumnValuePair) ([]ir.Column, error) {
	var targetColumns []ir.Column
	for _, cvPair := range cvPairs {
		col, ok := table.Column(cvPair.ColumnRef.Name)
		if !ok {
			return nil, fmt.Errorf("column %q does not appear in table %q", cvPair.ColumnRef.Name, table.Name)
		}
		column, err := ToIRColumn(table.Name, *col)
		if err != nil {
			return nil, err
		}
		targetColumns = append(targetColumns, column)
	}
	return targetColumns, nil
}

func collectAndInferUpdateParams(table *catalog.Table, s *parsing.UpdateStatement) ([]ir.Parameter, error) {
	cvParams, err := collectAndInferCVParams(table, s.CvPairs)
	if err != nil {
		return nil, err
	}

	var whereParams []Param
	if s.Where != nil {
		whereParams, err = collectAndInferWhereParams(table, s.Where)
		if err != nil {
			return nil, err
		}
	}

	params := append(cvParams, whereParams...)

	// look for mismatches
	parameters, err := validateParameters(params)
	if err != nil {
		return nil, err
	}

	return parameters, nil
}

func collectAndInferCVParams(table *catalog.Table, cvPairs []parsing.ColumnValuePair) ([]Param, error) {
	var params []Param
	for _, cvPair := range cvPairs {
		switch v := cvPair.Value.(type) {
		// it would be nice to switch this for an `if` statement
		case *parsing.Parameter:
			column, ok := table.Column(cvPair.ColumnRef.Name)
			if !ok {
				return nil, fmt.Errorf("column %q does not appear in table %q", cvPair.ColumnRef.Name, table.Name)
			}
			sqlType, err := toSQLType(column.SqlType)
			if err != nil {
				return nil, err
			}
			param := Param{index: v.Index, sqlType: sqlType, nullable: column.Nullable}
			params = append(params, param)
		case *parsing.Literal:
			column, ok := table.Column(cvPair.ColumnRef.Name)
			if !ok {
				return nil, fmt.Errorf("column %q does not appear in table %q", cvPair.ColumnRef.Name, table.Name)
			}
			irColumn, err := ToIRColumn(table.Name, *column)
			if err != nil {
				return nil, err
			}
			if err := validateLiteral(v, irColumn); err != nil {
				return nil, err
			}
		}
	}
	return params, nil
}
