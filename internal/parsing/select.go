package parsing

import (
	"fmt"
)

type SelectStatement struct {
	columns []Column
	from    *FromKeyword
	all     bool
}

func (s *SelectStatement) statementMarker() {}

func (s *SelectStatement) String() string {
	res := "SELECT"
	if s.all {
		res += " *,"
	}
	for _, col := range s.columns {
		res += fmt.Sprintf(" %s", col.name)
	}
	res = res[:len(res)-1]
	res += s.from.String()
	res += ";"
	return res
}
