package parsing

import (
	"fmt"
)

type LiteralKind int

const (
	IllegalLiteral = iota
	BooleanLiteral
	NullLiteral
)

func (l LiteralKind) String() string {
	strings := [...]string{
		"IllegalLiteral", "BooleanLiteral", "NullLiteral",
	}
	if l < 0 || int(l) >= len(strings) {
		return fmt.Sprintf("LiteralKind(%d)", l)
	}
	return strings[l]
}
