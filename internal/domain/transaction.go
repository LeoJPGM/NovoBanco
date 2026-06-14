package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDuplicateReference = errors.New("referencia única de transaccion ya existe (petición duplicada)")
)

type TransactionType string

const (
	Deposit     TransactionType = "DEPOSIT"
	Withdrawal  TransactionType = "WITHDRAWAL"
	TransferOut TransactionType = "TRANSFER_OUT"
	TransferIn  TransactionType = "TRANSFER_IN"
)

type TransactionStatus string

const (
	Successful TransactionStatus = "SUCCESSFUL"
	Failed     TransactionStatus = "FAILED"
	Reverted   TransactionStatus = "REVERTED"
)

type Transaction struct {
	ID              string            `json:"id"`
	AccountID       string            `json:"account_id"`
	ReferenceUnique string            `json:"reference_unique"`
	Amount          float64           `json:"amount"`
	Type            TransactionType   `json:"type"`
	Status          TransactionStatus `json:"status"`
	TransferDetails *string           `json:"transfer_details,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// TransactionRepository define los contratos para persistir transacciones
type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	GetByReferenceUnique(ctx context.Context, ref string) (*Transaction, error)
	GetLastMovements(ctx context.Context, accountID string, limit int) ([]*Transaction, error)
	GetOutgoingTransferCount(ctx context.Context, clientID string, days int) (int, error)
}

// TransactionManager maneja transacciones de BD de forma desacoplada usando Contexto
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
