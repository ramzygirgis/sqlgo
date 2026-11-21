package catalog

func New() *Catalog {
	tables := make(map[string]*Table)
	return &Catalog{Tables: tables}
}
