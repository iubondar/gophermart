package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"go.uber.org/zap"
)

type Withdrawer interface {
	Withdraw(ctx context.Context, userID uuid.UUID, orderNumber string, sum float32) (result constants.WithdrawResult, err error)
}

type WithdrawIn struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

type WithdrawHandler struct {
	withdrawer Withdrawer
}

func NewWithdrawHandler(withdrawer Withdrawer) *WithdrawHandler {
	return &WithdrawHandler{
		withdrawer: withdrawer,
	}
}

func (handler WithdrawHandler) Withdraw(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.GetUserIDFromReq(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	var in WithdrawIn
	var buf bytes.Buffer
	// читаем тело запроса
	_, err = buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// десериализуем JSON
	if err = json.Unmarshal(buf.Bytes(), &in); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate input
	if in.Order == "" {
		http.Error(res, "Order number is required", http.StatusBadRequest)
		return
	}

	if in.Sum <= 0 {
		http.Error(res, "Sum must be greater than 0", http.StatusBadRequest)
		return
	}

	result, err := handler.withdrawer.Withdraw(req.Context(), userID, in.Order, in.Sum)
	if err != nil {
		zap.L().Sugar().Debugln("Error withdrawing balance:", err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	switch result {
	case constants.Success:
		res.WriteHeader(http.StatusOK)
	case constants.InsufficientFunds:
		res.WriteHeader(http.StatusPaymentRequired)
	case constants.WrongOrderFormat:
		res.WriteHeader(http.StatusUnprocessableEntity)
	default:
		http.Error(res, "Wrong format", http.StatusBadRequest)
	}
}
