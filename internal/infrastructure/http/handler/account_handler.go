package handler

import (
	"encoding/json"
	"net/http"
	"novobanco/internal/domain"
	"novobanco/internal/usecase"
)

type AccountHandler struct {
	accountUsecase *usecase.AccountUsecase
}

func NewAccountHandler(u *usecase.AccountUsecase) *AccountHandler {
	return &AccountHandler{accountUsecase: u}
}

type CreateAccountRequest struct {
	ClientID string  `json:"client_id"`
	Type     string  `json:"type"`
	Balance  float64 `json:"balance"`
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de petición inválido")
		return
	}

	if req.ClientID == "" {
		respondWithError(w, http.StatusBadRequest, "el campo client_id es requerido")
		return
	}

	acc, err := h.accountUsecase.CreateAccount(r.Context(), req.ClientID, req.Type, req.Balance)
	if err != nil {
		if err == domain.ErrNegativeBalance {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, acc)
}

func (h *AccountHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.URL.Query().Get("account_number")
	if accountNumber == "" {
		respondWithError(w, http.StatusBadRequest, "el parámetro account_number es requerido")
		return
	}

	balance, err := h.accountUsecase.GetBalance(r.Context(), accountNumber)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"account_number": accountNumber,
		"balance":        balance,
		"currency":       "USD",
	})
}

func (h *AccountHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.URL.Query().Get("account_number")
	if accountNumber == "" {
		respondWithError(w, http.StatusBadRequest, "el parámetro account_number es requerido")
		return
	}

	acc, err := h.accountUsecase.GetAccountDetails(r.Context(), accountNumber)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, acc)
}

// Helpers para responder en formato JSON
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
