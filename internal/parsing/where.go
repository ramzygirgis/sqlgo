package parsing

import (
	"fmt"
)

type BinaryExpr struct {
	op    string
	left  Expr
	right Expr
}

func (b *BinaryExpr) expressionMarker() {}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", b.left.String(), b.op, b.right.String())
}
