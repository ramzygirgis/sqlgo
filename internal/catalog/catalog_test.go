package catalog

import (
	"strings"
	"testing"

	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

// build parses src and applies every statement to a fresh catalog.
func build(t *testing.T, src string) (*Catalog, error) {
	t.Helper()
	stmts, err := parsing.ParseAll(src)
	if err != nil {
		t.Fatalf("parse error (fix the test input, not the catalog): %v\ninput: %s", err, src)
	}
	c := New()
	for _, stmt := range stmts {
		if err := c.Apply(stmt); err != nil {
			return c, err
		}
	}
	return c, nil
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
			name:  "primary key implies not null",
			input: `CREATE TABLE t (id INT PRIMARY KEY);`,
			want: map[string][]Column{
				"t": {{Name: "id", SqlType: "INT", Nullable: false}},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := build(t, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(c.Tables) != len(tt.want) {
				t.Fatalf("got %d tables, want %d", len(c.Tables), len(tt.want))
			}
			for name, wantCols := range tt.want {
				table, ok := c.Tables[name]
				if !ok {
					t.Fatalf("table %q missing", name)
				}
				if table.Name != name {
					t.Errorf("table stored under %q has Name %q", name, table.Name)
				}
				if len(table.Columns) != len(wantCols) {
					t.Fatalf("table %q: got %d columns %v, want %d", name, len(table.Columns), table.Columns, len(wantCols))
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

func TestLookupMissingTable(t *testing.T) {
	c, err := build(t, `CREATE TABLE users (id INT);`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.Tables["comments"]; ok {
		t.Error("found table \"comments\", which was never created")
	}
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
			name:        "duplicate column",
			input:       `CREATE TABLE users (id INT, id TEXT);`,
			errContains: []string{"id", "users"},
			wantTables:  nil, // table must not be added
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
				t.Errorf("got %d tables after error, want %d (%v)", len(c.Tables), len(tt.wantTables), tt.wantTables)
			}
			for _, name := range tt.wantTables {
				if _, ok := c.Tables[name]; !ok {
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

func TestDuplicateTableKeepsOriginal(t *testing.T) {
	c, _ := build(t, `CREATE TABLE users (id INT);
	                  CREATE TABLE users (name TEXT, email TEXT);`)
	table, ok := c.Tables["users"]
	if !ok {
		t.Fatal("original table should still exist")
	}
	if len(table.Columns) != 1 || table.Columns[0].Name != "id" {
		t.Errorf("failed CREATE should not overwrite the original; got columns %+v", table.Columns)
	}
}
