package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/models"
)

type WithdrawalsRepository interface {
	Withdrawals(ctx context.Context, userID uuid.UUID) (withdrawal []models.Withdrawal, err error)
}

type WithdrawalsOut struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type WithdrawalsHandler struct {
	repo WithdrawalsRepository
}

func NewWithdrawalsHandler(repo WithdrawalsRepository) WithdrawalsHandler {
	return WithdrawalsHandler{
		repo: repo,
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

	withdrawals, err := handler.repo.Withdrawals(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	out := make([]WithdrawalsOut, 0, len(withdrawals))
	for i := range withdrawals {
		outElem := WithdrawalsOut{
			Order:       withdrawals[i].Number,
			Sum:         withdrawals[i].Sum,
			ProcessedAt: withdrawals[i].ProcessedAt.Format(time.RFC3339),
		}
		out = append(out, outElem)
	}

	resp, err := json.Marshal(out)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if len(out) == 0 {
		res.WriteHeader(http.StatusNoContent)
	} else {
		res.WriteHeader(http.StatusOK)
	}

	res.Write(resp)
}
