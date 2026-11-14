package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
)

func validateColRefs(colNames []string, table catalog.Table) error {
	for _, colName := range colNames {
		if _, ok := table.Column(colName); !ok {
			return fmt.Errorf("column %s does not belong to table %s", colName, table.Name)
		}
	}
	return nil
}

func validateParameters(params []Param) ([]ir.Parameter, error) {
	indexToParam := make(map[int]ir.Parameter)
	maxIndexSoFar := 0
	for _, param := range params {
		if param.index <= 0 {
			return []ir.Parameter{}, fmt.Errorf("illegal parameter $%d: all parameter indices must be positive", param.index)
		}
		t, err := toSQLType(param.sqlType)
		if err != nil {
			return []ir.Parameter{}, err
		}
		if cur, ok := indexToParam[param.index]; ok {
			if cur.SqlType != t {
				return []ir.Parameter{}, fmt.Errorf("type mismatch for parameter %d (%s and %s)", param.index, t, cur.SqlType)
			}
		}
		// by default, params in a select should not be nullable. (col = NULL is invalid)
		indexToParam[param.index] = ir.Parameter{Index: param.index, SqlType: t, Nullable: false}
		if param.index > maxIndexSoFar {
			maxIndexSoFar = param.index
		}
	}
	var res []ir.Parameter
	for i := 1; i <= maxIndexSoFar; i++ {
		p, ok := indexToParam[i]
		if !ok {
			return []ir.Parameter{}, fmt.Errorf("$%d missing", i)
		}
		res = append(res, p)

	}
	return res, nil
}
