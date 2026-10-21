currently supports standalone queries of the form:
`SELECT \[names \cup * \] FROM <tableName>;`
`CREATE TABLE <tableName> (
  rowName1 type1 (...constraints) (DEFAULT defaultVal),
  ...
);`

supported constraints:
- NOT NULL
- PRIMARY KEY


supported types:
- TEXT
- INT
