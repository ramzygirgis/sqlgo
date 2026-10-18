package parsing

import (
	"fmt"
)

type FromKeyword struct {
	table Table
}

func (s *FromKeyword) String() string {
	return fmt.Sprintf(" FROM %s", s.table.name)
}
