package parsing

import (
	"fmt"
)

type parser struct {
	toks []Token
	pos  int
}

func (p *parser) peek() Token {
	if p.pos >= len(p.toks) || p.pos < 0 {
		return Token{
			kind: EOF,
			body: "",
		}
	}
	return p.toks[p.pos]
}

func (p *parser) next() Token {
	tok := p.peek()
	p.pos += 1
	return tok
}

func (p *parser) expect(k TokenKind, body string) (Token, error) {
	if p.peek().kind != k || p.peek().body != body {
		return Token{}, fmt.Errorf("unexpected %v %s, got %v %s", k, body, p.peek().kind, p.peek().body)
	}
	return p.next(), nil
}

func (p *parser) parseSelect() (*SelectStatement, error) {
	if _, err := p.expect(Keyword, "SELECT"); err != nil {
		return nil, err
	}
	cols := make([]Column, 0)
	for p.peek().kind == Identifier {
		col := Column{
			name: p.next().body,
		}
		cols = append(cols, col)
	}
	from, err := p.parseFrom()
	if err != nil {
		return &SelectStatement{}, err
	}
	return &SelectStatement{
		columns: cols,
		from:    from,
	}, nil
}

func (p *parser) parseFrom() (*FromStatement, error) {}

func parseTokens(toks []Token) {

}
