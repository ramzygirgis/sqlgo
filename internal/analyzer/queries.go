package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func AnalyzeQuery(cat *catalog.Catalog, statement parsing.Statement) (*ir.Query, error) {
	switch s := statement.(type) {
	case *parsing.SelectStatement:
		return handleSelect(cat, s)
	case *parsing.InsertIntoStatement:
		return handleInsert(cat, s)
	default:
		return nil, fmt.Errorf("unsupported statement type %T", s)
	}
}
