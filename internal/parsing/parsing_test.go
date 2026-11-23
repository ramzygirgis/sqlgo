package parsing

import (
	"strings"
	"testing"
)

// normalize collapses all runs of whitespace into single spaces, so tests
// don't break over newlines/indentation in String() output.
func normalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

/* ---------- round trip: parse, then print back ---------- */

func TestParseRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// --- SELECT basics ---
		{"simple select", `SELECT name FROM users;`, `SELECT name FROM users;`},
		{"lowercase keywords", `select name from users;`, `SELECT name FROM users;`},
		{"mixed case keywords", `SeLeCt name FrOm users;`, `SELECT name FROM users;`},
		{"star", `SELECT * FROM users;`, `SELECT * FROM users;`},
		{"multiple columns", `SELECT id, name FROM users;`, `SELECT id, name FROM users;`},
		{"no spaces around commas", `SELECT id,name,email FROM users;`, `SELECT id, name, email FROM users;`},
		{"messy whitespace", "SELECT   id ,\n\tname\nFROM   users ;", `SELECT id, name FROM users;`},

		// --- comments are invisible to the parser ---
		{"comment before statement", "-- fetch names\nSELECT name FROM users;", `SELECT name FROM users;`},
		{"comment mid statement", "SELECT name -- the name\nFROM users;", `SELECT name FROM users;`},
		{"comment after statement", "SELECT name FROM users; -- done", `SELECT name FROM users;`},

		// --- WHERE: single comparisons ---
		{"where string", `SELECT name FROM users WHERE a='hello';`, `SELECT name FROM users WHERE (a = 'hello');`},
		{"where number", `SELECT name FROM users WHERE age > 18;`, `SELECT name FROM users WHERE (age > 18);`},
		{"where decimal", `SELECT name FROM users WHERE x < 5.3;`, `SELECT name FROM users WHERE (x < 5.3);`},
		{"where two columns", `SELECT name FROM users WHERE a = b;`, `SELECT name FROM users WHERE (a = b);`},
		{"where param", `SELECT name FROM users WHERE id = $1;`, `SELECT name FROM users WHERE (id = $1);`},
		{"op <=", `SELECT a FROM t WHERE a<=1;`, `SELECT a FROM t WHERE (a <= 1);`},
		{"op >=", `SELECT a FROM t WHERE a>=1;`, `SELECT a FROM t WHERE (a >= 1);`},
		{"op <>", `SELECT a FROM t WHERE a<>1;`, `SELECT a FROM t WHERE (a <> 1);`},
		{"op !=", `SELECT a FROM t WHERE a!=1;`, `SELECT a FROM t WHERE (a != 1);`},
		{"string with spaces", `SELECT a FROM t WHERE a = 'hello world';`, `SELECT a FROM t WHERE (a = 'hello world');`},
		{"escaped quote", `SELECT a FROM t WHERE a = 'O''Brien';`, `SELECT a FROM t WHERE (a = 'O''Brien');`},
		{"bare column", `SELECT a FROM t WHERE active;`, `SELECT a FROM t WHERE active;`},

		// --- WHERE: precedence and associativity ---
		{
			"AND binds tighter than OR (OR first)",
			`SELECT a FROM t WHERE a = 1 OR b = 2 AND c = 3;`,
			`SELECT a FROM t WHERE ((a = 1) OR ((b = 2) AND (c = 3)));`,
		},
		{
			"AND binds tighter than OR (AND first)",
			`SELECT a FROM t WHERE a = 1 AND b = 2 OR c = 3;`,
			`SELECT a FROM t WHERE (((a = 1) AND (b = 2)) OR (c = 3));`,
		},
		{
			"OR chain groups left",
			`SELECT a FROM t WHERE a = 1 OR b = 2 OR c = 3;`,
			`SELECT a FROM t WHERE (((a = 1) OR (b = 2)) OR (c = 3));`,
		},
		{
			"AND chain groups left",
			`SELECT a FROM t WHERE a = 1 AND b = 2 AND c = 3;`,
			`SELECT a FROM t WHERE (((a = 1) AND (b = 2)) AND (c = 3));`,
		},
		{
			"OR AND OR",
			`SELECT a FROM t WHERE a = 1 OR b = 2 AND c = 3 OR d = 4;`,
			`SELECT a FROM t WHERE (((a = 1) OR ((b = 2) AND (c = 3))) OR (d = 4));`,
		},

		// --- WHERE: parentheses ---
		{
			"parens override precedence (left)",
			`SELECT a FROM t WHERE (a = 1 OR b = 2) AND c = 3;`,
			`SELECT a FROM t WHERE (((a = 1) OR (b = 2)) AND (c = 3));`,
		},
		{
			"parens override precedence (right)",
			`SELECT a FROM t WHERE a = 1 AND (b = 2 OR c = 3);`,
			`SELECT a FROM t WHERE ((a = 1) AND ((b = 2) OR (c = 3)));`,
		},
		{"redundant parens vanish", `SELECT a FROM t WHERE ((a = 1));`, `SELECT a FROM t WHERE (a = 1);`},
		{
			"nested groups",
			`SELECT a FROM t WHERE ((a = 1 OR b = 2) AND (c = 3 OR d = 4));`,
			`SELECT a FROM t WHERE (((a = 1) OR (b = 2)) AND ((c = 3) OR (d = 4)));`,
		},

		// --- CREATE TABLE ---
		{
			"create one column",
			`CREATE TABLE users (name TEXT);`,
			`CREATE TABLE users ( name TEXT );`,
		},
		{
			"create multiline with tabs",
			"CREATE TABLE users (\n\tname TEXT\n\t);",
			`CREATE TABLE users ( name TEXT );`,
		},
		{
			"create with constraints",
			`CREATE TABLE users (
				id INT NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				city TEXT DEFAULT Toronto
			);`,
			`CREATE TABLE users ( id INT NOT NULL PRIMARY KEY, name TEXT NOT NULL, city TEXT DEFAULT Toronto );`,
		},
		{
			"create if not exists",
			`CREATE TABLE IF NOT EXISTS users (name TEXT);`,
			`CREATE TABLE IF NOT EXISTS users ( name TEXT );`,
		},

		// --- INSERT ---
		{
			"insert with columns",
			`INSERT INTO users (id, name) VALUES (1, 'bob');`,
			`INSERT INTO users (id, name) VALUES (1, 'bob');`,
		},
		{
			"insert without columns",
			`INSERT INTO users VALUES (1, 'bob');`,
			`INSERT INTO users VALUES (1, 'bob');`,
		},
		{
			"insert single value",
			`INSERT INTO t (a) VALUES (1);`,
			`INSERT INTO t (a) VALUES (1);`,
		},
		{
			"insert params",
			`INSERT INTO users (id, name) VALUES ($1, $2);`,
			`INSERT INTO users (id, name) VALUES ($1, $2);`,
		},
		{
			"insert mixed literals and params",
			`INSERT INTO t (a, b, c) VALUES ($1, 5.3, 'x');`,
			`INSERT INTO t (a, b, c) VALUES ($1, 5.3, 'x');`,
		},
		{
			"insert lowercase and no spaces",
			`insert into t(a,b)values(1,2);`,
			`INSERT INTO t (a, b) VALUES (1, 2);`,
		},
		{
			"insert escaped quote",
			`INSERT INTO t (name) VALUES ('O''Brien');`,
			`INSERT INTO t (name) VALUES ('O''Brien');`,
		},
		{
			"insert multiline",
			"INSERT INTO users (\n\tid,\n\tname\n) VALUES (\n\t1,\n\t'bob'\n);",
			`INSERT INTO users (id, name) VALUES (1, 'bob');`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v\ninput: %s", err, tt.input)
			}
			got := normalize(stmt.String())
			if got != normalize(tt.want) {
				t.Errorf("\ninput: %s\n got: %s\nwant: %s", tt.input, got, normalize(tt.want))
			}
		})
	}
}

