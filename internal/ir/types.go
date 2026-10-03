package internal_representation

type Column struct {
	Name string
}

type Query struct {
	Name    string
	Columns []Column
	Table   string
}
