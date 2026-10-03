package parsing

import (
	"fmt"
	"strings"
)

// this function strips a query of leading and trailing whitespace,
// and converts ending ';' into EOF.
func clean_query(query_string string) (string, error) {
	trimmed_query := strings.TrimSpace(query_string)
	length := len(trimmed_query) - 1
	if length == 0 {
		return "", fmt.Errorf("query is empty")
	}
	if trimmed_query[length-1] != ';' {
		return "", fmt.Errorf("query does not contain a terminator")
	}
	cleaned_query := trimmed_query[:length-1] + " SEMICOLON"
	return cleaned_query, nil
}

// this function takes in a query_string and returns a list of atoms
func query_to_atoms(query_string string) ([]string, error) {
	// TODO: convert (a, b, c) into [a, b, c]
	cleaned_query, err := clean_query(query_string)
	if err != nil {
		return []string{}, err
	}
	return strings.Fields(cleaned_query), nil
}

// this function takes in an atom and returns a token
func atom_to_token(atom string) string {
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
func query_to_tokens(query_string string) ([]string, error) {
	atoms, err := query_to_atoms(query_string)
	if err != nil {
		return []string{}, err
	}
	res := make([]string, len(atoms)+1)
	for i, atom := range atoms {
		res[i] = atom_to_token(atom)
	}
	return res, nil
}

// these above functions should eventually be portable to a whole page's queries. this means recursing and shit,
// but adding a EOF at the top level. for now, no EOF is needed as we don't have CTEs yet
