package codegen

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/analyzer"
	"github.com/ramzygirgis/sqlgo/internal/catalog"
)

func generateStruct(table *catalog.Table) (string, error) {
	structName := modelName(table.Name)
	structBody, err := generateStructBody(table)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("type %s struct {\n%s\n}", structName, structBody), nil
}

func generateStructBody(table *catalog.Table) (string, error) {
	body := ""
	for _, column := range table.Columns {
		colString, err := generateColumn(table, column)
		if err != nil {
			return "", err
		}
		body += fmt.Sprintf("\n%s", colString)
	}
	return body, nil
}

func generateColumn(table *catalog.Table, column catalog.Column) (string, error) {
	IRCol, err := analyzer.ToIRColumn(table.Name, column)
	if err != nil {
		return "", err
	}

	goName := snakeToCamel(IRCol.Name, IRCol.Nullable)
	goType, err := IRToGoType(IRCol.SqlType, IRCol.Nullable)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", goName, goType), nil
}
