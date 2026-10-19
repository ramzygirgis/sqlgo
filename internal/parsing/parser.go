package parsing

import (
	"fmt"
)

type parser struct {
	toks []Token
	pos  int
}

func (p *parser) peek() []Token {
	if p.pos >= len(p.toks) {
		return []Token{}
	}
	return p.toks[p.pos:]
}

func (p *parser) read() Token {
	if p.pos >= len(p.toks) || p.pos < 0 {
		// maybe we can throw here
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

func (p *parser) expectKind(tk TokenKind) (string, error) {
	if p.read().kind != tk {
		return "", fmt.Errorf("unexpected token kind %s, got %s", tk, p.read().kind)
	}
	return p.next().body, nil
}

func (p *parser) expectKeyword(body string) error {
	b, err := p.expectKind(Keyword)
	if err != nil {
		return err
	}
	if b != body {
		return fmt.Errorf("expected body %s, got %s.", body, b)
	}
	return nil
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
	case "CREATE":
		if len(toks) == 1 {
			return nil, fmt.Errorf("this query is ass")
		}
		switch toks[1].body {
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

func (p *parser) parseCreateTable() (*CreateTableStatement, error) {
	ifNotExists := false
	if _, err := p.expect(Keyword, "CREATE"); err != nil {
		return nil, err
	}
	if _, err := p.expect(Keyword, "TABLE"); err != nil {
		return nil, err
	}
	if (p.read() == Token{
		kind: Keyword,
		body: "IF",
	}) {
		p.next()
		if err := p.expectKeyword("NOT"); err != nil {
			return nil, err
		}
		if err := p.expectKeyword("EXISTS"); err != nil {
			return nil, err
		}
		ifNotExists = true
	}
	tableName, err := p.expectKind(Identifier)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(Identifier, "("); err != nil {
		return nil, err
	}
	var cols []Column
	col, err := p.parseColumnDef()
	if err != nil {
		return nil, err
	}
	cols = append(cols, col)
	for p.read().kind == Comma {
		p.next()
		col, err := p.parseColumnDef()
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	if _, err = p.expect(Identifier, ")"); err != nil {
		return nil, err
	}
	if _, err := p.expect(Semicolon, ";"); err != nil {
		return nil, err
	}
	return &CreateTableStatement{
		tableName:   tableName,
		columns:     cols,
		ifNotExists: ifNotExists,
	}, err
}

func (p *parser) parseColumnDef() (Column, error) {
	name, err := p.expectKind(Identifier)
	nullable := true
	primaryKey := false
	hasDefault := false
	defaultVal := ""
	if err != nil {
		return Column{}, err
	}
	sqlType, err := p.expectKind(Identifier)
	if err != nil {
		return Column{}, err
	}
constraints:
	for p.read().kind != EOF { // find cleanest way to do this
		switch p.read() {
		case Token{
			kind: Keyword,
			body: "NOT",
		}:
			p.next()
			err := p.expectKeyword("NULL")
			if err != nil {
				return Column{}, err
			}
			nullable = false
		case Token{
			kind: Keyword,
			body: "PRIMARY",
		}:
			p.next()
			err := p.expectKeyword("KEY")
			if err != nil {
				return Column{}, err
			}
			primaryKey = true
		case Token{
			kind: Keyword,
			body: "DEFAULT",
		}:
			p.next()
			body, err := p.expectKind(Identifier)
			if err != nil {
				return Column{}, err
			}
			hasDefault = true
			defaultVal = body
		default:
			break constraints
		}
	}
	return Column{
		name:       name,
		sqlType:    sqlType,
		hasDefault: hasDefault,
		defaultVal: defaultVal,
		nullable:   nullable,
		primaryKey: primaryKey,
	}, nil
}

func (p *parser) parseSelect() (*SelectStatement, error) {
	if _, err := p.expect(Keyword, "SELECT"); err != nil {
		return nil, err
	}
	cols := make([]Column, 0)
	all := false
	for p.read().kind == Identifier || p.read().kind == Star {
		switch p.read().kind {
		case Identifier:
			col := Column{
				name: p.next().body,
			}
			cols = append(cols, col)
		case Star:
			if _, err := p.expect(Star, "*"); err != nil {
				return nil, fmt.Errorf("expected *, got %s %s", p.read().kind, p.read().body)
			}
			all = true
		default:
			return nil, fmt.Errorf("expected column name or *, got %s %s", p.read().kind, p.read().body)
		}
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
		all:     all,
	}, nil
}

func (p *parser) parseFrom() (*FromKeyword, error) {
	if _, err := p.expect(Keyword, "FROM"); err != nil {
		return &FromKeyword{}, err
	}
	if p.read().kind != Identifier {
		return &FromKeyword{}, fmt.Errorf("expected table name identifier, got %v %s", p.read().kind, p.read().body)
	}
	table := p.next()
	from := FromKeyword{
		tableName: table.body,
	}
	return &from, nil
}

func (p *parser) parseEnd() error {
	if _, err := p.expect(Semicolon, ";"); err != nil {
		return err
	}
	return nil
}
