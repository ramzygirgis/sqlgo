package parsing

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strconv"
)

type Expr interface {
	expressionMarker()
	String() string // also for testing
}

type BinaryExpr struct {
	op    string
	left  Expr
	right Expr
}

func (b *BinaryExpr) expressionMarker() {}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.left.String(), b.op, b.right.String())
}

func (p *parser) parseExpr() (Expr, error) {
	return p.parseOrExpr()
}

// Column References

type ColumnRef struct {
	name string
}

func (c *ColumnRef) expressionMarker() {}

func (c *ColumnRef) String() string {
	return c.name
}

func (p *parser) parseColumnRef() (string, bool, error) {
	switch p.read().Kind {
	case lexing.Identifier:
		if len(p.read().Body) == 0 {
			return "", false, fmt.Errorf("expected column name identifier, got empty string")
		}
		return p.next().Body, false, nil
	case lexing.Star:
		if _, err := p.expect(lexing.Star, "*"); err != nil {
			return "", false, fmt.Errorf("expected *, got %s %s", p.read().Kind, p.read().Body)
		}
		return "", true, nil
	default:
		return "", false, fmt.Errorf("expected column name or *, got %s %s", p.read().Kind, p.read().Body)
	}
}

// Literals

type Literal struct {
	body string
}

func (l *Literal) expressionMarker() {}

func (l *Literal) String() string {
	return l.body
}

// Parameters

type Parameter struct {
	index int
}

func (p *Parameter) expressionMarker() {}

func (p *Parameter) String() string {
	return fmt.Sprintf("$%d", p.index)
}

func unpackParameter(s string) (*Parameter, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("expected first character of parameter to be $, got an empty string")
	}
	if s[0] != '$' {
		return nil, fmt.Errorf("expected first character of parameter to be $, got %s", string(s[0]))
	}
	num, err := strconv.Atoi(s[1:])
	if err != nil {
		return nil, fmt.Errorf("cannot cast %s to int", s[1:])
	}
	return &Parameter{index: num}, nil
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
	if p.read().Kind != lexing.Comparator {
		return left, nil
	}
	op := p.next().Body

	right, err := p.parseOperand()
	if err != nil {
		return nil, err
	}

	return &BinaryExpr{op: op, left: left, right: right}, nil
}

func (p *parser) parseOperand() (Expr, error) {
	switch p.read().Kind {
	case lexing.Identifier:
		return &ColumnRef{p.next().Body}, nil
	case lexing.String:
		return &Literal{p.next().Body}, nil
	case lexing.Number:
		return &Literal{p.next().Body}, nil
	case lexing.Param:
		node, err := unpackParameter(p.next().Body)
		if err != nil {
			return nil, err
		}
		return node, nil
	case lexing.LParen:
		p.next()
		node, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expectKind(lexing.RParen)
		if err != nil {
			return nil, err
		}
		return node, nil
	default:
		return nil, fmt.Errorf("Expected Identifier, String, Number, Parameter or (, got %s %s", p.read().Kind, p.read().Body)
	}
}
