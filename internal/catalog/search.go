package catalog

func (c *Catalog) Table(name string) (*Table, bool) {
	table, ok := c.Tables[name]
	if !ok {
		return nil, false
	}
	return table, true
}

func (t *Table) Column(colName string) (*Column, bool) {
	for _, col := range t.Columns {
		if colName == col.Name {
			return &col, true
		}
	}
	return nil, false
}
