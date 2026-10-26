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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err == nil {
				t.Errorf("expected an error, got none\ninput: %s\ntokens: %v", tt.input, toks)
			}
		})
	}
}
