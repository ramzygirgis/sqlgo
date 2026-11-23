package parsing

import (
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strings"
)

type DeleteFromStatement struct {
	TableName string
	Where     Expr
}

func (d *DeleteFromStatement) statementMarker() {}

func (d *DeleteFromStatement) String() string {
	var sb strings.Builder

	sb.WriteString("DELETE FROM ")
	sb.WriteString(d.TableName)
	if d.Where != nil {
		sb.WriteString("\nWHERE ")
		sb.WriteString(d.Where.String())
	}
	sb.WriteString(";")
	return sb.String()
}

func (p *parser) parseDelete() (*DeleteFromStatement, error) {
	if err := p.expectKeyword("DELETE"); err != nil {
		return nil, err
	}

	if err := p.expectKeyword("FROM"); err != nil {
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

	return &DeleteFromStatement{
		TableName: tableName,
		Where:     where,
	}, nil
}
