package parsing

import (
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
