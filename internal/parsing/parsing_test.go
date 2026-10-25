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
		{"missing semicolon", `SELECT name FROM users`},
		{"unknown statement", `DELETE FROM users;`},
		{"trailing junk after semicolon", `SELECT name FROM users; junk`},
		{"lex error surfaces", `SELECT @ FROM users;`},

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
