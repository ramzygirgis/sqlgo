package parsing

import (
	"fmt"
	"strings"
)

// this function strips a query of leading and trailing whitespace,
// and converts ending ';' into EOF.
func cleanQuery(queryString string) (string, error) {
	trimmedQuery := strings.TrimSpace(queryString)
	length := len(trimmedQuery) - 1
	if length == 0 {
		return "", fmt.Errorf("query is empty")
	}
	if trimmedQuery[length-1] != ';' {
		return "", fmt.Errorf("query does not contain a terminator")
	}
	cleanedQuery := trimmedQuery[:length-1] + " SEMICOLON"
	return cleanedQuery, nil
}

// this function takes in a query_string and returns a list of atoms
func queryToAtoms(queryString string) ([]string, error) {
	// TODO: convert (a, b, c) into [a, b, c]
	cleanedQuery, err := cleanQuery(queryString)
	if err != nil {
		return []string{}, err
	}
	return strings.Fields(cleanedQuery), nil
}

// this function takes in an atom and returns a token
func atomToToken(atom string) string {
	switch strings.ToLower(atom) {
	case "select":
		return `KEYWORD("SELECT")`
	case "from":
		return `KEYWORD("FROM")`
	case "where":
		return `KEYWORD("WHERE")`
	case "eof":
		return "SEMICOLON"
	}
	return fmt.Sprintf(`IDENTIFIER("%s")`, strings.ToLower(atom))
}

// this function takes in a query and returns a slice of tokens
func queryToTokens(queryString string) ([]string, error) {
	atoms, err := queryToAtoms(queryString)
	if err != nil {
		return []string{}, err
	}
	res := make([]string, len(atoms)+1)
	for i, atom := range atoms {
		res[i] = atomToToken(atom)
	}
	return res, nil
}

// these above functions should eventually be portable to a whole page's queries. this means recursing and shit,
// but adding a EOF at the top level. for now, no EOF is needed as we don't have CTEs yet
