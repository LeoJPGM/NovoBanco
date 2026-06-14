package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"novobanco/internal/domain"
	"novobanco/internal/usecase"
	"strconv"
)

type TransactionHandler struct {
	txUsecase *usecase.TransactionUsecase
}

func NewTransactionHandler(u *usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{txUsecase: u}
}

type DepositRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
	Reference     string  `json:"reference"`
}

func (h *TransactionHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de petición inválido")
		return
	}

	if req.AccountNumber == "" || req.Reference == "" {
		respondWithError(w, http.StatusBadRequest, "los campos account_number y reference son requeridos")
		return
	}

	t, err := h.txUsecase.Deposit(r.Context(), req.AccountNumber, req.Amount, req.Reference)
	if err != nil {
		h.handleTransactionError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

type WithdrawRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
	Reference     string  `json:"reference"`
}

func (h *TransactionHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de petición inválido")
		return
	}

	if req.AccountNumber == "" || req.Reference == "" {
		respondWithError(w, http.StatusBadRequest, "los campos account_number y reference son requeridos")
		return
	}

	t, err := h.txUsecase.Withdraw(r.Context(), req.AccountNumber, req.Amount, req.Reference)
	if err != nil {
		h.handleTransactionError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

type TransferRequest struct {
	FromAccountNumber string  `json:"from_account_number"`
	ToAccountNumber   string  `json:"to_account_number"`
	Amount            float64 `json:"amount"`
	Reference         string  `json:"reference"`
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de petición inválido")
		return
	}

	if req.FromAccountNumber == "" || req.ToAccountNumber == "" || req.Reference == "" {
		respondWithError(w, http.StatusBadRequest, "los campos from_account_number, to_account_number y reference son requeridos")
		return
	}

	t, err := h.txUsecase.Transfer(r.Context(), req.FromAccountNumber, req.ToAccountNumber, req.Amount, req.Reference)
	if err != nil {
		h.handleTransactionError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

func (h *TransactionHandler) GetMovements(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.URL.Query().Get("account_number")
	if accountNumber == "" {
		respondWithError(w, http.StatusBadRequest, "el parámetro account_number es requerido")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // Requerimiento: por defecto últimos 20 movimientos
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	movements, err := h.txUsecase.GetMovements(r.Context(), accountNumber, limit)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, movements)
}

// handleTransactionError mapea errores del dominio a códigos HTTP correctos
func (h *TransactionHandler) handleTransactionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		respondWithError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrAccountInactive):
		respondWithError(w, http.StatusForbidden, err.Error()) // Cuenta inactiva/bloqueada -> 403
	case errors.Is(err, domain.ErrInsufficientBalance):
		respondWithError(w, http.StatusUnprocessableEntity, err.Error()) // Saldo insuficiente -> 422
	case errors.Is(err, domain.ErrInvalidAmount):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrDuplicateReference):
		respondWithError(w, http.StatusConflict, err.Error()) // Petición duplicada (idempotencia) -> 409
	default:
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}
