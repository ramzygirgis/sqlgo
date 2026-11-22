package codegen

import (
	"strings"
	"testing"

	"github.com/ramzygirgis/sqlgo/internal/ir"
)

// mapType is the function under test. Rename the right-hand side if your
// mapping function has a different name.
var mapType = irToGoType

/* ---------- mapping ---------- */

func TestMapType(t *testing.T) {
	tests := []struct {
		name        string
		sqlType     ir.SQLType
		nullable    bool
		typeName    string
		packageName string
		importPath  string
		rendered    string // String()
	}{
		{"text not null", ir.TypeText, false, "string", "", "", "string"},
		{"text nullable", ir.TypeText, true, "NullString", "sql", "database/sql", "sql.NullString"},
		{"int not null", ir.TypeInt, false, "int32", "", "", "int32"},
		{"int nullable", ir.TypeInt, true, "NullInt32", "sql", "database/sql", "sql.NullInt32"},
		{"timestamp not null", ir.TypeTimestamp, false, "Time", "time", "time", "time.Time"},
		{"timestamp nullable", ir.TypeTimestamp, true, "NullTime", "sql", "database/sql", "sql.NullTime"},
		{"boolean not null", ir.TypeBoolean, false, "bool", "", "", "bool"},
		{"boolean nullable", ir.TypeBoolean, true, "NullBool", "sql", "database/sql", "sql.NullBool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapType(tt.sqlType, tt.nullable)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.TypeName != tt.typeName {
				t.Errorf("TypeName: got %q, want %q", got.TypeName, tt.typeName)
			}
			if got.PackageName != tt.packageName {
				t.Errorf("PackageName: got %q, want %q", got.PackageName, tt.packageName)
			}
			if got.ImportPath != tt.importPath {
				t.Errorf("ImportPath: got %q, want %q", got.ImportPath, tt.importPath)
			}
			if s := got.String(); s != tt.rendered {
				t.Errorf("String(): got %q, want %q", s, tt.rendered)
			}
		})
	}
}

func TestMapTypeErrors(t *testing.T) {
	tests := []struct {
		name        string
		sqlType     ir.SQLType
		nullable    bool
		errContains string
	}{
		{"illegal not null", ir.IllegalSQLType, false, "IllegalSQLType"},
		{"illegal nullable", ir.IllegalSQLType, true, "IllegalSQLType"},
		{"out of range", ir.SQLType(99), false, "SQLType(99)"},
		{"out of range nullable", ir.SQLType(99), true, "SQLType(99)"},
		{"negative", ir.SQLType(-1), false, "SQLType(-1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapType(tt.sqlType, tt.nullable)
			if err == nil {
				t.Fatalf("expected an error, got %+v", got)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error %q should mention %q", err, tt.errContains)
			}
		})
	}
}

// Every SQLType defined in ir (other than IllegalSQLType) must map in both
// nullabilities. Adding a type to ir without updating the mapping fails here.
// Types are enumerated by counting up until String() falls back to "SQLType(N)".
func TestEverySQLTypeMaps(t *testing.T) {
	count := 0
	for st := ir.IllegalSQLType + 1; !strings.HasPrefix(st.String(), "SQLType("); st++ {
		count++
		for _, nullable := range []bool{false, true} {
			got, err := mapType(st, nullable)
			if err != nil {
				t.Errorf("%v (nullable=%v): no mapping: %v", st, nullable, err)
				continue
			}
			if got.TypeName == "" {
				t.Errorf("%v (nullable=%v): empty TypeName", st, nullable)
			}
		}
	}
	if count == 0 {
		t.Fatal("found no SQL types to check; is SQLType.String() still falling back to SQLType(N)?")
	}
}

// A type has a package qualifier exactly when it has an import: a qualifier
// without an import won't compile, and an import without a qualifier is unused.
// (Not-null types aren't all builtins: TIMESTAMP is time.Time.)
func TestPackageAndImportAgree(t *testing.T) {
	for st := ir.IllegalSQLType + 1; !strings.HasPrefix(st.String(), "SQLType("); st++ {
		for _, nullable := range []bool{false, true} {
			got, err := mapType(st, nullable)
			if err != nil {
				t.Fatalf("%v (nullable=%v): unexpected error: %v", st, nullable, err)
			}
			if (got.PackageName == "") != (got.ImportPath == "") {
				t.Errorf("%v (nullable=%v): package %q but import %q", st, nullable, got.PackageName, got.ImportPath)
			}
		}
	}
}

// Nullable types must carry an import, or the generated file won't compile.
func TestNullableHasImport(t *testing.T) {
	for st := ir.IllegalSQLType + 1; !strings.HasPrefix(st.String(), "SQLType("); st++ {
		got, err := mapType(st, true)
		if err != nil {
			t.Fatalf("%v: unexpected error: %v", st, err)
		}
		if got.ImportPath == "" || got.PackageName == "" {
			t.Errorf("%v nullable: missing package or import: %+v", st, got)
		}
	}
}

func TestMapTypeIsPure(t *testing.T) {
	a, errA := mapType(ir.TypeText, true)
	b, errB := mapType(ir.TypeText, true)
	if errA != nil || errB != nil {
		t.Fatalf("unexpected errors: %v, %v", errA, errB)
	}
	if a != b {
		t.Errorf("same input gave different results: %+v vs %+v", a, b)
	}
}

/* ---------- rendering ---------- */

func TestGoTypeString(t *testing.T) {
	tests := []struct {
		name string
		t    GoType
		want string
	}{
		{"builtin", GoType{TypeName: "string"}, "string"},
		{"qualified", GoType{TypeName: "NullString", PackageName: "sql", ImportPath: "database/sql"}, "sql.NullString"},
		{"local type still qualified by String", GoType{TypeName: "User", PackageName: "db", ImportPath: "github.com/you/app/db"}, "db.User"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("String(): got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGoTypeIn(t *testing.T) {
	nullString := GoType{TypeName: "NullString", PackageName: "sql", ImportPath: "database/sql"}
	builtin := GoType{TypeName: "int32"}
	localUser := GoType{TypeName: "User", PackageName: "db", ImportPath: "github.com/you/app/db"}

	tests := []struct {
		name    string
		t       GoType
		current string // import path of the package being generated
		want    string
	}{
		{"builtin anywhere", builtin, "github.com/you/app/db", "int32"},
		{"builtin with empty current", builtin, "", "int32"},
		{"external type from another package", nullString, "github.com/you/app/db", "sql.NullString"},
		{"local type inside its own package", localUser, "github.com/you/app/db", "User"},
		{"local type from another package", localUser, "github.com/you/app/api", "db.User"},
		// Output package is *named* sql but is not database/sql:
		// the qualifier must stay, since import paths differ.
		{"package name collision keeps qualifier", nullString, "github.com/you/app/sql", "sql.NullString"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.In(tt.current); got != tt.want {
				t.Errorf("In(%q): got %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}
