package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
)

// this guy turns a catalog column into an IR column
func toIRColumn(tableName string, col catalog.Column) (ir.Column, error) {
	sqlType, err := toSQLType(col.SqlType)
	if err != nil {
		return ir.Column{}, fmt.Errorf("column %q in table %q: %w", col.Name, tableName, err)
	}
	return ir.Column{Name: col.Name, SqlType: sqlType, Nullable: col.Nullable}, nil
}
