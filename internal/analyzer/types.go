package analyzer

import (
	"github.com/ramzygirgis/sqlgo/internal/ir"
)

type Param struct {
	index    int
	sqlType  ir.SQLType
	nullable bool
}
