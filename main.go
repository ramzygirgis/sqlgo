package main

import (
	"bufio"
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/commands"
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
	"os"
	"strings"
)

const usage = `sqlgo <command>

commands:
	generate	generate Go code from schema/ and query (.sql) files`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "generate":
		err := commands.Generate()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		fmt.Println(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %s -- run `sqlgo -h` for all supported commands\n", os.Args[1])
	}
}

func lexAndParseCLITest() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("What query do u wanna test?")
	query, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading query input:", err)
		return
	}
	query = strings.TrimSpace(query)
	queryStrings := []string{query}
	runSuite(queryStrings)
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
	toks, err := lexing.Lex(query)
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
