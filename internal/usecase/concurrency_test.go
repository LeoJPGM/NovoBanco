package usecase

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"novobanco/internal/infrastructure/config"
	"novobanco/internal/infrastructure/database"
)

func TestConcurrencyWithdrawals(t *testing.T) {
	// 1. Carga configuración local e intenta conectar con la BD
	cfg := config.LoadConfig()
	db, err := database.NewPostgresDB(cfg.DBConn)
	if err != nil {
		// Si la base de datos no está corriendo en Docker, salta la prueba
		t.Skip("Saltando prueba de concurrencia: PostgreSQL en Docker no disponible.")
		return
	}
	defer db.Close()

	// 2. Inicializar repositorios y casos de uso
	accountRepo := database.NewPostgresAccountRepository(db)
	txRepo := database.NewPostgresTransactionRepository(db)
	txManager := database.NewSQLTxManager(db)
	txUsecase := NewTransactionUsecase(accountRepo, txRepo, txManager)
	accountUsecase := NewAccountUsecase(accountRepo)

	ctx := context.Background()

	// 3. Crear una cuenta de prueba con saldo de $100.00
	acc, err := accountUsecase.CreateAccount(ctx, "CONCURRENCY-CLIENT-TEST", "SAVINGS", 100.00)
	if err != nil {
		t.Fatalf("error al crear cuenta de prueba: %v", err)
	}

	// Limpieza de datos al finalizar el test
	defer func() {
		_, _ = db.Exec("DELETE FROM transactions WHERE account_id = $1", acc.ID)
		_, _ = db.Exec("DELETE FROM accounts WHERE id = $1", acc.ID)
	}()

	// 4. Lanzar 15 retiros de $10.00 en paralelo (Intentando debitar $150.00 de los $100.00 disponibles)
	numRequests := 15
	withdrawAmount := 10.00

	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Generar una referencia única para simular llamadas HTTP independientes
			ref := fmt.Sprintf("REF-CONC-%s-%d", acc.ID, index)

			_, err := txUsecase.Withdraw(ctx, acc.AccountNumber, withdrawAmount, ref)

			mu.Lock()
			if err != nil {
				failCount++
				t.Logf("Error en retiro %d %v", index, err)
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}

	// Espera a que terminen todas las goroutines
	wg.Wait()

	// 5. Validar estado final de la cuenta
	finalAcc, err := accountRepo.GetByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("error al consultar cuenta: %v", err)
	}

	t.Logf("=== Resultados del Test Concurrente ===")
	t.Logf("Intentos de retiro: %d ($10 cada uno)", numRequests)
	t.Logf("Retiros exitosos: %d (Esperado: 10)", successCount)
	t.Logf("Retiros fallidos: %d (Esperado: 5)", failCount)
	t.Logf("Saldo final en BD: $%.2f (Esperado: $0.00)", finalAcc.Balance)

	// Validaciones críticas:
	if finalAcc.Balance < 0 {
		t.Errorf("¡ERROR CRÍTICO! El saldo final es negativo: %f", finalAcc.Balance)
	}
	if successCount != 10 {
		t.Errorf("Se esperaba que exactamente 10 retiros tuvieran éxito, pero fueron %d", successCount)
	}
	if finalAcc.Balance != 0.00 {
		t.Errorf("Se esperaba saldo final de 0.00, pero se obtuvo %f", finalAcc.Balance)
	}
}
