package catalog

type Table struct {
	Name    string
	Columns []Column
}

func (t *Table) Column(colName string) (*Column, bool) {
	for _, col := range t.Columns {
		if colName == col.Name {
			return &col, true
		}
	}
	return nil, false
}

// todo: it's chopped to have this not be in types.go.
// find a cleaner way to organize these files
