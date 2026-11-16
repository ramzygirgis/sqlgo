package codegen

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"slices"
)

func getGoTypeImports(t GoType, packageImportPath string) ([]string, error) {
	var imports []string
	if t.ImportPath != "" && t.ImportPath != packageImportPath {
		imports = append(imports, t.ImportPath)
	}
	return imports, nil
}

func getParamsImports(params []ir.Parameter, packageImportPath string) ([]string, error) {
	var imports []string
	for _, p := range params {
		t, err := irToGoType(p.SqlType, p.Nullable)
		if err != nil {
			return nil, err
		}
		paramImports, err := getGoTypeImports(t, packageImportPath)
		imports = append(imports, paramImports...)
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
		paramImports, err := getParamsImports(query.Parameters, packageImportPath)
		if err != nil {
			return nil, err
		}
		imports = append(imports, paramImports...)
	}
	slices.Sort(imports)
	imports = slices.Compact(imports)
	return imports, nil
}

func getImportListModels(catalog catalog.Catalog, modelsImportPath string) ([]string, error) {
	var imports = []string{}
	for _, table := range catalog.Tables {
		for _, column := range table.Columns {
			t, err := columnToGoType(table.Name, column)
			if err != nil {
				return nil, err
			}
			colImports, err := getGoTypeImports(t, modelsImportPath)
			if err != nil {
				return nil, err
			}
			imports = append(imports, colImports...)
		}
	}
	imports = slices.Compact(imports)
	slices.Sort(imports)
	return imports, nil
}

func getImportStringQueries(queries []*ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	importList, err := getImportListQueries(queries, packageImportPath, modelsImportPath)
	if err != nil {
		return "", err
	}
	importString := "import (\n"
	for _, impStr := range importList {
		importString += fmt.Sprintf(`\t"%s"\n`, impStr)
	}
	importString += ")"
	return importString, nil
}

func getImportStringModels(catalog catalog.Catalog, modelsImportPath string) (string, error) {
	importList, err := getImportListModels(catalog, modelsImportPath)
	if err != nil {
		return "", err
	}
	importString := "import (\n"
	for _, impStr := range importList {
		importString += fmt.Sprintf(`\t"%s"\n`, impStr)
	}
	importString += ")"
	return importString, nil
}
