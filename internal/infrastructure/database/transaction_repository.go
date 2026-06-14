package database

import (
	"context"
	"database/sql"
	"errors"
	"novobanco/internal/domain"

	"github.com/lib/pq"
)

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	query := `
		INSERT INTO transactions (account_id, reference_unique, amount, type, status, transfer_details)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, timestamp
	`
	exec := getExecutor(ctx, r.db)
	err := exec.QueryRowContext(ctx, query, t.AccountID, t.ReferenceUnique, t.Amount, t.Type, t.Status, t.TransferDetails).
		Scan(&t.ID, &t.Timestamp)

	if err != nil {
		// Capturar violación de índice único (llave de idempotencia duplicada)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrDuplicateReference
		}
		return err
	}
	return nil
}

func (r *PostgresTransactionRepository) GetByReferenceUnique(ctx context.Context, ref string) (*domain.Transaction, error) {
	query := `
		SELECT id, account_id, reference_unique, amount, type, status, transfer_details, timestamp
		FROM transactions WHERE reference_unique = $1
	`
	t := &domain.Transaction{}
	exec := getExecutor(ctx, r.db)
	err := exec.QueryRowContext(ctx, query, ref).Scan(
		&t.ID, &t.AccountID, &t.ReferenceUnique, &t.Amount, &t.Type, &t.Status, &t.TransferDetails, &t.Timestamp,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("transacción no encontrada")
	}
	return t, err
}

func (r *PostgresTransactionRepository) GetLastMovements(ctx context.Context, accountID string, limit int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, account_id, reference_unique, amount, type, status, transfer_details, timestamp
		FROM transactions
		WHERE account_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	exec := getExecutor(ctx, r.db)
	rows, err := exec.QueryContext(ctx, query, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		err = rows.Scan(&t.ID, &t.AccountID, &t.ReferenceUnique, &t.Amount, &t.Type, &t.Status, &t.TransferDetails, &t.Timestamp)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func (r *PostgresTransactionRepository) GetOutgoingTransferCount(ctx context.Context, clientID string, days int) (int, error) {
	query := `
		SELECT COUNT(t.id)
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE a.client_id = $1
		  AND t.type = 'TRANSFER_OUT'
		  AND t.timestamp >= NOW() - ($2 || ' days')::INTERVAL
	`
	exec := getExecutor(ctx, r.db)
	var count int
	err := exec.QueryRowContext(ctx, query, clientID, days).Scan(&count)
	return count, err
}
