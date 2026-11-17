package project

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// tree builds an in-memory filesystem from path -> contents.
// Directories are implied by the paths; MapFS creates them automatically.
func tree(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for p, contents := range files {
		fsys[p] = &fstest.MapFile{Data: []byte(contents)}
	}
	return fsys
}

const goMod = "module github.com/you/app\n\ngo 1.27\n"

// pkg is shorthand for an expected query package.
func pkg(dir string, files ...string) Package {
	return Package{
		DirectoryInfo: DirectoryInfo{
			Dir:        dir,
			ImportPath: "github.com/you/app/" + dir,
		},
		SQLFiles: files,
	}
}

func mustDiscover(t *testing.T, fsys fstest.MapFS) Project {
	t.Helper()
	p, err := Discover(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return p
}

func checkPackages(t *testing.T, got, want []Package) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d packages %+v, want %d %+v", len(got), got, len(want), want)
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("package %d:\n got: %+v\nwant: %+v", i, got[i], want[i])
		}
	}
}

/* ---------- full projects ---------- */

func TestDiscoverTypicalProject(t *testing.T) {
	p := mustDiscover(t, tree(map[string]string{
		"go.mod":                        goMod,
		"schema/001_init.sql":           "CREATE TABLE users (id INT);",
		"schema/002_add_bio.sql":        "ALTER TABLE users ADD COLUMN bio TEXT;",
		"schema/002_add_bio.down.sql":   "ALTER TABLE users DROP COLUMN bio;",
		"internal/accounts/queries.sql": "-- name: GetUser :one",
		"internal/accounts/README.md":   "docs",
		"internal/billing/payments.sql": "-- name: ListPayments :many",
		"internal/billing/invoices.sql": "-- name: ListInvoices :many",
	}))

	if p.ModulePath != "github.com/you/app" {
		t.Errorf("ModulePath = %q", p.ModulePath)
	}
	wantSchema := []string{"schema/001_init.sql", "schema/002_add_bio.sql"}
	if !reflect.DeepEqual(p.SchemaFiles, wantSchema) {
		t.Errorf("SchemaFiles:\n got: %v\nwant: %v", p.SchemaFiles, wantSchema)
	}
	wantModels := DirectoryInfo{Dir: "internal/models", ImportPath: "github.com/you/app/internal/models"}
	if p.ModelsInfo != wantModels {
		t.Errorf("ModelsInfo:\n got: %+v\nwant: %+v", p.ModelsInfo, wantModels)
	}
	checkPackages(t, p.QueryPackages, []Package{
		pkg("internal/accounts", "queries.sql"),
		pkg("internal/billing", "invoices.sql", "payments.sql"),
	})
}

func TestDiscoverIgnoredDirectories(t *testing.T) {
	p := mustDiscover(t, tree(map[string]string{
		"go.mod":                     goMod,
		"schema/001_init.sql":        "",
		"internal/users/queries.sql": "",
		"vendor/lib/x.sql":           "",
		"testdata/fixture.sql":       "",
		".git/hooks/x.sql":           "",
		"internal/.hidden/x.sql":     "",
		"_scratch/try.sql":           "",
		"internal/_old/x.sql":        "",
		"tools/go.mod":               "module github.com/you/tools\n",
		"tools/gen/q.sql":            "",
		"tools/q.sql":                "",
	}))
	checkPackages(t, p.QueryPackages, []Package{pkg("internal/users", "queries.sql")})
}

func TestDiscoverNoQueries(t *testing.T) {
	p := mustDiscover(t, tree(map[string]string{
		"go.mod":              goMod,
		"schema/001_init.sql": "",
		"cmd/app/main.go":     "package main",
	}))
	if len(p.QueryPackages) != 0 {
		t.Errorf("expected no query packages, got %+v", p.QueryPackages)
	}
}

func TestDiscoverModelsDirNeedNotExist(t *testing.T) {
	p := mustDiscover(t, tree(map[string]string{
		"go.mod":              goMod,
		"schema/001_init.sql": "",
	}))
	if p.ModelsInfo.Dir != "internal/models" {
		t.Errorf("ModelsInfo.Dir = %q, want internal/models", p.ModelsInfo.Dir)
	}
}

func TestDiscoverIsDeterministic(t *testing.T) {
	fsys := tree(map[string]string{
		"go.mod":           goMod,
		"schema/003_c.sql": "",
		"schema/001_a.sql": "",
		"schema/002_b.sql": "",
		"z/queries.sql":    "",
		"a/queries.sql":    "",
		"m/n/queries.sql":  "",
		"m/queries.sql":    "",
		"a-c/queries.sql":  "",
		"a/b/queries.sql":  "",
	})
	first := mustDiscover(t, fsys)
	for i := 0; i < 20; i++ {
		if again := mustDiscover(t, fsys); !reflect.DeepEqual(first, again) {
			t.Fatalf("run %d differs:\nfirst: %+v\nagain: %+v", i+2, first, again)
		}
	}
}

