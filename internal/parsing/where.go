package parsing

type booleanExpression interface {
	evaluate()
}

type whereCondition struct {
}

// idea:
/*
* where
* col1
* binOp
*
*
*
 */
