package generator

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"path"
	"strings"
)

func codegen(query *ir.Query, packageImportPath string, modelsImportPath string) string {
	switch query.Cardinality {
	case ir.One:

	case ir.Many:
	case ir.Exec:
	}

}

func codegenOne() string {

}

func codegenMany() string {

}

func codegenExec() string {

}

func queryName(query *ir.Query) string {
	// adding \n\n at the end is the parent's job
	return strings.ToUpper(string(query.Name[0])) + query.Name[1:]
}

func functionContract(query *ir.Query, packageImportPath string, modelsImportPath string) (string, error) {
	// this guy writes func name(...) (... , ... , ...), but not the {.
	// that, and the } is the job of the parent
	// later, add support for smart naming function arguments
	res := fmt.Sprintf("func %s(ctx context.Context, db models.DBTX", query.Name)
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
	returnTypeString := ""
	switch query.Cardinality {
	case ir.One:
		returnTypeString += fmt.Sprintf("%s, error)", modelType.In(packageImportPath))
	case ir.Many:
		returnTypeString += fmt.Sprintf("(%s, error)", modelType.In(packageImportPath))
	case ir.Exec:
		returnTypeString += "error"
	default:
		return "", fmt.Errorf("unsupported cardinality %s", query.Cardinality)
	}
	res += returnTypeString
	return res, nil
}
