package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"novobanco/internal/domain"
	"time"
)

type AccountUsecase struct {
	accountRepo domain.AccountRepository
}

func NewAccountUsecase(accountRepo domain.AccountRepository) *AccountUsecase {
	return &AccountUsecase{accountRepo: accountRepo}
}

// CreateAccount crea una nueva cuenta bancaria
func (u *AccountUsecase) CreateAccount(ctx context.Context, clientID string, accountType string, initialBalance float64) (*domain.Account, error) {
	if initialBalance < 0 {
		return nil, domain.ErrNegativeBalance
	}

	accType := domain.AccountType(accountType)
	if accType != domain.Savings && accType != domain.Checking {
		return nil, fmt.Errorf("tipo de cuenta inválido: debe ser SAVINGS o CHECKING")
	}

	// Autogenerar número de cuenta único (formato CTA-XXXXXXXXXX)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	accNumber := fmt.Sprintf("CTA-%010d", r.Int63n(10000000000))

	acc := &domain.Account{
		AccountNumber: accNumber,
		ClientID:      clientID,
		Type:          accType,
		Currency:      "USD",
		Balance:       initialBalance,
		Status:        domain.Active,
	}

	err := u.accountRepo.Create(ctx, acc)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

// GetBalance obtiene el saldo actual de una cuenta
func (u *AccountUsecase) GetBalance(ctx context.Context, accountNumber string) (float64, error) {
	acc, err := u.accountRepo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return 0, err
	}
	return acc.Balance, nil
}

// GetAccountDetails obtiene toda la información de una cuenta
func (u *AccountUsecase) GetAccountDetails(ctx context.Context, accountNumber string) (*domain.Account, error) {
	return u.accountRepo.GetByAccountNumber(ctx, accountNumber)
}
