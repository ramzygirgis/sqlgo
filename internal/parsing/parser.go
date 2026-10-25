package parsing

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/lexing"
)

type parser struct {
	toks []lexing.Token
	pos  int
}

func (p *parser) peek() []lexing.Token {
	if p.pos >= len(p.toks) {
		return []lexing.Token{}
	}
	return p.toks[p.pos:]
}

func (p *parser) read() lexing.Token {
	if p.pos >= len(p.toks) || p.pos < 0 {
		// maybe we should treat this as an error
		return lexing.Token{
			Kind: lexing.EOF,
			Body: "EOF",
		}
	}
	return p.toks[p.pos]
}

func (p *parser) next() lexing.Token {
	tok := p.read()
	p.pos += 1
	return tok
}

func (p *parser) expect(k lexing.TokenKind, body string) (lexing.Token, error) {
	expected := lexing.Token{
		Kind: k,
		Body: body,
	}
	if p.read() != expected {
		return lexing.Token{}, fmt.Errorf("expected %v %s, got %v %s", k, body, p.read().Kind, p.read().Body)
	}
	return p.next(), nil
}

func (p *parser) expectKind(tk lexing.TokenKind) (string, error) {
	if p.read().Kind != tk {
		return "", fmt.Errorf("expected token kind %s, got %s", tk, p.read().Kind)
	}
	return p.next().Body, nil
}

func (p *parser) expectKeyword(body string) error {
	b, err := p.expectKind(lexing.Keyword)
	if err != nil {
		return err
	}
	if b != body {
		return fmt.Errorf("expected body %s, got %s.", body, b)
	}
	return nil
}

func (p *parser) acceptKeyword(body string) bool {
	tok := p.read()
	if tok.Kind != lexing.Keyword || tok.Body != body {
		return false
	}
	p.next()
	return true
}

func (p *parser) acceptComma() bool {
	tok := p.read()
	if tok.Kind != lexing.Comma || tok.Body != "," {
		return false
	}
	p.next()
	return true
}

func (p *parser) parseEnd() error {
	if _, err := p.expect(lexing.Semicolon, ";"); err != nil {
		return err
	}
	return nil
}

func Parse(query string) (Statement, error) {
	toks, err := lexing.Lex(query)
	p := parser{
		toks: toks,
		pos:  0,
	}
	statement, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	_, err = p.expectKind(lexing.EOF)
	if err != nil {
		return nil, err
	}
	return statement, nil
}
