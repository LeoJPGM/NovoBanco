package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"novobanco/internal/domain"
)

type TransactionUsecase struct {
	accountRepo     domain.AccountRepository
	transactionRepo domain.TransactionRepository
	txManager       domain.TransactionManager
}

func NewTransactionUsecase(
	accountRepo domain.AccountRepository,
	transactionRepo domain.TransactionRepository,
	txManager domain.TransactionManager,
) *TransactionUsecase {
	return &TransactionUsecase{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		txManager:       txManager,
	}
}

// Deposit abona dinero a una cuenta activa
func (u *TransactionUsecase) Deposit(ctx context.Context, accountNumber string, amount float64, reference string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	var transaction *domain.Transaction

	// Ejecuta toda la lógica dentro de una transacción ACID de base de datos
	err := u.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Buscar y bloquear la cuenta (SELECT FOR UPDATE)
		acc, err := u.accountRepo.GetByAccountNumberForUpdate(txCtx, accountNumber)
		if err != nil {
			return err
		}

		// 2. Validar estado de cuenta
		if !acc.CanOperate() {
			return domain.ErrAccountInactive
		}

		// 3. Modificar saldo en base de datos
		newBalance := acc.Balance + amount
		err = u.accountRepo.UpdateBalance(txCtx, acc.ID, newBalance)
		if err != nil {
			return err
		}

		// 4. Crear registro del movimiento
		t := &domain.Transaction{
			AccountID:       acc.ID,
			ReferenceUnique: reference,
			Amount:          amount,
			Type:            domain.Deposit,
			Status:          domain.Successful,
		}

		err = u.transactionRepo.Create(txCtx, t)
		if err != nil {
			return err
		}

		transaction = t
		return nil
	})

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// Withdraw retira dinero de una cuenta activa con saldo suficiente
func (u *TransactionUsecase) Withdraw(ctx context.Context, accountNumber string, amount float64, reference string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	var transaction *domain.Transaction

	err := u.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Buscar y bloquear la cuenta (SELECT FOR UPDATE)
		acc, err := u.accountRepo.GetByAccountNumberForUpdate(txCtx, accountNumber)
		if err != nil {
			return err
		}

		// 2. Validaciones de negocio
		if !acc.CanOperate() {
			return domain.ErrAccountInactive
		}
		if acc.Balance < amount {
			return domain.ErrInsufficientBalance
		}

		// 3. Modificar saldo
		newBalance := acc.Balance - amount
		err = u.accountRepo.UpdateBalance(txCtx, acc.ID, newBalance)
		if err != nil {
			return err
		}

		// 4. Crear registro del movimiento
		t := &domain.Transaction{
			AccountID:       acc.ID,
			ReferenceUnique: reference,
			Amount:          amount,
			Type:            domain.Withdrawal,
			Status:          domain.Successful,
		}

		err = u.transactionRepo.Create(txCtx, t)
		if err != nil {
			return err
		}

		transaction = t
		return nil
	})

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// Transfer mueve fondos entre dos cuentas de forma atómica y previene deadlocks
func (u *TransactionUsecase) Transfer(ctx context.Context, fromNumber, toNumber string, amount float64, reference string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if fromNumber == toNumber {
		return nil, fmt.Errorf("la cuenta de origen y destino deben ser distintas")
	}

	var txOutRecord *domain.Transaction

	err := u.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// =====================================================================
		// EVITAR DEADLOCKS: ORDENAMIENTO DETERMINISTA DE BLOQUEOS (LOCK ORDERING)
		// =====================================================================
		var firstAccount, secondAccount *domain.Account
		var err error

		// Siempre bloqueamos las cuentas en orden alfabético/numérico de su número de cuenta.
		// Así, si dos peticiones ocurren simultáneamente A->B y B->A, ambas intentarán
		// bloquear la misma cuenta primero (la menor), evitando un deadlock en Postgres.
		if fromNumber < toNumber {
			firstAccount, err = u.accountRepo.GetByAccountNumberForUpdate(txCtx, fromNumber)
			if err != nil {
				return err
			}
			secondAccount, err = u.accountRepo.GetByAccountNumberForUpdate(txCtx, toNumber)
			if err != nil {
				return err
			}
		} else {
			secondAccount, err = u.accountRepo.GetByAccountNumberForUpdate(txCtx, toNumber)
			if err != nil {
				return err
			}
			firstAccount, err = u.accountRepo.GetByAccountNumberForUpdate(txCtx, fromNumber)
			if err != nil {
				return err
			}
		}

		// Mapear de vuelta quién es el origen y quién es el destino
		var fromAcc, toAcc *domain.Account
		if firstAccount.AccountNumber == fromNumber {
			fromAcc = firstAccount
			toAcc = secondAccount
		} else {
			fromAcc = secondAccount
			toAcc = firstAccount
		}

		// =====================================================================
		// VALIDACIONES DE NEGOCIO
		// =====================================================================
		if !fromAcc.CanOperate() {
			return fmt.Errorf("cuenta origen inactiva: %w", domain.ErrAccountInactive)
		}
		if !toAcc.CanOperate() {
			return fmt.Errorf("cuenta destino inactiva: %w", domain.ErrAccountInactive)
		}
		if fromAcc.Balance < amount {
			return domain.ErrInsufficientBalance
		}

		// =====================================================================
		// ACTUALIZACIÓN DE SALDOS
		// =====================================================================
		err = u.accountRepo.UpdateBalance(txCtx, fromAcc.ID, fromAcc.Balance-amount)
		if err != nil {
			return err
		}

		err = u.accountRepo.UpdateBalance(txCtx, toAcc.ID, toAcc.Balance+amount)
		if err != nil {
			return err
		}

		// =====================================================================
		// REGISTRO DE TRANSACCIONES (Ambas Cuentas)
		// =====================================================================

		// Registro de salida (Débito)
		detailsFrom := map[string]string{"destination_account": toNumber}
		detailsBytesFrom, _ := json.Marshal(detailsFrom)
		detailsStrFrom := string(detailsBytesFrom)

		tOut := &domain.Transaction{
			AccountID:       fromAcc.ID,
			ReferenceUnique: reference,
			Amount:          amount,
			Type:            domain.TransferOut,
			Status:          domain.Successful,
			TransferDetails: &detailsStrFrom,
		}
		err = u.transactionRepo.Create(txCtx, tOut)
		if err != nil {
			return err
		}

		// Registro de entrada (Crédito)
		// Usamos un sufijo "-IN" para la referencia única para evitar colisionar con la de salida
		detailsTo := map[string]string{"source_account": fromNumber}
		detailsBytesTo, _ := json.Marshal(detailsTo)
		detailsStrTo := string(detailsBytesTo)

		tIn := &domain.Transaction{
			AccountID:       toAcc.ID,
			ReferenceUnique: reference + "-IN",
			Amount:          amount,
			Type:            domain.TransferIn,
			Status:          domain.Successful,
			TransferDetails: &detailsStrTo,
		}
		err = u.transactionRepo.Create(txCtx, tIn)
		if err != nil {
			return err
		}

		txOutRecord = tOut
		return nil
	})

	if err != nil {
		return nil, err
	}

	return txOutRecord, nil
}

// GetMovements obtiene los movimientos recientes de una cuenta
func (u *TransactionUsecase) GetMovements(ctx context.Context, accountNumber string, limit int) ([]*domain.Transaction, error) {
	acc, err := u.accountRepo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	return u.transactionRepo.GetLastMovements(ctx, acc.ID, limit)
}
