grammar should prob look smth like this:

STATEMENT: \[SELECT_STATEMENT, ;, EOF\]
SELECT_STATEMENT: \[SELECT, ...COLUMN_LIST | "*", FROM, TABLE_REF\]
COLUMN_LIST: ident ("," ident)*
TABLE_REF: ident
