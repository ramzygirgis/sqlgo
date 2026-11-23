package lexing

import (
	"strings"
)

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
	keywordSet["KEY"] = struct{}{}
	keywordSet["DEFAULT"] = struct{}{}
	keywordSet["WHERE"] = struct{}{}
	keywordSet["AND"] = struct{}{}
	keywordSet["OR"] = struct{}{}
	keywordSet["INSERT"] = struct{}{}
	keywordSet["INTO"] = struct{}{}
	keywordSet["VALUES"] = struct{}{}
	keywordSet["UPDATE"] = struct{}{}
	keywordSet["SET"] = struct{}{}
	keywordSet["DELETE"] = struct{}{}
	// storing true, false, null as keywords is kinda jank lol.
	// but i think this is correct
	keywordSet["TRUE"] = struct{}{}
	keywordSet["FALSE"] = struct{}{}
	keywordSet["NULL"] = struct{}{}

	_, ok := keywordSet[strings.ToUpper(s)]
	return ok
}

// note: we should have some sort of layer between lex and parse,
// and convert same types to eachother

func isType(s string) bool {
	typeSet := make(map[string]struct{})
	typeSet["TEXT"] = struct{}{}
	typeSet["INT"] = struct{}{}
	typeSet["TIMESTAMP"] = struct{}{}
	typeSet["BOOLEAN"] = struct{}{}

	_, ok := typeSet[strings.ToUpper(s)]
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
