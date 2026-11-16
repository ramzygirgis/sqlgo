package codegen

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"slices"
)

func getParamImports(params []ir.Parameter, packageImportPath string) ([]string, error) {
	var imports []string
	for _, p := range params {
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

func getImportListQueries(queries []*ir.Query, packageImportPath string, modelsImportPath string) ([]string, error) {
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
		paramImports, err := getParamImports(query.Parameters, packageImportPath)
		if err != nil {
			return nil, err
		}
		imports = append(imports, paramImports...)
	}
	slices.Sort(imports)
	imports = slices.Compact(imports)
	return imports, nil
}

func getImportStringQueries(queries []*ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	importList, err := getImportListQueries(queries, packageImportPath, modelsImportPath)
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
