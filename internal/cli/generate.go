package cli

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/analyzer"
	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
	"github.com/ramzygirgis/sqlgo/internal/project"
	"io/fs"
	"os"
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
		// we assume that content contains 1 query and no annotations
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
	// note: we should collect errors instead of early exiting when seeing one, so user gets multiple at once
	for _, queryPackage := range proj.QueryPackages {
		var analyzedFiles []AnalyzedFile
		for _, file := range queryPackage.SQLFiles {
			// todo: make sure that file is the relative path wrt root
			contents, err := os.ReadFile(file)
			if err != nil {
				return nil, err
			}
			queries, err := parsing.ParseQueries(string(contents))
			if err != nil {
				return nil, err
			}
			queryNames := make(map[string]struct{})
			var irQueries []*ir.Query
			for _, q := range queries {
				if _, ok := queryNames[q.Name]; ok {
					return nil, fmt.Errorf("query %s appears multiple times", q.Name)
				}
				queryNames[q.Name] = struct{}{}

				irQuery, err := analyzer.AnalyzeQuery(c, q)
				if err != nil {
					return nil, err
				}
				irQueries = append(irQueries, irQuery)
			}
			// generate file with same name as `file` but with a .go after
			analyzedFile := AnalyzedFile{
				SourcePath: file, // is this correct? maybe we should append a .go?
				Queries:    irQueries,
			}
			analyzedFiles = append(analyzedFiles, analyzedFile)
		}
		analyzedPackage := AnalyzedPackage{
			Dir:        queryPackage.Dir,
			ImportPath: queryPackage.ImportPath,
			SQLFiles:   queryPackage.SQLFiles,
			Files:      analyzedFiles,
		}
		analyzedPackages = append(analyzedPackages, analyzedPackage)
	}
	return analyzedPackages, nil
}

func Generate() error {
	root, fsys, proj, err := findProject()
	if err != nil {
		return err
	}

	cat, err := loadCatalog(fsys, proj)
	if err != nil {
		return err
	}

	analyzedPackages, err := analyzePackages(fsys, proj, cat)
	if err != nil {
		return err
	}

	stale, err := findStale(fsys, analyzedPackages) // generated files no longer produced
	if err != nil {
		return err
	}

	// generate models.go, using proj.ModelsInfo
	// generate db.go, just in same dir as models
	return write(root, analyzedPackages, stale)
}
