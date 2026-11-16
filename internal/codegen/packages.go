package codegen

import (
	"fmt"
	"path"
)

func getPackageDeclaration(packageImportPath string) string {
	packageName := path.Base(packageImportPath)
	return fmt.Sprintf("package %s", packageName)
}
