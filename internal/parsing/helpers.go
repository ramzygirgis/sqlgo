package parsing

import (
	"fmt"
	"strconv"
	"strings"
)

func isWordStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isWordChar(b byte) bool {
	return isWordStart(b) || isDigit(b)
}

func unpackParameter(s string) (*Parameter, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("expected first character of parameter to be $, got an empty string")
	}
	if s[0] != '$' {
		return nil, fmt.Errorf("expected first character of parameter to be $, got %s", string(s[0]))
	}
	num, err := strconv.Atoi(s[1:])
	if err != nil {
		return nil, fmt.Errorf("cannot cast %s to int", s[1:])
	}
	return &Parameter{index: num}, nil
}

func isKeyword(s string) bool {
	keywordSet := make(map[string]struct{})
	keywordSet["SELECT"] = struct{}{}
	keywordSet["FROM"] = struct{}{}
	keywordSet["CREATE"] = struct{}{}
	keywordSet["TABLE"] = struct{}{}
	keywordSet["IF"] = struct{}{}
	keywordSet["NOT"] = struct{}{}
	keywordSet["EXISTS"] = struct{}{}
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

/*
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
*/

func isType(s string) bool {
	typeSet := make(map[string]struct{})
	typeSet["TEXT"] = struct{}{}
	typeSet["INT"] = struct{}{}

	_, ok := typeSet[strings.ToUpper(s)]
	return ok
}
