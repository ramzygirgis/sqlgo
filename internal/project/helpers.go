package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

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

func ExtractModulePath(goModBody string) (string, error) {
	// maybe can optimize this, but prob super small returns
	lines := strings.Split(goModBody, "\n")
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
		//
		if strings.HasPrefix(modulePath, `"`) || strings.HasPrefix(modulePath, "`") {
			unquoted, err := strconv.Unquote(modulePath)
			if err != nil {
				return "", fmt.Errorf("go.mod: invalid module path %s: %w", modulePath, err)
			}
			modulePath = unquoted
		}
		return modulePath, nil
	}
	return "", fmt.Errorf("no module declaration found in go.mod")
}

// relative to project root -- something of the form:
// ["schema/users.sql", "schema/cars.sql"]
func collectSchemaFilepaths(entries []fs.DirEntry) ([]string, error) {
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
	if len(filenames) == 0 {
		return nil, fmt.Errorf("schema: no .sql files found (down migrations aren't included)")
	}
	return filenames, nil
}