/* ---------- inputs that must be rejected ---------- */

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// --- general ---
		{"empty", ``},
		{"whitespace only", "   \n\t "},
		{"comment only", `-- nothing here`},
		{"missing semicolon", `SELECT name FROM users`},
		{"unknown statement", `DROP TABLE users;`},
		{"trailing junk after semicolon", `SELECT name FROM users; junk`},
		{"lex error surfaces", `SELECT @ FROM users;`},
		{"annotation not allowed in Parse", "-- name: GetUser :one\nSELECT name FROM users;"},

		// --- SELECT list ---
		{"no columns", `SELECT FROM users;`},
		{"leading comma", `SELECT , name FROM users;`},
		{"trailing comma", `SELECT name, FROM users;`},
		{"double comma", `SELECT a,, b FROM users;`},
		{"missing comma", `SELECT a b FROM users;`},

		// --- FROM ---
		{"missing FROM", `SELECT name users;`},
		{"missing table", `SELECT name FROM;`},
		{"two tables", `SELECT name FROM users, orders;`},

		// --- WHERE ---
		{"empty where", `SELECT a FROM t WHERE;`},
		{"missing right operand", `SELECT a FROM t WHERE a =;`},
		{"missing left operand", `SELECT a FROM t WHERE = 1;`},
		{"double comparator", `SELECT a FROM t WHERE a = = 1;`},
		{"dangling AND", `SELECT a FROM t WHERE a = 1 AND;`},
		{"dangling OR", `SELECT a FROM t WHERE a = 1 OR;`},
		{"leading AND", `SELECT a FROM t WHERE AND a = 1;`},
		{"unclosed paren", `SELECT a FROM t WHERE (a = 1;`},
		{"extra close paren", `SELECT a FROM t WHERE a = 1);`},
		{"empty parens", `SELECT a FROM t WHERE ();`},

		// --- CREATE TABLE ---
		{"create no parens", `CREATE TABLE users name TEXT;`},
		{"create empty columns", `CREATE TABLE users ();`},
		{"create trailing comma", `CREATE TABLE users (name TEXT,);`},
		{"create missing type", `CREATE TABLE users (name);`},
		{"create missing comma", `CREATE TABLE users (a TEXT b TEXT);`},
		{"create NOT without NULL", `CREATE TABLE users (a TEXT NOT);`},
		{"create PRIMARY without KEY", `CREATE TABLE users (a TEXT PRIMARY);`},
		{"create unclosed", `CREATE TABLE users (a TEXT;`},
		{"create missing name", `CREATE TABLE (a TEXT);`},
		{"create without TABLE", `CREATE users (a TEXT);`},
		{"create IF without NOT EXISTS", `CREATE TABLE IF users (a TEXT);`},

		// --- INSERT ---
		{"insert missing INTO", `INSERT users VALUES (1);`},
		{"insert missing table", `INSERT INTO VALUES (1);`},
		{"insert missing VALUES", `INSERT INTO users (id) (1);`},
		{"insert missing values list", `INSERT INTO users (id) VALUES;`},
		{"insert values without parens", `INSERT INTO users VALUES 1, 2;`},
		{"insert empty values", `INSERT INTO users VALUES ();`},
		{"insert empty column list", `INSERT INTO users () VALUES (1);`},
		{"insert trailing comma in columns", `INSERT INTO users (id,) VALUES (1);`},
		{"insert trailing comma in values", `INSERT INTO users (id) VALUES (1,);`},
		{"insert missing comma in columns", `INSERT INTO users (id name) VALUES (1, 2);`},
		{"insert missing comma in values", `INSERT INTO users (id, name) VALUES (1 2);`},
		{"insert unclosed columns", `INSERT INTO users (id VALUES (1);`},
		{"insert unclosed values", `INSERT INTO users (id) VALUES (1;`},
		{"insert missing semicolon", `INSERT INTO users VALUES (1)`},
		{"insert trailing junk", `INSERT INTO users VALUES (1); junk`},
		{"insert star in columns", `INSERT INTO users (*) VALUES (1);`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := Parse(tt.input)
			if err == nil {
				t.Errorf("expected an error, got none\ninput: %s\nparsed as: %v", tt.input, stmt)
			}
		})
	}
}

