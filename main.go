package main

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

/*
func main() {
	sqlQuery3 := `SELECT *, *, name FROM users;`
	sqlQuery4 := `SELECT name FROM users;`
	sqlQuery1 := `CREATE TABLE users (
	name TEXT
	);`
	queryStrings := []string{sqlQuery1, sqlQuery2, sqlQuery3, sqlQuery4}
	runSuite(queryStrings)
}
*/

func main() {
	/*	sqlQuery2 := `CREATE TABLE users (
		id INT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		city TEXT DEFAULT Toronto
	);` */
	sqlQuery := []string{`SELECT name, * from users;`, `SELECT name, hello, *, * from users;`, `SELECT * frOm users;`}
	// sqlQuery := []string{sqlQuery2}
	runSuite(sqlQuery)
}

func runSuite(queryStrings []string) {
	var lexTest string
	var parseTest string
	fmt.Print("Run lex tests? (y/n)\n")
	// Scanln stops reading at a newline character
	_, err := fmt.Scanln(&lexTest)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	fmt.Print("Run parse tests? (y/n)\n")
	if _, err := fmt.Scanln(&parseTest); err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	for n, query := range queryStrings {
		if lexTest == "y" || lexTest == "Y" {
			fmt.Printf("\n*************\nLex Test %d:\n", n+1)
			testLex(query)
		}
		if parseTest == "y" || parseTest == "Y" {
			fmt.Printf("\n*************\nParse Test %d:\n", n+1)
			testParse(query)
		}
	}
}

func testLex(query string) {
	toks, err := parsing.Lex(query)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	for i, tok := range toks {
		fmt.Printf("token %d: %v\n", i+1, tok)
	}
}

func testParse(query string) {
	fmt.Printf("Your query: %s\n", query)
	statement, err := parsing.Parse(query)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("Reconstructed query: %s\n", statement)
}
