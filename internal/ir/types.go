package ir

import (
	"fmt"
)

type QueryKind int

const (
	IllegalQueryKind QueryKind = iota
	Select
	Insert
	Update
	Delete
)

func (q QueryKind) String() string {
	strings := [...]string{"IllegalQueryKind", "Select", "Insert", "Update", "Delete"}
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
	TypeTimestamp
	TypeBoolean
)

func (s SQLType) String() string { // should this be a Ptr method?
	strings := [...]string{
		"IllegalSQLType", "TypeText", "TypeInt",
		"TypeTimestamp", "TypeBoolean",
	}
	if s < 0 || int(s) >= len(strings) {
		return fmt.Sprintf("SQLType(%d)", s)
	}
	return strings[s]
}

type Cardinality int

const (
	IllegalCardinality Cardinality = iota
	One
	Many
	Exec
)

func (c Cardinality) String() string {
	strings := [...]string{"IllegalCardinality",
		"One", "Many", "Exec",
	}
	if c < 0 || int(c) >= len(strings) {
		return fmt.Sprintf("Cardinality(%d)", c)
	}
	return strings[c]
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
	Name          string
	Cardinality   Cardinality
	Kind          QueryKind
	TargetTable   string
	Parameters    []Parameter
	TargetColumns []Column
	ResultColumns []Column
	Sql           string
}
