package parsing

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/lexing"
)

type Statement interface {
	statementMarker()
	String() string // for testing
}

func (p *parser) parseStatement() (Statement, error) {
	toks := p.peek()
	if len(toks) == 0 {
		return nil, fmt.Errorf("empty query")
	}
	if toks[0].Kind != lexing.Keyword {
		return nil, fmt.Errorf("expected query to begin with a statement")
	}
	switch toks[0].Body {
	case "SELECT":
		selectStatementPtr, err := p.parseSelect()
		if err != nil {
			return nil, err
		}
		return selectStatementPtr, nil
	case "CREATE":
		if len(toks) == 1 {
			return nil, fmt.Errorf("this query is ass")
		}
		switch toks[1].Body {
		case "TABLE":
			createTableStatementPtr, err := p.parseCreateTable()
			if err != nil {
				return nil, err
			}
			return createTableStatementPtr, nil
		default:
			return nil, fmt.Errorf("expected CREATE keyword to be followed by another keyword.")
		}
	default:
		return nil, fmt.Errorf("expected query to begin with a statement")
	}
}
