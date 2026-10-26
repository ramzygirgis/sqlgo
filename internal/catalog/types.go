package catalog

type Column struct {
	Name     string
	SqlType  string
	Nullable bool
}

type Table struct {
	Name    string
	Columns []Column
}

type Catalog struct {
	Tables map[string]Table
}

// ideas:
// defining types as an iota? like why not use more enums,
// maybe even in parsing for these and for keywords
