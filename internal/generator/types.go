package generator

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/ir"
)

type GoType struct {
	TypeName    string
	PackageName string
	ImportPath  string
}

func (t GoType) String() string {
	if t.PackageName == "" {
		return t.TypeName
	}
	return t.PackageName + "." + t.TypeName
}

func (t GoType) In(currentImportPath string) string {
	if t.ImportPath == "" || t.ImportPath == currentImportPath {
		return t.TypeName
	}
	return t.PackageName + "." + t.TypeName
}

func IRtoGoType(irType ir.SQLType, nullable bool) (GoType, error) {
	switch irType {
	case ir.TypeText:
		if nullable {
			return GoType{
				TypeName:    "NullString",
				PackageName: "sql",
				ImportPath:  "database/sql",
			}, nil
		} else {
			return GoType{
				TypeName:    "string",
				PackageName: "",
				ImportPath:  "",
			}, nil
		}
	case ir.TypeInt:
		if nullable {
			return GoType{
				TypeName:    "NullInt32",
				PackageName: "sql",
				ImportPath:  "database/sql",
			}, nil
		} else {
			return GoType{
				TypeName:    "int32",
				PackageName: "",
				ImportPath:  "",
			}, nil
		}
	default:
		return GoType{}, fmt.Errorf("Invalid SQLType %s", irType)
	}
}
