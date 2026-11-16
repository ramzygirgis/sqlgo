package codegen

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ramzygirgis/sqlgo/internal/catalog"
	"github.com/ramzygirgis/sqlgo/internal/ir"
	"github.com/ramzygirgis/sqlgo/internal/parsing"
)

const (
	testModelsPath  = "github.com/you/app/internal/models"
	testQueriesPath = "github.com/you/app/internal/accounts"
)

const usersSchema = `
CREATE TABLE users (
	id INT PRIMARY KEY,
	name TEXT NOT NULL,
	bio TEXT
);`

/* ---------- helpers ---------- */

// mustDbFile generates db.go, failing the test on error.
func mustDbFile(t *testing.T, modelsImportPath string) string {
	t.Helper()
	src, err := generateDbFile(modelsImportPath)
	if err != nil {
		t.Fatalf("generateDbFile(%q): unexpected error: %v", modelsImportPath, err)
	}
	return src
}

func buildCatalog(t *testing.T, schema string) catalog.Catalog {
	t.Helper()
	stmts, err := parsing.ParseSchema(schema)
	if err != nil {
		t.Fatalf("schema parse error: %v", err)
	}
	cat := catalog.New()
	if err := cat.ApplyAll(stmts); err != nil {
		t.Fatalf("schema apply error: %v", err)
	}
	return *cat
}

// parseGo parses generated source, failing the test with the source attached.
func parseGo(t *testing.T, name, src string) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("%s does not parse: %v\n--- source ---\n%s", name, err, src)
	}
	return f
}

func importsOf(f *ast.File) []string {
	var paths []string
	for _, imp := range f.Imports {
		p, _ := strconv.Unquote(imp.Path.Value)
		paths = append(paths, p)
	}
	slices.Sort(paths)
	return paths
}

// structFields returns "Name Type" for every field of the named struct, in order.
func structFields(t *testing.T, f *ast.File, name string) []string {
	t.Helper()
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts := spec.(*ast.TypeSpec)
			if ts.Name.Name != name {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("%s is not a struct", name)
			}
			var out []string
			for _, field := range st.Fields.List {
				typ := types.ExprString(field.Type)
				for _, n := range field.Names {
					out = append(out, n.Name+" "+typ)
				}
			}
			return out
		}
	}
	t.Fatalf("struct %s not found", name)
	return nil
}

// topLevelNames returns every top-level func, type, const and var name.
func topLevelNames(f *ast.File) map[string]bool {
	names := map[string]bool{}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			names[d.Name.Name] = true
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					names[s.Name.Name] = true
				case *ast.ValueSpec:
					for _, n := range s.Names {
						names[n.Name] = true
					}
				}
			}
		}
	}
	return names
}

// testImporter resolves the generated models package from source we just
// type-checked, and everything else (stdlib) from GOROOT source.
type testImporter struct {
	local  map[string]*types.Package
	stdlib types.Importer
}

func (imp testImporter) Import(path string) (*types.Package, error) {
	if p, ok := imp.local[path]; ok {
		return p, nil
	}
	return imp.stdlib.Import(path)
}

func newImporter() testImporter {
	return testImporter{
		local:  map[string]*types.Package{},
		stdlib: importer.ForCompiler(token.NewFileSet(), "source", nil),
	}
}

// typecheck type-checks the given files as one package. A pass means the
// generated code really compiles: no unused imports, no missing fields,
// no mismatched signatures.
func typecheck(t *testing.T, pkgPath string, files map[string]string, imp testImporter) *types.Package {
	t.Helper()
	fset := token.NewFileSet()
	var parsed []*ast.File
	for name, src := range files {
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("%s does not parse: %v\n--- source ---\n%s", name, err, src)
		}
		parsed = append(parsed, f)
	}
	conf := types.Config{Importer: imp}
	pkg, err := conf.Check(pkgPath, fset, parsed, nil)
	if err != nil {
		var dump string
		for name, src := range files {
			dump += fmt.Sprintf("--- %s ---\n%s\n", name, src)
		}
		t.Fatalf("package %s does not type-check: %v\n%s", pkgPath, err, dump)
	}
	return pkg
}

/* ---------- db.go ---------- */

