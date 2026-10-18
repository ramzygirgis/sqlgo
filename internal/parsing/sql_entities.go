package parsing

type Statement interface {
	statementMarker()
	String() string // for testing
}

type sqlValue interface {
	// getter?????
	String() string
}

type Column struct {
	name       string
	sqlType    string
	contraints []string
	hasDefault bool
	defaultVal sqlValue // hmmmm, maybe actually use an interface?????????
}

func (c Column) GetName() string {
	return c.name
}

type Table struct {
	name    string
	columns []Column
}