/* ---------- annotated query files ---------- */

type wantQuery struct {
	name        string
	cardinality Cardinality
	sql         string // expected String() of the statement, whitespace-insensitive
}

func TestParseQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []wantQuery
	}{
		// --- empty-ish files ---
		{"empty file", ``, nil},
		{"whitespace only", "  \n\t\n  ", nil},
		{"comments only", "-- just notes\n-- nothing to see", nil},

		// --- one query ---
		{
			"single :many select",
			"-- name: ListUsers :many\nSELECT * FROM users;",
			[]wantQuery{{"ListUsers", Many, `SELECT * FROM users;`}},
		},
		{
			"single :one select with param",
			"-- name: GetUser :one\nSELECT id, name FROM users WHERE id = $1;",
			[]wantQuery{{"GetUser", One, `SELECT id, name FROM users WHERE (id = $1);`}},
		},
		{
			"single :exec insert",
			"-- name: CreateUser :exec\nINSERT INTO users (name, bio) VALUES ($1, $2);",
			[]wantQuery{{"CreateUser", Exec, `INSERT INTO users (name, bio) VALUES ($1, $2);`}},
		},
		{
			"annotation with extra spacing",
			"--    name:   GetUser    :one   \nSELECT id FROM users;",
			[]wantQuery{{"GetUser", One, `SELECT id FROM users;`}},
		},
		{
			"statement on the same line count doesn't matter",
			"-- name: GetUser :one\n\n\n\nSELECT id\nFROM users\nWHERE id = $1;",
			[]wantQuery{{"GetUser", One, `SELECT id FROM users WHERE (id = $1);`}},
		},
		{
			"name keeps case, digits and underscores",
			"-- name: get_User2 :one\nSELECT id FROM users;",
			[]wantQuery{{"get_User2", One, `SELECT id FROM users;`}},
		},
		{
			"no trailing newline after statement",
			"-- name: A :many\nSELECT a FROM t;",
			[]wantQuery{{"A", Many, `SELECT a FROM t;`}},
		},

		// --- many queries ---
		{
			"two queries keep order",
			`-- name: GetUser :one
SELECT id, name FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users;`,
			[]wantQuery{
				{"GetUser", One, `SELECT id, name FROM users WHERE (id = $1);`},
				{"ListUsers", Many, `SELECT * FROM users;`},
			},
		},
		{
			"all three cardinalities",
			`-- name: GetUser :one
SELECT id FROM users WHERE id = $1;
-- name: ListUsers :many
SELECT id FROM users;
-- name: CreateUser :exec
INSERT INTO users (id) VALUES ($1);`,
			[]wantQuery{
				{"GetUser", One, `SELECT id FROM users WHERE (id = $1);`},
				{"ListUsers", Many, `SELECT id FROM users;`},
				{"CreateUser", Exec, `INSERT INTO users (id) VALUES ($1);`},
			},
		},
		{
			"no blank lines between queries",
			"-- name: A :one\nSELECT a FROM t;\n-- name: B :one\nSELECT b FROM t;",
			[]wantQuery{
				{"A", One, `SELECT a FROM t;`},
				{"B", One, `SELECT b FROM t;`},
			},
		},
		{
			"next annotation on same line as previous semicolon",
			"-- name: A :one\nSELECT a FROM t; -- name: B :one\nSELECT b FROM t;",
			[]wantQuery{
				{"A", One, `SELECT a FROM t;`},
				{"B", One, `SELECT b FROM t;`},
			},
		},
		{
			"same statement twice under different names",
			"-- name: A :many\nSELECT a FROM t;\n-- name: B :many\nSELECT a FROM t;",
			[]wantQuery{
				{"A", Many, `SELECT a FROM t;`},
				{"B", Many, `SELECT a FROM t;`},
			},
		},
		{
			"five queries",
			`-- name: Q1 :one
SELECT a FROM t WHERE a = $1;
-- name: Q2 :many
SELECT * FROM t;
-- name: Q3 :exec
INSERT INTO t (a) VALUES ($1);
-- name: Q4 :many
SELECT a, b FROM t WHERE a = $1 AND b = $2;
-- name: Q5 :exec
INSERT INTO t VALUES (1, 'x');`,
			[]wantQuery{
				{"Q1", One, `SELECT a FROM t WHERE (a = $1);`},
				{"Q2", Many, `SELECT * FROM t;`},
				{"Q3", Exec, `INSERT INTO t (a) VALUES ($1);`},
				{"Q4", Many, `SELECT a, b FROM t WHERE ((a = $1) AND (b = $2));`},
				{"Q5", Exec, `INSERT INTO t VALUES (1, 'x');`},
			},
		},

		// --- ordinary comments around queries ---
		{
			"comment above annotation",
			"-- Fetches a single user by id.\n-- name: GetUser :one\nSELECT id FROM users WHERE id = $1;",
			[]wantQuery{{"GetUser", One, `SELECT id FROM users WHERE (id = $1);`}},
		},
		{
			"comment between annotation and statement",
			"-- name: GetUser :one\n-- uses the primary key\nSELECT id FROM users WHERE id = $1;",
			[]wantQuery{{"GetUser", One, `SELECT id FROM users WHERE (id = $1);`}},
		},
		{
			"comments inside statement",
			"-- name: GetUser :one\nSELECT id, -- the id\n  name -- the name\nFROM users\nWHERE id = $1; -- pk lookup",
			[]wantQuery{{"GetUser", One, `SELECT id, name FROM users WHERE (id = $1);`}},
		},
		{
			"comments between queries",
			"-- name: A :one\nSELECT a FROM t;\n\n-- ---------- section two ----------\n\n-- name: B :many\nSELECT b FROM t;",
			[]wantQuery{
				{"A", One, `SELECT a FROM t;`},
				{"B", Many, `SELECT b FROM t;`},
			},
		},
		{
			"trailing comment at end of file",
			"-- name: A :one\nSELECT a FROM t;\n-- end of file",
			[]wantQuery{{"A", One, `SELECT a FROM t;`}},
		},
		{
			"dashes inside a string literal",
			"-- name: A :many\nSELECT a FROM t WHERE a = 'x -- y';",
			[]wantQuery{{"A", Many, `SELECT a FROM t WHERE (a = 'x -- y');`}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQueries(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v\ninput:\n%s", err, tt.input)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d queries, want %d", len(got), len(tt.want))
			}
			for i, w := range tt.want {
				q := got[i]
				if q.Name != w.name {
					t.Errorf("query %d: Name = %q, want %q", i, q.Name, w.name)
				}
				if q.Cardinality != w.cardinality {
					t.Errorf("query %d: Cardinality = %v, want %v", i, q.Cardinality, w.cardinality)
				}
				if q.Statement == nil {
					t.Fatalf("query %d: Statement is nil", i)
				}
				if gotSQL := normalize(q.Statement.String()); gotSQL != normalize(w.sql) {
					t.Errorf("query %d: Statement\n got: %s\nwant: %s", i, gotSQL, normalize(w.sql))
				}
			}
		})
	}
}

