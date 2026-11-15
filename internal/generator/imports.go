package generator

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"slices"
)

func getParamImports(query *ir.Query, packageImportPath string) ([]string, error) {
	var imports []string
	for _, p := range query.Parameters {
		t, err := IRToGoType(p.SqlType, p.Nullable)
		if err != nil {
			return nil, err
		}
		if t.ImportPath != "" && t.ImportPath != packageImportPath {
			imports = append(imports, t.ImportPath)
		}
	}
	return imports, nil
}

func getImportList(queries []*ir.Query, packageImportPath string, modelsImportPath string) ([]string, error) {
	var imports = []string{"context"}
	for _, query := range queries {
		if query.Cardinality == ir.One {
			imports = append(imports, "fmt")
			imports = append(imports, "database/sql")
			break
		}
	}
	if packageImportPath != modelsImportPath {
		imports = append(imports, modelsImportPath)
	}
	for _, query := range queries {
		paramImports, err := getParamImports(query, packageImportPath)
		if err != nil {
			return nil, err
		}
		imports = append(imports, paramImports...)
	}
	slices.Sort(imports)
	imports = slices.Compact(imports)
	return imports, nil
}

func getImportString(queries []*ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	importList, err := getImportList(queries, packageImportPath, modelsImportPath)
	if err != nil {
		return "", err
	}
	importString := "import (\n"
	for _, i := range importList {
		importString += fmt.Sprintf(`\t"%s"\n`, i)
	}
	importString += ")"
	return importString, nil
}
