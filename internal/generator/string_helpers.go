package generator

import (
	"strings"
)

func isGoKeyword(s string) bool {
	keywords := make(map[string]struct{})
	keywords["break"] = struct{}{}
	keywords["case"] = struct{}{}
	keywords["chan"] = struct{}{}
	keywords["const"] = struct{}{}
	keywords["continue"] = struct{}{}
	keywords["default"] = struct{}{}
	keywords["defer"] = struct{}{}
	keywords["else"] = struct{}{}
	keywords["fallthrough"] = struct{}{}
	keywords["for"] = struct{}{}
	keywords["func"] = struct{}{}
	keywords["go"] = struct{}{}
	keywords["goto"] = struct{}{}
	keywords["if"] = struct{}{}
	keywords["import"] = struct{}{}
	keywords["interface"] = struct{}{}
	keywords["map"] = struct{}{}
	keywords["package"] = struct{}{}
	keywords["range"] = struct{}{}
	keywords["return"] = struct{}{}
	keywords["select"] = struct{}{}
	keywords["struct"] = struct{}{}
	keywords["switch"] = struct{}{}
	keywords["type"] = struct{}{}
	keywords["var"] = struct{}{}

	_, ok := keywords[s]
	return ok
}

func isInitialism(s string) bool {
	initialisms := make(map[string]struct{})
	initialisms["id"] = struct{}{}
	initialisms["api"] = struct{}{}
	initialisms["url"] = struct{}{}
	initialisms["http"] = struct{}{}
	initialisms["json"] = struct{}{}
	initialisms["sql"] = struct{}{}
	initialisms["uuid"] = struct{}{}
	initialisms["ip"] = struct{}{}

	_, ok := initialisms[strings.ToLower(s)]
	return ok
}

func snakeToCamel(identifier string, isExported bool) string {
	pieces := strings.Split(identifier, "_")
	res := ""
	for i, piece := range pieces {
		if len(piece) == 0 {
			continue
		}
		if !isExported && i == 0 {
			res += strings.ToLower(piece)
		} else if isInitialism(piece) {
			res += strings.ToUpper(piece)
		} else {
			res += strings.ToUpper(string(piece[0])) + strings.ToLower(string(piece[1:])) // is this correct?
		}
	}

	if isGoKeyword(res) {
		return "_" + res + "_"
	}
	return res
}