func TestParseQueriesStatementTypes(t *testing.T) {
	got, err := ParseQueries(`-- name: A :many
SELECT a FROM t;
-- name: B :exec
INSERT INTO t (a) VALUES ($1);`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d queries, want 2", len(got))
	}
	if _, ok := got[0].Statement.(*SelectStatement); !ok {
		t.Errorf("query 0: got %T, want *SelectStatement", got[0].Statement)
	}
	if _, ok := got[1].Statement.(*InsertIntoStatement); !ok {
		t.Errorf("query 1: got %T, want *InsertIntoStatement", got[1].Statement)
	}
}

func TestParseQueriesErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string // case-insensitive; empty means any error is fine
	}{
		// --- missing annotations ---
		{"statement without annotation", `SELECT a FROM t;`, ""},
		{"second statement without annotation", "-- name: A :one\nSELECT a FROM t;\nSELECT b FROM t;", ""},
		{"comment is not an annotation", "-- get all users\nSELECT a FROM t;", ""},

		// --- annotations without statements ---
		{"annotation alone", "-- name: A :one", ""},
		{"annotation alone with newline", "-- name: A :one\n", ""},
		{"annotation at end after a valid query", "-- name: A :one\nSELECT a FROM t;\n-- name: B :many", ""},
		{"two annotations in a row", "-- name: A :one\n-- name: B :one\nSELECT a FROM t;", ""},

		// --- duplicate names ---
		{"duplicate name", "-- name: A :one\nSELECT a FROM t;\n-- name: A :many\nSELECT b FROM t;", "A"},
		{"duplicate name not adjacent", "-- name: A :one\nSELECT a FROM t;\n-- name: B :one\nSELECT b FROM t;\n-- name: A :one\nSELECT c FROM t;", "A"},
		{"duplicate name same cardinality", "-- name: Dup :exec\nINSERT INTO t VALUES (1);\n-- name: Dup :exec\nINSERT INTO t VALUES (2);", "Dup"},

		// --- broken statements ---
		{"missing semicolon", "-- name: A :one\nSELECT a FROM t", ""},
		{"missing semicolon before next annotation", "-- name: A :one\nSELECT a FROM t\n-- name: B :one\nSELECT b FROM t;", ""},
		{"parse error in statement", "-- name: A :one\nSELECT FROM t;", ""},
		{"parse error in second statement", "-- name: A :one\nSELECT a FROM t;\n-- name: B :one\nSELECT a b FROM t;", ""},
		{"unsupported statement", "-- name: A :exec\nDROP TABLE t;", ""},
		{"two statements under one annotation", "-- name: A :many\nSELECT a FROM t;\nSELECT b FROM t;", ""},
		{"junk after last statement", "-- name: A :one\nSELECT a FROM t;\njunk", ""},

		// --- lexer errors surface ---
		{"malformed annotation", "-- name: A\nSELECT a FROM t;", ""},
		{"unknown cardinality", "-- name: A :all\nSELECT a FROM t;", ""},
		{"bad character in name", "-- name: Get-User :one\nSELECT a FROM t;", ""},
		{"lex error in statement", "-- name: A :one\nSELECT @ FROM t;", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQueries(tt.input)
			if err == nil {
				t.Fatalf("expected an error, got none\ninput:\n%s\nparsed %d queries", tt.input, len(got))
			}
			if tt.errContains != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errContains)) {
				t.Errorf("error %q should mention %q", err, tt.errContains)
			}
		})
	}
}

