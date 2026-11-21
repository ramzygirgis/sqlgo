package ir

import (
	"fmt"
)

func (q *Query) Validate() error {
	err := validateParameters(q.Parameters)
	if err != nil {
		return err
	}

	err = validateColumns(q.ResultColumns)
	if err != nil {
		return err
	}

	err = validateColumns(q.TargetColumns)
	if err != nil {
		return err
	}

	err = validateName(q.Name)
	if err != nil {
		return nil
	}

	err = validateCardinality(q.Cardinality)
	if err != nil {
		return err
	}

	switch q.Kind {
	case Select:
		if len(q.ResultColumns) == 0 {
			return fmt.Errorf("SELECT statement has no result columns")
		}
		if len(q.TargetColumns) != 0 {
			return fmt.Errorf("SELECT statement has target columns")
		}
		if q.Cardinality == Exec {
			return fmt.Errorf("SELECT statement cannot have cardinality :exec")
		}
	case Insert:
		if len(q.ResultColumns) != 0 {
			return fmt.Errorf("INSERT statement has a nonzero amount of result columns")
		}
		if len(q.TargetColumns) == 0 {
			return fmt.Errorf("INSERT statement has no target columns")
		}
		if q.Cardinality != Exec {
			return fmt.Errorf("INSERT statement can only have cardinality :exec -- got %s", q.Cardinality)
		}
	default:
		return fmt.Errorf("invalid query kind %s", q.Kind)
	}

	return nil
}

func validateParameters(params []Parameter) error {
	for i, param := range params {
		if param.Index != i+1 {
			return fmt.Errorf("parameter indexing must be ordered and contiguous from 1-N: expected $%d, got $%d", i+1, param.Index)
		}
		if !isType(param.SqlType) {
			return fmt.Errorf("invalid SQL type %s for parameter $%d", param.SqlType, param.Index)
		}
	}
	return nil
}

func validateColumns(cols []Column) error {
	for _, col := range cols {
		if !isType(col.SqlType) {
			return fmt.Errorf("invalid SQL type %s for column %s", col.SqlType, col.Name)
		}
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("query name is empty")
	}
	return nil
}

func validateCardinality(c Cardinality) error {
	if c != One && c != Many && c != Exec {
		return fmt.Errorf("unexpected query cardinality %s", c)
	}
	return nil
}
