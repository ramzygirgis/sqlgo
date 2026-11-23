package lexing

import (
	"strings"
	"testing"
)

/* ---------- token kinds ---------- */

func TestLexKinds(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenKind
	}{
		{"keywords and idents", `SELECT name FROM users;`,
			[]TokenKind{Keyword, Identifier, Keyword, Identifier, Semicolon, EOF}},
		{"comma without spaces", `a,b`,
			[]TokenKind{Identifier, Comma, Identifier, EOF}},
		{"star", `*`,
			[]TokenKind{Star, EOF}},
		{"parens", `(a)`,
			[]TokenKind{LParen, Identifier, RParen, EOF}},
		{"<= is one token", `a<=b`,
			[]TokenKind{Identifier, Comparator, Identifier, EOF}},
		{"<> is one token", `a<>b`,
			[]TokenKind{Identifier, Comparator, Identifier, EOF}},
		{"!= is one token", `a!=b`,
			[]TokenKind{Identifier, Comparator, Identifier, EOF}},
		{"string with spaces is one token", `'hello world'`,
			[]TokenKind{String, EOF}},
		{"escaped quote stays in one string", `'O''Brien'`,
			[]TokenKind{String, EOF}},
		{"decimal number", `5.3`,
			[]TokenKind{Number, EOF}},
		{"param", `$1`,
			[]TokenKind{Param, EOF}},
		{"identifier with digits", `col2`,
			[]TokenKind{Identifier, EOF}},
		{"underscore identifier", `_private`,
			[]TokenKind{Identifier, EOF}},
		{"tabs and newlines", "a\t\n\r b",
			[]TokenKind{Identifier, Identifier, EOF}},
		{"empty input", ``,
			[]TokenKind{EOF}},
		{"TRUE FALSE NULL are keywords", `TRUE false Null`,
			[]TokenKind{Keyword, Keyword, Keyword, EOF}},
		{"identifiers that merely contain TRUE or NULL", `is_true nullable truest`,
			[]TokenKind{Identifier, Identifier, Identifier, EOF}},
		{"boolean comparison", `active = true`,
			[]TokenKind{Identifier, Comparator, Keyword, EOF}},

		// --- INSERT ---
		{"insert with columns", `INSERT INTO users (id, name) VALUES (1, 'bob');`,
			[]TokenKind{
				Keyword, Keyword, Identifier,
				LParen, Identifier, Comma, Identifier, RParen,
				Keyword,
				LParen, Number, Comma, String, RParen,
				Semicolon, EOF,
			}},
		{"insert without columns", `INSERT INTO users VALUES ($1, $2);`,
			[]TokenKind{
				Keyword, Keyword, Identifier, Keyword,
				LParen, Param, Comma, Param, RParen,
				Semicolon, EOF,
			}},
		{"insert no spaces", `INSERT INTO t(a,b)VALUES(1,2);`,
			[]TokenKind{
				Keyword, Keyword, Identifier,
				LParen, Identifier, Comma, Identifier, RParen,
				Keyword,
				LParen, Number, Comma, Number, RParen,
				Semicolon, EOF,
			}},

		// --- ordinary comments are skipped ---
		{"comment only", `-- just a note`,
			[]TokenKind{EOF}},
		{"empty comment", `--`,
			[]TokenKind{EOF}},
		{"comment then query", "-- fetch users\nSELECT a FROM t;",
			[]TokenKind{Keyword, Identifier, Keyword, Identifier, Semicolon, EOF}},
		{"comment at end of line mid-query", "SELECT a -- the a column\nFROM t;",
			[]TokenKind{Keyword, Identifier, Keyword, Identifier, Semicolon, EOF}},
		{"comment at EOF without newline", "SELECT a FROM t; -- done",
			[]TokenKind{Keyword, Identifier, Keyword, Identifier, Semicolon, EOF}},
		{"several comments in a row", "-- one\n-- two\n-- three\na",
			[]TokenKind{Identifier, EOF}},
		{"comment mentioning name later", "-- this is not a name: annotation\na",
			[]TokenKind{Identifier, EOF}},
		{"dashes inside a string are not a comment", `'a -- b'`,
			[]TokenKind{String, EOF}},

		// --- annotations ---
		{"one annotation", "-- name: GetUser :one\nSELECT a FROM t;",
			[]TokenKind{OneAnnotation, Keyword, Identifier, Keyword, Identifier, Semicolon, EOF}},
		{"many annotation", "-- name: ListUsers :many\nSELECT a FROM t;",
			[]TokenKind{ManyAnnotation, Keyword, Identifier, Keyword, Identifier, Semicolon, EOF}},
		{"exec annotation", "-- name: CreateUser :exec\nINSERT INTO t VALUES (1);",
			[]TokenKind{
				ExecAnnotation, Keyword, Keyword, Identifier, Keyword,
				LParen, Number, RParen, Semicolon, EOF,
			}},
		{"annotation with extra spacing", "--   name:   GetUser   :one  \na",
			[]TokenKind{OneAnnotation, Identifier, EOF}},
		{"annotation at EOF without newline", "-- name: GetUser :one",
			[]TokenKind{OneAnnotation, EOF}},
		{"comment before annotation", "-- fetches a user\n-- name: GetUser :one\na",
			[]TokenKind{OneAnnotation, Identifier, EOF}},
		{"two annotated queries", "-- name: A :one\nSELECT a FROM t;\n\n-- name: B :many\nSELECT b FROM t;",
			[]TokenKind{
				OneAnnotation, Keyword, Identifier, Keyword, Identifier, Semicolon,
				ManyAnnotation, Keyword, Identifier, Keyword, Identifier, Semicolon,
				EOF,
			}},
		{"annotation not at start of file", "SELECT a FROM t;\n-- name: B :one\nb",
			[]TokenKind{Keyword, Identifier, Keyword, Identifier, Semicolon, OneAnnotation, Identifier, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(toks) != len(tt.want) {
				t.Fatalf("got %d tokens %v, want %d kinds %v", len(toks), toks, len(tt.want), tt.want)
			}
			for i, tok := range toks {
				if tok.Kind != tt.want[i] {
					t.Errorf("token %d (%q): got kind %v, want %v", i, tok.Body, tok.Kind, tt.want[i])
				}
			}
		})
	}
}

