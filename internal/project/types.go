package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const modelsDir = "internal/models"

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

func directoryHasFile(fsys fs.FS, curPath string, targetFileName string) (bool, error) {
	entries, err := fs.ReadDir(fsys, curPath)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.Name() == targetFileName {
			return true, nil
		}
	}
	return false, nil
}

func collectMatchingSuffixFiles(fsys fs.FS, curPath string, targetFileEnding string) ([]string, error) {
	var matches []string
	entries, err := fs.ReadDir(fsys, curPath)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), targetFileEnding) {
			matches = append(matches, entry.Name())
		}
	}
	slices.Sort(matches)
	return matches, nil
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

func FindRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		info, err := os.Stat(filepath.Join(dir, "go.mod"))
		switch {
		case err == nil && !info.IsDir():
			return dir, nil
		case err != nil && !errors.Is(err, fs.ErrNotExist):
			return "", fmt.Errorf("checking %s: %w", dir, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found in %s or any parent directory", startDir)
		}
		dir = parent
	}
}

func modelsInfo(modulePath string) DirectoryInfo {
	return DirectoryInfo{
		Dir:        modelsDir,
		ImportPath: path.Join(modulePath, modelsDir),
	}
}
