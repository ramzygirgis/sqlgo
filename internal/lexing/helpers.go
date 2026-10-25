package lexing

func isWordStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isWordChar(b byte) bool {
	return isWordStart(b) || isDigit(b)
}
