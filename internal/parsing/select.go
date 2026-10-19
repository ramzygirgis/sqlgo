package parsing

import (
	"fmt"
	"strings"
)

type SelectStatement struct {
	columns []Column
	from    *FromKeyword
	all     bool
}

func (s *SelectStatement) statementMarker() {}

func (s *SelectStatement) String() string {
	var sb strings.Builder
	lastIndex := len(s.columns) - 1
	sb.WriteString("SELECT")
	if s.all {
		sb.WriteString(" *")
		if lastIndex != -1 {
			sb.WriteString(",")
		}
	}
	for i, col := range s.columns {
		sb.WriteString(fmt.Sprintf(" %s", col.name))
		if i != lastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(s.from.String())
	sb.WriteString(";")
	return sb.String()
}
