package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

// Key del contexto para guardar la transacción
type txKey struct{}

var activeTxKey = txKey{}

// ConnectionPool envuelve el pool de conexiones sql.DB
type ConnectionPool struct {
	DB *sql.DB
}

// NewPostgresDB crea un nuevo pool de conexiones a la base de datos
func NewPostgresDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// SQLTxManager implementa domain.TransactionManager usando database/sql
type SQLTxManager struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) *SQLTxManager {
	return &SQLTxManager{db: db}
}

// ExecuteInTransaction ejecuta una función dentro de una transacción de BD SQL
func (m *SQLTxManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})

	if err != nil {
		return err
	}

	// Guardar la transacción en el contexto
	txCtx := context.WithValue(ctx, activeTxKey, tx)

	err = fn(txCtx)
	if err != nil {
		// Si hay un error, rollback la transacción
		_ = tx.Rollback()
		return err
	}

	// Si todo sale bien, commit la transacción
	return tx.Commit()
}

// GetTxOrDB extrae la transacción activa del contexto si existe, sino devuelve el pool DB original
type queryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func getExecutor(ctx context.Context, db *sql.DB) queryExecutor {
	if tx, ok := ctx.Value(activeTxKey).(*sql.Tx); ok {
		return tx
	}
	return db
}
