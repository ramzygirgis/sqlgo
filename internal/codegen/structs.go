package codegen

import (
	"fmt"
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
	goType, err := columnToGoType(table.Name, column)
	if err != nil {
		return "", err
	}
	goColName := snakeToCamel(column.Name, true)
	return fmt.Sprintf("%s %s", goColName, goType), nil
}
