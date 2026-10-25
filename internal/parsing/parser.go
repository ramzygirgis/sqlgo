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
		return Token{}, fmt.Errorf("expected %v %s, got %v %s", k, body, p.read().kind, p.read().body)
	}
	return p.next(), nil
}

func (p *parser) expectKind(tk TokenKind) (string, error) {
	if p.read().kind != tk {
		return "", fmt.Errorf("expected token kind %s, got %s", tk, p.read().kind)
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

func (p *parser) acceptKeyword(body string) bool {
	tok := p.read()
	if tok.kind != Keyword || tok.body != body {
		return false
	}
	p.next()
	return true
}

func (p *parser) acceptComma() bool {
	tok := p.read()
	if tok.kind != Comma || tok.body != ";" {
		return false
	}
	p.next()
	return true
}

func (p *parser) Parse() (Statement, error) {
	s, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	_, err = p.expectKind(EOF)
	if err != nil {
		return nil, err
	}
	return s, nil
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
	if _, err := p.expect(LParen, "("); err != nil {
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
	if _, err = p.expect(RParen, ")"); err != nil {
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
	sqlType, err := p.expectKind(Type)
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
	var where Expr = nil
	colNames := make([]string, 0)
	all := false

	colName, isStar, err := p.parseColumnRef()
	if err != nil {
		return nil, err
	}
	if isStar {
		all = true
	} else {
		colNames = append(colNames, colName)
	}

	for p.acceptComma() {
		colName, isStar, err = p.parseColumnRef()
		if err != nil {
			return nil, err
		}
		if isStar {
			all = true
		} else {
			colNames = append(colNames, colName)
		}
	}

	from, err := p.parseFrom()
	if err != nil {
		return &SelectStatement{}, err
	}

	if p.acceptKeyword("WHERE") {
		where, err = p.parseExpr()
		if err != nil {
			return &SelectStatement{}, err
		}
	}

	if err := p.parseEnd(); err != nil {
		return &SelectStatement{}, err
	}

	return &SelectStatement{
		columnNames: colNames,
		from:        from,
		all:         all,
		where:       where,
	}, nil
}

func (p *parser) parseColumnRef() (string, bool, error) {
	switch p.read().kind {
	case Identifier:
		if len(p.read().body) == 0 {
			return "", false, fmt.Errorf("expected column name identifier, got empty string")
		}
		return p.next().body, false, nil
	case Star:
		if _, err := p.expect(Star, "*"); err != nil {
			return "", false, fmt.Errorf("expected *, got %s %s", p.read().kind, p.read().body)
		}
		return "", true, nil
	default:
		return "", false, fmt.Errorf("expected column name or *, got %s %s", p.read().kind, p.read().body)
	}
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

func (p *parser) parseExpr() (Expr, error) {
	return p.parseOrExpr()
}

func (p *parser) parseOrExpr() (Expr, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}
	for p.acceptKeyword("OR") {
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{op: "OR", left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseAndExpr() (Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.acceptKeyword("AND") {
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{op: "AND", left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseComparison() (Expr, error) {
	left, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	if p.read().kind != Comparator {
		return left, nil
	}
	op := p.next().body

	right, err := p.parseOperand()
	if err != nil {
		return nil, err
	}

	return &BinaryExpr{op: op, left: left, right: right}, nil
}

func (p *parser) parseOperand() (Expr, error) {
	switch p.read().kind {
	case Identifier:
		return &ColumnRef{p.next().body}, nil
	case String:
		return &Literal{p.next().body}, nil
	case Number:
		return &Literal{p.next().body}, nil
	case Param:
		node, err := unpackParameter(p.next().body)
		if err != nil {
			return nil, err
		}
		return node, nil
	case LParen:
		p.next()
		node, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expectKind(RParen)
		if err != nil {
			return nil, err
		}
		return node, nil
	default:
		return nil, fmt.Errorf("Expected Identifier, String, Number, Parameter or (, got %s %s", p.read().kind, p.read().body)
	}
}
