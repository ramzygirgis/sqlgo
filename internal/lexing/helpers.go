package lexing

import (
	"fmt"
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

func validateFunctionName(name string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("function name is empty lmfao?")
	}
	if isDigit(name[0]) {
		return "", fmt.Errorf("function name cannot start with a digit")
	}
	for i := 0; i < len(name); i++ {
		if !isWordChar(name[i]) {
			return "", fmt.Errorf("unexpected character `%c` in query annotation's function name %q", name[i], name)
		}
	}
	return name, nil
}

func IsAnnotation(tk TokenKind) bool {
	return tk == OneAnnotation || tk == ManyAnnotation || tk == ExecAnnotation
}
