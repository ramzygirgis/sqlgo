package generator

import (
	"go/token"
	"testing"
)

/* ---------- exact conversions ---------- */

func TestSnakeToCamelExported(t *testing.T) {
	tests := []struct{ in, want string }{
		{"name", "Name"},
		{"id", "ID"},
		{"user_id", "UserID"},
		{"created_at", "CreatedAt"},
		{"api_url", "APIURL"},
		{"url", "URL"},
		{"uuid", "UUID"},
		{"http_status", "HTTPStatus"},
		{"json_data", "JSONData"},
		{"user_ip", "UserIP"},
		{"userName", "Username"}, // inner capitals kept
		{"USER_ID", "UserID"},    // all-caps pieces flattened
		{"_id", "ID"},            // leading underscore
		{"user__id", "UserID"},   // double underscore
		{"user_", "User"},        // trailing underscore
		{"x", "X"},               // single letter
		{"type", "Type"},         // keywords are lowercase, so exported is safe
		{"first_name_last", "FirstNameLast"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := snakeToCamel(tt.in, true); got != tt.want {
				t.Errorf("snakeToCamel(%q, true) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSnakeToCamelUnexported(t *testing.T) {
	tests := []struct{ in, want string }{
		{"name", "name"},
		{"id", "id"},
		{"user_id", "userID"},
		{"created_at", "createdAt"},
		{"api_url", "apiURL"},
		{"url_path", "urlPath"},
		{"http_status", "httpStatus"},
		{"userName", "username"}, // inner capitals kept
		{"USER_ID", "userID"},    // all-caps pieces flattened
		{"ID", "id"},             // initialism first piece fully lowercased
		{"_id", "id"},            // first NON-EMPTY piece is the lowercase one
		{"user__id", "userID"},
		{"user_", "user"},
		{"x", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := snakeToCamel(tt.in, false); got != tt.want {
				t.Errorf("snakeToCamel(%q, false) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

/* ---------- properties every output must have ---------- */

var namingInputs = []string{
	"name", "id", "user_id", "created_at", "api_url", "userName", "USER_ID",
	"_id", "user__id", "user_", "x", "http_status", "json_data",
	"type", "func", "range", "default", "map", "select", "case", "var",
	"2fa_code", "1st_place",
}

func TestOutputsAreValidIdentifiers(t *testing.T) {
	for _, in := range namingInputs {
		for _, exported := range []bool{true, false} {
			got := snakeToCamel(in, exported)
			if !token.IsIdentifier(got) {
				t.Errorf("snakeToCamel(%q, %v) = %q, which is not a valid Go identifier (keyword or bad start?)", in, exported, got)
			}
		}
	}
}

func TestExportedness(t *testing.T) {
	for _, in := range namingInputs {
		if got := snakeToCamel(in, true); !token.IsExported(got) {
			t.Errorf("snakeToCamel(%q, true) = %q, should be exported", in, got)
		}
		// A leading-digit fix may prepend "_" or an uppercase letter;
		// only check exportedness for inputs that start with a letter.
		if in[0] >= '0' && in[0] <= '9' {
			continue
		}
		if got := snakeToCamel(in, false); token.IsExported(got) {
			t.Errorf("snakeToCamel(%q, false) = %q, should be unexported", in, got)
		}
	}
}

func TestKeywordsAreEscaped(t *testing.T) {
	for _, kw := range []string{"type", "func", "range", "default", "map", "select", "case", "var", "go", "chan"} {
		got := snakeToCamel(kw, false)
		if token.IsKeyword(got) {
			t.Errorf("snakeToCamel(%q, false) = %q, which is still a Go keyword", kw, got)
		}
		if !token.IsIdentifier(got) {
			t.Errorf("snakeToCamel(%q, false) = %q, not a valid identifier", kw, got)
		}
	}
}

func TestLeadingDigitIsFixed(t *testing.T) {
	for _, in := range []string{"2fa_code", "1st_place", "3d"} {
		for _, exported := range []bool{true, false} {
			got := snakeToCamel(in, exported)
			if got == "" || (got[0] >= '0' && got[0] <= '9') {
				t.Errorf("snakeToCamel(%q, %v) = %q, must not start with a digit", in, exported, got)
			}
		}
	}
}

/* ---------- model names ---------- */

// Requires: func modelName(table string) string
// (e.g. snakeToCamel(singularize(table), true)).
func TestModelName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"users", "User"},
		{"posts", "Post"},
		{"user_accounts", "UserAccount"},
		{"categories", "Category"},
		{"addresses", "Address"},
		{"statuses", "Status"},
		{"status", "Status"},
		{"people", "Person"},
		{"api_keys", "APIKey"},
		{"user_ids", "UserID"},
		{"data", "Data"},
		{"Users", "User"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := modelName(tt.in); got != tt.want {
				t.Errorf("modelName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
