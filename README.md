goals:
- parse a wide range of raw sql queries into generated go code, with support for named parameters. this allows for better runtime performance than an ORM since you don't need to infer types or do any mapping at runtime.
- replay all migrations to construct the current schema in order to enable type safety.
- maybe generate docs?

---

repo structure:
- internal/parsing: contains all logic and types for turning raw sql into AST types.
- ...
