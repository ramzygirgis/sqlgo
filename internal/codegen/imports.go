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

// get imports for a given column representation

func getParamImports(param ir.Parameter, packageImportPath string) ([]string, error) {
	t, err := irToGoType(param.SqlType, param.Nullable)
	if err != nil {
		return nil, err
	}
	paramImports, err := getGoTypeImports(t, packageImportPath)
	return paramImports, nil
}

func getColumnImports(table *catalog.Table, col catalog.Column, modelsImportPath string) ([]string, error) {
	t, err := columnToGoType(table.Name, col)
	if err != nil {
		return nil, err
	}
	colImports, err := getGoTypeImports(t, modelsImportPath)
	if err != nil {
		return nil, err
	}
	return colImports, nil
}

// get imports for a list of a given column representation

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
		for _, param := range query.Parameters {
			paramImports, err := getParamImports(param, packageImportPath)
			if err != nil {
				return nil, err
			}
			imports = append(imports, paramImports...)
		}
	}
	slices.Sort(imports)
	imports = slices.Compact(imports)
	return imports, nil
}

func getImportListModels(catalog catalog.Catalog, modelsImportPath string) ([]string, error) {
	var imports = []string{}
	for _, table := range catalog.Tables {
		for _, column := range table.Columns {
			colImports, err := getColumnImports(table, column, modelsImportPath)
			if err != nil {
				return nil, err
			}
			imports = append(imports, colImports...)
		}
	}
	slices.Sort(imports)
	imports = slices.Compact(imports)
	return imports, nil
}

func getImportListDb() []string {
	imports := []string{"database/sql", "context", "errors"}
	slices.Sort(imports)
	return imports
}

// get imports for something high level (query, table)

func getImportStringQueries(queries []*ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	importList, err := getImportListQueries(queries, packageImportPath, modelsImportPath)
	if err != nil {
		return "", err
	}
	importString := "import (\n"
	for _, impStr := range importList {
		importString += fmt.Sprintf("\t%q\n", impStr)
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
		importString += fmt.Sprintf("\t%q\n", impStr)
	}
	importString += ")"
	return importString, nil
}

func getImportStringDb() string {
	importList := getImportListDb()
	importString := "import (\n"
	for _, impStr := range importList {
		importString += fmt.Sprintf("\t%q\n", impStr)
	}
	importString += ")"
	return importString
}
