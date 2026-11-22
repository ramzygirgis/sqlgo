package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
)

func validateColRefs(colNames []string, table catalog.Table) error {
	for _, colName := range colNames {
		if _, ok := table.Column(colName); !ok {
			return fmt.Errorf("column %s does not belong to table %s", colName, table.Name)
		}
	}
	return nil
}
