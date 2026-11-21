package project

import (
	"io/fs"
	"strings"
)

// for query package search
func shouldSkipDir(fsys fs.FS, curPath string, d fs.DirEntry) (bool, error) {
	if curPath == "." {
		// i think this is correct
		return false, nil
	}

	if shouldIgnoreDirName(d.Name()) {
		return true, nil
	}

	if shouldIgnorePath(d.Name()) {
		return true, nil
	}

	isDiffModule, err := isSeperateModule(fsys, curPath)
	if err != nil {
		return false, err
	}
	if isDiffModule {
		return true, nil
	}

	return false, nil
}

func shouldIgnoreDirName(dirName string) bool {
	ignoreSet := make(map[string]struct{})
	ignoreSet["vendor"] = struct{}{}
	ignoreSet["testdata"] = struct{}{}

	if _, ok := ignoreSet[dirName]; ok {
		return true
	}

	ignoredPrefixes := []string{".", "_"}
	for _, ignoredPrefix := range ignoredPrefixes {
		if strings.HasPrefix(dirName, ignoredPrefix) {
			return true
		}
	}

	return false
}

func shouldIgnorePath(curPath string) bool {
	// this function assumes that your path is NOT "."
	ignoreSet := make(map[string]struct{})
	ignoreSet["schema"] = struct{}{}

	_, ok := ignoreSet[curPath]
	return ok
}

func isSeperateModule(fsys fs.FS, curPath string) (bool, error) {
	hasGoMod, err := directoryHasFile(fsys, curPath, "go.mod")
	if err != nil {
		return false, err
	}
	if hasGoMod {
		return true, nil
	}
	return false, nil
}