/* ---------- go.mod ---------- */

func TestModulePathVariants(t *testing.T) {
	tests := []struct {
		name  string
		goMod string
		want  string
	}{
		{"plain", "module github.com/you/app\n", "github.com/you/app"},
		{"with go directive", "module github.com/you/app\n\ngo 1.27\n", "github.com/you/app"},
		{"quoted", "module \"github.com/you/app\"\n", "github.com/you/app"},
		{"trailing comment", "module github.com/you/app // the app\n", "github.com/you/app"},
		{"comment and blank lines first", "// my module\n\nmodule github.com/you/app\n", "github.com/you/app"},
		{"module not first directive", "go 1.27\nmodule github.com/you/app\n", "github.com/you/app"},
		{"with require block", "module github.com/you/app\n\ngo 1.27\n\nrequire (\n\tgithub.com/x/y v1.0.0\n)\n", "github.com/you/app"},
		{"extra spaces", "module    github.com/you/app   \n", "github.com/you/app"},
		{"single segment", "module example\n", "example"},
		{"no trailing newline", "module github.com/you/app", "github.com/you/app"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mustDiscover(t, tree(map[string]string{
				"go.mod":              tt.goMod,
				"schema/001_init.sql": "",
			}))
			if p.ModulePath != tt.want {
				t.Errorf("ModulePath = %q, want %q", p.ModulePath, tt.want)
			}
		})
	}
}

/* ---------- schema ---------- */

func TestSchemaFiles(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			"sorted by filename",
			map[string]string{"schema/010_c.sql": "", "schema/002_b.sql": "", "schema/001_a.sql": ""},
			[]string{"schema/001_a.sql", "schema/002_b.sql", "schema/010_c.sql"},
		},
		{
			"down migrations skipped",
			map[string]string{"schema/001_a.sql": "", "schema/001_a.down.sql": ""},
			[]string{"schema/001_a.sql"},
		},
		{
			"non-sql files ignored",
			map[string]string{"schema/001_a.sql": "", "schema/README.md": "", "schema/notes.sql.bak": ""},
			[]string{"schema/001_a.sql"},
		},
		{
			"subdirectories not searched",
			map[string]string{"schema/001_a.sql": "", "schema/old/000_legacy.sql": ""},
			[]string{"schema/001_a.sql"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.files["go.mod"] = goMod
			p := mustDiscover(t, tree(tt.files))
			if !reflect.DeepEqual(p.SchemaFiles, tt.want) {
				t.Errorf("SchemaFiles:\n got: %v\nwant: %v", p.SchemaFiles, tt.want)
			}
		})
	}
}

/* ---------- query packages ---------- */

func TestQueryPackages(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []Package
	}{
		{
			"deeply nested import path",
			map[string]string{"internal/a/b/c/q.sql": ""},
			[]Package{pkg("internal/a/b/c", "q.sql")},
		},
		{
			"files within a package are sorted",
			map[string]string{"x/c.sql": "", "x/a.sql": "", "x/b.sql": ""},
			[]Package{pkg("x", "a.sql", "b.sql", "c.sql")},
		},
		{
			"directory with only non-sql files is not a package",
			map[string]string{"docs/README.md": "", "docs/schema.png": ""},
			nil,
		},
		{
			"parent and child are separate packages",
			map[string]string{"a/q.sql": "", "a/b/q.sql": ""},
			[]Package{pkg("a", "q.sql"), pkg("a/b", "q.sql")},
		},
		{
			"packages sorted by dir, not walk order",
			map[string]string{"a/b/q.sql": "", "a-c/q.sql": ""},
			[]Package{pkg("a-c", "q.sql"), pkg("a/b", "q.sql")},
		},
		{
			"skip rules match exact names only",
			map[string]string{"vendored/q.sql": "", "my_testdata/q.sql": "", "schemas/q.sql": ""},
			[]Package{pkg("my_testdata", "q.sql"), pkg("schemas", "q.sql"), pkg("vendored", "q.sql")},
		},
		{
			"non-.sql suffixes ignored",
			map[string]string{"x/q.sql": "", "x/notes.sql.bak": "", "x/q.SQL.txt": ""},
			[]Package{pkg("x", "q.sql")},
		},
		{
			"directory named like a file is still a directory",
			map[string]string{"weird.sql/q.sql": ""},
			[]Package{pkg("weird.sql", "q.sql")},
		},
		{
			"nested schema dir is not the schema dir",
			map[string]string{"internal/schema/q.sql": ""},
			[]Package{pkg("internal/schema", "q.sql")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.files["go.mod"] = goMod
			tt.files["schema/001_init.sql"] = ""
			p := mustDiscover(t, tree(tt.files))
			checkPackages(t, p.QueryPackages, tt.want)
		})
	}
}

