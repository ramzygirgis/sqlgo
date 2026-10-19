package parsing

import (
	"fmt"
)

type FromKeyword struct {
	tableName string
}

func (s *FromKeyword) String() string {
	return fmt.Sprintf(" FROM %s", s.tableName)
}
