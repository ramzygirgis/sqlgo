package parsing

type Statement interface {
	statementMarker()
	String() string // for testing
}

type Column struct {
	name       string
	sqlType    string
	hasDefault bool
	defaultVal string
	nullable   bool
	primaryKey bool
}

type Table struct {
	name    string
	columns []Column
}
