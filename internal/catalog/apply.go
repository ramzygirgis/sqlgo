package catalog

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func (c *Catalog) Apply(statement parsing.Statement) error {
	if statement == nil {
		return fmt.Errorf("nil pointer was passed to catalog.Apply")
	}
	switch s := statement.(type) {
	case *parsing.CreateTableStatement:
		if err := c.applyCreateTable(s); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid statement type %T", s)
	}
	return nil
}

func (c *Catalog) ApplyAll(statements []parsing.Statement) error {
	var err error
	for _, s := range statements {
		err = c.Apply(s)
		if err != nil {
			return err
		}
	}
	return nil
}
