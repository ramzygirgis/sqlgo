package ir

import (
	"strings"
	"testing"
)

func validSelect() Query {
	return Query{
		Name:        "GetUser",
		Cardinality: One,
		Kind:        Select,
		TargetTable: "users",
		Parameters: []Parameter{
			{Index: 1, SqlType: TypeInt, Nullable: false},
			{Index: 2, SqlType: TypeText, Nullable: false},
		},
		TargetColumns: nil,
		ResultColumns: []Column{
			{Name: "id", SqlType: TypeInt, Nullable: false},
			{Name: "bio", SqlType: TypeText, Nullable: true},
		},
		Sql: "SELECT id, bio FROM users WHERE id = $1 AND name = $2;",
	}
}

func validInsert() Query {
	return Query{
		Name:        "CreateUser",
		Cardinality: Exec,
		Kind:        Insert,
		TargetTable: "users",
		Parameters: []Parameter{
			{Index: 1, SqlType: TypeInt, Nullable: false},
			{Index: 2, SqlType: TypeText, Nullable: false},
			{Index: 3, SqlType: TypeText, Nullable: true},
		},
		TargetColumns: []Column{
			{Name: "id", SqlType: TypeInt, Nullable: false},
			{Name: "name", SqlType: TypeText, Nullable: false},
			{Name: "bio", SqlType: TypeText, Nullable: true},
		},
		ResultColumns: nil,
		Sql:           "INSERT INTO users (id, name, bio) VALUES ($1, $2, $3);",
	}
}

func validUpdate() Query {
	return Query{
		Name:        "RenameUser",
		Cardinality: Exec,
		Kind:        Update,
		TargetTable: "users",
		Parameters: []Parameter{
			{Index: 1, SqlType: TypeText, Nullable: false},
			{Index: 2, SqlType: TypeInt, Nullable: false},
		},
		TargetColumns: []Column{
			{Name: "name", SqlType: TypeText, Nullable: false},
		},
		ResultColumns: nil,
		Sql:           "UPDATE users SET name = $1 WHERE id = $2;",
	}
}

func TestValidateAccepts(t *testing.T) {
	tests := []struct {
		name string
		q    Query
	}{
		{"select with columns and params", validSelect()},
		{"insert with params, no columns", validInsert()},
		{"select with no params", func() Query {
			q := validSelect()
			q.Parameters = nil
			return q
		}()},
		{"insert with no params", func() Query {
			q := validInsert()
			q.Parameters = nil
			return q
		}()},
		{"single param", func() Query {
			q := validSelect()
			q.Parameters = q.Parameters[:1]
			return q
		}()},
		{"nullable param is fine", func() Query {
			q := validSelect()
			q.Parameters[0].Nullable = true
			return q
		}()},

		// --- cardinality ---
		{"select :one", validSelect()},
		{"select :many", func() Query {
			q := validSelect()
			q.Cardinality = Many
			return q
		}()},
		{"insert :exec", validInsert()},

		// --- update ---
		{"update :exec", validUpdate()},
		{"update with no params", func() Query {
			q := validUpdate()
			q.Parameters = nil
			return q
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.q.Validate(); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}
}

func TestValidateRejects(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(q *Query)
		base   func() Query
	}{
		// --- invariant 1: sorted and contiguous from 1 ---
		{"params out of order", func(q *Query) {
			q.Parameters[0], q.Parameters[1] = q.Parameters[1], q.Parameters[0]
		}, validSelect},
		{"gap in indices", func(q *Query) {
			q.Parameters[1].Index = 3
		}, validSelect},
		{"does not start at 1", func(q *Query) {
			q.Parameters[0].Index = 2
			q.Parameters[1].Index = 3
		}, validSelect},
		{"index zero", func(q *Query) {
			q.Parameters = []Parameter{{Index: 0, SqlType: TypeInt}}
		}, validSelect},
		{"negative index", func(q *Query) {
			q.Parameters = []Parameter{{Index: -1, SqlType: TypeInt}}
		}, validSelect},
		{"gap in insert params", func(q *Query) {
			q.Parameters[2].Index = 4
		}, validInsert},

		// --- invariant 2: unique indices ---
		{"duplicate index", func(q *Query) {
			q.Parameters[1].Index = 1
		}, validSelect},
		{"duplicate in the middle", func(q *Query) {
			q.Parameters[2].Index = 2
		}, validInsert},

		// --- invariant 3: INSERT has no result columns ---
		{"insert with result columns", func(q *Query) {
			q.ResultColumns = []Column{{Name: "id", SqlType: TypeInt}}
		}, validInsert},

		// --- invariant 4: SELECT has at least one result column ---
		{"select with no columns (nil)", func(q *Query) {
			q.ResultColumns = nil
		}, validSelect},
		{"select with no columns (empty)", func(q *Query) {
			q.ResultColumns = []Column{}
		}, validSelect},

		// --- invariant 6: every param and column has a real SQL type ---
		{"param with illegal type", func(q *Query) {
			q.Parameters[0].SqlType = IllegalSQLType
		}, validSelect},
		{"column with illegal type", func(q *Query) {
			q.ResultColumns[1].SqlType = IllegalSQLType
		}, validSelect},
		{"insert param with illegal type", func(q *Query) {
			q.Parameters[2].SqlType = IllegalSQLType
		}, validInsert},

		// --- target columns ---
		{"select with target columns", func(q *Query) {
			q.TargetColumns = []Column{{Name: "id", SqlType: TypeInt}}
		}, validSelect},
		{"insert with no target columns (nil)", func(q *Query) {
			q.TargetColumns = nil
		}, validInsert},
		{"insert with no target columns (empty)", func(q *Query) {
			q.TargetColumns = []Column{}
		}, validInsert},
		{"target column with illegal type", func(q *Query) {
			q.TargetColumns[1].SqlType = IllegalSQLType
		}, validInsert},

		// --- query kind ---
		{"illegal query kind", func(q *Query) {
			q.Kind = IllegalQueryKind
		}, validSelect},

		// --- name ---
		{"empty name on select", func(q *Query) {
			q.Name = ""
		}, validSelect},
		{"empty name on insert", func(q *Query) {
			q.Name = ""
		}, validInsert},

		// --- cardinality ---
		{"illegal cardinality", func(q *Query) {
			q.Cardinality = IllegalCardinality
		}, validSelect},
		{"out-of-range cardinality", func(q *Query) {
			q.Cardinality = Cardinality(99)
		}, validSelect},
		{"insert :one", func(q *Query) {
			q.Cardinality = One
		}, validInsert},
		{"insert :many", func(q *Query) {
			q.Cardinality = Many
		}, validInsert},
		{"select :exec", func(q *Query) {
			q.Cardinality = Exec
		}, validSelect},

		// --- update ---
		{"update :one", func(q *Query) {
			q.Cardinality = One
		}, validUpdate},
		{"update :many", func(q *Query) {
			q.Cardinality = Many
		}, validUpdate},
		{"update with result columns", func(q *Query) {
			q.ResultColumns = []Column{{Name: "id", SqlType: TypeInt}}
		}, validUpdate},
		{"update with no target columns", func(q *Query) {
			q.TargetColumns = nil
		}, validUpdate},
		{"update with a gap in params", func(q *Query) {
			q.Parameters[1].Index = 3
		}, validUpdate},
		{"empty name on update", func(q *Query) {
			q.Name = ""
		}, validUpdate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.base()
			tt.mutate(&q)
			if err := q.Validate(); err == nil {
				t.Errorf("expected an error, got none\nquery: %+v", q)
			}
		})
	}
}

