package main

import (
	"fmt"
	"github.com/ramzygirgis/sqlgo/internal/commands"
	"os"
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
		os.Exit(1)
	}
}
