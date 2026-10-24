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
	keywordSet["CREATE"] = struct{}{}
	keywordSet["TABLE"] = struct{}{}
	keywordSet["NOT"] = struct{}{}
	keywordSet["IN"] = struct{}{}
	keywordSet["PRIMARY"] = struct{}{}
	keywordSet["NULL"] = struct{}{}
	keywordSet["KEY"] = struct{}{}
	keywordSet["DEFAULT"] = struct{}{}
	keywordSet["WHERE"] = struct{}{}
	keywordSet["AND"] = struct{}{}
	keywordSet["OR"] = struct{}{}
	keywordSet["LParen"] = struct{}{}
	keywordSet["RParen"] = struct{}{}

	_, ok := keywordSet[strings.ToUpper(s)]
	return ok
}

func isComparator(s string) bool {
	comparatorSet := make(map[string]struct{})
	comparatorSet["="] = struct{}{}
	comparatorSet[">"] = struct{}{}
	comparatorSet["<"] = struct{}{}
	comparatorSet["<="] = struct{}{}
	comparatorSet[">="] = struct{}{}
	comparatorSet["IS"] = struct{}{}
	comparatorSet["!="] = struct{}{}
	comparatorSet["<>"] = struct{}{}

	_, ok := comparatorSet[strings.ToUpper(s)]
	return ok
}

func isType(s string) bool {
	typeSet := make(map[string]struct{})
	typeSet["TEXT"] = struct{}{}
	typeSet["INT"] = struct{}{}

	_, ok := typeSet[strings.ToUpper(s)]
	return ok
}

// isContraint function????

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