func TestDbFileParses(t *testing.T) {
	src := mustDbFile(t, testModelsPath)
	f := parseGo(t, "db.go", src)
	if f.Name.Name != "models" {
		t.Errorf("package = %q, want %q", f.Name.Name, "models")
	}
}

func TestDbFileDeclarations(t *testing.T) {
	f := parseGo(t, "db.go", mustDbFile(t, testModelsPath))
	names := topLevelNames(f)
	for _, want := range []string{"DBTX", "TxStarter", "InTx"} {
		if !names[want] {
			t.Errorf("db.go is missing %s", want)
		}
	}
}

func TestDbFilePackageFollowsPath(t *testing.T) {
	f := parseGo(t, "db.go", mustDbFile(t, "github.com/you/app/internal/db"))
	if f.Name.Name != "db" {
		t.Errorf("package = %q, want %q", f.Name.Name, "db")
	}
}

func TestDbFileTypeChecks(t *testing.T) {
	typecheck(t, testModelsPath, map[string]string{
		"db.go": mustDbFile(t, testModelsPath),
	}, newImporter())
}

// DBTX must be satisfied by all three database/sql handles, and TxStarter
// must NOT be satisfied by *sql.Tx (that's what prevents nested transactions).
func TestDbFileInterfaces(t *testing.T) {
	imp := newImporter()
	models := typecheck(t, testModelsPath, map[string]string{
		"db.go": mustDbFile(t, testModelsPath),
	}, imp)
	sqlPkg, err := imp.Import("database/sql")
	if err != nil {
		t.Fatalf("importing database/sql: %v", err)
	}

	iface := func(name string) *types.Interface {
		obj := models.Scope().Lookup(name)
		if obj == nil {
			t.Fatalf("%s not declared", name)
		}
		return obj.Type().Underlying().(*types.Interface)
	}
	ptr := func(name string) types.Type {
		return types.NewPointer(sqlPkg.Scope().Lookup(name).Type())
	}

	dbtx := iface("DBTX")
	for _, h := range []string{"DB", "Conn", "Tx"} {
		if !types.Implements(ptr(h), dbtx) {
			t.Errorf("*sql.%s should implement DBTX", h)
		}
	}

	starter := iface("TxStarter")
	for _, h := range []string{"DB", "Conn"} {
		if !types.Implements(ptr(h), starter) {
			t.Errorf("*sql.%s should implement TxStarter", h)
		}
	}
	if types.Implements(ptr("Tx"), starter) {
		t.Error("*sql.Tx must NOT implement TxStarter (nested transactions would be allowed)")
	}
}

/* ---------- models.go ---------- */

