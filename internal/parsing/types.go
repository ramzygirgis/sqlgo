package parsing

import (
	"fmt"
)

type LiteralKind int

const (
	IllegalLiteral = iota
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
