package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAccountNotFound     = errors.New("cuenta no encontrada")
	ErrAccountInactive     = errors.New("la cuenta no está activa y no puede operar")
	ErrNegativeBalance     = errors.New("el saldo de la cuenta no puede ser negativo")
	ErrInsufficientBalance = errors.New("saldo insuficiente para realizar la transacción")
	ErrInvalidAmount       = errors.New("el monto de la transacción debe ser mayor a cero")
)

type AccountType string

const (
	Savings  AccountType = "SAVINGS"
	Checking AccountType = "CHECKING"
)

type AccountStatus string

const (
	Active  AccountStatus = "ACTIVE"
	Blocked AccountStatus = "BLOCKED"
	Closed  AccountStatus = "CLOSED"
)

// Account representa la entidad de cuenta bancaria
type Account struct {
	ID            string        `json:"id"`
	AccountNumber string        `json:"account_number"`
	ClientID      string        `json:"client_id"`
	Type          AccountType   `json:"type"`
	Currency      string        `json:"currency"`
	Balance       float64       `json:"balance"`
	Status        AccountStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// canOperate verifica si la cuenta está activa
func (a *Account) CanOperate() bool {
	return a.Status == Active
}

// AccountRepository define los contratos para persistir cuentas.
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*Account, error)

	// GetByAccountNumberForUpdate bloquea la fila en la BD para evitar condiciones de carrera (SELECT FOR UPDATE)
	GetByAccountNumberForUpdate(ctx context.Context, accountNumber string) (*Account, error)
	UpdateBalance(ctx context.Context, id string, newBalance float64) error
}
