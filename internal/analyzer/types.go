package analyzer

import (
	"github.com/ramzygirgis/sqlgo/internal/ir"
)

type Param struct {
	index    int
	sqlType  ir.SQLType
	nullable bool
}

type UnannotatedQuery struct {
	Kind          ir.QueryKind
	TargetTable   string
	Parameters    []ir.Parameter
	TargetColumns []ir.Column
	ResultColumns []ir.Column
	Sql           string
}