/* ---------- Cardinality ---------- */

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
	}
	for _, tt := range tests {
		if got := tt.c.String(); got != tt.want {
			t.Errorf("Cardinality(%d).String() = %q, want %q", int(tt.c), got, tt.want)
		}
	}
}

/* ---------- UPDATE ---------- */

func parseUpdate(t *testing.T, q string) *UpdateStatement {
	t.Helper()
	stmt, err := Parse(q)
	if err != nil {
		t.Fatalf("parse %q: %v", q, err)
	}
	u, ok := stmt.(*UpdateStatement)
	if !ok {
		t.Fatalf("got %T, want *UpdateStatement", stmt)
	}
	return u
}

func TestParseUpdate(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		table     string
		columns   []string
		wantWhere bool
	}{
		{"one pair with where", "UPDATE users SET name = $1 WHERE id = $2;", "users", []string{"name"}, true},
		{"several pairs", "UPDATE users SET name = $1, bio = $2, active = TRUE WHERE id = $3;", "users", []string{"name", "bio", "active"}, true},
		{"no where", "UPDATE users SET active = FALSE;", "users", []string{"active"}, false},
		{"lowercase keywords", "update users set name = 'x' where id = 1;", "users", []string{"name"}, true},
		{"compound where", "UPDATE users SET bio = NULL WHERE id = $1 AND (name = $2 OR name = $3);", "users", []string{"bio"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := parseUpdate(t, tt.query)
			if u.TableName != tt.table {
				t.Errorf("TableName = %q, want %q", u.TableName, tt.table)
			}
			var got []string
			for _, cv := range u.CvPairs {
				got = append(got, cv.ColumnRef.Name)
			}
			if len(got) != len(tt.columns) {
				t.Fatalf("columns = %v, want %v", got, tt.columns)
			}
			for i := range got {
				if got[i] != tt.columns[i] {
					t.Errorf("columns = %v, want %v", got, tt.columns)
					break
				}
			}
			if (u.Where != nil) != tt.wantWhere {
				t.Errorf("Where = %v, want present=%v", u.Where, tt.wantWhere)
			}
		})
	}
}

