package ir

import (
	"fmt"
)

func (q *Query) Validate() error {
	for i, param := range q.Parameters {
		if param.Index != i+1 {
			return fmt.Errorf("parameter indexing must be ordered and contiguous from 1-N: expected $%d, got $%d", i+1, param.Index)
		}
		if param.SqlType == IllegalSQLType {
			return fmt.Errorf("invalid SQL type for parameter $%d", param.Index)
		}
	}

	for _, col := range q.ResultColumns {
		if col.SqlType == IllegalSQLType {
			return fmt.Errorf("invalid SQL type for column %s", col.Name)
		}
	}
	for _, col := range q.TargetColumns {
		if col.SqlType == IllegalSQLType {
			return fmt.Errorf("invalid SQL type for column %s", col.Name)
		}
	}

	if q.Name == "" {
		return fmt.Errorf("query name is empty")
	}

	if q.Cardinality != One && q.Cardinality != Many && q.Cardinality != Exec {
		return fmt.Errorf("unexpected query cardinality %s", q.Cardinality)
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
			return fmt.Errorf("INSERT statement can only have cardinality :exec")
		}
	default:
		return fmt.Errorf("invalid query kind %s", q.Kind)
	}

	return nil
}
