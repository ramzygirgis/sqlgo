package parsing

import (
	"fmt"
	"strings"
	"unicode"
)

func isValidChar(b byte) bool {
	r := rune(b)
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

func isKeyword(s string) bool {
	keywordSet := make(map[string]struct{})
	keywordSet["SELECT"] = struct{}{}
	keywordSet["FROM"] = struct{}{}
	_, ok := keywordSet[s]
	return ok
}

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
