package catalog

import (
	"strings"
	"testing"

	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

// build parses src and applies it to a fresh catalog with ApplyAll.
func build(t *testing.T, src string) (*Catalog, error) {
	t.Helper()
	stmts, err := parsing.ParseAll(src)
	if err != nil {
		t.Fatalf("parse error (fix the test input, not the catalog): %v\ninput: %s", err, src)
	}
	c := New()
	return c, c.ApplyAll(stmts)
}

// mustBuild is build for inputs that must succeed.
func mustBuild(t *testing.T, src string) *Catalog {
	t.Helper()
	c, err := build(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v\ninput: %s", err, src)
	}
	return c
}

/* ---------- successful schemas ---------- */

func TestCreateTable(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string][]Column // table name -> columns in order
	}{
		{
			name:  "empty input",
			input: ``,
			want:  map[string][]Column{},
		},
		{
			name: "basic table",
			input: `CREATE TABLE users (
				id INT PRIMARY KEY,
				name TEXT NOT NULL,
				bio TEXT
			);`,
			want: map[string][]Column{
				"users": {
					{Name: "id", SqlType: "INT", Nullable: false},
					{Name: "name", SqlType: "TEXT", Nullable: false},
					{Name: "bio", SqlType: "TEXT", Nullable: true},
				},
			},
		},
		{
			name:  "single column",
			input: `CREATE TABLE t (x TEXT);`,
			want: map[string][]Column{
				"t": {{Name: "x", SqlType: "TEXT", Nullable: true}},
			},
		},
		{
			name:  "columns are nullable by default",
			input: `CREATE TABLE t (a INT, b TEXT);`,
			want: map[string][]Column{
				"t": {
					{Name: "a", SqlType: "INT", Nullable: true},
					{Name: "b", SqlType: "TEXT", Nullable: true},
				},
			},
		},
		{
			name:  "primary key implies not null",
			input: `CREATE TABLE t (id INT PRIMARY KEY);`,
			want: map[string][]Column{
				"t": {{Name: "id", SqlType: "INT", Nullable: false}},
			},
		},
		{
			name:  "NOT NULL and PRIMARY KEY together",
			input: `CREATE TABLE t (id INT NOT NULL PRIMARY KEY);`,
			want: map[string][]Column{
				"t": {{Name: "id", SqlType: "INT", Nullable: false}},
			},
		},
		{
			name:  "PRIMARY KEY before NOT NULL",
			input: `CREATE TABLE t (id INT PRIMARY KEY NOT NULL);`,
			want: map[string][]Column{
				"t": {{Name: "id", SqlType: "INT", Nullable: false}},
			},
		},
		{
			name:  "DEFAULT does not change nullability",
			input: `CREATE TABLE t (city TEXT DEFAULT Toronto);`,
			want: map[string][]Column{
				"t": {{Name: "city", SqlType: "TEXT", Nullable: true}},
			},
		},
		{
			name:  "DEFAULT with NOT NULL",
			input: `CREATE TABLE t (city TEXT NOT NULL DEFAULT Toronto);`,
			want: map[string][]Column{
				"t": {{Name: "city", SqlType: "TEXT", Nullable: false}},
			},
		},
		{
			name:  "lowercase types are normalized",
			input: `CREATE TABLE t (a int, b text);`,
			want: map[string][]Column{
				"t": {
					{Name: "a", SqlType: "INT", Nullable: true},
					{Name: "b", SqlType: "TEXT", Nullable: true},
				},
			},
		},
		{
			name:  "column order preserved",
			input: `CREATE TABLE t (c TEXT, a TEXT, b TEXT);`,
			want: map[string][]Column{
				"t": {
					{Name: "c", SqlType: "TEXT", Nullable: true},
					{Name: "a", SqlType: "TEXT", Nullable: true},
					{Name: "b", SqlType: "TEXT", Nullable: true},
				},
			},
		},
		{
			name:  "column names keep their case",
			input: `CREATE TABLE t (userName TEXT);`,
			want: map[string][]Column{
				"t": {{Name: "userName", SqlType: "TEXT", Nullable: true}},
			},
		},
		{
			name: "multiple tables",
			input: `CREATE TABLE users (id INT);
			        CREATE TABLE posts (id INT, title TEXT);`,
			want: map[string][]Column{
				"users": {{Name: "id", SqlType: "INT", Nullable: true}},
				"posts": {
					{Name: "id", SqlType: "INT", Nullable: true},
					{Name: "title", SqlType: "TEXT", Nullable: true},
				},
			},
		},
		{
			name: "same column name in different tables is fine",
			input: `CREATE TABLE a (id INT);
			        CREATE TABLE b (id INT);`,
			want: map[string][]Column{
				"a": {{Name: "id", SqlType: "INT", Nullable: true}},
				"b": {{Name: "id", SqlType: "INT", Nullable: true}},
			},
		},
		{
			name: "IF NOT EXISTS leaves original table alone",
			input: `CREATE TABLE users (id INT);
			        CREATE TABLE IF NOT EXISTS users (name TEXT);`,
			want: map[string][]Column{
				"users": {{Name: "id", SqlType: "INT", Nullable: true}},
			},
		},
		{
			name:  "IF NOT EXISTS on a new table creates it",
			input: `CREATE TABLE IF NOT EXISTS users (id INT);`,
			want: map[string][]Column{
				"users": {{Name: "id", SqlType: "INT", Nullable: true}},
			},
		},
		{
			name: "IF NOT EXISTS repeated",
			input: `CREATE TABLE IF NOT EXISTS users (id INT);
			        CREATE TABLE IF NOT EXISTS users (name TEXT);
			        CREATE TABLE IF NOT EXISTS users (email TEXT);`,
			want: map[string][]Column{
				"users": {{Name: "id", SqlType: "INT", Nullable: true}},
			},
		},
		{
			name: "IF NOT EXISTS skip does not affect later tables",
			input: `CREATE TABLE users (id INT);
			        CREATE TABLE IF NOT EXISTS users (name TEXT);
			        CREATE TABLE posts (id INT);`,
			want: map[string][]Column{
				"users": {{Name: "id", SqlType: "INT", Nullable: true}},
				"posts": {{Name: "id", SqlType: "INT", Nullable: true}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mustBuild(t, tt.input)
			if len(c.Tables) != len(tt.want) {
				t.Fatalf("got %d tables, want %d", len(c.Tables), len(tt.want))
			}
			for name, wantCols := range tt.want {
				table, ok := c.Table(name)
				if !ok {
					t.Fatalf("table %q missing", name)
				}
				if table.Name != name {
					t.Errorf("table stored under %q has Name %q", name, table.Name)
				}
				if len(table.Columns) != len(wantCols) {
					t.Fatalf("table %q: got %d columns %+v, want %d", name, len(table.Columns), table.Columns, len(wantCols))
				}
				for i, want := range wantCols {
					if got := table.Columns[i]; got != want {
						t.Errorf("table %q column %d: got %+v, want %+v", name, i, got, want)
					}
				}
			}
		})
	}
}

