package parsing

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strings"
)

type SelectStatement struct {
	ColumnNames []string
	TableName   string
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
	sb.WriteString(" FROM ")
	sb.WriteString(s.TableName)
	if s.Where != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(s.Where.String())
	}
	sb.WriteString(";")
	return sb.String()
}

func (p *parser) parseSelect() (*SelectStatement, error) {
	if err := p.expectKeyword("SELECT"); err != nil {
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

	_, err = p.expect(lexing.Keyword, "FROM")
	if err != nil {
		return nil, err
	}

	tableName, err := p.expectKind(lexing.Identifier)
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
		TableName:   tableName,
		All:         all,
		Where:       where,
	}, nil
}