// Errors for an UPDATE should say UPDATE, not some other statement.
func TestValidateUpdateErrorsNameUpdate(t *testing.T) {
	mutations := map[string]func(q *Query){
		"result columns":    func(q *Query) { q.ResultColumns = []Column{{Name: "id", SqlType: TypeInt}} },
		"no target columns": func(q *Query) { q.TargetColumns = nil },
		"cardinality":       func(q *Query) { q.Cardinality = Many },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			q := validUpdate()
			mutate(&q)
			err := q.Validate()
			if err == nil {
				t.Fatal("expected an error, got none")
			}
			if !strings.Contains(err.Error(), "UPDATE") {
				t.Errorf("error %q should mention UPDATE", err)
			}
		})
	}
}

// Validate must not modify the query it checks.
func TestValidateDoesNotMutate(t *testing.T) {
	q := validSelect()
	q.Parameters[0], q.Parameters[1] = q.Parameters[1], q.Parameters[0] // out of order on purpose
	before0, before1 := q.Parameters[0].Index, q.Parameters[1].Index

	_ = q.Validate()

	if q.Parameters[0].Index != before0 || q.Parameters[1].Index != before1 {
		t.Error("Validate reordered the params; it should only check, not fix")
	}
}

func TestQueryKindString(t *testing.T) {
	tests := []struct {
		kind QueryKind
		want string
	}{
		{IllegalQueryKind, "IllegalQueryKind"},
		{Select, "Select"},
		{Insert, "Insert"},
		{Update, "Update"},
		{QueryKind(99), "QueryKind(99)"},
		{QueryKind(-1), "QueryKind(-1)"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("QueryKind(%d).String() = %q, want %q", int(tt.kind), got, tt.want)
		}
	}
}

func TestCardinalityString(t *testing.T) {
	tests := []struct {
		c    Cardinality
		want string
	}{
		{IllegalCardinality, "IllegalCardinality"},
		{One, "One"},
		{Many, "Many"},
		{Exec, "Exec"},
		{Cardinality(99), "Cardinality(99)"},
		{Cardinality(-1), "Cardinality(-1)"},
	}
	for _, tt := range tests {
		if got := tt.c.String(); got != tt.want {
			t.Errorf("Cardinality(%d).String() = %q, want %q", int(tt.c), got, tt.want)
		}
	}
}
