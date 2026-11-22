package ir

import (
	"fmt"
)

func validateSelectInvariants(q *Query) error {
	if q.Kind != Select {
		return fmt.Errorf("non-SELECT statement passed into ir.validateSelectInvariants")
	}
	/*
		SELECT statements should have:
		1. no result columns
		2. nonzero number of target columns
		3. cardinality must be one of: One, Many
	*/
	if len(q.ResultColumns) == 0 {
		return fmt.Errorf("SELECT statement has no result columns")
	}
	if len(q.TargetColumns) != 0 {
		return fmt.Errorf("SELECT statement has target columns")
	}
	if !isSelectCardinality(q.Cardinality) {
		return fmt.Errorf("SELECT statement cannot have cardinality %s", q.Cardinality)
	}
	return nil
}

func validateInsertInvariants(q *Query) error {
	if q.Kind != Insert {
		return fmt.Errorf("non-INSERT statement passed into ir.validateInsertInvariants")
	}
	/*
	 INSERT statements should have:
	 1. no result columns
	 2. nonzero amount of target columns
	 3. cardinality must be one of: Exec
	*/
	if len(q.ResultColumns) != 0 {
		return fmt.Errorf("INSERT statement has result columns")
	}
	if len(q.TargetColumns) == 0 {
		return fmt.Errorf("INSERT statement has no target columns")
	}
	if !isInsertCardinality(q.Cardinality) {
		return fmt.Errorf("INSERT statement cannot have cardinality %s", q.Cardinality)
	}
	return nil
}

func validateUpdateInvariants(q *Query) error {
	if q.Kind != Update {
		return fmt.Errorf("non-UPDATE statement passed into ir.validateUpdateInvariants")
	}
	/*
	 UPDATE statements should have:
	 1. no result columns
	 2. nonzero amount of target columns
	 3. cardinality must be one of: Exec
	*/
	if len(q.ResultColumns) != 0 {
		return fmt.Errorf("INSERT statement has result columns")
	}
	if len(q.TargetColumns) == 0 {
		return fmt.Errorf("INSERT statement has no target columns")
	}
	if !isUpdateCardinality(q.Cardinality) {
		return fmt.Errorf("INSERT statement cannot have cardinality %s", q.Cardinality)
	}
	return nil
}
