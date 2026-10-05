package parsing

import (
	"fmt"
)

func parseStatement(queryString string) (AST, error) {
	queryToTokens, err := queryToTokens(queryString)
	if err != nil {
		return AST{}, err
	}
	// TODO
	return AST{}, err
}

func buildASTfromQuery(queryString string) (AST, error) {
	queryToTokens, err := queryToTokens(queryString)
	if err != nil {
		return AST{}, err
	}
	if queryToTokens[0] != `"KEYWORD("SELECT")` {
		return AST{}, fmt.Errorf("query does not begin with a SELECT keyword")
	}
}

// idea: when turning atoms into tokens, store in a struct
// type Token struct{
// 		body: string
// 		type: identifier | keyword
// 		parsing callback: func(???) ...
// }
