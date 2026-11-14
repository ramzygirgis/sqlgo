package parsing

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strings"
)

type SelectStatement struct {
	ColumnNames []string
	From        *FromKeyword
	Where       Expr
	All         bool
}

func (s *SelectStatement) statementMarker() {}

func (s *SelectStatement) String() string {
	var sb strings.Builder
	lastIndex := len(s.ColumnNames) - 1
	sb.WriteString("SELECT")
	if s.All {
		sb.WriteString(" *")
		if lastIndex != -1 {
			sb.WriteString(",")
		}
	}
	for i, colName := range s.ColumnNames {
		sb.WriteString(fmt.Sprintf(" %s", colName))
		if i != lastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(s.From.String())
	if s.Where != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(s.Where.String())
	}
	sb.WriteString(";")
	return sb.String()
}

type FromKeyword struct {
	TableName string
}

func (s *FromKeyword) String() string {
	return fmt.Sprintf(" FROM %s", s.TableName)
}

func (p *parser) parseSelect() (*SelectStatement, error) {
	if _, err := p.expect(lexing.Keyword, "SELECT"); err != nil {
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
		ColumnNames: colNames,
		From:        from,
		All:         all,
		Where:       where,
	}, nil
}

func (p *parser) parseFrom() (*FromKeyword, error) {
	if _, err := p.expect(lexing.Keyword, "FROM"); err != nil {
		return &FromKeyword{}, err
	}
	if p.read().Kind != lexing.Identifier {
		return &FromKeyword{}, fmt.Errorf("expected table name identifier, got %v %s", p.read().Kind, p.read().Body)
	}
	table := p.next()
	from := FromKeyword{
		TableName: table.Body,
	}
	return &from, nil
}
