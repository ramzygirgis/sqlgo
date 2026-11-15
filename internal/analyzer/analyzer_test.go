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

// testName is the query name used by the single-query helpers.
const testName = "TestQuery"

// buildCatalog parses schema and applies it to a fresh catalog.
func buildCatalog(t *testing.T, schema string) *catalog.Catalog {
	t.Helper()
	stmts, err := parsing.ParseSchema(schema)
	if err != nil {
		t.Fatalf("schema parse error: %v", err)
	}
	cat := catalog.New()
	if err := cat.ApplyAll(stmts); err != nil {
		t.Fatalf("schema apply error: %v", err)
	}
	return cat
}

// parseOne parses a single annotated query. Setup failures fail the test
// immediately, since they mean the test input is wrong, not the analyzer.
func parseOne(t *testing.T, annotated string) *parsing.AnnotatedQuery {
	t.Helper()
	qs, err := parsing.ParseQueries(annotated)
	if err != nil {
		t.Fatalf("query parse error: %v\ninput:\n%s", err, annotated)
	}
	if len(qs) != 1 {
		t.Fatalf("expected exactly 1 query, got %d", len(qs))
	}
	return qs[0]
}

// analyze annotates query as `-- name: TestQuery <card>` and analyzes it
// against the users schema.
func analyze(t *testing.T, query, card string) (*ir.Query, error) {
	t.Helper()
	cat := buildCatalog(t, usersSchema)
	q := parseOne(t, "-- name: "+testName+" "+card+"\n"+query)
	return AnalyzeQuery(cat, q)
}

// defaultCard picks the natural cardinality for a statement.
func defaultCard(query string) string {
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT") {
		return ":many"
	}
	return ":exec"
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
			q, err := analyze(t, tt.query, defaultCard(tt.query))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err := q.Validate(); err != nil {
				t.Errorf("result fails Validate(): %v", err)
			}

			if q.Name != testName {
				t.Errorf("Name: got %q, want %q", q.Name, testName)
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

/* ---------- cardinality ---------- */

func TestAnalyzeCardinality(t *testing.T) {
	tests := []struct {
		name  string
		query string
		card  string
		want  ir.Cardinality
	}{
		{"select :one", `SELECT id FROM users WHERE id = $1;`, ":one", ir.One},
		{"select :many", `SELECT id FROM users;`, ":many", ir.Many},
		{"select star :one", `SELECT * FROM users WHERE id = $1;`, ":one", ir.One},
		{"insert :exec", `INSERT INTO users (name) VALUES ($1);`, ":exec", ir.Exec},
		{"insert without column list :exec", `INSERT INTO users VALUES ($1, $2, $3);`, ":exec", ir.Exec},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := analyze(t, tt.query, tt.card)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if q.Cardinality != tt.want {
				t.Errorf("Cardinality: got %v, want %v", q.Cardinality, tt.want)
			}
			if err := q.Validate(); err != nil {
				t.Errorf("result fails Validate(): %v", err)
			}
		})
	}
}

/* ---------- error cases ---------- */

