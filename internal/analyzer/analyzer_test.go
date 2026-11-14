package analyzer

import (
	"strings"
	"testing"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

const usersSchema = `
CREATE TABLE users (
	id INT PRIMARY KEY,
	name TEXT NOT NULL,
	bio TEXT
);`

// analyze builds a catalog from schema, parses query, and runs AnalyzeQuery.
// Setup failures (bad schema, unparseable query) fail the test immediately,
// since they mean the test input is wrong, not the analyzer.
func analyze(t *testing.T, schema, query string) (*ir.Query, error) {
	t.Helper()

	stmts, err := parsing.ParseAll(schema)
	if err != nil {
		t.Fatalf("schema parse error: %v", err)
	}
	cat := catalog.New()
	if err := cat.ApplyAll(stmts); err != nil {
		t.Fatalf("schema apply error: %v", err)
	}

	stmt, err := parsing.Parse(query)
	if err != nil {
		t.Fatalf("query parse error: %v\nquery: %s", err, query)
	}
	return AnalyzeQuery(cat, stmt)
}

func normalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// shorthand constructors keep the tables readable
func col(name string, t ir.SQLType, nullable bool) ir.Column {
	return ir.Column{Name: name, SqlType: t, Nullable: nullable}
}

func param(i int, t ir.SQLType, nullable bool) ir.Parameter {
	return ir.Parameter{Index: i, SqlType: t, Nullable: nullable}
}

var (
	idCol   = col("id", ir.TypeInt, false)
	nameCol = col("name", ir.TypeText, false)
	bioCol  = col("bio", ir.TypeText, true)
)

/* ---------- success cases ---------- */

func TestAnalyzeSuccess(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		kind          ir.QueryKind
		resultColumns []ir.Column
		targetColumns []ir.Column
		params        []ir.Parameter
		sql           string // checked only if non-empty (whitespace-insensitive)
	}{
		// --- SELECT: result columns ---
		{
			name:          "select star",
			query:         `SELECT * FROM users;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol, nameCol, bioCol},
			sql:           `SELECT id, name, bio FROM users;`,
		},
		{
			name:          "select named columns",
			query:         `SELECT name, bio FROM users;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{nameCol, bioCol},
		},
		{
			name:          "select columns keep written order",
			query:         `SELECT bio, id FROM users;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{bioCol, idCol},
		},
		{
			name:          "select single column",
			query:         `SELECT id FROM users;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
		},

		// --- SELECT: WHERE without params ---
		{
			name:          "where with literal only",
			query:         `SELECT id FROM users WHERE id = 5;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
		},
		{
			name:          "where comparing two columns",
			query:         `SELECT id FROM users WHERE name = bio;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
		},

		// --- SELECT: param inference ---
		{
			name:          "param on right",
			query:         `SELECT name, bio FROM users WHERE id = $1;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{nameCol, bioCol},
			params:        []ir.Parameter{param(1, ir.TypeInt, false)},
		},
		{
			name:          "param on left",
			query:         `SELECT id FROM users WHERE $1 = name;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
			params:        []ir.Parameter{param(1, ir.TypeText, false)},
		},
		{
			name:          "where params are never nullable",
			query:         `SELECT id FROM users WHERE bio = $1;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
			params:        []ir.Parameter{param(1, ir.TypeText, false)},
		},
		{
			name:          "same param used twice with same type",
			query:         `SELECT id FROM users WHERE $1 = name OR bio = $1;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
			params:        []ir.Parameter{param(1, ir.TypeText, false)},
		},
		{
			name:          "two params",
			query:         `SELECT id FROM users WHERE id = $1 AND name = $2;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
			params:        []ir.Parameter{param(1, ir.TypeInt, false), param(2, ir.TypeText, false)},
		},
		{
			name:          "params sorted regardless of appearance",
			query:         `SELECT id FROM users WHERE name = $2 AND id = $1;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
			params:        []ir.Parameter{param(1, ir.TypeInt, false), param(2, ir.TypeText, false)},
		},
		{
			name:          "param inside parentheses",
			query:         `SELECT id FROM users WHERE (id = $1 OR name = $2) AND bio = $3;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
			params: []ir.Parameter{
				param(1, ir.TypeInt, false),
				param(2, ir.TypeText, false),
				param(3, ir.TypeText, false),
			},
		},
		{
			name:          "other comparators infer too",
			query:         `SELECT id FROM users WHERE id > $1 AND id <= $2;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol},
			params:        []ir.Parameter{param(1, ir.TypeInt, false), param(2, ir.TypeInt, false)},
		},
		{
			name:          "star with where keeps where in sql",
			query:         `SELECT * FROM users WHERE id = $1;`,
			kind:          ir.Select,
			resultColumns: []ir.Column{idCol, nameCol, bioCol},
			params:        []ir.Parameter{param(1, ir.TypeInt, false)},
			sql:           `SELECT id, name, bio FROM users WHERE (id = $1);`,
		},

		// --- INSERT ---
		{
			name:          "insert with column list",
			query:         `INSERT INTO users (name, bio) VALUES ($1, $2);`,
			kind:          ir.Insert,
			targetColumns: []ir.Column{nameCol, bioCol},
			params:        []ir.Parameter{param(1, ir.TypeText, false), param(2, ir.TypeText, true)},
		},
		{
			name:          "insert without column list targets every column",
			query:         `INSERT INTO users VALUES ($1, 'bob', $2);`,
			kind:          ir.Insert,
			targetColumns: []ir.Column{idCol, nameCol, bioCol},
			params:        []ir.Parameter{param(1, ir.TypeInt, false), param(2, ir.TypeText, true)},
		},
		{
			name:          "insert literals only",
			query:         `INSERT INTO users (id, name) VALUES (1, 'bob');`,
			kind:          ir.Insert,
			targetColumns: []ir.Column{idCol, nameCol},
		},
		{
			name:          "insert params out of column order",
			query:         `INSERT INTO users (id, name) VALUES ($2, $1);`,
			kind:          ir.Insert,
			targetColumns: []ir.Column{idCol, nameCol},
			params:        []ir.Parameter{param(1, ir.TypeText, false), param(2, ir.TypeInt, false)},
		},
		{
			name:          "insert column list in non-table order",
			query:         `INSERT INTO users (bio, id) VALUES ($1, $2);`,
			kind:          ir.Insert,
			targetColumns: []ir.Column{bioCol, idCol},
			params:        []ir.Parameter{param(1, ir.TypeText, true), param(2, ir.TypeInt, false)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := analyze(t, usersSchema, tt.query)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err := q.Validate(); err != nil {
				t.Errorf("result fails Validate(): %v", err)
			}

			if q.Kind != tt.kind {
				t.Errorf("Kind: got %v, want %v", q.Kind, tt.kind)
			}
			if q.TargetTable != "users" {
				t.Errorf("TargetTable: got %q, want %q", q.TargetTable, "users")
			}
			checkColumns(t, "ResultColumns", q.ResultColumns, tt.resultColumns)
			checkColumns(t, "TargetColumns", q.TargetColumns, tt.targetColumns)
			checkParams(t, q.Parameters, tt.params)

			if tt.sql != "" && normalize(q.Sql) != normalize(tt.sql) {
				t.Errorf("Sql:\n got: %s\nwant: %s", q.Sql, tt.sql)
			}
		})
	}
}

