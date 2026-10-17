package main

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func main() {
	sqlQuery := `SELECT * FROM users;`
	fmt.Printf("Lex Test:\n")
	testLex(sqlQuery)
	fmt.Printf("Parse Test:\n")
	testParse(sqlQuery)
}

func testLex(query string) {
	toks, err := parsing.Lex(query)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	for i, tok := range toks {
		fmt.Printf("token %d: %v\n", i+1, tok)
	}
}

func testParse(query string) {
	statement, err := parsing.Parse(query)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf("AST: %v\n", statement)
}
