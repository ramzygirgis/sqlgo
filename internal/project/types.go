package project

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

type DirectoryInfo struct {
	Dir        string
	ImportPath string
}

type Package struct {
	DirectoryInfo
	SQLFiles []string
}

type Project struct {
	ModulePath    string
	SchemaFiles   []string // schema files, in order of hwo to apply (schema/001_..., schema/002_...)
	ModelsInfo    DirectoryInfo
	QueryPackages []Package
}

/*
* module path:
* we should find go.mod. once we find go.mod, we should serach it for the quoted path (module "github.com/...")
* once you get modulePath, you look for a schema subdirectory. inside, you'll read every .sql file, and store them in sorted order
*	-- skip *.down.sql files. if you can't find a schema dir, or can't find any valid .sql files, then it's an error
*
 */

func getModulePath(fsys fs.FS) (string, error) {
	data, err := fs.ReadFile(fsys, "go.mod")
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("no go.mod found")
	}
	if err != nil {
		return "", err
	}

	body := string(data)
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "module" {
			continue
		}
		if len(fields) <= 1 {
			return "", fmt.Errorf("no module path found in module declaration")
		}
		modulePath := fields[1]
		// potentially have some validation for modulePath
		return modulePath, nil
	}
	return "", fmt.Errorf("no module declaration found in go.mod")
}

func findSchemaFiles(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, "schema")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("no schema directory found")
	}
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") || strings.HasSuffix(name, ".down.sql") {
			continue
		}
		full := path.Join("schema", name)
		filenames = append(filenames, full)
	}
	return filenames, nil
}
