package ir

func isType(s SQLType) bool {
	typeSet := make(map[SQLType]struct{})
	typeSet[TypeText] = struct{}{}
	typeSet[TypeInt] = struct{}{}
	typeSet[TypeTimestamp] = struct{}{}
	typeSet[TypeBoolean] = struct{}{}

	_, ok := typeSet[s]
	return ok
}
