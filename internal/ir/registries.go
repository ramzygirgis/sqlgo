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

func isCardinality(c Cardinality) bool {
	cardinalitySet := make(map[Cardinality]struct{})
	cardinalitySet[One] = struct{}{}
	cardinalitySet[Many] = struct{}{}
	cardinalitySet[Exec] = struct{}{}

	_, ok := cardinalitySet[c]
	return ok
}
