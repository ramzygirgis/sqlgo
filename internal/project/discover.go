package project

import (
	"io/fs"
)

func Discover(fsys fs.FS, modulePath string) (Project, error) {
	modulePath, err := getModulePath(fsys)
	if err != nil {
		return Project{}, err
	}
	schemaFiles, err := findSchemaFiles(fsys)
	if err != nil {
		return Project{}, err
	}
	modelsInfo := modelsInfo(modulePath)
	queryPackages, err := findQueryPackages(fsys, modulePath)
	if err != nil {
		return Project{}, err
	}
	return Project{
		ModulePath:    modulePath,
		SchemaFiles:   schemaFiles,
		ModelsInfo:    modelsInfo,
		QueryPackages: queryPackages,
	}, nil
}
