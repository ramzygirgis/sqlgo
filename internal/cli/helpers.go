package cli

import (
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
			contents, err := os.ReadFile(sourcePath)
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
		Path:     modelsDir,
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