func TestParseUpdateValues(t *testing.T) {
	u := parseUpdate(t, "UPDATE users SET a = $1, b = 'x', c = 5, d = TRUE, e = NULL, f = g;")
	want := []string{"*parsing.Parameter", "*parsing.Literal", "*parsing.Literal", "*parsing.Literal", "*parsing.Literal", "*parsing.ColumnRef"}
	if len(u.CvPairs) != len(want) {
		t.Fatalf("got %d pairs, want %d", len(u.CvPairs), len(want))
	}
	for i, cv := range u.CvPairs {
		if got := typeName(cv.Value); got != want[i] {
			t.Errorf("pair %d (%s): value is %s, want %s", i, cv.ColumnRef.Name, got, want[i])
		}
	}
}

func typeName(e Expr) string {
	switch e.(type) {
	case *Parameter:
		return "*parsing.Parameter"
	case *Literal:
		return "*parsing.Literal"
	case *ColumnRef:
		return "*parsing.ColumnRef"
	case *BinaryExpr:
		return "*parsing.BinaryExpr"
	default:
		return "unknown"
	}
}

// The printed SQL is what ends up in the generated code, so it has to be valid.
func TestUpdateRoundTrip(t *testing.T) {
	tests := []struct{ query, want string }{
		{
			"UPDATE users SET name = $1 WHERE id = $2;",
			"UPDATE users SET name = $1 WHERE (id = $2);",
		},
		{
			"UPDATE users SET name = $1, bio = $2 WHERE id = $3;",
			"UPDATE users SET name = $1, bio = $2 WHERE (id = $3);",
		},
		{
			"UPDATE users SET active = TRUE;",
			"UPDATE users SET active = TRUE;",
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := normalize(parseUpdate(t, tt.query).String())
			if got != tt.want {
				t.Errorf("\n got: %s\nwant: %s", got, tt.want)
			}
			// printing must be stable: parse what we printed and print again
			if again := normalize(parseUpdate(t, got).String()); again != got {
				t.Errorf("not stable:\n first: %s\nsecond: %s", got, again)
			}
		})
	}
}

