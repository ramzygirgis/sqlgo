package project

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
)

func getModulePath(fsys fs.FS) (string, error) {
	data, err := fs.ReadFile(fsys, "go.mod")
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("no go.mod found")
	}
	if err != nil {
		return "", err
	}

	body := string(data)
	return ExtractModulePath(body)
}

func findSchemaFiles(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, "schema")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("no schema directory found")
	}
	if err != nil {
		return nil, err
	}
	return collectSchemaFilepaths(entries)
}

func modelsInfo(modulePath string) DirectoryInfo {
	return DirectoryInfo{
		Dir:        modelsDir,
		ImportPath: path.Join(modulePath, modelsDir),
	}
}

func findQueryPackages(fsys fs.FS, modulePath string) ([]Package, error) {
	var packages []Package
	err := fs.WalkDir(fsys, ".", func(curPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		if curPath != "." {
			if d.Name() == "vendor" || d.Name() == "testdata" {
				return fs.SkipDir
			}
			if strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_") {
				return fs.SkipDir
			}
			if curPath == "schema" {
				return fs.SkipDir
			}
			// if you have a go.mod, you're a separate module
			hasGoMod, err := directoryHasFile(fsys, curPath, "go.mod")
			if err != nil {
				return err
			}
			if hasGoMod {
				return fs.SkipDir
			}
		}

		matches, err := collectMatchingSuffixFiles(fsys, curPath, ".sql")
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return nil
		}

		if curPath == "." {
			return fmt.Errorf("%s: .sql files at the project root aren't supported; move them into a package directory",
				strings.Join(matches, ", "))
		}
		if curPath == modelsDir {
			return fmt.Errorf("%s: .sql files aren't allowed in %s, since that package is generated",
				strings.Join(matches, ", "), modelsDir)
		}

		thisPackage := Package{
			Dir:        curPath,
			ImportPath: path.Join(modulePath, curPath),
			SQLFiles:   matches,
		}
		packages = append(packages, thisPackage)
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(packages, func(a, b Package) int {
		return strings.Compare(a.Dir, b.Dir)
	})
	return packages, nil
}
