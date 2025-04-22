package handler

import (
	"encoding/json"
	"net/http"

	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/usecase"
	"go.uber.org/zap"
)

type BalanceHandler struct {
	uc usecase.GetBalanceUsecase
}

func NewBalanceHandler(uc usecase.GetBalanceUsecase) BalanceHandler {
	return BalanceHandler{
		uc: uc,
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

	out, err := handler.uc.GetBalance(req.Context(), userID)
	if err != nil {
		zap.L().Sugar().Debugln("Error get balance:", err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(out)
	if err != nil {
		zap.L().Sugar().Debugln("Error marshal balance:", err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
