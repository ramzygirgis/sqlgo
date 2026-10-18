goals:
- parse a wide range of raw sql queries into generated go code, with support for named parameters.
- generate go code for executing these queries, allowing for better performance than an ORM since a lot of the type work is done at compile time.
- replay all migrations to construct the current schema in order to enable type safety.
- maybe generate docs?
