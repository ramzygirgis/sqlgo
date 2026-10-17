package parsing

func Parse(query string) (Statement, error) {
	toks, err := Lex(query)
	p := parser{
		toks: toks,
		pos:  0,
	}
	statement, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	return statement, nil
}
