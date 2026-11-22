package parsing

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strings"
)

/*
UPDATE table_name
SET column1 = value1, column2 = value2
WHERE condition;
*/

type ColumnValuePair struct {
	ColumnRef ColumnRef
	Value     Expr
}

type UpdateStatement struct {
	TableName string
	CvPairs   []ColumnValuePair
	Where     Expr
}

func (u *UpdateStatement) statementMarker() {}

func (u *UpdateStatement) String() string {
	var sb strings.Builder

	sb.WriteString("UPDATE ")
	sb.WriteString(u.TableName)
	sb.WriteString("\nSET ")
	for i, cvPair := range u.CvPairs {
		sb.WriteString(fmt.Sprintf("%s = %s", cvPair.ColumnRef.Name, cvPair.Value))
		if i != len(u.CvPairs) {
			sb.WriteString(", ")
		}
	}
	if u.Where != nil {
		sb.WriteString("\nWHERE ")
		sb.WriteString(u.Where.String())
	}
	sb.WriteString(";")
	return sb.String()
}

func (p *parser) parseUpdate() (*UpdateStatement, error) {
	if err := p.expectKeyword("UPDATE"); err != nil {
		return nil, err
	}
	tableName, err := p.expectKind(lexing.Identifier)

	if err != nil {
		return nil, err
	}
	if err := p.expectKeyword("SET"); err != nil {
		return nil, err
	}

	var cvPairs []ColumnValuePair
	cvPair, err := p.parseColumnValuePair()
	if err != nil {
		return nil, err
	}
	cvPairs = append(cvPairs, cvPair)
	for p.acceptComma() {
		cvPair, err := p.parseColumnValuePair()
		if err != nil {
			return nil, err
		}
		cvPairs = append(cvPairs, cvPair)
	}

	var where Expr = nil
	if p.acceptKeyword("WHERE") {
		where, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &UpdateStatement{
		TableName: tableName,
		CvPairs:   cvPairs,
		Where:     where,
	}, nil
}

func (p *parser) parseColumnValuePair() (ColumnValuePair, error) {
	colName, err := p.expectKind(lexing.Identifier)
	if err != nil {
		return ColumnValuePair{}, err
	}
	equals, err := p.expectKind(lexing.Comparator)
	if err != nil {
		return ColumnValuePair{}, err
	}
	if equals != "=" {
		return ColumnValuePair{}, fmt.Errorf("expected %s = <value>, got %s %s <value>.", colName, equals, colName)
	}

	v, err := p.parseExpr()
	if err != nil {
		return ColumnValuePair{}, err
	}

	return ColumnValuePair{
		ColumnRef: ColumnRef{
			Name: colName,
		},
		Value: v,
	}, nil
}
