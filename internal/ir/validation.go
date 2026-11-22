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
		return err
	}

	err = validateCardinality(q.Cardinality)
	if err != nil {
		return err
	}

	switch q.Kind {
	case Select:
		return validateSelectInvariants(q)
	case Insert:
		return validateInsertInvariants(q)
	case Update:
		return validateUpdateInvariants(q)
	default:
		return fmt.Errorf("invalid query kind %s", q.Kind)
	}
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
	if !isCardinality(c) {
		return fmt.Errorf("unexpected query cardinality %s", c)
	}
	return nil
}