/* ---------- lookup ---------- */

func TestTableLookup(t *testing.T) {
	c := mustBuild(t, `CREATE TABLE users (id INT, name TEXT);`)

	t.Run("existing table", func(t *testing.T) {
		table, ok := c.Table("users")
		if !ok {
			t.Fatal("expected to find users")
		}
		if table.Name != "users" || len(table.Columns) != 2 {
			t.Errorf("got %+v", table)
		}
	})

	t.Run("missing table", func(t *testing.T) {
		if _, ok := c.Table("comments"); ok {
			t.Error("found table \"comments\", which was never created")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		if _, ok := c.Table(""); ok {
			t.Error("found a table with an empty name")
		}
	})

	t.Run("lookup on empty catalog", func(t *testing.T) {
		if _, ok := New().Table("users"); ok {
			t.Error("empty catalog should have no tables")
		}
	})
}

/* ---------- errors ---------- */

func TestApplyErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains []string // every substring must appear in the error
		wantTables  []string // tables that should exist after the failure
	}{
		{
			name: "duplicate table",
			input: `CREATE TABLE users (id INT);
			        CREATE TABLE users (id INT);`,
			errContains: []string{"users", "already exists"},
			wantTables:  []string{"users"},
		},
		{
			name: "IF NOT EXISTS then plain duplicate",
			input: `CREATE TABLE IF NOT EXISTS users (id INT);
			        CREATE TABLE users (id INT);`,
			errContains: []string{"users", "already exists"},
			wantTables:  []string{"users"},
		},
		{
			name:        "duplicate column",
			input:       `CREATE TABLE users (id INT, id TEXT);`,
			errContains: []string{"id", "users"},
			wantTables:  nil, // table must not be added
		},
		{
			name:        "duplicate column not adjacent",
			input:       `CREATE TABLE users (id INT, name TEXT, id TEXT);`,
			errContains: []string{"id", "users"},
			wantTables:  nil,
		},
		{
			name:        "duplicate column with identical definition",
			input:       `CREATE TABLE users (id INT, id INT);`,
			errContains: []string{"id", "users"},
			wantTables:  nil,
		},
		{
			name: "duplicate column after a good table",
			input: `CREATE TABLE ok (id INT);
			        CREATE TABLE bad (x INT, x INT);`,
			errContains: []string{"x", "bad"},
			wantTables:  []string{"ok"},
		},
		{
			name:        "select is unsupported",
			input:       `SELECT * FROM users;`,
			errContains: []string{"SelectStatement"},
			wantTables:  nil,
		},
		{
			name: "select after a create",
			input: `CREATE TABLE users (id INT);
			        SELECT * FROM users;`,
			errContains: []string{"SelectStatement"},
			wantTables:  []string{"users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := build(t, tt.input)
			if err == nil {
				t.Fatalf("expected an error, got none")
			}
			for _, sub := range tt.errContains {
				if !strings.Contains(err.Error(), sub) {
					t.Errorf("error %q should mention %q", err, sub)
				}
			}
			if len(c.Tables) != len(tt.wantTables) {
				t.Errorf("got %d tables after error, want %d %v", len(c.Tables), len(tt.wantTables), tt.wantTables)
			}
			for _, name := range tt.wantTables {
				if _, ok := c.Table(name); !ok {
					t.Errorf("table %q should still exist after the error", name)
				}
			}
		})
	}
}

