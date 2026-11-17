package project

import (
	"io/fs"
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
