package parsing

import (
	"fmt"
	"strings"
)

// QUESTION:
// at what layer should we inject a "EOF" at the end of the text?
// maybe, we add this to queryString in pre-processing. need to figure this out

// this function strips a query of leading and trailing whitespace,
// and converts ending ';' into EOF.
func cleanQuery(queryString string) (string, error) {
	cleanedQuery := strings.TrimSpace(queryString)
	length := len(cleanedQuery)
	if length == 0 {
		return "", fmt.Errorf("query is empty")
	}
	if cleanedQuery[length-1] != ';' {
		return "", fmt.Errorf("query does not contain a terminator")
	}
	return cleanedQuery, nil
}

// this function takes in a query_string and returns a list of atoms
func queryToAtoms(queryString string) ([]string, error) {
	// TODO: convert (a, b, c) into [a, b, c]
	cleanedQuery, err := cleanQuery(queryString)
	if err != nil {
		return []string{}, err
	}
	atoms := strings.Fields(cleanedQuery)
	if len(atoms) == 0 {
		return []string{}, fmt.Errorf("query is empty")
	}
	return atoms, nil
}

// this function takes in an atom and returns a token
func atomToToken(atom string) Token {
	switch strings.ToUpper(atom) {
	case "SELECT":
		return Token{
			kind: Keyword,
			body: "SELECT",
		}
	case "FROM":
		return Token{
			kind: Keyword,
			body: "FROM",
		}
	case ";":
		return Token{
			kind: Semicolon,
			body: ";",
		}
	case "EOF":
		return Token{
			kind: EOF,
			body: "EOF",
		}
	}
	return Token{
		kind: Identifier,
		body: atom,
	}
}

// this function takes in a query and returns a slice of tokens
func queryToTokens(queryString string) ([]Token, error) {
	atoms, err := queryToAtoms(queryString)
	if err != nil {
		return []Token{}, err
	}
	res := make([]Token, len(atoms))
	for i, atom := range atoms {
		res[i] = atomToToken(atom)
	}
	return res, nil
}

// these above functions should eventually be portable to a whole page's queries. this means recursing and shit,
// but adding a EOF at the top level. for now, no EOF is needed as we don't have CTEs yet
