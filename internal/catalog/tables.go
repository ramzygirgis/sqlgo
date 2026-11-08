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
