package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/models"
)

type BalanceRepository interface {
	Account(ctx context.Context, userID uuid.UUID) (account models.Account, err error)
}

type BalanceOut struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type BalanceHandler struct {
	repo BalanceRepository
}

func NewBalanceHandler(repo BalanceRepository) BalanceHandler {
	return BalanceHandler{
		repo: repo,
	}
}

func (handler BalanceHandler) Balance(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.GetUserIDFromReq(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	account, err := handler.repo.Account(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	out := BalanceOut{
		Current:   account.Balance,
		Withdrawn: account.WithdrawalSum,
	}

	resp, err := json.Marshal(out)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
