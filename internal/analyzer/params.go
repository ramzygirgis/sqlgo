package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/ir"
)

func validateParameters(params []Param) ([]ir.Parameter, error) {
	indexToParam := make(map[int]ir.Parameter)
	maxIndexSoFar := 0
	for _, param := range params {
		if param.index <= 0 {
			return []ir.Parameter{}, fmt.Errorf("illegal parameter $%d: all parameter indices must be positive", param.index)
		}
		if cur, ok := indexToParam[param.index]; ok {
			if cur.SqlType != param.sqlType {
				return []ir.Parameter{}, fmt.Errorf("type mismatch for parameter %d (%s and %s)", param.index, cur.SqlType, param.sqlType)
			}
		}
		// by default, params in a select should not be nullable. (col = NULL is invalid)
		indexToParam[param.index] = ir.Parameter{Index: param.index, SqlType: param.sqlType, Nullable: param.nullable}
		if param.index > maxIndexSoFar {
			maxIndexSoFar = param.index
		}
	}
	var res []ir.Parameter
	for i := 1; i <= maxIndexSoFar; i++ {
		p, ok := indexToParam[i]
		if !ok {
			return []ir.Parameter{}, fmt.Errorf("parameter $%d missing", i)
		}
		res = append(res, p)

	}
	return res, nil
}
