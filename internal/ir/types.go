package ir

import (
	"fmt"
)

type QueryKind int

const (
	IllegalQueryKind QueryKind = iota
	Select
	Insert
)

func (q QueryKind) String() string { // should this be a Ptr method?
	strings := [...]string{"IllegalQueryKind", "Select", "Insert"}
	if q < 0 || int(q) >= len(strings) {
		return fmt.Sprintf("QueryKind(%d)", q)
	}
	return strings[q]
}

type SQLType int

const (
	IllegalSQLType SQLType = iota
	TypeText
	TypeInt
)

func (s SQLType) String() string { // should this be a Ptr method?
	strings := [...]string{"IllegalSQLType", "TypeText", "TypeInt"}
	if s < 0 || int(s) >= len(strings) {
		return fmt.Sprintf("SQLType(%d)", s)
	}
	return strings[s]
}

type Parameter struct {
	Index    int
	SqlType  SQLType
	Nullable bool
}

type Column struct {
	Name     string
	SqlType  SQLType
	Nullable bool
}

type Query struct {
	Kind          QueryKind
	TargetTable   string
	Parameters    []Parameter
	ResultColumns []Column
	Sql           string
}

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
			return fmt.Errorf("invalid SQL type for param %s", col.Name)
		}
	}

	switch q.Kind {
	case Select:
		if len(q.ResultColumns) == 0 {
			return fmt.Errorf("SELECT statement has no result columns")
		}
	case Insert:
		if len(q.ResultColumns) != 0 {
			return fmt.Errorf("INSERT statement has a nonzero amount of result columns")
		}
	default:
		return fmt.Errorf("invalid query kind %s", q.Kind)
	}

	return nil
}
