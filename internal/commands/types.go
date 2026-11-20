package commands

import (
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/project"
)

type AnalyzedFile struct {
	SourcePath string
	Queries    []*ir.Query
}

type AnalyzedPackage struct {
	project.DirectoryInfo
	Files []AnalyzedFile
}

type OutputFile struct {
	Path     string
	Contents []byte
}
