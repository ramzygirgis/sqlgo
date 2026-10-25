# sqlgo

Generate type-safe Go code from raw SQL, in the spirit of [sqlc](https://github.com/sqlc-dev/sqlc).

You rip raw SQL; sqlgo reads your schema and generates go functions with typed parameters and result structs. All the type inference and mapping happens statically, so there's no reflection or ORM overhead at runtime.

> **Status:** ongoing. The lexer and parser work for the supported SQL listed below. Schema replay and code generation are in progress.

## Goals

- Turn a wide range of raw SQL queries into generated Go code, with support for named params.
- Replay all migrations to reconstruct the current schema, so generated code can be typechecked against it.
- Maybe: generate docs from the schema and queries.

## Supported SQL

| Feature | Status |
|---|---|
| `SELECT col, ... FROM table` / `SELECT *` | ✅ |
| `WHERE` with `=`, `<>`, `!=`, `<`, `>`, `<=`, `>=` | ✅ |
| `AND` / `OR` with correct precedence, parentheses | ✅ |
| String, number, and `$1`-style parameter literals | ✅ |
| `CREATE TABLE [IF NOT EXISTS]` with `NOT NULL`, `PRIMARY KEY`, `DEFAULT` | ✅ |
| `-- name: Foo :many` query annotations | in progress |
| Migrations (`ALTER TABLE`, etc.) | in progress |
| Code generation | in progress |

## Repo structure

```
internal/
  parsing/     lexer, parser, and AST types: raw SQL → AST
  ir/          (planned) resolved, typed representation of queries
  generator/   (planned) IR → Go source
docs/          supported syntax and design notes
main.go        dev entry point: runs sample queries through the parser
```

The planned pipeline:

```
SQL → lexer → tokens → parser → AST → analyzer + catalog → IR → generator → Go
```

## Development

```sh
go test ./...            # run all tests
go run .                 # parse the sample queries in main.go
go run . -lex            # also print tokens
```
