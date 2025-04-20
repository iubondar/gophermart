package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/usecase"
	"go.uber.org/zap"
)

type WithdrawHandler struct {
	uc usecase.WithdrawUsecase
}

func NewWithdrawHandler(uc usecase.WithdrawUsecase) *WithdrawHandler {
	return &WithdrawHandler{
		uc: uc,
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

	var in usecase.WithdrawIn
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

	result, err := handler.uc.Withdraw(req.Context(), userID, in)
	if err != nil {
		zap.L().Sugar().Debugln("Error withdrawing balance:", err.Error())
		http.Error(res, "Internal server error", http.StatusInternalServerError)
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