/* ---------- errors ---------- */

func TestDiscoverErrors(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		errContains string
	}{
		{
			"no go.mod",
			map[string]string{"schema/001_init.sql": ""},
			"go.mod",
		},
		{
			"go.mod without module line",
			map[string]string{"go.mod": "go 1.27\n", "schema/001_init.sql": ""},
			"module",
		},
		{
			"empty go.mod",
			map[string]string{"go.mod": "", "schema/001_init.sql": ""},
			"module",
		},
		{
			"no schema directory",
			map[string]string{"go.mod": goMod, "internal/a/q.sql": ""},
			"schema",
		},
		{
			"schema without sql files",
			map[string]string{"go.mod": goMod, "schema/README.md": ""},
			"schema",
		},
		{
			"schema with only down migrations",
			map[string]string{"go.mod": goMod, "schema/001_a.down.sql": ""},
			"schema",
		},
		{
			"sql at project root",
			map[string]string{"go.mod": goMod, "schema/001_init.sql": "", "queries.sql": ""},
			"queries.sql",
		},
		{
			"sql in models dir",
			map[string]string{"go.mod": goMod, "schema/001_init.sql": "", "internal/models/oops.sql": ""},
			"internal/models",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Discover(tree(tt.files))
			if err == nil {
				t.Fatalf("expected an error, got none\nresult: %+v", p)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error %q should mention %q", err, tt.errContains)
			}
		})
	}
}

/* ---------- FindRoot (real filesystem) ---------- */

// mkfile creates a file (and its parent dirs) under root.
func mkfile(t *testing.T, root, rel string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("module example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, root, rel string) string {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(full, 0o755); err != nil {
		t.Fatal(err)
	}
	return full
}

func TestFindRootFromDeepInside(t *testing.T) {
	root := t.TempDir()
	mkfile(t, root, "go.mod")
	start := mkdir(t, root, "internal/accounts/sub")

	got, err := FindRoot(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("FindRoot = %q, want %q", got, root)
	}
}

func TestFindRootFromRootItself(t *testing.T) {
	root := t.TempDir()
	mkfile(t, root, "go.mod")

	got, err := FindRoot(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("FindRoot = %q, want %q", got, root)
	}
}

func TestFindRootNearestWins(t *testing.T) {
	root := t.TempDir()
	mkfile(t, root, "go.mod")
	mkfile(t, root, "tools/go.mod") // nested module
	start := mkdir(t, root, "tools/gen")

	got, err := FindRoot(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(root, "tools"); got != want {
		t.Errorf("FindRoot = %q, want the nearest module %q", got, want)
	}
}

func TestFindRootIgnoresGoModDirectory(t *testing.T) {
	root := t.TempDir()
	mkfile(t, root, "go.mod")
	mkdir(t, root, "sub/go.mod") // a directory named go.mod, not a file
	start := mkdir(t, root, "sub/inner")

	got, err := FindRoot(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("FindRoot = %q, want %q (a go.mod directory doesn't count)", got, root)
	}
}

func TestFindRootRelativeStart(t *testing.T) {
	root := t.TempDir()
	mkfile(t, root, "go.mod")
	start := mkdir(t, root, "internal/accounts")
	t.Chdir(start)

	got, err := FindRoot(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Compare after resolving symlinks: on macOS, t.TempDir lives under
	// /var, which is a symlink to /private/var, and Getwd may report either.
	gotReal, _ := filepath.EvalSymlinks(got)
	wantReal, _ := filepath.EvalSymlinks(root)
	if gotReal != wantReal {
		t.Errorf("FindRoot(\".\") = %q, want %q", got, root)
	}
}

func TestFindRootNotFound(t *testing.T) {
	// Assumes no ancestor of the temp directory has a go.mod, which holds on
	// normal machines (temp dirs live under /tmp or /var/folders).
	start := mkdir(t, t.TempDir(), "a/b")
	if got, err := FindRoot(start); err == nil {
		t.Errorf("expected an error, got root %q", got)
	} else if !strings.Contains(err.Error(), "go.mod") {
		t.Errorf("error %q should mention go.mod", err)
	}
}
