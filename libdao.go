package libdao

import (
	"context"
	"database/sql"
)

type DB interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	SqlExecutorWithContext
}

type SqlExecutor interface {
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Rollback() error
	Stmt(stmt *sql.Stmt) *sql.Stmt
}

type SqlExecutorWithContext interface {
	SqlExecutor
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
}
