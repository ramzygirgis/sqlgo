package parsing

import (
	"fmt"
)

type parser struct {
	toks []Token
	pos  int
}

func (p *parser) peek() []Token {
	return p.toks[p.pos:]
}

func (p *parser) read() Token {
	if p.pos >= len(p.toks) || p.pos < 0 {
		return Token{
			kind: EOF,
			body: "EOF",
		}
	}
	return p.toks[p.pos]
}

func (p *parser) next() Token {
	tok := p.read()
	p.pos += 1
	return tok
}

func (p *parser) expect(k TokenKind, body string) (Token, error) {
	expected := Token{
		kind: k,
		body: body,
	}
	if p.read() != expected {
		return Token{}, fmt.Errorf("unexpected %v %s, got %v %s", k, body, p.read().kind, p.read().body)
	}
	return p.next(), nil
}

func (p *parser) parseStatement() (Statement, error) {
	toks := p.peek()
	if len(toks) == 0 {
		return nil, fmt.Errorf("empty query")
	}
	if toks[0].kind != Keyword {
		return nil, fmt.Errorf("expected query to begin with a statement")
	}
	switch toks[0].body {
	case "SELECT":
		selectStatementPtr, err := p.parseSelect()
		if err != nil {
			return nil, err
		}
		return selectStatementPtr, nil
	default:
		return nil, fmt.Errorf("expected query to begin with a statement")
	}
}

func (p *parser) parseSelect() (*SelectStatement, error) {
	if _, err := p.expect(Keyword, "SELECT"); err != nil {
		return nil, err
	}
	cols := make([]Column, 0)
	for p.read().kind == Identifier {
		col := Column{
			name: p.next().body,
		}
		cols = append(cols, col)
	}

	from, err := p.parseFrom()
	if err != nil {
		return &SelectStatement{}, err
	}

	if err := p.parseEnd(); err != nil {
		return &SelectStatement{}, err
	}

	return &SelectStatement{
		columns: cols,
		from:    from,
	}, nil
}

func (p *parser) parseFrom() (*FromStatement, error) {
	if _, err := p.expect(Keyword, "FROM"); err != nil {
		return &FromStatement{}, err
	}
	if p.read().kind != Identifier {
		return &FromStatement{}, fmt.Errorf("expected table name identifier, got %v %s", p.read().kind, p.read().body)
	}
	table := p.next()
	from := FromStatement{
		table: Table{
			name: table.body,
		},
	}

	return &from, nil
}

func (p *parser) parseEnd() error {
	if _, err := p.expect(Semicolon, ";"); err != nil {
		return err
	}
	return nil
}
