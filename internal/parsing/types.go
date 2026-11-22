package parsing

import (
	"fmt"
)

type LiteralKind int

const (
	IllegalLiteral LiteralKind = iota
	StringLiteral
	NumberLiteral // maybe later, make this IntLiteral
	BooleanLiteral
	NullLiteral
)

func (l LiteralKind) String() string {
	strings := [...]string{
		"IllegalLiteral", "StringLiteral", "NumberLiteral",
		"BooleanLiteral", "NullLiteral",
	}
	if l < 0 || int(l) >= len(strings) {
		return fmt.Sprintf("LiteralKind(%d)", l)
	}
	return strings[l]
}

type Cardinality int

const (
	IllegalCardinality Cardinality = iota
	One
	Many
	Exec
)

func (c Cardinality) String() string {
	strings := [...]string{
		"IllegalCardinality", "One", "Many", "Exec",
	}
	if c < 0 || int(c) >= len(strings) {
		return fmt.Sprintf("Cardinality(%d)", c)
	}
	return strings[c]
}