func TestAnalyzeErrors(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		card        string   // empty means use defaultCard
		errContains []string // case-insensitive
	}{
		// --- SELECT ---
		{"select missing table", `SELECT * FROM posts;`, "", []string{"posts"}},
		{"select missing column", `SELECT email FROM users;`, "", []string{"email"}},
		{"select one good one bad column", `SELECT id, email FROM users;`, "", []string{"email"}},
		{"where missing column", `SELECT id FROM users WHERE email = $1;`, "", []string{"email"}},
		{"where missing column without params", `SELECT id FROM users WHERE email = 5;`, "", []string{"email"}},
		{"where missing column nested", `SELECT id FROM users WHERE id = 1 AND (name = 'x' OR email = 'y');`, "", []string{"email"}},
		{"conflicting param types", `SELECT id FROM users WHERE id = $1 AND name = $1;`, "", []string{"mismatch"}},
		{"param compared to param", `SELECT id FROM users WHERE $1 = $2;`, "", []string{"$1"}},
		{"param compared to literal", `SELECT id FROM users WHERE $1 = 5;`, "", []string{"$1"}},
		{"bare param", `SELECT id FROM users WHERE $1;`, "", []string{"$1"}},
		{"untyped param next to typed one", `SELECT id FROM users WHERE id = $1 AND $2 = 5;`, "", []string{"$2"}},
		{"missing $1", `SELECT id FROM users WHERE id = $2;`, "", []string{"$1"}},
		{"gap in params", `SELECT id FROM users WHERE id = $1 AND name = $3;`, "", []string{"$2"}},

		// --- INSERT ---
		{"insert missing table", `INSERT INTO posts VALUES (1);`, "", []string{"posts"}},
		{"insert missing column", `INSERT INTO users (email) VALUES ($1);`, "", []string{"email"}},
		{"insert duplicate column", `INSERT INTO users (id, id) VALUES ($1, $2);`, "", []string{"id"}},
		{"insert too few values", `INSERT INTO users (id, name) VALUES ($1);`, "", []string{"2", "1"}},
		{"insert too many values", `INSERT INTO users (id) VALUES ($1, $2);`, "", []string{"1", "2"}},
		{"insert no column list, too few values", `INSERT INTO users VALUES ($1);`, "", []string{"3", "1"}},
		{"insert column ref as value", `INSERT INTO users (id) VALUES (name);`, "", []string{"name"}},
		{"insert missing $1", `INSERT INTO users (id) VALUES ($2);`, "", []string{"$1"}},
		{"insert conflicting param types", `INSERT INTO users (id, name) VALUES ($1, $1);`, "", []string{"mismatch"}},

		// --- cardinality ---
		{"insert :one", `INSERT INTO users (name) VALUES ($1);`, ":one", []string{"one"}},
		{"insert :many", `INSERT INTO users (name) VALUES ($1);`, ":many", []string{"many"}},
		{"select :exec", `SELECT id FROM users;`, ":exec", []string{"exec"}},

		// --- unsupported ---
		{"create table is not a query", `CREATE TABLE t (a INT);`, ":exec", []string{"unsupported"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.card
			if card == "" {
				card = defaultCard(tt.query)
			}
			q, err := analyze(t, tt.query, card)
			if err == nil {
				t.Fatalf("expected an error, got none\nresult: %+v", q)
			}
			msg := strings.ToLower(err.Error())
			for _, sub := range tt.errContains {
				if !strings.Contains(msg, strings.ToLower(sub)) {
					t.Errorf("error %q should mention %q", err, sub)
				}
			}
			if !strings.HasPrefix(err.Error(), testName) {
				t.Errorf("error %q should start with the query name %q", err, testName)
			}
		})
	}
}

/* ---------- whole query files ---------- */

func TestAnalyzeQueryFile(t *testing.T) {
	cat := buildCatalog(t, usersSchema)
	qs, err := parsing.ParseQueries(`
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT id, name FROM users;

-- Insert a new user.
-- name: CreateUser :exec
INSERT INTO users (name, bio) VALUES ($1, $2);
`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	want := []struct {
		name   string
		card   ir.Cardinality
		kind   ir.QueryKind
		params int
	}{
		{"GetUser", ir.One, ir.Select, 1},
		{"ListUsers", ir.Many, ir.Select, 0},
		{"CreateUser", ir.Exec, ir.Insert, 2},
	}
	if len(qs) != len(want) {
		t.Fatalf("got %d queries, want %d", len(qs), len(want))
	}

	for i, w := range want {
		q, err := AnalyzeQuery(cat, qs[i])
		if err != nil {
			t.Fatalf("query %d (%s): unexpected error: %v", i, w.name, err)
		}
		if q.Name != w.name {
			t.Errorf("query %d: Name = %q, want %q", i, q.Name, w.name)
		}
		if q.Cardinality != w.card {
			t.Errorf("query %d: Cardinality = %v, want %v", i, q.Cardinality, w.card)
		}
		if q.Kind != w.kind {
			t.Errorf("query %d: Kind = %v, want %v", i, q.Kind, w.kind)
		}
		if len(q.Parameters) != w.params {
			t.Errorf("query %d: %d params, want %d", i, len(q.Parameters), w.params)
		}
		if err := q.Validate(); err != nil {
			t.Errorf("query %d: fails Validate(): %v", i, err)
		}
	}
}

func TestErrorNamesTheRightQuery(t *testing.T) {
	cat := buildCatalog(t, usersSchema)
	qs, err := parsing.ParseQueries(`
-- name: GoodOne :many
SELECT id FROM users;

-- name: BrokenOne :many
SELECT email FROM users;
`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if _, err := AnalyzeQuery(cat, qs[0]); err != nil {
		t.Fatalf("GoodOne: unexpected error: %v", err)
	}
	_, err = AnalyzeQuery(cat, qs[1])
	if err == nil {
		t.Fatal("BrokenOne: expected an error, got none")
	}
	if !strings.HasPrefix(err.Error(), "BrokenOne") {
		t.Errorf("error %q should start with %q", err, "BrokenOne")
	}
}

/* ---------- AST is left untouched ---------- */

func TestAnalyzeDoesNotMutateAST(t *testing.T) {
	cat := buildCatalog(t, usersSchema)
	q := parseOne(t, "-- name: All :many\nSELECT * FROM users;")
	before := q.Statement.String()

	if _, err := AnalyzeQuery(cat, q); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if after := q.Statement.String(); after != before {
		t.Errorf("AnalyzeQuery modified the statement:\nbefore: %s\n after: %s", before, after)
	}
}
