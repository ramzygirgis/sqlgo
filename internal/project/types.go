package project

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
