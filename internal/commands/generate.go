package commands

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/analyzer"
	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/codegen"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
	"github.com/ramzygirgis/sqlgo/internal/project"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

func findProject() (string, fs.FS, project.Project, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, project.Project{}, err
	}
	root, err := project.FindRoot(wd)
	if err != nil {
		return "", nil, project.Project{}, err
	}
	fsys := os.DirFS(root)
	proj, err := project.Discover(fsys)
	if err != nil {
		return "", nil, project.Project{}, err
	}
	return root, fsys, proj, nil
}

func loadCatalog(fsys fs.FS, proj project.Project) (*catalog.Catalog, error) {
	c := catalog.New()
	for _, schemaPath := range proj.SchemaFiles {
		content, err := fs.ReadFile(fsys, schemaPath)
		if err != nil {
			return nil, err
		}
		stmts, err := parsing.ParseSchema(string(content))
		if err != nil {
			return nil, err
		}
		// this is why it's important that SchemaFiles are in order
		for _, stmt := range stmts {
			err = c.Apply(stmt)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", schemaPath, err)
			}
		}
	}
	return c, nil
}

func analyzePackages(fsys fs.FS, proj project.Project, c *catalog.Catalog) ([]AnalyzedPackage, error) {
	var analyzedPackages []AnalyzedPackage
	var errs []error
	for _, queryPackage := range proj.QueryPackages {
		var analyzedFiles []AnalyzedFile
		queryNames := make(map[string]struct{})
		for _, file := range queryPackage.SQLFiles {
			// todo: make sure that file is the relative path wrt root
			sourcePath := path.Join(queryPackage.Dir, file)
			contents, err := fs.ReadFile(fsys, sourcePath)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s %s", sourcePath, err))
				continue
			}
			queries, err := parsing.ParseQueries(string(contents))
			if err != nil {
				errs = append(errs, fmt.Errorf("%s %s", sourcePath, err))
				continue
			}
			var irQueries []*ir.Query
			for _, q := range queries {
				if _, ok := queryNames[q.Name]; ok {
					errs = append(errs, fmt.Errorf("%s: query %s appears multiple times", sourcePath, q.Name))
					continue
				}
				queryNames[q.Name] = struct{}{}

				irQuery, err := analyzer.AnalyzeQuery(c, q)
				if err != nil {
					errs = append(errs, fmt.Errorf("%s %s", sourcePath, err))
					continue
				}
				irQueries = append(irQueries, irQuery)
			}
			// generate file with same name as `file` but with a .go after
			analyzedFile := AnalyzedFile{
				SourcePath: sourcePath,
				Queries:    irQueries,
			}
			analyzedFiles = append(analyzedFiles, analyzedFile)
		}
		analyzedPackage := AnalyzedPackage{
			Dir:        queryPackage.Dir,
			ImportPath: queryPackage.ImportPath,
			Files:      analyzedFiles,
		}
		analyzedPackages = append(analyzedPackages, analyzedPackage)
	}
	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}
	return analyzedPackages, nil
}

func render(proj project.Project, c *catalog.Catalog, analyzedPackages []AnalyzedPackage) ([]OutputFile, error) {
	var outputs []OutputFile

	modelsImportPath := proj.ModelsInfo.ImportPath
	modelsDir := proj.ModelsInfo.Dir
	// gen models
	modelsContents, err := codegen.GenerateModelsFile(*c, modelsImportPath)
	if err != nil {
		return nil, fmt.Errorf("models.go: %s", err)
	}
	modelsFile := OutputFile{
		Path:     path.Join(modelsDir, "models.go"),
		Contents: []byte(modelsContents),
	}
	outputs = append(outputs, modelsFile)

	dbContents, err := codegen.GenerateDbFile(modelsImportPath)
	if err != nil {
		return nil, fmt.Errorf("db.go: %s", err)
	}
	dbFile := OutputFile{
		Path:     path.Join(modelsDir, "db.go"),
		Contents: []byte(dbContents),
	}
	outputs = append(outputs, dbFile)

	for _, analyzedPackage := range analyzedPackages {
		packageImportPath := analyzedPackage.ImportPath
		files := analyzedPackage.Files
		for _, file := range files {
			if len(file.Queries) == 0 {
				continue
			}
			queryContents, err := codegen.GenerateQueriesFile(file.Queries, packageImportPath, modelsImportPath)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", file.SourcePath, err)
			}
			queryFile := OutputFile{
				Path:     file.SourcePath + ".go",
				Contents: []byte(queryContents),
			}
			outputs = append(outputs, queryFile)
		}
	}
	return outputs, nil
}

func findStale(fsys fs.FS, outputs []OutputFile) ([]string, error) {
	keep := make(map[string]struct{})
	for _, out := range outputs {
		keep[out.Path] = struct{}{}
	}

	var stale []string
	// watch out -- calling with "."
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if p == "." {
				return nil // we shouldn't skip root
			}
			name := d.Name()
			if name == "vendor" || name == "testdata" ||
				strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				return fs.SkipDir
			}
			// we shouldn't clean another module
			if _, err := fs.Stat(fsys, path.Join(p, "go.mod")); err == nil {
				return fs.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(p, ".go") {
			return nil
		}
		if _, ok := keep[p]; ok {
			return nil
		}

		file, err := fsys.Open(p)
		if err != nil {
			return err
		}
		defer file.Close()

		firstLine := ""
		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			firstLine = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		if firstLine != strings.TrimSpace(codegen.GetCodegenHeader()) {
			return nil
		}

		stale = append(stale, p)
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.Sort(stale)
	return stale, nil
}

func write(root string, outputs []OutputFile, stale []string) error {
	for _, out := range outputs {
		full := filepath.Join(root, filepath.FromSlash(out.Path))

		err := os.MkdirAll(filepath.Dir(full), 0o755)
		if err != nil {
			return fmt.Errorf("%s: %s", out.Path, err)
		}

		shouldWrite, err := shouldWriteFile(full, out)
		if err != nil {
			return err
		}
		if shouldWrite {
			if err = writeFileAtomic(full, string(out.Contents), ".sqlgo-*"); err != nil {
				return err
			}
		}
	}

	for _, stale := range stale {
		stalePath := filepath.Join(root, filepath.FromSlash(stale))
		shouldDelete, err := shouldDeleteStale(stalePath)
		if err != nil {
			// remove all generated files??
			return err
		}
		if shouldDelete {
			err = os.Remove(stalePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
