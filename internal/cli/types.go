package cli

import (
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/project"
)

type analyzedFile struct {
	SourcePath string
	Queries    []*ir.Query
}

type analyzedPackage struct {
	project.Package
	Files []analyzedFile
}
