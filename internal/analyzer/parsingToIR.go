package analyzer

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func toIRCardinality(c parsing.Cardinality) (ir.Cardinality, error) {
	switch c {
	case parsing.One:
		return ir.One, nil
	case parsing.Many:
		return ir.Many, nil
	case parsing.Exec:
		return ir.Exec, nil
	default:
		return ir.IllegalCardinality, fmt.Errorf("unexpected cardinality %s", c)
	}
}
