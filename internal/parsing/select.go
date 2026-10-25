package parsing

import (
	"fmt"
	"strings"
)

type SelectStatement struct {
	columnNames []string
	from        *FromKeyword
	where       Expr
	all         bool
}

func (s *SelectStatement) statementMarker() {}

func (s *SelectStatement) String() string {
	var sb strings.Builder
	lastIndex := len(s.columnNames) - 1
	sb.WriteString("SELECT")
	if s.all {
		sb.WriteString(" *")
		if lastIndex != -1 {
			sb.WriteString(",")
		}
	}
	for i, colName := range s.columnNames {
		sb.WriteString(fmt.Sprintf(" %s", colName))
		if i != lastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(s.from.String())
	if s.where != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(s.where.String())
	}
	sb.WriteString(";")
	return sb.String()
}
