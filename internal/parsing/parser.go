package parsing

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
