package main

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func main() {
	sqlQuery1 := `SELECT * FROM users;`
	sqlQuery2 := `SELECT *, name FROM users;`
	fmt.Printf("Lex Test 1:\n")
	testLex(sqlQuery1)
	fmt.Printf("Parse Test 2:\n")
	testParse(sqlQuery1)
	fmt.Printf("Lex Test 2:\n")
	testLex(sqlQuery2)
	fmt.Printf("Parse Test 2:\n")
	testParse(sqlQuery2)
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
	fmt.Printf("AST: %s\n", statement)
}
