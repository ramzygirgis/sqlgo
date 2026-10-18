package main

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

func main() {
	sqlQuery1 := `SELECT * FROM users;`
	sqlQuery2 := `SELECT *, name FROM users;`
	sqlQuery3 := `SELECT *, *, name FROM users;`
	sqlQuery4 := `SELECT *,*, name  FROM     users, hi;`
	queryStrings := []string{sqlQuery1, sqlQuery2, sqlQuery3, sqlQuery4}
	runSuite(queryStrings)
}

func runSuite(queryStrings []string) {
	for n, query := range queryStrings {
		fmt.Printf("*************\nLex Test %d:\n**************\n", n+1)
		testLex(query)
		fmt.Printf("*************\nParse Test %d:\n**************\n", n+1)
		testParse(query)

	}
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
