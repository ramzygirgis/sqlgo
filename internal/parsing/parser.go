package parsing

func (p *parser) parseSelect() (*SelectStatement, error) {
	if _, err := p.expectKeyword("SELECT"); err != nil {
		return nil, err
	}
	var col Column
	cols = make([]Column, 0)
	while p.peek().kind == Identifier {
		cols := Column{
			kind: ""
		}
		append(cols, )
	}
}

func parseTokens(toks []Token) {
	
}
