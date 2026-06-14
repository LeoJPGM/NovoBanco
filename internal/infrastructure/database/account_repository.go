package database

import (
	"context"
	"database/sql"
	"errors"
	"novobanco/internal/domain"
)

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) Create(ctx context.Context, acc *domain.Account) error {
	query := `
		INSERT INTO accounts (account_number, client_id, type, currency, balance, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	exec := getExecutor(ctx, r.db)
	err := exec.QueryRowContext(ctx, query, acc.AccountNumber, acc.ClientID, acc.Type, acc.Currency, acc.Balance, acc.Status).
		Scan(&acc.ID, &acc.CreatedAt, &acc.UpdatedAt)
	return err
}

func (r *PostgresAccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	query := `
		SELECT id, account_number, client_id, type, currency, balance, status, created_at, updated_at
		FROM accounts WHERE id = $1
	`
	acc := &domain.Account{}
	exec := getExecutor(ctx, r.db)
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&acc.ID, &acc.AccountNumber, &acc.ClientID, &acc.Type, &acc.Currency, &acc.Balance, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	return acc, err
}

func (r *PostgresAccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*domain.Account, error) {
	query := `
		SELECT id, account_number, client_id, type, currency, balance, status, created_at, updated_at
		FROM accounts WHERE account_number = $1
	`
	acc := &domain.Account{}
	exec := getExecutor(ctx, r.db)
	err := exec.QueryRowContext(ctx, query, accountNumber).Scan(
		&acc.ID, &acc.AccountNumber, &acc.ClientID, &acc.Type, &acc.Currency, &acc.Balance, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	return acc, err
}

func (r *PostgresAccountRepository) GetByAccountNumberForUpdate(ctx context.Context, accountNumber string) (*domain.Account, error) {
	// SELECT FOR UPDATE bloquea la fila en la BD hasta que termine la transacción
	query := `
		SELECT id, account_number, client_id, type, currency, balance, status, created_at, updated_at
		FROM accounts WHERE account_number = $1 FOR UPDATE
	`
	acc := &domain.Account{}
	exec := getExecutor(ctx, r.db)
	err := exec.QueryRowContext(ctx, query, accountNumber).Scan(
		&acc.ID, &acc.AccountNumber, &acc.ClientID, &acc.Type, &acc.Currency, &acc.Balance, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	return acc, err
}

func (r *PostgresAccountRepository) UpdateBalance(ctx context.Context, id string, newBalance float64) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	exec := getExecutor(ctx, r.db)
	_, err := exec.ExecContext(ctx, query, newBalance, id)
	return err
}