/* ---------- token bodies ---------- */

func TestLexBodies(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string // bodies, excluding EOF
	}{
		{"keywords uppercased", `select FROM wHeRe`, []string{"SELECT", "FROM", "WHERE"}},
		{"identifiers keep case", `userName`, []string{"userName"}},
		{"two-char operators", `<= >= <> != = < >`, []string{"<=", ">=", "<>", "!=", "=", "<", ">"}},
		{"string keeps quotes", `'hi there'`, []string{"'hi there'"}},
		{"escaped quote kept raw", `'O''Brien'`, []string{"'O''Brien'"}},
		{"param keeps dollar", `$12`, []string{"$12"}},
		{"insert keywords uppercased", `insert into t values`, []string{"INSERT", "INTO", "t", "VALUES"}},
		{"literal keywords uppercased", `true FALSE nUlL`, []string{"TRUE", "FALSE", "NULL"}},
		{"delete keywords uppercased", `delete from t`, []string{"DELETE", "FROM", "t"}},

		// --- annotations: body is the function name ---
		{"annotation body is the name", "-- name: GetUser :one", []string{"GetUser"}},
		{"annotation name keeps case", "-- name: listAllUsers :many", []string{"listAllUsers"}},
		{"annotation name with digits and underscore", "-- name: Get_User2 :exec", []string{"Get_User2"}},
		{"two annotations", "-- name: A :one\na;\n-- name: B :many\nb;", []string{"A", "a", ";", "B", "b", ";"}},
		{"comment body is not emitted", "-- hello world\nx", []string{"x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var got []string
			for _, tok := range toks {
				if tok.Kind != EOF {
					got = append(got, tok.Body)
				}
			}
			if strings.Join(got, "|") != strings.Join(tt.want, "|") {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

/* ---------- inputs the lexer must reject ---------- */

func TestLexErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unexpected character", `SELECT @ FROM users;`},
		{"unterminated string", `'oops`},
		{"unterminated after escape", `'it''s`},
		{"two decimal points", `1.2.3`},
		{"letters glued to number", `123abc`},
		{"bare dollar", `$`},
		{"letters glued to param", `$1abc`},
		{"lone bang", `!`},
		{"non-ascii outside string", `café`},

		// --- dashes ---
		{"single dash", `-`},
		{"single dash before word", `-a`},

		// --- malformed annotations ---
		{"annotation missing cardinality", "-- name: GetUser\na"},
		{"annotation missing name", "-- name: :one\na"},
		{"annotation with only name:", "-- name:\na"},
		{"annotation unknown cardinality", "-- name: GetUser :all\na"},
		{"annotation cardinality without colon", "-- name: GetUser one\na"},
		{"annotation extra field", "-- name: GetUser :one extra\na"},
		{"annotation name with dash", "-- name: Get-User :one\na"},
		{"annotation name with dot", "-- name: users.Get :one\na"},
		{"annotation name starts with digit", "-- name: 2GetUser :one\na"},
		// Fails until "is this an annotation?" is decided by a `name:` prefix
		// rather than an exact first field.
		{"annotation missing space after name:", "-- name:GetUser :one\na"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err == nil {
				t.Errorf("expected an error, got none\ninput: %q\ntokens: %v", tt.input, toks)
			}
		})
	}
}
