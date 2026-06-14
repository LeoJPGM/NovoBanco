package http

import (
	"net/http"
	"novobanco/internal/infrastructure/http/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(accountHandler *handler.AccountHandler, txHandler *handler.TransactionHandler) http.Handler {
	r := chi.NewRouter()

	// Middlewares estándar para logging y recuperación ante pánicos
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	// endpoints de Cuentas
	r.Route("/api/accounts", func(r chi.Router) {
		r.Post("/", accountHandler.CreateAccount)
		r.Get("/balance", accountHandler.GetBalance)
		r.Get("/details", accountHandler.GetDetails)
	})

	// endpoints de Transacciones
	r.Route("/api/transactions", func(r chi.Router) {
		r.Post("/deposit", txHandler.Deposit)
		r.Post("/withdraw", txHandler.Withdraw)
		r.Post("/transfer", txHandler.Transfer)
		r.Get("/movements", txHandler.GetMovements)
	})

	return r
}
