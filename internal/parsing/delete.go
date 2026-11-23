package parsing

import (
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strings"
)

type DeleteStatement struct {
	TableName string
	Where     Expr
}

func (d *DeleteStatement) statementMarker() {}

func (d *DeleteStatement) String() string {
	var sb strings.Builder

	sb.WriteString("DELETE ")
	sb.WriteString(d.TableName)
	sb.WriteString("\nWHERE ")
	if d.Where != nil {
		sb.WriteString(d.Where.String())
	}
	sb.WriteString(";")
	return sb.String()
}

func (p *parser) parseDelete() (*DeleteStatement, error) {
	if err := p.expectKeyword("DELETE"); err != nil {
		return nil, err
	}

	tableName, err := p.expectKind(lexing.Identifier)
	if err != nil {
		return nil, err
	}

	var where Expr = nil
	if p.acceptKeyword("WHERE") {
		where, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expectKind(lexing.Semicolon); err != nil {
		return nil, err
	}

	return &DeleteStatement{
		TableName: tableName,
		Where:     where,
	}, nil
}