func checkColumns(t *testing.T, field string, got, want []ir.Column) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d columns %+v, want %d %+v", field, len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: got %+v, want %+v", field, i, got[i], want[i])
		}
	}
}

func checkParams(t *testing.T, got, want []ir.Parameter) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("Parameters: got %d %+v, want %d %+v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Parameters[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

/* ---------- error cases ---------- */

func TestAnalyzeErrors(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		errContains []string // case-insensitive
	}{
		// --- SELECT ---
		{"select missing table", `SELECT * FROM posts;`, []string{"posts"}},
		{"select missing column", `SELECT email FROM users;`, []string{"email"}},
		{"select one good one bad column", `SELECT id, email FROM users;`, []string{"email"}},
		{"where missing column", `SELECT id FROM users WHERE email = $1;`, []string{"email"}},
		{"where missing column without params", `SELECT id FROM users WHERE email = 5;`, []string{"email"}},
		{"where missing column nested", `SELECT id FROM users WHERE id = 1 AND (name = 'x' OR email = 'y');`, []string{"email"}},
		{"conflicting param types", `SELECT id FROM users WHERE id = $1 AND name = $1;`, []string{"mismatch"}},
		{"param compared to param", `SELECT id FROM users WHERE $1 = $2;`, []string{"$1"}},
		{"param compared to literal", `SELECT id FROM users WHERE $1 = 5;`, []string{"$1"}},
		{"bare param", `SELECT id FROM users WHERE $1;`, []string{"$1"}},
		{"untyped param next to typed one", `SELECT id FROM users WHERE id = $1 AND $2 = 5;`, []string{"$2"}},
		{"missing $1", `SELECT id FROM users WHERE id = $2;`, []string{"$1"}},
		{"gap in params", `SELECT id FROM users WHERE id = $1 AND name = $3;`, []string{"$2"}},

		// --- INSERT ---
		{"insert missing table", `INSERT INTO posts VALUES (1);`, []string{"posts"}},
		{"insert missing column", `INSERT INTO users (email) VALUES ($1);`, []string{"email"}},
		{"insert duplicate column", `INSERT INTO users (id, id) VALUES ($1, $2);`, []string{"id"}},
		{"insert too few values", `INSERT INTO users (id, name) VALUES ($1);`, []string{"2", "1"}},
		{"insert too many values", `INSERT INTO users (id) VALUES ($1, $2);`, []string{"1", "2"}},
		{"insert no column list, too few values", `INSERT INTO users VALUES ($1);`, []string{"3", "1"}},
		{"insert column ref as value", `INSERT INTO users (id) VALUES (name);`, []string{"name"}},
		{"insert missing $1", `INSERT INTO users (id) VALUES ($2);`, []string{"$1"}},
		{"insert conflicting param types", `INSERT INTO users (id, name) VALUES ($1, $1);`, []string{"mismatch"}},

		// --- unsupported ---
		{"create table is not a query", `CREATE TABLE t (a INT);`, []string{"unsupported"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := analyze(t, usersSchema, tt.query)
			if err == nil {
				t.Fatalf("expected an error, got none\nresult: %+v", q)
			}
			msg := strings.ToLower(err.Error())
			for _, sub := range tt.errContains {
				if !strings.Contains(msg, strings.ToLower(sub)) {
					t.Errorf("error %q should mention %q", err, sub)
				}
			}
		})
	}
}

/* ---------- AST is left untouched ---------- */

func TestAnalyzeDoesNotMutateAST(t *testing.T) {
	stmts, err := parsing.ParseAll(usersSchema)
	if err != nil {
		t.Fatal(err)
	}
	cat := catalog.New()
	if err := cat.ApplyAll(stmts); err != nil {
		t.Fatal(err)
	}

	stmt, err := parsing.Parse(`SELECT * FROM users;`)
	if err != nil {
		t.Fatal(err)
	}
	before := stmt.String()

	if _, err := AnalyzeQuery(cat, stmt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if after := stmt.String(); after != before {
		t.Errorf("AnalyzeQuery modified the statement:\nbefore: %s\n after: %s", before, after)
	}
}
