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
		query, err := handleSelect(cat, s)
		if err != nil {
			return nil, err
		}
		return query, nil
	case *parsing.InsertIntoStatement:
		query, err := handleInsert(cat, s)
		if err != nil {
			return nil, err
		}
		return query, nil
	default:
		return nil, fmt.Errorf("unsupported statement type %T", s)
	}
}
