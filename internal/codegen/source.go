package codegen

func _DbTxInterfaceSource() string {
	return `type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}`
}

func _TxStarterInterfaceSource() string {
	return `type TxStarter interface {
	BeginTx(ctx context.Context, ops *sql.TxOptions) (*sql.Tx, error)
}`
}

func _ErrNotFound() string {
	return `var ErrNotFound = sql.ErrNoRows`
}

func _InTxSource() string {
	return `func InTx(ctx context.Context, db TxStarter, fn func(tx DBTX) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}
	return tx.Commit()
}`
}
