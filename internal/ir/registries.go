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

func isSelectCardinality(c Cardinality) bool {
	cardinalitySet := make(map[Cardinality]struct{})
	cardinalitySet[One] = struct{}{}
	cardinalitySet[Many] = struct{}{}

	_, ok := cardinalitySet[c]
	return ok
}

func isInsertCardinality(c Cardinality) bool {
	cardinalitySet := make(map[Cardinality]struct{})
	cardinalitySet[Exec] = struct{}{}

	_, ok := cardinalitySet[c]
	return ok
}

func isUpdateCardinality(c Cardinality) bool {
	cardinalitySet := make(map[Cardinality]struct{})
	cardinalitySet[Exec] = struct{}{}

	_, ok := cardinalitySet[c]
	return ok
}

func isDeleteCardinality(c Cardinality) bool {
	cardinalitySet := make(map[Cardinality]struct{})
	cardinalitySet[Exec] = struct{}{}

	_, ok := cardinalitySet[c]
	return ok
}
