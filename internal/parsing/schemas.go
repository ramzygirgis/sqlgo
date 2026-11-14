package parsing

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/lexing"
)

func ParseSchema(query string) ([]Statement, error) {
	toks, err := lexing.Lex(query)
	if err != nil {
		return nil, err
	}
	if len(toks) == 0 {
		return nil, fmt.Errorf("token stream is empty")
	}
	p := parser{
		toks: toks,
		pos:  0,
	}
	var statements []Statement
	for p.read().Kind != lexing.EOF && p.read().Kind != lexing.OutOfBounds {
		statement, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	_, err = p.expectKind(lexing.EOF)
	if err != nil {
		return nil, err
	}
	return statements, nil
}
