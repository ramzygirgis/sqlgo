package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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