func TestParseUpdateErrors(t *testing.T) {
	tests := []struct{ name, query string }{
		{"no table", "UPDATE SET name = $1;"},
		{"no SET", "UPDATE users name = $1;"},
		{"no pairs", "UPDATE users SET;"},
		{"pair without value", "UPDATE users SET name =;"},
		{"pair without =", "UPDATE users SET name $1;"},
		{"non-= comparator", "UPDATE users SET name < $1;"},
		{"trailing comma", "UPDATE users SET name = $1, WHERE id = $2;"},
		{"missing semicolon", "UPDATE users SET name = $1 WHERE id = $2"},
		{"empty WHERE", "UPDATE users SET name = $1 WHERE;"},
		{"junk after semicolon", "UPDATE users SET name = $1; garbage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.query); err == nil {
				t.Errorf("expected an error for %q", tt.query)
			}
		})
	}
}

// Several UPDATEs in one file: each must stop at its own semicolon.
func TestParseQueriesWithUpdates(t *testing.T) {
	qs, err := ParseQueries(`-- name: RenameUser :exec
UPDATE users SET name = $1 WHERE id = $2;

-- name: DeactivateAll :exec
UPDATE users SET active = FALSE;

-- name: GetUser :one
SELECT id FROM users WHERE id = $1;
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(qs) != 3 {
		t.Fatalf("got %d queries, want 3", len(qs))
	}
	for i, want := range []string{"RenameUser", "DeactivateAll"} {
		if qs[i].Name != want {
			t.Errorf("query %d: Name = %q, want %q", i, qs[i].Name, want)
		}
		if _, ok := qs[i].Statement.(*UpdateStatement); !ok {
			t.Errorf("query %d: got %T, want *UpdateStatement", i, qs[i].Statement)
		}
	}
}

/* ---------- DELETE ---------- */

func parseDelete(t *testing.T, q string) *DeleteFromStatement {
	t.Helper()
	stmt, err := Parse(q)
	if err != nil {
		t.Fatalf("parse %q: %v", q, err)
	}
	d, ok := stmt.(*DeleteFromStatement)
	if !ok {
		t.Fatalf("got %T, want *DeleteFromStatement", stmt)
	}
	return d
}

func TestParseDelete(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantWhere bool
	}{
		{"with where", "DELETE FROM users WHERE id = $1;", true},
		{"without where", "DELETE FROM users;", false},
		{"lowercase keywords", "delete from users where id = 1;", true},
		{"compound where", "DELETE FROM users WHERE active = FALSE AND (id = $1 OR name = $2);", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := parseDelete(t, tt.query)
			if d.TableName != "users" {
				t.Errorf("TableName = %q, want users", d.TableName)
			}
			if (d.Where != nil) != tt.wantWhere {
				t.Errorf("Where = %v, want present=%v", d.Where, tt.wantWhere)
			}
		})
	}
}

// The printed SQL is what ends up in the generated code, so it has to be valid.
func TestDeleteRoundTrip(t *testing.T) {
	tests := []struct{ query, want string }{
		{"DELETE FROM users WHERE id = $1;", "DELETE FROM users WHERE (id = $1);"},
		{"DELETE FROM users;", "DELETE FROM users;"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := normalize(parseDelete(t, tt.query).String())
			if got != tt.want {
				t.Errorf("\n got: %s\nwant: %s", got, tt.want)
			}
			if again := normalize(parseDelete(t, got).String()); again != got {
				t.Errorf("not stable:\n first: %s\nsecond: %s", got, again)
			}
		})
	}
}

func TestParseDeleteErrors(t *testing.T) {
	tests := []struct{ name, query string }{
		{"missing FROM", "DELETE users WHERE id = $1;"},
		{"no table", "DELETE FROM WHERE id = $1;"},
		{"no table, no where", "DELETE FROM;"},
		{"columns listed", "DELETE id FROM users;"},
		{"empty WHERE", "DELETE FROM users WHERE;"},
		{"missing semicolon", "DELETE FROM users WHERE id = $1"},
		{"junk after table", "DELETE FROM users garbage;"},
		{"junk after semicolon", "DELETE FROM users; garbage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.query); err == nil {
				t.Errorf("expected an error for %q", tt.query)
			}
		})
	}
}

// Several DELETEs in one file: each must stop at its own semicolon.
func TestParseQueriesWithDeletes(t *testing.T) {
	qs, err := ParseQueries(`-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: DeleteAll :exec
DELETE FROM users;

-- name: GetUser :one
SELECT id FROM users WHERE id = $1;
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(qs) != 3 {
		t.Fatalf("got %d queries, want 3", len(qs))
	}
	for i, want := range []string{"DeleteUser", "DeleteAll"} {
		if qs[i].Name != want {
			t.Errorf("query %d: Name = %q, want %q", i, qs[i].Name, want)
		}
		if _, ok := qs[i].Statement.(*DeleteFromStatement); !ok {
			t.Errorf("query %d: got %T, want *DeleteFromStatement", i, qs[i].Statement)
		}
	}
}