func TestModelsFileStruct(t *testing.T) {
	src, err := generateModelsFile(buildCatalog(t, usersSchema), testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := parseGo(t, "models.go", src)
	if f.Name.Name != "models" {
		t.Errorf("package = %q, want %q", f.Name.Name, "models")
	}

	got := structFields(t, f, "User")
	want := []string{"ID int32", "Name string", "Bio sql.NullString"}
	if !slices.Equal(got, want) {
		t.Errorf("User fields:\n got: %v\nwant: %v", got, want)
	}
}

func TestModelsFileNaming(t *testing.T) {
	src, err := generateModelsFile(buildCatalog(t, `
CREATE TABLE user_accounts (
	user_id INT NOT NULL,
	api_url TEXT NOT NULL
);`), testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := parseGo(t, "models.go", src)
	got := structFields(t, f, "UserAccount")
	want := []string{"UserID int32", "APIURL string"}
	if !slices.Equal(got, want) {
		t.Errorf("UserAccount fields:\n got: %v\nwant: %v", got, want)
	}
}

func TestModelsFileMultipleTables(t *testing.T) {
	src, err := generateModelsFile(buildCatalog(t, `
CREATE TABLE users (id INT PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE posts (id INT PRIMARY KEY, title TEXT NOT NULL, body TEXT);`), testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := parseGo(t, "models.go", src)
	if got := structFields(t, f, "User"); !slices.Equal(got, []string{"ID int32", "Name string"}) {
		t.Errorf("User fields: %v", got)
	}
	if got := structFields(t, f, "Post"); !slices.Equal(got, []string{"ID int32", "Title string", "Body sql.NullString"}) {
		t.Errorf("Post fields: %v", got)
	}
}

func TestModelsFileImports(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		want   []string
	}{
		{"nullable column needs database/sql", usersSchema, []string{"database/sql"}},
		{"all non-null needs nothing", `CREATE TABLE t (id INT PRIMARY KEY, name TEXT NOT NULL);`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := generateModelsFile(buildCatalog(t, tt.schema), testModelsPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := importsOf(parseGo(t, "models.go", src)); !slices.Equal(got, tt.want) {
				t.Errorf("imports: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelsFileIsDeterministic(t *testing.T) {
	cat := buildCatalog(t, `
CREATE TABLE users (id INT PRIMARY KEY);
CREATE TABLE posts (id INT PRIMARY KEY);
CREATE TABLE comments (id INT PRIMARY KEY);
CREATE TABLE tags (id INT PRIMARY KEY);`)
	first, err := generateModelsFile(cat, testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 20; i++ {
		again, err := generateModelsFile(cat, testModelsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if again != first {
			t.Fatalf("output changed between runs (map iteration order leaking?)\n--- first ---\n%s\n--- run %d ---\n%s", first, i+2, again)
		}
	}
}

func TestModelsPackageTypeChecks(t *testing.T) {
	models, err := generateModelsFile(buildCatalog(t, usersSchema), testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	typecheck(t, testModelsPath, map[string]string{
		"models.go": models,
		"db.go":     mustDbFile(t, testModelsPath),
	}, newImporter())
}

/* ---------- queries file ---------- */

func userQueries() []*ir.Query {
	idCol := ir.Column{Name: "id", SqlType: ir.TypeInt, Nullable: false}
	nameCol := ir.Column{Name: "name", SqlType: ir.TypeText, Nullable: false}
	bioCol := ir.Column{Name: "bio", SqlType: ir.TypeText, Nullable: true}
	all := []ir.Column{idCol, nameCol, bioCol}

	return []*ir.Query{
		{
			Name:          "GetUser",
			Cardinality:   ir.One,
			Kind:          ir.Select,
			TargetTable:   "users",
			Parameters:    []ir.Parameter{{Index: 1, SqlType: ir.TypeInt}},
			ResultColumns: all,
			Sql:           "SELECT id, name, bio FROM users WHERE (id = $1);",
		},
		{
			Name:          "ListUsers",
			Cardinality:   ir.Many,
			Kind:          ir.Select,
			TargetTable:   "users",
			ResultColumns: all,
			Sql:           "SELECT id, name, bio FROM users;",
		},
		{
			Name:        "CreateUser",
			Cardinality: ir.Exec,
			Kind:        ir.Insert,
			TargetTable: "users",
			Parameters: []ir.Parameter{
				{Index: 1, SqlType: ir.TypeText, Nullable: false},
				{Index: 2, SqlType: ir.TypeText, Nullable: true},
			},
			TargetColumns: []ir.Column{nameCol, bioCol},
			Sql:           "INSERT INTO users (name, bio) VALUES ($1, $2);",
		},
	}
}

func TestQueriesFileParses(t *testing.T) {
	src, err := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := parseGo(t, "queries.sql.go", src)
	if f.Name.Name != "accounts" {
		t.Errorf("package = %q, want %q", f.Name.Name, "accounts")
	}
}

func TestQueriesFileDeclarations(t *testing.T) {
	src, err := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := topLevelNames(parseGo(t, "queries.sql.go", src))
	for _, want := range []string{
		"GetUser", "ListUsers", "CreateUser", // functions
		"getUser", "listUsers", "createUser", // SQL constants
	} {
		if !names[want] {
			t.Errorf("queries file is missing %s", want)
		}
	}
}

func TestQueriesFileSingleImportBlock(t *testing.T) {
	src, err := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := parseGo(t, "queries.sql.go", src)

	blocks := 0
	for _, decl := range f.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
			blocks++
		}
	}
	if blocks != 1 {
		t.Errorf("got %d import blocks, want 1", blocks)
	}

	got := importsOf(f)
	if len(got) != len(slices.Compact(slices.Clone(got))) {
		t.Errorf("duplicate imports: %v", got)
	}
	for _, want := range []string{"context", "database/sql", "fmt", testModelsPath} {
		if !slices.Contains(got, want) {
			t.Errorf("missing import %q (got %v)", want, got)
		}
	}
}

func TestQueriesFileImportsMatchUsage(t *testing.T) {
	// Only an :exec with non-null params: no fmt, no database/sql.
	q := []*ir.Query{{
		Name:          "CreateUser",
		Cardinality:   ir.Exec,
		Kind:          ir.Insert,
		TargetTable:   "users",
		Parameters:    []ir.Parameter{{Index: 1, SqlType: ir.TypeText}},
		TargetColumns: []ir.Column{{Name: "name", SqlType: ir.TypeText}},
		Sql:           "INSERT INTO users (name) VALUES ($1);",
	}}
	src, err := generateQueriesFile(q, testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := importsOf(parseGo(t, "queries.sql.go", src))
	want := []string{"context", testModelsPath}
	if !slices.Equal(got, want) {
		t.Errorf("imports: got %v, want %v", got, want)
	}
}

func TestQueriesFileIsDeterministic(t *testing.T) {
	first, err := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 20; i++ {
		again, _ := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
		if again != first {
			t.Fatalf("output changed between runs\n--- first ---\n%s\n--- run %d ---\n%s", first, i+2, again)
		}
	}
}

func TestQueriesFileKeepsQueryOrder(t *testing.T) {
	src, err := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var funcs []string
	for _, decl := range parseGo(t, "queries.sql.go", src).Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			funcs = append(funcs, fd.Name.Name)
		}
	}
	want := []string{"GetUser", "ListUsers", "CreateUser"}
	if !slices.Equal(funcs, want) {
		t.Errorf("function order: got %v, want %v", funcs, want)
	}
}

// Empty input is an error: the CLI decides not to write (or to remove) the file.
func TestQueriesFileEmpty(t *testing.T) {
	for _, qs := range [][]*ir.Query{nil, {}} {
		if src, err := generateQueriesFile(qs, testQueriesPath, testModelsPath); err == nil {
			t.Errorf("expected an error for %d queries, got none\n%s", len(qs), src)
		}
	}
}

func TestModelsFileEmpty(t *testing.T) {
	cat := catalog.New()
	if src, err := generateModelsFile(*cat, testModelsPath); err == nil {
		t.Errorf("expected an error for an empty catalog, got none\n%s", src)
	}
}

// Every generated file starts with Go's standard generated-code marker,
// which the CLI relies on before overwriting or deleting a file.
func TestGeneratedHeader(t *testing.T) {
	const header = "// Code generated by sqlgo. DO NOT EDIT."

	models, err := generateModelsFile(buildCatalog(t, usersSchema), testModelsPath)
	if err != nil {
		t.Fatalf("models: %v", err)
	}
	queries, err := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("queries: %v", err)
	}
	files := map[string]string{
		"db.go":          mustDbFile(t, testModelsPath),
		"models.go":      models,
		"queries.sql.go": queries,
	}
	for name, src := range files {
		first, _, _ := strings.Cut(src, "\n")
		if first != header {
			t.Errorf("%s: first line = %q, want %q", name, first, header)
		}
	}
}

func TestQueriesFileIllegalType(t *testing.T) {
	q := userQueries()
	q[0].Parameters[0].SqlType = ir.IllegalSQLType
	if _, err := generateQueriesFile(q, testQueriesPath, testModelsPath); err == nil {
		t.Error("expected an error for an illegal param type, got none")
	}
}

/* ---------- everything together ---------- */

// The real proof: generate the models package and a queries package, then
// type-check both as Go would compile them. Catches wrong field names,
// return-type mismatches, missing or unused imports, and models./DBTX
// qualification bugs in one go.
func TestEndToEndTypeChecks(t *testing.T) {
	models, err := generateModelsFile(buildCatalog(t, usersSchema), testModelsPath)
	if err != nil {
		t.Fatalf("models: %v", err)
	}
	queries, err := generateQueriesFile(userQueries(), testQueriesPath, testModelsPath)
	if err != nil {
		t.Fatalf("queries: %v", err)
	}

	imp := newImporter()
	modelsPkg := typecheck(t, testModelsPath, map[string]string{
		"models.go": models,
		"db.go":     mustDbFile(t, testModelsPath),
	}, imp)
	imp.local[testModelsPath] = modelsPkg

	typecheck(t, testQueriesPath, map[string]string{
		"queries.sql.go": queries,
	}, imp)
}
