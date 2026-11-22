package analyzer

import (
	"strings"
	"testing"

	"github.com/ramzygirgis/sqlgo/internal/ir"
)

const flagsSchema = `
CREATE TABLE flags (
	id INT PRIMARY KEY,
	name TEXT NOT NULL,
	enabled BOOLEAN NOT NULL,
	note TEXT,
	created_at TIMESTAMP NOT NULL,
	deleted_at TIMESTAMP
);`

func analyzeFlags(t *testing.T, query, card string) (*ir.Query, error) {
	t.Helper()
	cat := buildCatalog(t, flagsSchema)
	q := parseOne(t, "-- name: "+testName+" "+card+"\n"+query)
	return AnalyzeQuery(cat, q)
}

// BOOLEAN and TIMESTAMP columns reach the IR with the right types.
func TestNewColumnTypes(t *testing.T) {
	q, err := analyzeFlags(t, "SELECT enabled, created_at, deleted_at FROM flags;", ":many")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkColumns(t, "ResultColumns", q.ResultColumns, []ir.Column{
		col("enabled", ir.TypeBoolean, false),
		col("created_at", ir.TypeTimestamp, false),
		col("deleted_at", ir.TypeTimestamp, true),
	})
}

func TestNewParamTypes(t *testing.T) {
	q, err := analyzeFlags(t, "SELECT id FROM flags WHERE enabled = $1 AND created_at > $2;", ":many")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkParams(t, q.Parameters, []ir.Parameter{
		param(1, ir.TypeBoolean, false),
		param(2, ir.TypeTimestamp, false),
	})
}

func TestInsertLiterals(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		errContains string // empty means success
	}{
		{"boolean literal", "INSERT INTO flags (id, name, enabled, created_at) VALUES ($1, 'x', TRUE, $2);", ""},
		{"NULL into nullable column", "INSERT INTO flags (id, name, enabled, note, created_at) VALUES ($1, 'x', FALSE, NULL, $2);", ""},
		{"NULL into NOT NULL column", "INSERT INTO flags (id, name, enabled, created_at) VALUES ($1, NULL, TRUE, $2);", "name"},
		{"NULL into NOT NULL boolean", "INSERT INTO flags (id, name, enabled, created_at) VALUES ($1, 'x', NULL, $2);", "enabled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := analyzeFlags(t, tt.query, ":exec")
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error, got none")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error %q should mention %q", err, tt.errContains)
			}
		})
	}
}

// Literals never become parameters: only the $n placeholders do.
func TestInsertLiteralsAreNotParams(t *testing.T) {
	q, err := analyzeFlags(t, "INSERT INTO flags (id, name, enabled, note, created_at) VALUES ($1, 'x', TRUE, NULL, $2);", ":exec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkParams(t, q.Parameters, []ir.Parameter{
		param(1, ir.TypeInt, false),
		param(2, ir.TypeTimestamp, false),
	})
}
