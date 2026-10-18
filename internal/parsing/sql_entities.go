package parsing

type Statement interface {
	statementMarker()
	String() string // for testing
}

type Column struct {
	name string
}

func (c Column) GetName() string {
	return c.name
}

type Table struct {
	name string
}
