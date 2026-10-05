package parsing

import (
// "fmt"
)

// QUESTIONS:
// how should i structure the AST?
// deciding what top level components should look like,
// and how to handle recursion is a pretty important problem to handle.
// i think using callbacks would be pretty nice, but need to come up
// with the best abstraction.
//
// how should i implement parseStatement? i think it makes
// most sense to treat tokens as a sort of stream.

func parseStatement(queryString string) (AST, error) {
	tokens, err := queryToTokens(queryString)
	if err != nil {
		return AST{}, err
	}
	tokens = tokens
	return AST{}, err
}

// idea: when turning atoms into tokens, store in a struct
// type Token struct{
// 		body: string
// 		type: identifier | keyword
// 		parsing callback: func(???) ...
// }
