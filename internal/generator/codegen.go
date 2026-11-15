package generator

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"path"
	"strings"
)

func codegen(query *ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	funcContract, err := functionContract(query, packageImportPath, modelsImportPath)
	if err != nil {
		return "", err
	}

	var funcBody string
	switch query.Cardinality {
	case ir.One:
		funcBody, err = functionBodyOne(query, packageImportPath, modelsImportPath)
		if err != nil {
			return "", err
		}
	case ir.Many:
		funcBody, err = functionBodyMany(query, packageImportPath, modelsImportPath)
		if err != nil {
			return "", err
		}
	case ir.Exec:
		funcBody, err = functionBodyExec(query)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unexpected query cardinality %s", query.Cardinality)
	}

	res := ""
	res += fmt.Sprintf("const %s = `%s`\n\n", queryName(query), queryBody(query))
	res += fmt.Sprintf("%s {\n\t%s\n}", funcContract, funcBody)
	return res, nil
}

func queryBody(query *ir.Query) string {
	return query.Sql
}

func queryName(query *ir.Query) string {
	// adding \n\n at the end is the parent's job
	return strings.ToLower(string(query.Name[0])) + query.Name[1:]
}

func functionName(query *ir.Query) string {
	return strings.ToUpper(string(query.Name[0])) + query.Name[1:]
}

func functionContract(query *ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	// this guy writes func name(...) (... , ... , ...), but not the {.
	// that, and the } is the job of the parent
	// later, add support for smart naming function arguments
	res := fmt.Sprintf("func %s(ctx context.Context, db models.DBTX", functionName(query))
	for _, param := range query.Parameters {
		goType, err := IRToGoType(param.SqlType, param.Nullable)
		if err != nil {
			return "", err
		}
		goTypeString := goType.In(packageImportPath)
		res += fmt.Sprintf(", arg%d %s", param.Index, goTypeString)
	}
	res += ") "
	modelType := GoType{
		TypeName:    modelName(query.TargetTable),
		PackageName: path.Base(modelsImportPath),
		ImportPath:  modelsImportPath,
	}
	modelNameGo := modelType.In(packageImportPath)
	returnTypeString := ""
	switch query.Cardinality {
	case ir.One:
		returnTypeString += fmt.Sprintf("%s, error)", modelNameGo)
	case ir.Many:
		returnTypeString += fmt.Sprintf("(%s, error)", modelNameGo)
	case ir.Exec:
		returnTypeString += "error"
	default:
		return "", fmt.Errorf("unsupported cardinality %s", query.Cardinality)
	}
	res += returnTypeString
	return res, nil
}

func functionBodyOne(query *ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	// note: when we do schema parsing, we will convert every row's name to a go type using snakeToCamel
	if query.Cardinality != ir.One {
		return "", fmt.Errorf("(functionBodyOne expects cardinality ir.One, got %s", query.Cardinality)
	}
	modelType := GoType{
		TypeName:    modelName(query.TargetTable),
		PackageName: path.Base(modelsImportPath),
		ImportPath:  modelsImportPath,
	}
	modelNameGo := modelType.In(packageImportPath)
	res := ""
	res += fmt.Sprintf("rows, err := db.QueryContext(ctx, %s", queryName(query))
	for _, param := range query.Parameters {
		res += fmt.Sprintf(", arg%d", param.Index)
	}
	res += ")\n\t"
	res += "if err != nil {\n\t\treturn nil, err\n\t}\n\t"
	res += "defer rows.Close()\n\t"
	res += "if !rows.Next() {\n\t\t"
	res += "if err := rows.Err(); err != nil {\n\t\t\t"
	res += fmt.Sprintf("return %s{}, err\n\t\t", modelNameGo)
	res += "}\n\t\t"
	res += fmt.Sprintf("return %s{}, sql.ErrNoRows\n\t", modelNameGo)
	res += fmt.Sprintf("var i %s\n\t", modelNameGo)
	res += "if err := rows.Scan("
	for i, col := range query.ResultColumns {
		colNameGo := snakeToCamel(col.Name, true)
		if i == 0 {
			res += fmt.Sprintf("&i.%s", colNameGo)
		} else {
			res += fmt.Sprintf(", &i.%s", colNameGo)
		}
	}
	res += "); err != nil {\n\t\t"
	res += fmt.Sprintf("return models.%s{}, err\n\t", modelNameGo)
	res += "}\n\t"
	res += "if rows.Next() {\n\t\t"
	res += fmt.Sprintf(`return models.%s{}, fmt.Errorf("%s: expected one row, got more")\n\t`, modelNameGo, functionName(query))
	res += "}\n\t"
	res += "return i, rows.Err()"

	return res, nil
}

func functionBodyMany(query *ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	if query.Cardinality != ir.Many {
		return "", fmt.Errorf("(functionBodyMany expects cardinality ir.Many, got %s", query.Cardinality)
	}
	modelType := GoType{
		TypeName:    modelName(query.TargetTable),
		PackageName: path.Base(modelsImportPath),
		ImportPath:  modelsImportPath,
	}
	modelNameGo := modelType.In(packageImportPath)
	res := ""
	res += fmt.Sprintf("rows, err := db.QueryContext(ctx, %s", queryName(query))
	for _, param := range query.Parameters {
		res += fmt.Sprintf(", arg%d", param.Index)
	}
	res += ")\n\t"
	res += "if err != nil {\n\t\treturn nil, err\n\t}\n\t"
	res += "defer rows.Close()\n\t"
	res += fmt.Sprintf("var items []%s\n\t", modelNameGo)
	res += "for rows.Next() {\n\t\t"
	res += fmt.Sprintf("var i %s\n\t\t", modelNameGo)
	res += "if err := rows.Scan("
	for i, col := range query.ResultColumns {
		colNameGo := snakeToCamel(col.Name, true)
		if i == 0 {
			res += fmt.Sprintf("&i.%s", colNameGo)
		} else {
			res += fmt.Sprintf(", &i.%s", colNameGo)
		}
	}
	res += "); err != nil {\n\t\t\t"
	res += fmt.Sprintf("return models.%s{}, err\n\t\t", modelNameGo)
	res += "}\n\t\t"
	res += "items = append(items, i)\n\t"
	res += "}\n\t"
	res += "if err := rows.Err(); err != nil {\n\t\t"
	res += "return nil, err\n\t"
	res += "}\n\t"
	res += "return items, nil"
	return res, nil
}

func functionBodyExec(query *ir.Query) (string, error) {
	if query.Cardinality != ir.Exec {
		return "", fmt.Errorf("(functionBodyExec expects cardinality ir.Exec, got %s", query.Cardinality)
	}

	res := ""
	res += fmt.Sprintf("_, err := db.ExecContext(ctx, %s", queryName(query))
	for _, param := range query.Parameters {
		res += fmt.Sprintf(", arg%d", param.Index)
	}
	res += ")\n\t"
	res += "return err"
	return res, nil
}
