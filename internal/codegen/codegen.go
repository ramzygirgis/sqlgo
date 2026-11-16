package codegen

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"go/format"
)

func generateQueriesFile(queries []*ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	packageDeclaration := getPackageDeclaration(packageImportPath)

	importString, err := getImportStringQueries(queries, packageImportPath, modelsImportPath)
	if err != nil {
		return "", err
	}
	rawCode := ""
	rawCode += packageDeclaration
	rawCode += "\n\n"
	rawCode += importString

	for _, query := range queries {
		generatedCode, err := generateQuery(query, packageImportPath, modelsImportPath)
		if err != nil {
			return "", err
		}
		rawCode += fmt.Sprintf("\n\n%s", generatedCode)
	}
	cleanedCode, err := format.Source([]byte(rawCode))
	if err != nil {
		return "", fmt.Errorf("file generation generated invalid Go: %w\n%s", err, rawCode)
	}
	return string(cleanedCode), nil
}

func generateModelsFile(catalog catalog.Catalog, modelsImportPath string) (string, error) {
	// make sure that the below call is correct
	packageDeclaration := getPackageDeclaration(modelsImportPath)
	// make sure that the above call is correct
	importString, err := getImportStringModels(catalog, modelsImportPath)
	if err != nil {
		return "", err
	}
	rawCode := ""
	rawCode += packageDeclaration
	rawCode += "\n\n"
	rawCode += importString

	for _, table := range catalog.Tables {
		generatedStruct, err := generateStruct(table)
		if err != nil {
			return "", err
		}
		rawCode += fmt.Sprintf("\n\n%s", generatedStruct)
	}
	cleanedCode, err := format.Source([]byte(rawCode))
	if err != nil {
		return "", err
	}
	return string(cleanedCode), nil
}
