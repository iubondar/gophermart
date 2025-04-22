package handler

import (
	"encoding/json"
	"net/http"

	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/usecase"
)

type WithdrawalsHandler struct {
	uc usecase.WithdrawalsUsecase
}

func NewWithdrawalsHandler(uc usecase.WithdrawalsUsecase) *WithdrawalsHandler {
	return &WithdrawalsHandler{
		uc: uc,
	}
}

func (handler WithdrawalsHandler) Withdrawals(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.GetUserIDFromReq(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	withdrawals, err := handler.uc.GetWithdrawals(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if len(withdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(res).Encode(withdrawals); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
