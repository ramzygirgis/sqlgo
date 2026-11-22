package parsing

import (
	"strings"
	"testing"
)

// whereOperand parses a SELECT and returns the right side of its WHERE comparison.
func whereOperand(t *testing.T, where string) Expr {
	t.Helper()
	stmt, err := Parse("SELECT id FROM users WHERE " + where + ";")
	if err != nil {
		t.Fatalf("parse %q: %v", where, err)
	}
	s, ok := stmt.(*SelectStatement)
	if !ok {
		t.Fatalf("got %T, want *SelectStatement", stmt)
	}
	b, ok := s.Where.(*BinaryExpr)
	if !ok {
		t.Fatalf("WHERE is %T, want *BinaryExpr", s.Where)
	}
	return b.Right
}

func TestLiteralKinds(t *testing.T) {
	tests := []struct {
		where    string
		wantKind LiteralKind
		wantBody string
	}{
		{"name = 'bob'", StringLiteral, "'bob'"},
		{"id = 5", NumberLiteral, "5"},
		{"score = 5.5", NumberLiteral, "5.5"},
		{"active = TRUE", BooleanLiteral, "TRUE"},
		{"active = false", BooleanLiteral, "FALSE"}, // keywords are uppercased by the lexer
		{"bio = NULL", NullLiteral, "NULL"},
		{"bio = null", NullLiteral, "NULL"},
		{"name = 'true'", StringLiteral, "'true'"}, // quoted: a string, not a boolean
	}
	for _, tt := range tests {
		t.Run(tt.where, func(t *testing.T) {
			e := whereOperand(t, tt.where)
			lit, ok := e.(*Literal)
			if !ok {
				t.Fatalf("got %T (%v), want *Literal", e, e)
			}
			if lit.Kind != tt.wantKind {
				t.Errorf("Kind = %s, want %s", lit.Kind, tt.wantKind)
			}
			if lit.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", lit.Body, tt.wantBody)
			}
		})
	}
}

// Words that only look like literals are still column references.
func TestLiteralLookalikesAreColumns(t *testing.T) {
	for _, where := range []string{"a = is_true", "a = nullable", "a = falsey"} {
		t.Run(where, func(t *testing.T) {
			if _, ok := whereOperand(t, where).(*ColumnRef); !ok {
				t.Errorf("%q: right side should be a *ColumnRef", where)
			}
		})
	}
}

func TestLiteralsInInsert(t *testing.T) {
	stmt, err := Parse("INSERT INTO users (name, active, bio) VALUES ('bob', TRUE, NULL);")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ins, ok := stmt.(*InsertIntoStatement)
	if !ok {
		t.Fatalf("got %T, want *InsertIntoStatement", stmt)
	}
	want := []LiteralKind{StringLiteral, BooleanLiteral, NullLiteral}
	if len(ins.Values) != len(want) {
		t.Fatalf("got %d values, want %d", len(ins.Values), len(want))
	}
	for i, v := range ins.Values {
		lit, ok := v.(*Literal)
		if !ok {
			t.Errorf("value %d: got %T, want *Literal", i, v)
			continue
		}
		if lit.Kind != want[i] {
			t.Errorf("value %d: Kind = %s, want %s", i, lit.Kind, want[i])
		}
	}
}

// Literal keywords print back as written, so the generated SQL keeps them.
func TestLiteralRoundTrip(t *testing.T) {
	tests := []struct{ query, want string }{
		{"SELECT id FROM users WHERE active = TRUE;", "TRUE"},
		{"SELECT id FROM users WHERE active = false;", "FALSE"},
		{"INSERT INTO users (name, bio) VALUES ('bob', NULL);", "NULL"},
	}
	for _, tt := range tests {
		stmt, err := Parse(tt.query)
		if err != nil {
			t.Fatalf("parse %q: %v", tt.query, err)
		}
		if got := stmt.String(); !strings.Contains(got, tt.want) {
			t.Errorf("%q printed back as %q, missing %s", tt.query, got, tt.want)
		}
	}
}

func TestKeywordInExpressionIsAnError(t *testing.T) {
	for _, where := range []string{"a = SELECT", "a = FROM", "a = WHERE"} {
		t.Run(where, func(t *testing.T) {
			if _, err := Parse("SELECT id FROM users WHERE " + where + ";"); err == nil {
				t.Errorf("expected an error for %q", where)
			}
		})
	}
}

func TestLiteralKindString(t *testing.T) {
	tests := []struct {
		k    LiteralKind
		want string
	}{
		{IllegalLiteral, "IllegalLiteral"},
		{StringLiteral, "StringLiteral"},
		{NumberLiteral, "NumberLiteral"},
		{BooleanLiteral, "BooleanLiteral"},
		{NullLiteral, "NullLiteral"},
		{LiteralKind(99), "LiteralKind(99)"},
		{LiteralKind(-1), "LiteralKind(-1)"},
	}
	for _, tt := range tests {
		if got := tt.k.String(); got != tt.want {
			t.Errorf("LiteralKind(%d).String() = %q, want %q", int(tt.k), got, tt.want)
		}
	}
}
