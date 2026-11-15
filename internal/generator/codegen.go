package generator

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"go/format"
	"path"
)

func generateFile(queries []*ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	packageDeclaration := getPackageDeclaration(packageImportPath)

	importString, err := getImportString(queries, packageImportPath, modelsImportPath)
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

func getPackageDeclaration(packageImportPath string) string {
	packageName := path.Base(packageImportPath)
	return fmt.Sprintf("package %s", packageName)
}