func TestApplyNil(t *testing.T) {
	c := New()
	if err := c.Apply(nil); err == nil {
		t.Error("expected an error for nil statement, got none")
	}
	if len(c.Tables) != 0 {
		t.Errorf("nil statement should not change the catalog, got %d tables", len(c.Tables))
	}
}

func TestFailedCreateKeepsOriginal(t *testing.T) {
	c, err := build(t, `CREATE TABLE users (id INT);
	                    CREATE TABLE users (name TEXT, email TEXT);`)
	if err == nil {
		t.Fatal("expected duplicate table error")
	}
	table, ok := c.Table("users")
	if !ok {
		t.Fatal("original table should still exist")
	}
	if len(table.Columns) != 1 || table.Columns[0].Name != "id" {
		t.Errorf("failed CREATE should not overwrite the original; got columns %+v", table.Columns)
	}
}

/* ---------- ApplyAll ---------- */

func TestApplyAllEmpty(t *testing.T) {
	c := New()
	if err := c.ApplyAll(nil); err != nil {
		t.Errorf("ApplyAll(nil): unexpected error %v", err)
	}
	if err := c.ApplyAll([]parsing.Statement{}); err != nil {
		t.Errorf("ApplyAll(empty): unexpected error %v", err)
	}
	if len(c.Tables) != 0 {
		t.Errorf("got %d tables, want 0", len(c.Tables))
	}
}

func TestApplyAllStopsAtFirstError(t *testing.T) {
	c, err := build(t, `CREATE TABLE a (id INT);
	                    CREATE TABLE a (id INT);
	                    CREATE TABLE b (id INT);`)
	if err == nil {
		t.Fatal("expected an error from the duplicate table")
	}
	if _, ok := c.Table("a"); !ok {
		t.Error("table a (before the error) should exist")
	}
	if _, ok := c.Table("b"); ok {
		t.Error("table b (after the error) should not have been applied")
	}
}

func TestApplyAllMatchesApply(t *testing.T) {
	src := `CREATE TABLE users (id INT PRIMARY KEY, name TEXT);
	        CREATE TABLE posts (id INT, body TEXT NOT NULL);`
	stmts, err := parsing.ParseAll(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	viaAll := New()
	if err := viaAll.ApplyAll(stmts); err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}

	viaOne := New()
	for _, s := range stmts {
		if err := viaOne.Apply(s); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	}

	for _, name := range []string{"users", "posts"} {
		a, _ := viaAll.Table(name)
		b, _ := viaOne.Table(name)
		if len(a.Columns) != len(b.Columns) {
			t.Fatalf("table %q: ApplyAll and Apply disagree on column count", name)
		}
		for i := range a.Columns {
			if a.Columns[i] != b.Columns[i] {
				t.Errorf("table %q column %d: ApplyAll %+v vs Apply %+v", name, i, a.Columns[i], b.Columns[i])
			}
		}
	}
}
