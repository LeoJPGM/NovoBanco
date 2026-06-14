package usecase

import (
	"context"
	"errors"
	"novobanco/internal/domain"
	"testing"
)

// mockAccountRepository implementa domain.AccountRepository en memoria para pruebas unitarias
type mockAccountRepository struct {
	accounts  map[string]*domain.Account
	createErr error
}

func (m *mockAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.accounts[account.AccountNumber] = account
	return nil
}

func (m *mockAccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	for _, acc := range m.accounts {
		if acc.ID == id {
			return acc, nil
		}
	}
	return nil, domain.ErrAccountNotFound
}

func (m *mockAccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*domain.Account, error) {
	acc, exists := m.accounts[accountNumber]
	if !exists {
		return nil, domain.ErrAccountNotFound
	}
	return acc, nil
}

func (m *mockAccountRepository) GetByAccountNumberForUpdate(ctx context.Context, accountNumber string) (*domain.Account, error) {
	return m.GetByAccountNumber(ctx, accountNumber)
}

func (m *mockAccountRepository) UpdateBalance(ctx context.Context, id string, newBalance float64) error {
	for _, acc := range m.accounts {
		if acc.ID == id {
			acc.Balance = newBalance
			return nil
		}
	}
	return errors.New("cuenta no encontrada")
}

// Tests unitarios de cuenta
func TestCreateAccount(t *testing.T) {
	repo := &mockAccountRepository{accounts: make(map[string]*domain.Account)}
	uc := NewAccountUsecase(repo)

	// Test caso exitoso
	acc, err := uc.CreateAccount(context.Background(), "CLIENT-1", "SAVINGS", 100.0)
	if err != nil {
		t.Fatalf("se esperaba que no falle, obtuvo error: %v", err)
	}

	if acc.ClientID != "CLIENT-1" {
		t.Errorf("se esperaba cliente CLIENT-1, obtuvo %s", acc.ClientID)
	}
	if acc.Balance != 100.0 {
		t.Errorf("se esperaba saldo 100, obtuvo %f", acc.Balance)
	}

	// Test saldo inicial negativo
	_, err = uc.CreateAccount(context.Background(), "CLIENT-1", "SAVINGS", -50.0)
	if !errors.Is(err, domain.ErrNegativeBalance) {
		t.Errorf("se esperaba error de saldo negativo, obtuvo: %v", err)
	}
}
